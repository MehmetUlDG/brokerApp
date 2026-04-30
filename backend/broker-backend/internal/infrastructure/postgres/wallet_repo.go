package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"github.com/yourusername/broker-backend/internal/domain"
)


// =============================================================================
// walletRepository — WalletRepository implementasyonu (PostgreSQL + sqlx)
// =============================================================================

// walletRepository, domain.WalletRepository interface'ini PostgreSQL üzerinde
// sqlx kütüphanesi ile uygular.
//
// Thread Safety:
//   - sqlx.DB bağlantı havuzu thread-safe'dir; struct alanları immutable'dır.
//   - Eş zamanlı çağrılar güvenle yapılabilir.
//
// Race Condition Stratejisi:
//   - Bakiye okuma+yazma operasyonları her zaman tek bir sqlx.Tx içinde yapılır.
//   - Satır kilidi: SELECT ... FOR UPDATE → diğer tx'ler satırı kilitli bulur,
//     transaction commit/rollback olana kadar bekler (Pessimistic Locking).
//   - "Lost Update" problemi bu yaklaşımla tamamen engellenir.
type walletRepository struct {
	db *sqlx.DB
}

// NewWalletRepository, yeni bir walletRepository örneği döner.
// db sıfır değer olamaz; aksi hâlde panic oluşur.
func NewWalletRepository(db *sqlx.DB) domain.WalletRepository {
	if db == nil {
		panic("wallet repository: sqlx.DB nil olamaz")
	}
	return &walletRepository{db: db}
}

// =============================================================================
// GetByUserID
// =============================================================================

// GetByUserID, verilen kullanıcıya ait cüzdanı okur.
// Kilit ALMAZ — salt okunur sorgular için kullanılmalıdır.
func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	const query = `
		SELECT id, user_id, balance, btc_balance, updated_at
		FROM   wallets
		WHERE  user_id = $1
		LIMIT  1
	`

	var w domain.Wallet
	err := r.db.GetContext(ctx, &w, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("wallet GetByUserID: %w", err)
	}
	return &w, nil
}

// =============================================================================
// GetByUserIDForUpdate — KRİTİK: Pessimistic Lock
// =============================================================================

// GetByUserIDForUpdate, cüzdanı okur VE satır kilidini alır.
// Yalnızca aktif bir sqlx.Tx içinden çağrılmalıdır; aksi hâlde kilit
// hemen serbest bırakılır ve koruma anlamsız olur.
//
// SELECT ... FOR UPDATE davranışı:
//   - Satırı okuyan diğer tx'ler kilidi alana kadar BLOKE olur.
//   - NOWAIT veya SKIP LOCKED ile gereksiz bekleme önlenebilir (gerekirse eklenebilir).
//   - Bu tx commit veya rollback olduğunda kilit otomatik kaldırılır.
func (r *walletRepository) GetByUserIDForUpdate(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	// Bu metot yalnızca bir tx bağlamında çalışır — ancak tx nesnesini
	// doğrudan almaz; çağıran UpdateBalance içinde tx üzerinden devredilir.
	// Bağımsız kullanım için internal yardımcı metot kullanılır.
	return r.getByUserIDForUpdateOnTx(ctx, r.db, userID)
}

// getByUserIDForUpdateOnTx, sqlx.ExtContext arayüzü üzerinden çalışır.
// Hem *sqlx.DB hem *sqlx.Tx bu arayüzü karşıladığından kod tekrarı önlenir.
func (r *walletRepository) getByUserIDForUpdateOnTx(
	ctx context.Context,
	ext sqlx.ExtContext,
	userID uuid.UUID,
) (*domain.Wallet, error) {
	// SELECT ... FOR UPDATE:
	// PostgreSQL'de bu ifade, eşleşen satırlara exclusive row-level lock uygular.
	// Aynı satırı güncellemek isteyen başka transaction'lar bu lock serbest
	// bırakılana kadar bekler → "Lost Update" impossible.
	const query = `
		SELECT id, user_id, balance, btc_balance, updated_at
		FROM   wallets
		WHERE  user_id = $1
		LIMIT  1
		FOR UPDATE
	`

	var w domain.Wallet
	err := sqlx.GetContext(ctx, ext, &w, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("wallet GetByUserIDForUpdate: %w", err)
	}
	return &w, nil
}

// =============================================================================
// Create
// =============================================================================

// Create, sıfır bakiyeyle yeni bir cüzdan oluşturur.
func (r *walletRepository) Create(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	const query = `
		INSERT INTO wallets (id, user_id, balance, btc_balance, updated_at)
		VALUES ($1, $2, 0, 0, NOW())
		RETURNING id, user_id, balance, btc_balance, updated_at
	`

	newID := uuid.New()
	var w domain.Wallet
	err := r.db.QueryRowxContext(ctx, query, newID, userID).StructScan(&w)
	if err != nil {
		// UNIQUE kısıtı ihlali: user_id zaten kayıtlı
		if isUniqueViolation(err) {
			return nil, domain.ErrWalletAlreadyExists
		}
		return nil, fmt.Errorf("wallet Create: %w", err)
	}
	return &w, nil
}

// =============================================================================
// CreateWithTx
// =============================================================================

// CreateWithTx, aktif bir transaction içinde cüzdan oluşturur.
// tx parametresi *sqlx.Tx olmalıdır; tip casting hatalıysa error döner.
func (r *walletRepository) CreateWithTx(ctx context.Context, txArg interface{}, userID uuid.UUID) (*domain.Wallet, error) {
	tx, ok := txArg.(*sqlx.Tx)
	if !ok {
		return nil, fmt.Errorf("wallet CreateWithTx: geçersiz transaction tipi (%T)", txArg)
	}

	const query = `
		INSERT INTO wallets (id, user_id, balance, btc_balance, updated_at)
		VALUES ($1, $2, 0, 0, NOW())
		RETURNING id, user_id, balance, btc_balance, updated_at
	`

	newID := uuid.New()
	var w domain.Wallet
	err := tx.QueryRowxContext(ctx, query, newID, userID).StructScan(&w)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrWalletAlreadyExists
		}
		return nil, fmt.Errorf("wallet CreateWithTx: %w", err)
	}
	return &w, nil
}

// =============================================================================
// UpdateBalance — ANA METOT: Transaction + Pessimistic Lock
// =============================================================================

// UpdateBalance, bir kullanıcının belirtilen bakiye alanını güvenli şekilde günceller.
//
// # Eşzamanlılık Güvenliği (Race Condition Prevention)
//
// Problem: İki eş zamanlı istek aynı kullanıcının bakiyesini okur (her ikisi 100$
// görür), ikisi de 80$ düşerek -60$'a düşme riski yaratır.
//
// Çözüm — Pessimistic Locking ile adım adım:
//
//  1. BEGIN → Yeni bir transaction başlatılır.
//  2. SELECT ... FOR UPDATE → Satır exclusive olarak kilitlenir.
//     İkinci eş zamanlı tx aynı satırı kilitlemeye çalışır ve BLOKE olur.
//  3. Bakiye kontrolü (DEBIT için): uygulama katmanında yeterlilik doğrulanır.
//  4. UPDATE → Bakiye güncellenir, updated_at anında tazenlenir.
//  5. COMMIT → Kilit serbest bırakılır; bekleyen tx devam edebilir.
//
// Bu yaklaşım, "Optimistic Locking" (version/CAS) alternatifine kıyasla
// daha basıl ve çekişme düşükken daha az retry gerektirdiği için tercih edilmiştir.
func (r *walletRepository) UpdateBalance(ctx context.Context, params domain.UpdateBalanceParams) (*domain.Wallet, error) {
	// Miktar doğrulama
	if params.Amount.IsNegative() || params.Amount.IsZero() {
		return nil, domain.ErrInvalidAmount
	}

	// ── ADIM 1: Transaction başlat ────────────────────────────────────────────
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // PostgreSQL varsayılanı; yeterlidir
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("wallet UpdateBalance: transaction başlatılamadı: %w", err)
	}

	// Panic dahil tüm çıkış yollarında rollback güvencesi
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // panik'i yeniden fırlat
		}
	}()

	// rollback helper — hata yollarında kullanılır
	rollback := func(cause error) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback hatası: %v", cause, rbErr)
		}
		return cause
	}

	// ── ADIM 2: SELECT ... FOR UPDATE (Pessimistic Lock) ──────────────────────
	wallet, err := r.getByUserIDForUpdateOnTx(ctx, tx, params.UserID)
	if err != nil {
		return nil, rollback(err)
	}

	// ── ADIM 3: Yeni bakiye hesapla ve doğrula ────────────────────────────────
	var (
		currentBalance decimal.Decimal
		newBalance     decimal.Decimal
	)

	switch params.Field {
	case domain.BalanceFieldUSD:
		currentBalance = wallet.Balance
	case domain.BalanceFieldBTC:
		currentBalance = wallet.BTCBalance
	default:
		return nil, rollback(fmt.Errorf("bilinmeyen bakiye alanı: %s", params.Field))
	}

	switch params.Type {
	case domain.TransferTypeCredit:
		newBalance = currentBalance.Add(params.Amount)

	case domain.TransferTypeDebit:
		if currentBalance.LessThan(params.Amount) {
			return nil, rollback(domain.ErrInsufficientBalance)
		}
		newBalance = currentBalance.Sub(params.Amount)

	default:
		return nil, rollback(fmt.Errorf("bilinmeyen transfer tipi: %s", params.Type))
	}

	// ── ADIM 4: UPDATE — dinamik alan adı güvenli şekilde seçilir ──────────────
	// NOT: SQL injection riski yok — params.Field yalnızca domain sabitleri alır
	//      (BalanceFieldUSD = "balance", BalanceFieldBTC = "btc_balance").
	//      Bu değerler harici kullanıcı girdisi DEĞİLDİR.
	updateQuery := fmt.Sprintf(`
		UPDATE wallets
		SET    %s      = $1,
		       updated_at = $2
		WHERE  user_id  = $3
		RETURNING id, user_id, balance, btc_balance, updated_at
	`, string(params.Field))

	now := time.Now().UTC()
	var updated domain.Wallet
	err = tx.QueryRowxContext(ctx, updateQuery, newBalance, now, params.UserID).StructScan(&updated)
	if err != nil {
		return nil, rollback(fmt.Errorf("wallet UpdateBalance: UPDATE sorgusu başarısız: %w", err))
	}

	// ── ADIM 5: COMMIT — kilit serbest bırakılır ──────────────────────────────
	if err = tx.Commit(); err != nil {
		return nil, rollback(fmt.Errorf("wallet UpdateBalance: commit hatası: %w", err))
	}

	return &updated, nil
}

// =============================================================================
// TransferForOrder — Atomik bakiye takası
// =============================================================================

// TransferForOrder, emir gerçekleştirme sırasında USD ve BTC bakiyesini
// TEK bir veritabanı transaction'ı içinde ve tek seferde günceller.
func (r *walletRepository) TransferForOrder(
	ctx context.Context,
	userID uuid.UUID,
	side string,
	quantity, price decimal.Decimal,
) (*domain.Wallet, error) {
	total := quantity.Mul(price)

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("wallet TransferForOrder: transaction başlatılamadı: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	rollback := func(cause error) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback hatası: %v", cause, rbErr)
		}
		return cause
	}

	wallet, err := r.getByUserIDForUpdateOnTx(ctx, tx, userID)
	if err != nil {
		return nil, rollback(err)
	}

	var newUSD, newBTC decimal.Decimal

	switch side {
	case "BUY":
		if wallet.Balance.LessThan(total) {
			return nil, rollback(domain.ErrInsufficientBalance)
		}
		newUSD = wallet.Balance.Sub(total)
		newBTC = wallet.BTCBalance.Add(quantity)

	case "SELL":
		if wallet.BTCBalance.LessThan(quantity) {
			return nil, rollback(domain.ErrInsufficientBalance)
		}
		newBTC = wallet.BTCBalance.Sub(quantity)
		newUSD = wallet.Balance.Add(total)

	default:
		return nil, rollback(fmt.Errorf("geçersiz emir yönü: %s", side))
	}

	updateQuery := `
		UPDATE wallets
		SET    balance = $1,
		       btc_balance = $2,
		       updated_at = $3
		WHERE  user_id = $4
		RETURNING id, user_id, balance, btc_balance, updated_at
	`

	now := time.Now().UTC()
	var updated domain.Wallet
	err = tx.QueryRowxContext(ctx, updateQuery, newUSD, newBTC, now, userID).StructScan(&updated)
	if err != nil {
		return nil, rollback(fmt.Errorf("wallet TransferForOrder: UPDATE sorgusu başarısız: %w", err))
	}

	if err = tx.Commit(); err != nil {
		return nil, rollback(fmt.Errorf("wallet TransferForOrder: commit hatası: %w", err))
	}

	return &updated, nil
}

// =============================================================================
// Yardımcı Fonksiyonlar
// =============================================================================

// isUniqueViolation, PostgreSQL unique constraint ihlali hatasını (SQLSTATE 23505) kontrol eder.
// lib/pq driverının *pq.Error tipi üzerinden güvenli tip doğrulaması yapılır.
// String matching yerine SQLSTATE kodu kullanılır — yanlış pozitif riski yok.
func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

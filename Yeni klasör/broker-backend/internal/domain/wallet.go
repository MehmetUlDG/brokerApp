package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// User Entity
// =============================================================================

// User, sisteme kayıtlı bir kullanıcıyı temsil eder.
// PasswordHash alanı hiçbir zaman JSON olarak dışarıya sızdırılmamalıdır;
// delivery katmanındaki response DTO'larına bu alan dahil edilmemelidir.
type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"` // JSON'a açılmaz — delivery katmanında atlanır
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	CreatedAt    time.Time `db:"created_at"`
}

// FullName, kullanıcının tam adını döner (Ad + Soyad).
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// =============================================================================
// Wallet Entity
// =============================================================================

// Wallet, bir kullanıcının finansal varlıklarını temsil eder.
// Bakiye alanları için shopspring/decimal kullanılır:
//   - Kayan nokta (float64) hataları yoktur.
//   - DECIMAL(18,8) ↔ decimal.Decimal dönüşümü kayıpsızdır.
//   - Finansal hesaplarda standart yaklaşımdır.
type Wallet struct {
	ID         uuid.UUID       `db:"id"`
	UserID     uuid.UUID       `db:"user_id"`
	Balance    decimal.Decimal `db:"balance"`     // USD bakiyesi
	BTCBalance decimal.Decimal `db:"btc_balance"` // BTC bakiyesi (demo)
	UpdatedAt  time.Time       `db:"updated_at"`
}

// =============================================================================
// Transfer Types
// =============================================================================

// TransferType, bir bakiye değişikliğinin yönünü tanımlar.
type TransferType string

const (
	TransferTypeDebit  TransferType = "DEBIT"  // Bakiyeden düşme (satın alma, çekim)
	TransferTypeCredit TransferType = "CREDIT" // Bakiyeye ekleme (satış, yatırım)
)

// BalanceField, hangi bakiye alanının güncelleneceğini belirtir.
type BalanceField string

const (
	BalanceFieldUSD BalanceField = "balance"      // USD bakiyesi
	BalanceFieldBTC BalanceField = "btc_balance"  // BTC bakiyesi
)

// UpdateBalanceParams, bakiye güncelleme parametrelerini taşır.
// Repository katmanına geçilir; Usecase katmanı bu struct'ı doldurur.
type UpdateBalanceParams struct {
	UserID uuid.UUID
	Field  BalanceField
	Amount decimal.Decimal // Her zaman pozitif olmalıdır
	Type   TransferType    // DEBIT | CREDIT
}

// =============================================================================
// WalletRepository Interface (Infrastructure → Domain bağımlılığı)
// =============================================================================

// WalletRepository, cüzdan veritabanı işlemlerini soyutlar.
// Bağımlılık tersine çevirme (DIP) gereği bu interface domain katmanında
// tanımlanır; implementasyonu infrastructure katmanındadır.
//
// KRİTİK: UpdateBalance implementasyonu mutlaka aşağıdaki garantileri sağlamalıdır:
//  1. SELECT ... FOR UPDATE ile satır-düzeyinde Pessimistic Lock alınmalı.
//  2. Tüm operasyonlar tek bir sqlx.Tx içinde gerçekleştirilmeli.
//  3. Bakiye asla negatife düşmemeli (DEBIT için yeterlilik kontrolü).
type WalletRepository interface {
	// GetByUserID, verilen kullanıcıya ait cüzdanı döner.
	// Cüzdan bulunamazsa ErrWalletNotFound hatası döner.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	// GetByUserIDForUpdate, aynı işi yapar ancak çağıran transaction içinde
	// SELECT ... FOR UPDATE kilidi alır. Yalnızca aktif bir tx içinden çağrılmalıdır.
	GetByUserIDForUpdate(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	// Create, yeni bir cüzdan kaydı oluşturur. Sıfır bakiye ile başlar.
	Create(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	// UpdateBalance, bir kullanıcının belirtilen bakiye alanına
	// atomik bir güncelleme uygular.
	//
	// Implementasyon garantileri:
	//   - Transaction + SELECT FOR UPDATE (Pessimistic Lock)
	//   - DEBIT durumunda yetersiz bakiye → ErrInsufficientBalance
	//   - Başarıda güncel Wallet döner
	UpdateBalance(ctx context.Context, params UpdateBalanceParams) (*Wallet, error)

	// CreateWithTx, aktif bir transaction içinde cüzdan oluşturur.
	// Kullanıcı kaydı ile cüzdan oluşturma aynı tx içinde yapılacaksa kullanılır.
	CreateWithTx(ctx context.Context, tx interface{}, userID uuid.UUID) (*Wallet, error)
}

// =============================================================================
// WalletUsecase Interface (Delivery → Usecase bağımlılığı)
// =============================================================================

// WalletUsecase, cüzdan iş kurallarını soyutlar.
// HTTP handler'lar (delivery katmanı) bu interface'e bağımlıdır;
// concrete implementasyona değil.
type WalletUsecase interface {
	// GetWallet, kimliği doğrulanmış kullanıcının cüzdanını döner.
	GetWallet(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	// Deposit, kullanıcının USD bakiyesine amount ekler.
	// amount > 0 olmalıdır, aksi hâlde ErrInvalidAmount döner.
	Deposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*Wallet, error)

	// Withdraw, kullanıcının USD bakiyesinden amount düşer.
	// Yetersiz bakiye durumunda ErrInsufficientBalance döner.
	Withdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*Wallet, error)

	// TransferForOrder, emir gerçekleştirme sırasında çağrılır.
	// BUY  → USD düşer, BTC artar
	// SELL → BTC düşer, USD artar
	// Tüm işlemler tek bir veritabanı transaction'ı içinde gerçekleşir.
	TransferForOrder(ctx context.Context, userID uuid.UUID, side string, quantity, price decimal.Decimal) (*Wallet, error)
}

// =============================================================================
// Domain Errors
// =============================================================================

// Wallet domain'e özgü hata sentinel değerleri.
// errors.Is() ile karşılaştırılabilirler.
var (
	// ErrWalletNotFound, belirtilen kullanıcıya ait cüzdan bulunamadığında döner.
	ErrWalletNotFound = NewDomainError("wallet_not_found", "cüzdan bulunamadı")

	// ErrInsufficientBalance, DEBIT sırasında bakiye yetersizse döner.
	ErrInsufficientBalance = NewDomainError("insufficient_balance", "yetersiz bakiye")

	// ErrInvalidAmount, amount ≤ 0 ise döner.
	ErrInvalidAmount = NewDomainError("invalid_amount", "tutar sıfırdan büyük olmalıdır")

	// ErrWalletAlreadyExists, aynı kullanıcı için ikinci cüzdan oluşturulmaya
	// çalışıldığında döner (UNIQUE kısıtı ihlali).
	ErrWalletAlreadyExists = NewDomainError("wallet_already_exists", "kullanıcının zaten bir cüzdanı var")
)

// DomainError, domain katmanına özgü yapılandırılmış hata tipidir.
// Code alanı HTTP yanıtlarında ve loglamada kullanılır.
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError, yeni bir DomainError oluşturur.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

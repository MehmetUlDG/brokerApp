package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"payment-service/internal/domain"
)

// =============================================================================
// WalletRepo
// =============================================================================

// WalletRepo, domain.WalletRepository'nin pgx implementasyonudur.
// Broker-backend ile paylaşılan wallets tablosunu doğrudan okur/yazar.
type WalletRepo struct {
	pool *pgxpool.Pool
}

// NewWalletRepo, yeni bir WalletRepo örneği döner.
func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

// GetByUserID, bir kullanıcıya ait cüzdanı döner.
func (r *WalletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	const query = `
		SELECT id, user_id, balance, btc_balance, updated_at
		FROM wallets
		WHERE user_id = $1
	`
	row := r.pool.QueryRow(ctx, query, userID)
	return scanWallet(row)
}

// UpdateBalance, cüzdan bakiyesine usdDelta ve btcDelta uygular.
// SELECT … FOR UPDATE ile pessimistic lock alınır; bakiye negatife düşmesi engellenir.
func (r *WalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, usdDelta, btcDelta decimal.Decimal) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("WalletRepo.UpdateBalance: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Pessimistic lock — aynı anda birden fazla işlemin aynı satırı değiştirmesini engeller.
	const lockQuery = `SELECT id, balance, btc_balance FROM wallets WHERE user_id = $1 FOR UPDATE`
	var (
		walletID   uuid.UUID
		usdBalance decimal.Decimal
		btcBalance decimal.Decimal
		usdStr     string
		btcStr     string
	)
	err = tx.QueryRow(ctx, lockQuery, userID).Scan(&walletID, &usdStr, &btcStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrWalletNotFound
		}
		return fmt.Errorf("WalletRepo.UpdateBalance: lock: %w", err)
	}
	usdBalance, _ = decimal.NewFromString(usdStr)
	btcBalance, _ = decimal.NewFromString(btcStr)

	newUSD := usdBalance.Add(usdDelta)
	newBTC := btcBalance.Add(btcDelta)

	if newUSD.IsNegative() || newBTC.IsNegative() {
		return domain.ErrInsufficientBalance
	}

	const updateQuery = `
		UPDATE wallets
		SET balance = $2, btc_balance = $3, updated_at = NOW()
		WHERE id = $1
	`
	if _, err = tx.Exec(ctx, updateQuery, walletID, newUSD.String(), newBTC.String()); err != nil {
		return fmt.Errorf("WalletRepo.UpdateBalance: update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("WalletRepo.UpdateBalance: commit: %w", err)
	}
	return nil
}

// =============================================================================
// scanWallet — ortak satır tarayıcı
// =============================================================================

func scanWallet(row rowScanner) (*domain.Wallet, error) {
	var (
		w       domain.Wallet
		usdStr  string
		btcStr  string
	)
	err := row.Scan(&w.ID, &w.UserID, &usdStr, &btcStr, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("scanWallet: %w", err)
	}
	w.Balance, _ = decimal.NewFromString(usdStr)
	w.BTCBalance, _ = decimal.NewFromString(btcStr)
	return &w, nil
}

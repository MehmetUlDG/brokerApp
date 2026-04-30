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
		SELECT id, user_id, balance, btc_balance, version, updated_at
		FROM wallets
		WHERE user_id = $1
	`
	row := r.pool.QueryRow(ctx, query, userID)
	return scanWallet(row)
}

// UpdateBalance, cüzdan bakiyesine usdDelta ve btcDelta uygular.
// Optimistic locking kullanır; version kontrolü yapar.
func (r *WalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, usdDelta, btcDelta decimal.Decimal) error {
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.updateBalanceWithOptimisticLock(ctx, userID, usdDelta, btcDelta)
		if err == nil {
			return nil
		}
		if !errors.Is(err, domain.ErrWalletConflict) {
			return err
		}
	}
	return domain.ErrWalletConflict
}

// updateBalanceWithOptimisticLock, tek deneme yapar ve version kontrolü uygular.
func (r *WalletRepo) updateBalanceWithOptimisticLock(ctx context.Context, userID uuid.UUID, usdDelta, btcDelta decimal.Decimal) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("WalletRepo.UpdateBalance: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const lockQuery = `SELECT id, balance, btc_balance, version FROM wallets WHERE user_id = $1 FOR UPDATE`
	var (
		walletID   uuid.UUID
		usdStr     string
		btcStr     string
		version    int
	)
	err = tx.QueryRow(ctx, lockQuery, userID).Scan(&walletID, &usdStr, &btcStr, &version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrWalletNotFound
		}
		return fmt.Errorf("WalletRepo.UpdateBalance: lock: %w", err)
	}

	usdBalance, _ := decimal.NewFromString(usdStr)
	btcBalance, _ := decimal.NewFromString(btcStr)

	newUSD := usdBalance.Add(usdDelta)
	newBTC := btcBalance.Add(btcDelta)

	if newUSD.IsNegative() || newBTC.IsNegative() {
		return domain.ErrInsufficientBalance
	}

	const updateQuery = `
		UPDATE wallets
		SET balance = $2, btc_balance = $3, version = version + 1, updated_at = NOW()
		WHERE id = $1 AND version = $4
	`
	ct, err := tx.Exec(ctx, updateQuery, walletID, newUSD.String(), newBTC.String(), version)
	if err != nil {
		return fmt.Errorf("WalletRepo.UpdateBalance: update: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrWalletConflict
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
	err := row.Scan(&w.ID, &w.UserID, &usdStr, &btcStr, &w.Version, &w.UpdatedAt)
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

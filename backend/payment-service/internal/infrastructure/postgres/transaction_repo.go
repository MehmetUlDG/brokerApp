/*
Schema for transactions table (run once against shared broker_db):

CREATE TABLE transactions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  type        TEXT NOT NULL,
  amount      NUMERIC(20,8) NOT NULL,
  currency    TEXT NOT NULL DEFAULT 'USD',
  status      TEXT NOT NULL DEFAULT 'PENDING',
  stripe_ref  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id   ON transactions(user_id);
CREATE INDEX idx_transactions_stripe_ref ON transactions(stripe_ref) WHERE stripe_ref IS NOT NULL;
*/

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
// TransactionRepo
// =============================================================================

// TransactionRepo, domain.TransactionRepository'nin pgx implementasyonudur.
type TransactionRepo struct {
	pool *pgxpool.Pool
}

// NewTransactionRepo, yeni bir TransactionRepo örneği döner.
func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

// Create, yeni bir Transaction kaydı oluşturur.
func (r *TransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	const query = `
		INSERT INTO transactions (id, user_id, type, amount, currency, status, stripe_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, query,
		tx.ID,
		tx.UserID,
		string(tx.Type),
		tx.Amount.String(),
		tx.Currency,
		string(tx.Status),
		tx.StripeRef,
	)
	if err != nil {
		return fmt.Errorf("TransactionRepo.Create: %w", err)
	}
	return nil
}

// GetByID, ID'ye göre Transaction döner.
func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	const query = `
		SELECT id, user_id, type, amount, currency, status, COALESCE(stripe_ref, ''), created_at, updated_at
		FROM transactions
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanTransaction(row)
}

// ListByUser, kullanıcıya ait işlemleri oluşturulma tarihine göre azalan sırada döner.
func (r *TransactionRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, error) {
	const query = `
		SELECT id, user_id, type, amount, currency, status, COALESCE(stripe_ref, ''), created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.ListByUser: %w", err)
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("TransactionRepo.ListByUser scan: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("TransactionRepo.ListByUser rows: %w", err)
	}
	return txs, nil
}

// UpdateStatus, bir Transaction'ın status, stripe_ref ve updated_at alanlarını günceller.
func (r *TransactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus, stripeRef string) error {
	const query = `
		UPDATE transactions
		SET status = $2, stripe_ref = $3, updated_at = NOW()
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, query, id, string(status), stripeRef)
	if err != nil {
		return fmt.Errorf("TransactionRepo.UpdateStatus: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrTransactionNotFound
	}
	return nil
}

// GetByStripeRef, StripeRef'e göre Transaction döner (webhook lookup).
func (r *TransactionRepo) GetByStripeRef(ctx context.Context, stripeRef string) (*domain.Transaction, error) {
	const query = `
		SELECT id, user_id, type, amount, currency, status, COALESCE(stripe_ref, ''), created_at, updated_at
		FROM transactions
		WHERE stripe_ref = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, stripeRef)
	return scanTransaction(row)
}

// =============================================================================
// scanTransaction — ortak satır tarayıcı
// =============================================================================

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(row rowScanner) (*domain.Transaction, error) {
	var (
		t         domain.Transaction
		amountStr string
	)
	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.Type,
		&amountStr,
		&t.Currency,
		&t.Status,
		&t.StripeRef,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("scanTransaction: %w", err)
	}
	t.Amount, err = decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("scanTransaction: bad amount %q: %w", amountStr, err)
	}
	return &t, nil
}

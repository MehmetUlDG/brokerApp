package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// TransactionType
// =============================================================================

// TransactionType, bir işlemin türünü belirtir.
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeRefund     TransactionType = "REFUND"
)

// =============================================================================
// TransactionStatus
// =============================================================================

// TransactionStatus, bir işlemin durumunu belirtir.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

// =============================================================================
// Transaction Entity
// =============================================================================

// Transaction, bir para giriş/çıkış işlemini temsil eder.
// Amount alanı her zaman decimal.Decimal olarak taşınır; string'e yalnızca
// proto sınırında dönüştürülür.
type Transaction struct {
	ID        uuid.UUID         `db:"id"`
	UserID    uuid.UUID         `db:"user_id"`
	Type      TransactionType   `db:"type"`
	Amount    decimal.Decimal   `db:"amount"`
	Currency  string            `db:"currency"`
	Status    TransactionStatus `db:"status"`
	StripeRef string            `db:"stripe_ref"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

// =============================================================================
// TransactionRepository Interface
// =============================================================================

// TransactionRepository, işlem veritabanı operasyonlarını soyutlar.
// Implementasyonu infrastructure/postgres katmanındadır.
type TransactionRepository interface {
	// Create, yeni bir Transaction kaydı oluşturur.
	Create(ctx context.Context, tx *Transaction) error

	// GetByID, ID'ye göre Transaction döner.
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)

	// ListByUser, kullanıcıya ait işlemleri oluşturulma tarihine göre azalan
	// sırada, sayfalayarak döner.
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Transaction, error)

	// UpdateStatus, bir Transaction'ın status, stripe_ref ve updated_at
	// alanlarını günceller.
	UpdateStatus(ctx context.Context, id uuid.UUID, status TransactionStatus, stripeRef string) error

	// GetByStripeRef, StripeRef'e göre Transaction döner (webhook lookup için).
	GetByStripeRef(ctx context.Context, stripeRef string) (*Transaction, error)
}

// =============================================================================
// Domain Errors
// =============================================================================

var (
	ErrTransactionNotFound = NewPaymentError("transaction_not_found", "transaction not found")
)

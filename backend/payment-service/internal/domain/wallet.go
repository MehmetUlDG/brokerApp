package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// Wallet Entity
// =============================================================================

// Wallet, bir kullanıcının finansal varlıklarını temsil eder.
// Bakiye alanları decimal.Decimal ile taşınır; float64 kesinlik hataları yoktur.
type Wallet struct {
	ID         uuid.UUID       `db:"id"`
	UserID     uuid.UUID       `db:"user_id"`
	Balance    decimal.Decimal `db:"balance"`     // USD bakiyesi
	BTCBalance decimal.Decimal `db:"btc_balance"` // BTC bakiyesi
	Version    int             `db:"version"`     // Optimistic locking versiyonu
	UpdatedAt  time.Time       `db:"updated_at"`
}

// =============================================================================
// WalletRepository Interface
// =============================================================================

// WalletRepository, cüzdan okuma/yazma operasyonlarını soyutlar.
// Broker-backend ile aynı wallets tablosunu paylaşır.
type WalletRepository interface {
	// GetByUserID, kullanıcıya ait cüzdanı döner.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Wallet, error)

	// UpdateBalance, usdDelta ve btcDelta kadar bakiye değişikliği uygular.
	// Pozitif delta → credit, negatif delta → debit.
	// Bakiye asla negatife düşmemeli; implementasyon bu kontrolü yapmalıdır.
	UpdateBalance(ctx context.Context, userID uuid.UUID, usdDelta, btcDelta decimal.Decimal) error
}

// =============================================================================
// Domain Errors
// =============================================================================

var (
	ErrWalletNotFound      = NewPaymentError("wallet_not_found", "wallet not found")
	ErrInsufficientBalance = NewPaymentError("insufficient_balance", "insufficient balance")
	ErrInvalidAmount       = NewPaymentError("invalid_amount", "amount must be greater than zero")
	ErrWalletConflict      = NewPaymentError("wallet_conflict", "wallet version conflict, please retry")
)

// =============================================================================
// PaymentError — domain hata tipi
// =============================================================================

// PaymentError, payment-service domain'ine özgü yapılandırılmış hata tipidir.
type PaymentError struct {
	Code    string
	Message string
}

func (e *PaymentError) Error() string { return e.Message }

// NewPaymentError, yeni bir PaymentError oluşturur.
func NewPaymentError(code, message string) *PaymentError {
	return &PaymentError{Code: code, Message: message}
}

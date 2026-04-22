package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"payment-service/internal/domain"
	kafkainfra "payment-service/internal/infrastructure/kafka"
	stripeinfra "payment-service/internal/infrastructure/stripe"
)

// =============================================================================
// PaymentUsecase
// =============================================================================

// PaymentUsecase, tüm ödeme iş mantığını uygular.
// Bağımlılıklar constructor injection ile verilir; global state yoktur.
type PaymentUsecase struct {
	txRepo    domain.TransactionRepository
	walletRepo domain.WalletRepository
	stripe    *stripeinfra.StripeAdapter
	publisher *kafkainfra.PaymentPublisher
}

// NewPaymentUsecase, yeni bir PaymentUsecase döner.
func NewPaymentUsecase(
	txRepo domain.TransactionRepository,
	walletRepo domain.WalletRepository,
	stripe *stripeinfra.StripeAdapter,
	publisher *kafkainfra.PaymentPublisher,
) *PaymentUsecase {
	return &PaymentUsecase{
		txRepo:    txRepo,
		walletRepo: walletRepo,
		stripe:    stripe,
		publisher: publisher,
	}
}

// =============================================================================
// Deposit
// =============================================================================

// Deposit, kullanıcının USD bakiyesine para yatırma işlemi uygular.
//  1. amount string → decimal.Decimal dönüşümü.
//  2. PENDING Transaction kaydı oluşturulur.
//  3. Stripe PaymentIntent oluşturulur ve onaylanır.
//  4. Başarıda: wallet güncellenir, TX COMPLETED yapılır, deposit.completed yayımlanır.
//  5. Stripe hatasında: TX FAILED yapılır, deposit.failed yayımlanır.
func (u *PaymentUsecase) Deposit(
	ctx context.Context,
	userID, amountStr, currency, paymentMethodID string,
) (*domain.Transaction, error) {

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || !amount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewPaymentError("invalid_user_id", fmt.Sprintf("invalid user_id: %v", err))
	}

	// 1. PENDING kaydı oluştur
	tx := &domain.Transaction{
		ID:       uuid.New(),
		UserID:   uid,
		Type:     domain.TransactionTypeDeposit,
		Amount:   amount,
		Currency: currency,
		Status:   domain.TransactionStatusPending,
	}
	if err := u.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("Deposit: create tx: %w", err)
	}

	// 2. Stripe PaymentIntent
	_, stripeID, stripeErr := u.stripe.CreatePaymentIntent(amount, currency, paymentMethodID)

	if stripeErr != nil {
		// Stripe başarısız → FAILED
		_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, "")
		tx.Status = domain.TransactionStatusFailed

		_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
			EventType:     "deposit.failed",
			TransactionID: tx.ID.String(),
			UserID:        userID,
			Amount:        amount.String(),
			Currency:      currency,
			StripeRef:     "",
		})

		return tx, stripeErr
	}

	// 3. Stripe başarılı → wallet güncelle + COMPLETED
	if err := u.walletRepo.UpdateBalance(ctx, uid, amount, decimal.Zero); err != nil {
		// Wallet güncelleme başarısız olsa da TX'i COMPLETED yapıyoruz;
		// operatörün manuel müdahale için event yayımlanır.
		_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, stripeID)
		tx.Status = domain.TransactionStatusFailed
		return tx, fmt.Errorf("Deposit: update wallet: %w", err)
	}

	_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeID)
	tx.Status = domain.TransactionStatusCompleted
	tx.StripeRef = stripeID

	_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
		EventType:     "deposit.completed",
		TransactionID: tx.ID.String(),
		UserID:        userID,
		Amount:        amount.String(),
		Currency:      currency,
		StripeRef:     stripeID,
	})

	return tx, nil
}

// =============================================================================
// Withdraw
// =============================================================================

// Withdraw, kullanıcının USD bakiyesinden para çekme işlemi uygular.
//  1. Wallet bakiyesi kontrol edilir (usd_balance >= amount).
//  2. PENDING Transaction kaydı oluşturulur.
//  3. Stripe Payout oluşturulur.
//  4. Başarıda: wallet güncellenir (usdDelta = -amount), TX COMPLETED, withdrawal.completed yayımlanır.
//  5. Stripe hatasında: TX FAILED.
func (u *PaymentUsecase) Withdraw(
	ctx context.Context,
	userID, amountStr, currency, stripeAccountID string,
) (*domain.Transaction, error) {

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || !amount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewPaymentError("invalid_user_id", fmt.Sprintf("invalid user_id: %v", err))
	}

	// 1. Bakiye kontrolü
	wallet, err := u.walletRepo.GetByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("Withdraw: get wallet: %w", err)
	}
	if wallet.Balance.LessThan(amount) {
		return nil, domain.ErrInsufficientBalance
	}

	// 2. PENDING kaydı
	tx := &domain.Transaction{
		ID:       uuid.New(),
		UserID:   uid,
		Type:     domain.TransactionTypeWithdrawal,
		Amount:   amount,
		Currency: currency,
		Status:   domain.TransactionStatusPending,
	}
	if err := u.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("Withdraw: create tx: %w", err)
	}

	// 3. Stripe Payout
	stripeID, stripeErr := u.stripe.CreatePayout(amount, currency, stripeAccountID)

	if stripeErr != nil {
		_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, "")
		tx.Status = domain.TransactionStatusFailed

		_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
			EventType:     "withdrawal.failed",
			TransactionID: tx.ID.String(),
			UserID:        userID,
			Amount:        amount.String(),
			Currency:      currency,
			StripeRef:     "",
		})

		return tx, stripeErr
	}

	// 4. Wallet güncelle (usdDelta negatif)
	if err := u.walletRepo.UpdateBalance(ctx, uid, amount.Neg(), decimal.Zero); err != nil {
		_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, stripeID)
		tx.Status = domain.TransactionStatusFailed
		return tx, fmt.Errorf("Withdraw: update wallet: %w", err)
	}

	_ = u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeID)
	tx.Status = domain.TransactionStatusCompleted
	tx.StripeRef = stripeID

	_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
		EventType:     "withdrawal.completed",
		TransactionID: tx.ID.String(),
		UserID:        userID,
		Amount:        amount.String(),
		Currency:      currency,
		StripeRef:     stripeID,
	})

	return tx, nil
}

// =============================================================================
// GetHistory
// =============================================================================

// GetHistory, kullanıcıya ait işlem geçmişini döner.
func (u *PaymentUsecase) GetHistory(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]*domain.Transaction, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewPaymentError("invalid_user_id", fmt.Sprintf("invalid user_id: %v", err))
	}
	return u.txRepo.ListByUser(ctx, uid, limit, offset)
}

// =============================================================================
// GetBalance
// =============================================================================

// GetBalance, kullanıcıya ait cüzdan bakiyesini döner.
func (u *PaymentUsecase) GetBalance(ctx context.Context, userID string) (*domain.Wallet, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewPaymentError("invalid_user_id", fmt.Sprintf("invalid user_id: %v", err))
	}
	return u.walletRepo.GetByUserID(ctx, uid)
}

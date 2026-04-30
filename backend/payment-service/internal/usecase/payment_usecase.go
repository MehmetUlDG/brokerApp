package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"payment-service/internal/domain"
	kafkainfra "payment-service/internal/infrastructure/kafka"
)

const defaultTimeout = 30 * time.Second

// StripeAdapterInterface, Stripe API çağrılarını soyutlar.
type StripeAdapterInterface interface {
	CreatePaymentIntent(amount decimal.Decimal, currency, paymentMethodID string) (string, string, error)
	CreatePaymentIntentWithContext(ctx context.Context, amount decimal.Decimal, currency, paymentMethodID string) (string, string, error)
	CreatePayout(amount decimal.Decimal, currency, stripeAccountID string) (string, error)
	CreatePayoutWithContext(ctx context.Context, amount decimal.Decimal, currency, stripeAccountID string) (string, error)
	RefundPayment(paymentIntentID string) error
}

// PaymentPublisherInterface, Kafka yayıncısını soyutlar.
type PaymentPublisherInterface interface {
	Publish(ctx context.Context, msg kafkainfra.PaymentEventMsg) error
	Close() error
}

// PaymentUsecase, tüm ödeme iş mantığını uygular.
// Bağımlılıklar constructor injection ile verilir; global state yoktur.
type PaymentUsecase struct {
	txRepo     domain.TransactionRepository
	walletRepo domain.WalletRepository
	stripe     StripeAdapterInterface
	publisher  PaymentPublisherInterface
	logger     *zap.Logger
}

// NewPaymentUsecase, yeni bir PaymentUsecase döner.
func NewPaymentUsecase(
	txRepo domain.TransactionRepository,
	walletRepo domain.WalletRepository,
	stripe StripeAdapterInterface,
	publisher PaymentPublisherInterface,
	logger *zap.Logger,
) *PaymentUsecase {
	return &PaymentUsecase{
		txRepo:     txRepo,
		walletRepo: walletRepo,
		stripe:     stripe,
		publisher:  publisher,
		logger:     logger,
	}
}

// Deposit, kullanıcının USD bakiyesine para yatırma işlemi uygular.
//  1. amount string → decimal.Decimal dönüşümü.
//  2. PENDING Transaction kaydı oluşturulur.
//  3. Stripe PaymentIntent oluşturulur ve onaylanır.
//  4. Başarıda: wallet güncellenir, TX COMPLETED yapılır, deposit.completed yayımlanır.
//  5. Stripe başarılı/Wallet başarısız: Refund (Compensating Transaction) + TX FAILED.
//  6. Stripe hatasında: TX FAILED yapılır, deposit.failed yayımlanır.
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

	// 2. Stripe PaymentIntent (30sn timeout)
	stripeCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	_, stripeID, stripeErr := u.stripe.CreatePaymentIntentWithContext(stripeCtx, amount, currency, paymentMethodID)

	if stripeErr != nil {
		// Stripe başarısız → FAILED
		if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, ""); err != nil {
			u.logger.Error("Deposit: UpdateStatus failed after Stripe error",
				zap.String("tx_id", tx.ID.String()),
				zap.Error(err))
		}
		tx.Status = domain.TransactionStatusFailed

		_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
			EventType:     "deposit.failed",
			TransactionID: tx.ID.String(),
			UserID:        userID,
			Amount:        amount.String(),
			Currency:      currency,
			StripeRef:     "",
		})

		u.logger.Warn("Deposit failed: Stripe error",
			zap.String("tx_id", tx.ID.String()),
			zap.String("user_id", userID),
			zap.Error(stripeErr))

		return tx, stripeErr
	}

	// 3. Stripe başarılı → wallet güncelle (30sn timeout)
	walletCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	if err := u.walletRepo.UpdateBalance(walletCtx, uid, amount, decimal.Zero); err != nil {
		// Wallet başarısız → Refund (Compensating Transaction)
		u.logger.Error("Deposit: Wallet update failed, initiating refund",
			zap.String("tx_id", tx.ID.String()),
			zap.String("stripe_id", stripeID),
			zap.Error(err))

		refundErr := u.processRefund(ctx, tx.ID, stripeID)
		if refundErr != nil {
			u.logger.Error("Deposit: Refund failed, manual intervention required",
				zap.String("tx_id", tx.ID.String()),
				zap.String("stripe_id", stripeID),
				zap.Error(refundErr))
		}

		if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, stripeID); err != nil {
			u.logger.Error("Deposit: UpdateStatus failed after refund",
				zap.String("tx_id", tx.ID.String()),
				zap.Error(err))
		}
		tx.Status = domain.TransactionStatusFailed
		tx.StripeRef = stripeID

		_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
			EventType:     "deposit.failed",
			TransactionID: tx.ID.String(),
			UserID:        userID,
			Amount:        amount.String(),
			Currency:      currency,
			StripeRef:     stripeID,
		})

		return tx, fmt.Errorf("Deposit: update wallet: %w", err)
	}

	// 4. Wallet başarılı → COMPLETED
	if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeID); err != nil {
		u.logger.Error("Deposit: UpdateStatus failed after wallet success",
			zap.String("tx_id", tx.ID.String()),
			zap.Error(err))
	}
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

	u.logger.Info("Deposit completed successfully",
		zap.String("tx_id", tx.ID.String()),
		zap.String("user_id", userID),
		zap.String("amount", amount.String()))

	return tx, nil
}

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
	walletCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	wallet, err := u.walletRepo.GetByUserID(walletCtx, uid)
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

	// 3. Stripe Payout (30sn timeout)
	payoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	stripeID, stripeErr := u.stripe.CreatePayoutWithContext(payoutCtx, amount, currency, stripeAccountID)

	if stripeErr != nil {
		if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, ""); err != nil {
			u.logger.Error("Withdraw: UpdateStatus failed after Stripe error",
				zap.String("tx_id", tx.ID.String()),
				zap.Error(err))
		}
		tx.Status = domain.TransactionStatusFailed

		_ = u.publisher.Publish(ctx, kafkainfra.PaymentEventMsg{
			EventType:     "withdrawal.failed",
			TransactionID: tx.ID.String(),
			UserID:        userID,
			Amount:        amount.String(),
			Currency:      currency,
			StripeRef:     "",
		})

		u.logger.Warn("Withdraw failed: Stripe error",
			zap.String("tx_id", tx.ID.String()),
			zap.String("user_id", userID),
			zap.Error(stripeErr))

		return tx, stripeErr
	}

	// 4. Wallet güncelle (30sn timeout)
	walletCtx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	if err := u.walletRepo.UpdateBalance(walletCtx, uid, amount.Neg(), decimal.Zero); err != nil {
		if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed, stripeID); err != nil {
			u.logger.Error("Withdraw: UpdateStatus failed after wallet error",
				zap.String("tx_id", tx.ID.String()),
				zap.Error(err))
		}
		tx.Status = domain.TransactionStatusFailed
		return tx, fmt.Errorf("Withdraw: update wallet: %w", err)
	}

	// 5. COMPLETED
	if err := u.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeID); err != nil {
		u.logger.Error("Withdraw: UpdateStatus failed after wallet success",
			zap.String("tx_id", tx.ID.String()),
			zap.Error(err))
	}
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

	u.logger.Info("Withdraw completed successfully",
		zap.String("tx_id", tx.ID.String()),
		zap.String("user_id", userID),
		zap.String("amount", amount.String()))

	return tx, nil
}

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

// GetBalance, kullanıcıya ait cüzdan bakiyesini döner.
func (u *PaymentUsecase) GetBalance(ctx context.Context, userID string) (*domain.Wallet, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewPaymentError("invalid_user_id", fmt.Sprintf("invalid user_id: %v", err))
	}
	return u.walletRepo.GetByUserID(ctx, uid)
}

// processRefund, Stripe'dan refund (iade) işlemi yapar ve refund transaction kaydı oluşturur.
func (u *PaymentUsecase) processRefund(ctx context.Context, txID uuid.UUID, stripeID string) error {
	refundCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if err := u.stripe.RefundPayment(stripeID); err != nil {
		return fmt.Errorf("processRefund: stripe refund: %w", err)
	}

	refundTx := &domain.Transaction{
		ID:        uuid.New(),
		UserID:    txID,
		Type:      domain.TransactionTypeRefund,
		Amount:    decimal.Zero,
		Currency:  "USD",
		Status:    domain.TransactionStatusCompleted,
		StripeRef: stripeID,
	}

	if err := u.txRepo.Create(refundCtx, refundTx); err != nil {
		return fmt.Errorf("processRefund: create refund tx: %w", err)
	}

	u.logger.Info("Refund processed successfully",
		zap.String("original_tx_id", txID.String()),
		zap.String("refund_tx_id", refundTx.ID.String()),
		zap.String("stripe_id", stripeID))

	return nil
}

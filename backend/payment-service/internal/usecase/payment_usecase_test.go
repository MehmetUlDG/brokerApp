package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"payment-service/internal/domain"
	"payment-service/internal/infrastructure/kafka"
	"payment-service/internal/usecase"
)

type testTransactionRepo struct {
	createdTX      *domain.Transaction
	updatedStatus  domain.TransactionStatus
	updatedStripe  string
	getByStripeRef *domain.Transaction
}

func (r *testTransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	r.createdTX = tx
	return nil
}

func (r *testTransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	return nil, nil
}

func (r *testTransactionRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, error) {
	return nil, nil
}

func (r *testTransactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus, stripeRef string) error {
	r.updatedStatus = status
	r.updatedStripe = stripeRef
	return nil
}

func (r *testTransactionRepo) GetByStripeRef(ctx context.Context, stripeRef string) (*domain.Transaction, error) {
	return r.getByStripeRef, nil
}

type testWalletRepo struct {
	wallet       *domain.Wallet
	updateErr    error
	getErr       error
	lastUsdDelta decimal.Decimal
}

func (r *testWalletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.wallet, nil
}

func (r *testWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, usdDelta, btcDelta decimal.Decimal) error {
	r.lastUsdDelta = usdDelta
	return r.updateErr
}

type testStripeAdapter struct {
	createPIErr     error
	createPIWithCtxErr error
	createPOErr     error
	createPOWithCtxErr error
	refundErr       error
	stripeID        string
	clientSecret   string
}

func (a *testStripeAdapter) CreatePaymentIntent(amount decimal.Decimal, currency, paymentMethodID string) (string, string, error) {
	return a.clientSecret, a.stripeID, a.createPIErr
}

func (a *testStripeAdapter) CreatePaymentIntentWithContext(ctx context.Context, amount decimal.Decimal, currency, paymentMethodID string) (string, string, error) {
	return a.clientSecret, a.stripeID, a.createPIWithCtxErr
}

func (a *testStripeAdapter) CreatePayout(amount decimal.Decimal, currency, stripeAccountID string) (string, error) {
	return a.stripeID, a.createPOErr
}

func (a *testStripeAdapter) CreatePayoutWithContext(ctx context.Context, amount decimal.Decimal, currency, stripeAccountID string) (string, error) {
	return a.stripeID, a.createPOWithCtxErr
}

func (a *testStripeAdapter) RefundPayment(paymentIntentID string) error {
	return a.refundErr
}

type testPaymentPublisher struct {
	published bool
	err       error
}

func (p *testPaymentPublisher) Publish(ctx context.Context, msg kafka.PaymentEventMsg) error {
	p.published = true
	return p.err
}

func (p *testPaymentPublisher) Close() error {
	return nil
}

func TestPaymentUsecase_Deposit_StripeSuccess_WalletSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{}
	stripeAdapter := &testStripeAdapter{
		stripeID:     "pi_test123",
		clientSecret: "client_secret",
	}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	userID := uuid.New().String()
	amountStr := "100.00"
	currency := "USD"
	paymentMethodID := "pm_test123"

	ctx := context.Background()
	result, err := uc.Deposit(ctx, userID, amountStr, currency, paymentMethodID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TransactionStatusCompleted, result.Status)
	assert.Equal(t, "pi_test123", result.StripeRef)
	assert.True(t, publisher.published)
	assert.Equal(t, domain.TransactionStatusCompleted, txRepo.updatedStatus)
}

func TestPaymentUsecase_Deposit_StripeSuccess_WalletFail_Refund(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{
		updateErr: assert.AnError,
	}
	stripeAdapter := &testStripeAdapter{
		stripeID:     "pi_test123",
		clientSecret: "client_secret",
	}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	userID := uuid.New().String()
	amountStr := "100.00"
	currency := "USD"
	paymentMethodID := "pm_test123"

	ctx := context.Background()
	result, err := uc.Deposit(ctx, userID, amountStr, currency, paymentMethodID)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TransactionStatusFailed, result.Status)
	assert.Equal(t, "pi_test123", result.StripeRef)
	assert.True(t, publisher.published)
}

func TestPaymentUsecase_Deposit_StripeFail(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{}
	stripeAdapter := &testStripeAdapter{
		createPIWithCtxErr: domain.NewPaymentError("stripe_card_declined", "card declined"),
	}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	userID := uuid.New().String()
	amountStr := "100.00"
	currency := "USD"
	paymentMethodID := "pm_test123"

	ctx := context.Background()
	result, err := uc.Deposit(ctx, userID, amountStr, currency, paymentMethodID)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TransactionStatusFailed, result.Status)
	assert.True(t, publisher.published)
	assert.Equal(t, domain.TransactionStatusFailed, txRepo.updatedStatus)
}

func TestPaymentUsecase_Withdraw_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	uid := uuid.New()
	balance, _ := decimal.NewFromString("100.00")

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{
		wallet: &domain.Wallet{
			ID:      uid,
			UserID:  uid,
			Balance: balance,
		},
	}
	stripeAdapter := &testStripeAdapter{
		stripeID: "po_test123",
	}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	userID := uid.String()
	amountStr := "50.00"
	currency := "USD"
	stripeAccountID := "acct_test123"

	ctx := context.Background()
	result, err := uc.Withdraw(ctx, userID, amountStr, currency, stripeAccountID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TransactionStatusCompleted, result.Status)
	assert.True(t, publisher.published)
	assert.Equal(t, domain.TransactionStatusCompleted, txRepo.updatedStatus)
}

func TestPaymentUsecase_Withdraw_InsufficientBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	uid := uuid.New()
	balance, _ := decimal.NewFromString("50.00")

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{
		wallet: &domain.Wallet{
			ID:      uid,
			UserID:  uid,
			Balance: balance,
		},
	}
	stripeAdapter := &testStripeAdapter{}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	userID := uid.String()
	amountStr := "100.00"
	currency := "USD"
	stripeAccountID := "acct_test123"

	ctx := context.Background()
	result, err := uc.Withdraw(ctx, userID, amountStr, currency, stripeAccountID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
}

func TestPaymentUsecase_Deposit_InvalidAmount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{}
	stripeAdapter := &testStripeAdapter{}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	ctx := context.Background()
	result, err := uc.Deposit(ctx, uuid.New().String(), "invalid", "USD", "pm_test")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestPaymentUsecase_Deposit_InvalidUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{}
	stripeAdapter := &testStripeAdapter{}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	ctx := context.Background()
	result, err := uc.Deposit(ctx, "not-a-uuid", "100.00", "USD", "pm_test")

	assert.Error(t, err)
	assert.Nil(t, result)
	var pe *domain.PaymentError
	assert.ErrorAs(t, err, &pe)
	assert.Equal(t, "invalid_user_id", pe.Code)
}

func TestPaymentUsecase_GetBalance_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	uid := uuid.New()
	balance, _ := decimal.NewFromString("250.00")

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{
		wallet: &domain.Wallet{
			ID:      uid,
			UserID:  uid,
			Balance: balance,
		},
	}
	stripeAdapter := &testStripeAdapter{}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	ctx := context.Background()
	wallet, err := uc.GetBalance(ctx, uid.String())

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.True(t, wallet.Balance.Equal(balance))
}

func TestPaymentUsecase_ContextTimeout(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	txRepo := &testTransactionRepo{}
	walletRepo := &testWalletRepo{
		updateErr: context.DeadlineExceeded,
	}
	stripeAdapter := &testStripeAdapter{
		stripeID:     "pi_test",
		clientSecret: "secret",
		refundErr:    nil,
	}
	publisher := &testPaymentPublisher{}

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	ctx := context.Background()
	result, err := uc.Deposit(ctx, uuid.New().String(), "100.00", "USD", "pm_test")

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TransactionStatusFailed, result.Status)
}
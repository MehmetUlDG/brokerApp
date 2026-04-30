package stripe

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/payout"
	"github.com/stripe/stripe-go/v76/refund"

	"payment-service/internal/domain"
)

// =============================================================================
// StripeAdapter
// =============================================================================

// StripeAdapter, Stripe API çağrılarını soyutlar.
// Global stripe.Key yerine her istekte secretKey açıkça set edilir.
type StripeAdapter struct {
	secretKey string
}

// NewStripeAdapter, yeni bir StripeAdapter döner.
func NewStripeAdapter(secretKey string) *StripeAdapter {
	return &StripeAdapter{secretKey: secretKey}
}

// CreatePaymentIntent, Stripe PaymentIntent oluşturur ve onaylar.
// amount: ondalıklı string olarak USD miktarı (örn. "100.50").
// Stripe API'si kuruş cinsinden int64 bekler; dönüşüm burada yapılır.
// Returns: clientSecret, stripeID (PaymentIntent.ID), err.
func (a *StripeAdapter) CreatePaymentIntent(
	amount decimal.Decimal,
	currency, paymentMethodID string,
) (clientSecret, stripeID string, err error) {
	return a.CreatePaymentIntentWithContext(context.Background(), amount, currency, paymentMethodID)
}

// CreatePaymentIntentWithContext, Stripe PaymentIntent oluşturur ve onaylar (context ile).
func (a *StripeAdapter) CreatePaymentIntentWithContext(
	ctx context.Context,
	amount decimal.Decimal,
	currency, paymentMethodID string,
) (clientSecret, stripeID string, err error) {
	stripe.Key = a.secretKey

	// Stripe kuruş (cents) cinsinden çalışır: 1 USD = 100 cents.
	cents := amount.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(cents),
		Currency:      stripe.String(currency),
		PaymentMethod: stripe.String(paymentMethodID),
		Confirm:       stripe.Bool(true),
		// Off-session = kullanıcı aktif oturum yokken ödeme onayı
		OffSession: stripe.Bool(true),
		Params: stripe.Params{
			Context: ctx,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", "", wrapStripeError("CreatePaymentIntent", err)
	}

	return pi.ClientSecret, pi.ID, nil
}

// CreatePayout, Stripe Payout (çekim) oluşturur.
// amount: ondalıklı string olarak USD miktarı.
// stripeAccountID: hedef connected account veya bank account ID.
// Returns: stripeID (Payout.ID), err.
func (a *StripeAdapter) CreatePayout(
	amount decimal.Decimal,
	currency, stripeAccountID string,
) (stripeID string, err error) {
	return a.CreatePayoutWithContext(context.Background(), amount, currency, stripeAccountID)
}

// CreatePayoutWithContext, Stripe Payout (çekim) oluşturur (context ile).
func (a *StripeAdapter) CreatePayoutWithContext(
	ctx context.Context,
	amount decimal.Decimal,
	currency, stripeAccountID string,
) (stripeID string, err error) {
	stripe.Key = a.secretKey

	cents := amount.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.PayoutParams{
		Amount:      stripe.Int64(cents),
		Currency:    stripe.String(currency),
		Destination: stripe.String(stripeAccountID),
		Params: stripe.Params{
			Context: ctx,
		},
	}

	po, err := payout.New(params)
	if err != nil {
		return "", wrapStripeError("CreatePayout", err)
	}

	return po.ID, nil
}

// RefundPayment, Stripe PaymentIntent için tam refund (iade) oluşturur.
// paymentIntentID: Stripe PaymentIntent ID (pi_...).
func (a *StripeAdapter) RefundPayment(paymentIntentID string) error {
	stripe.Key = a.secretKey

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	_, err := refund.New(params)
	if err != nil {
		return wrapStripeError("RefundPayment", err)
	}
	return nil
}

// =============================================================================
// Error handling
// =============================================================================

// wrapStripeError, Stripe SDK hatalarını domain-friendly PaymentError'a dönüştürür.
func wrapStripeError(op string, err error) error {
	if stripeErr, ok := err.(*stripe.Error); ok {
		return domain.NewPaymentError(
			fmt.Sprintf("stripe_%s", stripeErr.Code),
			fmt.Sprintf("%s: stripe error [%s]: %s", op, stripeErr.Code, stripeErr.Msg),
		)
	}
	return domain.NewPaymentError(
		"stripe_unknown",
		fmt.Sprintf("%s: %v", op, err),
	)
}

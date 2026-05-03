package handler

import (
	"encoding/json"
	"net/http"

	"payment-service/gen/payment"

	"github.com/yourusername/broker-backend/internal/delivery/http/middleware"
	"github.com/yourusername/broker-backend/internal/domain"
)

// WalletHandler, cüzdan endpoint'lerini yönetir.
// Tüm route'lar JWT middleware ile korunur.
type WalletHandler struct {
	usecase       domain.WalletUsecase
	paymentClient payment.PaymentServiceClient
}

// NewWalletHandler, yeni bir WalletHandler örneği döner.
func NewWalletHandler(usecase domain.WalletUsecase, paymentClient payment.PaymentServiceClient) *WalletHandler {
	return &WalletHandler{usecase: usecase, paymentClient: paymentClient}
}

// depositRequest, Stripe ile para yatırma için istek yapısıdır.
type depositRequest struct {
	Amount          string `json:"amount"`
	PaymentMethodID string `json:"payment_method_id"`
}

// withdrawRequest, para çekme için istek yapısıdır.
type withdrawRequest struct {
	Amount          string `json:"amount"`
	StripeAccountID string `json:"stripe_account_id"`
}

// GetWallet godoc
//
//	GET /api/wallet
//	Yanıt: 200 OK → { wallet }
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	wallet, err := h.usecase.GetWallet(r.Context(), userID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, wallet)
}

// Deposit godoc
//
//	POST /api/wallet/deposit
//	Body: { "amount": "100.50", "payment_method_id": "pm_..." }
//	Yanıt: 200 OK → { "transaction_id": "...", "status": "..." }
func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	if req.Amount == "" || req.PaymentMethodID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "amount ve payment_method_id zorunludur",
		})
		return
	}

	// gRPC çağrısı
	resp, err := h.paymentClient.Deposit(r.Context(), &payment.DepositRequest{
		UserId:                userID.String(),
		Amount:                req.Amount,
		Currency:              "USD",
		StripePaymentMethodId: req.PaymentMethodID,
	})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Withdraw godoc
//
//	POST /api/wallet/withdraw
//	Body: { "amount": "50.00", "stripe_account_id": "acct_..." }
//	Yanıt: 200 OK → { "transaction_id": "...", "status": "..." }
func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	if req.Amount == "" || req.StripeAccountID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "amount ve stripe_account_id zorunludur",
		})
		return
	}

	// gRPC çağrısı
	resp, err := h.paymentClient.Withdraw(r.Context(), &payment.WithdrawRequest{
		UserId:          userID.String(),
		Amount:          req.Amount,
		Currency:        "USD",
		StripeAccountId: req.StripeAccountID,
	})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetTransactions godoc
//
//	GET /api/transactions
//	Yanıt: 200 OK → { transactions }
func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	// gRPC çağrısı
	resp, err := h.paymentClient.GetHistory(r.Context(), &payment.HistoryRequest{
		UserId: userID.String(),
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	transactions := resp.GetTransactions()
	if transactions == nil {
		transactions = []*payment.Transaction{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
	})
}

// GetBalance godoc
//
//	GET /api/balance
//	Yanıt: 200 OK → { usd_balance, btc_balance }
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	// gRPC çağrısı
	resp, err := h.paymentClient.GetBalance(r.Context(), &payment.BalanceRequest{
		UserId: userID.String(),
	})
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

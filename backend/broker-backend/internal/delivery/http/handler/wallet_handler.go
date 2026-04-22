package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/yourusername/broker-backend/internal/delivery/http/middleware"
	"github.com/yourusername/broker-backend/internal/domain"
)

// WalletHandler, cüzdan endpoint'lerini yönetir.
// Tüm route'lar JWT middleware ile korunur.
type WalletHandler struct {
	usecase domain.WalletUsecase
}

// NewWalletHandler, yeni bir WalletHandler örneği döner.
func NewWalletHandler(usecase domain.WalletUsecase) *WalletHandler {
	return &WalletHandler{usecase: usecase}
}

// amountRequest, para yatırma/çekme için ortak istek yapısıdır.
// Decimal hassasiyeti için amount string olarak alınır.
type amountRequest struct {
	Amount string `json:"amount"`
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
//	Body: { "amount": "100.50" }
//	Yanıt: 200 OK → { güncel wallet }
func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	var req amountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "geçerli ve sıfırdan büyük bir tutar giriniz",
		})
		return
	}

	wallet, err := h.usecase.Deposit(r.Context(), userID, amount)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, wallet)
}

// Withdraw godoc
//
//	POST /api/wallet/withdraw
//	Body: { "amount": "50.00" }
//	Yanıt: 200 OK → { güncel wallet }
func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	var req amountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "geçerli ve sıfırdan büyük bir tutar giriniz",
		})
		return
	}

	wallet, err := h.usecase.Withdraw(r.Context(), userID, amount)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, wallet)
}

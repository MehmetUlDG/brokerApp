package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/yourusername/broker-backend/internal/delivery/http/middleware"
	"github.com/yourusername/broker-backend/internal/domain"
)

// OrderHandler, emir endpoint'lerini yönetir.
// Tüm route'lar JWT middleware ile korunur.
type OrderHandler struct {
	usecase domain.OrderUsecase
}

// NewOrderHandler, yeni bir OrderHandler örneği döner.
func NewOrderHandler(usecase domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: usecase}
}

// placeOrderRequest, yeni emir için istek yapısıdır.
// Decimal hassasiyeti için sayısal alanlar string olarak alınır.
type placeOrderRequest struct {
	Symbol   string `json:"symbol"`   // Örn: "BTCUSDT"
	Side     string `json:"side"`     // "BUY" | "SELL"
	Type     string `json:"type"`     // "MARKET" | "LIMIT"
	Quantity string `json:"quantity"` // Örn: "0.001"
	Price    string `json:"price"`    // LIMIT için hedef fiyat; MARKET'te boş bırakılabilir
}

// PlaceOrder godoc
//
//	POST /api/orders
//	Body: { "symbol", "side", "type", "quantity", "price" }
//	Yanıt: 201 Created → { oluşturulan order }
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	var req placeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	// Zorunlu alan kontrolü
	if req.Symbol == "" || req.Side == "" || req.Type == "" || req.Quantity == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "symbol, side, type ve quantity zorunludur",
		})
		return
	}

	// Side doğrulama
	if req.Side != "BUY" && req.Side != "SELL" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "side değeri BUY veya SELL olmalıdır",
		})
		return
	}

	// Type doğrulama
	if req.Type != "MARKET" && req.Type != "LIMIT" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "type değeri MARKET veya LIMIT olmalıdır",
		})
		return
	}

	// Miktar parse
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "geçerli ve sıfırdan büyük bir miktar giriniz",
		})
		return
	}

	// Fiyat parse (opsiyonel — MARKET emirlerde boş bırakılabilir)
	price := decimal.Zero
	if req.Price != "" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz fiyat formatı"})
			return
		}
	}

	// LIMIT emirlerde fiyat zorunlu
	if req.Type == "LIMIT" && price.IsZero() {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "LIMIT emirlerde price alanı zorunludur",
		})
		return
	}

	order, err := h.usecase.PlaceOrder(r.Context(), domain.PlaceOrderParams{
		UserID:   userID,
		Symbol:   req.Symbol,
		Side:     domain.OrderSide(req.Side),
		Type:     domain.OrderType(req.Type),
		Quantity: quantity,
		Price:    price,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

// GetOrders godoc
//
//	GET /api/orders
//	Yanıt: 200 OK → [ { order }, ... ] — giriş yapmış kullanıcıya ait emirler (en yeni önce)
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "yetkilendirme hatası"})
		return
	}

	orders, err := h.usecase.GetUserOrders(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "emirler getirilemedi"})
		return
	}

	// nil slice yerine boş dizi döndür (frontend dostu)
	if orders == nil {
		orders = []*domain.Order{}
	}

	respondJSON(w, http.StatusOK, orders)
}

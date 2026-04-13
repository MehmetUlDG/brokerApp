package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yourusername/broker-backend/internal/domain"
)

// respondJSON, HTTP yanıtına JSON formatında veri yazar.
// Content-Type otomatik olarak "application/json" yapılır.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Encoding hatası nadirdir, loglamak yeterlidir
		http.Error(w, `{"error":"yanıt kodlaması başarısız"}`, http.StatusInternalServerError)
	}
}

// handleDomainError, domain katmanı hatalarını uygun HTTP durum koduna dönüştürür.
// Bilinmeyen hatalar için 500 Internal Server Error döner.
func handleDomainError(w http.ResponseWriter, err error) {
	var domErr *domain.DomainError
	if errors.As(err, &domErr) {
		respondJSON(w, domainErrorToHTTPStatus(domErr.Code), map[string]string{
			"error": domErr.Message,
			"code":  domErr.Code,
		})
		return
	}
	// Bilinmeyen/beklenmeyen hata → 500 (detay istemciye sızdırılmaz)
	respondJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "dahili sunucu hatası",
	})
}

// domainErrorToHTTPStatus, domain hata kodunu HTTP durum koduna çevirir.
func domainErrorToHTTPStatus(code string) int {
	switch code {
	case "user_not_found", "wallet_not_found":
		return http.StatusNotFound // 404
	case "user_already_exists", "wallet_already_exists":
		return http.StatusConflict // 409
	case "invalid_credentials":
		return http.StatusUnauthorized // 401
	case "insufficient_balance":
		return http.StatusUnprocessableEntity // 422
	case "invalid_amount":
		return http.StatusBadRequest // 400
	default:
		return http.StatusBadRequest // 400
	}
}

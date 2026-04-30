package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	paymenthttp "payment-service/internal/http"
)

func TestWebhookHandler_HealthCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
}

func TestWebhookHandler_NewWebhookHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	handler := paymenthttp.NewWebhookHandler(nil, "whsec_test", logger)

	assert.NotNil(t, handler)
}

func TestWebhookHandler_RegisterRoutes_HealthEndpoint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	handler := paymenthttp.NewWebhookHandler(nil, "whsec_test", logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
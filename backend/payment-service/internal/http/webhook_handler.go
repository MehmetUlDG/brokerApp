package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/stripe/stripe-go/v76/webhook"

	"payment-service/internal/domain"
)

const webhookTimeout = 30 * time.Second

// WebhookHandler, Stripe'ın HTTP webhook callback'lerini işler.
// İmza doğrulaması her istek için yapılır.
type WebhookHandler struct {
	txRepo        domain.TransactionRepository
	webhookSecret string
	logger        *zap.Logger
}

// NewWebhookHandler, yeni bir WebhookHandler döner.
func NewWebhookHandler(txRepo domain.TransactionRepository, webhookSecret string, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		txRepo:        txRepo,
		webhookSecret: webhookSecret,
		logger:        logger,
	}
}

// RegisterRoutes, HTTP mux'a webhook rotasını kaydeder.
func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/stripe/webhook", h.HandleWebhook)
}

// HandleWebhook, Stripe webhook POST isteklerini işler.
//
//   - Stripe imzasını doğrular (Stripe-Signature header).
//   - payment_intent.succeeded → TX COMPLETED.
//   - payout.paid             → TX COMPLETED.
//   - Diğer event tipler sessizce görmezden gelinir.
//   - Her durumda 200 döner (Stripe yeniden denemesini önlemek için).
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("Webhook: body okunamadı", zap.Error(err))
		http.Error(w, "request body error", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(body, sigHeader, h.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		h.logger.Warn("Webhook: imza doğrulama başarısız", zap.Error(err))
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	go h.processEvent(string(event.Type), event.Data.Raw)

	w.WriteHeader(http.StatusOK)
}

// processEvent, event tipine göre ilgili DB güncellemesini yapar.
// 30 saniye timeout ile çalışır.
func (h *WebhookHandler) processEvent(eventType string, raw json.RawMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), webhookTimeout)
	defer cancel()

	switch eventType {
	case "payment_intent.succeeded":
		var data struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			h.logger.Error("Webhook: payment_intent.succeeded parse hatası", zap.Error(err))
			return
		}
		h.markCompleted(ctx, data.ID)

	case "payout.paid":
		var data struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			h.logger.Error("Webhook: payout.paid parse hatası", zap.Error(err))
			return
		}
		h.markCompleted(ctx, data.ID)

	default:
		h.logger.Debug("Webhook: bilinmeyen event tipi", zap.String("event_type", eventType))
	}
}

// markCompleted, StripeRef'e göre Transaction'ı bulur ve COMPLETED yapar (idempotent).
func (h *WebhookHandler) markCompleted(ctx context.Context, stripeRef string) {
	tx, err := h.txRepo.GetByStripeRef(ctx, stripeRef)
	if err != nil {
		h.logger.Error("Webhook: TX bulunamadı", zap.String("stripe_ref", stripeRef), zap.Error(err))
		return
	}

	if tx.Status == domain.TransactionStatusCompleted {
		h.logger.Debug("Webhook: TX zaten COMPLETED", zap.String("tx_id", tx.ID.String()))
		return
	}

	if err := h.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeRef); err != nil {
		h.logger.Error("Webhook: UpdateStatus başarısız",
			zap.String("tx_id", tx.ID.String()),
			zap.String("stripe_ref", stripeRef),
			zap.Error(err))
		return
	}

	h.logger.Info("Webhook: TX COMPLETED yapıldı",
		zap.String("tx_id", tx.ID.String()),
		zap.String("stripe_ref", stripeRef))
}

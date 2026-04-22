package http

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v76/webhook"

	"payment-service/internal/domain"
)

// =============================================================================
// WebhookHandler
// =============================================================================

// WebhookHandler, Stripe'ın HTTP webhook callback'lerini işler.
// İmza doğrulaması her istek için yapılır.
type WebhookHandler struct {
	txRepo        domain.TransactionRepository
	webhookSecret string
}

// NewWebhookHandler, yeni bir WebhookHandler döner.
func NewWebhookHandler(txRepo domain.TransactionRepository, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		txRepo:        txRepo,
		webhookSecret: webhookSecret,
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("⚠  Webhook: body okunamadı: %v", err)
		http.Error(w, "request body error", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(body, sigHeader, h.webhookSecret, webhook.ConstructEventOptions{
    IgnoreAPIVersionMismatch: true,
})
	if err != nil {
		log.Printf("⚠  Webhook: imza doğrulama başarısız: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// Non-blocking: 200 hemen döner; Stripe SLA'sı korunur.
	go h.processEvent(string(event.Type), event.Data.Raw)

	w.WriteHeader(http.StatusOK)
}

// processEvent, event tipine göre ilgili DB güncellemesini yapar.
// Background context kullanılır; request yaşam döngüsünden bağımsızdır.
func (h *WebhookHandler) processEvent(eventType string, raw json.RawMessage) {
	ctx := context.Background()

	switch eventType {
	case "payment_intent.succeeded":
		var data struct {
			ID string `json:"id"` // PaymentIntent ID = StripeRef
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			log.Printf("❌ Webhook: payment_intent.succeeded parse hatası: %v", err)
			return
		}
		h.markCompleted(ctx, data.ID)

	case "payout.paid":
		var data struct {
			ID string `json:"id"` // Payout ID = StripeRef
		}
		if err := json.Unmarshal(raw, &data); err != nil {
			log.Printf("❌ Webhook: payout.paid parse hatası: %v", err)
			return
		}
		h.markCompleted(ctx, data.ID)

	default:
		log.Printf("ℹ  Webhook: bilinmeyen event tipi, atlanıyor: %s", eventType)
	}
}

// markCompleted, StripeRef'e göre Transaction'ı bulur ve COMPLETED yapar (idempotent).
func (h *WebhookHandler) markCompleted(ctx context.Context, stripeRef string) {
	tx, err := h.txRepo.GetByStripeRef(ctx, stripeRef)
	if err != nil {
		log.Printf("❌ Webhook: TX bulunamadı (stripeRef=%s): %v", stripeRef, err)
		return
	}

	if tx.Status == domain.TransactionStatusCompleted {
		log.Printf("ℹ  Webhook: TX zaten COMPLETED, atlanıyor (id=%s)", tx.ID)
		return
	}

	if err := h.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted, stripeRef); err != nil {
		log.Printf("❌ Webhook: UpdateStatus başarısız (id=%s): %v", tx.ID, err)
		return
	}

	log.Printf("✅ Webhook: TX COMPLETED yapıldı (id=%s, stripeRef=%s)", tx.ID, stripeRef)
}

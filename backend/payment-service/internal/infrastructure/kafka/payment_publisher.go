package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// =============================================================================
// PaymentEventMsg — Kafka mesaj şeması
// =============================================================================

// PaymentEventMsg, payment-events topic'ine yazılan olayın JSON temsilidir.
type PaymentEventMsg struct {
	EventType     string `json:"event_type"`     // "deposit.completed" | "withdrawal.completed" | "deposit.failed"
	TransactionID string `json:"transaction_id"` // UUID string
	UserID        string `json:"user_id"`        // UUID string
	Amount        string `json:"amount"`         // decimal string, örn. "100.50"
	Currency      string `json:"currency"`       // "USD", "BTC" vs.
	StripeRef     string `json:"stripe_ref"`     // Stripe PaymentIntent/Payout ID
}

// =============================================================================
// PaymentPublisher
// =============================================================================

// PaymentPublisher, payment-events Kafka topic'ine mesaj yayımlar.
type PaymentPublisher struct {
	writer *kafkago.Writer
}

// NewPaymentPublisher, yeni bir PaymentPublisher döner.
// brokers: "host:port" listesi (virgülle ayrılmış single string veya slice).
func NewPaymentPublisher(brokers []string) *PaymentPublisher {
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Topic:                  "payment-events",
		Balancer:               &kafkago.LeastBytes{},
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		AllowAutoTopicCreation: true,
	}
	return &PaymentPublisher{writer: w}
}

// Publish, verilen PaymentEventMsg'yi JSON olarak Kafka'ya yazar.
// Mesaj anahtarı olarak TransactionID kullanılır (aynı işlemlerin aynı
// partition'a gitmesi için).
func (p *PaymentPublisher) Publish(ctx context.Context, msg PaymentEventMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("PaymentPublisher.Publish: marshal: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(msg.TransactionID),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("PaymentPublisher.Publish: write: %w", err)
	}
	return nil
}

// Close, Kafka writer'ı kapatır. Graceful shutdown sırasında çağrılmalıdır.
func (p *PaymentPublisher) Close() error {
	return p.writer.Close()
}

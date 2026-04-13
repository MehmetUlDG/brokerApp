package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	kafkago "github.com/segmentio/kafka-go"
)

// Processor, outbox_events tablosundaki PENDING event'leri okuyarak
// Kafka'ya ileten arka plan işçisidir (background worker).
//
// Transactional Outbox Pattern akışı:
//  1. DB'den PENDING event'ler çekilir (SELECT FOR UPDATE SKIP LOCKED → çakışma yok)
//  2. Her event Kafka'ya yazılır
//  3. Başarıysa status → "PROCESSED", başarısızsa → "FAILED"
//  4. Tüm güncellemeler aynı transaction'da commit edilir
//
// SKIP LOCKED garantisi: Birden fazla processor instance çalışsa bile
// aynı event iki kez işlenmez (at-least-once ≈ etkin çoğunlukla exactly-once).
type Processor struct {
	db     *sqlx.DB
	writer *kafkago.Writer
}

// outboxRow, veritabanından okunan outbox_events satırını temsil eder.
type outboxRow struct {
	ID            string          `db:"id"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   string          `db:"aggregate_id"`
	EventType     string          `db:"event_type"`
	Payload       json.RawMessage `db:"payload"` // JSONB → json.RawMessage
	Status        string          `db:"status"`
}

// kafkaEnvelope, Kafka'ya yazılacak mesajın standart formatıdır.
type kafkaEnvelope struct {
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

// NewProcessor, yeni bir Outbox Processor örneği döner.
// Kafka writer'ı AllowAutoTopicCreation ile başlatılır (geliştirme kolaylığı).
func NewProcessor(db *sqlx.DB, brokers []string, topic string) *Processor {
	return &Processor{
		db: db,
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

// Start, her 5 saniyede bir PENDING outbox event'lerini işler.
// ctx iptal edilene kadar çalışır; graceful shutdown için idealdir.
func (p *Processor) Start(ctx context.Context) {
	log.Println("📤 Outbox Processor başlatıldı (5s polling aralığı)")
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
		if err := p.writer.Close(); err != nil {
			log.Printf("[Outbox] Kafka writer kapatılamadı: %v", err)
		}
		log.Println("📤 Outbox Processor durduruldu")
	}()

	// İlk başlangıçta hemen bir kez çalıştır (tick beklemeden)
	if err := p.processPending(ctx); err != nil {
		log.Printf("[Outbox] İlk işlem hatası: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.processPending(ctx); err != nil {
				log.Printf("[Outbox] Polling hatası: %v", err)
			}
		}
	}
}

// processPending, PENDING durumundaki event'leri okuyup Kafka'ya gönderir.
// FOR UPDATE SKIP LOCKED: Birden fazla instance çalışıyorsa çakışma olmaz.
func (p *Processor) processPending(ctx context.Context) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback() // Commit başarılıysa etkisiz olur

	const selectQ = `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, status
		FROM   outbox_events
		WHERE  status = 'PENDING'
		ORDER  BY created_at ASC
		LIMIT  100
		FOR UPDATE SKIP LOCKED
	`

	var rows []outboxRow
	if err := tx.SelectContext(ctx, &rows, selectQ); err != nil {
		return fmt.Errorf("event seçimi başarısız: %w", err)
	}

	if len(rows) == 0 {
		return tx.Commit() // Boş transaction commit
	}

	log.Printf("[Outbox] %d PENDING event işleniyor...", len(rows))

	for _, row := range rows {
		newStatus := "PROCESSED"

		envelope := kafkaEnvelope{
			EventType:     row.EventType,
			AggregateType: row.AggregateType,
			AggregateID:   row.AggregateID,
			Payload:       row.Payload,
		}

		msgBytes, err := json.Marshal(envelope)
		if err != nil {
			log.Printf("[Outbox] Envelope marshal hatası (id=%s): %v", row.ID, err)
			newStatus = "FAILED"
		} else if err := p.writer.WriteMessages(ctx, kafkago.Message{
			Key:   []byte(row.AggregateID),
			Value: msgBytes,
		}); err != nil {
			log.Printf("[Outbox] Kafka gönderimi başarısız (id=%s): %v", row.ID, err)
			newStatus = "FAILED"
		}

		const updateQ = `UPDATE outbox_events SET status = $1 WHERE id = $2`
		if _, updateErr := tx.ExecContext(ctx, updateQ, newStatus, row.ID); updateErr != nil {
			log.Printf("[Outbox] Durum güncellenemedi (id=%s): %v", row.ID, updateErr)
		}

		if newStatus == "PROCESSED" {
			log.Printf("[Outbox] ✅ Kafka'ya gönderildi: %s/%s", row.EventType, row.ID)
		}
	}

	return tx.Commit()
}

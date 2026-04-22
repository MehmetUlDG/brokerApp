package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"

	"github.com/yourusername/broker-backend/internal/domain"
)

// =============================================================================
// TradeConsumer — trade-executed Kafka topic tüketicisi
// =============================================================================

// TradeConsumer, matching engine'in yayımladığı trade-executed olaylarını
// tüketerek emir durumunu COMPLETED'a getirir ve ilgili cüzdan transferlerini
// gerçekleştirir.
type TradeConsumer struct {
	orderRepo     domain.OrderRepository
	walletUsecase domain.WalletUsecase
	reader        *kafkago.Reader
}

// NewTradeConsumer, yeni bir TradeConsumer örneği döner.
func NewTradeConsumer(
	orderRepo domain.OrderRepository,
	walletUsecase domain.WalletUsecase,
	reader *kafkago.Reader,
) *TradeConsumer {
	return &TradeConsumer{
		orderRepo:    orderRepo,
		walletUsecase: walletUsecase,
		reader:       reader,
	}
}

// =============================================================================
// TradeExecutedMsg — Kafka mesaj şeması
// =============================================================================

// TradeExecutedMsg, matching engine'den gelen trade-executed Kafka mesajının
// JSON temsilidir.
//   - ExecPrice float64 olarak gelir (matching engine kaynaklı); decimal'e çevrilir.
//   - Quantity   string  olarak gelir; decimal.RequireFromString ile parse edilir.
type TradeExecutedMsg struct {
	OrderID   string  `json:"order_id"`
	ExecPrice float64 `json:"exec_price"`
	Quantity  string  `json:"quantity"`
}

// =============================================================================
// Start — Ana tüketim döngüsü
// =============================================================================

// Start, trade-executed topic'ini sürekli dinler ve her mesaj için:
//  1. JSON'u TradeExecutedMsg'ye açar.
//  2. OrderID'yi uuid.UUID'ye parse eder.
//  3. orderRepo.GetByID ile ilgili emri çeker.
//  4. Emir PENDING değilse mesajı commit eder ve atlar (idempotency koruması).
//  5. ExecPrice ve Quantity'yi decimal.Decimal'e dönüştürür.
//  6. walletUsecase.TransferForOrder ile bakiye transferini gerçekleştirir.
//  7. orderRepo.UpdateStatus ile emri COMPLETED yapar.
//  8. Hata durumunda log'lar ama döngüyü kırmaz (servis devam eder).
//  9. Mesajı commit eder.
//
// ctx iptal edildiğinde (graceful shutdown) döngüden çıkar.
func (c *TradeConsumer) Start(ctx context.Context) {
	log.Println("✅ Trade Consumer başlatıldı — trade-executed topic'i dinleniyor")

	for {
		// ctx iptal kontrolü — her iterasyonda erken çıkış
		select {
		case <-ctx.Done():
			log.Println("⏹  Trade Consumer durduruluyor (context iptal)")
			return
		default:
		}

		// ── Mesaj al ──────────────────────────────────────────────────────────
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			// ctx iptal edilmişse sessizce çık
			if ctx.Err() != nil {
				log.Println("⏹  Trade Consumer durduruluyor (context iptal)")
				return
			}
			log.Printf("❌ Trade Consumer: mesaj alınamadı: %v", err)
			continue
		}

		log.Printf("📨 Trade Consumer: mesaj alındı (offset=%d, key=%s)", msg.Offset, string(msg.Key))

		// ── JSON parse ────────────────────────────────────────────────────────
		var tradeMsg TradeExecutedMsg
		if err := json.Unmarshal(msg.Value, &tradeMsg); err != nil {
			log.Printf("❌ Trade Consumer: JSON parse hatası: %v | ham değer: %s", err, string(msg.Value))
			_ = c.commitMsg(ctx, msg)
			continue
		}

		// ── OrderID parse ─────────────────────────────────────────────────────
		orderID, err := uuid.Parse(tradeMsg.OrderID)
		if err != nil {
			log.Printf("❌ Trade Consumer: geçersiz order_id=%q: %v", tradeMsg.OrderID, err)
			_ = c.commitMsg(ctx, msg)
			continue
		}

		// ── Emri getir ────────────────────────────────────────────────────────
		order, err := c.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			log.Printf("❌ Trade Consumer: emir getirilemedi (id=%s): %v", orderID, err)
			_ = c.commitMsg(ctx, msg)
			continue
		}

		// ── Idempotency: sadece PENDING emirleri işle ─────────────────────────
		if order.Status != domain.OrderStatusPending {
			log.Printf("⏭  Trade Consumer: emir zaten işlenmiş, atlanıyor (id=%s, status=%s)",
				orderID, order.Status)
			_ = c.commitMsg(ctx, msg)
			continue
		}

		// ── Decimal dönüşümleri ───────────────────────────────────────────────
		execPrice := decimal.NewFromFloat(tradeMsg.ExecPrice)
		quantity := decimal.RequireFromString(tradeMsg.Quantity)

		// ── Cüzdan transferi ──────────────────────────────────────────────────
		if _, err := c.walletUsecase.TransferForOrder(ctx, order.UserID, string(order.Side), quantity, execPrice); err != nil {
			log.Printf("❌ Trade Consumer: cüzdan transferi başarısız (id=%s): %v", orderID, err)
			// Hata loglanır ama döngü devam eder; mesaj commit edilmez —
			// bir sonraki iterasyonda yeniden deneme mümkün değil (zaten commit
			// edilmediği için Kafka mesajı kalmaya devam eder). İdempotency
			// koruması yeniden işleme güvencesi sağlar.
			// NOT: Burada commit YAPILMIYOR → mesaj yeniden teslim edilebilir.
			continue
		}

		// ── Emir durumunu güncelle ────────────────────────────────────────────
		if err := c.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusCompleted); err != nil {
			log.Printf("❌ Trade Consumer: emir durumu güncellenemedi (id=%s): %v", orderID, err)
			// Aynı şekilde commit yapılmıyor — yeniden deneme imkânı korunur.
			continue
		}

		log.Printf("✅ Trade Consumer: emir tamamlandı (id=%s, side=%s, qty=%s, price=%s)",
			orderID, order.Side, quantity.String(), execPrice.String())

		// ── Mesajı commit et ──────────────────────────────────────────────────
		_ = c.commitMsg(ctx, msg)
	}
}

// commitMsg, verilen Kafka mesajını commit eder ve oluşan hatayı loglar.
func (c *TradeConsumer) commitMsg(ctx context.Context, msg kafkago.Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("⚠  Trade Consumer: mesaj commit edilemedi (offset=%d): %v", msg.Offset, err)
		return err
	}
	return nil
}

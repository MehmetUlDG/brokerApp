package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// ============================================================================
// Modeller
// ============================================================================

// Ingestion servisinden gelen anlık fiyat modeli
type LivePriceMsg struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"` // Kolaylık için string, ancak engine içinde parse edeceğiz
	Timestamp int64  `json:"timestamp"`
}

// Order servisinden gelen yeni (PENDING) emir modeli
type OrderMsg struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Symbol   string `json:"symbol"`
	Side     string `json:"side"`     // "BUY" veya "SELL"
	Type     string `json:"type"`     // "MARKET" veya "LIMIT"
	Quantity string `json:"quantity"` // Kolaylık için string
	Price    string `json:"price"`    // LIMIT emirler için hedef fiyat
}

// Eşleşme gerçekleştiğinde Kafka'ya gönderilecek event modeli
type TradeExecutedMsg struct {
	OrderID   string  `json:"order_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Quantity  string  `json:"quantity"`
	ExecPrice float64 `json:"exec_price"`
	Timestamp int64   `json:"timestamp"`
}

// ============================================================================
// Engine State (Stateful Yapı)
// ============================================================================

type MatchingEngine struct {
	// Anlık fiyatları tutmak için RWMutex. Birden fazla okuyucu aynı anda okuyabilir;
	// Sadece 'live-prices' goroutine'i güncelleme (yazma) yaparken kilitler.
	priceMu sync.RWMutex
	prices  map[string]float64 // Örn: {"BTCUSDT": 65000.50}

	// Bekleyen Limit ve Market emirler. Gerçek bir sistemde Heap / OrderBook kullanılır,
	// Burada basit bir Slice ile demostrasyon yapılmıştır. Slice üzerinde thread-safe
	// çalışmak için kendi Mutex'i bulunur.
	ordersMu sync.Mutex
	orders   []OrderMsg

	// Çıktı için Kafka Writer
	producer *kafka.Writer

	// İletişim Kanalları
	priceChan chan LivePriceMsg
	orderChan chan OrderMsg
}

func NewMatchingEngine(brokerAddr string) *MatchingEngine {
	return &MatchingEngine{
		prices: make(map[string]float64),
		orders: make([]OrderMsg, 0),
		producer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Topic:    "trade-executed",
			Balancer: &kafka.LeastBytes{},
		},
		priceChan: make(chan LivePriceMsg, 1000), // Buffer
		orderChan: make(chan OrderMsg, 1000),     // Buffer
	}
}

// ============================================================================
// Engine Start: Tüketiciler, Üreticiler ve Matcher Goroutine'leri
// ============================================================================

func (e *MatchingEngine) Start(ctx context.Context, brokerAddr string) {
	log.Println("⚡ Matching Engine başlatılıyor... (Stateful & RAM-based)")

	// 1. Live Prices Dinleyicisi (Goroutine)
	go e.consumeLivePrices(ctx, brokerAddr)

	// 2. New Orders Dinleyicisi (Goroutine)
	go e.consumeNewOrders(ctx, brokerAddr)

	// 3. Eşleştirme Motoru Döngüsü (Goroutine)
	go e.matchLoop(ctx)
}

func (e *MatchingEngine) Stop() {
	e.producer.Close()
	close(e.priceChan)
	close(e.orderChan)
}

// ----------------------------------------------------------------------------
// Tüketiciler (Consumers) - Kafka -> Channel
// ----------------------------------------------------------------------------

func (e *MatchingEngine) consumeLivePrices(ctx context.Context, brokerAddr string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   "live-prices",
		GroupID: "matching-engine-prices-group", // Tüm engine'ler okumalıysa ayrı group olabilir
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Err] live-prices okunamadı: %v", err)
			continue
		}

		var priceMsg LivePriceMsg
		if err := json.Unmarshal(msg.Value, &priceMsg); err == nil {
			// İşlenmesi için channel'a yolla
			e.priceChan <- priceMsg
		}
	}
}

func (e *MatchingEngine) consumeNewOrders(ctx context.Context, brokerAddr string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   "new-orders",
		GroupID: "matching-engine-orders-group",
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Err] new-orders okunamadı: %v", err)
			continue
		}

		var orderMsg OrderMsg
		if err := json.Unmarshal(msg.Value, &orderMsg); err == nil {
			// İşlenmesi için channel'a yolla
			e.orderChan <- orderMsg
		}
	}
}

// ----------------------------------------------------------------------------
// Çekirdek Engine (Matching Process)
// ----------------------------------------------------------------------------

func (e *MatchingEngine) matchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Match Loop sonlandırılıyor.")
			return

		// Durum 1: Yeni Fiyat Geldi
		case p := <-e.priceChan:
			priceF, _ := strconv.ParseFloat(p.Price, 64)

			// 1. Gelen fiyatı State'e yaz (RWMutex Writer Lock)
			e.priceMu.Lock()
			e.prices[p.Symbol] = priceF
			e.priceMu.Unlock()

			// 2. Fiyat değişimiyle bekleyen emirleri tara
			e.scanAndMatchOrders(p.Symbol, priceF)

		// Durum 2: Yeni Emir Geldi
		case o := <-e.orderChan:
			// Önce anlık piyasa fiyatını al (RWMutex Reader Lock)
			e.priceMu.RLock()
			currentPrice, exists := e.prices[o.Symbol]
			e.priceMu.RUnlock()

			if !exists {
				// Eşleştirme henüz mümkün değil, sıraya (Slice) at.
				e.ordersMu.Lock()
				e.orders = append(e.orders, o)
				e.ordersMu.Unlock()
				log.Printf("Fiyat verisi olmadığı için emir sırda bekliyor: %s", o.ID)
				continue
			}

			// Emir bu fiyatta hemen eşleşebilir mi kontrol et
			matched := e.tryMatchSingleOrder(o, currentPrice)
			if !matched {
				// Eşleşmedi, Slice'a ekle ve bekle
				e.ordersMu.Lock()
				e.orders = append(e.orders, o)
				e.ordersMu.Unlock()
				log.Printf("Limit emri piyasa fiyatına (%v) uymadı, havuza eklendi: %s", currentPrice, o.ID)
			}
		}
	}
}

func (e *MatchingEngine) tryMatchSingleOrder(o OrderMsg, currentPrice float64) bool {
	// Market emriyse her türlü eşleşir
	if o.Type == "MARKET" {
		e.executeTrade(o, currentPrice)
		return true
	}

	// Limit emriyse şarta bakılır
	limitVal, _ := strconv.ParseFloat(o.Price, 64)

	if o.Side == "BUY" && currentPrice <= limitVal {
		e.executeTrade(o, currentPrice) // Düşükten aldık
		return true
	} else if o.Side == "SELL" && currentPrice >= limitVal {
		e.executeTrade(o, currentPrice) // Yüksekten sattık
		return true
	}

	return false
}

// Slice tarayarak gerçekleşenleri tespit etme
func (e *MatchingEngine) scanAndMatchOrders(symbol string, currentPrice float64) {
	e.ordersMu.Lock()
	defer e.ordersMu.Unlock()

	var remaining []OrderMsg

	for _, o := range e.orders {
		if o.Symbol != symbol {
			remaining = append(remaining, o)
			continue
		}

		matched := e.tryMatchSingleOrder(o, currentPrice)
		if !matched {
			remaining = append(remaining, o) // Eşleşmeyenleri listede tutmaya devam et
		}
	}

	// Sadece çalışmayan/eşleşmeyen emirleri güncel listeye taşı (Slicing)
	e.orders = remaining
}

// ----------------------------------------------------------------------------
// Çıktı Üretici (Producer) - Kafka'ya trade-executed fırlatma
// ----------------------------------------------------------------------------

func (e *MatchingEngine) executeTrade(o OrderMsg, execPrice float64) {
	log.Printf("[MATCH!] '%s' emri %.2f fiyatından eşleşti. Kafka'ya gönderiliyor...", o.ID, execPrice)

	tradeEvent := TradeExecutedMsg{
		OrderID:   o.ID,
		Symbol:    o.Symbol,
		Side:      o.Side,
		Quantity:  o.Quantity,
		ExecPrice: execPrice,
		Timestamp: time.Now().UnixMilli(),
	}

	payload, _ := json.Marshal(tradeEvent)

	// trade-executed topic'ini tetikleyerek Wallet Service / Outbox Process tarafında bakiyeleri güncelletir
	err := e.producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(o.ID),
			Value: payload,
		},
	)

	if err != nil {
		log.Printf("[Err] Trade Kafka'ya basılamadı: %v", err)
	}
}

// ============================================================================
// MAIN ENTRYPOINT
// ============================================================================

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	brokerAddr := "localhost:9092" // Sistem ayağa kalktığında Docker'daki adress
	engine := NewMatchingEngine(brokerAddr)

	engine.Start(ctx, brokerAddr)

	// Servis arka planda goroutine'lerde çalıştığı için main() fonksiyonunu ayakta tut
	// Normalde os.Signal ile graceful shutdown yapılır.
	select {}
}

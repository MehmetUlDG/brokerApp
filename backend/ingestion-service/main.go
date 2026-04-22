package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

// BinanceTradeMsg stream'den gelen ham verinin (trade event) yapısı
type BinanceTradeMsg struct {
	EventType string `json:"e"` // Event type
	EventTime int64  `json:"E"` // Event time
	Symbol    string `json:"s"` // Symbol (Örn: BTCUSDT)
	TradeID   int64  `json:"t"` // Trade ID
	Price     string `json:"p"` // Price
	Quantity  string `json:"q"` // Quantity
	BuyerID   int64  `json:"b"` // Buyer order ID
	SellerID  int64  `json:"a"` // Seller order ID
	TradeTime int64  `json:"T"` // Trade time
	IsMarket  bool   `json:"m"` // Is the buyer the market maker?
}

// LivePriceMsg sistemin iç kısmına (Kafka'ya) gönderilecek normalize edilmiş mesaj
type LivePriceMsg struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Timestamp int64  `json:"timestamp"`
}

var kafkaWriter *kafka.Writer

func initKafkaWriter() {
	// Kafka Broker adresi Docker Compose üzerinden sağlanır (Örn: localhost:9092)
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"), // Geliştirme ortamı için localhost
		Topic:    "live-prices",
		Balancer: &kafka.LeastBytes{},
	}
}

func main() {
	log.Println("🚀 Ingestion Service başlıyor...")
	initKafkaWriter()
	defer kafkaWriter.Close()

	url := "wss://stream.binance.com:443/ws/btcusdt@trade"
	var retryCount float64 = 0

	for {
		log.Printf("[Connecting] %s adresine bağlanılıyor...", url)

		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("[Error] Bağlantı kurulamadı: %v", err)
			
			// Exponential Backoff Mantığı: 2^retryCount saniye bekle
			backoffDuration := time.Duration(math.Pow(2, retryCount)) * time.Second
			
			// En fazla 60 saniye beklet
			if backoffDuration > 60*time.Second {
				backoffDuration = 60 * time.Second
			}
			
			log.Printf("[Backoff] %v saniye sonra tekrar denenecek...", backoffDuration.Seconds())
			time.Sleep(backoffDuration)
			retryCount++
			continue
		}

		// Başarılı bağlandıysa retryCount'u sıfırla
		log.Println("[Success] Başarıyla bağlanıldı, trade verileri dinleniyor...")
		retryCount = 0

		// Verileri dinleyen fonksiyonu başlat
		processStream(conn)

		// processStream biterse = bağlantı bir sebeple kapandı demektir
		conn.Close()
		log.Println("[Warning] Bağlantı koptu, yeniden bağlanma döngüsüne giriliyor...")
		time.Sleep(1 * time.Second)
	}
}

// processStream, websocket bağlantısından verileri aralıksız okur ve Kafka'ya yazar
func processStream(conn *websocket.Conn) {
	for {
		// WebSocket üzerinden mesaj oku
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Disconnect] WebSocket okuma hatası: %v", err)
			break // Read hatası fırlatıldıysa döngüyü kır, dış loop bağlantıyı yeniden deneyecek
		}

		// Sadece JSON parse işlemi
		var tradeMsg BinanceTradeMsg
		if err := json.Unmarshal(message, &tradeMsg); err != nil {
			log.Printf("[Warning] JSON dönüştürme hatası: %v", err)
			continue
		}

		// Sadece Sembol, Fiyat ve Timestamp bilgilerini al
		livePrice := LivePriceMsg{
			Symbol:    tradeMsg.Symbol,
			Price:     tradeMsg.Price,
			Timestamp: time.Now().UnixMilli(),
		}

		payload, err := json.Marshal(livePrice)
		if err != nil {
			log.Printf("[Warning] LivePriceMsg marshal hatası: %v", err)
			continue
		}

		// Kafka'ya Produce et
		// Aynı sembole ait olan fiyatlar sırayla okunsun diye Partition Key olarak Symbol verilebilir.
		err = kafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(livePrice.Symbol),
				Value: payload,
			},
		)

		if err != nil {
			log.Printf("[Kafka Error] Mesaj üretilemedi: %v", err)
		} else {
			// Başarılı gönderimleri ekrana yazdır (Development amaçlı)
			log.Printf("[Live Price -> Kafka] %s: %s", livePrice.Symbol, livePrice.Price)
		}
	}
}

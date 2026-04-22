package kafka

import (
	"context"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

// Producer, Kafka'ya mesaj yazan bileşeni sarmalayan yapıdır.
// kafka-go paketini doğrudan kullanmak yerine bu wrapper,
// bağımlılık tersine çevirme prensibine uygun soyutlama sağlar.
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer, belirtilen broker adresleri ve topic üzerinde çalışan
// yeni bir Kafka Producer örneği döner.
//
// AllowAutoTopicCreation: geliştirme ortamında topic'lerin otomatik
// oluşturulmasını sağlar. Production'da false yapılması önerilir.
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

// Publish, verilen key ve value ile Kafka'ya senkron bir mesaj yazar.
// ctx iptal edilirse veya Kafka bağlanamıyorsa hata döner.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	if err := p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   key,
		Value: value,
	}); err != nil {
		return fmt.Errorf("kafka mesaj gönderilemedi (topic=%s): %w", p.writer.Topic, err)
	}
	return nil
}

// Close, Kafka producer bağlantısını temiz şekilde kapatır.
// Graceful shutdown içinde çağrılmalıdır.
func (p *Producer) Close() error {
	return p.writer.Close()
}

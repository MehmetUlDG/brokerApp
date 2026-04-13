package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/yourusername/broker-backend/internal/domain"
)


type orderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository yaratır.
func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

// PlaceOrder fonksiyonu, hem Order'ı kaydeder hem de OrderCreated event'ini
// outbox_events tablosuna yazar. Bu işlemler AYNI TRANSACTION içinde gerçekleşir.
// Böylece sistemin herhangi bir yerinde çökme yaşanırsa, veri bütünlüğü bozulmaz.
func (r *orderRepository) PlaceOrder(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error {
	// 1. Transaction başlatıyoruz.
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// Hata olursa veya panik yaşanırsa işlemleri geri alması için defer kurgusu.
	// tx.Commit() başarılı olduğunda tx.Rollback() etkisiz olacaktır.
	defer tx.Rollback()

	// 2. Order'ı PENDING olarak orders tablosuna kaydet
	orderQuery := `
		INSERT INTO orders (id, user_id, symbol, side, type, quantity, price, status, created_at, updated_at)
		VALUES (:id, :user_id, :symbol, :side, :type, :quantity, :price, :status, :created_at, :updated_at)
	`
	_, err = tx.NamedExecContext(ctx, orderQuery, order)
	if err != nil {
		return err // Dönüş öncesi defer tx.Rollback() çalışır
	}

	// 3. Outbox event'i, aynı transaction kapsamında tablolara kaydet
	outboxQuery := `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status, created_at)
		VALUES (:id, :aggregate_type, :aggregate_id, :event_type, :payload, :status, :created_at)
	`
	_, err = tx.NamedExecContext(ctx, outboxQuery, event)
	if err != nil {
		return err // Dönüş öncesi defer tx.Rollback() çalışır
	}

	// 4. Her iki kayıt da başarılıysa, Transaction'ı Commit et.
	// Bu aşamadan sonra veriler diğer transactionlar tarafından okunabilir hale gelir.
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

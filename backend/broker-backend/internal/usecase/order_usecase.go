package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourusername/broker-backend/internal/domain"
)

type orderUsecase struct {
	repo domain.OrderRepository
}

func NewOrderUsecase(repo domain.OrderRepository) domain.OrderUsecase {
	return &orderUsecase{repo: repo}
}

// PlaceOrder, kullanıcıdan gelen emir isteğiyle yeni bir emir oluşturur
// ve bu emri repository katmanına göndererek kaydedilmesini sağlar.
func (u *orderUsecase) PlaceOrder(ctx context.Context, params domain.PlaceOrderParams) (*domain.Order, error) {
	now := time.Now()

	// 1. Domain nesnesini oluştur ("PENDING" durumunda)
	newOrder := &domain.Order{
		ID:        uuid.New(),
		UserID:    params.UserID,
		Symbol:    params.Symbol,
		Side:      params.Side,
		Type:      params.Type,
		Quantity:  params.Quantity,
		Price:     params.Price,
		Status:    domain.OrderStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 2. Outbox Event için Payload'ı (JSON) hazırla
	payloadBytes, err := json.Marshal(newOrder)
	if err != nil {
		return nil, err
	}

	outboxEvent := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "ORDER",
		AggregateID:   newOrder.ID.String(),
		EventType:     "OrderCreated",
		Payload:       payloadBytes,
		Status:        "PENDING",
		CreatedAt:     now,
	}

	// 3. Atomik kayıt için (hem sipariş hem de event) Repository'e pasla
	if err := u.repo.PlaceOrder(ctx, newOrder, outboxEvent); err != nil {
		return nil, err
	}

	// Bilgileri dön
	return newOrder, nil
}

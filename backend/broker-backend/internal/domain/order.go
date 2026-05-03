package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ErrOrderNotFound, belirtilen ID'ye ait emir bulunamadığında döner.
var ErrOrderNotFound = NewDomainError("order_not_found", "emir bulunamadı")

// =============================================================================
// Order Enums
// =============================================================================

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusFailed    OrderStatus = "FAILED"
	OrderStatusCanceled  OrderStatus = "CANCELED"
)

// =============================================================================
// Order Entity
// =============================================================================

type Order struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	UserID    uuid.UUID       `db:"user_id" json:"user_id"`
	Symbol    string          `db:"symbol" json:"symbol"`   // Örn: 'BTCUSDT'
	Side      OrderSide       `db:"side" json:"side"`       // BUY veya SELL
	Type      OrderType       `db:"type" json:"type"`       // MARKET veya LIMIT
	Quantity  decimal.Decimal `db:"quantity" json:"quantity"`
	Price     decimal.Decimal `db:"price" json:"price"`     // LIMIT için emir fiyatı; MARKET=0
	Status    OrderStatus     `db:"status" json:"status"`   // PENDING, vb.
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// =============================================================================
// Outbox Event Entity (Genel Amaçlı)
// =============================================================================

type OutboxEvent struct {
	ID            uuid.UUID `db:"id" json:"id"`
	AggregateType string    `db:"aggregate_type" json:"aggregate_type"`
	AggregateID   string    `db:"aggregate_id" json:"aggregate_id"`
	EventType     string    `db:"event_type" json:"event_type"`
	Payload       []byte    `db:"payload" json:"payload"` // JSON tutulur
	Status        string    `db:"status" json:"status"`   // PENDING, PROCESSED, FAILED
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// =============================================================================
// Usecase ve Repository Interfaceleri
// =============================================================================

type PlaceOrderParams struct {
	UserID   uuid.UUID
	Symbol   string          // İşlem sembolü (örn: 'BTCUSDT') — zorunlu
	Side     OrderSide
	Type     OrderType
	Quantity decimal.Decimal
	Price    decimal.Decimal
}

type OrderRepository interface {
	// PlaceOrder emir girme işlemini (orders tablosuna) ve event fırlatmayı
	// (outbox_events tablosuna) AYNI transaction içerisinde atomik olarak gerçekleştirir.
	// Outbox Pattern: herhangi bir çökmede veri bütünlüğü bozulmaz.
	PlaceOrder(ctx context.Context, order *Order, event *OutboxEvent) error

	// GetByID, verilen UUID'ye ait emri döner.
	// Emir bulunamazsa ErrOrderNotFound döner.
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)

	// GetUserOrders, belirli bir kullanıcıya ait tüm emirleri döner (en yeni önce).
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*Order, error)

	// UpdateStatus, belirtilen emrin durumunu günceller ve updated_at'i tazeler.
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
}

type OrderUsecase interface {
	PlaceOrder(ctx context.Context, params PlaceOrderParams) (*Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*Order, error)
}

package ports

import (
	"context"
	"order-service/internal/models"
)

type OrderCache interface {
	SetOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
	PreloadOrders(ctx context.Context, order []models.Order) error
	DeleteOrder(ctx context.Context, orderUID string) error
	Close() error
}
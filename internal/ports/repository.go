package ports

import (
	"context"
	"order-service/internal/models"
)

//go:generate mockgen -source=repository.go -destination=../mocks/mock_repository.go -package=mocks

// For default operations
type OrderRepository interface {
	// Basic operations
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
	SaveOrder(ctx context.Context, order *models.Order) error
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	DeleteOrder(ctx context.Context, orderUID string) error

	// Transaction operations
	BeginTx(ctx context.Context) (OrderTx, error)
	WithTransaction(ctx context.Context, fn func(tx OrderTx) error) error
}

// For transactions
type OrderTx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
	SaveOrder(ctx context.Context, order *models.Order) error
	DeleteOrder(ctx context.Context, orderUID string) error
}
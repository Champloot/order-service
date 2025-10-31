package ports

import (
	"context"
)

//go:generate mockgen -source=consumer.go -destination=../mocks/mock_consumer.go -package=mocks

type OrderConsumer interface {
	Start(ctx context.Context)
	Close() error
}
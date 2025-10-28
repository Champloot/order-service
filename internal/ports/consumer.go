package ports

import (
	"context"
)

type OrderConsumer interface {
	Start(ctx context.Context)
	Close() error
}
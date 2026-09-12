package ports

import (
	"context"
)

type Repository[T any, ID comparable] interface {
	Read(ctx context.Context, id ID) (T, error)
	Write(ctx context.Context, entity T) error
}

package book

import "context"

type Store interface {
	CurrentBook(ctx context.Context, serverID string) (Book, error)
	SaveCurrentBook(ctx context.Context, serverID string, book Book) error
}

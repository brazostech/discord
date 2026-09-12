package book

import "context"

type BookRepository struct{}

func (r BookRepository) Read(ctx context.Context, id string) (book Book, err error) {
	panic("unimplemented")
}

func (r BookRepository) Write(ctx context.Context, book Book) (err error) {
	panic("unimplemented")
}

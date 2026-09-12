package book

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/brazostech/discord/discord/interactions"
)

type FakeRepository struct {
	data map[string]Book
}

func (r *FakeRepository) Read(ctx context.Context, id string) (Book, error) {
	if book, ok := r.data[id]; ok {
		return book, nil
	}

	return Book{}, errors.New("no book :(")
}

func (r *FakeRepository) Write(ctx context.Context, book Book) error {
	if book.Id == "" {
		book.Id = uuid.New().String()
	}

	r.data[book.Id] = book
	return nil
}

// TODO: implement me
func TestBookService(t *testing.T) {
	fr := FakeRepository{}
	bs := NewBookService(&fr)

	var interactionPacket interactions.InteractionPacket

	_, err := bs.CommandBookHandler(t.Context(), interactionPacket)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

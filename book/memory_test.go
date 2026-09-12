package book

import (
	"errors"
	"testing"
)

func TestMemoryStoreRoundTrip(t *testing.T) {
	store := NewMemoryStore()
	want := Book{Name: "Dune", Url: "https://example.com/dune", Chapter: 3}

	if err := store.SaveCurrentBook(t.Context(), "server-1", want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMemoryStoreCurrentBookNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.CurrentBook(t.Context(), "server-1")
	if !errors.Is(err, ErrNoCurrentBook) {
		t.Fatalf("error = %v, want %v", err, ErrNoCurrentBook)
	}
}

func TestMemoryStoreSaveOverwrites(t *testing.T) {
	store := NewMemoryStore()
	ctx := t.Context()

	if err := store.SaveCurrentBook(ctx, "server-1", Book{Name: "Dune", Chapter: 4}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.SaveCurrentBook(ctx, "server-1", Book{Name: "Neuromancer", Chapter: 0}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.CurrentBook(ctx, "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if got.Name != "Neuromancer" || got.Chapter != 0 {
		t.Fatalf("got %+v, want the overwritten book", got)
	}
}

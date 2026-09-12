package book

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterStoresNewCurrentBook(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)

	got, err := service.Register(t.Context(), "server-1", "Dune", "https://example.com/dune")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	want := Book{Name: "Dune", Url: "https://example.com/dune", Chapter: 0}
	if got != want {
		t.Fatalf("register returned %+v, want %+v", got, want)
	}

	stored, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored != want {
		t.Fatalf("stored %+v, want %+v", stored, want)
	}
}

func TestRegisterReplacesCurrentBookAndResetsChapter(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := t.Context()

	if _, err := service.Register(ctx, "server-1", "Dune", ""); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if _, err := service.UpdateChapter(ctx, "server-1", 9); err != nil {
		t.Fatalf("update chapter: %v", err)
	}

	got, err := service.Register(ctx, "server-1", "Neuromancer", "https://example.com/neuromancer")
	if err != nil {
		t.Fatalf("second register: %v", err)
	}

	want := Book{Name: "Neuromancer", Url: "https://example.com/neuromancer", Chapter: 0}
	if got != want {
		t.Fatalf("register returned %+v, want %+v", got, want)
	}

	stored, err := store.CurrentBook(ctx, "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored != want {
		t.Fatalf("stored %+v, want %+v", stored, want)
	}
}

func TestRegisterRejectsEmptyName(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)

	_, err := service.Register(t.Context(), "server-1", "", "https://example.com")
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("register error = %v, want %v", err, ErrNameRequired)
	}

	if _, err := store.CurrentBook(t.Context(), "server-1"); !errors.Is(err, ErrNoCurrentBook) {
		t.Fatalf("store error = %v, want %v", err, ErrNoCurrentBook)
	}
}

func TestUpdateChapterUpdatesCurrentBook(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := t.Context()

	if _, err := service.Register(ctx, "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}

	got, err := service.UpdateChapter(ctx, "server-1", 7)
	if err != nil {
		t.Fatalf("update chapter: %v", err)
	}

	want := Book{Name: "Dune", Chapter: 7}
	if got != want {
		t.Fatalf("update returned %+v, want %+v", got, want)
	}

	stored, err := store.CurrentBook(ctx, "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored != want {
		t.Fatalf("stored %+v, want %+v", stored, want)
	}
}

func TestUpdateChapterMovesBackward(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := t.Context()

	if _, err := service.Register(ctx, "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := service.UpdateChapter(ctx, "server-1", 8); err != nil {
		t.Fatalf("update to 8: %v", err)
	}

	got, err := service.UpdateChapter(ctx, "server-1", 6)
	if err != nil {
		t.Fatalf("update to 6: %v", err)
	}
	if got.Chapter != 6 {
		t.Fatalf("chapter = %d, want 6", got.Chapter)
	}
}

func TestUpdateChapterRejectsNegative(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := t.Context()

	if _, err := service.Register(ctx, "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err := service.UpdateChapter(ctx, "server-1", -1)
	if !errors.Is(err, ErrNegativeChapter) {
		t.Fatalf("error = %v, want %v", err, ErrNegativeChapter)
	}

	stored, err := store.CurrentBook(ctx, "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored.Chapter != 0 {
		t.Fatalf("chapter = %d, want 0 left unchanged", stored.Chapter)
	}
}

func TestUpdateChapterWithoutCurrentBook(t *testing.T) {
	service := NewService(NewMemoryStore())

	_, err := service.UpdateChapter(t.Context(), "server-1", 3)
	if !errors.Is(err, ErrNoCurrentBook) {
		t.Fatalf("error = %v, want %v", err, ErrNoCurrentBook)
	}
}

func TestServersHaveIndependentCurrentBooks(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := t.Context()

	if _, err := service.Register(ctx, "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := store.CurrentBook(ctx, "server-2"); !errors.Is(err, ErrNoCurrentBook) {
		t.Fatalf("server-2 error = %v, want %v", err, ErrNoCurrentBook)
	}
}

var errStore = errors.New("store unavailable")

type failingStore struct{ err error }

func (s failingStore) CurrentBook(context.Context, string) (Book, error) {
	return Book{}, s.err
}

func (s failingStore) SaveCurrentBook(context.Context, string, Book) error {
	return s.err
}

func TestRegisterPropagatesStoreFailure(t *testing.T) {
	service := NewService(failingStore{err: errStore})

	_, err := service.Register(t.Context(), "server-1", "Dune", "")
	if !errors.Is(err, errStore) {
		t.Fatalf("error = %v, want %v", err, errStore)
	}
}

func TestUpdateChapterPropagatesStoreFailure(t *testing.T) {
	service := NewService(failingStore{err: errStore})

	_, err := service.UpdateChapter(t.Context(), "server-1", 1)
	if !errors.Is(err, errStore) {
		t.Fatalf("error = %v, want %v", err, errStore)
	}
}

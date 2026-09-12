package book

import (
	"context"
	"fmt"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// Register sets or replaces the server's Current Book, starting at chapter 0.
func (s *Service) Register(ctx context.Context, serverID, name, url string) (Book, error) {
	if name == "" {
		return Book{}, ErrNameRequired
	}

	book := Book{Name: name, Url: url}

	if err := s.store.SaveCurrentBook(ctx, serverID, book); err != nil {
		return Book{}, fmt.Errorf("save current book: %w", err)
	}

	return book, nil
}

// UpdateChapter changes the Current Book's chapter; it fails with
// ErrNoCurrentBook when the server has no Current Book.
func (s *Service) UpdateChapter(ctx context.Context, serverID string, chapter int) (Book, error) {
	if chapter < 0 {
		return Book{}, ErrNegativeChapter
	}

	book, err := s.store.CurrentBook(ctx, serverID)
	if err != nil {
		return Book{}, err
	}

	book.Chapter = chapter

	if err := s.store.SaveCurrentBook(ctx, serverID, book); err != nil {
		return Book{}, fmt.Errorf("save current book: %w", err)
	}

	return book, nil
}

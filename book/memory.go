package book

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	books map[string]Book
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		books: make(map[string]Book),
	}
}

func (s *MemoryStore) CurrentBook(_ context.Context, serverID string) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, ok := s.books[serverID]
	if !ok {
		return Book{}, ErrNoCurrentBook
	}

	return book, nil
}

func (s *MemoryStore) SaveCurrentBook(_ context.Context, serverID string, book Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.books[serverID] = book

	return nil
}

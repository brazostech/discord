package book

import "errors"

var (
	ErrNoCurrentBook   = errors.New("no current book")
	ErrNameRequired    = errors.New("book name is required")
	ErrNegativeChapter = errors.New("chapter must be non-negative")
)

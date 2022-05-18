package store

import (
	"context"
	"fmt"

	"github.com/haleyrc/bookshelf/library"
	"github.com/jmoiron/sqlx"
)

type LibraryStore struct {
	DB *sqlx.DB
}

func (s *LibraryStore) CreateBook(ctx context.Context, book *library.Book) error {
	q := `INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id;`

	err := s.DB.GetContext(ctx, &book.ID, q, book.Title, book.Author)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}

	return nil
}

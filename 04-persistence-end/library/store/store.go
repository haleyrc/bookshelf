package store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/haleyrc/bookshelf/library"
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

func (s *LibraryStore) GetBookByID(ctx context.Context, id int64) (*library.Book, error) {
	q := `SELECT id, title, author FROM books WHERE id = $1;`

	var book library.Book
	err := s.DB.QueryRowxContext(ctx, q, id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		return nil, fmt.Errorf("get book by id: %w", err)
	}

	return &book, nil
}

func (s *LibraryStore) GetBooks(ctx context.Context) ([]*library.Book, error) {
	q := `SELECT id, title, author FROM books ORDER BY id ASC;`

	rows, err := s.DB.QueryxContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("get books: %w", err)
	}
	defer rows.Close()

	books := []*library.Book{}
	for rows.Next() {
		var book library.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			return nil, fmt.Errorf("get books: %w", err)
		}
		books = append(books, &book)
	}

	return books, nil
}

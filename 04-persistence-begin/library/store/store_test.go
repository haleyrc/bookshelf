package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/haleyrc/bookshelf/internal/test"
	"github.com/haleyrc/bookshelf/library"
	"github.com/haleyrc/bookshelf/library/store"
)

func TestLibraryStore_CreateBook(t *testing.T) {
	ctx := context.Background()
	db := sqlx.MustConnect("postgres", "postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable")
	ls := store.LibraryStore{DB: db}

	defer db.Close()

	book := library.Book{
		Title:  "The Lean Startup",
		Author: "Eric Ries",
	}
	if err := ls.CreateBook(ctx, &book); err != nil {
		t.Fatal("unexpected error:", err)
	}
	test.MustCleanup(t, func() error {
		return deleteBook(ctx, t, ls.DB, book.ID)
	})

	if book.ID == 0 {
		t.Errorf("expected id to not be blank, but it was")
	}
	if book.Title != "The Lean Startup" {
		t.Errorf("expected title to be \"The Lean Startup\" but got %q", book.Title)
	}
	if book.Author != "Eric Ries" {
		t.Errorf("expected author to be \"Eric Ries\" but got %q", book.Author)
	}
}

func deleteBook(ctx context.Context, t *testing.T, db *sqlx.DB, id int64) error {
	q := `DELETE FROM books WHERE id = $1;`
	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	return nil
}

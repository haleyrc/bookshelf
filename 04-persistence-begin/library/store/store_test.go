package store_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/haleyrc/bookshelf/internal/test"
	"github.com/haleyrc/bookshelf/library"
	"github.com/haleyrc/bookshelf/library/store"
)

var ls store.LibraryStore

func TestMain(m *testing.M) {
	path := filepath.Join("..", "..", ".env")
	godotenv.Load(path)

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		fmt.Println("set the TEST_DATABASE_URL environment variable to run this test suite")
		os.Exit(0)
	}
	ls.DB = sqlx.MustConnect("postgres", url)

	code := m.Run()

	ls.DB.Close()
	os.Exit(code)
}

func TestLibraryStore_CreateBook(t *testing.T) {
	ctx := context.Background()

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

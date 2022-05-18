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

func TestLibraryStore_GetBookByID(t *testing.T) {
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

	gotBook, err := ls.GetBookByID(ctx, book.ID)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if gotBook.ID != book.ID {
		t.Errorf("expected id to be %d but got %d", book.ID, gotBook.ID)
	}
	if gotBook.Title != "The Lean Startup" {
		t.Errorf("expected title to be \"The Lean Startup\" but got %q", gotBook.Title)
	}
	if gotBook.Author != "Eric Ries" {
		t.Errorf("expected author to be \"Eric Ries\" but got %q", gotBook.Author)
	}
}

func TestLibraryStore_GetBooks(t *testing.T) {
	ctx := context.Background()

	params := [][]string{
		{"Norse Mythology", "Neil Gaiman"},
		{"The Divine Comedy", "Dante Alighieri"},
		{"2001: A Space Odyssey", "Arthur C. Clarke"},
	}
	ids, err := createManyBooks(ctx, t, ls, params)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	books, err := ls.GetBooks(ctx)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	for idx, book := range books {
		id := ids[idx]
		title := params[idx][0]
		author := params[idx][1]

		if book.ID != id {
			t.Errorf("expected book %d id to be %d, but got %d", idx, id, book.ID)
		}
		if book.Title != title {
			t.Errorf("expected book %d title to be %q, but got %q", idx, title, book.Title)
		}
		if book.Author != author {
			t.Errorf("expected book %d author to be %q, but got %q", idx, author, book.Author)
		}
	}
}

func createManyBooks(ctx context.Context, t *testing.T, ls store.LibraryStore, params [][]string) ([]int64, error) {
	ids := []int64{}
	for _, p := range params {
		b := library.Book{Title: p[0], Author: p[1]}
		if err := ls.CreateBook(ctx, &b); err != nil {
			return nil, err
		}
		test.MustCleanup(t, func() error {
			return deleteBook(ctx, t, ls.DB, b.ID)
		})
		ids = append(ids, b.ID)
	}
	return ids, nil
}

func deleteBook(ctx context.Context, t *testing.T, db *sqlx.DB, id int64) error {
	q := `DELETE FROM books WHERE id = $1;`
	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	return nil
}

package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/haleyrc/bookshelf/internal/test"
	"github.com/haleyrc/bookshelf/library/store"
)

var ls store.LibraryStore

func TestMain(m *testing.M) {
	ls.DB = sqlx.MustConnect("postgres", "postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable")
	defer ls.DB.Close()

	code := m.Run()

	os.Exit(code)
}

func TestLibraryStore_CreateBook(t *testing.T) {
	ctx := context.Background()

	createdBook, err := ls.CreateBook(ctx, "The Lean Startup", "Eric Ries")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	test.MustCleanup(t, func() error {
		return ls.DeleteBook(ctx, createdBook.ID)
	})

	if createdBook.ID == 0 {
		t.Errorf("expected id to not be blank, but it was")
	}
	if createdBook.Title != "The Lean Startup" {
		t.Errorf("expected title to be \"The Lean Startup\" but got %q", createdBook.Title)
	}
	if createdBook.Author != "Eric Ries" {
		t.Errorf("expected author to be \"Eric Ries\" but got %q", createdBook.Author)
	}
}

func TestLibraryStore_DeleteBook(t *testing.T) {
	ctx := context.Background()

	createdBook, err := ls.CreateBook(ctx, "The Lean Startup", "Eric Ries")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if err := ls.DeleteBook(ctx, createdBook.ID); err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestLibraryStore_GetBookByID(t *testing.T) {
	ctx := context.Background()

	createdBook, err := ls.CreateBook(ctx, "The Lean Startup", "Eric Ries")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	test.MustCleanup(t, func() error {
		return ls.DeleteBook(ctx, createdBook.ID)
	})

	gotBook, err := ls.GetBookByID(ctx, createdBook.ID)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if gotBook.ID != createdBook.ID {
		t.Errorf("expected id to be %d but got %d", createdBook.ID, gotBook.ID)
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
		b, err := ls.CreateBook(ctx, p[0], p[1])
		if err != nil {
			return nil, err
		}
		test.MustCleanup(t, func() error {
			return ls.DeleteBook(ctx, b.ID)
		})
		ids = append(ids, b.ID)
	}
	return ids, nil
}

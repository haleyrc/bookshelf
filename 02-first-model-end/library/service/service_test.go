package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/haleyrc/bookshelf/library"
	"github.com/haleyrc/bookshelf/library/service"
)

var ErrMockNotImplemented = fmt.Errorf("mock not implemented")

func TestLibraryService_AddBook(t *testing.T) {
	ctx := context.Background()
	title := "Dune"
	author := "Frank Herbert"

	store := &mockStore{
		CreateBookFn: func(_ context.Context, title, author string) (*library.Book, error) {
			return &library.Book{ID: 123, Title: title, Author: author}, nil
		},
	}
	svc := service.LibraryService{Store: store}

	req := service.AddBookRequest{Title: title, Author: author}
	resp, err := svc.AddBook(ctx, req)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(store.CreateBookCalledWith) != 1 {
		t.Fatalf("expected CreateBook to be called once, but was called %d times", len(store.CreateBookCalledWith))
	}

	createBookArgs := store.CreateBookCalledWith[0]
	if createBookArgs.Title != title {
		t.Errorf("expected CreateBook to be called with title %q, but got %q", title, createBookArgs.Title)
	}
	if createBookArgs.Author != author {
		t.Errorf("expected CreateBook to be called with author %q, but got %q", author, createBookArgs.Author)
	}

	if resp.Book == nil {
		t.Fatal("expected book to not be nil, but it was")
	}
	if resp.Book.ID == 0 {
		t.Errorf("expected response to include an id, but it was blank")
	}
	if resp.Book.Title != title {
		t.Errorf("expected response to have title %q, but got %q", title, resp.Book.Title)
	}
	if resp.Book.Author != author {
		t.Errorf("expected response to have author %q, but got %q", author, resp.Book.Author)
	}
}

type mockStore struct {
	CreateBookCalledWith []createBookArgs
	CreateBookFn         func(ctx context.Context, title, author string) (*library.Book, error)
}

type createBookArgs struct {
	Title  string
	Author string
}

func (ms *mockStore) CreateBook(ctx context.Context, title, author string) (*library.Book, error) {
	if ms.CreateBookFn == nil {
		return nil, ErrMockNotImplemented
	}
	ms.CreateBookCalledWith = append(ms.CreateBookCalledWith, createBookArgs{Title: title, Author: author})
	return ms.CreateBookFn(ctx, title, author)
}

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
	testcases := map[string]struct {
		Request   service.AddBookRequest
		ShouldErr bool
	}{
		"Empty Request": {
			Request:   service.AddBookRequest{},
			ShouldErr: true,
		},
		"Blank Title": {
			Request:   service.AddBookRequest{Title: "", Author: "Frank Herbert"},
			ShouldErr: true,
		},
		"Blank Author": {
			Request:   service.AddBookRequest{Title: "Dune", Author: ""},
			ShouldErr: true,
		},
		"Happy Path": {
			Request:   service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"},
			ShouldErr: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			store := &mockStore{
				CreateBookFn: func(_ context.Context, title, author string) (*library.Book, error) {
					return &library.Book{ID: 123, Title: title, Author: author}, nil
				},
			}
			svc := service.LibraryService{Store: store}

			resp, err := svc.AddBook(ctx, tc.Request)
			if tc.ShouldErr {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			if len(store.CreateBookCalledWith) != 1 {
				t.Fatalf("expected CreateBook to be called once, but was called %d times", len(store.CreateBookCalledWith))
			}

			args := store.CreateBookCalledWith[0]
			if args.Title != tc.Request.Title {
				t.Errorf("expected CreateBook to be called with title %q, but got %q", tc.Request.Title, args.Title)
			}
			if args.Author != tc.Request.Author {
				t.Errorf("expected CreateBook to be called with author %q, but got %q", tc.Request.Author, args.Author)
			}

			if resp.Book == nil {
				t.Fatal("expected book to not be nil, but it was")
			}
			if resp.Book.ID != 123 {
				t.Errorf("expected book to have id 123, but got %d", resp.Book.ID)
			}
			if resp.Book.Title != tc.Request.Title {
				t.Errorf("expected book to have title %q, but got %q", tc.Request.Title, resp.Book.Title)
			}
			if resp.Book.Author != tc.Request.Author {
				t.Errorf("expected book to have author %q, but got %q", tc.Request.Author, resp.Book.Author)
			}
		})
	}
}

func TestLibraryService_GetBook(t *testing.T) {
	testcases := map[string]struct {
		Request   service.GetBookRequest
		ShouldErr bool
	}{
		"Empty Request": {
			Request:   service.GetBookRequest{},
			ShouldErr: true,
		},
		"Happy Path": {
			Request:   service.GetBookRequest{ID: 123},
			ShouldErr: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			store := &mockStore{
				GetBookFn: func(_ context.Context, id int64) (*library.Book, error) {
					return &library.Book{ID: id, Title: "Dune", Author: "Frank Herbert"}, nil
				},
			}
			svc := service.LibraryService{Store: store}

			resp, err := svc.GetBook(ctx, tc.Request)
			if tc.ShouldErr {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			if len(store.GetBookCalledWith) != 1 {
				t.Fatalf("expected GetBook to be called once, but was called %d times", len(store.GetBookCalledWith))
			}

			args := store.GetBookCalledWith[0]
			if args.ID != tc.Request.ID {
				t.Errorf("expected GetBook to be called with id %d, but got %q", tc.Request.ID, args.ID)
			}

			if resp.Book == nil {
				t.Fatal("expected book to not be nil, but it was")
			}
			if resp.Book.ID != tc.Request.ID {
				t.Errorf("expected book to have id %d, but got %d", tc.Request.ID, resp.Book.ID)
			}
			if resp.Book.Title != "Dune" {
				t.Errorf("expected book to have title \"Dune\", but got %q", resp.Book.Title)
			}
			if resp.Book.Author != "Frank Herbert" {
				t.Errorf("expected book to have author \"Frank Herbert\", but got %q", resp.Book.Author)
			}

		})
	}
}

func TestLibraryService_GetBooks(t *testing.T) {
	ctx := context.Background()
	store := &mockStore{
		GetBooksFn: func(_ context.Context) ([]*library.Book, error) {
			return []*library.Book{
				{ID: 123, Title: "Dune", Author: "Frank Herbert"},
				{ID: 456, Title: "The Silmarillion", Author: "J.R.R. Tolkien"},
			}, nil
		},
	}
	svc := service.LibraryService{Store: store}

	resp, err := svc.GetBooks(ctx, service.GetBooksRequest{})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(store.GetBooksCalledWith) != 1 {
		t.Fatalf("expected GetBooks to be called once, but was called %d times", len(store.GetBookCalledWith))
	}

	if resp.Books == nil {
		t.Fatal("expected book to not be nil, but it was")
	}
	if len(resp.Books) != 2 {
		t.Fatalf("expected to get 2 books, but got %d", len(resp.Books))
	}

	book1 := resp.Books[0]
	if book1.ID != 123 {
		t.Errorf("expected book to have id 123, but got %d", book1.ID)
	}
	if book1.Title != "Dune" {
		t.Errorf("expected book to have title \"Dune\", but got %q", book1.Title)
	}
	if book1.Author != "Frank Herbert" {
		t.Errorf("expected book to have author \"Frank Herbert\", but got %q", book1.Author)
	}

	book2 := resp.Books[1]
	if book2.ID != 456 {
		t.Errorf("expected book to have id 456, but got %d", book2.ID)
	}
	if book2.Title != "The Silmarillion" {
		t.Errorf("expected book to have title \"The Silmarillion\", but got %q", book2.Title)
	}
	if book2.Author != "J.R.R. Tolkien" {
		t.Errorf("expected book to have author \"J.R.R. Tolkien\", but got %q", book2.Author)
	}
}

type mockStore struct {
	CreateBookCalledWith []struct {
		Ctx    context.Context
		Title  string
		Author string
	}
	CreateBookFn func(ctx context.Context, title, author string) (*library.Book, error)

	GetBookCalledWith []struct {
		Ctx context.Context
		ID  int64
	}
	GetBookFn func(ctx context.Context, id int64) (*library.Book, error)

	GetBooksCalledWith []struct {
		Ctx context.Context
	}
	GetBooksFn func(ctx context.Context) ([]*library.Book, error)
}

func (ms *mockStore) CreateBook(ctx context.Context, title, author string) (*library.Book, error) {
	if ms.CreateBookFn == nil {
		return nil, ErrMockNotImplemented
	}
	ms.CreateBookCalledWith = append(ms.CreateBookCalledWith, struct {
		Ctx    context.Context
		Title  string
		Author string
	}{
		Ctx:    ctx,
		Title:  title,
		Author: author,
	})
	return ms.CreateBookFn(ctx, title, author)
}

func (ms *mockStore) GetBookByID(ctx context.Context, id int64) (*library.Book, error) {
	if ms.GetBookFn == nil {
		return nil, ErrMockNotImplemented
	}
	ms.GetBookCalledWith = append(ms.GetBookCalledWith, struct {
		Ctx context.Context
		ID  int64
	}{
		Ctx: ctx,
		ID:  id,
	})
	return ms.GetBookFn(ctx, id)
}

func (ms *mockStore) GetBooks(ctx context.Context) ([]*library.Book, error) {
	if ms.GetBooksFn == nil {
		return nil, ErrMockNotImplemented
	}
	ms.GetBooksCalledWith = append(ms.GetBooksCalledWith, struct {
		Ctx context.Context
	}{Ctx: ctx})
	return ms.GetBooksFn(ctx)
}

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

type mockStore struct {
	CreateBookCalledWith []struct {
		Ctx    context.Context
		Title  string
		Author string
	}
	CreateBookFn func(ctx context.Context, title, author string) (*library.Book, error)
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
	return nil, fmt.Errorf("TODO")
}

func (ms *mockStore) GetBooks(ctx context.Context) ([]*library.Book, error) {
	return nil, fmt.Errorf("TODO")
}

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/haleyrc/bookshelf/library"
)

type LibraryStore interface {
	CreateBook(ctx context.Context, title, author string) (*library.Book, error)
	GetBookByID(ctx context.Context, id int64) (*library.Book, error)
	GetBooks(ctx context.Context) ([]*library.Book, error)
}

type LibraryService struct {
	Store LibraryStore
}

type AddBookRequest struct {
	Title  string
	Author string
}

type AddBookResponse struct {
	Book *library.Book
}

func (ls *LibraryService) AddBook(ctx context.Context, req AddBookRequest) (*AddBookResponse, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("add book: title can't be blank")
	}
	if req.Author == "" {
		return nil, fmt.Errorf("add book: author can't be blank")
	}

	book, err := ls.Store.CreateBook(ctx, req.Title, req.Author)
	if err != nil {
		return nil, fmt.Errorf("add book: %w", err)
	}
	log.Printf("added book %d: %s\n", book.ID, book.Title)

	return &AddBookResponse{Book: book}, nil
}

type GetBookRequest struct {
	ID int64
}

type GetBookResponse struct {
	Book *library.Book
}

func (ls *LibraryService) GetBook(ctx context.Context, req GetBookRequest) (*GetBookResponse, error) {
	if req.ID == 0 {
		return nil, fmt.Errorf("get book: id is required")
	}

	book, err := ls.Store.GetBookByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	return &GetBookResponse{Book: book}, nil
}

type GetBooksRequest struct{}

type GetBooksResponse struct {
	Books []*library.Book
}

func (ls *LibraryService) GetBooks(ctx context.Context, req GetBooksRequest) (*GetBooksResponse, error) {
	books, err := ls.Store.GetBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get books: %w", err)
	}

	return &GetBooksResponse{Books: books}, nil
}

type RateBookRequest struct{}
type RateBookResponse struct{}

func (ls *LibraryService) RateBook(ctx context.Context, req RateBookRequest) (*RateBookResponse, error) {
	return nil, fmt.Errorf("TODO")
}

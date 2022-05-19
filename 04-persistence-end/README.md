# Persistence - Solution

> Full code:
>
> - [`store_test.go`](./library/store/store_test.go)
> - [`store.go`](./library/store/store.go)

This assignment required a bit of critical thinking when it comes to both
interacting with the database and writing the tests for that behavior. What made
this challenging was the shift from dealing with single model instances to
dealing with collections. On the plus side, once you've internalized the
patterns that you like to use, this divide becomes much easier to cross. So
let's jump right in and see how I approached the task.

## Getting a book

The implementation of `GetBookByID` is fairly simple especially if you were able
to follow the implementation of `CreateBook`. The only substantial difference
between the two is that in this case, we already have an ID and are returning
a book:

```go
func (s *LibraryStore) GetBookByID(ctx context.Context, id int64) (*library.Book, error) {
	q := `SELECT id, title, author FROM books WHERE id = $1;`

	var book library.Book
	err := s.DB.QueryRowxContext(ctx, q, id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		return nil, fmt.Errorf("get book by id: %w", err)
	}

	return &book, nil
}
```

One thing to note here is that we aren't making any distinction between the
various reasons that this method can return an error. When we talk about error
handling later we'll see how this can start to take even simple methods and
make them much more interesting. Now let's see how I tested this:

```go
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
```

As with the implementation, this should look fairly familiar. One notable item
is that we had to first create a book in order to them retrieve it from the
database. One trap that people fall into at this point is making assertions
on the created book, but what we really want to test is our getter method. Since
we already test the `CreateBook` method independently, we can assume for this
test that it functions as intended and move on.

## Getting a list of books

Unlike the single book case, retrieving multiple books from our database
requires a different approach:

```go
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
```

There are a couple of approaches that would work here, but I went with the
"rows" approach for my version. Apart from the need for a loop to build our list
of books, however, the underlying logic is nearly identical to the single book
case. Once again, let's look at how we test this:

```go
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
```

If you used the hint in the assignment, you might have ended up with something
similar to what I wrote, but if you attempted it on your own that's great too.
As with everything, there are infinite ways to approach this problem. What's
important isn't the exact details of how you wrote your test, but rather the
principles you used to guide your implementation. As with the previous case,
we want to focus only on asserting against the behavior of our function under
test and ignore the rest. Here though, I would also caution you to ensure that
any details of the setup process not salient to the test itself are pulled out
of the test itself. This was the guiding principle behind the `createManyBooks`
test helper. By moving this functionality into a helper, our test code is short
enough to take in at a glance.

## Next

Once you feel confident that you understand what's happening in this lesson, you
should be ready to move on to lesson 5: [TODO](#).

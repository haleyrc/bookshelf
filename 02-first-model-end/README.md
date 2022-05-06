# First model - Solution

Hopefully you didn't find the assignment too difficult, but in case you
struggled or just for a different perspective, I'll walk through how I
implemented the "get a book" and "get a list of books" actions.

## Getting a book

First let's take a look at the code:

```go
type GetBookRequest struct {
	ID int64
}

type GetBookResponse struct {
	Book *library.Book
}

func (ls *LibraryService) GetBook(ctx context.Context, req GetBookRequest) (*GetBookResponse, error) {
	book, err := ls.Store.GetBookByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	return &GetBookResponse{Book: book}, nil
}
```

Here you can see that we're accepting an integer ID and returning a single book.
There are a couple different schools of thought when it comes to querying for
objects.

Some people prefer a single method that always returns an array and
it's up to the client to handle it however they see fit. In my opinion, this is
sub-optimal because it exposes your database to your user. What I mean by that
is that it doesn't really reflect a real use-case, but is more of a convenience
for the backend developer who only thinks in terms of a result set. Whenever
possible, however, I prefer to shoulder the burden of adding additional
complexity to give my upstream developers a more practical API.

I've also seen a number of APIs that take multiple different query parameters
when retrieving even a single entity. I find that this rarely works well due to
the fact that most database columns are not unique so you quickly run into
decisions about how to handle the case where a query returns more than one
result where only one was expected. In addition, if you think through the
frontend use-case, you are usually making a "get one" request for something like
a details page where you have selected one entity from a list and already know
its ID. In that case, why not keep our backend API simple while still matching
our usage pattern?

The only other thing of note in this method is that if you were following my
"add a book" example very closely, you may have added a log line where I did
not. This is largely down to taste, but in general I won't log success messages
in pure queries, but I will in mutations. There's an argument for keeping things
consistent, but there is rarely anything interesting to log when a client simply
fetches data.

> It's worth pointing out that this is not the final logging pattern we'll be
> using in this course, but I want to start thinking about _what_ to log as
> early as possible even if we don't quite have the _how_ figured out.

## Getting a list of books

The code for getting a list of books is largely the same as what you've already
seen:

```go
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
```

The most interesting thing to note here is that I haven't added any fields to
our request. This is perfectly acceptable. Not all of our requests will have
fields, but we still fall back on our service pattern because it allows us to
add fields later without breaking compatibility.

## Next

Once you feel comfortable with this lesson, you can move on to [Lesson 3: TODO](#).

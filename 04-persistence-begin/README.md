# Persistence

> Before we start with the lesson proper, I need to address a small change I
> made to some of the code from the previous lessons. If you'll remember, we
> had previously defined a `LibraryStore` interface in our service package that
> looked something like this:
>
> ```go
> type LibraryStore interface {
> 	CreateBook(ctx context.Context, title, author string) (*library.Book, error)
> 	GetBookByID(ctx context.Context, id int64) (*library.Book, error)
> 	GetBooks(ctx context.Context) ([]*library.Book, error)
> }
> ```
>
> There's nothing wrong with this setup and I wrote many database layers that
> followed this pattern. Recently, however, I have changed how I think about the
> parameters and return values of store methods and today I would prefer to
> write this as:
>
> ```diff
> type LibraryStore interface {
> -	CreateBook(ctx context.Context, title, author string) (*library.Book, error)
> +	CreateBook(ctx context.Context, book *library.Book) error
> 	GetBookByID(ctx context.Context, id int64) (*library.Book, error)
> 	GetBooks(ctx context.Context) ([]*library.Book, error)
> }
> ```
>
> With this pattern, we are passing a domain object to the persistence layer
> fully-formed, rather than relying on the persistence layer to return the
> object to us. In our small example, the differences are pretty minimal, but I
> find that this matches my mental model for how this operation should work much
> more closely.
>
> Future lessons will use the new pattern, but I'll be leaving the existing
> lessons as-is for now to prevent any confusion. Make sure you take a look at
> the full service implementaion, since our call to `CreateBook` also has to
> change to keep everything working.

Following from our previous lesson, we've got a decent service started and at
this point I think it's time to start talking about persisting our "models" to a
database. There are a lot of different databases to choose from and often the
choice has already been made for you, but assuming I have some measure of
control over infrastructure decisions, I will almost always reach for
PostgreSQL. Your own personal preferences may vary, but PostgreSQL offers a lot
of benefits for my use-case:

1. Widely supported
1. Open source
1. Enterprise ready
1. I already know it

It's probably obvious at this point that we'll be using PostgreSQL for this
course, so let's get started by writing our very first migration.

## Migrations

Since we started with our domain modeling, we already have a `Book` model that
we can save to our database, but we need to write the migration for it. To keep
things focused on the deployable bits for this course, I'm not going to get into
the tooling for authoring, running, or rolling back migrations. There are too
many variables that go into that decision for me to give hard-and-fast rules so
instead, we'll be writing our migrations by hand. This has the added benefit of
making it very apparent what we need to be doing to do zero-downtime deploys
later.

Let's start by looking at the migration itself and working backwards from there:

```sql
CREATE TABLE books (
    id         BIGSERIAL PRIMARY KEY,
    title      TEXT NOT NULL CHECK (length(title) > 0),
    author     TEXT NOT NULL CHECK (length(author) > 0),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE books IS 'User-submitted books that can be rated';
COMMENT ON COLUMN books.id IS 'A unique identifier for the book';
COMMENT ON COLUMN books.title IS 'The title of the book';
COMMENT ON COLUMN books.author IS 'The author of the book';
COMMENT ON COLUMN books.created_at IS 'The time the book was added to the database';
```

If you've done any SQL before, there shouldn't be a whole lot of surprises here.
You can see the three fields from our domain model showing up as database
columns: `id`, `title`, and `author`.

You'll also notice that I've added a `created_at` column that does not appear on
our domain model. This is a fairly standard "default" column to add to any
database table, but it serves an additional purpose for this course. Namely, it
illustrates that we aren't required to expose all of our database internals to
our domain. In fact, I make it a point whenever I'm crossing package boundaries
to consider whether I'm accurately modeling my domain or if I'm being lazy and
simply exposing my database as-is to my clients. In this case, we haven't
outlined a domain reason for having a timestamp here, so we don't need to bubble
that up. When we start expanding our domain models you'll see how this applies
to things like value objects, cross-context models, and more to keep our domain
clean regardless of the underlying infrastructure.

There are a couple of other decisions worth calling out here. First is the use
of a serial key for our `id` column. In general, I prefer using UUIDs for this
purpose since they aren't subject to wraparound issues, they're easier to work
with in distributed systems, and they can even be used to enable optimistic UI
in a really bulletproof way. I decided against using them for this course,
however, since they introduce some additional complexity without really adding
anything in terms of teachability. Realistically, this is one of those decisions
that may be influenced by other factors so there isn't one right answer.

Next, I'll note quickly that the choice of database-level constraints on the
`title` and `author` columns is a bit of a personal preference, but I will
usually add these if and only if the constraint is so basic to our model that
I'm comfortable making a decision up-front for the lifetime of the app. If you
are at all uncertain, I would hedge and omit them in your migration, preferring
instead to validate at the domain layer. This almost always results in the same
outcome, but is vastly simpler to change course on later.

Finally, I will almost always add database comments to my tables and columns. I
find that the extra lift is usually fairly small and the dividends it pays when
you're troubleshooting or onboarding a new engineer is well worth it. As we get
further in the course, you'll see that I don't use many other database features
(e.g. triggers, stored procedures, etc.) so I'm really only commenting on the
data.

> Since database comments don't seem to be all that popular, I'll take a short
> aside to explain how they work. Basically, comments can be added to almost
> everything you can create in PostgreSQL. By default, these won't show up when
> describing a table in `psql`:
>
> ```sql
> bookshelf=# \dt
>          List of relations
>  Schema | Name  | Type  |  Owner
> --------+-------+-------+----------
>  public | books | table | postgres
> (1 row)
> ```
>
> If you use the extended version of the command, however, they will appear alongside everything else making it much easier to get a feel for the data without having to search through the code:
>
> ```sql
> bookshelf=# \dt+
>          List of relations
>  Schema | Name  | Type  |  Owner   |              Description
> --------+-------+-------+----------+----------------------------------------
>  public | books | table | postgres | User-submitted books that can be rated
> (1 row)
> ```
>
> Many tools will also have a place to display these which makes them even more
> valuable in my opinion.

Now we've got a database migration, but nowhere to run it. Enter the tool
everyone loves to hate: Docker.

## Docker

Let me preface this section by saying that I have a love-hate relationship with
Docker. On one hand, it enables a lot of stuff that would otherwise be a huge
chore during local development. On the other hand, it never quite works as
intended right out of the box. That said, for the purposes of this course it
means I only have to rely on one local dependency. It's also worth looking at
early since I can pretty much guarantee that in 2022, some portion of your stack
is using Docker, even if you don't know it.

If you look at the root of the project, you'll see we already have a
`Dockerfile` as well as a bare-bones `docker-compose.yml` file. You can ignore
the `Dockerfile` for now, we'll have more to say on that when we actually start
up a server. Just for shits and giggles though, you can go ahead and run:

```
$ docker-compose up server
```

This should build the hello world app we've got and after the usual ten pages of
output, you should see something like the following:

```
Creating 04-persistence-end_server_1 ... done
Attaching to 04-persistence-end_server_1
server_1    | Hello, world!
04-persistence-end_server_1 exited with code 0
```

This is great since it means we've got Docker working already, but it's not all
that useful. What we really need is a database. So let's add a new service to
our `docker-compose.yml` file just below the `server` section:

```yaml
postgres:
  image: postgres:14.2
  restart: always
  environment:
    - POSTGRES_PASSWORD=password
    - POSTGRES_DB=bookshelf
  volumes:
    - ./migrations:/docker-entrypoint-initdb.d
  ports:
    - "5555:5432"
```

There's a bit going on here and this isn't intended to be a Docker tutorial, but
the gist of it is that when we start this service, we'll have a PostgreSQL
instance available for use by our application. For now, we'll be connecting to
it externally, so we open it up on host port `:5555`.

The most immediately interesting part of this configuration is the `volumes`
section. Here we're specifying that everything in the `./migrations` directory
should be mounted into the container at this weird path:
`/docker-entrypoint-initdb.d`. The `postgres` Docker image will look for any
SQL files in this location and run them at start up, basically giving us
immediate migrations for free when we're developing locally.

Since seeing is believing, we can go ahead and test this for ourselves. We
already have a migration, so let's start up our `postgres` service and see what
happens:

```
$ docker-compose up postgres
```

Once again you'll get a bunch of Docker gibberish, but buried in that output you
should see lines like this:

```
postgres_1  | /usr/local/bin/docker-entrypoint.sh: running /docker-entrypoint-initdb.d/20220406.0-add-book-table.sql
postgres_1  | CREATE TABLE
postgres_1  | COMMENT
postgres_1  | COMMENT
postgres_1  | COMMENT
postgres_1  | COMMENT
postgres_1  | COMMENT
```

This is exactly what we want and if you were to connect to this database you
could see our table sitting there empty.

> Note that right now we only have a single file our migrations directory, so
> order isn't really an issue. As we continue though, we'll be adding more and
> more migrations that need to run in a specific order. This is the reason for
> our somewhat cumbersome naming scheme for migrations. By default, they will
> appear and be run in numerical order. Since we use timestamps as our prefix,
> this is exactly the behavior we want.

With a database to start throwing bytes at, it's time to figure out how we get
connected. The first step in that process is getting the right packages.

## Dependencies

Here we're faced with another hot-button topic: how to connect to a database. In
the Go ecosystem as with most other languages, there exists a huge spectrum of
options for connecting to and interacting with a database. At the lowest layers,
there is the standard library's `sql` package that most of the others are built
from. This is far too low-level for general use though.

On the other end of the spectrum are the ORMs. While widely used in other
languages and frameworks, ORMs are often frowned upon in the Go community as
they tend to run counter to the Go-ish way of doing things. I tend to fall into
this camp as well, having been bitten by a number of object-relational impedance
mismatch issues in the past.

Even in the middle grounds there are a fair number of options with the most
popular being the `sqlx` and `pgx` packages. (Note that while `sqlx` is database
agnostic, `pgx` is PostgreSQL-specific). Both of these packages are great at
their jobs, but I have far more experience with `sqlx` so that's what we'll be
using here. Since we still need a PostgreSQL driver, we'll also be using the
`pq` package.

We can add both of these to our project with the following command:

```
$ go get github.com/lib/pq github.com/jmoiron/sqlx
```

Now we're all set to start connecting to our database and to do that we'll start
by writing a test.

## Testing at the database

Alright, so we've got a database with our new table all set up and we installed
our dependencies, so the next step is to make sure we can actually connect to
the database. We'll start with a test for a (so far) fictional `CreateBook`
method and to get going, all we want to do is verify that we can connect without
error. To do that, we create a new file at `library/store/store_test.go` and add
the following:

```go
package store_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestLibraryStore_CreateBook(t *testing.T) {
	db := sqlx.MustConnect("postgres", "postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable")
	defer db.Close()
}
```

This is really ugly but we'll clean it up in a bit. All that matters is that
we're making a connection using the parameters for our Docker-based PostgreSQL
instance. Make note of the anonymous import. If you omit this line, the test
will run but will fail with a runtime error because we haven't registered our
driver.

If you run this test, you'll get one of two outcomes depending on whether you've
already started up your Docker service. Either the test will pass or you'll see
a message like the following:

```
=== RUN   TestLibraryStore_CreateBook
--- FAIL: TestLibraryStore_CreateBook (0.01s)
panic: dial tcp [::1]:5555: connect: connection refused [recovered]
	panic: dial tcp [::1]:5555: connect: connection refused

...

FAIL	github.com/username/bookshelf/library/store	0.516s
```

If you see this error, the solution is simple. You can either run your Docker
service in the foreground or more commonly I will start it in daemon mode:

```
$ docker-compose up -d postgres
```

> You will have to wait a couple of seconds either way for the container to
> start and the migrations to run. If you're running in daemon mode, this can be
> easy to forget and can cause some headaches when you try to run your code. We'll
> deal with this in a later lesson, but for now, if your test fails try running it
> again once the database is all prepped and ready.

If we run our test now, we should get all green which is useful, but not very
interesting. To get something to work with, I'll go ahead and give you the full
test code and then I'll walk through it before we move on to making it pass:

```go
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
```

This test code is making a lot of assumptions about the method we're going to
write, but there shouldn't be too many surprises. We're connecting to our
database, creating a `LibraryStore` instance, passing a book to our method, and
then asserting on the values. There are really only three things to point out.

First, I've made a bit of a leap creating a store object that we haven't written
yet, but if you've been following along with the patterns so far, there's
nothing really new here. We know we need a _something_ and that _something_
probably needs a connection to the database. If the actual structure of our
_something_ ends up needing to change, that's easy enough to do.

The next thing of note is that we're asserting that our ID has been set. You may
remember from the previous lesson that setting IDs was something we considered a
database concern. This is where we actually test that behavior. Since we're
using a serial ID for our book model, we expect that we get a new unique ID any
time we save a new row to the database.

Finally, we're testing the contents of the title and author fields, even though
we set those manually in the lines above. This may seem redundant and the case
could certainly be made that these tests are extraneous, however, I prefer my
persistence layer to return data unmodified so I include them here as a sort of
contract test. If the fields come back blank, or in all caps, or reversed then
we've got something going on in our store that we need to address.

Now that we've got a failing test, let's make it pass and the first thing we
need to address is the fact that we don't even have a `store` package. To
remedy that, we can create a new file `library/store/store.go` and add the
following:

```go
package store

import "github.com/jmoiron/sqlx"

type LibraryStore struct {
	DB *sqlx.DB
}
```

This gets us past our first couple errors, but we still need a `CreateBook`
method. There are a number of ways to go about this, but here's how I would
write this:

```go
func (s *LibraryStore) CreateBook(ctx context.Context, book *library.Book) error {
	q := `INSERT INTO books (title, author) VALUES ($1, $2) RETURNING id;`

	err := s.DB.GetContext(ctx, &book.ID, q, book.Title, book.Author)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}

	return nil
}
```

The only real trick here is that we're `RETURNING` the ID that the database
creates and immediately assigning to the ID field of our book. At this point,
our test should be passing. In fact, if you connect to your database and do a
`SELECT * FROM books;` you'll see the book we created for our test. Obviously,
we need a way to clean up after ourselves, but this is pretty good for only a
few lines of real code.

Since we're green, the next step in the TDD cycle is refactoring, but before we
do that, let's make sure that we can clean up our database appropriately. To do
that, I'm going to introduce a couple quick test helpers. The first is a quick
and dirty function in our test package to do the actual deleting. We could make
this a method on our `LibraryStore`, but we don't have a use-case for deleting a
book right now and if we add it there, we then have to maintain it indefinitely.
Instead, we'll just drop it in our test and move on. At the bottom of
`store_test.go` then, add:

```go
func deleteBook(ctx context.Context, t *testing.T, db *sqlx.DB, id int64) error {
	q := `DELETE FROM books WHERE id = $1;`
	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	return nil
}
```

We could just run this in a defer in our test, but we want to make sure it gets
run even if our test panics, so we'll use the new(-ish) `t.Cleanup` method. One
issue with this method, however, is that I'd really like to know if our cleanup
fails, but I don't want to write the same `if err != nil { t.Log(err) }` line
everywhere. So let's write another test helper to wrap that behavior. Unlike the
`deleteBook` function, we're going to put this in a new package since it's one
of those things with a good use-case in many places.

I'll start by creating a package for various testing bits and bobs. I could dump
it in our `common` package, but as above, I don't really want this to become a
public API that I have to maintain. So I'll create a new `internal/` directory
and nest our `test/` directory under that. So in `internal/test/test.go` I'll
add the following:

```go
func MustCleanup(t *testing.T, f func() error) {
	t.Cleanup(func() {
		if err := f(); err != nil {
			t.Log("failed to clean up after test:", err)
		}
	})
}
```

Perhaps not the prettiest thing, but it does the job and now we can add the
following lines right below where we create our book:

```go
test.MustCleanup(t, func() error {
    return deleteBook(ctx, t, ls.DB, book.ID)
})
```

Now if we run our test again, we shouldn't have any new rows in the database at
the end...or rather...it should work that way. If we actually run it, we'll get
an error: `failed to clean up after test: delete book: sql: database is closed`.
This happens because our `defer` is running when the test function returns, but
_before_ the test suite ends and cleanup functions are executed. We can fix this
by finally getting to that refactoring.

## Refactor

At this point, I want to stop and take some time to get our tests into a cleaner
state. This will not only help fix the issue we ran into above, it will give us
a way to do a portion of our test setup once and then re-use that setup across
each of our tests functions. To achieve this, we're going to make use of Go's
`TestMain` function and a global variable.

> It's important to note here that test code is not compiled into production
> binaries, to the use of a global here doesn't dirty up our namespace or
> anything like that. Plus, sometimes you just need a global.

To get started, I'll add the following to the top of my test file:

```go
var ls store.LibraryStore

func TestMain(m *testing.M) {
	ls.DB = sqlx.MustConnect("postgres", "postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable")

	code := m.Run()

	ls.DB.Close()
	os.Exit(code)
}
```

This follows the usual `TestMain` layout so no surprises there. What's important
is that we're setting up a single database connection and a single instance of
our store. You can now remove the related lines from our
`TestLibraryStore_CreateBook` function:

```diff
-db := sqlx.MustConnect("postgres", "postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable")
-ls := store.LibraryStore{DB: db}
-
-defer db.Close()
```

If you run our test again, you should see that we're back to passing and if you
inspect the database, you won't see a new row added. So what's going on here?
We're creating our database connection before any tests run in `TestMain`, then
we're running our test function that performs some cleanup when it exits
(specifically, deleting our test book), and _then_ we're closing our database
connection. By moving the code to close our connection to `TestMain`, we're
ensuring that all of our cleanup happens as it should. Note that we actually
removed the defer entirely, since otherwise `os.Exit` would prevent the database
from being closed. This probably isn't a huge deal since tests don't run
indefinitely, but it's good practice anyway.

The final refactoring to do is to clean up how we're dealing with our database
connection string. At the same time, we can put in a safeguard that will prevent
these tests from being run if there's no database to connect to. To get started,
I'll create a `.env` file at the root of the project and move our connection
string into it:

```
export TEST_DATABASE_URL=postgres://postgres:password@localhost:5555/bookshelf?sslmode=disable
```

From here, we could do a `source .env` and update the test runner to look for
the new environment variable and everything should work, but that's pretty\
ungainly and in a CI environment, we expect our environment variable to already
be set. To make things easier, we'll pull in another dependency that will parse
a `.env` file for us and add it to the running context. To do that, run:

```
$ go get github.com/joho/godotenv
```

and then modify `TestMain` to look like this:

```go
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
```

The relative paths here are a bit of a bummer, but it's something I can usually
live with in a test setup. Now we're pulling in our environment variable and
using that to prevent a test run if we don't have a database to point to.

Now we've got ourselves setup to write as many tests as we need without having
to manage a database connection every time.

## Assignment

The assignment for this lesson is quite a bit more challenging than we've seen
previously, because it really demands that you think critically about how you
are interacting with the database and how you would test that interaction. The
assignment itself is to implement the `GetBookByID` and `GetBooks` methods on
the store as well as the tests.

Since this assignment is already challenging, I've included a hint in the form
of another test helper that I wrote for my version. Feel free to ignore it and
try to tackle it on your own, but if you get stuck, hopefully this will help
frame your thinking around the testing side of things.

<details>
<summary>Hint</summary>

In order to test the `GetBooks` method, you need to be able to populate a number
of test rows and clean them up successfully. To help with this, I added the
following function to my `store_test.go` file:

```go
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
```

You can use it similar to the following snippet:

```go
params := [][]string{
    {"Norse Mythology", "Neil Gaiman"},
    {"The Divine Comedy", "Dante Alighieri"},
    {"2001: A Space Odyssey", "Arthur C. Clarke"},
}
ids, err := createManyBooks(ctx, t, ls, params)
if err != nil {
    t.Fatal("unexpected error:", err)
}
```

With this, you should be able to make the necessary assertions to complete the
test function.

</details>

When you've completed the assignment, compare your approach to mine
[here](../04-persistence-end).

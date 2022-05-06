# First model

Now that we have the framework for our application, we can start filling in the
details. At this point, you can really go in a number of different directions
depending on what your goal is and who else you'll be working with. If you're
collaborating with a frontend team, it may make sense to start by solidifying
the API and working back to the database. On the other hand, if you have an
existing database or you just happen to be data-minded, the reverse path may
yield better results. For my money, I like to start at the service layer since
this is where most of the "interesting" modeling happens and that's where we'll
start for this lesson.

With that decided, we now need to decide where in our model to start. Luckily,
if we've broken our domain into truly independent contexts, it's largely
irrelevant which context we model first. I've made the executive decision to
start with the library side of things, but ultimately this is just another one
of those judgement calls that you'll end up making every time you're in front
of your editor.

## Model

Since we've already decided to start with our library context, it makes sense to
going to consider a very simple model with only three fields:

- ID
- Title
- Author

> Note that while we could have started with the assumption that authors are a
> separate model and that we'll need to normalize our database, we may never
> need that level of normalization so in the spirit of YAGNI we're taking the
> easier road here and just storing the author name as a string per book.
> Conveniently, this also gives us a scenario to explore live data migrations in
> a future lesson.

To get the ball rolling, I'll create a file `library/book.go` and populate it
with the following:

```go
package library

type Book struct {
	ID     int64
	Title  string
	Author string
}
```

We could have put this in the `service` package since that's where all of the
business logic will live, but remember that our `api` and `store` packages also
send and receive our domain models. So to prevent circular dependencies and
confusing imports we put all of our models at the context root (`library/` in
this case).

Now we have our first model, but not a whole lot to do with it. For that, we
need "actions" and those live in a service.

## Service

The idea of a service is a fairly universal concept, but one that still somehow
manages to be difficult to quantify. For the purposes of this lesson, however,
a service is just a state container for hanging domain methods on. This is a bit
esoteric, but it's pretty simple once you see an example, so let's dig in.

I'll start by creating a barebones service type. This takes the form of a struct
that will hold handles to dependencies, configuration, etc.:

```go
type LibraryStore interface {}

type LibraryService struct {
	Store LibraryStore
}
```

In this example, you can see that I've included a `Store` field that right now
is just an empty interface. This is another early optimization I'll do because
almost any service I write is going to need access to the database and this sets
us up to do that.

Next, we need to start thinking about the actions we need to implement. Here is
where I can go one of two different directions depending on the steps that have
gotten me here. If you're of the camp that you should only implement the bare
minimum to move forward, then it may make the most sense to only stub out one
action at a time. On the other hand, by the time I've gotten to this step I've
usually already done a large amount of domain modeling offline and I have a
pretty good idea of what actions I absolutely have to support. In those cases
(and for the sake of this lesson), I will usually stub out all of the actions I
can in one pass. Then I can incrementally implement each in turn while still
having a view of what's left.

For the sake of the lesson, let's assume that domain modeling has already been
done and we've determined that at minimum, we need to support the following
actions:

- Adding a book
- Getting a list of books (likely with some filtering capabilities, but that
  comes later)
- Viewing detail for a single book
- Rating a book

With that in hand, my next step is to stub each service method out. It may seem
intractable to stub functions out when you don't know what the inputs or outputs
are going to be, but here again I have a particular way I write all of my
services that lends itself well to stubs. Basically, every service method has
the same shape for its signature. Each method is named for the action it
implements, takes a context and a request type as arguments, and returns a
response type or an error. Let's see how this looks for the "add a book" action:

```go
type AddBookRequest struct{}
type AddBookResponse struct{}

func (ls *LibraryService) AddBook(ctx context.Context, req AddBookRequest) (*AddBookResponse, error) {
	return nil, fmt.Errorf("TODO")
}
```

Here you can see the shape described above, as well as what the entire stubbing
process produces. Since we don't know what input we need just yet or what we'll
be returning, we've left our request and response structs empty, but we can
still write our method stub without worrying that the signature will change down
the road. For this lesson, I've gone ahead and added stubs for each of the
actions outlined above that follow the same pattern.

So now that you've seen what a sample service method stub looks like, it's
important to understand why. When we get to the database layer, you'll see that
I don't follow a similar convention there and our API methods all have the usual
Go handler shape: `func (http.ResponseWriter, *http.Request)`. So why do my
services always use request and response types? There are two main reasons:
ease of compatibility and ease of translation.

---

### Compatibility

Whenever you write any method, function, type, etc. you are creating an API
whether you want to or not. Any code that uses your type or calls your function
depends explicitly on the contract you created. Consider a function like the
following:

```go
func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
```

This is a fairly simple function and it's not even exported, so we would expect
to be pretty safe from the world of compatibility guarantees. Say, however, that
our simple function is used in 20 places throughout the package where it lives.
If we were to add a second parameter to specify a language to be greeted in, we
would have 20 call-sites to change, even if most of them only ever use the
default language. For a simple function like this, we could just create a second
version `greetInLanguage` and limit the blast radius of the refactoring. We
could even refactor the original to call `greetInLanguage` with English
hard-coded as the second parameter and for anything this simple that option is
probably the first you should reach for.

When we're talking about service calls, however, it's often impossible to
enumerate all of the possible use-cases and the situation only gets more and
more difficult as we add or modify parameters to meet new requirements. This
difficulty stems from the fact that our domain layer is the most subject to
change. By comparison, the API layer has to change much more slowly since we
also have compatibility guarantees to our clients and updating our upstream
dependents is a very slow process. Similarly, the database structure changes
fairly slowly and calls into this layer are usually well known and are largely
dictated by the needs of the domain.

By wrapping all of our inputs and outputs in request and response types
respectively, we can change or add parameters without modifying our function
signature. By carefully considering what potential breaking changes we're
introducing, we can also provide default values or behavior to maintain previous
invariants until we can do a larger-scale refactoring. Consider this alternative
version of the `greet` function:

```go
type greetRequest struct {
	Name     string
}

type greetResponse struct {
	Greeting string
}

func greet(params greetRequest) (*greetResponse, error) {
    return &greetResponse{fmt.Sprintf("Hello, %s!", params.Name)}, nil
}
```

This is much more verbose, but now let's say we're using our function all over
the place and we want to add a language parameter _without_ breaking existing
call-sites. This is how that could work:

```go
type greetRequest struct {
	Name     string
	Language string
}

type greetResponse struct {
	Greeting string
}

func greet(params greetRequest) (*greetResponse, error) {
	if params.Language == "" {
		params.Language = "en"
	}
	switch params.Language {
	case "en":
		return &greetResponse{fmt.Sprintf("Hello, %s!", params.Name)}, nil
	case "de":
		return &greetResponse{fmt.Sprintf("Hallo, %s!", params.Name)}, nil
	default:
		return nil, fmt.Errorf("invalid language: %s", params.Language)
	}
}
```

We haven't changed our function signature so the program will still build and
we've provided a default for the language parameter that matches the previous
behavior so compatibility is maintained. All a new caller has to do to get the
German version is pass the parameter explicitly.

This example is pretty contrived, but imagine your service call to retrieve a
list of books. You might start by returning every book, but eventually you want
to add filtering and sorting so you need to take new parameters. Then you need
to support pagination so you have to take even more inputs to determine what
page of results to return, but you also need to return metadata for the total
count of books so the client can display a nice UI. With standard function
signatures, you'd quickly have an explosion of parameters, wrapper functions,
return values, and refactorings on your hand.

Using request and response types doesn't save you from having to think
critically when writing functions, but it can make the process a bit smoother.

### Translation

The other primary benefit of using request and response types is at the
incoming boundary between your service and the layer (or layers) above it. If
you remember from the first lesson, the service layer is the primary
entrypoint to our business logic. So far we've largely assumed that above that
sits a web API, but a service request could come from a number of places such as
a CLI, a background worker process, etc. Each of these has its own semantics for
collecting the inputs to our business logic, but what we would like is a single
unifying interface into our domain and that's where service requests and
responses come in.

By creating a single type for holding all of our parameters, we can very easily
construct and test translation functions. For a web API, you might have a set of
functions to convert from an HTTP request to a service request and from a
service response back to HTTP. For a CLI, the inputs may be a combination of
command line flags, file-based configuration, and live input. A background
worker may pull everything from the database. What's important is that once the
inputs are known, the process becomes identical no matter where we started and
we can test each of these steps in isolation.

---

So now that you know why I structure things the way I do, we can start filling
in our "add a book" types and method. We'll start with the request type and here
the question is: what do we need to create a new book? If we look at our model
we can see that we really only have two "user settable" fields. There may be a
case to make for allowing user-specified IDs (for instance if we were using
UUIDs and wanted to allow our clients to do some optimistic UI tricks), but to
keep things simple we'll just set that server-side. Completing our request type
then is as simple as adding those fields:

```go
type AddBookRequest struct {
	Title  string
	Author string
}
```

We can do the same kind of thing on the response side, but here we're not
returning individual fields, we're returning the entire book. Note that in the
outgoing direction, our book has already been persisted to the database and as
such should have an ID set, which we'll verify in our tests later. So with that
said, here's our very simple response:

```go
type AddBookResponse struct {
	Book *library.Book
}
```

With those types in hand, we can move to the method itself. This will start out
pretty simple, but as we start to make our application "production-ready" even a
simple method can start to get rather large. We'll leave most of that off for
now to focus on the service itself though:

```go
func (ls *LibraryService) AddBook(ctx context.Context, req AddBookRequest) (*AddBookResponse, error) {
	book, err := ls.Store.CreateBook(ctx, req.Title, req.Author)
	if err != nil {
		return nil, fmt.Errorf("add book: %w", err)
	}
	log.Printf("added book %d: %s\n", book.ID, book.Title)

	return &AddBookResponse{Book: book}, nil
}
```

The most interesting thing to note in this otherwise simple method is the
`CreateBook` method we're calling on our `Store`. If you add this right now and
try to compile, you'll get an error:

```
ls.Store.CreateBook undefined (type LibraryStore is interface with no methods)
```

Looking back at our interface, this makes sense. We're attempting to call a
method where we've expressly stated none exist. This is almost always how I will
build interfaces when I'm working. I find that by deferring the decision to the
point where it's a build error to continue, I can keep my interfaces as small as
possible without having to do a lot of work ahead of time.

To make this right again, we simply add the method we need to our interface:

```go
type LibraryStore interface {
	CreateBook(ctx context.Context, title, author string) (*library.Book, error)
}
```

At the moment, it doesn't matter that we don't have a concrete implementation of
this interface. We've told Go that our service has a handle to something that
does and until we try to run it without setting a value for the `Store` field,
we can continue as if everything will just work out in the end. As you'll see in
the next lesson, this won't affect our ability to test our service and, in fact,
will make that process more tightly focused and less brittle.

For now though, we've added everything we need to add a book to our library...at
least in terms of the business logic to do so. In subsequent lessons we'll take
a look at how we can persist our book to a real database as well as how we can
get meaningful values for the title and author.

## Assignment

Take a look at the code I've already added for adding a book and make sure that
you understand the syntax and, more importantly, the motivations for the design
decisions we've made thus far. Once you feel like you have a solid grasp of the
state of things, see if you can implement basic versions of the "get a book" and
"get a list of books" methods. To keep from getting too complex too quickly,
only consider requests for all books without any additional filtering, sorting,
or scoping. You can ignore the "rate a book" action for now since that involves
some design decisions we haven't quite made yet.

When you're done, you can look at the [solution](../02-first-model-end) to
see how I implemented these methods.

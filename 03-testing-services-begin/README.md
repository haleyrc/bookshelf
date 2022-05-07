# Testing services

> Before we get started with testing, I wanted to say that if you're still
> getting familiar with Go and haven't gotten a lot of tooling setup, you can
> definitely run all of the tests in this lesson with the standard Go commands
> e.g.:
>
> ```
> $ go test ./library/service
> ```
>
> That said, I would definitely take the time to setup a more streamlined method
> such as the Go extension for VSCode or the vim-go plugin for example.

So after last lesson, we've got a model and a service and we _think_ we've made
something that will work, but how can we verify it?

In this lesson, we'll look at how I approach testing at the service layer and
along the way we'll see a little bit about how I do validation. I'll preface any
discussion with a pretty hefty disclaimer: testing strategy is one of those
topics in computer science that nobody agrees on. The way I construct my tests,
how I use mocks, and the types of assertions I make are all largely based on my
own personal preference, but it is a preference based on trying a number of
setups that didn't give me the confidence I wanted in my code. I encourage you
to work through this lesson, but then to experiment a bit to see how you would
improve your own confidence that you were shipping something high quality. Just
take care to avoid the common testing pitfalls like testing implementation
details or testing the standard library. There's confident and there's neurotic
and the line is often pretty gray between the two.

To get started with testing, let's create a test file next to our service
(`library/service/service_test.go`) with an empty test function for our add book
action:

```go
package service_test

func TestLibraryService_AddBook(t *testing.T) {
	// TODO
}
```

If you run the tests for this package now, Go should trivially report a `PASS`.
If you've made any syntax errors or anything else that might cause compilation
to fail, however, you should see that now so you can go correct it.

As far as naming tests go, I will generally try to stick to a pattern in this
layer of `func Test{Service Name}_{Method Name}` whenever possible. Remember
that your test name is the first thing a developer will see when they run it to
let them know where a problem occurred. Now we need some actual test code.

If you've done any testing before, you're likely familiar with the standard test
pattern:

```go
func TestLibraryService_AddBook(t *testing.T) {
	// Setup

	// Call

	// Assert

	// Cleanup
}
```

We'll follow this exact pattern for our own tests and I think it makes sense to
start with our setup code.

> For the rest of this lesson and in subsequent lessons, I'll generally elide
> the surrounding code for brevity's sake except where it would be confusing to
> do so. If you follow along it should all make sense, but if you get lost you
> can always refer to the code for the lesson itself.

For our test, we know we're going to need a service since that's our "object
under test". We also need a context and a request since those are the arguments
to our "function under test", so we'll add all of that to the top of our test
function:

```go
ctx := context.Background()
svc := service.LibraryService{}
req := service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"}
```

For the fields of the request object I've just filled in some dummy data. I will
usually hard-code this kind of input (unless I'm doing something like
table-driven tests) because it prevents accidental scenarios where I end up
comparing two variables who can't possibly be different, but not by virtue of
my live code.

You may also be wondering about the `Store` field of the service and you'd be
right to be curious, but the details of how we're going to approach that part of
our test are a bit hairy so we'll save them for a bit later.

Moving on to the "call" part of our test, we can add the following to invoke
our service method and verify that it returns without error:

```go
resp, err := svc.AddBook(ctx, req)
if err != nil {
	t.Fatal("unexpected error:", err)
}
```

This is exactly the call pattern we would expect a client to use, except in our
test we bail early if we don't get a response since all of our assertions assume
that the call succeeded.

So now we've got a response object we can start making assertions against, but
what should those assertions look like? Well that depends on what our FUT is
supposed to do. For adding a book, we can sum it up by saying: adding a book
should take a title and an author, save the object in the database, and return
the persisted book. Since we're starting with the response, we're interested in
what the book that we're receiving looks like. Ideally it would match the title
and author we provided, but if you recall we also have an ID that we would
expect to get back as a result. So let's go ahead and add those assertions:

```go
if resp.Book == nil {
	t.Fatal("expected book to not be nil, but it was")
}
if resp.Book.ID == 0 {
	t.Errorf("expected book to include an id, but it was blank")
}
if resp.Book.Title != "Dune" {
	t.Errorf("expected book to have title \"Dune\", but got %q", resp.Book.Title)
}
if resp.Book.Author != "Frank Herbert" {
	t.Errorf("expected book to have author \"Frank Herbert\", but got %q", resp.Book.Author)
}
```

Hopefully this looks fairly straightforward, but there are a couple bits of
nuance I want to point out. First, we start by checking that we receive a book
at all. This is important because part of the contract of this method is that we
get back the persisted book object, but it's also a fatal error if we don't
since we can't make any more assumptions after that point.

Next, we're comparing the received ID to 0 to check that it wasn't set. This
kind of assumption is a bit dangerous because it's possible for a random integer
to be 0 and still be valid. Since this is expressly a database ID, however, we
can be sure that if the field has the zero value, it wasn't set. In the end
though, be sure to match your assertions to the actual behavior of your system.

Finally, you'll see that I once again hard-code the expected values. It's more
likely that a typo will creep in doing it this way, but that would be
immediately apparent in the test output and once again we've ensured that the
value can only be "correct" if it matches a static value exactly.

So now we've got all the makings of a test, but there's one problem. If you run
this now, you'll get a panic. If you dig into the stacktrace, you'll see the
offending line is this one:

```go
book, err := ls.Store.CreateBook(ctx, req.Title, req.Author)
```

Ah, so there's that `Store` thing again. Luckily, this all makes sense. In our
service we're saying we have a dependency on a `Store` interface that exposes a
`CreateBook` method, but we haven't actually set that field to anything in our
test. We're literally calling methods on `nil`. But we definitely don't want to
set up a whole database and write code to interact with it. That's the realm of
an integration or end-to-end test. We just want to test domain logic as simply
as possible.

> **Mocks**
>
> I feel the need to break for a bit to talk about my thoughts on mocks et al _as
> it pertains to Go_. First, I'm not staunchly for or against TDD. I think that
> like most things, there are pros and cons to TDD as a whole and also like most
> things that the correct solution lies somewhere in the middle. That said, I do
> think that _testing_ is critical, so anything that encourages testing has at
> least some upside.
>
> I say all that to say that I will use terms like "mock" in a very loosely
> defined way because I'm not trying to teach TDD in this course. If you have a
> different view of what a mock is and think I'm using stubs, or doubles, or spies
> that's fine, but it's not the point. Instead of focusing on the terminology,
> focus instead on the meaning behind the code. If you can understand why a thing
> is done that's a lot more crucial than knowing what it's called.

Alright, we're almost to a working test, but we need a mock for our `Store`.
There are a million different ways to go about this and the way I use is
definitely on the border of over-engineered and there's a lot of boilerplate to
boot. Of course, I do it this way because I feel the trade-offs still favor it
for my use-case. So let's dig in to how I construct a mock. Here's the code
(don't worry if it's a bit confusing, I'll explain all the bits):

```go

type mockStore struct {}

func (ms *mockStore) CreateBook(ctx context.Context, title, author string) (*library.Book, error) {
	return nil, fmt.Errorf("TODO")
}

func (ms *mockStore) GetBookByID(ctx context.Context, id int64) (*library.Book, error) {
	return nil, fmt.Errorf("TODO")
}

func (ms *mockStore) GetBooks(ctx context.Context) ([]*library.Book, error) {
	return nil, fmt.Errorf("TODO")
}
```

So we've got a new type and a few new methods. If you look at the methods, you
might notice that they match the functions in our `Store` interface. In the
usual Go way then, our mock store implements the `Store` interface and can be
used in our service. We can wire that up by modifying our setup code:

```go
store := &mockStore{}
svc := service.LibraryService{Store: store}
```

After you make this change, we've finally got a test that will run! Of course,
if you do you'll see something like the following:

```
=== RUN   TestLibraryService_AddBook
    /Users/username/path/to/bookshelf/03-testing-services-begin/library/service/service_test.go:20: unexpected error: add book: TODO
--- FAIL: TestLibraryService_AddBook (0.00s)
FAIL
FAIL	github.com/username/bookshelf/library/service	0.325s
```

This is a bit underwhelming, but it's also important. We're actually already
testing something: what does our service do if the database just returns an
error. That's not a super interesting test though. How we can improve? Well as
before, we start by determing what it is we _should_ be testing. In our previous
assertions it was fairly simple to test because we were getting back a response.
Here, we don't have a return value that we can use so we'll need to do a bit of
trickery to get something of value to test against.

Now at this point it's tempting to start implementing an in-memory database and
then writing additional wiring code to verify that if we save something to that
database and then retrieve it, etc. etc. etc. This is appealing because it feels
like we're testing something real, but what we end up doing is testing how good
an in-memory database regurgitates data. Not all that useful after all.

When we get to this point, I encourage people to think about "testing at the
interfaces". What I mean by that is this: if we think of our FUT as a black-box
all we can see is the data that enters and exits. These are the interfaces to
our function. Two obvious interfaces are the parameters and the return values,
but many functions have other avenues for data to escape and I find that it's
important to test these as well since they define the interactions within the
system as a whole. In this case, the additional interface is the one we defined:
the `Store`. We send data to the `Store` via the `CreateBook` method (and we
get stuff back, but we'll see how we model this later). So how can we test this
interface without writing a whole database. Well I do it with a mix of behavior
injection and plain old bookkeeping. Before we get to the implementation of our
mock store, however, let's see how we actually use it so we can get a sense of
the purpose of it all. Here are the assertions I would add to our test (above
the response tests if you're curious; that's where it happens in the code so I
try to keep things in unison in the test):

```go
if len(store.CreateBookCalledWith) != 1 {
	t.Fatalf("expected CreateBook to be called once, but was called %d times", len(store.CreateBookCalledWith))
}

args := store.CreateBookCalledWith[0]
if args.Title != "Dune" {
	t.Errorf("expected CreateBook to be called with title \"Dune\", but got %q", args.Title)
}
if args.Author != "Frank Herbert" {
	t.Errorf("expected CreateBook to be called with author \"Frank Herbert\", but got %q", args.Author)
}
```

Immediately, we run into another possible area of contention with other
testing strategies. I like to verify that my store methods are called in
accordance with my business logic. For this method, I know that one of my
contract items is that I save the book to the database, so this method _must_ be
called for me to fulfill that contract. This does border on testing
implementation details, but if you think of the database _layer_ as a downstream
dependency, it's more akin to observing the effects of a particular function
call.

The next two assertions are for a similar purpose. Not only do we want to know
that we sent a message to the right place, we want to verify that the message
contains the data we expect; in this case that's the title and author we passed
to the service.

So now we know what assertions we want to write, but we don't get this behavior
for free, so let's flesh out our mock a bit:

```go
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
```

This might look a bit intimidating, but the work it's doing is pretty mundane.
When our mock's `CreateBook` method is called, we're first checking for the
presence of a function in this weird `CreateBookFn` field. This is where our
behavior injection will come in. We're simply storing an implementation of the
function with this signature and if there isn't an implementation stored, we're
throwing an error. If you compare the function signature of `CreateBook` and
`CreateBookFn`, you'll see they're identical. This is an important part of the
design of our mock. By storing additional behavior, we can get all of the spying
and tracking in one package, but swap out alternate behaviors for various test
scenarios. You'll see how powerful this is when we start to test failing cases.

Next, we're saving the arguments that our function was called with. This will
allow us to test the data that passed between our service and our store against
our expectations. The only really interesting bit of this is that I prefer to
use anonymous structs here to keep from polluting the namespace with a million
one-off struct types.

Finally, we call into the "behavior" function with the arguments we were given.
This is why we have the `nil` check above. If we don't have an implementation,
this would immediately panic which isn't useful as a test goes. So we warn the
developer early that they need to provide one. In fact, if you run the tests now
you'll see something like this:

```
=== RUN   TestLibraryService_AddBook
    /Users/username/path/to/bookshelf/03-testing-services-begin/library/service/service_test.go:22: unexpected error: add book: mock not implemented
--- FAIL: TestLibraryService_AddBook (0.00s)
FAIL
FAIL	github.com/username/bookshelf/library/service	0.344s
```

If you remember from above, we created an instance of our mock store and wired
it into our service, but we never provided an implementation of our
`CreateBookFn` function. Let's do that now:

```go
store := &mockStore{
	CreateBookFn: func(_ context.Context, title, author string) (*library.Book, error) {
		return &library.Book{ID: 123, Title: title, Author: author}, nil
	},
}
```

This may look a little weird if you haven't used first-class functions all that
much in Go, but all we're doing is assigning `CreateBookFn` to an anonymous
function with a matching signature. Inside of our lambda, we're just returning
a book instance with the title and author we were provided. Note that we're also
returning an ID that we made up. This simulates what we expect from our
database, but the specific value isn't important. In fact, each of these values
is arbitrary and they don't pertain to our store tests. They _are_ important,
however, since they pertain directly to our response tests. The values we return
here are then returned as part of our response, which is what lets us verify
that the service fulfills the portion of the contract that dictates it return
the persisted book unchanged. In fact, if you run the test now it will pass:

```
=== RUN   TestLibraryService_AddBook
--- PASS: TestLibraryService_AddBook (0.00s)
PASS
ok  	github.com/username/bookshelf/library/service	0.194s
```

Before we call our test complete, however, there's a final change we can make to
hammer home how the return value of our "behavior" function lets us write more
meaningful tests of our response. To finish up, lets replace the existing test
that read `if resp.Book.ID == 0` with the following assertion:

```go
if resp.Book.ID != 123 {
	t.Errorf("expected book to have id 123, but got %d", resp.Book.ID)
}
```

Remember how above I mentioned that testing against 0 was potentially dangerous?
Well by structuring our mock the way we did, we can avoid the situation entirely
by giving us a way to check the much stronger assertion: given a book with a
valid ID returned from the store, the response includes the same ID. If all goes
well, you should be able to run the test and see the same `PASS` as before.

I started this lesson with a disclaimer and I hope that I've at least somewhat
justified how I write service tests. If you're still unconvinced, I would love
to hear alternative approaches to testing here. I think any additional
perspective has the potential to make testing even simpler and more useful and
that's always a Good Thing.

With one passing, happy-path test in hand, lets now talk about validation and
along with that two more important testing concepts: table driven tests and
sad-path tests.

## Table driven tests

Before we get started with our validation, let's do a bit of refactoring in our
test file that will set us up for adding additional test to this one function.

We'll start by introducing the idea of table driven tests. If you haven't seen
table driven tests before, they can look a bit odd, but in practice they're
fairly simple. Essentially, you provide a list of inputs and expected outputs to
a function and then loop over each of these tuples, call the FUT with the inputs
and compare against the outputs. Once again this will make more sense in
practice, so lets add the first bit of the puzzle. Start by adding the following
to the top of our existing test function:

```go
testcases := map[string]struct {
	Request service.AddBookRequest
}{}

for name, tc := range testcases {
	t.Run(name, func(t *testing.T) {
		// TODO
	})
}
```

This outlines the two main components of a table driven tests: a collection of
test cases and a loop to process them. For each test case, we're setting a
request that will be passed to our service call. In the general case, you can
put whatever you need in your test cases.

Below our test cases, we loop over each in turn, running a sub-test that forms a
closure over our single case and allows us to perform all of the same
assumptions we would in a standard test. But what goes in our sub-test and what
should we put in our request? Well if the request in our test case is what gets
passed to our service call, we can simply make it match the request we created
earlier:

```go
testcases := map[string]struct {
	Request service.AddBookRequest
}{
	"Happy Path": {
		Request: service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"},
	},
}
```

Now we have some inputs, we need some assertions. Well for now we can just use
the ones we already wrote. I won't paste the entire resultant code here because
it's a fairly mechanical transformation, but all you need to do is move the rest
of the code up into the sub-test.

If you do this and try to run the tests you'll get a compiler error warning you
that we aren't using the `tc` variable from our loop. This gives us a good idea
of where to go next and the final transformation to get back to passing is to
replace all of the hard-coded title and author strings with the values from our
test case. We can also drop the manually constructed request object since we get
that from our test case as well. The resulting function looks like the
following:

```go
func TestLibraryService_AddBook(t *testing.T) {
	testcases := map[string]struct {
		Request service.AddBookRequest
	}{
		"Happy Path": {
			Request: service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"},
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
```

If you recall from earlier in the lesson, I mentioned that I will usually
hard-code input values unless I'm using table driven tests. Now you can see why.
We just don't need to when we're wrapping our inputs up in a neat package.
Running our tests again shows that we're back to passing and also highlights how
nice the output is when we pass a meaningful name to `t.Run`:

```
=== RUN   TestLibraryService_AddBook
=== RUN   TestLibraryService_AddBook/Happy_Path
--- PASS: TestLibraryService_AddBook (0.00s)
    --- PASS: TestLibraryService_AddBook/Happy_Path (0.00s)
PASS
```

## Validation

Now that we've got our table driven tests setup we're ready to start breaking
stuff with validation. In the spirit of TDD we'll go ahead and start by writing
the tests and here you'll see another benefit of table driven tests (TDT?): it's
now incrediby simple to add more test cases.

There are a number of validations we could consider here, depending on the exact
business requirements, but for the purposes of this lesson we'll just say that
both title and author are required (that is: they can't be blank strings). Those
test cases are pretty trivial to add to our suite so let's start there:

```diff
testcases := map[string]struct {
	Request service.AddBookRequest
}{
+	"Empty Request": {
+		Request: service.AddBookRequest{},
+	},
+	"Blank Title": {
+		Request: service.AddBookRequest{Title: "", Author: "Frank Herbert"},
+	},
+	"Blank Author": {
+		Request: service.AddBookRequest{Title: "Dune", Author: ""},
+	},
	"Happy Path": {
		Request: service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"},
	},
}
```

This gets us part of the way there, but if we run our tests we'll get three
failures. Our current assertions are always expecting the service call to
succeed, but we're _trying_ to get an error. To get this to work, we'll need to
add an extra field to our test cases and update our assertions. Starting with
our test cases, we add the following field:

```diff
testcases := map[string]struct {
	Request service.AddBookRequest
+	ShouldErr bool
}{
```

Now for each test case, we set the value of `ShouldErr` according to whether it
should pass validation. I only show two examples here, but the others should
match the "Empty Request" case:

```diff
"Empty Request": {
	Request:   service.AddBookRequest{},
+	ShouldErr: true,
},
// ...
"Happy Path": {
	Request:   service.AddBookRequest{Title: "Dune", Author: "Frank Herbert"},
+	ShouldErr: false,
},
```

Finally, we need to update the error check after our service call to take
advantage of our new field:

```diff
resp, err := svc.AddBook(ctx, tc.Request)
+if tc.ShouldErr {
+	if err == nil {
+		t.Fatal("expected an error, but got nil")
+	}
+	return
+}
if err != nil {
	t.Fatal("unexpected error:", err)
}
```

Now we're checking to see if we should expect an error and if so, that we got
one. If we don't get one, we bail immediately because whatever response we got
is irrelevant. If we do get an error we exit our test case because we got the
behavior we wanted.

The following conditional is still required for the happy path tests, but now
we're covering all of our bases. Now when we run our tests, we're finally
getting somewhere and we've got some failing validation tests:

```
=== RUN   TestLibraryService_AddBook
=== RUN   TestLibraryService_AddBook/Blank_Author
    /Users/username/path/to/bookshelf/03-testing-services-begin/library/service/service_test.go:50: expected an error, but got nil
=== RUN   TestLibraryService_AddBook/Happy_Path
=== RUN   TestLibraryService_AddBook/Empty_Request
    /Users/username/path/to/bookshelf/03-testing-services-begin/library/service/service_test.go:50: expected an error, but got nil
=== RUN   TestLibraryService_AddBook/Blank_Title
    /Users/username/path/to/bookshelf/03-testing-services-begin/library/service/service_test.go:50: expected an error, but got nil
--- FAIL: TestLibraryService_AddBook (0.00s)
    --- FAIL: TestLibraryService_AddBook/Blank_Author (0.00s)
    --- PASS: TestLibraryService_AddBook/Happy_Path (0.00s)
    --- FAIL: TestLibraryService_AddBook/Empty_Request (0.00s)
    --- FAIL: TestLibraryService_AddBook/Blank_Title (0.00s)
FAIL
```

The next step is to get back to green by adding in the validation. At the top of
our service method we can just add:

```go
if req.Title == "" {
	return nil, fmt.Errorf("add book: title can't be blank")
}
if req.Author == "" {
	return nil, fmt.Errorf("add book: author can't be blank")
}
```

In later lessons we'll take a critical look at how we're reporting errors, but
for now, this is enough to get us back to green:

```
=== RUN   TestLibraryService_AddBook
=== RUN   TestLibraryService_AddBook/Happy_Path
=== RUN   TestLibraryService_AddBook/Empty_Request
=== RUN   TestLibraryService_AddBook/Blank_Title
=== RUN   TestLibraryService_AddBook/Blank_Author
--- PASS: TestLibraryService_AddBook (0.00s)
    --- PASS: TestLibraryService_AddBook/Happy_Path (0.00s)
    --- PASS: TestLibraryService_AddBook/Empty_Request (0.00s)
    --- PASS: TestLibraryService_AddBook/Blank_Title (0.00s)
    --- PASS: TestLibraryService_AddBook/Blank_Author (0.00s)
PASS
```

Once again, you can see how nice the output is when we make good use of sub-test
naming. If you run `go test` with coverage, you should see that we're now testing
every line of our service method except for the error return. This is fairly
common since we're just wrapping and returning an error and aiming for 100%
coverage would be a bit of a waste.

## Assignment

Now that you've seen how I structure my mocks and how I test at various levels,
it's up to you to implement the tests for the get book and get books cases. As a
bit of a hint, you'll need to figure out how to implement the mock store
method and how to assert against a list of results. You'll also need to make a
determination in both cases about whether it makes sense to use table driven
tests or not.

When you're done, you can check my implementation
[here](../03-testing-services-end).

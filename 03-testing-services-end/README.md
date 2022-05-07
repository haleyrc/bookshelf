# Testing services - Solution

> Code: [`service_test.go`](./library/service/service_test.go)

This assignment was definitely a step up from previous ones and required a good
amount of decision-making, so don't be discouraged if what you ended up with
doesn't look exactly like what I wrote. The important part here is internalizing
the concepts more than the specific implementation details. That said, let's
look at some of the more interesting points of my solution:

## Getting a book

This test ended up looking very similar to the "add a book" test. One major
deviation is that since we don't have a title and author as part of our service
request (remember these would come from the database layer in this method), I
once again hard-coded them into the return value from our `GetBookFn` lambda.
On the other hand, we _do_ have an ID in our request so we use that in
`GetBookFn` as well as our assertions.

## Getting a list of books

This test also looks familiar, but that's because it's closer to our initial
run for the `AddBook` test. Since we don't have any inputs to our service
method, we also don't have any need to run our assertions multiple times. For
that reason, we can eschew the table driven test and just test the happy path.

We also have the new situation where we need to handle a list of results instead
of a single book in both the `GetBooksFn` method and our response. The store
case is fairly simple as we just return multiple entities to test against. We
can be reasonably sure that if our logic works for two results, it will work for
the `n+1` case.

For our response assertions, we can just peel off each book and assert against
them individually. Once again, what works for two likely works for more.

## Next

Hopefully you were able to get to a working solution, but if not, take some
extra time to work back through the solution code and solidify any parts that
remain unclear in your head before moving on.

When you feel ready to proceed, you can start with lesson 4: [TODO](#).

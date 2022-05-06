# Scaffolding

In this lesson, we're going to focus on getting the bones of our project laid
out. I'll be doing most of the grunt work for you, so the important part as
you're following along is making sure that you understand not only the layout of
the directories and files, but also their purpose in the final product.

## What we're building

For this course, we're going to be working within a pretty small domain: a book
rating application. I wanted to pick something that was universal enough to not
need a long explanation, simple enough to not get bogged down in domain
modeling, but complicated enough to give us some interesting migration problems
to tackle later on.

With all that said, here's the basic premise of the application. Users can sign
up which allows them to add and rate books that they've read. They can also see
books that other users have added and rate those as well. That's it for now. I
haven't made any attempt to flesh these out as user stories since I want to
focus on building the app and not on project management terms and processes.

As an aside, don't feel like you need to follow along with this example. Feel
free to take some creative liberties and invent your own application to use
as you learn these concepts. Take note though that if your domain is too small,
you may not have a lot of options for expanding it in future lessons so give
yourself some wiggle room.

## File structure

There are a lot of conflicting opinions on how much design is too much when
starting a new Go project. Perhaps the most common advice is to start with a
single package that encapsulates everything and only expand out from there when
you've determined a need. This is okay advice, especially for someone new to Go
or for a proof-of-concept with only very vaguely defined scope. That said, I can
say with some certainty that you aren't going to write an entire API in a single
package and have it be of any reasonable maintainability.

The most important part of scaffolding any new project is just finding a set of
conventions that work for you. The scaffold I'll lay out here is based on my
personal experiences writing APIs in Go and may not make sense for every
project, but this is how **I** would start a new project today. Hopefully, the
explanations in this lesson will give you a good idea of my thought process so
that you can decide how much (or how little) design makes sense for your own
projects.

```
.
├── Dockerfile
├── README.md
├── app
├── auth
│   ├── api
│   ├── service
│   └── store
├── common
├── docker-compose.yml
├── go.mod
├── library
│   ├── api
│   ├── service
│   └── store
├── main.go
└── migrations
```

### `app/`

This is the main container for all of the "moving parts" of our application.
Eventually, this is where we'll define our top-level application structure and
handle parsing configuration, setting up database connections, and wiring up all
of our individual services into a runnable artifact.

By encapsulating all of this in our `app` package, we can run end-to-end tests
easily against our actual server.

### `common/`

This is where we'll put shared functionality such as JSON marshaling and
un-marshaling that should be consistent across
[bounded contexts](https://martinfowler.com/bliki/BoundedContext.html). If a
package is only used by a single context, it should instead live within that
package as close to the call-site as possible.

You should always put the actual code in separate packages specific to the
use-case (e.g. `json`, `date`) and not turn this directory into a grab-bag of
random functionality.

### `auth/`

This is our first package specific to this domain (although auth is likely to be
a part of any domain you might be working in). I create these "bounded context
packages" at the top-level since it should be fairly obvious what they are for
from the naming. _Inside_ of these packages is where most of the fun stuff
happens.

#### `auth/api/`

The `api` package provides a place to store the logic for converting from an
HTTP request to a service request and from a service response to an HTTP
response. This package includes types specific to marshaling and unmarshaling
that represent the shape of the data expected in incoming requests and returned
in responses. These types may include things like `json` or `xml` struct tags.

The primary takeaway for this package is that by separating our API types from
the domain types, we can more easily make changes to business logic without
breaking existing contracts.

#### `auth/service/`

The `service` package provides a central location for all of our business logic.
You can consider the service layer as the primary entry point for the
application; everything boils down to a service request/response loop
independent of the transport mechanism (API call, CLI command, etc.) and the
database implementation (SQL, NoSQL, filesystem).

This is where we define our domain types that correspond to the language we use
within this context. These types represent the concepts within the domain and as
such may be composed of multiple persistence or API types. Domain types may have
methods on them for doing transformations or calculations, but they should not
have any JSON or database struct tags; those are the responsibility of their
respective packages.

Both the `api` and `store` packages deal in domain types directly so there is a
dependency from the outer layers to the core logic, but not vice versa. This is
done so that every package "speaks" in terms of the domain, even if they
immediately convert to an internal type for their own purposes. Otherwise, your
service layer has tight coupling with all of the ephemeral layers around it
instead of providing a single bedrock for everything else to work from.

#### `auth/store/`

The `store` package is where we do the nitty-gritty of pulling data from and
storing data in the database/filesystem/etc. As with the API side, we create new
types that match the structure we need for efficient querying. This may or may
not look like the domain types, but we can more explicitly prevent our API from
looking like our database by requiring this kind of type conversion between
layers.

### `library/`

In our sample application, this is the second bounded context. The layout
matches that of the `auth` package but now contains concerns around managing the
books, authors, and ratings and not the authentication mechanisms.

### `migrations/`

This is where all of our database migrations go. In the past, I've used both
`up` and `down` directories and that may well be a strategy to explore in a
larger context. For small projects with ephemeral tests databases however, it's
probably overkill. It's also not incredibly useful in production since doing a
down migration can be extremely dangerous and it's almost always easier to
simply roll the affected code back and possibly do another "forward" migration
to get back to a good state.

### `docker-compose.yml`

The provided Docker Compose configuration lets us run our whole backend stack
locally, without having to install a bunch of extra dependencies and manage
ports, etc. Despite the fiddly nature of Docker, this can increase productivity
so much that I still recommend just starting here. Chances are you're going to
need some Docker skills at some point anyway.

### `Dockerfile`

The Dockerfile builds our application and provides a container image for Docker
Compose to use, but if done correctly also provides a production-ready image for
upload to whatever provider you're using for deploys (assuming they use a
Docker-based system which most do).

### `go.mod`

This file is required to build and run your application. For now, just follow
the instructions in the Assignment section to update it to match your setup.
You will almost never have to interact with `go.mod` manually.

### `main.go`

This file provides the entrypoint into our application. In order to keep things
testable, this package contains almost no code, preferring instead to offload
the bulk of the work to the `app` package.

### `README.md`

Every good project should have a good README.

## Assignment

For this inaugural assignment, we'll just get setup with a bare repo that we can
use for the rest of the course. Please note that this course assumes a basic
knowledge of development concepts and tooling, so I'll only be posting commands
without any real explanation or ceremony and there won't be any tutorials on
installing Git, Go, or any other tool I use. In addition, the commands in this
course are the commands that **I** would run on my MacBook, in iTerm2, using the
zsh shell. Your exact setup is almost certainly different so make sure that you
are comfortable with your own development environment before proceeding.

Start by creating a new directory for our project. Go doesn't have any
official scaffolding tools or generators, we'll be doing things the
old-fashioned way:

```
$ mkdir bookshelf
$ cd bookshelf
```

I'm assuming you'll want to save to version control, so go ahead and create a
new repository on your provider of choice and then run your standard Git
commands to wire everything up:

```
$ git init
$ git remote add origin git@github.com/GITHUB_USERNAME/GITHUB_REPO.git
```

replacing `GITHUB_USERNAME` and `GITHUB_REPO` with the appropriate values for
your new repository.

Once that's done, you'll need to create a new Go module. Run the following
command to create the `go.mod` file that both marks this as a Go module and
serves as your "dependencies" file a la `package.json`, `requirements.txt`,
etc.:

```
$ go mod init github.com/GITHUB_USERNAME/GITHUB_REPO.git
```

Now we're going to create an empty Go file just to verify that our environment
is setup correctly. Run:

```
$ echo "package main" > main.go
$ go test .
```

You should see output similar to the following:

```
?   github.com/{GITHUB_USERNAME}/{GITHUB_REPO}  [no test files]
```

To complete the setup process, you can copy the contents of this lesson's
solution directory to your own, taking care not to replace the `go.mod` file we
created earlier (it's fine to replace the `main.go` file as it doesn't contain
any information specific to my setup).

Now we're all set so we can create our first commit and push to Github:

```
$ git add .
$ git commit -am "Initial commit"
$ git push -u origin main
```

Now that you have a working environment for the rest of the course, take some
time to walk through the directory structure and files provided and make sure
that you understand their purpose in context. We'll be filling these out fairly
quickly as we proceed so if you're not clear on the "why", now is the time to
ask questions and solidify that understanding.

---

Once you feel comfortable with this lesson, you can see how it looks in practice
[here](../01-scaffolding-end).

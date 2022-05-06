from golang:1.18 as build

ENV CGO_ENABLED=0

WORKDIR /bookshelf

ADD . .

RUN go build -ldflags="-s -w" -o bin/bookshelf

FROM alpine:3.15

COPY --from=build /bookshelf/bin/bookshelf /bookshelf

EXPOSE 8080

ENTRYPOINT ["/bookshelf"]

FROM golang:1.22-alpine

RUN apk add --no-cache git

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

WORKDIR /migrations

ENTRYPOINT ["goose"]

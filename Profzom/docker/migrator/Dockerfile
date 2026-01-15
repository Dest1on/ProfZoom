FROM golang:1.24-alpine

RUN apk add --no-cache git

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

WORKDIR /migrations

ENV GOOSE_DRIVER=postgres
ENV GOOSE_DBSTRING=postgres://profzoom:profzoom@postgres:5432/profzoom?sslmode=disable

CMD ["goose", "up"]

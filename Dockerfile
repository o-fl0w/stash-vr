# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY ./cmd ./cmd/
COPY ./internal ./internal/
COPY ./pkg ./pkg/

RUN go generate ./cmd/stash-vr/ && go build -o ./stash-vr ./cmd/stash-vr/

FROM alpine:3.16

WORKDIR /deploy

COPY --from=build /app/stash-vr ./

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666

ENTRYPOINT ["./stash-vr"]
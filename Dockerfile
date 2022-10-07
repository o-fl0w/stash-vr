# syntax=docker/dockerfile:1
FROM golang:1.19-alpine as env-vips-dev

RUN apk update && apk add build-base vips-dev

FROM env-vips-dev as build

ARG BUILD_VERSION

WORKDIR /build

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY ./cmd ./cmd/
COPY ./internal ./internal/

RUN go generate ./cmd/stash-vr/ && go build -ldflags "-X stash-vr/internal/application.BuildVersion=$BUILD_VERSION" -o ./stash-vr ./cmd/stash-vr/

FROM alpine:3.16 as env-vips

RUN apk update && apk add vips

FROM env-vips as deploy

WORKDIR /app

COPY ./web ./web/
COPY --from=build /build/stash-vr ./

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666

ENTRYPOINT ["./stash-vr"]


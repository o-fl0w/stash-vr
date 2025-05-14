# syntax=docker/dockerfile:1

FROM golang:1.24-alpine as build

ARG BUILD_VERSION
ARG BUILD_SHA

WORKDIR /build

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY ./cmd ./cmd/
COPY ./internal ./internal/
#COPY ./pkg ./pkg/

RUN go generate ./cmd/stash-vr/ && go build -ldflags "-X stash-vr/internal/build.Version=$BUILD_VERSION -X stash-vr/internal/build.SHA=$BUILD_SHA" -o ./stash-vr ./cmd/stash-vr/

FROM alpine:3.21

WORKDIR /app

COPY --from=build /build/stash-vr ./

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666

ENTRYPOINT ["./stash-vr"]
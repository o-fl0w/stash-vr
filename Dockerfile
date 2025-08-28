# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build

ARG BUILD_VERSION=dev
ARG BUILD_SHA=unknown

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY cmd/ internal/ ./

RUN go build -ldflags "-X stash-vr/internal/build.Version=$BUILD_VERSION -X stash-vr/internal/build.SHA=$BUILD_SHA" -o ./stash-vr ./cmd/stash-vr/

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=build /build/stash-vr ./

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666
USER nonroot:nonroot

ENTRYPOINT ["./stash-vr"]
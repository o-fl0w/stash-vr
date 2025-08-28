# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

ARG BUILD_VERSION=dev
ARG BUILD_SHA=unknown

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY cmd/ internal/ ./

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags "-s -w \
        -X stash-vr/internal/build.Version=${BUILD_VERSION} \
        -X stash-vr/internal/build.SHA=${BUILD_SHA}" \
      -o /out/app ./cmd/stash-vr

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=build /out/app /app/stash-vr

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666
USER nonroot:nonroot

ENTRYPOINT ["/app/stash-vr"]
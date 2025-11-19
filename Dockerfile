# syntax=docker/dockerfile:1.7

ARG BUILDPLATFORM

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

ARG TARGETOS
ARG TARGETARCH

ARG BUILD_VERSION=dev
ARG BUILD_SHA=unknown

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X stash-vr/internal/build.Version=$BUILD_VERSION -X stash-vr/internal/build.SHA=$BUILD_SHA" \
      -o /out/stash-vr ./cmd/stash-vr

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/stash-vr /app/stash-vr
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

ENV STASH_GRAPHQL_URL=http://localhost:9999/graphql

EXPOSE 9666
USER nonroot:nonroot
ENTRYPOINT ["/app/stash-vr"]
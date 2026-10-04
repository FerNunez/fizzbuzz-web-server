# syntax=docker/dockerfile:1
# Build
FROM golang:1.27-alpine AS build
WORKDIR /src

# NOTE: add go.sum if 3rd party dep
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /out/server ./cmd/server

# Copy from Build
FROM alpine:3.24
WORKDIR /app
RUN adduser -D -u 1001 app

COPY --from=build /out/server ./server

USER app

EXPOSE 8081
ENTRYPOINT ["./server"]
# TODO: add CMD for params? PORT, Max values?

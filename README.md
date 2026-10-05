# FizzBuzz Web Server

> Server for REST API implementation of FizzBuzz.

[![CI](https://github.com/FerNunez/fizzbuzz-web-server/actions/workflows/ci.yaml/badge.svg)](https://github.com/FerNunez/fizzbuzz-web-server/actions/workflows/ci.yaml) ![Go version](https://img.shields.io/github/go-mod/go-version/FerNunez/fizzbuzz-web-server) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

A small Go HTTP service that generates customizable FizzBuzz sequences. It also records every valid request and reports the most frequent one.

Given `int1`, `int2`, `limit`, `str1` and `str2`, it returns the numbers from 1 to `limit`, where:
- multiples of `int1` are replaced by `str1`
- multiples of `int2` are replaced by `str2`
- multiples of both are replaced by `str1` + `str2`

## Features

- REST API:
  - `POST /fizzbuzz` generates a sequence from a JSON body
  - `GET /fizzbuzz/statistics` returns the most frequent request and its hit count
  - `GET /health` for liveness checks
- Standard library only
- Multi-stage Docker image
- CI on GitHub Actions: lint, vet, format, race-enabled tests, Docker build
- OpenAPI 3.0 contract in [`api/openapi.yaml`](api/openapi.yaml)

## Quick Start

**Requirements:** Go 1.27+ or Docker.

### Local

```bash
go run ./cmd/server
# fizzbuzz server listening address=:8081
```

### Docker

```bash
docker build -t fizzbuzz-web-server .
docker run --rm -p 8081:8081 fizzbuzz-web-server
```

Settings are passed as environment variables (see [Configuration](#configuration)). If you change `HTTP_ADDR`, map the matching port:

```bash
docker run --rm -p 9000:9000 -e HTTP_ADDR=:9000 -e LOG_LEVEL=DEBUG fizzbuzz-web-server
```

Then check it is up:

```bash
curl -s localhost:8081/health
# ok
```

## Configuration

All settings are environment variables with defaults. An invalid value (for example `MAX_LIMIT=abc` or `MAX_LIMIT=0`) stops the server at startup with an error listing every problem.

| Variable         | Default | Description                                   |
|------------------|---------|-----------------------------------------------|
| `HTTP_ADDR`      | `:8081` | Address the server listens on                 |
| `LOG_LEVEL`      | `INFO`  | `DEBUG`, `INFO`, `WARN` or `ERROR`            |
| `MAX_LIMIT`      | `10000` | Maximum accepted `limit`                      |
| `MAX_STR_LENGTH` | `100`   | Maximum accepted length of `str1` and `str2`  |

```bash
LOG_LEVEL=DEBUG HTTP_ADDR=:9000 go run ./cmd/server
```

## API

The full contract is in [`api/openapi.yaml`](api/openapi.yaml). To browse it, paste it into [editor.swagger.io](https://editor.swagger.io).

| Method | Path                   | Description                                   |
|--------|------------------------|-----------------------------------------------|
| `POST` | `/fizzbuzz`            | Generate a FizzBuzz sequence                  |
| `GET`  | `/fizzbuzz/statistics` | Most frequent request and its number of hits  |
| `GET`  | `/health`              | Health check                                  |

All JSON responses use the same envelope:
- on success:`{"data": <any data>}`
- on failure: `{"error": {"code": <string machine_code>, "message": <string human_message>}}`.

### POST /fizzbuzz

```bash
curl -s -X POST localhost:8081/fizzbuzz \
  -H 'Content-Type: application/json' \
  -d '{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"buzz"}'
```

```json
{"data":["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]}
```

| Field   | Type    | Rule                     |
|---------|---------|--------------------------|
| `int1`  | integer | ≥ 1                      |
| `int2`  | integer | ≥ 1                      |
| `limit` | integer | 1 to 10000 (`MAX_LIMIT`) |
| `str1`  | string  | 1 to 100 characters (`MAX_STR_LENGTH`) |
| `str2`  | string  | 1 to 100 characters (`MAX_STR_LENGTH`) |

All fields are required, and unknown fields are rejected.

| Status | `error.code`     | When                                                   |
|--------|------------------|--------------------------------------------------------|
| 400    | `bad_request`    | Body is not valid JSON, has an unknown field or a wrong type |
| 400    | `invalid_params` | A value breaks one of the rules above                  |
| 500    | `internal`       | Unexpected server error                                |

### GET /fizzbuzz/statistics

```bash
curl -s localhost:8081/fizzbuzz/statistics
```

```json
{"data":{"params":{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"buzz"},"hits":1}}
```

Before any request, it returns `{"data":{"params":null,"hits":0}}`.

## Development

The [`Makefile`](Makefile) provides shortcuts for the commands below (`make run`, `make test`, `make race`, `make build`, `make lint`, ...). Run `make check` to run everything CI checks before pushing.

```bash
### Run
go run ./cmd/server

### Test
go test ./...

### Race detector
go test -race ./...

### Build
go build -o bin/server ./cmd/server

### Lint
test -z "$(gofmt -l .)"
go vet ./...
go mod tidy -diff
golangci-lint run   # v2.13, same version as CI
```

CI runs the same checks on every push and pull request on main (see [`.github/workflows/ci.yaml`](.github/workflows/ci.yaml)).

## Assumptions & Limitations

- Parameters are sent as a JSON body to `POST /fizzbuzz`, not as query parameters on `GET`. A GET would fit a read-only computation, but this endpoint has a side effect: every call updates the request statistics. GET responses can be cached by a load balancer or a cache such as nginx, and cached hits never reach the server, so the statistics would undercount. POST makes the side effect explicit and is not cached by default.
- `int1` and `int2` may be equal, which will just print `str1str2`.
- Inputs are bounded to protect the server from very large requests (CPU, memory, response size). `limit` is capped by `MAX_LIMIT` and `str1`/`str2` lengths by `MAX_STR_LENGTH` (see [Configuration](#configuration)).
- Two requests are "the same" when all five parameters are equal. Statistics count identical inputs, with no normalization: `{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"fizz"}` and `{"int1":5,"int2":3,"limit":15,"str1":"fizz","str2":"fizz"}` are counted as two different requests, even though the sequence output will look like the same.
- Only valid requests are counted in the statistics.
- When several requests share the highest count, any one of them may be returned.
- Statistics do not need to survive a restart.
- Unknown query parameters on `GET` endpoints are ignored.

## Architecture & Project Layout

```
.
├── api/openapi.yaml              # HTTP contract (OpenAPI 3.0)
├── cmd/server/main.go            # entry point: config, wiring, HTTP server, shutdown
├── internal/config/               # loads and validates environment variables
├── internal/fizzbuzz/
│   ├── domain/                   # business types, rules and interfaces (no dependencies)
│   ├── service/                  # use cases: generate a sequence, read statistics
│   ├── handler/                  # HTTP layer: decode, call service, encode JSON
│   └── infrastructure/repository # statistics storage (in-memory)
├── docs/                         # roadmap and design notes
└── Dockerfile
```

The layout follows the official Go guide [Organizing a Go module](https://go.dev/doc/modules/layout): commands under `cmd/`, private packages under `internal/`.

Code tries to be maintainable and easy to understand. If you want to know more in detail about the architecture, patterns, please check: [`docs/architecture.md`](docs/architecture.md)

## Contributions

To contribute to this project, see [`docs/project_evolution.md`](docs/project_evolution.md).

There you will find our desired way to extend the service, new repositories, events and queues, cloud providers, etc.

## Next Steps

You can check the listed future improvements at [`docs/next_steps.md`](docs/next_steps.md).

## License

[MIT](LICENSE)

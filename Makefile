# Thin aliases over the go toolchain. Every recipe can be copied and run
# without make; see the Development section of the README.

.PHONY: run test race build fmt vet lint tidy check clean

run:
	go run ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

build:
	go build -o bin/server ./cmd/server

# Fails if any file is not gofmt-formatted (gofmt -l alone always exits 0).
fmt:
	test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet:
	go vet ./...

# Requires golangci-lint v2.13 (the version pinned in CI).
lint:
	golangci-lint run

tidy:
	go mod tidy -diff

# Everything CI checks.
check: tidy fmt vet lint race build

clean:
	rm -rf bin/

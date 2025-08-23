.PHONY: help test test-verbose build clean lint e2e all fmt test-coverage

# Default target
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests"
	@echo "  test-quick   - Run quick tests only (no downloads)"
	@echo "  test-verbose - Run all tests with verbose output"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  fmt          - Format code with go fmt and goimports"
	@echo "  build        - Build the binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  lint         - Run golangci-lint"
	@echo "  e2e          - Run end-to-end tests"
	@echo "  all          - Run test, lint, and build"

# Run all tests
test:
	go test ./...

# Run quick tests only (no downloads)
test-quick:
	go test ./... -run "Test.*Quick|Test.*Detection" -v

# Run all tests with verbose output
test-verbose:
	go test -v ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Build the binary
build:
	go build -o updatecursor ./cmd/updatecursor

# Clean build artifacts
clean:
	rm -f updatecursor
	go clean -cache

# Run linting (requires golangci-lint)
lint:
	golangci-lint run

# Run end-to-end tests (CLI integration tests)
e2e:
	go test -v ./internal/cli

# Run all checks: test, lint, and build
all: test lint build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run with race detection
test-race:
	go test -race ./...

# Run with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n1
	go tool cover -html=coverage.out

# Build for release (stripped binary)
build-release:
	go build -ldflags="-s -w" -o updatecursor ./cmd/updatecursor

# Cross-compile for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o updatecursor-linux-amd64 ./cmd/updatecursor

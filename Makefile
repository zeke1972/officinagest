.PHONY: build run clean test fmt vet lint install help

# Variables
BINARY_NAME=officina
GO=go
GOFLAGS=-v

# Build info
VERSION?=2.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# LDFLAGS for version info
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete!"

## run: Run the application
run: build
	@./$(BINARY_NAME)

## install: Install the application to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME) v$(VERSION)..."
	$(GO) install $(LDFLAGS) .
	@echo "Install complete!"

## clean: Remove build artifacts and temporary files
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f debug.log
	@rm -f officina.db
	@$(GO) clean
	@echo "Clean complete!"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete!"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete!"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## mod-tidy: Tidy and verify dependencies
mod-tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy
	$(GO) mod verify
	@echo "Modules tidy complete!"

## mod-download: Download dependencies
mod-download:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded!"

## check: Run fmt, vet, and test
check: fmt vet test
	@echo "All checks passed!"

## dev: Quick development cycle (fmt + build + run)
dev: fmt build run

## release: Build optimized binary for release
release:
	@echo "Building release binary..."
	$(GO) build $(LDFLAGS) -trimpath -o $(BINARY_NAME) .
	@echo "Release build complete!"

## info: Show build information
info:
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Version:     $(VERSION)"
	@echo "Build Time:  $(BUILD_TIME)"
	@echo "Commit:      $(COMMIT)"
	@echo "Go Version:  $(shell $(GO) version)"

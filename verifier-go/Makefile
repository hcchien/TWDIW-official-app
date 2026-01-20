.PHONY: help build test run clean install dev api-server

# Default target
help:
	@echo "Taiwan Digital Wallet - Go Implementation"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build API server binary"
	@echo "  make test        - Run all tests"
	@echo "  make run         - Start API server"
	@echo "  make dev         - Run API server in development mode"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make install     - Install dependencies"
	@echo "  make api-server  - Build and run API server"
	@echo ""
	@echo "Test targets:"
	@echo "  make test-vp        - Test VP service"
	@echo "  make test-oidvp     - Test OID4VP service"
	@echo "  make test-issuer    - Test issuer service"
	@echo "  make test-coverage  - Run tests with coverage report"

# Build the API server
build:
	@echo "Building API server..."
	@go build -o bin/api-server ./cmd/api-server
	@echo "Build complete: bin/api-server"

# Build with optimizations for production
build-prod:
	@echo "Building API server (production)..."
	@go build -ldflags="-s -w" -o bin/api-server ./cmd/api-server
	@echo "Production build complete: bin/api-server"

# Run all tests
test:
	@echo "Running all tests..."
	@go test ./... -v

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Test specific packages
test-vp:
	@echo "Testing VP service..."
	@go test ./pkg/vp -v

test-oidvp:
	@echo "Testing OID4VP service..."
	@go test ./pkg/oidvp -v

test-issuer:
	@echo "Testing issuer service..."
	@go test ./issuer-go/pkg/credential -v

# Run the API server
run: build
	@echo "Starting API server..."
	@./bin/api-server

# Development mode (no build, direct run)
dev:
	@echo "Running API server in development mode..."
	@go run ./cmd/api-server

# Start API server on custom port
run-port:
	@echo "Starting API server on port $(PORT)..."
	@PORT=$(PORT) ./bin/api-server

# Build and run
api-server: build run

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed"

# Quick development cycle
quick: fmt build run

# Show current status
status:
	@echo "Taiwan Digital Wallet - Status"
	@echo ""
	@echo "Go version:"
	@go version
	@echo ""
	@echo "Module info:"
	@go list -m
	@echo ""
	@echo "Test count:"
	@go test ./... -v 2>&1 | grep -c "^--- PASS" || true
	@echo ""
	@echo "Build status:"
	@if [ -f bin/api-server ]; then echo "✓ API server binary exists"; else echo "✗ API server not built"; fi

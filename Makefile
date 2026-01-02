.PHONY: build build-arm64 clean deploy test

# Build the binary for Lambda (Linux AMD64)
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip function.zip bootstrap

# Build the binary for Lambda (Linux ARM64/Graviton2)
build-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip function.zip bootstrap

# Build locally for testing
build-local:
	go build -o ec2_manager main.go

# Clean build artifacts
clean:
	rm -f bootstrap ec2_manager ec2_manager.zip ec2_manager-arm64.zip function.zip coverage.out coverage.html

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Check code formatting
fmt:
	go fmt ./...

# Run linter (requires golangci-lint to be installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Display help
help:
	@echo "Available targets:"
	@echo "  build          - Build Lambda deployment package for AMD64 (bootstrap binary + zip)"
	@echo "  build-arm64    - Build Lambda deployment package for ARM64/Graviton2 (bootstrap binary + zip)"
	@echo "  build-local    - Build binary for local testing"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  deps           - Install and tidy dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  help           - Display this help message"

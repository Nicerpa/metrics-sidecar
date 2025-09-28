# Variables
BINARY_NAME=metrics-sidecard
BUILD_DIR=bin
MAIN_PATH=./main.go
SERVER_PATH=./cmd/server

# Build the application (main.go)
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build the sidecard server
build-server:
	go build -o $(BUILD_DIR)/$(BINARY_NAME)-server $(SERVER_PATH)

# Run the application (main.go)
run:
	go run $(MAIN_PATH)

# Run the sidecard server with example parameters
run-server:
	go run $(SERVER_PATH) -listen-port 8080 -proxy-port 3000

# Run the example test server
run-test-server:
	go run examples/test-server.go

# Show help for the sidecard server
help-server:
	go run $(SERVER_PATH) -help

# Demo: run both test server and sidecard (requires two terminals)
demo:
	@echo "Starting test server on :3000..."
	@echo "In another terminal, run: make run-server"
	@echo "Then test with: curl http://localhost:8080/"
	go run examples/test-server.go

# Test the application
test:
	go test -v ./...

# Test with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	go clean
	rm -rf $(BUILD_DIR)

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Build sidecard for multiple platforms
build-server-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-server-linux-amd64 $(SERVER_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-server-darwin-amd64 $(SERVER_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-server-windows-amd64.exe $(SERVER_PATH)

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the basic application"
	@echo "  build-server  - Build the sidecard server"
	@echo "  run           - Run the basic application"
	@echo "  run-server    - Run the sidecard server (proxy-port 3000)"
	@echo "  run-test-server- Run example test server"
	@echo "  help-server   - Show sidecard server help"
	@echo "  demo          - Run demo setup instructions"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  deps          - Install dependencies"
	@echo "  build-all     - Build basic app for multiple platforms"
	@echo "  build-server-all - Build sidecard for multiple platforms"
	@echo "  help          - Show this help message"

.PHONY: build run test test-coverage clean fmt lint deps build-all help
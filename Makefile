.PHONY: build install test clean build-all fmt lint

BIN_DIR=bin
BINARY_NAME=autotitle
BUILD_CMD=go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/autotitle

# Build the CLI binary
build:
	@mkdir -p $(BIN_DIR)
	$(BUILD_CMD)
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all:
	@mkdir -p $(BIN_DIR)
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/autotitle-linux-amd64 ./cmd/autotitle
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/autotitle-windows-amd64.exe ./cmd/autotitle
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/autotitle-darwin-amd64 ./cmd/autotitle
	@echo "Done - Binaries are in $(BIN_DIR)/"

# Install the CLI to GOPATH/bin
install:
	go install ./cmd/autotitle

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	rm -f autotitle
	rm -f test/*.mkv
	rm -rf test/_backup

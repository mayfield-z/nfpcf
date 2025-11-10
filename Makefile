.PHONY: all build clean run

BINARY_NAME=nfpcf
BUILD_DIR=bin

all: build

build:
	@echo "Building NFPCF..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

run: build
	@echo "Running NFPCF..."
	./$(BUILD_DIR)/$(BINARY_NAME) -c ./config/nfpcfcfg.yaml

mod:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies ready"

test:
	@echo "Running tests..."
	@go test -v ./...

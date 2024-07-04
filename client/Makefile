# Define variables
BINARY_NAME=client
SRC_DIR=client
MAIN_FILE=$(SRC_DIR)/main.go

# Default target to build, vet, format, and run
.PHONY: all
all: build vet fmt test run

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@go build

# Run `go vet` to check for suspicious constructs
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run `go fmt` to format source code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v $(TEST_DIR) || (echo "Tests failed"; exit 1)

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

# Clean the build files
.PHONY: clean
clean:
	@echo "Cleaning $(BINARY_NAME) build files..."
	@rm -f $(BINARY_NAME)
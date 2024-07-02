# Define variables
BINARY_NAME=server
SRC_DIR=server
MAIN_FILE=$(SRC_DIR)/main.go

# Default target to build and run
.PHONY: all
all: build run

# Build the binary
.PHONY: build
build:
	@echo "Building the binary..."
	go build

# Run the application
.PHONY: run
run:
	@echo "Running the application..."
	./$(BINARY_NAME)

# Clean the build files
.PHONY: clean
clean:
	@echo "Cleaning the build files..."
	rm -f $(BINARY_NAME)

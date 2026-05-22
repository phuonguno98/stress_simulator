# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint
BINARY_NAME=stress_simulator
BINARY_UNIX=stress_simulator_unix

# Build the project
build: 
	$(GOBUILD) -o bin/$(BINARY_NAME) -v

# Run all tests
test:
	$(GOTEST) ./... -v

# Run tests with coverage
test-coverage:
	$(GOTEST) ./... -coverprofile=coverage.out -v
	$(GOTEST) -coverprofile=coverage.out -v
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	$(GOLINT) run

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html

# Run the application
run:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v
	./bin/$(BINARY_NAME)

.PHONY: build test test-coverage clean run lint
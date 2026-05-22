# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint
BINARY_NAME=stress_simulator
BINARY_UNIX=stress_simulator_unix

# Version information
VERSION ?= $(shell git describe --tags --always --dirty || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD || echo "none")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Build the project
build: 
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) -v

# Run all tests
test:
	$(GOTEST) -race -cover ./... -v

# Run tests with coverage
test-coverage:
	$(GOTEST) -race ./... -coverprofile=coverage.out -v
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	$(GOLINT) run

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_NAME).exe
	rm -f coverage.out
	rm -f coverage.html

# Run the application
run:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) -v
	./bin/$(BINARY_NAME)

.PHONY: build test test-coverage clean run lint
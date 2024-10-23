BINARY_NAME=admina-sysutils
VERSION=0.1.0
BUILD_DIR=bin
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build test clean lint vet fmt build-all deps

all: deps clean lint test build

build: clean
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/admina-sysutils

test:
	go test -v -race -cover ./...

clean:
	go clean
	rm -rf $(BUILD_DIR)

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

vet:
	go vet ./...

fmt:
	go fmt ./...

# Cross-compilation
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v ./cmd/admina-sysutils
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v ./cmd/admina-sysutils
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v ./cmd/admina-sysutils

# Update dependencies
deps:
	go mod tidy
	go mod verify

BINARY_NAME=admina-sysutils
VERSION=0.1.0
BUILD_DIR=bin

.PHONY: all build test clean

all: clean test build

build: clean
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/admina-sysutils

test:
	go test -v ./...

clean:
	go clean
	rm -rf $(BUILD_DIR)

# クロスコンパイル
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v ./cmd/admina-sysutils
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v ./cmd/admina-sysutils
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v ./cmd/admina-sysutils

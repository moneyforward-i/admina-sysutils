BINARY_NAME=admina-sysutils
VERSION=0.1.0
BUILD_DIR=bin
COVERAGE_DIR=coverage
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build test clean lint vet fmt build-all deps test-coverage test-verbose test-race

all: deps clean lint test build

build: clean
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/admina-sysutils

# 基本的なテスト実行
test:
	go test ./...

# 詳細なテスト出力
test-verbose:
	go test -v ./...

# レースコンディションチェック付きテスト
test-race:
	go test -race ./...

# カバレッジレポート付きテスト
test-coverage:
	mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out

# 全テストの実行（カバレッジ、レースチェック含む）
test-all: test-race test-coverage

clean:
	go clean
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

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

# CI用のテストターゲット
test-ci: lint vet test-all

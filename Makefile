SHELL := /bin/bash

BINARY_NAME=admina-sysutils
VERSION=0.1.0
BUILD_DIR=bin
OUT_DIR=out
COVERAGE_DIR=$(OUT_DIR)/test
REPORT_DIR=$(OUT_DIR)/test
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
GOBIN := $(shell go env GOBIN)
GOPATH := $(shell go env GOPATH)
PATH := $(GOBIN):$(GOPATH)/bin:$(PATH)

.PHONY: all build test clean lint vet fmt build-all deps test-ci build-cd dev test-e2e test-e2e-identity

## シチュエーションごとのコマンド
# CI用のテストターゲット
test-ci: clean deps lint vet test

# CD用のビルドターゲット
build-cd: clean deps build-all

# ローカル開発用のターゲット
dev: all

## E2Eテスト関連のコマンド
# test-e2e: すべてのE2Eテストを実行します
test-e2e: test-e2e-identity

# test-e2e-identity: Identity関連のE2Eテストを実行します
test-e2e-identity:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please create .env file with required environment variables."; \
		exit 1; \
	fi
	mkdir -p $(COVERAGE_DIR) $(REPORT_DIR)
	set -a && source .env && set +a && E2E_TEST=1 go test -v -coverprofile=$(COVERAGE_DIR)/e2e_identity_coverage.out ./internal/identity -run TestE2E_Identity | tee >(go-junit-report > $(REPORT_DIR)/e2e_identity_report.xml)
	npx xunit-viewer --results=$(REPORT_DIR)/e2e_identity_report.xml --output=$(REPORT_DIR)/e2e_identity_report.html
	go tool cover -html=$(COVERAGE_DIR)/e2e_identity_coverage.out -o $(COVERAGE_DIR)/e2e_identity_coverage.html
	go tool cover -func=$(COVERAGE_DIR)/e2e_identity_coverage.out

## 基本コマンド
# all: すべてのビルド、テスト、静的解析を実行します。
all: fmt deps clean lint vet test build

# build: プロジェクトのビルドを行います。
build: clean
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/admina-sysutils

# test: テストを実行し、カバレッジレポートとJUnit形式のテストレポートを生成します。
test:
	mkdir -p $(COVERAGE_DIR) $(REPORT_DIR)
	go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./... | tee >(go-junit-report > $(REPORT_DIR)/report.xml)
	npx xunit-viewer --results=$(REPORT_DIR)/report.xml --output=$(REPORT_DIR)/report.html
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out

# clean: ビルド成果物や中間ファイルを削除します。
clean:
	go clean
	rm -rf $(BUILD_DIR)
	rm -rf $(OUT_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -rf $(REPORT_DIR)

# lint: コードの静的解析を行います。
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# vet: コードの潜在的な問題を検出します。
vet:
	go vet ./...

# fmt: コードのフォーマットを行います。
fmt:
	go fmt ./...

# build-all: クロスコンパイルを行い、異なるプラットフォーム向けにバイナリを生成します。
build-all:
	mkdir -p $(BUILD_DIR)
	mkdir -p $(OUT_DIR)/bin
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/admina/$(BINARY_NAME) -v ./cmd/admina-sysutils
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin/admina/$(BINARY_NAME) -v ./cmd/admina-sysutils
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows/admina/$(BINARY_NAME).exe -v ./cmd/admina-sysutils
	# PowerShellスクリプトをUTF-16LEに変換してコピー
	nkf -w16L0 scripts/windows/samemerge.ps1 > $(BUILD_DIR)/windows/admina/samemerge.ps1
	cd $(BUILD_DIR)/linux && zip -r ../../$(OUT_DIR)/bin/$(BINARY_NAME)-linux-amd64.zip admina
	cd $(BUILD_DIR)/darwin && zip -r ../../$(OUT_DIR)/bin/$(BINARY_NAME)-darwin-amd64.zip admina
	cd $(BUILD_DIR)/windows && zip -r ../../$(OUT_DIR)/bin/$(BINARY_NAME)-windows-amd64.zip admina

# deps: 依存関係を更新します。
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/jstemmer/go-junit-report@latest
	npm install xunit-viewer
	go mod tidy
	go mod verify


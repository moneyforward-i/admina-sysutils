//go:build tools
// +build tools

package tools

// このファイルは開発ツールの依存関係を管理するためのものです
// 実際のアプリケーションコードでは使用されません

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

// このファイルは空でも問題ありません
// インポートされたパッケージは go.mod に記録されます

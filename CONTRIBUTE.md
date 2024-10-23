# 貢献ガイド

Admina SysUtils プロジェクトへの貢献に興味をお持ちいただき、ありがとうございます。以下は、プロジェクトに貢献する際のガイドラインです。

## 開発環境のセットアップ

### 要件

- Go 1.20 以上

### 開発中の実行

開発中にプログラムを実行するには、以下のコマンドを使用します：

go run ./cmd/admina-sysutils/main.go

### ビルド

ローカル環境用にビルドするには：

make build

すべてのプラットフォーム（Windows、Mac、Linux）用にビルドするには：

make build-all

### ビルド出力の場所

- ローカルビルド：プロジェクトの ./bin ディレクトリに admina-sysutils バイナリ（Windows の場合は admina-sysutils.exe）が生成されます。
- クロスコンパイル：プロジェクトの ./bin ディレクトリに以下のファイルが生成されます：
  - Linux: admina-sysutils-linux-amd64
  - Mac: admina-sysutils-darwin-amd64
  - Windows: admina-sysutils-windows-amd64.exe

### ビルドしたファイルの実行

ビルドしたファイルを実行するには、ターミナルで以下のコマンドを使用します：

Linux と Mac の場合：

./bin/admina-sysutils

Windows の場合：

.\bin\admina-sysutils.exe

クロスコンパイルされたファイルの場合は、ファイル名を適宜置き換えてください。

### テスト

テストを実行するには：

make test

## コーディング規約

- Go の標準的なコーディング規約に従ってください。
- gofmt を使用してコードをフォーマットしてください。
- golangci-lint を使用して静的解析を行ってください。

## プルリクエストのプロセス

1. 新しい機能やバグ修正のためのブランチを作成します。
2. 変更を加え、適切なテストを追加します。
3. すべてのテストが通過することを確認します。
4. プルリクエストを作成し、変更内容を詳細に説明してください。

## 問題の報告

バグを見つけた場合や新機能のアイデアがある場合は、GitHub の Issue を作成してください。

## ライセンス

このプロジェクトに貢献することで、あなたの貢献が Apache License 2.0 の下でライセンスされることに同意したものとみなされます。

# Admina SysUtils

Admina SysUtils は管理タスクを自動化するためのコマンドラインツールです。

## インストール

```
go install github.com/yourusername/admina-sysutils/cmd/admina-sysutils@latest
```

## 使用方法

```
admina-sysutils --help
```

## 開発

### 必要条件

- Go 1.20 以上

### 開発時の実行方法

開発中にプログラムを実行するには、以下のコマンドを使用します：

```
go run ./cmd/admina-sysutils/main.go
```

### ビルド方法

ローカル環境用にビルドするには：

```
make build
```

すべてのプラットフォーム（Windows、Mac、Linux）用にビルドするには：

```
make build-all
```

### ビルドファイルの出力先

- ローカルビルド: プロジェクトの`bin`ディレクトリに`admina-sysutils`（Windows の場合は`admina-sysutils.exe`）が生成されます。
- クロスコンパイル: プロジェクトの`bin`ディレクトリに以下のファイルが生成されます：
  - Linux: `admina-sysutils-linux-amd64`
  - Mac: `admina-sysutils-darwin-amd64`
  - Windows: `admina-sysutils-windows-amd64.exe`

### ビルドしたファイルの実行方法

ビルドしたファイルを実行するには、ターミナルで以下のコマンドを使用します：

# Linux および Mac の場合

```
./bin/admina-sysutils
```

# Windows の場合

```
.\bin\admina-sysutils.exe
```

クロスコンパイルしたファイルの場合は、ファイル名を適切に置き換えてください。

### テスト

```
make test
```

## ライセンス

MIT ライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルを参照してください。

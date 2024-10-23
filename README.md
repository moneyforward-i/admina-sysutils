# Admina SysUtils

Admina SysUtils は、管理タスクを自動化するためのコマンドラインツールです。

## インストール

以下のコマンドを使用してインストールできます：

go install github.com/moneyforward-i/admina-sysutils/cmd/admina-sysutils@latest

## 使用方法

基本的な使用方法は以下の通りです：

admina-sysutils [グローバルオプション] <コマンド> [サブコマンド]

グローバルオプション：

- --help: ヘルプを表示
- --debug: デバッグモードを有効化
- --output <format>: 出力フォーマットを指定（json, markdown, pretty）

## サポートされているコマンド

Admina SysUtils は以下のコマンドをサポートしています：

| コマンド | サブコマンド | オプション | 説明                                     |
| -------- | ------------ | ---------- | ---------------------------------------- |
| identity | matrix       | なし       | 組織のアイデンティティマトリックスを表示 |

注: グローバルオプション（--help, --debug, --output）はすべてのコマンドで使用可能です。

## 設定

Admina SysUtils を使用するには、以下の環境変数を設定する必要があります：

- `ADMINA_ORGANIZATION_ID`: あなたの組織 ID
- `ADMINA_API_KEY`: API キー

オプションで以下の環境変数も設定できます：

- `ADMINA_BASE_URL`: API のベース URL（デフォルトは https://api.itmc.i.moneyforward.com/api/v1）

## 例

アイデンティティマトリックスを JSON 形式で表示：

> admina-sysutils --output pretty identity matrix

## ライセンス

このプロジェクトは Apache License 2.0 の下でリリースされています。詳細は LICENSE ファイルを参照してください。

## 貢献

プロジェクトへの貢献に興味がある場合は、CONTRIBUTE.md を参照してください。

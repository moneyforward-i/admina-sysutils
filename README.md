# Admina SysUtils

[![CI](https://github.com/moneyforward-i/admina-sysutils/actions/workflows/test.yml/badge.svg)](https://github.com/moneyforward-i/admina-sysutils/actions/workflows/test.yml)
[![Latest Release](https://img.shields.io/github/v/release/moneyforward-i/admina-sysutils?include_prereleases)](https://github.com/moneyforward-i/admina-sysutils/releases/latest)

Admina SysUtils は、管理タスクを自動化するためのコマンドラインツールです。

## インストール

最新バージョンは [こちら](https://github.com/moneyforward-i/admina-sysutils/releases) から取得してください。

以下は各 OS でのセットアップ方法です：

### Windows

1. リリースページから最新の Windows 用バイナリをダウンロードします。
2. ダウンロードしたファイルをパスの通った場所に配置します。（またはパスを通します）

詳細なセットアップ手順とタスクスケジューラーの設定方法については、[Windows セットアップガイド](docs/windows-setup-guide.md)を参照してください。

### Mac

1. リリースページから最新の Mac 用バイナリをダウンロードします。
2. ダウンロードしたファイルをパスの通った場所に配置します。（またはパスを通します）

## 使用方法

基本的な使用方法は以下の通りです：

admina-sysutils [グローバルオプション] <コマンド> [サブコマンド] [オプション]

グローバルオプション：

- --help: ヘルプを表示
- --debug: デバッグモードを有効化
- --output <format>: 出力フォーマットを指定（json, markdown, pretty）

## サポートされているコマンド

Admina SysUtils は以下のコマンドをサポートしています：

| コマンド | サブコマンド | オプション                             | 必須 | デフォルト値 | 説明                                     | サンプル                                          |
| -------- | ------------ | -------------------------------------- | ---- | ------------ | ---------------------------------------- | ------------------------------------------------- |
| identity | matrix       | --output format (json/markdown/pretty) |      | pretty       | 組織のアイデンティティマトリックスを表示 | --output pretty                                   |
| identity | samemerge    | --output format (json/markdown/pretty) |      | pretty       | 出力フォーマットを指定                   | --output json                                     |
|          |              | --parent-domain << domain >>           | ◯    | -            | 親ドメインを指定                         | --parent-domain example.com                       |
|          |              | --child-domains << domains >>          | ◯    | -            | 子ドメインをカンマ区切りで指定           | --child-domains sub1.example.com,sub2.example.com |
|          |              | --dry-run                              |      | false        | 実際のマージを実行せずに確認のみ         | --dry-run                                         |
|          |              | --y                                    |      | false        | 確認プロンプトをスキップ                 | --y                                               |
|          |              | --nomask                               |      | false        | メールアドレスをマスクしない             | --nomask                                          |
|          |              | --outdir << path >>                    |      | ./out        | 出力ディレクトリのパスを指定             | --outdir /path/to/output                          |
| identity | help         | なし                                   |      | -            | アイデンティティコマンドのヘルプを表示   | identity help                                     |

## 設定

Admina SysUtils を使用するには、以下の環境変数を設定する必要があります：

- `ADMINA_ORGANIZATION_ID`: あなたの組織 ID
- `ADMINA_API_KEY`: API キー

オプションで以下の環境変数も設定できます：

- `ADMINA_CLI_ROOT`: プロジェクトのルートディレクトリ. この環境変数が設定されていない場合、バイナリを呼び出しているディレクトリをプロジェクトルートとして使用します。
- `ADMINA_BASE_URL`: API のベース URL（デフォルトは https://api.itmc.i.moneyforward.com/api/v1）
- `HTTPS_PROXY`/`HTTP_PROXY`: プロキシサーバーを経由して API にアクセスする場合に設定（例: http://proxy.example.com:8080）

## 出力ファイル

### `samemerge`コマンド

`samemerge`コマンドを実行すると、以下のファイルが出力されます：

出力ディレクトリ構造：

```
<outdir>/
└── data-{timestamp}/          # タイムスタンプ付きのディレクトリ
    ├── merge_candidates.csv   # マージ候補のリスト
    └── unmapped.csv          # マッピングされなかったアイデンティティ
```

### 出力ファイル一覧

| ファイル名           | 説明                                             | 出力項目                                                                                                                                                                             |
| -------------------- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| merge_candidates.csv | マージ候補となるアイデンティティのペアとその状態 | ・親アイデンティティ情報（メールアドレス、ID 等）<br>・子アイデンティティ情報（メールアドレス、ID 等）<br>・マージ状態（成功、スキップ、エラー）<br>・理由（スキップやエラーの場合） |
| unmapped.csv         | マッピングされなかったアイデンティティの一覧     | ・メールアドレス<br>・管理タイプ<br>・その他の属性情報                                                                                                                               |

注意：

- デフォルトの出力先は`<current>/out/`です
- CLI 自体はこれらのローテーションを行いません。ラップしたスクリプトで管理してください。

## 例

### アイデンティティマトリックスを表示する例：

#### Pretty 形式で表示

> ./admina-sysutils --output pretty identity matrix

![identity matrix](./img/identity_matrix_command.gif)

### 同一メールアドレスのマージ例：

#### ドライランでマージ候補を確認

> ./admina-sysutils identity samemerge --parent-domain example.com --child-domains sub1.example.com,sub2.example.com --dry-run

![samemerge](./img/identity_merge_command.gif)

#### 確認プロンプトなしで実行

> ./admina-sysutils identity samemerge --parent-domain example.com --child-domains sub1.example.com,sub2.example.com -y

#### メールアドレスをマスクせずに JSON 形式で出力

> ./admina-sysutils identity samemerge --parent-domain example.com --child-domains sub1.example.com,sub2.example.com --nomask --output json

#### カスタム出力ディレクトリを指定して実行

> ./admina-sysutils identity samemerge --parent-domain example.com --child-domains sub1.example.com,sub2.example.com --outdir /path/to/output

## 標準出力と標準エラー出力

Admina SysUtils のコマンドを実行する際、標準出力にはコマンドの結果が、標準エラー出力にはコマンドの実行ログが出力されます。

### 標準出力と標準エラー出力を別々のファイルに出力する例：

> ./admina-sysutils identity samemerge --parent-domain example.com --child-domains sub1.example.com,sub2.example.com --output json > result.json 2> log.txt

## ヘルプの表示例：

> ./admina-sysutils identity help

## ライセンス

このプロジェクトは Apache License 2.0 の下でリリースされています。詳細は LICENSE ファイルを参照してください。

## 貢献

プロジェクトへの貢献に興味がある場合は、CONTRIBUTE.md を参照してください。

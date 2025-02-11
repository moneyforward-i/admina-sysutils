# スクリプトの実行に必要な設定
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

#-----------------------------------
# ユーザー設定
#-----------------------------------
# 必須設定
$PARENT_DOMAIN = "parent.com"
$CHILD_DOMAINS = "child1.com,child2.com"

# オプション設定
# コメントアウトを外すと設定が有効になります
$DRY_RUN = $true        # --dry-run オプションを有効にする(デフォルト: DryRun実行しない)
#$MASK_EMAIL = $true    # メールアドレスをマスクする（デフォルト: マスクなし）
#$DEBUG = $true         # デバッグモードを有効にする（デフォルト: デバッグなし）

# 環境変数設定
# 環境変数で指定するか、以下の値を直接指定することもできます
#$env:ADMINA_ORGANIZATION_ID = "99999999"
#$env:ADMINA_API_KEY = "your-api-key"

#-----------------------------------
# システム設定（必要な場合のみ変更）
#-----------------------------------
# ログ設定
$KEEP_LOG_FILES = 45    # 保持するログファイル数
$TIMEOUT_HOURS = 6      # 実行タイムアウト時間

# イベントログ設定
$EVENT_SOURCE = "AdminaSysUtils"
$EVENT_LOG = "Application"
<#
Custom Application Event ID Recommended Ranges:
1000-1999: Information messages (正常系メッセージ)
2000-2999: Warning messages (警告メッセージ)
3000-3999: Error messages (エラーメッセージ)
4000-4999: Audit success (監査成功)
5000-5999: Audit failure (監査失敗)
Note: 0-999は Microsoft により予約されているため使用を避けること
#>
#-----------------------------------
# 内部処理用設定（変更不要）
#-----------------------------------
# パス設定
$SCRIPT_DIR = $PSScriptRoot
$ADMINA_EXE = Join-Path $SCRIPT_DIR "admina-sysutils.exe"
$LOG_DIR = Join-Path $SCRIPT_DIR "logs"
$OUT_DIR = Join-Path $SCRIPT_DIR "out"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$DATA_DIR = Join-Path $OUT_DIR "data-$timestamp"

# デフォルト値の設定
if (-not (Get-Variable -Name MASK_EMAIL -ErrorAction SilentlyContinue)) {
    $MASK_EMAIL = $false
}
if (-not (Get-Variable -Name DRY_RUN -ErrorAction SilentlyContinue)) {
    $DRY_RUN = $false
}
if (-not (Get-Variable -Name DEBUG -ErrorAction SilentlyContinue)) {
    $DEBUG = $false
}

# 環境変数のチェック
$requiredEnvVars = @(
    "ADMINA_API_KEY",
    "ADMINA_ORGANIZATION_ID"
)

foreach ($var in $requiredEnvVars) {
    if (-not (Get-Item env:$var -ErrorAction SilentlyContinue)) {
        Write-Error "必要な環境変数 ${var} が設定されていません。"
        exit 1
    }
}

# イベントログのソース登録確認とスクリプト起動ログを記録
try {
    # コマンドの構築
    $commandArgs = @()
    $modeInfo = @()

    # グローバルオプションを追加
    if ($DEBUG) {
        $commandArgs += "--debug"
        $modeInfo += "デバッグモード"
    }

    # サブコマンドとオプションを追加
    $commandArgs += @(
        "identity",
        "samemerge",
        "--parent-domain",
        $PARENT_DOMAIN,
        "--child-domains",
        $CHILD_DOMAINS,
        "--output",
        "pretty",
        "--outdir",
        $DATA_DIR,
        "--y"
    )

    if ($DRY_RUN) {
        $commandArgs += "--dry-run"
        $modeInfo += "DRY_RUNモード"
    }
    if (-not $MASK_EMAIL) {
        $commandArgs += "--nomask"
        $modeInfo += "マスクなし"
    }

    $startMessage = @"
同一人物マージ処理を開始します。
PARENT_DOMAIN: $PARENT_DOMAIN
CHILD_DOMAINS: $CHILD_DOMAINS
ORGANIZATION_ID: $env:ADMINA_ORGANIZATION_ID
出力ディレクトリ: $DATA_DIR
実行モード: $($modeInfo -join ", ")
"@
    Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $startMessage
}
catch {
    try {
        New-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE
        Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $startMessage
    }
    catch {
        Write-Error "イベントログのソース登録に失敗しました: $_"
        exit 1
    }
}

# ログディレクトリの作成
if (-not (Test-Path $LOG_DIR)) {
    New-Item -ItemType Directory -Path $LOG_DIR | Out-Null
}

# 出力ディレクトリの作成（存在しない場合）
if (-not (Test-Path $OUT_DIR)) {
    New-Item -ItemType Directory -Path $OUT_DIR | Out-Null
}

# 古いログファイルの削除（最新の45ファイルのみ保持）
Get-ChildItem $LOG_DIR -Filter "*.log" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip $KEEP_LOG_FILES |
    Remove-Item -Force

# 古い出力ディレクトリの削除（最新の45ディレクトリのみ保持）
Get-ChildItem $OUT_DIR -Directory |
    Where-Object { $_.Name -like "data-*" } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip $KEEP_LOG_FILES |
    Remove-Item -Recurse -Force

# タイムスタンプ付きのログファイル名を生成
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$logFile = Join-Path $LOG_DIR "samemerge_${timestamp}.log"

# 実行コマンドをログに出力
$commandLine = "$ADMINA_EXE $($commandArgs -join ' ')"
"実行コマンド: $commandLine" | Out-File -FilePath $logFile -Encoding Default

try {
    # CLIコマンドの実行（タイムアウト付き）
    $job = Start-Job -ScriptBlock {
        param($exePath, $commandArgs)
        Set-Location (Split-Path $exePath)
        & $exePath @commandArgs 2>&1
        $LASTEXITCODE
    } -ArgumentList $ADMINA_EXE, (,$commandArgs)

    $completed = Wait-Job $job -Timeout ($TIMEOUT_HOURS * 3600)

    # まず出力を取得して記録
    $output = Receive-Job $job
    if ($output) {
        $output | Out-File -FilePath $logFile -Encoding Default -Append
    }

    if ($completed) {
        $exitCode = $output[-1]  # 最後の要素が終了コード

        # 終了コードの確認
        if ($exitCode -eq 0) {
            # ログファイルの最後の40行を取得
            $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
            $successMessage = @"
同一人物マージ処理が正常に完了しました。
ログファイル: $logFile

ログ:
(...)
$lastLines
"@
            Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $successMessage
            Write-Output "同一人物マージ処理が正常に完了しました。"
        } else {
            # ログファイルの最後の40行を取得
            $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
            $errorMessage = @"
同一人物マージ処理でエラーが発生しました。終了コード: $exitCode
ログファイル: $logFile

ログy:
$lastLines
"@
            Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $errorMessage
            Write-Error "同一人物マージ処理でエラーが発生しました。終了コード: $exitCode。詳細はログファイルを確認してください。"
            exit $exitCode
        }
    } else {
        # ログファイルの最後の40行を取得
        $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
        $timeoutMessage = @"
${TIMEOUT_HOURS}時間を超えたため、処理を中断しました。
ログファイル: $logFile

ログ:
(...)
$lastLines
"@
        Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $timeoutMessage
        if ($job) {
            Stop-Job $job
            Remove-Job $job -Force
        }
        Write-Error "${TIMEOUT_HOURS}時間を超えたため、処理を中断しました。"
        exit 1
    }
}
catch {
    # ログファイルの最後の40行を取得（ログファイルが存在する場合）
    $lastLines = ""
    if (Test-Path $logFile) {
        $lastLines = "`n`nログ:
(...)`n" + (Get-Content -Path $logFile -Tail 40 | Out-String)
    }
    $errorMessage = @"
予期せぬエラーが発生しました: $_
ログファイル: $logFile$lastLines
"@
    Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $errorMessage
    Write-Error "予期せぬエラーが発生しました: $_"
    exit 1
}
finally {
    if ($job) {
        Remove-Job $job -Force -ErrorAction SilentlyContinue
    }
    # デバッグモードの環境変数をクリア
    if ($DEBUG) {
        Remove-Item Env:\ADMINA_DEBUG -ErrorAction SilentlyContinue
    }
}
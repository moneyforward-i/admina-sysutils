# Required settings for script execution
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

#-----------------------------------
# User settings
#-----------------------------------
# Required settings
$PARENT_DOMAIN = "parent.com"
$CHILD_DOMAINS = "child1.com,child2.com"

# Optional settings
# Uncomment to enable these settings
$DRY_RUN = $true        # Enable --dry-run option (Default: No DryRun)
#$MASK_EMAIL = $true    # Mask email addresses (Default: No masking)
#$DEBUG = $true         # Enable debug mode (Default: No debugging)

# Proxy settings
# Uncomment and modify the following lines if you need to use a proxy
#$env:HTTP_PROXY = "http://username:password@proxy.example.com:8080"
#$env:HTTPS_PROXY = "http://username:password@proxy.example.com:8080"

# Environment variables
# You can set these as environment variables or directly specify them below
#$env:ADMINA_ORGANIZATION_ID = "99999999"
#$env:ADMINA_API_KEY = "your-api-key"

#-----------------------------------
# System settings (change only if necessary)
#-----------------------------------
# Log settings
$KEEP_LOG_FILES = 45    # Number of log files to retain
$TIMEOUT_HOURS = 6      # Execution timeout in hours

# Event log settings
$EVENT_SOURCE = "AdminaSysUtils"
$EVENT_LOG = "Application"
<#
Custom Application Event ID Recommended Ranges:
1000-1999: Information messages
2000-2999: Warning messages
3000-3999: Error messages
4000-4999: Audit success
5000-5999: Audit failure
Note: Avoid using 0-999 as they are reserved by Microsoft
#>
#-----------------------------------
# Internal processing settings (do not change)
#-----------------------------------
# Path settings
$SCRIPT_DIR = $PSScriptRoot
$ADMINA_EXE = Join-Path $SCRIPT_DIR "admina-sysutils.exe"
$LOG_DIR = Join-Path $SCRIPT_DIR "logs"
$OUT_DIR = Join-Path $SCRIPT_DIR "out"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$DATA_DIR = Join-Path $OUT_DIR "data-$timestamp"

# Setting default values
if (-not (Get-Variable -Name MASK_EMAIL -ErrorAction SilentlyContinue)) {
    $MASK_EMAIL = $false
}
if (-not (Get-Variable -Name DRY_RUN -ErrorAction SilentlyContinue)) {
    $DRY_RUN = $false
}
if (-not (Get-Variable -Name DEBUG -ErrorAction SilentlyContinue)) {
    $DEBUG = $false
}

# Validate and fix proxy URL format
function Validate-ProxyUrl {
    param (
        [string]$proxyUrl
    )

    if ([string]::IsNullOrEmpty($proxyUrl)) {
        return $null
    }

    # Check if URL format is correct when it contains @ symbol
    if ($proxyUrl -match '@') {
        try {
            # Try to parse as a valid URL
            $uri = [System.Uri]::new($proxyUrl)
            return $proxyUrl
        } catch {
            # If invalid format, try to fix it
            Write-Warning "Invalid proxy URL format: $proxyUrl"

            # Basic format correction (expecting http://user:pass@host:port format)
            if ($proxyUrl -match '(http[s]?)://([^:]+):([^@]+)@(.+)') {
                $protocol = $matches[1]
                $username = [System.Web.HttpUtility]::UrlEncode($matches[2])
                $password = [System.Web.HttpUtility]::UrlEncode($matches[3])
                $hostPort = $matches[4]

                $fixedUrl = "${protocol}://${username}:${password}@${hostPort}"
                Write-Warning "Fixed proxy URL: $fixedUrl"
                return $fixedUrl
            }
        }
    }

    return $proxyUrl
}

# Validate and fix proxy environment variables
$httpProxy = $env:HTTP_PROXY
$httpsProxy = $env:HTTPS_PROXY

if (-not [string]::IsNullOrEmpty($httpProxy)) {
    $fixedHttpProxy = Validate-ProxyUrl -proxyUrl $httpProxy
    if ($fixedHttpProxy -ne $httpProxy) {
        $env:HTTP_PROXY = $fixedHttpProxy
        Write-Warning "Fixed HTTP_PROXY environment variable"
    }
}

if (-not [string]::IsNullOrEmpty($httpsProxy)) {
    $fixedHttpsProxy = Validate-ProxyUrl -proxyUrl $httpsProxy
    if ($fixedHttpsProxy -ne $httpsProxy) {
        $env:HTTPS_PROXY = $fixedHttpsProxy
        Write-Warning "Fixed HTTPS_PROXY environment variable"
    }
}

# Check environment variables
$requiredEnvVars = @(
    "ADMINA_API_KEY",
    "ADMINA_ORGANIZATION_ID"
)

foreach ($var in $requiredEnvVars) {
    if (-not (Get-Item env:$var -ErrorAction SilentlyContinue)) {
        Write-Error "Required environment variable ${var} is not set."
        exit 1
    }
}

# Check event log source registration and record script startup log
try {
    # Build command
    $commandArgs = @()
    $modeInfo = @()

    # Add global options
    if ($DEBUG) {
        $commandArgs += "--debug"
        $modeInfo += "Debug mode"
    }

    # Add subcommand and options
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
        $modeInfo += "DRY_RUN mode"
    }
    if (-not $MASK_EMAIL) {
        $commandArgs += "--nomask"
        $modeInfo += "No masking"
    }

    $startMessage = @"
Starting identity merge process.
PARENT_DOMAIN: $PARENT_DOMAIN
CHILD_DOMAINS: $CHILD_DOMAINS
ORGANIZATION_ID: $env:ADMINA_ORGANIZATION_ID
Output directory: $DATA_DIR
Execution mode: $($modeInfo -join ", ")
"@
    Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $startMessage
}
catch {
    try {
        New-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE
        Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $startMessage
    }
    catch {
        Write-Error "Failed to register event log source: $_"
        exit 1
    }
}

# Create log directory
if (-not (Test-Path $LOG_DIR)) {
    New-Item -ItemType Directory -Path $LOG_DIR | Out-Null
}

# Create output directory (if it doesn't exist)
if (-not (Test-Path $OUT_DIR)) {
    New-Item -ItemType Directory -Path $OUT_DIR | Out-Null
}

# Delete old log files (keep only the latest 45 files)
Get-ChildItem $LOG_DIR -Filter "*.log" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip $KEEP_LOG_FILES |
    Remove-Item -Force

# Delete old output directories (keep only the latest 45 directories)
Get-ChildItem $OUT_DIR -Directory |
    Where-Object { $_.Name -like "data-*" } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -Skip $KEEP_LOG_FILES |
    Remove-Item -Recurse -Force

# Generate log filename with timestamp
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$logFile = Join-Path $LOG_DIR "samemerge_${timestamp}.log"

# Output execution command to log
$commandLine = "$ADMINA_EXE $($commandArgs -join ' ')"
"Execution command: $commandLine" | Out-File -FilePath $logFile -Encoding Default

try {
    # Execute CLI command (with timeout)
    $job = Start-Job -ScriptBlock {
        param($exePath, $commandArgs)
        Set-Location (Split-Path $exePath)
        & $exePath @commandArgs 2>&1
        $LASTEXITCODE
    } -ArgumentList $ADMINA_EXE, (,$commandArgs)

    $completed = Wait-Job $job -Timeout ($TIMEOUT_HOURS * 3600)

    # First get and record the output
    $output = Receive-Job $job
    if ($output) {
        $output | Out-File -FilePath $logFile -Encoding Default -Append
    }

    if ($completed) {
        $exitCode = $output[-1]  # The last element is the exit code

        # Check exit code
        if ($exitCode -eq 0) {
            # Get the last 40 lines of the log file
            $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
            $successMessage = @"
Identity merge process completed successfully.
Log file: $logFile

Log:
(...)
$lastLines
"@
            Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Information -EventId 1000 -Message $successMessage
            Write-Output "Identity merge process completed successfully."
        } else {
            # Get the last 40 lines of the log file
            $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
            $errorMessage = @"
Error occurred during identity merge process. Exit code: $exitCode
Log file: $logFile

Log:
$lastLines
"@
            Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $errorMessage
            Write-Error "Error occurred during identity merge process. Exit code: $exitCode. Check log file for details."
            exit $exitCode
        }
    } else {
        # Get the last 40 lines of the log file
        $lastLines = Get-Content -Path $logFile -Tail 40 | Out-String
        $timeoutMessage = @"
Process terminated after exceeding ${TIMEOUT_HOURS} hours.
Log file: $logFile

Log:
(...)
$lastLines
"@
        Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $timeoutMessage
        if ($job) {
            Stop-Job $job
            Remove-Job $job -Force
        }
        Write-Error "Process terminated after exceeding ${TIMEOUT_HOURS} hours."
        exit 1
    }
}
catch {
    # Get the last 40 lines of the log file (if it exists)
    $lastLines = ""
    if (Test-Path $logFile) {
        $lastLines = "`n`nLog:
(...)`n" + (Get-Content -Path $logFile -Tail 40 | Out-String)
    }
    $errorMessage = @"
Unexpected error occurred: $_
Log file: $logFile$lastLines
"@
    Write-EventLog -LogName $EVENT_LOG -Source $EVENT_SOURCE -EntryType Error -EventId 3000 -Message $errorMessage
    Write-Error "Unexpected error occurred: $_"
    exit 1
}
finally {
    if ($job) {
        Remove-Job $job -Force -ErrorAction SilentlyContinue
    }
    # Clear debug mode environment variable
    if ($DEBUG) {
        Remove-Item Env:\ADMINA_DEBUG -ErrorAction SilentlyContinue
    }
}
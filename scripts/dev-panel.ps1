# Dev mode: panel hot-reload UI + auto rebuild/restart saat file Go berubah.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$env:ZAUTO_PANEL_DEV = "1"
$panel = Join-Path $Root "zautopanel.exe"

function Build-Panel {
    Write-Host "[dev] Building zautopanel.exe..."
    $env:CGO_ENABLED = "1"
    go build -tags "desktop,production" -ldflags "-w -s" -o $panel ./cmd/zautopanel
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Stop-Panel {
    $p = Get-Process -Name "zautopanel" -ErrorAction SilentlyContinue
    if ($p) {
        Write-Host "[dev] Stopping panel..."
        taskkill /F /IM zautopanel.exe 2>$null | Out-Null
        Start-Sleep -Milliseconds 700
    }
}

function Start-Panel {
    Write-Host "[dev] Starting panel..."
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $panel
    $psi.WorkingDirectory = $Root
    $psi.UseShellExecute = $false
    $psi.EnvironmentVariables["ZAUTO_PANEL_DEV"] = "1"
    [System.Diagnostics.Process]::Start($psi) | Out-Null
}

Build-Panel
Stop-Panel
Start-Panel

Write-Host ""
Write-Host "=== zauto Panel DEV ==="
Write-Host "  UI (html/css/js): simpan -> auto reload ~1 detik (badge DEV di footer)"
Write-Host "  Go/backend:       simpan *.go -> rebuild + restart otomatis"
Write-Host "  Manual:           F5 di panel"
Write-Host ""

$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = $Root
$watcher.IncludeSubdirectories = $true
$watcher.Filter = "*.go"
$watcher.EnableRaisingEvents = $true

Write-Host "Watching Go files — Ctrl+C to stop."
while ($true) {
    $null = $watcher.WaitForChanged([System.IO.WatcherChangeTypes]::Changed, 3600000)
    Start-Sleep -Milliseconds 1600
    try {
        Build-Panel
        Stop-Panel
        Start-Panel
    } catch {
        Write-Host "[dev] error: $_"
    }
}

# Start zauto Panel (native Wails app).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$panel = Join-Path $Root "zautopanel.exe"

function Build-Panel {
    Write-Host "Building zautopanel.exe (Wails tags desktop,production)..."
    $env:CGO_ENABLED = "1"
    go build -tags "desktop,production" -ldflags "-w -s" -o $panel ./cmd/zautopanel
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

if (-not (Test-Path $panel)) {
    Build-Panel
}

Start-Process -FilePath $panel -WorkingDirectory $Root

# Build zauto CLI + native Wails panel.
$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root

Write-Host "Building zauto.exe (CLI)..."
go build -o zauto.exe ./cmd/zauto
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Building adbwrap.exe (silent ADB)..."
go build -ldflags "-w -s -H windowsgui" -o adbwrap.exe ./cmd/adbwrap
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$env:CGO_ENABLED = "1"
Write-Host "Building zautopanel.exe (Wails)..."
go build -tags "desktop,production" -ldflags "-w -s" -o zautopanel.exe ./cmd/zautopanel
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "OK: zauto.exe + zautopanel.exe + adbwrap.exe"

# Start zauto: PostgreSQL + Panel native (Wails).
$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
Set-Location $Root

Write-Host "=== zauto ==="
Write-Host "[1/2] PostgreSQL..."
docker compose -f docker-compose.db.yml up -d
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[2/2] Panel (Wails)..."
& powershell -ExecutionPolicy Bypass -File "$Root\build.ps1"
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& "$Root\scripts\start-panel.ps1"

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$deploy = Join-Path $root 'deploy'
New-Item -ItemType Directory -Force -Path $deploy, (Join-Path $deploy 'data'), (Join-Path $deploy 'uploads') | Out-Null

Write-Host 'Building frontend...'
npm run build
if (-not (Test-Path 'dist\index.html')) { throw 'frontend build failed' }

$distDest = Join-Path $deploy 'dist'
if (Test-Path $distDest) { Remove-Item $distDest -Recurse -Force }
Copy-Item 'dist' $distDest -Recurse

Write-Host 'Building Windows backend...'
Push-Location 'backend-go'
$env:CGO_ENABLED = '0'
go build -ldflags '-s -w' -o (Join-Path $deploy 'travel-server.exe') .
Write-Host 'Building Linux amd64 backend...'
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
go build -ldflags '-s -w' -o (Join-Path $deploy 'travel-server') .
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Pop-Location

Get-ChildItem $deploy -File | Where-Object { $_.Name -match '^(install|start|start-caddy)\.sh$|^Caddyfile$' } | ForEach-Object {
  $text = [System.IO.File]::ReadAllText($_.FullName) -replace "`r`n", "`n" -replace "`r", "`n"
  [System.IO.File]::WriteAllText($_.FullName, $text)
}

Write-Host "Deploy folder ready: $deploy"

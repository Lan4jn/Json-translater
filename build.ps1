$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force dist | Out-Null

go mod tidy
go test ./...
go build -o dist/json2table-windows-amd64.exe .

$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o dist/json2table-linux-amd64 .

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host "Build complete:"
Write-Host "  dist/json2table-windows-amd64.exe"
Write-Host "  dist/json2table-linux-amd64"

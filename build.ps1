$ErrorActionPreference = "Stop"
Push-Location "src"

# One-time: go install github.com/akavel/rsrc@latest
go generate
New-Item -ItemType Directory -Force "../build" | Out-Null
go build -ldflags "-H=windowsgui" -o "../build/cybersaver.exe"

Pop-Location
Write-Host "Built build/cybersaver.exe"

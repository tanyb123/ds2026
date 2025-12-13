# PowerShell build script for Remote Shell RPC System

Write-Host "Building Remote Shell RPC System..." -ForegroundColor Green

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "Found: $goVersion" -ForegroundColor Cyan
} catch {
    Write-Host "ERROR: Go is not installed or not in PATH!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "After installation, restart PowerShell and try again." -ForegroundColor Yellow
    exit 1
}

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Cyan
}

# Build server
Write-Host "`nBuilding server..." -ForegroundColor Yellow
go build -o bin\server.exe ./server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build server" -ForegroundColor Red
    exit 1
}
Write-Host "Server built successfully!" -ForegroundColor Green

# Build client
Write-Host "`nBuilding client..." -ForegroundColor Yellow
go build -o bin\client.exe ./client
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build client" -ForegroundColor Red
    exit 1
}
Write-Host "Client built successfully!" -ForegroundColor Green

# Build admin tool
Write-Host "`nBuilding admin tool..." -ForegroundColor Yellow
go build -o bin\admin.exe ./admin
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build admin" -ForegroundColor Red
    exit 1
}
Write-Host "Admin tool built successfully!" -ForegroundColor Green

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green
Write-Host "Binaries are in the bin\ directory" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run:" -ForegroundColor Yellow
Write-Host "  Server:  .\bin\server.exe" -ForegroundColor White
Write-Host "  Client:  .\bin\client.exe -id my-client" -ForegroundColor White
Write-Host "  Admin:   .\bin\admin.exe" -ForegroundColor White
Write-Host ""





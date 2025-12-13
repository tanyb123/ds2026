# PowerShell script to run the server

if (-not (Test-Path "bin\server.exe")) {
    Write-Host "Server not found. Building..." -ForegroundColor Yellow
    .\build.ps1
}

Write-Host "Starting Remote Shell RPC Server..." -ForegroundColor Green
Write-Host "Press Ctrl+C to stop`n" -ForegroundColor Yellow
.\bin\server.exe





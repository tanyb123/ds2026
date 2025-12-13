# PowerShell script to run a client

param(
    [string]$ClientID = "",
    [string]$Command = ""
)

if (-not (Test-Path "bin\client.exe")) {
    Write-Host "Client not found. Building..." -ForegroundColor Yellow
    .\build.ps1
}

$args = @()
if ($ClientID) {
    $args += "-id", $ClientID
}
if ($Command) {
    $args += "-cmd", $Command
}

Write-Host "Starting Remote Shell RPC Client..." -ForegroundColor Green
.\bin\client.exe $args





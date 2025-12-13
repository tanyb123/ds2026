# Script to check if your computer has public IP

Write-Host "=== Checking Network Configuration ===" -ForegroundColor Cyan
Write-Host ""

# Get local IP addresses
Write-Host "Local IP Addresses:" -ForegroundColor Yellow
$localIPs = Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*"} | Select-Object -ExpandProperty IPAddress
foreach ($ip in $localIPs) {
    Write-Host "  - $ip" -ForegroundColor Green
}

Write-Host ""

# Get public IP
Write-Host "Checking Public IP..." -ForegroundColor Yellow
try {
    $publicIP = (Invoke-WebRequest -Uri "https://api.ipify.org" -UseBasicParsing).Content
    Write-Host "Public IP (from internet): $publicIP" -ForegroundColor Green
    Write-Host ""
    
    # Compare
    $hasPublicIP = $false
    foreach ($localIP in $localIPs) {
        if ($localIP -eq $publicIP) {
            $hasPublicIP = $true
            Write-Host "✓ Your computer HAS a public IP!" -ForegroundColor Green
            Write-Host "  Local IP matches Public IP: $localIP" -ForegroundColor Green
            break
        }
    }
    
    if (-not $hasPublicIP) {
        Write-Host "✗ Your computer is behind NAT/Router" -ForegroundColor Yellow
        Write-Host "  Local IP: $($localIPs[0])" -ForegroundColor Yellow
        Write-Host "  Public IP (Router): $publicIP" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "To allow external connections:" -ForegroundColor Cyan
        Write-Host "  1. Configure port forwarding on your router" -ForegroundColor White
        Write-Host "  2. Forward port 8080 to: $($localIPs[0])" -ForegroundColor White
        Write-Host "  3. Clients connect to: $publicIP:8080" -ForegroundColor White
    }
    
} catch {
    Write-Host "Could not check public IP. Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "You can check manually at: https://whatismyipaddress.com" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "=== Network Information ===" -ForegroundColor Cyan
Write-Host "Default Gateway:" -ForegroundColor Yellow
$gateway = (Get-NetRoute -DestinationPrefix "0.0.0.0/0" | Where-Object {$_.NextHop -ne "0.0.0.0"}).NextHop
if ($gateway) {
    Write-Host "  $gateway" -ForegroundColor Green
    Write-Host ""
    Write-Host "This is likely your router IP." -ForegroundColor Cyan
    Write-Host "If you have public IP, it should match the public IP above." -ForegroundColor Cyan
} else {
    Write-Host "  Not found" -ForegroundColor Yellow
}

Write-Host ""

# Run as Administrator
$setupc = "C:\Program Files (x86)\com0com\setupc.exe"

Write-Host "=== Current com0com configuration ===" -ForegroundColor Cyan
& $setupc list

Write-Host "`n=== Attempting to remove existing pair ===" -ForegroundColor Yellow
& $setupc remove 0

Write-Host "`n=== Creating new port pair COM10<->COM11 ===" -ForegroundColor Green
& $setupc install PortName=COM10 PortName=COM11

Write-Host "`n=== New configuration ===" -ForegroundColor Cyan
& $setupc list

Write-Host "`n=== Verifying ports are visible ===" -ForegroundColor Cyan
[System.IO.Ports.SerialPort]::getportnames() | Sort-Object

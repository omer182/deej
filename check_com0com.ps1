# Check if com0com is installed
Write-Host "=== Checking com0com Installation ===" -ForegroundColor Cyan

# Check Program Files
$com0comPaths = @(
    "C:\Program Files\com0com",
    "C:\Program Files (x86)\com0com",
    "$env:ProgramFiles\com0com",
    "${env:ProgramFiles(x86)}\com0com"
)

foreach ($path in $com0comPaths) {
    if (Test-Path $path) {
        Write-Host "Found com0com at: $path" -ForegroundColor Green
        Get-ChildItem $path | Select-Object Name, Length, LastWriteTime
    }
}

# Check for com0com drivers
Write-Host "`n=== Checking for com0com Drivers ===" -ForegroundColor Cyan
Get-PnpDevice | Where-Object {$_.FriendlyName -like "*com0com*"} | Select-Object FriendlyName, Status, InstanceId | Format-Table

# Check for CNCA/CNCB ports (default com0com naming)
Write-Host "`n=== Checking for CNCA/CNCB ports ===" -ForegroundColor Cyan
Get-PnpDevice -Class Ports | Where-Object {$_.FriendlyName -like "*CNC*"} | Select-Object FriendlyName, Status | Format-Table

# List setupc command if exists
Write-Host "`n=== Looking for setupc.exe ===" -ForegroundColor Cyan
foreach ($path in $com0comPaths) {
    $setupc = Join-Path $path "setupc.exe"
    if (Test-Path $setupc) {
        Write-Host "Found setupc.exe at: $setupc" -ForegroundColor Green
    }
}

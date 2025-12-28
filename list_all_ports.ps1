# Method 1: .NET SerialPort list
Write-Host "=== Method 1: .NET SerialPort ===" -ForegroundColor Cyan
[System.IO.Ports.SerialPort]::getportnames() | Sort-Object

# Method 2: Registry check
Write-Host "`n=== Method 2: Registry ===" -ForegroundColor Cyan
Get-ItemProperty -Path "HKLM:\HARDWARE\DEVICEMAP\SERIALCOMM\" -ErrorAction SilentlyContinue | Format-List

# Method 3: Device Manager
Write-Host "`n=== Method 3: Device Manager ===" -ForegroundColor Cyan
Get-PnpDevice -Class Ports -ErrorAction SilentlyContinue | Where-Object {$_.Status -eq "OK"} | Select-Object FriendlyName, InstanceId | Format-Table

# Method 4: WMI Query
Write-Host "`n=== Method 4: WMI Query ===" -ForegroundColor Cyan
Get-WmiObject Win32_SerialPort -ErrorAction SilentlyContinue | Select-Object DeviceID, Description, Status | Format-Table

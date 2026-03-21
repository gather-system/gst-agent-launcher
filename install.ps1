$ErrorActionPreference = "Stop"

$installDir = "D:\WorkSpace\GatherTech\PM\scripts"
$configDir = Join-Path $env:USERPROFILE ".config\gst-launcher"
$configFile = Join-Path $configDir "agents.json"

# Step 1: Build
Write-Host "=== Step 1: Build ===" -ForegroundColor Cyan
& "$PSScriptRoot\build.ps1"

# Step 2: Install to PATH directory
Write-Host "`n=== Step 2: Install ===" -ForegroundColor Cyan
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}
Copy-Item -Path "gst-launcher.exe" -Destination $installDir -Force
Write-Host "Installed to: $installDir\gst-launcher.exe" -ForegroundColor Green

# Step 3: Create user config if not exists
Write-Host "`n=== Step 3: Config ===" -ForegroundColor Cyan
if (Test-Path $configFile) {
    Write-Host "Config already exists: $configFile" -ForegroundColor Yellow
} else {
    if (-not (Test-Path $configDir)) {
        New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    }
    Copy-Item -Path "$PSScriptRoot\config\default.json" -Destination $configFile -Force
    Write-Host "Created config: $configFile" -ForegroundColor Green
}

# Done
Write-Host "`n=== Installation Complete ===" -ForegroundColor Cyan
Write-Host "Run 'gst-launcher' from any directory to start." -ForegroundColor Green
Write-Host "Run 'gst-launcher --version' to verify." -ForegroundColor Green

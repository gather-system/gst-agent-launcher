$ErrorActionPreference = "Stop"

$version = "0.2.0"
$commit = git rev-parse --short HEAD
$date = Get-Date -Format "yyyy-MM-dd"

Write-Host "Building gst-launcher v$version ($commit, $date)..." -ForegroundColor Cyan

go build -ldflags "-X main.version=$version -X main.commit=$commit -X main.date=$date" -o gst-launcher.exe .

if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item gst-launcher.exe).Length / 1MB, 1)
    Write-Host "Build successful: gst-launcher.exe (${size} MB)" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

# Build script for Mobile version (Android/iOS)
# Run: ./build-mobile.ps1

Write-Host "Building Match-3 for Mobile..." -ForegroundColor Cyan

# Check if gomobile is installed
$gomobilePath = (go env GOPATH) + "\bin\gomobile.exe"
if (-not (Test-Path $gomobilePath)) {
    Write-Host "Installing gomobile..." -ForegroundColor Yellow
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to install gomobile!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Initializing gomobile..." -ForegroundColor Yellow
    & "$gomobilePath" init
}

# Build for Android
Write-Host "`nBuilding Android APK..." -ForegroundColor Cyan
& "$gomobilePath" build -o match3.apk -target=android ./cmd

if ($LASTEXITCODE -eq 0) {
    Write-Host "Android build successful!" -ForegroundColor Green
    Write-Host "APK location: match3.apk" -ForegroundColor Cyan
} else {
    Write-Host "Android build failed!" -ForegroundColor Red
}

# Build for iOS (requires macOS)
if ($IsMacOS) {
    Write-Host "`nBuilding iOS package..." -ForegroundColor Cyan
    & "$gomobilePath" build -o match3.ipa -target=ios ./cmd
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "iOS build successful!" -ForegroundColor Green
        Write-Host "IPA location: match3.ipa" -ForegroundColor Cyan
    } else {
        Write-Host "iOS build failed!" -ForegroundColor Red
    }
} else {
    Write-Host "`niOS build skipped (requires macOS)" -ForegroundColor Yellow
}

Write-Host "`nTo install APK on device:" -ForegroundColor Cyan
Write-Host "  adb install match3.apk" -ForegroundColor White

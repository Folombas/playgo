# Build script to copy sprites to assets folder
# Run: .\copy-sprites.ps1

$spritesRoot = "D:\Projects\sprites"
$assetsDir = "assets"

# Create assets directory if not exists
if (-not (Test-Path $assetsDir)) {
    New-Item -ItemType Directory -Path $assetsDir | Out-Null
}

# Create subdirectories
@('gems', 'backgrounds', 'ui') | ForEach-Object {
    $path = Join-Path $assetsDir $_
    if (-not (Test-Path $path)) {
        New-Item -ItemType Directory -Path $path | Out-Null
    }
}

Write-Host "Copying sprites to assets..." -ForegroundColor Cyan

# Gem sprites (32x32 food sprites)
$gems = @{
    "red"    = "03_Food\Food_Packs\32x32\apple.png"
    "green"  = "03_Food\Food_Packs\32x32\pear.png"
    "yellow" = "03_Food\Food_Packs\32x32\lemon.png"
    "blue"   = "03_Food\Food_Packs\32x32\blueberries.png"
    "purple" = "03_Food\Food_Packs\32x32\grapes.png"
    "orange" = "03_Food\Food_Packs\32x32\orange.png"
}

foreach ($entry in $gems.GetEnumerator()) {
    $src = Join-Path $spritesRoot $entry.Value
    $dst = Join-Path $assetsDir "gems\$($entry.Key).png"
    if (Test-Path $src) {
        Copy-Item $src $dst
        Write-Host "  Copied gem: $($entry.Key)" -ForegroundColor Green
    }
}

# Background
$bgSrc = Join-Path $spritesRoot "05_Backgrounds\Horizontal-2D-BG-PNG\game_background_1\game_background_1.png"
$bgDst = Join-Path $assetsDir "backgrounds\game_bg.png"
if (Test-Path $bgSrc) {
    Copy-Item $bgSrc $bgDst
    Write-Host "  Copied background" -ForegroundColor Green
}

# UI elements
$uiSrc = Join-Path $spritesRoot "07_Effects\rad-rainbow-lifebar.png"
$uiDst = Join-Path $assetsDir "ui\progress_bar.png"
if (Test-Path $uiSrc) {
    Copy-Item $uiSrc $uiDst
    Write-Host "  Copied UI element" -ForegroundColor Green
}

Write-Host "`nSprite copy complete!" -ForegroundColor Cyan
Get-ChildItem -Recurse $assetsDir | Where-Object { -not $_.PSIsContainer } | Format-Table Name, Length

# Build script for WebAssembly version
# Run: ./build-web.ps1

Write-Host "Building Match-3 for WebAssembly..." -ForegroundColor Cyan

# Copy wasm_exec.js from Go installation
$goRoot = $env:GOROOT
if (-not $goRoot) {
    $goRoot = (go env GOROOT)
}

$wasmExecSrc = Join-Path $goRoot "misc/wasm/wasm_exec.js"
$wasmExecDst = "web/wasm_exec.js"

if (Test-Path $wasmExecSrc) {
    Copy-Item $wasmExecSrc $wasmExecDst
    Write-Host "Copied wasm_exec.js" -ForegroundColor Green
} else {
    Write-Host "Warning: wasm_exec.js not found at $wasmExecSrc" -ForegroundColor Yellow
}

# Build WASM
$env:GOOS = "js"
$env:GOARCH = "wasm"

Write-Host "Compiling to WASM..." -ForegroundColor Cyan
go build -o web/match3.wasm ./cmd

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Files in web/:" -ForegroundColor Cyan
    Get-ChildItem web | Format-Table Name, Length
    Write-Host "`nTo run locally:" -ForegroundColor Cyan
    Write-Host "  cd web" -ForegroundColor White
    Write-Host "  python -m http.server 8080" -ForegroundColor White
    Write-Host "  Then open http://localhost:8080" -ForegroundColor White
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

#!/bin/bash
# Build for WebAssembly

echo "🌐 Building for WebAssembly..."

# Build WASM
GOOS=js GOARCH=wasm go build -o web/crystal-cascade.wasm ./cmd

if [ $? -eq 0 ]; then
    echo "✅ WASM build successful!"
    echo "📁 Output: web/crystal-cascade.wasm"
    echo ""
    echo "To run locally:"
    echo "  1. Copy wasm_exec.js: cp \$(go env GOROOT)/misc/wasm/wasm_exec.js web/"
    echo "  2. Start server: python3 -m http.server 8080"
    echo "  3. Open: http://localhost:8080/web/"
else
    echo "❌ Build failed!"
    exit 1
fi

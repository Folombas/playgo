#!/bin/bash
# Build script for Crystal Cascade - Linux/macOS

echo "🔨 Building Crystal Cascade for current platform..."

# Get platform
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
OUTPUT="crystal-cascade"

if [[ "$PLATFORM" == "darwin" ]]; then
    echo "🍎 Building for macOS..."
elif [[ "$PLATFORM" == "linux" ]]; then
    echo "🐧 Building for Linux..."
fi

go build -o "$OUTPUT" ./cmd

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Run with: ./$OUTPUT"
else
    echo "❌ Build failed!"
    exit 1
fi

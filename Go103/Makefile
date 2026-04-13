.PHONY: build run test clean wasm android

# Default build for current platform
build:
	go build -o match3

# Run the game
run:
	go run .

# Run unit tests
test:
	go test -v -count=1 ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build for Web (WASM)
wasm:
	GOOS=js GOARCH=wasm go build -o match3.wasm
	@echo "WASM build complete. Serve with:"
	@echo "  python3 -m http.server 8080"

# Build for Android (requires gomobile)
android:
	@gomobile version || (echo "gomobile not installed"; exit 1)
	gomobile build -target=android -o match3.apk

# Clean build artifacts
clean:
	rm -f match3 match3.exe match3.wasm match3.apk coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Lint
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

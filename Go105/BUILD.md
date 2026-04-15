# 🛠️ Build Instructions — Geometric Match Game

## Требования
- Go 1.21+
- Ebitengine v2.8+

## Установка зависимостей
```bash
go mod tidy
```

## Сборка

### Windows
```bash
go build -o matchgame.exe
./matchgame.exe
```

### Linux / macOS
```bash
go build -o matchgame
./matchgame
```

### Web (WASM)
```bash
# Копирование wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# Сборка WASM
GOOS=js GOARCH=wasm go build -o matchgame.wasm

# Открыть index.html в браузере (нужен локальный сервер)
python -m http.server 8080
# Затем открыть http://localhost:8080
```

### Android
```bash
# Установка gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Сборка APK
gomobile build -target=android -o matchgame.apk
```

## Запуск в режиме разработки
```bash
go run .
```

## Управление
| Клавиша | Действие |
|---------|----------|
| ← → | Движение влево/вправо |
| ↑ | Поворот |
| ↓ | Ускоренное падение |
| Пробел | Мгновенная установка |
| P / Esc | Пауза |
| Enter / Space | Старт / Рестарт |

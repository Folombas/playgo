package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WindowWidth  = 800
	WindowHeight = 800
)

func main() {
	// Загружаем спрайты
	LoadSprites()

	// Создаём игру
	game := NewGame(WindowWidth, WindowHeight)

	// Настраиваем окно
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Match-3 - Go365 Challenge")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Запускаем игру
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

/*
Инструкция по сборке:

1. Windows:
   GOOS=windows GOARCH=amd64 go build -o match3.exe

2. Android (gomobile):
   gomobile bind -target=android -o match3.aar
   или используйте ebitenmobile для создания APK

3. Web (WASM):
   GOOS=js GOARCH=wasm go build -o game.wasm
   Скопируйте wasm_exec.js из GOROOT (misc/wasm/wasm_exec.js)
   Создайте index.html с загрузкой WASM

Примечания:
- Для добавления новых типов фишек увеличьте константу NumColors в board.go
- Для изменения размера поля измените константу BoardSize
- Все спрайты встраиваются через go:embed
- Для кроссплатформенной сборки используйте соответствующие GOOS/GOARCH
*/

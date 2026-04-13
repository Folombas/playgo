// Match-3 Game - Казуальная карманная игра "Три в ряд"
// Построена на движке Ebitengine (v2.6+)
// Кроссплатформенная: Windows, Android, Web (WASM)
//
// Сборка:
//   Windows: GOOS=windows GOARCH=amd64 go build -o match3.exe
//   Web:     GOOS=js GOARCH=wasm go build -o game.wasm
//   Android: gomobile build -target=android -o match3.apk
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800
	screenHeight = 800
)

func main() {
	// Загружаем спрайты (с fallback на программную графику)
	loadSprites()

	// Инициализируем звуки
	initSounds()

	// Создаём игру
	game := NewGame()

	// Настраиваем окно
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Match-3 - Три в ряд")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Запускаем игровой цикл
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("Ошибка запуска игры:", err)
	}
}

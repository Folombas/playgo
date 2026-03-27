// Go365 - День 87 (27 марта 2026)
// GO MARIO v1.0.0 - Классический платформер на Go + Ebitengine
// Полная переработка с нуля

package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"go_mario/internal/assets"
	"go_mario/internal/game"
)

const (
	// Версия игры
	Version = "1.0.0"

	// Размеры окна игры
	WindowWidth  = 1024
	WindowHeight = 768
)

func main() {
	// Парсинг аргументов командной строки
	version := flag.Bool("version", false, "показать версию")
	flag.Parse()

	if *version {
		println("Go Mario v" + Version)
		println("Go365 Challenge - Day 87")
		return
	}

	// Загрузка ассетов
	log.Println("Загрузка ассетов...")
	gameAssets := assets.Get()
	if err := gameAssets.Load(); err != nil {
		log.Fatalf("Ошибка загрузки ассетов: %v", err)
	}
	log.Println("Ассеты загружены успешно!")

	// Создание игры
	log.Println("Инициализация игры...")
	g := game.NewGame()

	// Настройка окна
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Go Mario v" + Version + " | Go365 Day 87")

	// Запуск игры
	log.Println("Запуск игры...")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

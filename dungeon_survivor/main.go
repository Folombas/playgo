// Package main - точка входа игры Village Platformer
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/playgo/go90/internal/entity"
	"github.com/playgo/go90/internal/game"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

func init() {
	// Загрузка спрайтов из Platformer Complete!
	
	// Спрайт игрока - используем p1_spritesheet.png
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/player/p1_spritesheet.png")
	if err == nil {
		entity.PlayerSprite = playerImg
	}

	// Спрайт ландшафта - используем tiles_spritesheet.png
	terrainImg, _, err := ebitenutil.NewImageFromFile("assets/tiles/tiles_spritesheet.png")
	if err == nil {
		entity.TerrainSprite = terrainImg
	}

	// Спрайты объектов
	objectsImg, _, err := ebitenutil.NewImageFromFile("assets/items/items_spritesheet.png")
	if err == nil {
		entity.ObjectsSprite = objectsImg
	}

	// Полоска здоровья
	lifebarImg, _, err := ebitenutil.NewImageFromFile("assets/rad-rainbow-lifebar.png")
	if err == nil {
		entity.LifebarSprite = lifebarImg
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("🏠 Village Platformer - 2D Платформер")

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

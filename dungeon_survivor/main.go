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
	// Загрузка спрайтов
	terrainImg, _, err := ebitenutil.NewImageFromFile("assets/Jungle_terrain.png")
	if err == nil {
		entity.TerrainSprite = terrainImg
	}

	objectsImg, _, err := ebitenutil.NewImageFromFile("assets/objects.PNG")
	if err == nil {
		entity.ObjectsSprite = objectsImg
	}

	// Спрайт игрока
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/8-bits-characters-gaming-assets.jpg")
	if err == nil {
		entity.PlayerSprite = playerImg
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

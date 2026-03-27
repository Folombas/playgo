// Package assets предоставляет систему загрузки и управления игровыми ассетами
package assets

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
)

// Assets хранит все загруженные ресурсы игры
type Assets struct {
	// Игрок (Марио)
	PlayerStand  *ebiten.Image
	PlayerWalk1  *ebiten.Image
	PlayerWalk2  *ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerDuck   *ebiten.Image
	PlayerHurt   *ebiten.Image

	// Враги
	SlimeWalk1 *ebiten.Image
	SlimeWalk2 *ebiten.Image
	SlimeDead  *ebiten.Image
	FlyWalk1   *ebiten.Image
	FlyWalk2   *ebiten.Image
	FlyDead    *ebiten.Image

	// Тайлы
	GrassTop    *ebiten.Image
	GrassMid    *ebiten.Image
	GrassLeft   *ebiten.Image
	GrassRight  *ebiten.Image
	DirtTop     *ebiten.Image
	DirtMid     *ebiten.Image
	Brick       *ebiten.Image
	BoxEmpty    *ebiten.Image
	BoxItem     *ebiten.Image
	BoxCoin     *ebiten.Image
	BoxUsed     *ebiten.Image

	// Предметы
	CoinGold   *ebiten.Image
	CoinSilver *ebiten.Image
	CoinBronze *ebiten.Image
	Mushroom   *ebiten.Image
	Star       *ebiten.Image
	FlagGreen  *ebiten.Image
	FlagRed    *ebiten.Image

	// Декорации
	Cloud1 *ebiten.Image
	Cloud2 *ebiten.Image
	Cloud3 *ebiten.Image
	Bush   *ebiten.Image
	Plant  *ebiten.Image

	// Фон
	Background *ebiten.Image

	// Шрифт
	GameFont font.Face
}

var instance *Assets

// Get возвращает глобальный экземпляр ассетов
func Get() *Assets {
	if instance == nil {
		instance = &Assets{}
	}
	return instance
}

// Load загружает все ассеты игры
func (a *Assets) Load() error {
	basePath := "assets/sprites"

	// Загрузка игрока
	var err error
	a.PlayerStand, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "mario/p1_stand.png"))
	if err != nil {
		return fmt.Errorf("failed to load player stand: %w", err)
	}

	a.PlayerWalk1, _, err = loadWalkFrame(filepath.Join(basePath, "mario/walk/p1_walk.png"), 0)
	if err != nil {
		return fmt.Errorf("failed to load player walk1: %w", err)
	}

	a.PlayerWalk2, _, err = loadWalkFrame(filepath.Join(basePath, "mario/walk/p1_walk.png"), 1)
	if err != nil {
		return fmt.Errorf("failed to load player walk2: %w", err)
	}

	a.PlayerJump, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "mario/p1_jump.png"))
	if err != nil {
		return fmt.Errorf("failed to load player jump: %w", err)
	}

	a.PlayerDuck, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "mario/p1_duck.png"))
	if err != nil {
		return fmt.Errorf("failed to load player duck: %w", err)
	}

	a.PlayerHurt, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "mario/p1_hurt.png"))
	if err != nil {
		return fmt.Errorf("failed to load player hurt: %w", err)
	}

	// Загрузка врагов
	a.SlimeWalk1, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/slimeWalk1.png"))
	if err != nil {
		return fmt.Errorf("failed to load slime walk1: %w", err)
	}

	a.SlimeWalk2, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/slimeWalk2.png"))
	if err != nil {
		return fmt.Errorf("failed to load slime walk2: %w", err)
	}

	a.SlimeDead, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/slimeDead.png"))
	if err != nil {
		return fmt.Errorf("failed to load slime dead: %w", err)
	}

	a.FlyWalk1, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/flyFly1.png"))
	if err != nil {
		return fmt.Errorf("failed to load fly walk1: %w", err)
	}

	a.FlyWalk2, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/flyFly2.png"))
	if err != nil {
		return fmt.Errorf("failed to load fly walk2: %w", err)
	}

	a.FlyDead, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "enemies/flyDead.png"))
	if err != nil {
		return fmt.Errorf("failed to load fly dead: %w", err)
	}

	// Загрузка тайлов
	a.GrassTop, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/grassMid.png"))
	if err != nil {
		return fmt.Errorf("failed to load grass top: %w", err)
	}

	a.GrassMid, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/grass.png"))
	if err != nil {
		return fmt.Errorf("failed to load grass mid: %w", err)
	}

	a.GrassLeft, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/grassLeft.png"))
	if err != nil {
		return fmt.Errorf("failed to load grass left: %w", err)
	}

	a.GrassRight, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/grassRight.png"))
	if err != nil {
		return fmt.Errorf("failed to load grass right: %w", err)
	}

	a.DirtTop, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/dirtMid.png"))
	if err != nil {
		return fmt.Errorf("failed to load dirt top: %w", err)
	}

	a.DirtMid, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/dirt.png"))
	if err != nil {
		return fmt.Errorf("failed to load dirt mid: %w", err)
	}

	a.Brick, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/brickWall.png"))
	if err != nil {
		return fmt.Errorf("failed to load brick: %w", err)
	}

	a.BoxEmpty, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/boxEmpty.png"))
	if err != nil {
		return fmt.Errorf("failed to load box empty: %w", err)
	}

	a.BoxItem, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/boxItem.png"))
	if err != nil {
		return fmt.Errorf("failed to load box item: %w", err)
	}

	a.BoxCoin, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/boxCoin.png"))
	if err != nil {
		return fmt.Errorf("failed to load box coin: %w", err)
	}

	a.BoxUsed, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "tiles/boxItem_disabled.png"))
	if err != nil {
		return fmt.Errorf("failed to load box used: %w", err)
	}

	// Загрузка предметов
	a.CoinGold, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/coinGold.png"))
	if err != nil {
		return fmt.Errorf("failed to load coin gold: %w", err)
	}

	a.CoinSilver, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/coinSilver.png"))
	if err != nil {
		return fmt.Errorf("failed to load coin silver: %w", err)
	}

	a.CoinBronze, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/coinBronze.png"))
	if err != nil {
		return fmt.Errorf("failed to load coin bronze: %w", err)
	}

	a.Mushroom, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/mushroomRed.png"))
	if err != nil {
		return fmt.Errorf("failed to load mushroom: %w", err)
	}

	a.Star, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/star.png"))
	if err != nil {
		return fmt.Errorf("failed to load star: %w", err)
	}

	a.FlagGreen, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/flagGreen.png"))
	if err != nil {
		return fmt.Errorf("failed to load flag green: %w", err)
	}

	a.FlagRed, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "items/flagRed.png"))
	if err != nil {
		return fmt.Errorf("failed to load flag red: %w", err)
	}

	// Загрузка декораций
	a.Cloud1, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "decorations/cloud1.png"))
	if err != nil {
		return fmt.Errorf("failed to load cloud1: %w", err)
	}

	a.Cloud2, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "decorations/cloud2.png"))
	if err != nil {
		return fmt.Errorf("failed to load cloud2: %w", err)
	}

	a.Cloud3, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "decorations/cloud3.png"))
	if err != nil {
		return fmt.Errorf("failed to load cloud3: %w", err)
	}

	a.Bush, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "decorations/bush.png"))
	if err != nil {
		return fmt.Errorf("failed to load bush: %w", err)
	}

	a.Plant, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "decorations/plant.png"))
	if err != nil {
		return fmt.Errorf("failed to load plant: %w", err)
	}

	// Загрузка фона
	a.Background, _, err = ebitenutil.NewImageFromFile(filepath.Join(basePath, "backgrounds/bg.png"))
	if err != nil {
		return fmt.Errorf("failed to load background: %w", err)
	}

	// Загрузка шрифта
	// Используем встроенный шрифт Ebitengine
	a.GameFont = nil // Будет использован шрифт по умолчанию

	return nil
}

// loadWalkFrame загружает кадр анимации ходьбы из спрайт-листа
func loadWalkFrame(path string, frameIndex int) (*ebiten.Image, image.Point, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, image.Point{}, err
	}

	// Спрайт-лист p1_walk.png содержит 2 кадра
	// Каждый кадр примерно 32x32 пикселя
	bounds := img.Bounds()
	frameWidth := bounds.Dx() / 2

	// Вырезаем нужный кадр
	rect := image.Rect(frameIndex*frameWidth, 0, (frameIndex+1)*frameWidth, bounds.Dy())
	frame := img.SubImage(rect).(*ebiten.Image)

	return frame, image.Point{frameWidth, bounds.Dy()}, nil
}

// loadFont загружает шрифт
// Пока возвращает nil, так как шрифты в ZIP
// В будущем можно реализовать распаковку
func loadFont(path string, size int) (font.Face, error) {
	return nil, fmt.Errorf("font loading not implemented")
}

// Reset сбрасывает глобальный экземпляр (для тестов)
func Reset() {
	instance = nil
}

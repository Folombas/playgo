// Package sprite - загрузка и управление спрайтами
// Go365 Day 91 - City Platformer
package sprite

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// SpriteType - типы спрайтов
type SpriteType string

const (
	SpritePlayer      SpriteType = "player"
	SpriteEnemyMutant SpriteType = "mutant"
	SpriteEnemyRobot  SpriteType = "robot"
	SpriteEnemyZombie SpriteType = "zombie"
	SpriteItemMedkit  SpriteType = "medkit"
	SpriteItemAmmo    SpriteType = "ammo"
	SpriteItemFood    SpriteType = "food"
	SpriteItemParts   SpriteType = "parts"
	SpriteTileGround  SpriteType = "ground"
	SpriteTileBuilding SpriteType = "building"
	SpriteTileRubble  SpriteType = "rubble"
	SpriteBullet      SpriteType = "bullet"
)

// SpriteSheet - атлас спрайтов
type SpriteSheet struct {
	Player      map[string][]*ebiten.Image
	Enemies     map[string][]*ebiten.Image
	Items       map[string][]*ebiten.Image
	Tiles       map[string]*ebiten.Image
	Bullets     []*ebiten.Image
	Backgrounds []*ebiten.Image
}

// LoadSpriteSheet - загрузка спрайт-листа
func LoadSpriteSheet() *SpriteSheet {
	ss := &SpriteSheet{
		Player:  make(map[string][]*ebiten.Image),
		Enemies: make(map[string][]*ebiten.Image),
		Items:   make(map[string][]*ebiten.Image),
		Tiles:   make(map[string]*ebiten.Image),
	}

	// Инициализация спрайтов
	ss.initPlayer()
	ss.initEnemies()
	ss.initItems()
	ss.initTiles()
	ss.initBullets()

	return ss
}

// initPlayer - инициализация спрайтов игрока
func (ss *SpriteSheet) initPlayer() {
	// Создаём спрайты программно (заглушки)
	// Формат: [состояние][кадр]
	
	// Стойка
	ss.Player["stand"] = []*ebiten.Image{
		ss.createPlayerImage("stand", 0),
	}
	
	// Бег (4 кадра)
	ss.Player["run"] = make([]*ebiten.Image, 4)
	for i := 0; i < 4; i++ {
		ss.Player["run"][i] = ss.createPlayerImage("run", i)
	}
	
	// Прыжок
	ss.Player["jump"] = []*ebiten.Image{
		ss.createPlayerImage("jump", 0),
	}
	
	// Присед
	ss.Player["crouch"] = []*ebiten.Image{
		ss.createPlayerImage("crouch", 0),
	}
	
	// Стрельба
	ss.Player["shoot"] = []*ebiten.Image{
		ss.createPlayerImage("shoot", 0),
	}
}

// createPlayerImage - создание спрайта игрока
func (ss *SpriteSheet) createPlayerImage(state string, frame int) *ebiten.Image {
	width, height := 32, 48
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Цвета для разных частей тела
	skinColor := color.RGBA{255, 220, 180, 255}
	shirtColor := color.RGBA{60, 80, 100, 255}
	pantsColor := color.RGBA{40, 40, 40, 255}
	bootColor := color.RGBA{30, 30, 30, 255}
	
	// Анимация бега
	offsetY := 0
	if state == "run" {
		offsetY = (frame % 2) * 2
	}
	
	// Ноги
	for y := 32; y < height; y++ {
		for x := 8; x < 24; x++ {
			img.Set(x, y+offsetY, pantsColor)
		}
	}
	
	// Ботинки
	for y := 44; y < height; y++ {
		for x := 8; x < 16; x++ {
			img.Set(x, y+offsetY, bootColor)
		}
		for x := 16; x < 24; x++ {
			img.Set(x, y+offsetY, bootColor)
		}
	}
	
	// Тело
	for y := 16 + offsetY; y < 32+offsetY; y++ {
		for x := 6; x < 26; x++ {
			img.Set(x, y, shirtColor)
		}
	}
	
	// Голова
	for y := offsetY; y < 16+offsetY; y++ {
		for x := 8; x < 24; x++ {
			img.Set(x, y, skinColor)
		}
	}
	
	// Руки с оружием
	if state == "shoot" {
		for y := 18 + offsetY; y < 24+offsetY; y++ {
			for x := 20; x < 32; x++ {
				img.Set(x, y, shirtColor)
			}
		}
	} else {
		for y := 18 + offsetY; y < 28+offsetY; y++ {
			for x := 6; x < 10; x++ {
				img.Set(x, y, skinColor)
			}
			for x := 22; x < 26; x++ {
				img.Set(x, y, skinColor)
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// initEnemies - инициализация спрайтов врагов
func (ss *SpriteSheet) initEnemies() {
	// Мутант
	ss.Enemies["mutant"] = make([]*ebiten.Image, 2)
	for i := 0; i < 2; i++ {
		ss.Enemies["mutant"][i] = ss.createMutantImage(i)
	}
	
	// Робот
	ss.Enemies["robot"] = make([]*ebiten.Image, 2)
	for i := 0; i < 2; i++ {
		ss.Enemies["robot"][i] = ss.createRobotImage(i)
	}
	
	// Зомби
	ss.Enemies["zombie"] = make([]*ebiten.Image, 2)
	for i := 0; i < 2; i++ {
		ss.Enemies["zombie"][i] = ss.createZombieImage(i)
	}
}

// createMutantImage - создание спрайта мутанта
func (ss *SpriteSheet) createMutantImage(frame int) *ebiten.Image {
	width, height := 40, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	bodyColor := color.RGBA{80, 120, 60, 255}
	eyeColor := color.RGBA{255, 50, 50, 255}
	
	// Тело
	for y := 10; y < height; y++ {
		for x := 5; x < width-5; x++ {
			img.Set(x, y, bodyColor)
		}
	}
	
	// Глаза
	img.Set(12, 18, eyeColor)
	img.Set(28, 18, eyeColor)
	
	return ebiten.NewImageFromImage(img)
}

// createRobotImage - создание спрайта робота
func (ss *SpriteSheet) createRobotImage(frame int) *ebiten.Image {
	width, height := 36, 44
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	metalColor := color.RGBA{100, 100, 120, 255}
	lightColor := color.RGBA{200, 200, 220, 255}
	eyeColor := color.RGBA{255, 0, 0, 255}
	
	// Тело
	for y := 15; y < height; y++ {
		for x := 8; x < width-8; x++ {
			img.Set(x, y, metalColor)
		}
	}
	
	// Голова
	for y := 5; y < 15; y++ {
		for x := 10; x < width-10; x++ {
			img.Set(x, y, metalColor)
		}
	}
	
	// Глаз
	img.Set(18, 10, eyeColor)
	
	// Ноги
	for y := 35; y < height; y++ {
		for x := 10; x < 16; x++ {
			img.Set(x, y, lightColor)
		}
		for x := 20; x < 26; x++ {
			img.Set(x, y, lightColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createZombieImage - создание спрайта зомби
func (ss *SpriteSheet) createZombieImage(frame int) *ebiten.Image {
	width, height := 32, 44
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	skinColor := color.RGBA{100, 120, 80, 255}
	clothColor := color.RGBA{80, 60, 60, 255}
	
	// Тело
	for y := 18; y < height; y++ {
		for x := 6; x < width-6; x++ {
			img.Set(x, y, clothColor)
		}
	}
	
	// Голова и руки
	for y := 8; y < 18; y++ {
		for x := 8; x < width-8; x++ {
			img.Set(x, y, skinColor)
		}
	}
	
	// Руки вперёд
	for y := 14; y < 24; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, skinColor)
		}
		for x := width-10; x < width; x++ {
			img.Set(x, y, skinColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// initItems - инициализация спрайтов предметов
func (ss *SpriteSheet) initItems() {
	ss.Items["medkit"] = []*ebiten.Image{ss.createMedkitImage()}
	ss.Items["ammo"] = []*ebiten.Image{ss.createAmmoImage()}
	ss.Items["food"] = []*ebiten.Image{ss.createFoodImage()}
	ss.Items["parts"] = []*ebiten.Image{ss.createPartsImage()}
}

// createMedkitImage - создание спрайта аптечки
func (ss *SpriteSheet) createMedkitImage() *ebiten.Image {
	width, height := 24, 20
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	boxColor := color.RGBA{200, 50, 50, 255}
	crossColor := color.RGBA{255, 255, 255, 255}
	
	// Коробка
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, boxColor)
		}
	}
	
	// Крест
	for y := 4; y < 16; y++ {
		for x := 10; x < 14; x++ {
			img.Set(x, y, crossColor)
		}
	}
	for y := 8; y < 12; y++ {
		for x := 4; x < 20; x++ {
			img.Set(x, y, crossColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createAmmoImage - создание спрайта патронов
func (ss *SpriteSheet) createAmmoImage() *ebiten.Image {
	width, height := 20, 16
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	caseColor := color.RGBA{200, 180, 100, 255}
	bulletColor := color.RGBA{100, 100, 100, 255}
	
	// Гильзы
	for i := 0; i < 3; i++ {
		for y := 4; y < height; y++ {
			for x := i*6 + 2; x < i*6+6; x++ {
				img.Set(x, y, caseColor)
			}
		}
		for y := 0; y < 4; y++ {
			for x := i*6 + 2; x < i*6+6; x++ {
				img.Set(x, y, bulletColor)
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createFoodImage - создание спрайта еды
func (ss *SpriteSheet) createFoodImage() *ebiten.Image {
	width, height := 20, 20
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	foodColor := color.RGBA{180, 100, 60, 255}
	
	// Консервная банка
	for y := 4; y < height; y++ {
		for x := 4; x < width-4; x++ {
			img.Set(x, y, foodColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createPartsImage - создание спрайта деталей
func (ss *SpriteSheet) createPartsImage() *ebiten.Image {
	width, height := 24, 24
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	metalColor := color.RGBA{120, 120, 140, 255}
	
	// Шестерёнка
	centerX, centerY := 12, 12
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := dx*dx + dy*dy
			if dist < 36 && dist > 16 {
				img.Set(x, y, metalColor)
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// initTiles - инициализация тайлов
func (ss *SpriteSheet) initTiles() {
	ss.Tiles["ground"] = ss.createGroundTile()
	ss.Tiles["building"] = ss.createBuildingTile()
	ss.Tiles["rubble"] = ss.createRubbleTile()
	ss.Tiles["metal"] = ss.createMetalTile()
}

// createGroundTile - создание тайла земли
func (ss *SpriteSheet) createGroundTile() *ebiten.Image {
	width, height := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	groundColor := color.RGBA{80, 70, 60, 255}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, groundColor)
		}
	}
	
	// Добавим немного текстуры
	img.Set(10, 10, color.RGBA{70, 60, 50, 255})
	img.Set(30, 20, color.RGBA{90, 80, 70, 255})
	img.Set(50, 40, color.RGBA{70, 60, 50, 255})
	
	return ebiten.NewImageFromImage(img)
}

// createBuildingTile - создание тайла здания
func (ss *SpriteSheet) createBuildingTile() *ebiten.Image {
	width, height := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	wallColor := color.RGBA{100, 100, 110, 255}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, wallColor)
		}
	}
	
	// Окно
	for y := 10; y < 30; y++ {
		for x := 20; x < 44; x++ {
			img.Set(x, y, color.RGBA{50, 50, 60, 255})
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createRubbleTile - создание тайла обломков
func (ss *SpriteSheet) createRubbleTile() *ebiten.Image {
	width, height := 64, 32
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	rubbleColor := color.RGBA{90, 80, 70, 255}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, rubbleColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// createMetalTile - создание тайла металла
func (ss *SpriteSheet) createMetalTile() *ebiten.Image {
	width, height := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	metalColor := color.RGBA{120, 120, 130, 255}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, metalColor)
		}
	}
	
	// Заклёпки
	img.Set(5, 5, color.RGBA{100, 100, 110, 255})
	img.Set(59, 5, color.RGBA{100, 100, 110, 255})
	img.Set(5, 59, color.RGBA{100, 100, 110, 255})
	img.Set(59, 59, color.RGBA{100, 100, 110, 255})
	
	return ebiten.NewImageFromImage(img)
}

// initBullets - инициализация пуль
func (ss *SpriteSheet) initBullets() {
	ss.Bullets = []*ebiten.Image{ss.createBulletImage()}
}

// createBulletImage - создание спрайта пули
func (ss *SpriteSheet) createBulletImage() *ebiten.Image {
	width, height := 12, 6
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	bulletColor := color.RGBA{255, 200, 50, 255}
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bulletColor)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// GetPlayerSprite - получение спрайта игрока
func (ss *SpriteSheet) GetPlayerSprite(state string, frame int) *ebiten.Image {
	if sprites, ok := ss.Player[state]; ok && len(sprites) > 0 {
		if frame >= 0 && frame < len(sprites) {
			return sprites[frame]
		}
		return sprites[0]
	}
	return ss.Player["stand"][0]
}

// GetEnemySprite - получение спрайта врага
func (ss *SpriteSheet) GetEnemySprite(enemyType string, frame int) *ebiten.Image {
	if sprites, ok := ss.Enemies[enemyType]; ok && len(sprites) > 0 {
		return sprites[frame%len(sprites)]
	}
	return ss.Enemies["mutant"][0]
}

// GetItemSprite - получение спрайта предмета
func (ss *SpriteSheet) GetItemSprite(itemType string) *ebiten.Image {
	if sprites, ok := ss.Items[itemType]; ok && len(sprites) > 0 {
		return sprites[0]
	}
	return ss.Items["parts"][0]
}

// GetTile - получение тайла
func (ss *SpriteSheet) GetTile(tileType string) *ebiten.Image {
	if tile, ok := ss.Tiles[tileType]; ok {
		return tile
	}
	return ss.Tiles["ground"]
}

// GetBullet - получение пули
func (ss *SpriteSheet) GetBullet() *ebiten.Image {
	if len(ss.Bullets) > 0 {
		return ss.Bullets[0]
	}
	return ss.createBulletImage()
}

// LoadImage - загрузка изображения из файла
func LoadImage(path string) (*ebiten.Image, error) {
	fullPath := filepath.Join("assets", "sprites", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: Sprite file not found: %s\n", fullPath)
		return nil, err
	}

	img, _, err := ebitenutil.NewImageFromFile(fullPath)
	return img, err
}

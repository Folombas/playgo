// Package sprite - загрузка и управление спрайтами
// Go365 Day 92 - Cyber City Runner
package sprite

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet хранит все загруженные спрайты
type SpriteSheet struct {
	// Игрок
	PlayerStand *ebiten.Image
	PlayerWalk  []*ebiten.Image
	PlayerJump  *ebiten.Image
	PlayerHurt  *ebiten.Image

	// Тайлы
	Tiles map[string]*ebiten.Image

	// Предметы
	Items map[string]*ebiten.Image

	// Враги
	Enemies map[string]*ebiten.Image

	// Фон
	Background *ebiten.Image
}

// LoadSpriteSheet загружает все спрайты из папки assets/sprites
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		Tiles:   make(map[string]*ebiten.Image),
		Items:   make(map[string]*ebiten.Image),
		Enemies: make(map[string]*ebiten.Image),
	}

	basePath := "assets/sprites"

	// Загрузка спрайта игрока
	if err := ss.loadPlayerSprites(basePath); err != nil {
		fmt.Println("Warning: player sprite error:", err)
	}

	// Загрузка тайлов
	if err := ss.loadTiles(basePath); err != nil {
		fmt.Println("Warning: tiles error:", err)
	}

	// Загрузка предметов
	if err := ss.loadItems(basePath); err != nil {
		fmt.Println("Warning: items error:", err)
	}

	// Загрузка врагов
	if err := ss.loadEnemies(basePath); err != nil {
		fmt.Println("Warning: enemies error:", err)
	}

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites(basePath string) error {
	// Загрузка спрайтшита
	spriteSheetPath := filepath.Join(basePath, "p1_spritesheet.png")
	img, err := loadEbitenImage(spriteSheetPath)
	if err != nil {
		return err
	}

	// Нарезка спрайтов согласно p1_spritesheet.txt
	// p1_stand = 67 196 66 92
	ss.PlayerStand = cropImage(img, 67, 196, 66, 92)

	// p1_jump = 438 93 67 94
	ss.PlayerJump = cropImage(img, 438, 93, 67, 94)

	// p1_hurt = 438 0 69 92
	ss.PlayerHurt = cropImage(img, 438, 0, 69, 92)

	// Анимация walk (11 кадров)
	// p1_walk01-11 = 72x97
	ss.PlayerWalk = make([]*ebiten.Image, 0, 11)
	walkFrames := [][4]int{
		{0, 0, 72, 97},
		{73, 0, 72, 97},
		{146, 0, 72, 97},
		{0, 98, 72, 97},
		{73, 98, 72, 97},
		{146, 98, 72, 97},
		{219, 0, 72, 97},
		{292, 0, 72, 97},
		{219, 98, 72, 97},
		{365, 0, 72, 97},
		{292, 98, 72, 97},
	}

	for _, frame := range walkFrames {
		ss.PlayerWalk = append(ss.PlayerWalk, cropImage(img, frame[0], frame[1], frame[2], frame[3]))
	}

	return nil
}

// loadTiles загружает тайлы
func (ss *SpriteSheet) loadTiles(basePath string) error {
	tilesPath := filepath.Join(basePath, "Tiles")

	// Загрузка основных тайлов
	tileFiles := []string{
		"grass.png", "grassMid.png", "grassLeft.png", "grassRight.png",
		"dirt.png", "dirtMid.png", "dirtLeft.png", "dirtRight.png",
		"stone.png", "stoneMid.png", "stoneLeft.png", "stoneRight.png",
		"brickWall.png", "castle.png", "castleMid.png",
		"ladder_mid.png", "ladder_top.png",
		"box.png", "boxEmpty.png",
		"door_closedMid.png", "door_openMid.png",
		"signExit.png",
	}

	for _, file := range tileFiles {
		path := filepath.Join(tilesPath, file)
		tile, err := loadEbitenImage(path)
		if err != nil {
			continue
		}
		name := file[:len(file)-4] // без .png
		ss.Tiles[name] = tile
	}

	return nil
}

// loadItems загружает предметы
func (ss *SpriteSheet) loadItems(basePath string) error {
	itemsPath := filepath.Join(basePath, "items")

	itemFiles := []string{
		"coinGold.png", "coinSilver.png", "coinBronze.png",
		"gemRed.png", "gemBlue.png", "gemGreen.png", "gemYellow.png",
		"star.png", "mushroomRed.png", "mushroomBrown.png",
		"bomb.png", "keyBlue.png", "keyGreen.png", "keyRed.png",
	}

	for _, file := range itemFiles {
		path := filepath.Join(itemsPath, file)
		item, err := loadEbitenImage(path)
		if err != nil {
			continue
		}
		name := file[:len(file)-4]
		ss.Items[name] = item
	}

	return nil
}

// loadEnemies загружает врагов
func (ss *SpriteSheet) loadEnemies(basePath string) error {
	enemiesPath := filepath.Join(basePath, "enemies")

	enemyFiles := []string{
		"bat_fly.png", "bat_hit.png", "bat_dead.png",
		"bee_fly.png", "bee_hit.png", "bee_dead.png",
		"fly_fly.png", "fly_hit.png", "fly_dead.png",
		"ghost_normal.png", "ghost_hit.png", "ghost_dead.png",
		"slime_walk.png", "slime_hit.png", "slime_dead.png",
		"snake_walk.png", "snake_hit.png", "snake_dead.png",
		"spider_walk1.png", "spider_walk2.png", "spider_hit.png",
		"ladyBug_walk.png", "ladyBug_hit.png",
		"frog.png", "frog_hit.png", "frog_leap.png",
		"snail_walk.png", "snail_hit.png",
	}

	for _, file := range enemyFiles {
		path := filepath.Join(enemiesPath, file)
		enemy, err := loadEbitenImage(path)
		if err != nil {
			continue
		}
		name := file[:len(file)-4]
		ss.Enemies[name] = enemy
	}

	return nil
}

// loadEbitenImage загружает изображение как ebiten.Image
func loadEbitenImage(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

// cropImage вырезает часть изображения
func cropImage(src *ebiten.Image, x, y, width, height int) *ebiten.Image {
	bounds := src.Bounds()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+width > bounds.Max.X {
		width = bounds.Max.X - x
	}
	if y+height > bounds.Max.Y {
		height = bounds.Max.Y - y
	}

	if width <= 0 || height <= 0 {
		return ebiten.NewImage(1, 1)
	}

	dst := ebiten.NewImage(width, height)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(x), -float64(y))
	dst.DrawImage(src, opts)

	return dst
}

// GetPlayerSprite возвращает спрайт игрока по имени
func (ss *SpriteSheet) GetPlayerSprite(name string) *ebiten.Image {
	switch name {
	case "stand":
		return ss.PlayerStand
	case "jump":
		return ss.PlayerJump
	case "hurt":
		return ss.PlayerHurt
	}
	return nil
}

// GetPlayerWalkFrame возвращает кадр анимации ходьбы
func (ss *SpriteSheet) GetPlayerWalkFrame(frame int) *ebiten.Image {
	if frame < 0 || frame >= len(ss.PlayerWalk) {
		return ss.PlayerStand
	}
	return ss.PlayerWalk[frame]
}

// GetTile возвращает тайл по имени
func (ss *SpriteSheet) GetTile(name string) *ebiten.Image {
	if tile, ok := ss.Tiles[name]; ok {
		return tile
	}
	return nil
}

// GetItem возвращает предмет по имени
func (ss *SpriteSheet) GetItem(name string) *ebiten.Image {
	if item, ok := ss.Items[name]; ok {
		return item
	}
	return nil
}

// GetEnemySprite возвращает спрайт врага по имени
func (ss *SpriteSheet) GetEnemySprite(name string) *ebiten.Image {
	if enemy, ok := ss.Enemies[name]; ok {
		return enemy
	}
	return nil
}

// CreatePlaceholder создаёт изображение-заглушку
func CreatePlaceholder(width, height int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	fillImage(img, c)
	return img
}

// fillImage заполняет изображение цветом
func fillImage(img *ebiten.Image, c color.Color) {
	w, h := img.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
}

// Package sprite - загрузка и управление спрайтами из PNG файлов
// Go365 Day 90 - City Survivor
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

// SpriteRect - прямоугольник спрайта на атласе
type SpriteRect struct {
	Name   string
	X, Y   int
	Width  int
	Height int
}

// Animation - последовательность кадров анимации
type Animation struct {
	Name      string
	Frames    []*ebiten.Image
	Duration  float64
	Loop      bool
	FrameTime float64
}

// NewAnimation создаёт новую анимацию
func NewAnimation(name string, frames []*ebiten.Image, fps float64, loop bool) *Animation {
	return &Animation{
		Name:      name,
		Frames:    frames,
		Duration:  float64(len(frames)) / fps,
		Loop:      loop,
		FrameTime: 1.0 / fps,
	}
}

// SpriteSheet - атлас спрайтов с анимациями
type SpriteSheet struct {
	PlayerSprites  map[string]*ebiten.Image
	PlayerAnims    map[string]*Animation
	EnemySprites   map[string]*ebiten.Image
	EnemyAnims     map[string]*Animation
	ItemSprites    map[string]*ebiten.Image
	TileSprites    map[string]*ebiten.Image
	BGSprite       *ebiten.Image
}

// LoadSpriteSheet загружает все спрайты из файлов
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		PlayerSprites: make(map[string]*ebiten.Image),
		PlayerAnims:   make(map[string]*Animation),
		EnemySprites:  make(map[string]*ebiten.Image),
		EnemyAnims:    make(map[string]*Animation),
		ItemSprites:   make(map[string]*ebiten.Image),
		TileSprites:   make(map[string]*ebiten.Image),
	}

	// Загрузка спрайтов игрока (используем набор p1)
	if err := ss.loadPlayerSprites(); err != nil {
		return nil, fmt.Errorf("player sprites: %w", err)
	}

	// Загрузка врагов
	if err := ss.loadEnemySprites(); err != nil {
		return nil, fmt.Errorf("enemy sprites: %w", err)
	}

	// Загрузка предметов
	if err := ss.loadItemSprites(); err != nil {
		return nil, fmt.Errorf("item sprites: %w", err)
	}

	// Загрузка тайлов
	if err := ss.loadTileSprites(); err != nil {
		return nil, fmt.Errorf("tile sprites: %w", err)
	}

	// Загрузка фона
	if err := ss.loadBackground(); err != nil {
		// Фон не критичен
		fmt.Fprintf(os.Stderr, "Warning: background not loaded: %v\n", err)
	}

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites() error {
	basePath := "assets/Player"

	// Загрузка спрайтов монстра-паука (furr_walking_monster)
	// Idle анимация (16 кадров)
	for i := 0; i < 16; i++ {
		filename := fmt.Sprintf("skeleton-idle_%d.png", i)
		path := filepath.Join(basePath, "monster_idle", filename)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: monster idle %d not loaded: %v\n", i, err)
			continue
		}
		if i == 0 {
			ss.PlayerSprites["stand"] = img
		}
	}

	// Walk анимация (13 кадров)
	walkFrames := make([]*ebiten.Image, 0, 13)
	for i := 0; i < 13; i++ {
		filename := fmt.Sprintf("skeleton-walking_%d.png", i)
		path := filepath.Join(basePath, "monster_walk", filename)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: monster walk %d not loaded: %v\n", i, err)
			continue
		}
		walkFrames = append(walkFrames, img)
	}

	if len(walkFrames) > 0 {
		ss.PlayerAnims["walk"] = NewAnimation("walk", walkFrames, 12, true)
	}

	// Hurt анимация (8 кадров)
	hurtFrames := make([]*ebiten.Image, 0, 8)
	for i := 0; i < 8; i++ {
		filename := fmt.Sprintf("skeleton-got_hit_%d.png", i)
		path := filepath.Join(basePath, "monster_hurt", filename)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: monster hurt %d not loaded: %v\n", i, err)
			continue
		}
		hurtFrames = append(hurtFrames, img)
	}

	if len(hurtFrames) > 0 {
		ss.PlayerAnims["hurt"] = NewAnimation("hurt", hurtFrames, 15, false)
	}

	// Death анимация (26 кадров)
	deathFrames := make([]*ebiten.Image, 0, 26)
	for i := 0; i < 26; i++ {
		filename := fmt.Sprintf("skeleton-defeated_%d.png", i)
		path := filepath.Join(basePath, "monster_death", filename)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: monster death %d not loaded: %v\n", i, err)
			continue
		}
		deathFrames = append(deathFrames, img)
	}

	if len(deathFrames) > 0 {
		ss.PlayerAnims["death"] = NewAnimation("death", deathFrames, 15, false)
	}

	// Загрузка отдельных спрайтов для совместимости
	spriteFiles := []string{
		"monster_idle/skeleton-idle_0.png", // stand
		"monster_walk/skeleton-walking_0.png", // front
	}

	for i, file := range spriteFiles {
		path := filepath.Join(basePath, file)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: monster sprite %s not loaded: %v\n", file, err)
			continue
		}
		if i == 0 {
			ss.PlayerSprites["stand"] = img
		} else {
			ss.PlayerSprites["front"] = img
		}
	}

	// Прыжок и присед используем как stand (заглушка)
	if ss.PlayerSprites["stand"] != nil {
		ss.PlayerSprites["jump"] = ss.PlayerSprites["stand"]
		ss.PlayerSprites["duck"] = ss.PlayerSprites["stand"]
		ss.PlayerSprites["hurt"] = ss.PlayerSprites["stand"]
	}

	return nil
}

// loadEnemySprites загружает спрайты врагов
func (ss *SpriteSheet) loadEnemySprites() error {
	basePath := "assets/Enemies"

	// Загрузка основных врагов
	enemies := []string{
		"fishDead.png",
		"fishSwim1.png",
		"fishSwim2.png",
		"flyDead.png",
		"flyFly1.png",
		"flyFly2.png",
		"slimeDead.png",
		"slimeWalk1.png",
		"slimeWalk2.png",
		"snailWalk1.png",
		"snailWalk2.png",
		"blockerMad.png",
		"blockerSad.png",
		"pokerMad.png",
		"pokerSad.png",
	}

	for _, file := range enemies {
		path := filepath.Join(basePath, file)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: enemy %s not loaded: %v\n", file, err)
			continue
		}
		key := file[:len(file)-4]
		ss.EnemySprites[key] = img
	}

	// Анимация плавания рыбы
	fishFrames := []*ebiten.Image{
		ss.EnemySprites["fishSwim1"],
		ss.EnemySprites["fishSwim2"],
	}
	if fishFrames[0] != nil && fishFrames[1] != nil {
		ss.EnemyAnims["fishSwim"] = NewAnimation("fishSwim", fishFrames, 8, true)
	}

	// Анимация полёта мухи
	flyFrames := []*ebiten.Image{
		ss.EnemySprites["flyFly1"],
		ss.EnemySprites["flyFly2"],
	}
	if flyFrames[0] != nil && flyFrames[1] != nil {
		ss.EnemyAnims["flyFly"] = NewAnimation("flyFly", flyFrames, 12, true)
	}

	// Анимация ходьбы слайма
	slimeFrames := []*ebiten.Image{
		ss.EnemySprites["slimeWalk1"],
		ss.EnemySprites["slimeWalk2"],
	}
	if slimeFrames[0] != nil && slimeFrames[1] != nil {
		ss.EnemyAnims["slimeWalk"] = NewAnimation("slimeWalk", slimeFrames, 6, true)
	}

	// Анимация ходьбы улитки
	snailFrames := []*ebiten.Image{
		ss.EnemySprites["snailWalk1"],
		ss.EnemySprites["snailWalk2"],
	}
	if snailFrames[0] != nil && snailFrames[1] != nil {
		ss.EnemyAnims["snailWalk"] = NewAnimation("snailWalk", snailFrames, 5, true)
	}

	return nil
}

// loadItemSprites загружает спрайты предметов
func (ss *SpriteSheet) loadItemSprites() error {
	basePath := "assets/Items"

	items := []string{
		"coinBronze.png",
		"coinSilver.png",
		"coinGold.png",
		"gemBlue.png",
		"gemGreen.png",
		"gemRed.png",
		"gemYellow.png",
		"keyBlue.png",
		"keyGreen.png",
		"keyRed.png",
		"keyYellow.png",
		"star.png",
		"bomb.png",
		"mushroomRed.png",
		"mushroomBrown.png",
	}

	for _, file := range items {
		path := filepath.Join(basePath, file)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: item %s not loaded: %v\n", file, err)
			continue
		}
		key := file[:len(file)-4]
		ss.ItemSprites[key] = img
	}

	// Предметы из папки Food
	foodPath := "assets/Food"
	foodItems := []string{
		"food-OCAL.png",
		"food-Glitch.png",
		"food-DCSS.png",
	}
	for _, file := range foodItems {
		path := filepath.Join(foodPath, file)
		img, err := loadImage(path)
		if err != nil {
			continue
		}
		key := file[:len(file)-4]
		ss.ItemSprites[key] = img
	}

	return nil
}

// loadTileSprites загружает тайлы
func (ss *SpriteSheet) loadTileSprites() error {
	basePath := "assets/Tiles"

	// Основные тайлы земли и травы
	tiles := []string{
		"grass.png",
		"grassLeft.png",
		"grassRight.png",
		"grassMid.png",
		"grassCenter.png",
		"grassHalf.png",
		"grassHalfLeft.png",
		"grassHalfMid.png",
		"grassHalfRight.png",
		"dirt.png",
		"dirtLeft.png",
		"dirtRight.png",
		"dirtMid.png",
		"dirtCenter.png",
		"brickWall.png",
		"box.png",
		"boxAlt.png",
		"boxCoin.png",
		"boxItem.png",
		"boxEmpty.png",
		"fence.png",
		"ladder_mid.png",
		"ladder_top.png",
		"door_closedTop.png",
		"door_closedMid.png",
		"lock_blue.png",
		"bridge.png",
		"bridgeLogs.png",
	}

	for _, file := range tiles {
		path := filepath.Join(basePath, file)
		img, err := loadImage(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: tile %s not loaded: %v\n", file, err)
			continue
		}
		key := file[:len(file)-4]
		ss.TileSprites[key] = img
	}

	return nil
}

// loadBackground загружает фон
func (ss *SpriteSheet) loadBackground() error {
	// Пробуем несколько вариантов фона
	bgFiles := []string{
		"assets/bg_forest.png",
		"assets/bg_forest_layers/bg_forecast.png",
	}

	for _, path := range bgFiles {
		img, err := loadImage(path)
		if err == nil {
			ss.BGSprite = img
			return nil
		}
	}

	// Если не нашли, создаём заглушку
	ss.BGSprite = createPlaceholderBG()
	return nil
}

// loadImage загружает изображение из файла
func loadImage(path string) (*ebiten.Image, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return img, nil
}

// createPlaceholderBG создаёт простой фон-заглушку
func createPlaceholderBG() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Градиент неба
	for y := 0; y < 720; y++ {
		ratio := float64(y) / 720.0
		r := uint8(60 + ratio*40)
		g := uint8(60 + ratio*30)
		b := uint8(80 + ratio*50)
		for x := 0; x < 1280; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return ebiten.NewImageFromImage(img)
}

// GetPlayerSprite возвращает спрайт игрока по имени
func (ss *SpriteSheet) GetPlayerSprite(name string) *ebiten.Image {
	if sprite, ok := ss.PlayerSprites[name]; ok {
		return sprite
	}
	// Заглушка
	return ss.PlayerSprites["stand"]
}

// GetPlayerAnim возвращает анимацию игрока
func (ss *SpriteSheet) GetPlayerAnim(name string) *Animation {
	return ss.PlayerAnims[name]
}

// GetEnemySprite возвращает спрайт врага
func (ss *SpriteSheet) GetEnemySprite(name string) *ebiten.Image {
	if sprite, ok := ss.EnemySprites[name]; ok {
		return sprite
	}
	// Первая доступная заглушка
	for _, v := range ss.EnemySprites {
		return v
	}
	return nil
}

// GetEnemyAnim возвращает анимацию врага
func (ss *SpriteSheet) GetEnemyAnim(name string) *Animation {
	return ss.EnemyAnims[name]
}

// GetItemSprite возвращает спрайт предмета
func (ss *SpriteSheet) GetItemSprite(name string) *ebiten.Image {
	if sprite, ok := ss.ItemSprites[name]; ok {
		return sprite
	}
	// Заглушка - монета
	return ss.ItemSprites["coinGold"]
}

// GetTileSprite возвращает спрайт тайла
func (ss *SpriteSheet) GetTileSprite(name string) *ebiten.Image {
	if sprite, ok := ss.TileSprites[name]; ok {
		return sprite
	}
	// Заглушка - земля
	return ss.TileSprites["dirt"]
}

// GetBackground возвращает фон
func (ss *SpriteSheet) GetBackground() *ebiten.Image {
	return ss.BGSprite
}

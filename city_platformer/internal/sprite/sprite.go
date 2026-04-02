// Package sprite - загрузка и управление спрайтами для City Platformer
// Go365 Day 93 - Neon Runner: Cyber Escape
package sprite

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet хранит все загруженные спрайты
type SpriteSheet struct {
	// Игрок
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerHurt   *ebiten.Image
	PlayerDuck   *ebiten.Image
	PlayerFront  *ebiten.Image

	// Тайлы
	Tiles map[string]*ebiten.Image

	// Предметы
	Items map[string]*ebiten.Image

	// Враги
	Enemies map[string]*ebiten.Image

	// HUD
	HUD map[string]*ebiten.Image

	// Фон
	Background      *ebiten.Image
	BackgroundCastle *ebiten.Image
}

// LoadSpriteSheet загружает все спрайты из папки assets/sprites
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		Tiles:   make(map[string]*ebiten.Image),
		Items:   make(map[string]*ebiten.Image),
		Enemies: make(map[string]*ebiten.Image),
		HUD:     make(map[string]*ebiten.Image),
	}

	basePath := "assets/sprites"

	// Загрузка спрайтов игрока
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

	// Загрузка HUD
	if err := ss.loadHUD(basePath); err != nil {
		fmt.Println("Warning: HUD error:", err)
	}

	// Загрузка фона
	if err := ss.loadBackground(basePath); err != nil {
		fmt.Println("Warning: background error:", err)
	}

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites(basePath string) error {
	playerPath := filepath.Join(basePath, "Player")

	// Загрузка отдельных спрайтов
	ss.PlayerStand = ss.loadImage(filepath.Join(playerPath, "p1_stand.png"))
	ss.PlayerJump = ss.loadImage(filepath.Join(playerPath, "p1_jump.png"))
	ss.PlayerHurt = ss.loadImage(filepath.Join(playerPath, "p1_hurt.png"))
	ss.PlayerDuck = ss.loadImage(filepath.Join(playerPath, "p1_duck.png"))
	ss.PlayerFront = ss.loadImage(filepath.Join(playerPath, "p1_front.png"))

	// Загрузка анимации ходьбы (11 кадров)
	walkPath := filepath.Join(playerPath, "p1_walk", "PNG")
	ss.PlayerWalk = make([]*ebiten.Image, 0, 11)

	for i := 1; i <= 11; i++ {
		frameName := fmt.Sprintf("p1_walk%02d.png", i)
		framePath := filepath.Join(walkPath, frameName)
		if img := ss.loadImage(framePath); img != nil {
			ss.PlayerWalk = append(ss.PlayerWalk, img)
		}
	}

	// Сортировка кадров по имени файла
	if len(ss.PlayerWalk) > 0 {
		sort.Slice(ss.PlayerWalk, func(i, j int) bool {
			boundsI := ss.PlayerWalk[i].Bounds()
			boundsJ := ss.PlayerWalk[j].Bounds()
			return boundsI.Dx() < boundsJ.Dx() || (boundsI.Dx() == boundsJ.Dx() && boundsI.Dy() < boundsJ.Dy())
		})
	}

	return nil
}

// loadTiles загружает тайлы
func (ss *SpriteSheet) loadTiles(basePath string) error {
	tilesPath := filepath.Join(basePath, "Tiles")

	// Основные тайлы
	tileFiles := []string{
		// Земля
		"grass.png", "grassMid.png", "grassLeft.png", "grassRight.png",
		"grassHalf.png", "grassHalfLeft.png", "grassHalfRight.png",
		"dirt.png", "dirtMid.png", "dirtLeft.png", "dirtRight.png",
		"dirtHalf.png", "dirtHalfLeft.png", "dirtHalfRight.png",
		// Камень
		"stone.png", "stoneMid.png", "stoneLeft.png", "stoneRight.png",
		"stoneHalf.png", "stoneHalfLeft.png", "stoneHalfRight.png",
		// Кирпичи
		"brickWall.png",
		// Замок
		"castle.png", "castleMid.png", "castleLeft.png", "castleRight.png",
		"castleHalf.png", "castleHalfLeft.png", "castleHalfRight.png",
		// Лестницы
		"ladder_mid.png", "ladder_top.png",
		// Двери
		"door_closedMid.png", "door_openMid.png",
		// Выход
		"signExit.png",
		// Опасности
		"spikes.png",
	}

	for _, file := range tileFiles {
		path := filepath.Join(tilesPath, file)
		if tile := ss.loadImage(path); tile != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Tiles[name] = tile
		}
	}

	return nil
}

// loadItems загружает предметы
func (ss *SpriteSheet) loadItems(basePath string) error {
	itemsPath := filepath.Join(basePath, "Items")

	itemFiles := []string{
		// Монеты
		"coinGold.png", "coinSilver.png", "coinBronze.png",
		// Кристаллы
		"gemRed.png", "gemBlue.png", "gemGreen.png", "gemYellow.png",
		// Бонусы
		"star.png", "mushroomRed.png", "mushroomBrown.png",
		// Ключи
		"keyBlue.png", "keyGreen.png", "keyRed.png", "keyYellow.png",
		// Облака
		"cloud1.png", "cloud2.png", "cloud3.png",
		// Флаги
		"flagBlue.png", "flagGreen.png", "flagRed.png", "flagYellow.png",
	}

	for _, file := range itemFiles {
		path := filepath.Join(itemsPath, file)
		if item := ss.loadImage(path); item != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Items[name] = item
		}
	}

	return nil
}

// loadEnemies загружает врагов
func (ss *SpriteSheet) loadEnemies(basePath string) error {
	enemiesPath := filepath.Join(basePath, "Enemies")

	enemyFiles := []string{
		// Муха
		"flyFly1.png", "flyFly2.png",
		// Летучая мышь
		"bat_fly.png", "bat.png", "bat_hit.png", "bat_dead.png",
		// Слайм
		"slimeWalk1.png", "slimeWalk2.png", "slime.png",
		// Змея
		"snakeWalk.png", "snake.png", "snake_hit.png", "snake_dead.png",
		// Паук
		"spider_walk1.png", "spider_walk2.png", "spider.png", "spider_hit.png",
		// Призрак
		"ghost_normal.png", "ghost.png", "ghost_hit.png", "ghost_dead.png",
		// Пчела
		"bee_fly.png", "bee.png", "bee_hit.png", "bee_dead.png",
		// Божья коровка
		"ladyBug_walk.png", "ladyBug.png", "ladyBug_hit.png",
		// Лягушка
		"frog.png", "frog_hit.png", "frog_leap.png", "frog_dead.png",
		// Улитка
		"snailWalk1.png", "snailWalk2.png", "snail_walk.png", "snail_hit.png",
	}

	for _, file := range enemyFiles {
		path := filepath.Join(enemiesPath, file)
		if enemy := ss.loadImage(path); enemy != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Enemies[name] = enemy
		}
	}

	return nil
}

// loadHUD загружает элементы интерфейса
func (ss *SpriteSheet) loadHUD(basePath string) error {
	hudPath := filepath.Join(basePath, "HUD")

	hudFiles := []string{
		"hud_heartFull.png", "hud_heartHalf.png", "hud_heartEmpty.png",
		"hud_gem_red.png", "hud_gem_blue.png", "hud_gem_green.png", "hud_gem_yellow.png",
		"hud_coins.png", "hud_x.png",
		"hud_0.png", "hud_1.png", "hud_2.png", "hud_3.png",
		"hud_4.png", "hud_5.png", "hud_6.png", "hud_7.png", "hud_8.png", "hud_9.png",
	}

	for _, file := range hudFiles {
		path := filepath.Join(hudPath, file)
		if hud := ss.loadImage(path); hud != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.HUD[name] = hud
		}
	}

	return nil
}

// loadBackground загружает фоны
func (ss *SpriteSheet) loadBackground(basePath string) error {
	bgPath := filepath.Join(basePath, "Background")

	ss.Background = ss.loadImage(filepath.Join(bgPath, "bg.png"))
	ss.BackgroundCastle = ss.loadImage(filepath.Join(bgPath, "bg_castle.png"))

	return nil
}

// loadImage загружает одно изображение
func (ss *SpriteSheet) loadImage(path string) *ebiten.Image {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
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
	case "duck":
		return ss.PlayerDuck
	case "front":
		return ss.PlayerFront
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

// GetHUD возвращает элемент HUD по имени
func (ss *SpriteSheet) GetHUD(name string) *ebiten.Image {
	if hud, ok := ss.HUD[name]; ok {
		return hud
	}
	return nil
}

// GetBackground возвращает фон
func (ss *SpriteSheet) GetBackground(castle bool) *ebiten.Image {
	if castle && ss.BackgroundCastle != nil {
		return ss.BackgroundCastle
	}
	return ss.Background
}

// CreatePlaceholder создаёт изображение-заглушку
func CreatePlaceholder(width, height int, r, g, b uint8) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(colorRGBA(r, g, b, 255))
	return img
}

func colorRGBA(r, g, b, a uint8) color.Color {
	return color.NRGBA{r, g, b, a}
}

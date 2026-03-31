// Package sprite - загрузка и управление спрайтами
// Go365 Day 91 - Cyber City Runner
package sprite

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// Animation - анимация спрайта
type Animation struct {
	Frames    []*ebiten.Image
	FrameTime float64
	Loop      bool
	Name      string
}

// SpriteSheet - коллекция спрайтов и анимаций
type SpriteSheet struct {
	playerSprites  map[string]*ebiten.Image
	playerAnims    map[string]*Animation
	enemySprites   map[string]*ebiten.Image
	enemyAnims     map[string]*Animation
	itemSprites    map[string]*ebiten.Image
	background     *ebiten.Image
	tileSprites    map[string]*ebiten.Image
}

// LoadSpriteSheet загружает спрайты из assets
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		playerSprites: make(map[string]*ebiten.Image),
		playerAnims:   make(map[string]*Animation),
		enemySprites:  make(map[string]*ebiten.Image),
		enemyAnims:    make(map[string]*Animation),
		itemSprites:   make(map[string]*ebiten.Image),
		tileSprites:   make(map[string]*ebiten.Image),
	}

	// Загрузка спрайтов игрока
	ss.loadPlayerSprites()

	// Загрузка спрайтов врагов
	ss.loadEnemySprites()

	// Загрузка предметов
	ss.loadItemSprites()

	// Загрузка фона
	ss.loadBackground()

	// Загрузка тайлов
	ss.loadTileSprites()

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites() {
	// Пробуем загрузить из PlatformerComplete
	basePath := "assets/Base pack/Player"

	// Загрузка статичных спрайтов
	ss.playerSprites["stand"] = ss.loadImage(filepath.Join(basePath, "p1_stand.png"))
	ss.playerSprites["jump"] = ss.loadImage(filepath.Join(basePath, "p1_jump.png"))
	ss.playerSprites["duck"] = ss.loadImage(filepath.Join(basePath, "p1_duck.png"))
	ss.playerSprites["hurt"] = ss.loadImage(filepath.Join(basePath, "p1_hurt.png"))

	// Загрузка анимации ходьбы
	walkFrames := ss.loadAnimationFrames(filepath.Join(basePath, "p1_walk"), 15)
	if len(walkFrames) > 0 {
		ss.playerAnims["walk"] = &Animation{
			Frames:    walkFrames,
			FrameTime: 0.1,
			Loop:      true,
			Name:      "walk",
		}
	}

	// Загрузка анимации бега (используем walk как заглушку)
	ss.playerAnims["run"] = ss.playerAnims["walk"]
}

// loadEnemySprites загружает спрайты врагов
func (ss *SpriteSheet) loadEnemySprites() {
	basePath := "assets/Base pack/Enemies"

	// Слайм
	ss.enemySprites["slimeWalk1"] = ss.loadImage(filepath.Join(basePath, "slimeWalk1.png"))
	ss.enemySprites["slimeWalk2"] = ss.loadImage(filepath.Join(basePath, "slimeWalk2.png"))
	ss.enemySprites["slimeDead"] = ss.loadImage(filepath.Join(basePath, "slimeDead.png"))

	// Рыба
	ss.enemySprites["fishSwim1"] = ss.loadImage(filepath.Join(basePath, "fishSwim1.png"))
	ss.enemySprites["fishSwim2"] = ss.loadImage(filepath.Join(basePath, "fishSwim2.png"))
	ss.enemySprites["fishDead"] = ss.loadImage(filepath.Join(basePath, "fishDead.png"))

	// Муха
	ss.enemySprites["flyFly1"] = ss.loadImage(filepath.Join(basePath, "flyFly1.png"))
	ss.enemySprites["flyFly2"] = ss.loadImage(filepath.Join(basePath, "flyFly2.png"))

	// Улитка
	ss.enemySprites["snailWalk1"] = ss.loadImage(filepath.Join(basePath, "snailWalk1.png"))
	ss.enemySprites["snailWalk2"] = ss.loadImage(filepath.Join(basePath, "snailWalk2.png"))

	// Blocker
	ss.enemySprites["blockerBody"] = ss.loadImage(filepath.Join(basePath, "blockerBody.png"))
	ss.enemySprites["blockerMad"] = ss.loadImage(filepath.Join(basePath, "blockerMad.png"))

	// Анимация ходьбы слайма
	ss.enemyAnims["slimeWalk"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["slimeWalk1"], ss.enemySprites["slimeWalk2"]},
		FrameTime: 0.15,
		Loop:      true,
		Name:      "slimeWalk",
	}

	// Анимация плавания рыбы
	ss.enemyAnims["fishSwim"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["fishSwim1"], ss.enemySprites["fishSwim2"]},
		FrameTime: 0.15,
		Loop:      true,
		Name:      "fishSwim",
	}

	// Анимация полёта мухи
	ss.enemyAnims["flyFly"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["flyFly1"], ss.enemySprites["flyFly2"]},
		FrameTime: 0.1,
		Loop:      true,
		Name:      "flyFly",
	}

	// Анимация ходьбы улитки
	ss.enemyAnims["snailWalk"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["snailWalk1"], ss.enemySprites["snailWalk2"]},
		FrameTime: 0.2,
		Loop:      true,
		Name:      "snailWalk",
	}
}

// loadItemSprites загружает спрайты предметов
func (ss *SpriteSheet) loadItemSprites() {
	basePath := "assets/Base pack/Items"

	ss.itemSprites["coinGold"] = ss.loadImage(filepath.Join(basePath, "coinGold.png"))
	ss.itemSprites["coinSilver"] = ss.loadImage(filepath.Join(basePath, "coinSilver.png"))
	ss.itemSprites["coinBronze"] = ss.loadImage(filepath.Join(basePath, "coinBronze.png"))
	ss.itemSprites["gemRed"] = ss.loadImage(filepath.Join(basePath, "gemRed.png"))
	ss.itemSprites["gemBlue"] = ss.loadImage(filepath.Join(basePath, "gemBlue.png"))
	ss.itemSprites["gemGreen"] = ss.loadImage(filepath.Join(basePath, "gemGreen.png"))
	ss.itemSprites["gemYellow"] = ss.loadImage(filepath.Join(basePath, "gemYellow.png"))
	ss.itemSprites["star"] = ss.loadImage(filepath.Join(basePath, "star.png"))
	ss.itemSprites["mushroomRed"] = ss.loadImage(filepath.Join(basePath, "mushroomRed.png"))
	ss.itemSprites["mushroomBrown"] = ss.loadImage(filepath.Join(basePath, "mushroomBrown.png"))
}

// loadBackground загружает фон
func (ss *SpriteSheet) loadBackground() {
	ss.background = ss.loadImage("assets/bg.png")
}

// loadTileSprites загружает спрайты тайлов
func (ss *SpriteSheet) loadTileSprites() {
	basePath := "assets/Base pack/Tiles"

	ss.tileSprites["dirt"] = ss.loadImage(filepath.Join(basePath, "dirt.png"))
	ss.tileSprites["grass"] = ss.loadImage(filepath.Join(basePath, "grass.png"))
	ss.tileSprites["brickWall"] = ss.loadImage(filepath.Join(basePath, "brickWall.png"))
	ss.tileSprites["box"] = ss.loadImage(filepath.Join(basePath, "box.png"))
	ss.tileSprites["ladder_mid"] = ss.loadImage(filepath.Join(basePath, "ladder_mid.png"))
	ss.tileSprites["ladder_top"] = ss.loadImage(filepath.Join(basePath, "ladder_top.png"))
	ss.tileSprites["spikes"] = ss.loadImage(filepath.Join(basePath, "spikes.png"))
}

// loadImage загружает изображение
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

// loadAnimationFrames загружает кадры анимации
func (ss *SpriteSheet) loadAnimationFrames(dir string, maxFrames int) []*ebiten.Image {
	frames := make([]*ebiten.Image, 0)

	for i := 1; i <= maxFrames; i++ {
		path := filepath.Join(dir, "p1_walk"+padNumber(i)+".png")
		if img := ss.loadImage(path); img != nil {
			frames = append(frames, img)
		}
	}

	return frames
}

// padNumber дополняет номер нулями
func padNumber(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// GetPlayerSprite возвращает спрайт игрока
func (ss *SpriteSheet) GetPlayerSprite(name string) *ebiten.Image {
	return ss.playerSprites[name]
}

// GetPlayerAnim возвращает анимацию игрока
func (ss *SpriteSheet) GetPlayerAnim(name string) *Animation {
	return ss.playerAnims[name]
}

// GetEnemySprite возвращает спрайт врага
func (ss *SpriteSheet) GetEnemySprite(name string) *ebiten.Image {
	return ss.enemySprites[name]
}

// GetEnemyAnim возвращает анимацию врага
func (ss *SpriteSheet) GetEnemyAnim(name string) *Animation {
	return ss.enemyAnims[name]
}

// GetItemSprite возвращает спрайт предмета
func (ss *SpriteSheet) GetItemSprite(name string) *ebiten.Image {
	return ss.itemSprites[name]
}

// GetTileSprite возвращает спрайт тайла
func (ss *SpriteSheet) GetTileSprite(name string) *ebiten.Image {
	return ss.tileSprites[name]
}

// GetBackground возвращает фон
func (ss *SpriteSheet) GetBackground() *ebiten.Image {
	return ss.background
}

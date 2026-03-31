// Package sprite - загрузка и управление спрайтами
// Go365 Day 91 - Sunny Adventure
package sprite

import (
	"fmt"
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
	backgrounds    map[string]*ebiten.Image
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
		backgrounds:   make(map[string]*ebiten.Image),
		tileSprites:   make(map[string]*ebiten.Image),
	}

	// Загрузка спрайтов игрока (солнышко)
	ss.loadPlayerSprites()

	// Загрузка спрайтов врагов
	ss.loadEnemySprites()

	// Загрузка предметов
	ss.loadItemSprites()

	// Загрузка фонов
	ss.loadBackgrounds()

	// Загрузка тайлов
	ss.loadTileSprites()

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites() {
	basePath := "assets/Base pack/Player"

	// Загружаем все спрайты персонажа p1
	ss.playerSprites["stand"] = ss.loadImage(filepath.Join(basePath, "p1_stand.png"))
	ss.playerSprites["jump"] = ss.loadImage(filepath.Join(basePath, "p1_jump.png"))
	ss.playerSprites["duck"] = ss.loadImage(filepath.Join(basePath, "p1_duck.png"))
	ss.playerSprites["hurt"] = ss.loadImage(filepath.Join(basePath, "p1_hurt.png"))
	ss.playerSprites["front"] = ss.loadImage(filepath.Join(basePath, "p1_front.png"))

	// Загружаем анимацию ходьбы из PNG файлов
	walkPath := filepath.Join(basePath, "p1_walk", "PNG")
	walkFrames := make([]*ebiten.Image, 0)
	for i := 1; i <= 11; i++ {
		frame := ss.loadImage(filepath.Join(walkPath, fmt.Sprintf("p1_walk%02d.png", i)))
		if frame != nil {
			walkFrames = append(walkFrames, frame)
		}
	}
	if len(walkFrames) > 0 {
		ss.playerAnims["walk"] = &Animation{
			Frames:    walkFrames,
			FrameTime: 0.1,
			Loop:      true,
			Name:      "walk",
		}
	}
	ss.playerAnims["run"] = ss.playerAnims["walk"]
}

// loadEnemySprites загружает спрайты врагов и друзей
func (ss *SpriteSheet) loadEnemySprites() {
	basePath := "assets/Base pack/Enemies"
	extraPath := "assets/Extra animations and enemies/Enemy sprites"

	// Враги из Base pack
	ss.enemySprites["fly_fly"] = ss.loadImage(filepath.Join(basePath, "flyFly1.png"))
	ss.enemySprites["slime_walk"] = ss.loadImage(filepath.Join(basePath, "slimeWalk1.png"))
	ss.enemySprites["snail_walk"] = ss.loadImage(filepath.Join(basePath, "snailWalk1.png"))

	// Друзья и враги из Extra
	ss.enemySprites["bee_fly"] = ss.loadImage(filepath.Join(extraPath, "bee_fly.png"))
	ss.enemySprites["bee"] = ss.loadImage(filepath.Join(extraPath, "bee.png"))
	ss.enemySprites["ladyBug_walk"] = ss.loadImage(filepath.Join(extraPath, "ladyBug_walk.png"))
	ss.enemySprites["ladyBug"] = ss.loadImage(filepath.Join(extraPath, "ladyBug.png"))
	ss.enemySprites["frog"] = ss.loadImage(filepath.Join(extraPath, "frog.png"))
	ss.enemySprites["frog_leap"] = ss.loadImage(filepath.Join(extraPath, "frog_leap.png"))
	ss.enemySprites["ghost_normal"] = ss.loadImage(filepath.Join(extraPath, "ghost_normal.png"))
	ss.enemySprites["ghost"] = ss.loadImage(filepath.Join(extraPath, "ghost.png"))
	ss.enemySprites["bat_fly"] = ss.loadImage(filepath.Join(extraPath, "bat_fly.png"))
	ss.enemySprites["bat"] = ss.loadImage(filepath.Join(extraPath, "bat.png"))
	ss.enemySprites["spider_walk1"] = ss.loadImage(filepath.Join(extraPath, "spider_walk1.png"))
	ss.enemySprites["spider"] = ss.loadImage(filepath.Join(extraPath, "spider.png"))
	ss.enemySprites["snake_walk"] = ss.loadImage(filepath.Join(extraPath, "snake_walk.png"))
	ss.enemySprites["snake"] = ss.loadImage(filepath.Join(extraPath, "snake.png"))
}

// loadItemSprites загружает спрайты предметов
func (ss *SpriteSheet) loadItemSprites() {
	basePath := "assets/Base pack/Items"

	// Монеты
	ss.itemSprites["coinGold"] = ss.loadImage(filepath.Join(basePath, "coinGold.png"))
	ss.itemSprites["coinSilver"] = ss.loadImage(filepath.Join(basePath, "coinSilver.png"))
	ss.itemSprites["coinBronze"] = ss.loadImage(filepath.Join(basePath, "coinBronze.png"))

	// Кристаллы
	ss.itemSprites["gemRed"] = ss.loadImage(filepath.Join(basePath, "gemRed.png"))
	ss.itemSprites["gemBlue"] = ss.loadImage(filepath.Join(basePath, "gemBlue.png"))
	ss.itemSprites["gemGreen"] = ss.loadImage(filepath.Join(basePath, "gemGreen.png"))
	ss.itemSprites["gemYellow"] = ss.loadImage(filepath.Join(basePath, "gemYellow.png"))

	// Звезда
	ss.itemSprites["star"] = ss.loadImage(filepath.Join(basePath, "star.png"))

	// Грибы
	ss.itemSprites["mushroomRed"] = ss.loadImage(filepath.Join(basePath, "mushroomRed.png"))
	ss.itemSprites["mushroomBrown"] = ss.loadImage(filepath.Join(basePath, "mushroomBrown.png"))

	// Облака
	ss.itemSprites["cloud1"] = ss.loadImage(filepath.Join(basePath, "cloud1.png"))
	ss.itemSprites["cloud2"] = ss.loadImage(filepath.Join(basePath, "cloud2.png"))
	ss.itemSprites["cloud3"] = ss.loadImage(filepath.Join(basePath, "cloud3.png"))

	// Флаги (для выхода)
	ss.itemSprites["flagGreen"] = ss.loadImage(filepath.Join(basePath, "flagGreen.png"))
	ss.itemSprites["flagGreen2"] = ss.loadImage(filepath.Join(basePath, "flagGreen2.png"))
}

// loadBackgrounds загружает фоны
func (ss *SpriteSheet) loadBackgrounds() {
	// Основной фон - голубое небо
	ss.background = ss.loadImage("assets/bg.png")

	// Фоны из Mushroom expansion
	mushroomBg := "assets/Mushroom expansion/Backgrounds"
	ss.backgrounds["grasslands"] = ss.loadImage(filepath.Join(mushroomBg, "bg_grasslands.png"))
	ss.backgrounds["castle"] = ss.loadImage(filepath.Join(mushroomBg, "bg_castle.png"))
	ss.backgrounds["shroom"] = ss.loadImage(filepath.Join(mushroomBg, "bg_shroom.png"))
	ss.backgrounds["desert"] = ss.loadImage(filepath.Join(mushroomBg, "bg_desert.png"))
}

// loadTileSprites загружает спрайты тайлов
func (ss *SpriteSheet) loadTileSprites() {
	basePath := "assets/Base pack/Tiles"

	// Земля и трава
	ss.tileSprites["dirt"] = ss.loadImage(filepath.Join(basePath, "dirt.png"))
	ss.tileSprites["grass"] = ss.loadImage(filepath.Join(basePath, "grass.png"))
	ss.tileSprites["grassHalf"] = ss.loadImage(filepath.Join(basePath, "grassHalf.png"))
	ss.tileSprites["grassLeft"] = ss.loadImage(filepath.Join(basePath, "grassLeft.png"))
	ss.tileSprites["grassRight"] = ss.loadImage(filepath.Join(basePath, "grassRight.png"))
	ss.tileSprites["grassMid"] = ss.loadImage(filepath.Join(basePath, "grassMid.png"))
	ss.tileSprites["dirtMid"] = ss.loadImage(filepath.Join(basePath, "dirtMid.png"))
	ss.tileSprites["dirtHalf"] = ss.loadImage(filepath.Join(basePath, "dirtHalf.png"))

	// Холмы с травой
	ss.tileSprites["grassHillLeft"] = ss.loadImage(filepath.Join(basePath, "grassHillLeft.png"))
	ss.tileSprites["grassHillRight"] = ss.loadImage(filepath.Join(basePath, "grassHillRight.png"))
	ss.tileSprites["grassHillLeft2"] = ss.loadImage(filepath.Join(basePath, "grassHillLeft2.png"))
	ss.tileSprites["grassHillRight2"] = ss.loadImage(filepath.Join(basePath, "grassHillRight2.png"))

	// Кирпич и замок
	ss.tileSprites["brickWall"] = ss.loadImage(filepath.Join(basePath, "brickWall.png"))
	ss.tileSprites["castle"] = ss.loadImage(filepath.Join(basePath, "castle.png"))
	ss.tileSprites["castleCenter"] = ss.loadImage(filepath.Join(basePath, "castleCenter.png"))
	ss.tileSprites["castleHalf"] = ss.loadImage(filepath.Join(basePath, "castleHalf.png"))
	ss.tileSprites["castleMid"] = ss.loadImage(filepath.Join(basePath, "castleMid.png"))

	// Коробки
	ss.tileSprites["box"] = ss.loadImage(filepath.Join(basePath, "box.png"))
	ss.tileSprites["boxAlt"] = ss.loadImage(filepath.Join(basePath, "boxAlt.png"))
	ss.tileSprites["boxCoin"] = ss.loadImage(filepath.Join(basePath, "boxCoin.png"))
	ss.tileSprites["boxItem"] = ss.loadImage(filepath.Join(basePath, "boxItem.png"))
	ss.tileSprites["boxEmpty"] = ss.loadImage(filepath.Join(basePath, "boxEmpty.png"))

	// Лестницы
	ss.tileSprites["ladder_mid"] = ss.loadImage(filepath.Join(basePath, "ladder_mid.png"))
	ss.tileSprites["ladder_top"] = ss.loadImage(filepath.Join(basePath, "ladder_top.png"))

	// Заборы
	ss.tileSprites["fence"] = ss.loadImage(filepath.Join(basePath, "fence.png"))
	ss.tileSprites["fenceBroken"] = ss.loadImage(filepath.Join(basePath, "fenceBroken.png"))

	// Шипы
	ss.tileSprites["spikes"] = ss.loadImage(filepath.Join(basePath, "spikes.png"))

	// Мосты
	ss.tileSprites["bridge"] = ss.loadImage(filepath.Join(basePath, "bridge.png"))
	ss.tileSprites["bridgeLogs"] = ss.loadImage(filepath.Join(basePath, "bridgeLogs.png"))

	// Вода и лава
	ss.tileSprites["liquidWater"] = ss.loadImage(filepath.Join(basePath, "liquidWater.png"))
	ss.tileSprites["liquidWaterTop"] = ss.loadImage(filepath.Join(basePath, "liquidWaterTop.png"))
	ss.tileSprites["liquidLava"] = ss.loadImage(filepath.Join(basePath, "liquidLava.png"))
	ss.tileSprites["liquidLavaTop"] = ss.loadImage(filepath.Join(basePath, "liquidLavaTop.png"))

	// Камни
	ss.tileSprites["rock"] = ss.loadImage(filepath.Join(basePath, "rock.png"))

	// Кусты и растения
	ss.tileSprites["bush"] = ss.loadImage(filepath.Join(basePath, "bush.png"))
	ss.tileSprites["plant"] = ss.loadImage(filepath.Join(basePath, "plant.png"))
	ss.tileSprites["plantPurple"] = ss.loadImage(filepath.Join(basePath, "plantPurple.png"))
	ss.tileSprites["cactus"] = ss.loadImage(filepath.Join(basePath, "cactus.png"))

	// Грибы (из Items)
	ss.tileSprites["mushroomRed"] = ss.loadImage(filepath.Join("assets/Base pack/Items", "mushroomRed.png"))
	ss.tileSprites["mushroomBrown"] = ss.loadImage(filepath.Join("assets/Base pack/Items", "mushroomBrown.png"))

	// Холмы большие и маленькие
	ss.tileSprites["hill_large"] = ss.loadImage(filepath.Join(basePath, "hill_large.png"))
	ss.tileSprites["hill_small"] = ss.loadImage(filepath.Join(basePath, "hill_small.png"))

	// Лёд (из Ice expansion)
	icePath := "assets/Ice expansion/Tiles"
	ss.tileSprites["ice"] = ss.loadImage(filepath.Join(icePath, "ice.png"))
	ss.tileSprites["iceCenter"] = ss.loadImage(filepath.Join(icePath, "iceCenter.png"))
	ss.tileSprites["iceHalf"] = ss.loadImage(filepath.Join(icePath, "iceHalf.png"))

	// Конфеты (из Candy expansion)
	candyPath := "assets/Candy expansion/Tiles"
	ss.tileSprites["candy"] = ss.loadImage(filepath.Join(candyPath, "candy.png"))
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

// GetPlayerSprite возвращает спрайт игрока
func (ss *SpriteSheet) GetPlayerSprite(name string) *ebiten.Image {
	return ss.playerSprites[name]
}

// GetPlayerAnim возвращает анимацию игрока
func (ss *SpriteSheet) GetPlayerAnim(name string) *Animation {
	return ss.playerAnims[name]
}

// GetEnemySprite возвращает спрайт врага/друга
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

// GetBackground возвращает основной фон
func (ss *SpriteSheet) GetBackground() *ebiten.Image {
	return ss.background
}

// GetBackgroundByTheme возвращает фон по теме
func (ss *SpriteSheet) GetBackgroundByTheme(themeName string) *ebiten.Image {
	if bg, ok := ss.backgrounds[themeName]; ok {
		return bg
	}
	return ss.background
}

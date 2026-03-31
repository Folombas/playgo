// Package sprite - загрузка ВСЕХ спрайтов из PlatformerComplete
// Go365 Day 91 - Sunny Adventure
package sprite

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// Animation - анимация спрайта
type Animation struct {
	Frames    []*ebiten.Image
	FrameTime float64
	Loop      bool
	Name      string
}

// SpriteSheet - коллекция ВСЕХ спрайтов
type SpriteSheet struct {
	playerSprites  map[string]*ebiten.Image
	playerAnims    map[string]*Animation
	enemySprites   map[string]*ebiten.Image
	enemyAnims     map[string]*Animation
	itemSprites    map[string]*ebiten.Image
	backgrounds    map[string]*ebiten.Image
	tileSprites    map[string]*ebiten.Image
	hudSprites     map[string]*ebiten.Image
}

// LoadSpriteSheet загружает ВСЕ спрайты
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		playerSprites: make(map[string]*ebiten.Image),
		playerAnims:   make(map[string]*Animation),
		enemySprites:  make(map[string]*ebiten.Image),
		enemyAnims:    make(map[string]*Animation),
		itemSprites:   make(map[string]*ebiten.Image),
		backgrounds:   make(map[string]*ebiten.Image),
		tileSprites:   make(map[string]*ebiten.Image),
		hudSprites:    make(map[string]*ebiten.Image),
	}

	ss.loadPlayerSprites()
	ss.loadEnemySprites()
	ss.loadItemSprites()
	ss.loadBackgrounds()
	ss.loadTileSprites()
	ss.loadHUDSprites()

	return ss, nil
}

// loadPlayerSprites - ВСЕ спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites() {
	basePath := "assets/Base pack/Player"

	// Статичные спрайты p1, p2, p3
	for _, p := range []string{"p1", "p2", "p3"} {
		ss.playerSprites[p+"_stand"] = ss.loadImage(filepath.Join(basePath, p+"_stand.png"))
		ss.playerSprites[p+"_jump"] = ss.loadImage(filepath.Join(basePath, p+"_jump.png"))
		ss.playerSprites[p+"_duck"] = ss.loadImage(filepath.Join(basePath, p+"_duck.png"))
		ss.playerSprites[p+"_hurt"] = ss.loadImage(filepath.Join(basePath, p+"_hurt.png"))
		ss.playerSprites[p+"_front"] = ss.loadImage(filepath.Join(basePath, p+"_front.png"))

		// Анимации ходьбы
		walkPath := filepath.Join(basePath, p+"_walk", "PNG")
		walkFrames := make([]*ebiten.Image, 0)
		for i := 1; i <= 11; i++ {
			frame := ss.loadImage(filepath.Join(walkPath, fmt.Sprintf("%s_walk%02d.png", p, i)))
			if frame != nil {
				walkFrames = append(walkFrames, frame)
			}
		}
		if len(walkFrames) > 0 {
			ss.playerAnims[p+"_walk"] = &Animation{
				Frames:    walkFrames,
				FrameTime: 0.1,
				Loop:      true,
				Name:      p + "_walk",
			}
		}
	}

	ss.playerAnims["walk"] = ss.playerAnims["p1_walk"]
	ss.playerAnims["run"] = ss.playerAnims["p1_walk"]
}

// loadEnemySprites - ВСЕ спрайты врагов
func (ss *SpriteSheet) loadEnemySprites() {
	basePath := "assets/Base pack/Enemies"
	extraPath := "assets/Extra animations and enemies/Enemy sprites"

	// Base pack враги
	enemies := []string{
		"blockerBody", "blockerMad", "blockerSad",
		"fishDead", "fishSwim1", "fishSwim2",
		"flyDead", "flyFly1", "flyFly2",
		"pokerMad", "pokerSad",
		"slimeDead", "slimeWalk1", "slimeWalk2",
		"snailShell", "snailShell_upsidedown", "snailWalk1", "snailWalk2",
	}
	for _, e := range enemies {
		ss.enemySprites[e] = ss.loadImage(filepath.Join(basePath, e+".png"))
	}

	// Анимации
	ss.enemyAnims["slimeWalk"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["slimeWalk1"], ss.enemySprites["slimeWalk2"]},
		FrameTime: 0.15,
		Loop:      true,
	}
	ss.enemyAnims["fishSwim"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["fishSwim1"], ss.enemySprites["fishSwim2"]},
		FrameTime: 0.15,
		Loop:      true,
	}
	ss.enemyAnims["flyFly"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["flyFly1"], ss.enemySprites["flyFly2"]},
		FrameTime: 0.1,
		Loop:      true,
	}
	ss.enemyAnims["snailWalk"] = &Animation{
		Frames:    []*ebiten.Image{ss.enemySprites["snailWalk1"], ss.enemySprites["snailWalk2"]},
		FrameTime: 0.2,
		Loop:      true,
	}

	// Extra враги
	extraEnemies := []string{
		"bee", "bee_fly", "bee_dead", "bee_hit",
		"ladyBug", "ladyBug_walk", "ladyBug_hit", "ladyBug_fly",
		"frog", "frog_leap", "frog_dead", "frog_hit",
		"ghost", "ghost_normal", "ghost_dead", "ghost_hit",
		"bat", "bat_fly", "bat_dead", "bat_hit", "bat_hang",
		"spider", "spider_walk1", "spider_walk2", "spider_dead", "spider_hit",
		"snake", "snake_walk", "snake_dead", "snake_hit",
		"worm", "worm_walk", "worm_dead", "worm_hit",
		"fishGreen", "fishGreen_swim", "fishGreen_dead", "fishGreen_hit",
		"fishPink", "fishPink_swim", "fishPink_dead", "fishPink_hit",
		"slime", "slime_walk", "slime_dead", "slime_hit", "slime_squashed",
		"slimeBlue", "slimeBlue_blue", "slimeBlue_dead", "slimeBlue_hit", "slimeBlue_squashed",
		"slimeGreen", "slimeGreen_walk", "slimeGreen_dead", "slimeGreen_hit", "slimeGreen_squashed",
		"snakeLava", "snakeLava_ani", "snakeLava_dead", "snakeLava_hit",
		"snakeSlime", "snakeSlime_ani", "snakeSlime_dead", "snakeSlime_hit",
		"spinner", "spinner_spin", "spinner_dead", "spinner_hit",
		"spinnerHalf", "spinnerHalf_spin", "spinnerHalf_dead", "spinnerHalf_hit",
		"mouse", "mouse_walk", "mouse_dead", "mouse_hit",
		"piranha", "piranha_down", "piranha_dead", "piranha_hit",
		"barnacle", "barnacle_bite", "barnacle_dead", "barnacle_hit",
		"grassBlock", "grassBlock_jump", "grassBlock_dead", "grassBlock_hit",
	}
	for _, e := range extraEnemies {
		ss.enemySprites[e] = ss.loadImage(filepath.Join(extraPath, e+".png"))
	}
}

// loadItemSprites - ВСЕ спрайты предметов
func (ss *SpriteSheet) loadItemSprites() {
	basePath := "assets/Base pack/Items"

	// Все предметы
	items := []string{
		"coinBronze", "coinGold", "coinSilver",
		"gemBlue", "gemGreen", "gemRed", "gemYellow",
		"star", "bomb", "bombFlash",
		"mushroomRed", "mushroomBrown",
		"keyBlue", "keyGreen", "keyRed", "keyYellow",
		"flagBlue", "flagBlue2", "flagBlueHanging",
		"flagGreen", "flagGreen2", "flagGreenHanging",
		"flagRed", "flagRed2", "flagRedHanging",
		"flagYellow", "flagYellow2", "flagYellowHanging",
		"cloud1", "cloud2", "cloud3",
		"bush", "plant", "plantPurple", "cactus", "rock",
		"spikes", "springboardUp", "springboardDown",
		"buttonBlue", "buttonBlue_pressed",
		"buttonGreen", "buttonGreen_pressed",
		"buttonRed", "buttonRed_pressed",
		"buttonYellow", "buttonYellow_pressed",
		"switchLeft", "switchMid", "switchRight",
		"weight", "weightChained", "chain",
		"fireball", "particleBrick1a", "particleBrick1b",
		"particleBrick2a", "particleBrick2b",
		"snowhill",
	}
	for _, item := range items {
		ss.itemSprites[item] = ss.loadImage(filepath.Join(basePath, item+".png"))
	}
}

// loadBackgrounds - ВСЕ фоны
func (ss *SpriteSheet) loadBackgrounds() {
	// Base фоны
	ss.backgrounds["bg"] = ss.loadImage("assets/Base pack/bg.png")
	ss.backgrounds["bg_castle"] = ss.loadImage("assets/Base pack/bg_castle.png")

	// Mushroom expansion фоны
	mushroomBg := "assets/Mushroom expansion/Backgrounds"
	ss.backgrounds["bg_grasslands"] = ss.loadImage(filepath.Join(mushroomBg, "bg_grasslands.png"))
	ss.backgrounds["bg_castle"] = ss.loadImage(filepath.Join(mushroomBg, "bg_castle.png"))
	ss.backgrounds["bg_shroom"] = ss.loadImage(filepath.Join(mushroomBg, "bg_shroom.png"))
	ss.backgrounds["bg_desert"] = ss.loadImage(filepath.Join(mushroomBg, "bg_desert.png"))
}

// loadTileSprites - ВСЕ тайлы
func (ss *SpriteSheet) loadTileSprites() {
	basePath := "assets/Base pack/Tiles"

	// Загружаем ВСЕ файлы PNG из папки Tiles
	tileFiles, _ := filepath.Glob(filepath.Join(basePath, "*.png"))
	for _, file := range tileFiles {
		name := strings.TrimSuffix(filepath.Base(file), ".png")
		ss.tileSprites[name] = ss.loadImage(file)
	}

	// Ice expansion
	icePath := "assets/Ice expansion/Tiles"
	iceFiles, _ := filepath.Glob(filepath.Join(icePath, "*.png"))
	for _, file := range iceFiles {
		name := "ice_" + strings.TrimSuffix(filepath.Base(file), ".png")
		ss.tileSprites[name] = ss.loadImage(file)
	}

	// Candy expansion
	candyPath := "assets/Candy expansion/Tiles"
	candyFiles, _ := filepath.Glob(filepath.Join(candyPath, "*.png"))
	for _, file := range candyFiles {
		name := "candy_" + strings.TrimSuffix(filepath.Base(file), ".png")
		ss.tileSprites[name] = ss.loadImage(file)
	}

	// Buildings expansion
	buildPath := "assets/Buildings expansion/Tiles"
	buildFiles, _ := filepath.Glob(filepath.Join(buildPath, "*.png"))
	for _, file := range buildFiles {
		name := "build_" + strings.TrimSuffix(filepath.Base(file), ".png")
		ss.tileSprites[name] = ss.loadImage(file)
	}
}

// loadHUDSprites - ВСЕ спрайты HUD
func (ss *SpriteSheet) loadHUDSprites() {
	hudPath := "assets/Base pack/HUD"

	// Все HUD элементы
	hudFiles := []string{
		"hud_0", "hud_1", "hud_2", "hud_3", "hud_4",
		"hud_5", "hud_6", "hud_7", "hud_8", "hud_9",
		"hud_x", "hud_coins",
		"hud_heartEmpty", "hud_heartFull", "hud_heartHalf",
		"hud_gem_blue", "hud_gem_green", "hud_gem_red", "hud_gem_yellow",
		"hud_keyBlue", "hud_keyBlue_disabled",
		"hud_keyGreen", "hud_keyGreem_disabled",
		"hud_keyRed", "hud_keyRed_disabled",
		"hud_keyYellow", "hud_keyYellow_disabled",
		"hud_p1", "hud_p1Alt",
		"hud_p2", "hud_p2Alt",
		"hud_p3", "hud_p3Alt",
	}
	for _, h := range hudFiles {
		ss.hudSprites[h] = ss.loadImage(filepath.Join(hudPath, h+".png"))
	}
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

// GetPlayerSprite - спрайт игрока
func (ss *SpriteSheet) GetPlayerSprite(name string) *ebiten.Image {
	return ss.playerSprites[name]
}

// GetPlayerAnim - анимация игрока
func (ss *SpriteSheet) GetPlayerAnim(name string) *Animation {
	return ss.playerAnims[name]
}

// GetEnemySprite - спрайт врага
func (ss *SpriteSheet) GetEnemySprite(name string) *ebiten.Image {
	return ss.enemySprites[name]
}

// GetEnemyAnim - анимация врага
func (ss *SpriteSheet) GetEnemyAnim(name string) *Animation {
	return ss.enemyAnims[name]
}

// GetItemSprite - спрайт предмета
func (ss *SpriteSheet) GetItemSprite(name string) *ebiten.Image {
	return ss.itemSprites[name]
}

// GetTileSprite - спрайт тайла
func (ss *SpriteSheet) GetTileSprite(name string) *ebiten.Image {
	return ss.tileSprites[name]
}

// GetBackground - фон
func (ss *SpriteSheet) GetBackground() *ebiten.Image {
	return ss.backgrounds["bg"]
}

// GetBackgroundByTheme - фон по теме
func (ss *SpriteSheet) GetBackgroundByTheme(themeName string) *ebiten.Image {
	if bg, ok := ss.backgrounds[themeName]; ok {
		return bg
	}
	return ss.backgrounds["bg"]
}

// GetHUDSprite - спрайт HUD
func (ss *SpriteSheet) GetHUDSprite(name string) *ebiten.Image {
	return ss.hudSprites[name]
}

// GetAllTileSprites - ВСЕ спрайты тайлов
func (ss *SpriteSheet) GetAllTileSprites() map[string]*ebiten.Image {
	return ss.tileSprites
}

// GetAllEnemySprites - ВСЕ спрайты врагов
func (ss *SpriteSheet) GetAllEnemySprites() map[string]*ebiten.Image {
	return ss.enemySprites
}

// GetAllItemSprites - ВСЕ спрайты предметов
func (ss *SpriteSheet) GetAllItemSprites() map[string]*ebiten.Image {
	return ss.itemSprites
}

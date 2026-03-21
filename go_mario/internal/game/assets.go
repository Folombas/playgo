package game

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// AssetManager - менеджер всех игровых ассетов
type AssetManager struct {
	tiles       map[string]*ebiten.Image
	enemies     map[string][]*ebiten.Image
	players     map[string][]*ebiten.Image
	items       map[string]*ebiten.Image
	backgrounds map[string]*ebiten.Image
	loaded      bool
}

// NewAssetManager создаёт менеджер ассетов
func NewAssetManager() *AssetManager {
	return &AssetManager{
		tiles:       make(map[string]*ebiten.Image),
		enemies:     make(map[string][]*ebiten.Image),
		players:     make(map[string][]*ebiten.Image),
		items:       make(map[string]*ebiten.Image),
		backgrounds: make(map[string]*ebiten.Image),
		loaded:      false,
	}
}

// LoadAll загружает ВСЕ ассеты из папки assets
func (am *AssetManager) LoadAll() bool {
	log.Println("🎨 Loading all game assets...")

	// Load tiles
	am.loadTiles("assets/PNG/Tiles")

	// Load enemies
	am.loadEnemies("assets/PNG/Enemies")

	// Load player sprites (all colors)
	am.loadPlayers("assets/PNG/Players/128x256")

	// Load backgrounds
	am.loadBackgrounds("assets/PNG/Backgrounds")

	// Load items
	am.loadItems("assets/PNG/Tiles")

	am.loaded = true
	log.Printf("✅ Assets loaded! Tiles: %d, Enemies: %d, Players: %d",
		len(am.tiles), len(am.enemies), len(am.players))

	return true
}

// loadTiles загружает все тайлы
func (am *AssetManager) loadTiles(dir string) {
	tileFiles := []string{
		"grass.png", "dirt.png", "rock.png", "snow.png",
		"brickBrown.png", "brickGrey.png", "boxCrate.png", "boxCrate_double.png",
		"boxCoin.png", "boxItem.png", "boxExplosive.png",
		"bomb.png", "bombWhite.png",
		"lava.png", "lavaTop_high.png", "lavaTop_low.png",
		"water.png", "waterTop_high.png", "waterTop_low.png",
		"spikes.png", "spring.png", "sprung.png",
		"ladderMid.png", "ladderTop.png",
		"doorClosed_mid.png", "doorClosed_top.png",
		"doorOpen_mid.png", "doorOpen_top.png",
		"fence.png", "fenceBroken.png",
		"bush.png", "cactus.png", "mushroomBrown.png", "mushroomRed.png",
		"plantPurple.png", "chain.png",
		"bridgeA.png", "bridgeB.png",
		"switchBlue.png", "switchGreen.png", "switchRed.png", "switchYellow.png",
		"lockBlue.png", "lockGreen.png", "lockRed.png", "lockYellow.png",
		"leverLeft.png", "leverMid.png", "leverRight.png",
		"torch1.png", "torch2.png", "torchOff.png",
		"sign.png", "signExit.png", "signLeft.png", "signRight.png",
		"window.png", "weight.png", "weightAttached.png",
	}

	for _, file := range tileFiles {
		path := filepath.Join(dir, file)
		if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
			name := file[:len(file)-4] // Remove .png
			am.tiles[name] = img
		}
	}
}

// loadEnemies загружает все спрайты врагов
func (am *AssetManager) loadEnemies(dir string) {
	enemyTypes := []string{
		"slimeGreen", "slimeBlue", "slimePurple", "slimeBlock",
		"fishBlue", "fishGreen", "fishPink",
		"wormGreen", "wormPink",
		"snail", "mouse", "frog", "fly", "ladybug",
		"saw", "sawHalf",
		"barnacle", "bee",
	}

	for _, enemy := range enemyTypes {
		frames := am.loadEnemyFrames(dir, enemy)
		if len(frames) > 0 {
			am.enemies[enemy] = frames
		}
	}
}

// loadEnemyFrames загружает кадры анимации врага
func (am *AssetManager) loadEnemyFrames(dir, name string) []*ebiten.Image {
	frames := make([]*ebiten.Image, 0)

	// Try different frame suffixes
	suffixes := []string{"_move", "", "_dead", "_hit", "_fly", "_fall", "_attack"}

	for _, suffix := range suffixes {
		path := filepath.Join(dir, name+suffix+".png")
		if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
			frames = append(frames, img)
		}
	}

	// Also try numbered variants
	for i := 1; i <= 2; i++ {
		path := filepath.Join(dir, name+"_move"+string(rune('0'+i))+".png")
		if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
			frames = append(frames, img)
		}
	}

	return frames
}

// loadPlayers загружает спрайты игроков (все цвета)
func (am *AssetManager) loadPlayers(baseDir string) {
	colors := []string{"Blue", "Green", "Pink", "Yellow"}

	for _, color := range colors {
		dir := filepath.Join(baseDir, color)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		frames := make([]*ebiten.Image, 11)

		// Load specific frames
		frameFiles := map[string]int{
			"alien" + color + "_stand.png":  0,
			"alien" + color + "_walk1.png":  1,
			"alien" + color + "_walk2.png":  2,
			"alien" + color + "_jump.png":   3,
			"alien" + color + "_duck.png":   4,
			"alien" + color + "_hit.png":    5,
			"alien" + color + "_climb1.png": 6,
			"alien" + color + "_climb2.png": 7,
			"alien" + color + "_swim1.png":  8,
			"alien" + color + "_swim2.png":  9,
			"alien" + color + "_front.png":  10,
		}

		for file, idx := range frameFiles {
			path := filepath.Join(dir, file)
			if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
				frames[idx] = img
			}
		}

		am.players[color] = frames
	}
}

// loadBackgrounds загружает фоны
func (am *AssetManager) loadBackgrounds(dir string) {
	bgFiles := []string{
		"blue_grass.png", "blue_desert.png", "blue_land.png", "blue_shroom.png",
		"colored_grass.png", "colored_desert.png", "colored_land.png", "colored_shroom.png",
	}

	for _, file := range bgFiles {
		path := filepath.Join(dir, file)
		if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
			name := file[:len(file)-4]
			am.backgrounds[name] = img
		}
	}
}

// loadItems загружает предметы из тайлов
func (am *AssetManager) loadItems(dir string) {
	itemFiles := []string{
		"boxCoin.png", "boxItem.png", "boxCrate.png",
		"mushroomBrown.png", "mushroomRed.png",
		"spring.png", "bomb.png",
	}

	for _, file := range itemFiles {
		path := filepath.Join(dir, file)
		if img, _, err := ebitenutil.NewImageFromFile(path); err == nil {
			name := file[:len(file)-4]
			am.items[name] = img
		}
	}
}

// GetTile возвращает тайл по имени
func (am *AssetManager) GetTile(name string) *ebiten.Image {
	return am.tiles[name]
}

// GetEnemy возвращает кадры врага
func (am *AssetManager) GetEnemy(name string) []*ebiten.Image {
	return am.enemies[name]
}

// GetPlayer возвращает кадры игрока по цвету
func (am *AssetManager) GetPlayer(color string) []*ebiten.Image {
	if frames, ok := am.players[color]; ok {
		return frames
	}
	return am.players["Blue"] // Default
}

// GetItem возвращает предмет
func (am *AssetManager) GetItem(name string) *ebiten.Image {
	return am.items[name]
}

// GetBackground возвращает фон
func (am *AssetManager) GetBackground(name string) *ebiten.Image {
	return am.backgrounds[name]
}

// IsLoaded проверяет, загружены ли ассеты
func (am *AssetManager) IsLoaded() bool {
	return am.loaded
}

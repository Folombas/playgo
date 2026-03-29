// Package sprite - загрузка и управление спрайтами
// Go365 Day 88 - PlatformerComplete Pack
package sprite

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// SpriteSheet - атлас спрайтов PlatformerComplete
type SpriteSheet struct {
	// Игрок (Player 1)
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerDuck   *ebiten.Image
	PlayerHurt   *ebiten.Image
	PlayerFront  *ebiten.Image
	
	// Враги
	EnemySlime   []*ebiten.Image
	EnemyFly     []*ebiten.Image
	EnemySnail   []*ebiten.Image
	EnemyFish    []*ebiten.Image
	
	// Платформы и земля
	Tiles        map[string]*ebiten.Image
	
	// Предметы
	CoinGold     *ebiten.Image
	CoinSilver   *ebiten.Image
	CoinBronze   *ebiten.Image
	GemRed       *ebiten.Image
	GemBlue      *ebiten.Image
	GemGreen     *ebiten.Image
	GemYellow    *ebiten.Image
	Star         *ebiten.Image
	MushroomRed  *ebiten.Image
	MushroomBrown *ebiten.Image
	
	// Декорации
	Cloud1       *ebiten.Image
	Cloud2       *ebiten.Image
	Cloud3       *ebiten.Image
	Bush         *ebiten.Image
	Plant        *ebiten.Image
	PlantPurple  *ebiten.Image
	Rock         *ebiten.Image
	Cactus       *ebiten.Image
}

// LoadSpriteSheet - загрузка спрайт-листа PlatformerComplete
func LoadSpriteSheet() *SpriteSheet {
	ss := &SpriteSheet{
		PlayerWalk: make([]*ebiten.Image, 11),
		EnemySlime: make([]*ebiten.Image, 2),
		EnemyFly:   make([]*ebiten.Image, 2),
		EnemySnail: make([]*ebiten.Image, 2),
		EnemyFish:  make([]*ebiten.Image, 2),
		Tiles:      make(map[string]*ebiten.Image),
	}
	
	// Загрузка игрока
	ss.loadPlayer()
	
	// Загрузка врагов
	ss.loadEnemies()
	
	// Загрузка тайлов
	ss.loadTiles()
	
	// Загрузка предметов
	ss.loadItems()
	
	// Загрузка декораций
	ss.loadDecorations()
	
	return ss
}

// loadPlayer - загрузка спрайтов игрока
func (ss *SpriteSheet) loadPlayer() {
	// Стойка
	ss.PlayerStand = ss.loadImage("Player/p1_stand.png")
	
	// Ходьба (11 кадров)
	for i := 1; i <= 11; i++ {
		filename := fmt.Sprintf("Player/p1_walk/PNG/p1_walk%02d.png", i)
		ss.PlayerWalk[i-1] = ss.loadImage(filename)
	}
	
	// Прыжок
	ss.PlayerJump = ss.loadImage("Player/p1_jump.png")
	
	// Присед
	ss.PlayerDuck = ss.loadImage("Player/p1_duck.png")
	
	// Получение урона
	ss.PlayerHurt = ss.loadImage("Player/p1_hurt.png")
	
	// Вид спереди
	ss.PlayerFront = ss.loadImage("Player/p1_front.png")
}

// loadEnemies - загрузка спрайтов врагов
func (ss *SpriteSheet) loadEnemies() {
	// Слайм (2 кадра)
	ss.EnemySlime[0] = ss.loadImage("Enemies/slimeWalk1.png")
	ss.EnemySlime[1] = ss.loadImage("Enemies/slimeWalk2.png")
	
	// Муха
	ss.EnemyFly[0] = ss.loadImage("Enemies/flyFly1.png")
	ss.EnemyFly[1] = ss.loadImage("Enemies/flyFly2.png")
	
	// Улитка
	ss.EnemySnail[0] = ss.loadImage("Enemies/snailWalk1.png")
	ss.EnemySnail[1] = ss.loadImage("Enemies/snailWalk2.png")
	
	// Рыба
	ss.EnemyFish[0] = ss.loadImage("Enemies/fishSwim1.png")
	ss.EnemyFish[1] = ss.loadImage("Enemies/fishSwim2.png")
}

// loadTiles - загрузка тайлов
func (ss *SpriteSheet) loadTiles() {
	// Земля с травой
	ss.Tiles["grassMid"] = ss.loadImage("Tiles/grassMid.png")
	ss.Tiles["grassLeft"] = ss.loadImage("Tiles/grassLeft.png")
	ss.Tiles["grassRight"] = ss.loadImage("Tiles/grassRight.png")
	ss.Tiles["grassHalf"] = ss.loadImage("Tiles/grassHalf.png")
	ss.Tiles["grassHalfLeft"] = ss.loadImage("Tiles/grassHalfLeft.png")
	ss.Tiles["grassHalfRight"] = ss.loadImage("Tiles/grassHalfRight.png")
	ss.Tiles["grassCenter"] = ss.loadImage("Tiles/grassCenter.png")
	
	// Dirt
	ss.Tiles["dirtMid"] = ss.loadImage("Tiles/dirtMid.png")
	ss.Tiles["dirtLeft"] = ss.loadImage("Tiles/dirtLeft.png")
	ss.Tiles["dirtRight"] = ss.loadImage("Tiles/dirtRight.png")
	ss.Tiles["dirtCenter"] = ss.loadImage("Tiles/dirtCenter.png")
	
	// Камень
	ss.Tiles["stoneMid"] = ss.loadImage("Tiles/stoneMid.png")
	ss.Tiles["stoneLeft"] = ss.loadImage("Tiles/stoneLeft.png")
	ss.Tiles["stoneRight"] = ss.loadImage("Tiles/stoneRight.png")
	ss.Tiles["stoneCenter"] = ss.loadImage("Tiles/stoneCenter.png")
	
	// Замок
	ss.Tiles["castleMid"] = ss.loadImage("Tiles/castleMid.png")
	ss.Tiles["castleLeft"] = ss.loadImage("Tiles/castleLeft.png")
	ss.Tiles["castleRight"] = ss.loadImage("Tiles/castleRight.png")
	ss.Tiles["castleCenter"] = ss.loadImage("Tiles/castleCenter.png")
	
	// Стены
	ss.Tiles["brickWall"] = ss.loadImage("Tiles/brickWall.png")
	ss.Tiles["stoneWall"] = ss.loadImage("Tiles/stoneWall.png")
	
	// Холмы
	ss.Tiles["grassHillLeft"] = ss.loadImage("Tiles/grassHillLeft.png")
	ss.Tiles["grassHillRight"] = ss.loadImage("Tiles/grassHillRight.png")
	ss.Tiles["grassHillLeft2"] = ss.loadImage("Tiles/grassHillLeft2.png")
	ss.Tiles["grassHillRight2"] = ss.loadImage("Tiles/grassHillRight2.png")
	
	// Лестницы
	ss.Tiles["ladder_mid"] = ss.loadImage("Tiles/ladder_mid.png")
	ss.Tiles["ladder_top"] = ss.loadImage("Tiles/ladder_top.png")
	
	// Вода/лава
	ss.Tiles["liquidWaterTop"] = ss.loadImage("Tiles/liquidWaterTop.png")
	ss.Tiles["liquidWater"] = ss.loadImage("Tiles/liquidWater.png")
	ss.Tiles["liquidLavaTop"] = ss.loadImage("Tiles/liquidLavaTop.png")
	ss.Tiles["liquidLava"] = ss.loadImage("Tiles/liquidLava.png")
	
	// Ящики
	ss.Tiles["box"] = ss.loadImage("Tiles/box.png")
	ss.Tiles["boxCoin"] = ss.loadImage("Tiles/boxCoin.png")
	ss.Tiles["boxItem"] = ss.loadImage("Tiles/boxItem.png")
	ss.Tiles["boxExplosive"] = ss.loadImage("Tiles/boxExplosive.png")
	
	// Забор
	ss.Tiles["fence"] = ss.loadImage("Tiles/fence.png")
	
	// Шипы
	ss.Tiles["spikes"] = ss.loadImage("Items/spikes.png")
}

// loadItems - загрузка предметов
func (ss *SpriteSheet) loadItems() {
	ss.CoinGold = ss.loadImage("Items/coinGold.png")
	ss.CoinSilver = ss.loadImage("Items/coinSilver.png")
	ss.CoinBronze = ss.loadImage("Items/coinBronze.png")
	
	ss.GemRed = ss.loadImage("Items/gemRed.png")
	ss.GemBlue = ss.loadImage("Items/gemBlue.png")
	ss.GemGreen = ss.loadImage("Items/gemGreen.png")
	ss.GemYellow = ss.loadImage("Items/gemYellow.png")
	
	ss.Star = ss.loadImage("Items/star.png")
	
	ss.MushroomRed = ss.loadImage("Items/mushroomRed.png")
	ss.MushroomBrown = ss.loadImage("Items/mushroomBrown.png")
}

// loadDecorations - загрузка декораций
func (ss *SpriteSheet) loadDecorations() {
	ss.Cloud1 = ss.loadImage("Items/cloud1.png")
	ss.Cloud2 = ss.loadImage("Items/cloud2.png")
	ss.Cloud3 = ss.loadImage("Items/cloud3.png")
	
	ss.Bush = ss.loadImage("Items/bush.png")
	ss.Plant = ss.loadImage("Items/plant.png")
	ss.PlantPurple = ss.loadImage("Items/plantPurple.png")
	ss.Rock = ss.loadImage("Items/rock.png")
	ss.Cactus = ss.loadImage("Items/cactus.png")
}

// loadImage - загрузка изображения
func (ss *SpriteSheet) loadImage(path string) *ebiten.Image {
	fullPath := "assets/" + strings.ReplaceAll(path, "\\", "/")
	img, _, err := ebitenutil.NewImageFromFile(fullPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load %s: %v\n", fullPath, err)
		return nil
	}
	return img
}

// LoadImageFromFile - публичная функция загрузки
func LoadImageFromFile(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	return img, err
}

// GetPlayerFrame - получение кадра анимации игрока
func (ss *SpriteSheet) GetPlayerFrame(state string, frame int) *ebiten.Image {
	switch state {
	case "stand":
		return ss.PlayerStand
	case "walk":
		if frame >= 0 && frame < len(ss.PlayerWalk) {
			return ss.PlayerWalk[frame]
		}
	case "jump":
		return ss.PlayerJump
	case "duck":
		return ss.PlayerDuck
	case "hurt":
		return ss.PlayerHurt
	case "front":
		return ss.PlayerFront
	}
	return ss.PlayerStand
}

// GetEnemyFrame - получение кадра анимации врага
func (ss *SpriteSheet) GetEnemyFrame(enemyType string, frame int) *ebiten.Image {
	switch enemyType {
	case "slime":
		if frame >= 0 && frame < len(ss.EnemySlime) {
			return ss.EnemySlime[frame%len(ss.EnemySlime)]
		}
	case "fly":
		if frame >= 0 && frame < len(ss.EnemyFly) {
			return ss.EnemyFly[frame%len(ss.EnemyFly)]
		}
	case "snail":
		if frame >= 0 && frame < len(ss.EnemySnail) {
			return ss.EnemySnail[frame%len(ss.EnemySnail)]
		}
	case "fish":
		if frame >= 0 && frame < len(ss.EnemyFish) {
			return ss.EnemyFish[frame%len(ss.EnemyFish)]
		}
	}
	if len(ss.EnemySlime) > 0 {
		return ss.EnemySlime[0]
	}
	return nil
}

// GetCoin - получение монеты
func (ss *SpriteSheet) GetCoin(coinType string) *ebiten.Image {
	switch coinType {
	case "gold":
		return ss.CoinGold
	case "silver":
		return ss.CoinSilver
	case "bronze":
		return ss.CoinBronze
	default:
		return ss.CoinGold
	}
}

// GetGem - получение гема
func (ss *SpriteSheet) GetGem(gemType string) *ebiten.Image {
	switch gemType {
	case "red":
		return ss.GemRed
	case "blue":
		return ss.GemBlue
	case "green":
		return ss.GemGreen
	case "yellow":
		return ss.GemYellow
	default:
		return ss.GemRed
	}
}

// GetTile - получение тайла
func (ss *SpriteSheet) GetTile(tileType string) *ebiten.Image {
	if img, ok := ss.Tiles[tileType]; ok {
		return img
	}
	return ss.Tiles["grassMid"]
}

// GetDecoration - получение декорации
func (ss *SpriteSheet) GetDecoration(decType string) *ebiten.Image {
	switch decType {
	case "cloud1":
		return ss.Cloud1
	case "cloud2":
		return ss.Cloud2
	case "cloud3":
		return ss.Cloud3
	case "bush":
		return ss.Bush
	case "plant":
		return ss.Plant
	case "plantPurple":
		return ss.PlantPurple
	case "rock":
		return ss.Rock
	case "cactus":
		return ss.Cactus
	default:
		return ss.Cloud1
	}
}

// GetTileSize - размер тайла
func (ss *SpriteSheet) GetTileSize() (int, int) {
	if tile := ss.Tiles["grassMid"]; tile != nil {
		bounds := tile.Bounds()
		return bounds.Dx(), bounds.Dy()
	}
	return 70, 70 // Размер по умолчанию
}

// GetPlayerSize - размер игрока
func (ss *SpriteSheet) GetPlayerSize() (int, int) {
	if ss.PlayerStand != nil {
		bounds := ss.PlayerStand.Bounds()
		return bounds.Dx(), bounds.Dy()
	}
	return 66, 92 // Размер по умолчанию из спрайт-листа
}

// GetEnemySize - размер врага
func (ss *SpriteSheet) GetEnemySize(enemyType string) (int, int) {
	switch enemyType {
	case "slime":
		if ss.EnemySlime[0] != nil {
			bounds := ss.EnemySlime[0].Bounds()
			return bounds.Dx(), bounds.Dy()
		}
	case "fly":
		if ss.EnemyFly[0] != nil {
			bounds := ss.EnemyFly[0].Bounds()
			return bounds.Dx(), bounds.Dy()
		}
	case "snail":
		if ss.EnemySnail[0] != nil {
			bounds := ss.EnemySnail[0].Bounds()
			return bounds.Dx(), bounds.Dy()
		}
	case "fish":
		if ss.EnemyFish[0] != nil {
			bounds := ss.EnemyFish[0].Bounds()
			return bounds.Dx(), bounds.Dy()
		}
	}
	return 70, 70
}

// CreateColorImage - создание цветного изображения для частиц
func CreateColorImage(width, height int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return ebiten.NewImageFromImage(img)
}

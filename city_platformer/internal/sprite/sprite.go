// Package sprite - загрузка и управление спрайтами для Village Platformer
// Go365 Day 93 - Деревенский платформер: Домики, деревья, холмы
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
	// Игрок - Солнышко
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerHappy  *ebiten.Image

	// Тайлы - природа
	Tiles map[string]*ebiten.Image

	// Тайлы - грибы
	Mushrooms map[string]*ebiten.Image

	// Тайлы - конфеты
	Candy map[string]*ebiten.Image

	// Предметы
	Items map[string]*ebiten.Image

	// Друзья - милые существа
	Friends map[string]*ebiten.Image

	// Враги
	Enemies map[string]*ebiten.Image

	// Фон
	Backgrounds map[string]*ebiten.Image
}

// LoadSpriteSheet загружает все спрайты из папки assets/sprites
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		Tiles:       make(map[string]*ebiten.Image),
		Mushrooms:   make(map[string]*ebiten.Image),
		Candy:       make(map[string]*ebiten.Image),
		Items:       make(map[string]*ebiten.Image),
		Friends:     make(map[string]*ebiten.Image),
		Enemies:     make(map[string]*ebiten.Image),
		Backgrounds: make(map[string]*ebiten.Image),
	}

	basePath := "assets/sprites"

	// Загрузка спрайтов игрока (Солнышко)
	if err := ss.loadPlayerSprites(basePath); err != nil {
		fmt.Println("Warning: player sprite error:", err)
	}

	// Загрузка тайлов
	if err := ss.loadTiles(basePath); err != nil {
		fmt.Println("Warning: tiles error:", err)
	}

	// Загрузка грибов
	if err := ss.loadMushrooms(basePath); err != nil {
		fmt.Println("Warning: mushrooms error:", err)
	}

	// Загрузка конфет
	if err := ss.loadCandy(basePath); err != nil {
		fmt.Println("Warning: candy error:", err)
	}

	// Загрузка предметов
	if err := ss.loadItems(basePath); err != nil {
		fmt.Println("Warning: items error:", err)
	}

	// Загрузка друзей
	if err := ss.loadFriends(basePath); err != nil {
		fmt.Println("Warning: friends error:", err)
	}

	// Загрузка врагов
	if err := ss.loadEnemies(basePath); err != nil {
		fmt.Println("Warning: enemies error:", err)
	}

	// Загрузка фонов
	if err := ss.loadBackgrounds(basePath); err != nil {
		fmt.Println("Warning: backgrounds error:", err)
	}

	return ss, nil
}

// loadPlayerSprites загружает спрайты игрока
func (ss *SpriteSheet) loadPlayerSprites(basePath string) error {
	playerPath := filepath.Join(basePath, "Player")

	// Загрузка отдельных спрайтов
	ss.PlayerStand = ss.loadImage(filepath.Join(playerPath, "p1_stand.png"))
	ss.PlayerJump = ss.loadImage(filepath.Join(playerPath, "p1_jump.png"))
	ss.PlayerHappy = ss.loadImage(filepath.Join(playerPath, "p1_front.png"))

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

	return nil
}

// loadTiles загружает тайлы (трава, земля, камень, кирпичи, замок)
func (ss *SpriteSheet) loadTiles(basePath string) error {
	tilesPath := filepath.Join(basePath, "Tiles")

	// Основные тайлы - природа
	tileFiles := []string{
		// Трава
		"grass.png", "grassMid.png", "grassLeft.png", "grassRight.png",
		"grassHalf.png", "grassHalfLeft.png", "grassHalfRight.png",
		"grassHillLeft.png", "grassHillLeft2.png", "grassHillRight.png", "grassHillRight2.png",
		// Земля
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
		// Декор
		"bush.png", "plant.png", "plantPurple.png", "flower.png",
		// Конфетные тайлы
		"cake.png", "cakeMid.png", "cakeLeft.png", "cakeRight.png",
		"choco.png", "chocoMid.png", "chocoLeft.png", "chocoRight.png",
		// Сладости
		"candyRed.png", "candyBlue.png", "candyGreen.png", "candyYellow.png",
		"canePink.png", "lollipopRed.png", "lollipopGreen.png",
		"cherry.png", "heart.png", "cupCake.png",
	}

	for _, file := range tileFiles {
		path := filepath.Join(tilesPath, file)
		if tile := ss.loadImage(path); tile != nil {
			name := file[:len(file)-4] // без .png
			ss.Tiles[name] = tile
		}
	}

	return nil
}

// loadMushrooms загружает грибы
func (ss *SpriteSheet) loadMushrooms(basePath string) error {
	itemsPath := filepath.Join(basePath, "Items")

	mushroomFiles := []string{
		// Грибы красные
		"shroomRedLeft.png", "shroomRedMid.png", "shroomRedRight.png",
		"shroomRedAltLeft.png", "shroomRedAltMid.png", "shroomRedAltRight.png",
		// Грибы коричневые
		"shroomBrownLeft.png", "shroomBrownMid.png", "shroomBrownRight.png",
		"shroomBrownAltLeft.png", "shroomBrownAltMid.png", "shroomBrownAltRight.png",
		// Грибы бежевые
		"shroomTanLeft.png", "shroomTanMid.png", "shroomTanRight.png",
		"shroomTanAltLeft.png", "shroomTanAltMid.png", "shroomTanAltRight.png",
		// Высокие грибы
		"tallShroom_red.png", "tallShroom_brown.png", "tallShroom_tan.png",
		// Маленькие грибы
		"tinyShroom_red.png", "tinyShroom_brown.png", "tinyShroom_tan.png",
		// Стебли
		"stem.png", "stemTop.png", "stemBase.png", "stemCrown.png",
	}

	for _, file := range mushroomFiles {
		path := filepath.Join(itemsPath, file)
		if mushroom := ss.loadImage(path); mushroom != nil {
			name := file[:len(file)-4]
			ss.Mushrooms[name] = mushroom
		}
	}

	return nil
}

// loadCandy загружает конфеты
func (ss *SpriteSheet) loadCandy(basePath string) error {
	tilesPath := filepath.Join(basePath, "Tiles")

	candyFiles := []string{
		"cookieBrown.png", "cookiePink.png", "cookieChoco.png",
		"creamPink.png", "creamVanilla.png", "creamChoco.png",
		"wafflePink.png", "waffleWhite.png",
		"gummyWormRedHead.png", "gummyWormRedMid.png", "gummyWormRedEnd.png",
		"gummyWormGreenHead.png", "gummyWormGreenMid.png", "gummyWormGreenEnd.png",
	}

	for _, file := range candyFiles {
		path := filepath.Join(tilesPath, file)
		if candy := ss.loadImage(path); candy != nil {
			name := file[:len(file)-4]
			ss.Candy[name] = candy
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
			name := file[:len(file)-4]
			ss.Items[name] = item
		}
	}

	return nil
}

// loadFriends загружает друзей (милые существа)
func (ss *SpriteSheet) loadFriends(basePath string) error {
	enemiesPath := filepath.Join(basePath, "Enemies")

	friendFiles := []string{
		// Пчёлка
		"bee_fly.png", "bee.png",
		// Божья коровка
		"ladyBug_walk.png", "ladyBug.png",
		// Лягушка
		"frog.png", "frog_leap.png",
		// Улитка
		"snailWalk1.png", "snailWalk2.png", "snail_walk.png",
		// Рыбки
		"fishSwim1.png", "fishSwim2.png",
	}

	for _, file := range friendFiles {
		path := filepath.Join(enemiesPath, file)
		if friend := ss.loadImage(path); friend != nil {
			name := file[:len(file)-4]
			ss.Friends[name] = friend
		}
	}

	return nil
}

// loadEnemies загружает врагов
func (ss *SpriteSheet) loadEnemies(basePath string) error {
	enemiesPath := filepath.Join(basePath, "Enemies")

	enemyFiles := []string{
		// Муха
		"flyFly1.png", "flyFly2.png", "fly.png",
		// Летучая мышь
		"bat_fly.png", "bat.png", "bat_hit.png", "bat_dead.png",
		// Слайм
		"slimeWalk1.png", "slimeWalk2.png", "slime.png",
		// Змея
		"snakeWalk.png", "snake.png",
		// Паук
		"spider_walk1.png", "spider_walk2.png", "spider.png",
		// Призрак
		"ghost_normal.png", "ghost.png",
	}

	for _, file := range enemyFiles {
		path := filepath.Join(enemiesPath, file)
		if enemy := ss.loadImage(path); enemy != nil {
			name := file[:len(file)-4]
			ss.Enemies[name] = enemy
		}
	}

	return nil
}

// loadBackgrounds загружает фоны
func (ss *SpriteSheet) loadBackgrounds(basePath string) error {
	bgPath := filepath.Join(basePath, "Background")

	bgFiles := []string{
		"bg.png",
		"bg_castle.png",
		"bg_grasslands.png",
		"bg_desert.png",
		"bg_shroom.png",
	}

	for _, file := range bgFiles {
		path := filepath.Join(bgPath, file)
		if bg := ss.loadImage(path); bg != nil {
			name := file[:len(file)-4]
			ss.Backgrounds[name] = bg
		}
	}

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
	case "happy":
		return ss.PlayerHappy
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

// GetMushroom возвращает гриб по имени
func (ss *SpriteSheet) GetMushroom(name string) *ebiten.Image {
	if mushroom, ok := ss.Mushrooms[name]; ok {
		return mushroom
	}
	return nil
}

// GetCandy возвращает конфету по имени
func (ss *SpriteSheet) GetCandy(name string) *ebiten.Image {
	if candy, ok := ss.Candy[name]; ok {
		return candy
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

// GetFriend возвращает друга по имени
func (ss *SpriteSheet) GetFriend(name string) *ebiten.Image {
	if friend, ok := ss.Friends[name]; ok {
		return friend
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

// GetBackground возвращает фон по имени
func (ss *SpriteSheet) GetBackground(name string) *ebiten.Image {
	if bg, ok := ss.Backgrounds[name]; ok {
		return bg
	}
	return nil
}

// GetTree возвращает дерево по имени
func (ss *SpriteSheet) GetTree(name string) *ebiten.Image {
	// Деревья из папки Trees
	if tree, ok := ss.Tiles["tree"+name]; ok {
		return tree
	}
	// Сосны
	if name == "pine0" || name == "pine1" || name == "pine2" || name == "pine3" || name == "pine4" {
		pineName := "Tree - Pine " + name[len(name)-1:] + ".png"
		if tree, ok := ss.Tiles[pineName]; ok {
			return tree
		}
	}
	return nil
}

// CreatePlaceholder создаёт изображение-заглушку
func CreatePlaceholder(width, height int, r, g, b uint8) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{r, g, b, 255})
	return img
}

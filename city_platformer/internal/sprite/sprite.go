// Package sprite - загрузка пиксельных спрайтов для Pixel Platformer
// Go365 Day 93 - Полностью пиксельная игра!
package sprite

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet хранит все загруженные пиксельные спрайты
type SpriteSheet struct {
	// Игрок
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerHurt   *ebiten.Image

	// Тайлы
	Tiles map[string]*ebiten.Image

	// Предметы
	Items map[string]*ebiten.Image

	// Грибы
	Mushrooms map[string]*ebiten.Image

	// Детали (бабочки, лягушки, кактусы)
	Details map[string]*ebiten.Image

	// Враги
	Enemies map[string]*ebiten.Image

	// Деревья
	Trees map[string]*ebiten.Image

	// Домики
	Houses map[string]*ebiten.Image

	// Фон
	Backgrounds map[string]*ebiten.Image
}

// LoadSpriteSheet загружает все пиксельные спрайты
func LoadSpriteSheet() (*SpriteSheet, error) {
	ss := &SpriteSheet{
		Tiles:       make(map[string]*ebiten.Image),
		Mushrooms:   make(map[string]*ebiten.Image),
		Details:     make(map[string]*ebiten.Image),
		Items:       make(map[string]*ebiten.Image),
		Enemies:     make(map[string]*ebiten.Image),
		Trees:       make(map[string]*ebiten.Image),
		Houses:      make(map[string]*ebiten.Image),
		Backgrounds: make(map[string]*ebiten.Image),
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

	// Загрузка грибов
	if err := ss.loadMushrooms(basePath); err != nil {
		fmt.Println("Warning: mushrooms error:", err)
	}

	// Загрузка деталей
	if err := ss.loadDetails(basePath); err != nil {
		fmt.Println("Warning: details error:", err)
	}

	// Загрузка врагов
	if err := ss.loadEnemies(basePath); err != nil {
		fmt.Println("Warning: enemies error:", err)
	}

	// Загрузка деревьев
	if err := ss.loadTrees(basePath); err != nil {
		fmt.Println("Warning: trees error:", err)
	}

	// Загрузка домиков
	if err := ss.loadHouses(basePath); err != nil {
		fmt.Println("Warning: houses error:", err)
	}

	// Загрузка фона
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
	ss.PlayerHurt = ss.loadImage(filepath.Join(playerPath, "p1_hurt.png"))

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

// loadTiles загружает пиксельные тайлы
func (ss *SpriteSheet) loadTiles(basePath string) error {
	tilesPath := filepath.Join(basePath, "Tiles")

	// Основные тайлы
	tileFiles := []string{
		// Трава
		"grass.png", "grassMid.png", "grassLeft.png", "grassRight.png",
		"grassHalf.png", "grassHalfLeft.png", "grassHalfRight.png",
		// Земля
		"dirt.png", "dirtMid.png", "dirtLeft.png", "dirtRight.png",
		// Камень
		"stone.png", "stoneMid.png", "stoneLeft.png", "stoneRight.png",
		// Кирпичи
		"brickWall.png",
		// Замок
		"castle.png", "castleMid.png", "castleLeft.png", "castleRight.png",
		// Лестницы
		"ladder_mid.png", "ladder_top.png",
		// Выход
		"signExit.png",
	}

	for _, file := range tileFiles {
		path := filepath.Join(tilesPath, file)
		if tile := ss.loadImage(path); tile != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Tiles[name] = tile
		}
	}

	// Загрузка тайлов из Forest pack
	forestPath := filepath.Join(basePath, "Forest")
	forestFiles := []string{
		"forest_pack_03.png", "forest_pack_05.png", "forest_pack_07.png",
		"forest_pack_09.png", "forest_pack_11.png", "forest_pack_13.png",
		"forest_pack_15.png", "forest_pack_17.png", "forest_pack_19.png",
		"forest_pack_21.png", "forest_pack_33.png", "forest_pack_34.png",
		"forest_pack_35.png", "forest_pack_36.png", "forest_pack_37.png",
		"forest_pack_38.png", "forest_pack_39.png", "forest_pack_40.png",
		"forest_pack_41.png", "forest_pack_51.png", "forest_pack_52.png",
		"forest_pack_53.png", "forest_pack_54.png", "forest_pack_55.png",
		"forest_pack_56.png", "forest_pack_57.png", "forest_pack_58.png",
		"forest_pack_59.png", "forest_pack_60.png",
	}

	for _, file := range forestFiles {
		path := filepath.Join(forestPath, file)
		if tile := ss.loadImage(path); tile != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Tiles[name] = tile
		}
	}

	return nil
}

// loadItems загружает пиксельные предметы
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
		"keyBlue.png", "keyGreen.png", "keyRed.png",
		// Облака
		"cloud1.png", "cloud2.png", "cloud3.png",
		// Кактусы
		"cactus.png",
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

// loadMushrooms загружает грибы
func (ss *SpriteSheet) loadMushrooms(basePath string) error {
	mushroomsPath := filepath.Join(basePath, "Mushrooms")

	mushroomFiles := []string{
		// Грибы красные
		"shroomRedLeft.png", "shroomRedMid.png", "shroomRedRight.png",
		// Грибы коричневые
		"shroomBrownLeft.png", "shroomBrownMid.png", "shroomBrownRight.png",
		// Грибы бежевые
		"shroomTanLeft.png", "shroomTanMid.png", "shroomTanRight.png",
		// Высокие грибы
		"tallShroom_red.png", "tallShroom_brown.png", "tallShroom_tan.png",
		// Маленькие грибы
		"tinyShroom_red.png", "tinyShroom_brown.png", "tinyShroom_tan.png",
		// Стебли
		"stem.png", "stemTop.png", "stemBase.png",
	}

	for _, file := range mushroomFiles {
		path := filepath.Join(mushroomsPath, file)
		if mushroom := ss.loadImage(path); mushroom != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Mushrooms[name] = mushroom
		}
	}

	return nil
}

// loadDetails загружает детали (бабочки, лягушки)
func (ss *SpriteSheet) loadDetails(basePath string) error {
	detailsPath := filepath.Join(basePath, "Details")

	detailFiles := []string{
		"GrassLand_Butterfly.png",
		"GrassLand_Frog.png",
		"GrassLand_Flower.png",
		"GrassLand_Bush.png",
	}

	for _, file := range detailFiles {
		path := filepath.Join(detailsPath, file)
		if detail := ss.loadImage(path); detail != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Details[name] = detail
		}
	}

	return nil
}

// loadEnemies загружает пиксельных врагов
func (ss *SpriteSheet) loadEnemies(basePath string) error {
	enemiesPath := filepath.Join(basePath, "Enemies")

	enemyFiles := []string{
		// Муха
		"flyFly1.png", "flyFly2.png",
		// Летучая мышь
		"bat_fly.png", "bat_hit.png", "bat_dead.png",
		// Слайм
		"slimeWalk1.png", "slimeWalk2.png", "slime_dead.png",
		// Змея
		"snakeWalk.png", "snake_hit.png", "snake_dead.png",
		// Паук
		"spider_walk1.png", "spider_walk2.png", "spider_hit.png",
		// Призрак
		"ghost_normal.png", "ghost_hit.png", "ghost_dead.png",
		// Пчела
		"bee_fly.png", "bee_hit.png", "bee_dead.png",
		// Божья коровка
		"ladyBug_walk.png", "ladyBug_hit.png",
		// Лягушка
		"frog.png", "frog_hit.png", "frog_leap.png",
		// Улитка
		"snailWalk1.png", "snailWalk2.png", "snail_hit.png",
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

// loadTrees загружает пиксельные деревья
func (ss *SpriteSheet) loadTrees(basePath string) error {
	treesPath := filepath.Join(basePath, "Trees")

	// Сосны
	pineFiles := []string{
		"Tree - Pine 00.png", "Tree - Pine 01.png", "Tree - Pine 02.png",
		"Tree - Pine 03.png", "Tree - Pine 04.png",
	}

	for _, file := range pineFiles {
		path := filepath.Join(treesPath, file)
		if tree := ss.loadImage(path); tree != nil {
			name := "pine_" + strings.TrimSuffix(file, ".png")
			ss.Trees[name] = tree
		}
	}

	// Деревья из Ice expansion
	treeFiles := []string{
		"tree.png", "treeTop.png", "treeTrunk.png",
		"deadTree.png", "pineSapling.png",
	}

	for _, file := range treeFiles {
		path := filepath.Join(treesPath, file)
		if tree := ss.loadImage(path); tree != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Trees[name] = tree
		}
	}

	return nil
}

// loadHouses загружает пиксельные домики
func (ss *SpriteSheet) loadHouses(basePath string) error {
	housesPath := filepath.Join(basePath, "Tiles")

	// Домики (3 типа x 9 частей = 27 спрайтов)
	houseTypes := []string{"houseBeige", "houseDark", "houseGray"}
	houseParts := []string{
		"TopLeft", "TopMid", "TopRight",
		"MidLeft", "MidRight",
		"BottomLeft", "BottomMid", "BottomRight",
		"Alt", "Alt2",
	}

	for _, houseType := range houseTypes {
		for _, part := range houseParts {
			file := houseType + part + ".png"
			path := filepath.Join(housesPath, file)
			if house := ss.loadImage(path); house != nil {
				name := houseType + "_" + part
				ss.Houses[name] = house
			}
		}
	}

	return nil
}

// loadBackgrounds загружает пиксельные фоны
func (ss *SpriteSheet) loadBackgrounds(basePath string) error {
	bgPath := filepath.Join(basePath, "Background")

	// Основные фоны
	bgFiles := []string{
		"bg.png", "bg_castle.png", "bg_grasslands.png",
		"bg_desert.png", "bg_shroom.png",
	}

	for _, file := range bgFiles {
		path := filepath.Join(bgPath, file)
		if bg := ss.loadImage(path); bg != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Backgrounds[name] = bg
		}
	}

	// Фон из Forest pack
	forestBgPath := filepath.Join(basePath, "Forest")
	if bg := ss.loadImage(filepath.Join(forestBgPath, "bg_forest.png")); bg != nil {
		ss.Backgrounds["bg_forest"] = bg
	}

	// Слои фона
	layersPath := filepath.Join(forestBgPath, "bg_forest_layers")
	layerFiles := []string{"bg_forest_a.png", "bg_forest_b.png", "bg_forest_c.png"}
	for _, file := range layerFiles {
		path := filepath.Join(layersPath, file)
		if layer := ss.loadImage(path); layer != nil {
			name := strings.TrimSuffix(file, ".png")
			ss.Backgrounds[name] = layer
		}
	}

	return nil
}

// loadImage загружает одно изображение с масштабированием
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

// GetTile возвращает тайл
func (ss *SpriteSheet) GetTile(name string) *ebiten.Image {
	if tile, ok := ss.Tiles[name]; ok {
		return tile
	}
	return nil
}

// GetItem возвращает предмет
func (ss *SpriteSheet) GetItem(name string) *ebiten.Image {
	if item, ok := ss.Items[name]; ok {
		return item
	}
	return nil
}

// GetMushroom возвращает гриб
func (ss *SpriteSheet) GetMushroom(name string) *ebiten.Image {
	if mushroom, ok := ss.Mushrooms[name]; ok {
		return mushroom
	}
	return nil
}

// GetDetail возвращает деталь
func (ss *SpriteSheet) GetDetail(name string) *ebiten.Image {
	if detail, ok := ss.Details[name]; ok {
		return detail
	}
	return nil
}

// GetEnemySprite возвращает спрайт врага
func (ss *SpriteSheet) GetEnemySprite(name string) *ebiten.Image {
	if enemy, ok := ss.Enemies[name]; ok {
		return enemy
	}
	return nil
}

// GetTree возвращает дерево
func (ss *SpriteSheet) GetTree(name string) *ebiten.Image {
	if tree, ok := ss.Trees[name]; ok {
		return tree
	}
	return nil
}

// GetHousePart возвращает часть домика
func (ss *SpriteSheet) GetHousePart(houseType, part string) *ebiten.Image {
	name := houseType + "_" + part
	if house, ok := ss.Houses[name]; ok {
		return house
	}
	return nil
}

// GetBackground возвращает фон
func (ss *SpriteSheet) GetBackground(name string) *ebiten.Image {
	if bg, ok := ss.Backgrounds[name]; ok {
		return bg
	}
	return nil
}

// CreatePlaceholder создаёт пиксельную заглушку
func CreatePlaceholder(width, height int, r, g, b uint8) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{r, g, b, 255})
	return img
}

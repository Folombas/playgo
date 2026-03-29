// Package sprite - загрузка и управление спрайтами
// Go365 Day 88 - Food Platformer
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

// FoodType - типы еды
type FoodType int

const (
	FoodFruit FoodType = iota
	FoodVegetable
	FoodMeat
	FoodDairy
	FoodBakery
	FoodJunk
	FoodDrink
	FoodSweet
)

// SpriteSheet - атлас спрайтов Food Platformer
type SpriteSheet struct {
	// Игрок (повар)
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerCook   *ebiten.Image
	
	// Еда (разные стили)
	Food7Soul1    []*ebiten.Image
	FoodDCSS      []*ebiten.Image
	FoodBluecarrot []*ebiten.Image
	FoodGlitch    []*ebiten.Image
	FoodOCAL      []*ebiten.Image
	FoodOther     []*ebiten.Image
	
	// Категории еды
	Fruits      []*ebiten.Image
	Vegetables  []*ebiten.Image
	Meats       []*ebiten.Image
	Dairy       []*ebiten.Image
	Bakery      []*ebiten.Image
	Junk        []*ebiten.Image
	Drinks      []*ebiten.Image
	
	// Враги (испорченная еда)
	EnemyRotten  []*ebiten.Image
	EnemyBug     []*ebiten.Image
	
	// Платформы
	Tiles        map[string]*ebiten.Image
	
	// Декорации
	Kitchen      []*ebiten.Image
}

// LoadSpriteSheet - загрузка спрайт-листа Food Platformer
func LoadSpriteSheet() *SpriteSheet {
	ss := &SpriteSheet{
		PlayerWalk:   make([]*ebiten.Image, 4),
		Food7Soul1:   make([]*ebiten.Image, 0),
		FoodDCSS:     make([]*ebiten.Image, 0),
		FoodBluecarrot: make([]*ebiten.Image, 0),
		FoodGlitch:   make([]*ebiten.Image, 0),
		FoodOCAL:     make([]*ebiten.Image, 0),
		FoodOther:    make([]*ebiten.Image, 0),
		Fruits:       make([]*ebiten.Image, 0),
		Vegetables:   make([]*ebiten.Image, 0),
		Meats:        make([]*ebiten.Image, 0),
		Dairy:        make([]*ebiten.Image, 0),
		Bakery:       make([]*ebiten.Image, 0),
		Junk:         make([]*ebiten.Image, 0),
		Drinks:       make([]*ebiten.Image, 0),
		EnemyRotten:  make([]*ebiten.Image, 2),
		EnemyBug:     make([]*ebiten.Image, 2),
		Tiles:        make(map[string]*ebiten.Image),
	}
	
	// Загрузка игрока (создаём заглушки)
	ss.loadPlayer()
	
	// Загрузка еды из спрайт-листов
	ss.loadFoodSheets()
	
	// Загрузка тайлов
	ss.loadTiles()
	
	// Загрузка врагов
	ss.loadEnemies()
	
	return ss
}

// loadPlayer - загрузка спрайтов игрока
func (ss *SpriteSheet) loadPlayer() {
	// Создаём цветные заглушки для игрока-повара
	ss.PlayerStand = ss.createColoredImage(40, 50, color.RGBA{255, 200, 150, 255})
	
	for i := 0; i < 4; i++ {
		ss.PlayerWalk[i] = ss.createColoredImage(40, 50, color.RGBA{255, 200, 150, 255})
	}
	
	ss.PlayerJump = ss.createColoredImage(40, 50, color.RGBA{255, 200, 150, 255})
	ss.PlayerCook = ss.createColoredImage(40, 50, color.RGBA{255, 255, 255, 255})
}

// loadFoodSheets - загрузка листов с едой
func (ss *SpriteSheet) loadFoodSheets() {
	// 7Soul1 стиль (16x16)
	ss.loadFoodFromSheet("Food/food-7Soul1.png", 16, 16, ss.Food7Soul1)
	
	// DCSS стиль
	ss.loadFoodFromSheet("Food/food-DCSS.png", 32, 32, ss.FoodDCSS)
	
	// bluecarrot16 стиль (32x32)
	ss.loadFoodFromSheet("Food/food-bluecarrot16.png", 32, 32, ss.FoodBluecarrot)
	
	// Glitch стиль
	ss.loadFoodFromSheet("Food/food-Glitch.png", 32, 32, ss.FoodGlitch)
	
	// OCAL стиль
	ss.loadFoodFromSheet("Food/food-OCAL.png", 32, 32, ss.FoodOCAL)
	
	// Other стиль
	ss.loadFoodFromSheet("Food/food-other.png", 32, 32, ss.FoodOther)
}

// loadFoodFromSheet - загрузка еды из спрайт-листа
func (ss *SpriteSheet) loadFoodFromSheet(path string, tileW, tileH int, target []*ebiten.Image) {
	fullPath := "assets/" + strings.ReplaceAll(path, "\\", "/")
	img, _, err := ebitenutil.NewImageFromFile(fullPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load %s: %v\n", fullPath, err)
		return
	}
	
	bounds := img.Bounds()
	cols := bounds.Dx() / tileW
	rows := bounds.Dy() / tileH
	
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			rect := image.Rect(col*tileW, row*tileH, (col+1)*tileW, (row+1)*tileH)
			sprite := img.SubImage(rect).(*ebiten.Image)
			ss.FoodGlitch = append(ss.FoodGlitch, sprite)
		}
	}
}

// loadTiles - загрузка тайлов
func (ss *SpriteSheet) loadTiles() {
	// Кухонные тайлы - создаём заглушки
	ss.Tiles["counter"] = ss.createColoredImage(64, 64, color.RGBA{180, 140, 100, 255})
	ss.Tiles["floor"] = ss.createColoredImage(64, 64, color.RGBA{200, 200, 200, 255})
	ss.Tiles["wall"] = ss.createColoredImage(64, 64, color.RGBA{220, 220, 220, 255})
	ss.Tiles["shelf"] = ss.createColoredImage(64, 48, color.RGBA{160, 120, 80, 255})
	ss.Tiles["stove"] = ss.createColoredImage(64, 64, color.RGBA{100, 100, 100, 255})
}

// loadEnemies - загрузка врагов
func (ss *SpriteSheet) loadEnemies() {
	// Испорченная еда
	ss.EnemyRotten[0] = ss.createColoredImage(32, 32, color.RGBA{100, 150, 50, 255})
	ss.EnemyRotten[1] = ss.createColoredImage(32, 32, color.RGBA{120, 160, 60, 255})
	
	// Насекомые
	ss.EnemyBug[0] = ss.createColoredImage(24, 24, color.RGBA{80, 60, 40, 255})
	ss.EnemyBug[1] = ss.createColoredImage(24, 24, color.RGBA{90, 70, 50, 255})
}

// createColoredImage - создание цветного изображения
func (ss *SpriteSheet) createColoredImage(width, height int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return ebiten.NewImageFromImage(img)
}

// GetFood - получение спрайта еды по типу
func (ss *SpriteSheet) GetFood(foodType FoodType, index int) *ebiten.Image {
	var foods []*ebiten.Image
	
	switch foodType {
	case FoodFruit:
		foods = ss.Fruits
	case FoodVegetable:
		foods = ss.Vegetables
	case FoodMeat:
		foods = ss.Meats
	case FoodDairy:
		foods = ss.Dairy
	case FoodBakery:
		foods = ss.Bakery
	case FoodJunk:
		foods = ss.Junk
	case FoodDrink:
		foods = ss.Drinks
	case FoodSweet:
		foods = ss.FoodGlitch
	}
	
	if len(foods) > 0 && index >= 0 && index < len(foods) {
		return foods[index]
	}
	
	// Возвращаем случайную еду из Glitch набора
	if len(ss.FoodGlitch) > 0 {
		return ss.FoodGlitch[index%len(ss.FoodGlitch)]
	}
	
	return nil
}

// GetRandomFood - получение случайной еды
func (ss *SpriteSheet) GetRandomFood(index int) *ebiten.Image {
	allFoods := []*ebiten.Image{
		ss.Food7Soul1[0],
		ss.FoodDCSS[0],
		ss.FoodBluecarrot[0],
		ss.FoodGlitch[0],
		ss.FoodOCAL[0],
	}
	
	for _, foods := range allFoods {
		if foods != nil {
			return foods
		}
	}
	
	return ss.createColoredImage(32, 32, color.RGBA{255, 200, 50, 255})
}

// GetEnemy - получение спрайта врага
func (ss *SpriteSheet) GetEnemy(enemyType string, frame int) *ebiten.Image {
	switch enemyType {
	case "rotten":
		if frame >= 0 && frame < len(ss.EnemyRotten) {
			return ss.EnemyRotten[frame%len(ss.EnemyRotten)]
		}
	case "bug":
		if frame >= 0 && frame < len(ss.EnemyBug) {
			return ss.EnemyBug[frame%len(ss.EnemyBug)]
		}
	}
	return ss.EnemyRotten[0]
}

// GetTile - получение тайла
func (ss *SpriteSheet) GetTile(tileType string) *ebiten.Image {
	if img, ok := ss.Tiles[tileType]; ok {
		return img
	}
	return ss.Tiles["counter"]
}

// GetPlayerFrame - получение кадра игрока
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
	case "cook":
		return ss.PlayerCook
	}
	return ss.PlayerStand
}

// LoadImageFromFile - публичная функция загрузки
func LoadImageFromFile(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	return img, err
}

// GetFoodValue - получение ценности еды
func GetFoodValue(foodType FoodType) int {
	switch foodType {
	case FoodFruit, FoodVegetable:
		return 10 // Полезная еда
	case FoodMeat, FoodDairy, FoodBakery:
		return 15 // Средняя ценность
	case FoodSweet:
		return 20 // Сладости
	case FoodDrink:
		return 5 // Напитки
	case FoodJunk:
		return -10 // Вредная еда (штраф)
	default:
		return 10
	}
}

// GetFoodColor - получение цвета для еды
func GetFoodColor(foodType FoodType) color.RGBA {
	switch foodType {
	case FoodFruit:
		return color.RGBA{255, 100, 100, 255} // Красный
	case FoodVegetable:
		return color.RGBA{100, 200, 100, 255} // Зелёный
	case FoodMeat:
		return color.RGBA{180, 80, 80, 255} // Коричневый
	case FoodDairy:
		return color.RGBA{255, 255, 200, 255} // Белый
	case FoodBakery:
		return color.RGBA{200, 150, 80, 255} // Золотой
	case FoodJunk:
		return color.RGBA{150, 100, 50, 255} // Коричневый
	case FoodDrink:
		return color.RGBA{100, 150, 255, 255} // Синий
	case FoodSweet:
		return color.RGBA{255, 100, 200, 255} // Розовый
	default:
		return color.RGBA{255, 200, 50, 255}
	}
}

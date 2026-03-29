// Package sprite - загрузка и управление спрайтами
// Go365 Day 88
package sprite

import (
	_ "embed"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// SpriteSheet - атлас спрайтов
type SpriteSheet struct {
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerCrouch *ebiten.Image
	
	EnemySlime   []*ebiten.Image
	EnemyFly     []*ebiten.Image
	EnemySnail   []*ebiten.Image
	
	Tiles        map[string]*ebiten.Image
	Items        map[string]*ebiten.Image
}

// LoadSpriteSheet - загрузка спрайт-листа
// Пока возвращает пустую структуру, спрайты будут добавлены позже
func LoadSpriteSheet() *SpriteSheet {
	ss := &SpriteSheet{
		PlayerWalk: make([]*ebiten.Image, 4),
		EnemySlime: make([]*ebiten.Image, 2),
		EnemyFly:   make([]*ebiten.Image, 2),
		EnemySnail: make([]*ebiten.Image, 2),
		Tiles:      make(map[string]*ebiten.Image),
		Items:      make(map[string]*ebiten.Image),
	}
	
	// Создаём временные спрайты (заглушки)
	ss.createPlaceholderSprites()
	
	return ss
}

// createPlaceholderSprites - создание временных спрайтов-заглушек
func (ss *SpriteSheet) createPlaceholderSprites() {
	// Игрок - стоя
	ss.PlayerStand = createPlaceholderImage(32, 32, color.RGBA{50, 150, 255, 255})
	
	// Игрок - бег (4 кадра)
	for i := 0; i < 4; i++ {
		ss.PlayerWalk[i] = createPlaceholderImage(32, 32, color.RGBA{50, 150, 255, 255})
	}
	
	// Игрок - прыжок
	ss.PlayerJump = createPlaceholderImage(32, 32, color.RGBA{50, 150, 255, 255})
	
	// Игрок - присед
	ss.PlayerCrouch = createPlaceholderImage(32, 20, color.RGBA{50, 150, 255, 255})
	
	// Враги
	for i := 0; i < 2; i++ {
		ss.EnemySlime[i] = createPlaceholderImage(32, 32, color.RGBA{50, 200, 50, 255})
		ss.EnemyFly[i] = createPlaceholderImage(32, 24, color.RGBA{200, 200, 50, 255})
		ss.EnemySnail[i] = createPlaceholderImage(36, 28, color.RGBA{200, 100, 50, 255})
	}
	
	// Предметы
	ss.Items["coin"] = createPlaceholderImage(24, 24, color.RGBA{255, 215, 0, 255})
	ss.Items["gem_red"] = createPlaceholderImage(28, 28, color.RGBA{255, 50, 50, 255})
	ss.Items["gem_blue"] = createPlaceholderImage(28, 28, color.RGBA{50, 150, 255, 255})
	ss.Items["gem_green"] = createPlaceholderImage(28, 28, color.RGBA{50, 255, 100, 255})
	
	// Плитки
	ss.Tiles["ground"] = createPlaceholderImage(32, 32, color.RGBA{100, 180, 100, 255})
	ss.Tiles["stone"] = createPlaceholderImage(32, 32, color.RGBA{150, 150, 150, 255})
	ss.Tiles["brick"] = createPlaceholderImage(32, 32, color.RGBA{180, 100, 80, 255})
	ss.Tiles["metal"] = createPlaceholderImage(32, 32, color.RGBA{200, 200, 220, 255})
}

// createPlaceholderImage - создание изображения-заглушки
func createPlaceholderImage(width, height int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Заполняем цветом
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// LoadImageFromFile - загрузка изображения из файла
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
	case "crouch":
		return ss.PlayerCrouch
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
	}
	return ss.EnemySlime[0]
}

// GetItem - получение предмета
func (ss *SpriteSheet) GetItem(itemType string) *ebiten.Image {
	if img, ok := ss.Items[itemType]; ok {
		return img
	}
	return ss.Items["coin"]
}

// GetTile - получение плитки
func (ss *SpriteSheet) GetTile(tileType string) *ebiten.Image {
	if img, ok := ss.Tiles[tileType]; ok {
		return img
	}
	return ss.Tiles["stone"]
}

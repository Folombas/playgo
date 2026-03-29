// Package sprite - загрузка и управление спрайтами
// Go365 Day 88 - Forest Pack
package sprite

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// SpriteSheet - атлас спрайтов
type SpriteSheet struct {
	// Игрок
	PlayerStand  *ebiten.Image
	PlayerWalk   []*ebiten.Image
	PlayerJump   *ebiten.Image
	PlayerCrouch *ebiten.Image
	
	// Враги
	EnemySlime   []*ebiten.Image
	EnemyFly     []*ebiten.Image
	EnemySnail   []*ebiten.Image
	
	// Платформы и земля
	Tiles        map[string]*ebiten.Image
	GroundTop    *ebiten.Image
	GroundMid    *ebiten.Image
	
	// Декорации
	Tree         *ebiten.Image
	Flower       *ebiten.Image
	Mushroom     *ebiten.Image
	Rock         *ebiten.Image
	Stone        *ebiten.Image
	
	// Фон
	BgForest     *ebiten.Image
	BgLayerA     *ebiten.Image
	BgLayerB     *ebiten.Image
	BgLayerC     *ebiten.Image
	
	// Предметы
	Items        map[string]*ebiten.Image
}

// LoadSpriteSheet - загрузка спрайт-листа
func LoadSpriteSheet() *SpriteSheet {
	ss := &SpriteSheet{
		PlayerWalk: make([]*ebiten.Image, 4),
		EnemySlime: make([]*ebiten.Image, 2),
		EnemyFly:   make([]*ebiten.Image, 2),
		EnemySnail: make([]*ebiten.Image, 2),
		Tiles:      make(map[string]*ebiten.Image),
		Items:      make(map[string]*ebiten.Image),
	}
	
	// Загрузка тайлов земли
	ss.loadGroundTiles()
	
	// Загрузка фона
	ss.loadBackground()
	
	// Загрузка декораций
	ss.loadDecorations()
	
	// Создание временных спрайтов для сущностей
	ss.createEntitySprites()
	
	return ss
}

// loadGroundTiles - загрузка тайлов земли
func (ss *SpriteSheet) loadGroundTiles() {
	// Основной тайл земли с травой
	groundTop, err := LoadImageFromFile("assets/forest_pack_03.png")
	if err == nil {
		ss.GroundTop = groundTop
		ss.Tiles["ground_top"] = groundTop
	}
	
	// Средний тайл земли
	groundMid, err := LoadImageFromFile("assets/forest_pack_35.png")
	if err == nil {
		ss.GroundMid = groundMid
		ss.Tiles["ground_mid"] = groundMid
	}
	
	// Тайлы для платформ
	ss.Tiles["grass"] = ss.loadTile("forest_pack_03.png", 32, 32)
	ss.Tiles["dirt"] = ss.loadTile("forest_pack_35.png", 32, 32)
}

// loadBackground - загрузка фона
func (ss *SpriteSheet) loadBackground() {
	// Основной фон леса
	bg, err := LoadImageFromFile("assets/bg_forest.png")
	if err == nil {
		ss.BgForest = bg
	}
	
	// Слои параллакса
	ss.BgLayerA = ss.loadLayer("bg_forest_layers/bg_forest_a.png")
	ss.BgLayerB = ss.loadLayer("bg_forest_layers/bg_forest_b.png")
	ss.BgLayerC = ss.loadLayer("bg_forest_layers/bg_forest_c.png")
}

// loadLayer - загрузка слоя фона
func (ss *SpriteSheet) loadLayer(path string) *ebiten.Image {
	img, err := LoadImageFromFile("assets/" + path)
	if err != nil {
		return nil
	}
	return img
}

// loadDecorations - загрузка декораций
func (ss *SpriteSheet) loadDecorations() {
	ss.Tree = ss.loadDecoration("forest_pack_92.png")
	ss.Flower = ss.loadDecoration("forest_pack_103.png")
	ss.Mushroom = ss.loadDecoration("forest_pack_104.png")
	ss.Rock = ss.loadDecoration("forest_pack_105.png")
	ss.Stone = ss.loadDecoration("forest_pack_106.png")
}

// loadDecoration - загрузка элемента декорации
func (ss *SpriteSheet) loadDecoration(filename string) *ebiten.Image {
	img, err := LoadImageFromFile("assets/" + filename)
	if err != nil {
		return nil
	}
	return img
}

// loadTile - загрузка тайла с указанным размером
func (ss *SpriteSheet) loadTile(filename string, width, height int) *ebiten.Image {
	img, err := LoadImageFromFile("assets/" + filename)
	if err != nil {
		return nil
	}
	
	// Возвращаем суб-изображение нужного размера
	return img.SubImage(image.Rect(0, 0, width, height)).(*ebiten.Image)
}

// createEntitySprites - создание спрайтов для сущностей
func (ss *SpriteSheet) createEntitySprites() {
	// Игрок - используем цвета Forest Pack (зелёные/коричневые тона)
	ss.PlayerStand = createPlaceholderImage(32, 32, color.RGBA{100, 180, 100, 255})
	
	for i := 0; i < 4; i++ {
		ss.PlayerWalk[i] = createPlaceholderImage(32, 32, color.RGBA{100, 180, 100, 255})
	}
	
	ss.PlayerJump = createPlaceholderImage(32, 32, color.RGBA{100, 180, 100, 255})
	ss.PlayerCrouch = createPlaceholderImage(32, 20, color.RGBA{100, 180, 100, 255})
	
	// Враги - лесные цвета
	for i := 0; i < 2; i++ {
		ss.EnemySlime[i] = createPlaceholderImage(32, 32, color.RGBA{50, 150, 50, 255})
		ss.EnemyFly[i] = createPlaceholderImage(32, 24, color.RGBA{180, 150, 50, 255})
		ss.EnemySnail[i] = createPlaceholderImage(36, 28, color.RGBA{150, 100, 50, 255})
	}
	
	// Предметы
	ss.Items["coin"] = createPlaceholderImage(24, 24, color.RGBA{255, 200, 50, 255})
	ss.Items["gem_red"] = createPlaceholderImage(28, 28, color.RGBA{255, 80, 80, 255})
	ss.Items["gem_blue"] = createPlaceholderImage(28, 28, color.RGBA{80, 150, 255, 255})
	ss.Items["gem_green"] = createPlaceholderImage(28, 28, color.RGBA{80, 255, 100, 255})
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
	return ss.Tiles["grass"]
}

// DrawTiled - отрисовка тайлами
func (ss *SpriteSheet) DrawTiled(screen *ebiten.Image, tile *ebiten.Image, x, y, width, height float64, cameraX, cameraY float64) {
	if tile == nil {
		return
	}
	
	tileW := float64(tile.Bounds().Dx())
	tileH := float64(tile.Bounds().Dy())
	
	startX := int((x - cameraX) / tileW) * int(tileW)
	startY := int((y - cameraY) / tileH) * int(tileH)
	
	for ty := startY; ty < int(y+height-cameraY); ty += int(tileH) {
		for tx := startX; tx < int(x+width-cameraX); tx += int(tileW) {
			if tx >= -int(tileW) && tx < int(width) && ty >= -int(tileH) && ty < int(height) {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(tx), float64(ty))
				screen.DrawImage(tile, op)
			}
		}
	}
}

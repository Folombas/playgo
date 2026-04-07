package render

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/puzzle_go/internal/entity"
)

//go:embed assets/sprites/*.png assets/sprites/backtiles/*.png
var spritesFS embed.FS

// SpriteManager управляет загрузкой и кэшированием спрайтов
type SpriteManager struct {
	sprites map[string]*ebiten.Image
	mu      sync.RWMutex
}

// NewSpriteManager создает новый менеджер спрайтов
func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		sprites: make(map[string]*ebiten.Image),
	}
}

// Load загружает все спрайты
func (sm *SpriteManager) Load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Загрузка кристаллов
	crystalFiles := []string{
		"gem0.png", "gem1.png", "gem2.png", "gem3.png", "gem4.png", "gem5.png",
	}

	crystalNames := []string{
		"red", "blue", "green", "yellow", "violet", "orange",
	}

	for i, file := range crystalFiles {
		img, err := loadPNG("assets/sprites/" + file)
		if err != nil {
			log.Printf("Warning: could not load %s: %v", file, err)
			continue
		}
		sm.sprites["crystal_"+crystalNames[i]] = img
		log.Printf("Loaded sprite: crystal_%s", crystalNames[i])
	}

	// Загрузка фона
	for i := 1; i <= 18; i++ {
		filename := "assets/sprites/backtiles/BackTile_" + formatNum(i) + ".png"
		img, err := loadPNG(filename)
		if err != nil {
			continue
		}
		sm.sprites["background_tile_"+string(rune('0'+i))] = img
	}

	log.Printf("Loaded %d sprites", len(sm.sprites))
	return nil
}

func formatNum(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// GetCrystalImage возвращает изображение кристалла
func (sm *SpriteManager) GetCrystalImage(crystalType entity.CrystalType) *ebiten.Image {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	crystalNames := []string{
		"crystal_red", "crystal_blue", "crystal_green",
		"crystal_yellow", "crystal_violet", "crystal_orange",
	}

	if crystalType >= entity.CrystalRed && crystalType <= entity.CrystalOrange {
		if img, ok := sm.sprites[crystalNames[crystalType]]; ok {
			return img
		}
	}

	// fallback на gem0
	if img, ok := sm.sprites["crystal_red"]; ok {
		return img
	}

	// Создаем placeholder
	return sm.createPlaceholder()
}

func (sm *SpriteManager) createPlaceholder() *ebiten.Image {
	img := ebiten.NewImage(64, 64)
	img.Fill(color.RGBA{100, 100, 255, 255})
	return img
}

func loadPNG(path string) (*ebiten.Image, error) {
	data, err := spritesFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

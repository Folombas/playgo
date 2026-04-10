package game

import (
	"bytes"
	"embed"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assetFS embed.FS

// SpriteManager управляет загрузкой и хранением спрайтов
type SpriteManager struct {
	gems       [6]*ebiten.Image
	background *ebiten.Image
	loaded     bool
}

// NewSpriteManager создаёт новый менеджер спрайтов
func NewSpriteManager() *SpriteManager {
	return &SpriteManager{}
}

// Load загружает все спрайты из embedded файлов
func (sm *SpriteManager) Load() {
	if sm.loaded {
		return
	}

	log.Println("Loading sprites...")

	// Загрузка камней
	gemNames := []string{"red", "green", "yellow", "blue", "purple", "orange"}
	for i, name := range gemNames {
		sm.gems[i] = sm.loadGemSprite("assets/gems/" + name + ".png")
	}

	// Загрузка фона
	sm.background = sm.loadImage("assets/backgrounds/game_bg.png")

	sm.loaded = true
	log.Println("Sprites loaded successfully!")
}

// loadGemSprite загружает спрайт камня из PNG файла
func (sm *SpriteManager) loadGemSprite(path string) *ebiten.Image {
	data, err := assetFS.ReadFile(path)
	if err != nil {
		log.Printf("Warning: Failed to load gem sprite %s: %v", path, err)
		// Возвращаем пустой спрайт
		return ebiten.NewImage(32, 32)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("Warning: Failed to decode gem sprite %s: %v", path, err)
		return ebiten.NewImage(32, 32)
	}

	return ebiten.NewImageFromImage(img)
}

// loadImage загружает произвольное изображение
func (sm *SpriteManager) loadImage(path string) *ebiten.Image {
	data, err := assetFS.ReadFile(path)
	if err != nil {
		log.Printf("Warning: Failed to load image %s: %v", path, err)
		return ebiten.NewImage(640, 960)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("Warning: Failed to decode image %s: %v", path, err)
		return ebiten.NewImage(640, 960)
	}

	return ebiten.NewImageFromImage(img)
}

// GetGemSprite возвращает спрайт камня по типу
func (sm *SpriteManager) GetGemSprite(gemType int) *ebiten.Image {
	if !sm.loaded {
		sm.Load()
	}
	if gemType >= 0 && gemType < 6 {
		return sm.gems[gemType]
	}
	return sm.gems[0]
}

// GetBackground возвращает фоновое изображение
func (sm *SpriteManager) GetBackground() *ebiten.Image {
	if !sm.loaded {
		sm.Load()
	}
	return sm.background
}

// IsLoaded проверяет, загружены ли спрайты
func (sm *SpriteManager) IsLoaded() bool {
	return sm.loaded
}

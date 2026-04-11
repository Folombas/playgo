package logic

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// ImageCache кэширует часто используемые изображения
// чтобы избежать постоянных аллокаций каждый кадр
type ImageCache struct {
	tileBackgrounds map[int]*ebiten.Image
	highlights      map[int]*ebiten.Image
	bombIndicators  map[int]*ebiten.Image
	mu              sync.RWMutex
}

// NewImageCache создаёт новый кэш изображений
func NewImageCache() *ImageCache {
	return &ImageCache{
		tileBackgrounds: make(map[int]*ebiten.Image),
		highlights:      make(map[int]*ebiten.Image),
		bombIndicators:  make(map[int]*ebiten.Image),
	}
}

// GetTileBackground возвращает или создаёт фон для ячейки
func (ic *ImageCache) GetTileBackground(size int) *ebiten.Image {
	ic.mu.RLock()
	if img, ok := ic.tileBackgrounds[size]; ok {
		ic.mu.RUnlock()
		return img
	}
	ic.mu.RUnlock()

	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Двойная проверка после блокировки
	if img, ok := ic.tileBackgrounds[size]; ok {
		return img
	}

	img := ebiten.NewImage(size, size)
	img.Fill(color.RGBA{200, 200, 200, 255})
	ic.tileBackgrounds[size] = img
	return img
}

// GetHighlight возвращает или создаёт изображение выделения
func (ic *ImageCache) GetHighlight(size int) *ebiten.Image {
	ic.mu.RLock()
	if img, ok := ic.highlights[size]; ok {
		ic.mu.RUnlock()
		return img
	}
	ic.mu.RUnlock()

	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Двойная проверка после блокировки
	if img, ok := ic.highlights[size]; ok {
		return img
	}

	img := ebiten.NewImage(size, size)
	img.Fill(color.White)
	ic.highlights[size] = img
	return img
}

// GetBombIndicator возвращает или создаёт индикатор бомбы
func (ic *ImageCache) GetBombIndicator(size int) *ebiten.Image {
	ic.mu.RLock()
	if img, ok := ic.bombIndicators[size]; ok {
		ic.mu.RUnlock()
		return img
	}
	ic.mu.RUnlock()

	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Двойная проверка после блокировки
	if img, ok := ic.bombIndicators[size]; ok {
		return img
	}

	img := ebiten.NewImage(size, size)
	img.Fill(color.RGBA{255, 50, 50, 255})
	ic.bombIndicators[size] = img
	return img
}

// Clear очищает кэш (если нужно освободить память)
func (ic *ImageCache) Clear() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.tileBackgrounds = make(map[int]*ebiten.Image)
	ic.highlights = make(map[int]*ebiten.Image)
	ic.bombIndicators = make(map[int]*ebiten.Image)
}

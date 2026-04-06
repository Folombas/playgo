package sprite

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteManager handles loading and caching sprites
type SpriteManager struct {
	cache map[string]*ebiten.Image
}

func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		cache: make(map[string]*ebiten.Image),
	}
}

// Load loads a sprite from file and caches it
func (sm *SpriteManager) Load(path string) (*ebiten.Image, error) {
	if img, ok := sm.cache[path]; ok {
		return img, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	eImg := ebiten.NewImageFromImage(img)
	sm.cache[path] = eImg

	return eImg, nil
}

// LoadDir loads all PNG sprites from a directory
func (sm *SpriteManager) LoadDir(dir string) (map[string]*ebiten.Image, error) {
	sprites := make(map[string]*ebiten.Image)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		img, err := sm.Load(path)
		if err != nil {
			continue // Skip failed loads
		}

		// Use filename without extension as key
		key := entry.Name()[:len(entry.Name())-4]
		sprites[key] = img
	}

	return sprites, nil
}

// Get returns cached sprite by path
func (sm *SpriteManager) Get(path string) *ebiten.Image {
	return sm.cache[path]
}

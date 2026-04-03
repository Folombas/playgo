package sprite

import (
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// Loader manages loading and caching of sprite images.
type Loader struct {
	cache   map[string]*ebiten.Image
	baseDir string
}

// NewLoader creates a new sprite loader.
func NewLoader(baseDir string) *Loader {
	return &Loader{
		cache:   make(map[string]*ebiten.Image),
		baseDir: baseDir,
	}
}

// Load loads a sprite from path relative to baseDir and caches it.
func (l *Loader) Load(relPath string) (*ebiten.Image, error) {
	if img, ok := l.cache[relPath]; ok {
		return img, nil
	}

	fullPath := filepath.Join(l.baseDir, relPath)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	eImg := ebiten.NewImageFromImage(img)
	l.cache[relPath] = eImg
	return eImg, nil
}

// LoadPNG decodes a PNG file and returns ebiten.Image without caching.
func LoadPNG(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// SubImage returns a sub-image from the given bounds.
func SubImage(src *ebiten.Image, bounds image.Rectangle) *ebiten.Image {
	return src.SubImage(bounds).(*ebiten.Image)
}

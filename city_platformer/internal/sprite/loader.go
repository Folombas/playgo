package sprite

import (
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// Load loads a PNG file and returns an ebiten.Image.
func Load(path string) (*ebiten.Image, error) {
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

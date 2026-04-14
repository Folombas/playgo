package main

import (
	"bytes"
	"embed"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed sprites/*.png
var spritesFS embed.FS

// TileImages хранит загруженные спрайты для фишек
var TileImages []*ebiten.Image

// LoadSprites загружает все спрайты из embedded файловой системы
func LoadSprites() {
	TileImages = make([]*ebiten.Image, 6)

	for i := 0; i < 6; i++ {
		data, err := spritesFS.ReadFile("sprites/tile_" + string(rune('0'+i)) + ".png")
		if err != nil {
			log.Printf("Warning: could not load sprite tile_%d.png, using fallback", i)
			TileImages[i] = createFallbackTile(i)
			continue
		}

		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: could not decode sprite tile_%d.png, using fallback", i)
			TileImages[i] = createFallbackTile(i)
			continue
		}

		TileImages[i] = ebiten.NewImageFromImage(img)
	}
}

// createFallbackTile создаёт цветную заглушку, если спрайт не загрузился
func createFallbackTile(colorIndex int) *ebiten.Image {
	img := ebiten.NewImage(64, 64)

	// Цвета для разных фишек
	colors := []struct{ r, g, b byte }{
		{255, 50, 50},   // 0: красный
		{50, 50, 255},   // 1: синий
		{50, 255, 50},   // 2: зелёный
		{255, 255, 50},  // 3: жёлтый
		{180, 50, 255},  // 4: фиолетовый
		{255, 165, 0},   // 5: оранжевый
	}

	c := colors[colorIndex%len(colors)]
	// Заполняем простым цветом (в реальной игре лучше использовать vector.DrawFilledCircle)
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, &colorRGBA{r: c.r, g: c.g, b: c.b, a: 255})
		}
	}

	return img
}

// colorRGBA реализует image.Color для ручной отрисовки
type colorRGBA struct {
	r, g, b, a byte
}

func (c *colorRGBA) RGBA() (r, g, b, a uint32) {
	return uint32(c.r) * 0x101, uint32(c.g) * 0x101, uint32(c.b) * 0x101, uint32(c.a) * 0x101
}

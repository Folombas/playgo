// Package render — загрузка спрайтов и утилиты отрисовки.
package render

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/playgo/puzzle_go/internal/config"
	"github.com/playgo/puzzle_go/internal/entity"
)

// SpriteCache хранит все загруженные спрайты.
type SpriteCache struct {
	Gems       [config.GemTypes]*ebiten.Image
	Coin       *ebiten.Image
	Selector   *ebiten.Image
	BGTile     *ebiten.Image
	HUDBG      *ebiten.Image
	WhitePixel *ebiten.Image
}

// LoadSprites загружает все спрайты.
func LoadSprites() *SpriteCache {
	sc := &SpriteCache{}
	sc.WhitePixel = ebiten.NewImage(1, 1)
	sc.WhitePixel.Fill(color.White)

	jewelNames := []string{"jewelblue_0", "jewelred", "jewelgreen", "jewelyellow", "jewelviolet", "jewelorange"}
	fallback := []string{"gem0", "gem1", "gem2", "gem3", "gem4", "gem5"}
	for i := 0; i < config.GemTypes; i++ {
		img, _, _ := ebitenutil.NewImageFromFile(fmt.Sprintf("assets/sprites/%s.png", jewelNames[i]))
		if img == nil {
			img, _, _ = ebitenutil.NewImageFromFile(fmt.Sprintf("assets/sprites/%s.png", fallback[i]))
		}
		if img != nil {
			w, h := img.Size()
			if w != config.Tile || h != config.Tile {
				sc.Gems[i] = scaleTo(img, config.Tile, config.Tile)
			} else {
				sc.Gems[i] = img
			}
		}
	}

	sc.Coin, _, _ = ebitenutil.NewImageFromFile("assets/sprites/coin.png")
	sc.Selector, _, _ = ebitenutil.NewImageFromFile("assets/sprites/selector.png")
	sc.BGTile, _, _ = ebitenutil.NewImageFromFile("assets/sprites/ground.png")
	if sc.BGTile == nil {
		sc.BGTile, _, _ = ebitenutil.NewImageFromFile("assets/sprites/backtiles/BackTile_01.png")
	}
	if sc.BGTile == nil {
		sc.BGTile = ebiten.NewImage(config.Tile, config.Tile)
		sc.BGTile.Fill(color.RGBA{30, 30, 50, 255})
	}
	sc.HUDBG = ebiten.NewImage(config.WinW, config.HUD)
	sc.HUDBG.Fill(color.RGBA{20, 20, 40, 255})
	return sc
}

func scaleTo(src *ebiten.Image, w, h int) *ebiten.Image {
	dst := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	sw, sh := src.Size()
	op.GeoM.Scale(float64(w)/float64(sw), float64(h)/float64(sh))
	dst.DrawImage(src, op)
	return dst
}

// GemColor возвращает цвет гема.
func GemColor(id int) color.Color {
	c := []color.Color{
		color.RGBA{200, 220, 255, 255}, color.RGBA{255, 80, 80, 255},
		color.RGBA{80, 220, 80, 255}, color.RGBA{255, 220, 80, 255},
		color.RGBA{180, 80, 255, 255}, color.RGBA{255, 160, 40, 255},
	}
	if id >= 0 && id < len(c) { return c[id] }
	return color.White
}

// DrawRect рисует заполненный прямоугольник через белый пиксель.
func DrawRect(s, wp *ebiten.Image, x, y, w, h int, c color.Color, a float32) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.GeoM.Scale(float64(w), float64(h))
	if col, ok := c.(color.RGBA); ok {
		op.ColorScale.SetR(float32(col.R)/255)
		op.ColorScale.SetG(float32(col.G)/255)
		op.ColorScale.SetB(float32(col.B)/255)
		op.ColorScale.SetA(float32(col.A)/255 * a)
	}
	s.DrawImage(wp, op)
}

// SpawnParts создаёт частицы.
func SpawnParts(x, y float64, clr color.Color, n int, out *[]entity.Particle) {
	for i := 0; i < n; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1 + rand.Float64()*3
		*out = append(*out, entity.Particle{
			X: x, Y: y, VX: math.Cos(angle)*speed, VY: math.Sin(angle)*speed - 2,
			Life: 30 + rand.Intn(30), MaxLife: 60, Clr: clr,
			Sz: 2 + rand.Intn(5), RotSpeed: (rand.Float64()-0.5)*0.2,
		})
	}
}

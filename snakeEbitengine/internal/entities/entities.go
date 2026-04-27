package entities

import (
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func LoadPNG(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func LoadSpriteSheet(path string, frameW, frameH, cols, rows int, removeBg bool, bgColor color.Color) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	sheetW := bounds.Dx()
	sheetH := bounds.Dy()
	var frames []*ebiten.Image
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * frameW
			y := row * frameH
			if x+frameW > sheetW || y+frameH > sheetH {
				continue
			}
			subImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+frameW, y+frameH))
			ebImg := ebiten.NewImageFromImage(subImg)
			if removeBg {
				ebImg = MakeColorTransparent(ebImg, bgColor)
			}
			frames = append(frames, ebImg)
		}
	}
	return frames, nil
}

func LoadSpriteSheetAuto(path string, cols, rows int, removeBg bool, bgColor color.Color) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	frameW := bounds.Dx() / cols
	frameH := bounds.Dy() / rows
	return LoadSpriteSheet(path, frameW, frameH, cols, rows, removeBg, bgColor)
}

func MakeColorTransparent(img *ebiten.Image, targetColor color.Color) *ebiten.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rgba := image.NewRGBA(bounds)
	drawImg := ebiten.NewImageFromImage(rgba)
	drawImg.Clear()
	drawImg.DrawImage(img, nil)
	pix := rgba.Pix
	stride := rgba.Stride
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := y*stride + x*4
			r, g, b, a := pix[off], pix[off+1], pix[off+2], pix[off+3]
			tr, tg, tb, _ := targetColor.RGBA()
			tr8 := uint8(tr >> 8)
			tg8 := uint8(tg >> 8)
			tb8 := uint8(tb >> 8)
			if absDiff(r, tr8) < 5 && absDiff(g, tg8) < 5 && absDiff(b, tb8) < 5 && a > 0 {
				pix[off+3] = 0
			}
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

func absDiff(a, b uint8) int {
	diff := int(a) - int(b)
	if diff < 0 {
		return -diff
	}
	return diff
}

func LoadFont(path string, size float64) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tt, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	const dpi = 72
	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}
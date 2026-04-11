package logic

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Star представляет звезду на фоне
type Star struct {
	X, Y     float64
	Size     float64
	TwinkleSpeed float64
	TwinklePhase float64
	Brightness float64
}

// BackgroundEffect управляет фоновыми эффектами
type BackgroundEffect struct {
	stars []*Star
	width int
	height int
}

// NewBackgroundEffect создаёт новый фоновый эффект
func NewBackgroundEffect(width, height int) *BackgroundEffect {
	be := &BackgroundEffect{
		stars:  make([]*Star, 0),
		width:  width,
		height: height,
	}
	
	be.generateStars(50)
	
	return be
}

// generateStars генерирует случайные звёзды
func (be *BackgroundEffect) generateStars(count int) {
	for i := 0; i < count; i++ {
		star := &Star{
			X:            rand.Float64() * float64(be.width),
			Y:            rand.Float64() * float64(be.height),
			Size:         1 + rand.Float64()*2,
			TwinkleSpeed: 0.5 + rand.Float64()*2,
			TwinklePhase: rand.Float64() * math.Pi * 2,
			Brightness:   0.5 + rand.Float64()*0.5,
		}
		be.stars = append(be.stars, star)
	}
}

// Update обновляет фоновые эффекты
func (be *BackgroundEffect) Update(deltaTime float64) {
	for _, star := range be.stars {
		star.TwinklePhase += star.TwinkleSpeed * deltaTime
	}
}

// Draw отрисовывает фоновые эффекты
func (be *BackgroundEffect) Draw(screen *ebiten.Image) {
	for _, star := range be.stars {
		// Рассчитываем яркость с пульсацией
		brightness := star.Brightness * (0.5 + 0.5*math.Sin(star.TwinklePhase))
		alpha := uint8(brightness * 255)
		
		starColor := color.RGBA{255, 255, 255, alpha}
		
		// Рисуем звезду
		starImg := ebiten.NewImage(int(star.Size), int(star.Size))
		starImg.Fill(starColor)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(star.X, star.Y)
		screen.DrawImage(starImg, op)
	}
}

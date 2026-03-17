package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	groundHeight = 100
)

type Game struct {
	playerX float64
	playerY float64
}

func NewGame() *Game {
	return &Game{
		playerX: 100,
		playerY: screenHeight - groundHeight - 50,
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw sky (blue background)
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw sun
	g.drawSun(screen)

	// Draw clouds
	g.drawClouds(screen)

	// Draw ground
	g.drawGround(screen)
}

func (g *Game) drawSun(screen *ebiten.Image) {
	// Sun position (top right)
	sunX := float32(screenWidth - 80)
	sunY := float32(80)
	sunRadius := float32(40)

	// Outer glow (lighter yellow)
	vector.DrawFilledCircle(screen, sunX, sunY, sunRadius+10, color.RGBA{255, 255, 200, 100}, false)

	// Main sun body (bright yellow)
	vector.DrawFilledCircle(screen, sunX, sunY, sunRadius, color.RGBA{255, 255, 0, 255}, false)

	// Inner bright core
	vector.DrawFilledCircle(screen, sunX, sunY, sunRadius-10, color.RGBA{255, 255, 150, 255}, false)

	// Sun rays (8 rays around the sun)
	rayLength := float32(15)
	rayWidth := float32(3)
	for i := 0; i < 8; i++ {
		angle := float32(i) * 3.14159 / 4
		rayStartX := sunX + (sunRadius+5)*angle
		rayStartY := sunY + (sunRadius+5)*angle
		rayEndX := sunX + (sunRadius+rayLength)*angle
		rayEndY := sunY + (sunRadius+rayLength)*angle
		vector.StrokeLine(screen, rayStartX, rayStartY, rayEndX, rayEndY, rayWidth, color.RGBA{255, 255, 100, 200}, false)
	}
}

func (g *Game) drawClouds(screen *ebiten.Image) {
	// Cloud 1 (left)
	g.drawCloud(screen, 100, 80, 60)
	// Cloud 2 (center-left)
	g.drawCloud(screen, 300, 120, 50)
	// Cloud 3 (center-right)
	g.drawCloud(screen, 550, 60, 70)
	// Cloud 4 (right)
	g.drawCloud(screen, 700, 100, 45)
}

func (g *Game) drawCloud(screen *ebiten.Image, x, y, size int) {
	xFloat := float32(x)
	yFloat := float32(y)
	sizeFloat := float32(size)

	// Cloud color (white)
	cloudColor := color.RGBA{255, 255, 255, 255}

	// Main cloud body (multiple overlapping circles for fluffy look)
	// Bottom left puff
	vector.DrawFilledCircle(screen, xFloat-sizeFloat/3, yFloat+sizeFloat/6, sizeFloat/3, cloudColor, false)
	// Bottom center puff
	vector.DrawFilledCircle(screen, xFloat, yFloat+sizeFloat/4, sizeFloat/2.5, cloudColor, false)
	// Bottom right puff
	vector.DrawFilledCircle(screen, xFloat+sizeFloat/2, yFloat+sizeFloat/6, sizeFloat/3.5, cloudColor, false)
	// Top left puff
	vector.DrawFilledCircle(screen, xFloat-sizeFloat/6, yFloat-sizeFloat/6, sizeFloat/3, cloudColor, false)
	// Top center puff (largest)
	vector.DrawFilledCircle(screen, xFloat, yFloat-sizeFloat/8, sizeFloat/2, cloudColor, false)
	// Top right puff
	vector.DrawFilledCircle(screen, xFloat+sizeFloat/3, yFloat-sizeFloat/6, sizeFloat/3.5, cloudColor, false)
}

func (g *Game) drawGround(screen *ebiten.Image) {
	// Ground position
	groundY := float32(screenHeight - groundHeight)

	// Dirt layer (brown)
	dirtColor := color.RGBA{139, 69, 19, 255}
	vector.DrawFilledRect(screen, 0, groundY+20, screenWidth, groundHeight-20, dirtColor, false)

	// Grass layer (green) on top of dirt
	grassColor := color.RGBA{34, 139, 34, 255}
	vector.DrawFilledRect(screen, 0, groundY, screenWidth, 25, grassColor, false)

	// Grass details (individual grass blades)
	grassBladeColor := color.RGBA{0, 100, 0, 255}
	for x := 0; x < screenWidth; x += 15 {
		bladeHeight := float32(8 + (x % 5))
		vector.StrokeLine(screen,
			float32(x), groundY,
			float32(x), groundY-bladeHeight,
			2, grassBladeColor, false)
	}

	// Add some grass variation (lighter green patches)
	lightGrassColor := color.RGBA{100, 180, 100, 255}
	for x := 5; x < screenWidth; x += 40 {
		vector.DrawFilledRect(screen, float32(x), groundY+5, 20, 8, lightGrassColor, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Go Mario - 2D Platformer")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

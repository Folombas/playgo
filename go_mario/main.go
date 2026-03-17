package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	groundHeight = 100
)

type Game struct {
	playerX   float64
	playerY   float64
	frameCount int
}

func NewGame() *Game {
	return &Game{
		playerX:   100,
		playerY:   screenHeight - groundHeight - 50,
		frameCount: 0,
	}
}

func (g *Game) Update() error {
	g.frameCount++
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

	// Animated sun rays (16 rays with pulsing effect)
	// Rays rotate slowly and pulse in/out
	rayBaseLength := float32(20)
	pulseSpeed := 0.05
	
	for i := 0; i < 16; i++ {
		// Base angle for this ray
		baseAngle := float32(i) * 2 * math.Pi / 16
		
		// Add slow rotation
		rotationOffset := float32(g.frameCount) * 0.01
		angle := baseAngle + rotationOffset
		
		// Each ray pulses with a phase offset for wave effect
		rayPhase := float32(math.Sin(float64(g.frameCount)*pulseSpeed + float64(i)*0.4))
		rayLength := rayBaseLength + rayPhase*10
		
		// Ray width pulses too
		rayWidth := float32(2 + rayPhase*1.5)
		
		// Calculate ray start and end positions
		rayStartX := sunX + (sunRadius+3)*float32(math.Cos(float64(angle)))
		rayStartY := sunY + (sunRadius+3)*float32(math.Sin(float64(angle)))
		rayEndX := sunX + (sunRadius+rayLength)*float32(math.Cos(float64(angle)))
		rayEndY := sunY + (sunRadius+rayLength)*float32(math.Sin(float64(angle)))
		
		// Ray color with pulsing alpha
		alpha := uint8(150 + rayPhase*50)
		rayColor := color.RGBA{255, 255, 100, alpha}
		
		vector.StrokeLine(screen, float32(rayStartX), float32(rayStartY), float32(rayEndX), float32(rayEndY), rayWidth, rayColor, false)
	}
	
	// Inner rotating glow ring
	ringPhase := float32(g.frameCount) * 0.02
	for i := 0; i < 3; i++ {
		ringAngle := ringPhase + float32(i)*2*math.Pi/3
		ringX := sunX + (sunRadius-5)*float32(math.Cos(float64(ringAngle)))
		ringY := sunY + (sunRadius-5)*float32(math.Sin(float64(ringAngle)))
		vector.DrawFilledCircle(screen, float32(ringX), float32(ringY), 3, color.RGBA{255, 255, 200, 200}, false)
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

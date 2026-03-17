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

type Cloud struct {
	x     float32
	y     float32
	size  float32
	speed float32
}

type Apple struct {
	x      float32
	y      float32
	offset float32 // для анимации покачивания
}

type Tree struct {
	x      float32
	y      float32
	height float32
	apples []Apple
}

type Game struct {
	playerX    float64
	playerY    float64
	frameCount int
	clouds     []Cloud
	trees      []Tree
}

func NewGame() *Game {
	// Initialize clouds with random positions and speeds
	clouds := []Cloud{
		{x: 100, y: 80, size: 60, speed: 0.3},
		{x: 300, y: 120, size: 50, speed: 0.5},
		{x: 550, y: 60, size: 70, speed: 0.2},
		{x: 700, y: 100, size: 45, speed: 0.4},
	}
	
	// Initialize apple trees
	trees := []Tree{
		createTree(150, screenHeight-groundHeight, 120),
		createTree(400, screenHeight-groundHeight, 140),
		createTree(650, screenHeight-groundHeight, 130),
	}
	
	return &Game{
		playerX:    100,
		playerY:    screenHeight - groundHeight - 50,
		frameCount: 0,
		clouds:     clouds,
		trees:      trees,
	}
}

func createTree(x, y float32, height float32) Tree {
	// Calculate trunk height and canopy position
	trunkHeight := height * 0.6
	canopyY := y - trunkHeight - 10
	
	// Create apples at positions within the tree canopy
	apples := []Apple{
		{x: x - 25, y: canopyY + 20, offset: 0},
		{x: x + 30, y: canopyY + 15, offset: 0.5},
		{x: x, y: canopyY + 35, offset: 1.0},
		{x: x - 15, y: canopyY + 45, offset: 1.5},
		{x: x + 20, y: canopyY + 40, offset: 2.0},
	}
	
	return Tree{
		x:      x,
		y:      y,
		height: height,
		apples: apples,
	}
}

func (g *Game) Update() error {
	g.frameCount++
	
	// Update cloud positions
	for i := range g.clouds {
		g.clouds[i].x += g.clouds[i].speed
		
		// Wrap around when cloud goes off screen
		if g.clouds[i].x - g.clouds[i].size > screenWidth {
			g.clouds[i].x = -g.clouds[i].size
		}
	}
	
	// Update apple sway animation
	for i := range g.trees {
		for j := range g.trees[i].apples {
			g.trees[i].apples[j].offset = float32(g.frameCount)*0.02 + float32(j)*0.5
		}
	}
	
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw sky (blue background)
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw sun
	g.drawSun(screen)

	// Draw clouds
	g.drawClouds(screen)

	// Draw trees
	g.drawTrees(screen)

	// Draw ground
	g.drawGround(screen)
}

func (g *Game) drawSun(screen *ebiten.Image) {
	// Sun position (top right)
	sunX := float32(screenWidth - 80)
	sunY := float32(80)
	sunRadius := float32(40)

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
	// Draw all clouds from the clouds array
	for _, cloud := range g.clouds {
		g.drawCloud(screen, cloud.x, cloud.y, cloud.size)
	}
}

func (g *Game) drawCloud(screen *ebiten.Image, x, y, size float32) {
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

func (g *Game) drawTrees(screen *ebiten.Image) {
	for _, tree := range g.trees {
		g.drawTree(screen, tree)
	}
}

func (g *Game) drawTree(screen *ebiten.Image, tree Tree) {
	// Draw tree trunk
	g.drawTreeTrunk(screen, tree.x, tree.y, tree.height)
	
	// Draw tree canopy (foliage)
	g.drawTreeCanopy(screen, tree.x, tree.y, tree.height)
	
	// Draw apples
	for _, apple := range tree.apples {
		g.drawApple(screen, apple.x, apple.y, apple.offset)
	}
}

func (g *Game) drawTreeTrunk(screen *ebiten.Image, x, y, height float32) {
	trunkWidth := float32(20)
	trunkHeight := height * 0.6

	// Main trunk (brown) - extends from ground up
	trunkColor := color.RGBA{101, 67, 33, 255}
	vector.DrawFilledRect(screen, x-trunkWidth/2, y-trunkHeight, trunkWidth, trunkHeight, trunkColor, false)

	// Bark texture (darker lines)
	barkColor := color.RGBA{60, 40, 20, 255}
	for i := 0; i < 3; i++ {
		lineY := y - trunkHeight + float32(i)*8 + 5
		vector.StrokeLine(screen, x-trunkWidth/2+3, lineY, x+trunkWidth/2-3, lineY+2, 2, barkColor, false)
	}
}

func (g *Game) drawTreeCanopy(screen *ebiten.Image, x, y, height float32) {
	// Position canopy at top of trunk (not floating)
	trunkHeight := height * 0.6
	canopyY := y - trunkHeight - 10 // Sit on top of trunk with slight overlap
	canopyRadius := float32(50)

	// Main foliage (dark green)
	foliageColor := color.RGBA{34, 139, 34, 255}
	vector.DrawFilledCircle(screen, x, canopyY, canopyRadius, foliageColor, false)

	// Add depth with overlapping circles
	vector.DrawFilledCircle(screen, x-25, canopyY+10, canopyRadius-15, foliageColor, false)
	vector.DrawFilledCircle(screen, x+25, canopyY+10, canopyRadius-15, foliageColor, false)
	vector.DrawFilledCircle(screen, x, canopyY-15, canopyRadius-20, foliageColor, false)

	// Lighter green highlights for volume
	highlightColor := color.RGBA{100, 180, 100, 255}
	vector.DrawFilledCircle(screen, x-15, canopyY-10, 15, highlightColor, false)
	vector.DrawFilledCircle(screen, x+10, canopyY-20, 12, highlightColor, false)
}

func (g *Game) drawApple(screen *ebiten.Image, x, y, offset float32) {
	// Animate apple sway (gentle swinging)
	sway := float32(math.Sin(float64(offset))) * 2
	
	appleX := x + sway
	appleY := y
	
	appleRadius := float32(6)
	
	// Apple body (red)
	appleColor := color.RGBA{220, 20, 60, 255}
	vector.DrawFilledCircle(screen, appleX, appleY, appleRadius, appleColor, false)
	
	// Apple shine (lighter red highlight)
	highlightX := appleX - 2
	highlightY := appleY - 2
	vector.DrawFilledCircle(screen, highlightX, highlightY, appleRadius/2, color.RGBA{255, 100, 100, 255}, false)
	
	// Apple stem (brown)
	stemX := appleX
	stemY := appleY - appleRadius
	vector.StrokeLine(screen, stemX, stemY, stemX, stemY-4, 1.5, color.RGBA{139, 69, 19, 255}, false)
	
	// Green leaf
	leafX := stemX + 3
	leafY := stemY - 2
	leafColor := color.RGBA{34, 139, 34, 255}
	
	// Leaf shape (small oval)
	vector.DrawFilledCircle(screen, leafX, leafY, 3, leafColor, false)
	vector.DrawFilledCircle(screen, leafX+2, leafY+1, 2, leafColor, false)
	
	// Leaf vein (lighter green)
	vector.StrokeLine(screen, leafX, leafY, leafX+3, leafY+1, 0.5, color.RGBA{100, 200, 100, 255}, false)
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

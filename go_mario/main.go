package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	groundHeight = 100
	gravity      = 0.5
	jumpForce    = -12
	moveSpeed    = 4
)

type GameState int

const (
	Menu GameState = iota
	Playing
)

type TimeOfDay int

const (
	Day TimeOfDay = iota
	Night
)

type Weather int

const (
	Clear Weather = iota
	Stormy
)

type Raindrop struct {
	x      float32
	y      float32
	speed  float32
	length float32
}

type Lightning struct {
	active    bool
	timer     int
	branches  []LightningBranch
}

type LightningBranch struct {
	points []struct{ x, y float32 }
	width  float32
}

type Cloud struct {
	x     float32
	y     float32
	size  float32
	speed float32
}

type Star struct {
	x      float32
	y      float32
	size   float32
	twinkle float32
}

type Apple struct {
	x         float32
	y         float32
	offset    float32 // для анимации покачивания
	collected bool
}

type Tree struct {
	x      float32
	y      float32
	height float32
	apples []Apple
}

type Player struct {
	x         float64
	y         float64
	vy        float64 // vertical velocity
	width     float32
	height    float32
	onGround  bool
	score     int
	facing    int // -1 = left, 1 = right
	animFrame int
}

type Game struct {
	playerX    float64
	playerY    float64
	frameCount int
	clouds     []Cloud
	trees      []Tree
	player     Player
	state      GameState
	timeOfDay  TimeOfDay
	weather    Weather
	stars      []Star
	moonX      float32
	moonY      float32
	raindrops  []Raindrop
	lightning  Lightning
	stormClouds []Cloud
}

func NewGame() *Game {
	// Initialize clouds with random positions and speeds
	clouds := []Cloud{
		{x: 100, y: 80, size: 60, speed: 0.3},
		{x: 300, y: 120, size: 50, speed: 0.5},
		{x: 550, y: 60, size: 70, speed: 0.2},
		{x: 700, y: 100, size: 45, speed: 0.4},
	}

	// Initialize storm clouds (dark, for stormy weather)
	stormClouds := []Cloud{
		{x: 50, y: 30, size: 80, speed: 0.4},
		{x: 200, y: 50, size: 100, speed: 0.3},
		{x: 400, y: 20, size: 90, speed: 0.5},
		{x: 600, y: 40, size: 85, speed: 0.35},
		{x: 750, y: 25, size: 75, speed: 0.4},
	}

	// Initialize raindrops
	raindrops := make([]Raindrop, 300)
	for i := range raindrops {
		raindrops[i] = Raindrop{
			x:      float32(i%30) * 27,
			y:      float32(i%20) * 30,
			speed:  float32(i%5+10) + float32(i%3)*2,
			length: float32(i%10+10),
		}
	}

	// Initialize apple trees
	trees := []Tree{
		createTree(150, screenHeight-groundHeight, 120),
		createTree(400, screenHeight-groundHeight, 140),
		createTree(650, screenHeight-groundHeight, 130),
	}

	// Initialize stars for night sky
	stars := make([]Star, 100)
	for i := range stars {
		stars[i] = Star{
			x:       float32(i%20) * 40 + float32(i%7)*13,
			y:       float32(i/20) * 25 + float32(i%5)*7,
			size:    float32(i%3+1),
			twinkle: float32(i) * 0.1,
		}
	}

	// Initialize player (bunny)
	player := Player{
		x:         50,
		y:         float64(screenHeight - groundHeight - 40),
		vy:        0,
		width:     30,
		height:    40,
		onGround:  true,
		score:     0,
		facing:    1,
		animFrame: 0,
	}

	return &Game{
		playerX:    100,
		playerY:    screenHeight - groundHeight - 50,
		frameCount: 0,
		clouds:     clouds,
		stormClouds: stormClouds,
		trees:      trees,
		player:     player,
		state:      Menu,
		timeOfDay:  Day,
		weather:    Clear,
		stars:      stars,
		moonX:      100,
		moonY:      80,
		raindrops:  raindrops,
		lightning:  Lightning{active: false, timer: 0, branches: []LightningBranch{}},
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
	// Handle menu state
	if g.state == Menu {
		// Navigate time of day with Up/Down arrows
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			g.timeOfDay = Day
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			g.timeOfDay = Night
		}
		// Navigate weather with Left/Right arrows
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			g.weather = Clear
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			g.weather = Stormy
		}
		// Start game with Enter or Space
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = Playing
		}
		return nil
	}

	g.frameCount++

	// Update cloud positions
	for i := range g.clouds {
		g.clouds[i].x += g.clouds[i].speed

		// Wrap around when cloud goes off screen
		if g.clouds[i].x - g.clouds[i].size > screenWidth {
			g.clouds[i].x = -g.clouds[i].size
		}
	}

	// Update storm clouds
	for i := range g.stormClouds {
		g.stormClouds[i].x += g.stormClouds[i].speed
		if g.stormClouds[i].x - g.stormClouds[i].size > screenWidth {
			g.stormClouds[i].x = -g.stormClouds[i].size
		}
	}

	// Update raindrops
	for i := range g.raindrops {
		g.raindrops[i].y += g.raindrops[i].speed
		if g.raindrops[i].y > screenHeight {
			g.raindrops[i].y = -g.raindrops[i].length
			g.raindrops[i].x = float32(g.frameCount%30 + i%10) * 27
		}
	}

	// Update lightning
	if g.weather == Stormy {
		g.updateLightning()
	}

	// Update apple sway animation
	for i := range g.trees {
		for j := range g.trees[i].apples {
			g.trees[i].apples[j].offset = float32(g.frameCount)*0.02 + float32(j)*0.5
		}
	}

	// Player movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.x -= moveSpeed
		g.player.facing = -1
		g.player.animFrame++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.x += moveSpeed
		g.player.facing = 1
		g.player.animFrame++
	}

	// Jumping
	if (ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.player.onGround {
		g.player.vy = jumpForce
		g.player.onGround = false
	}

	// Apply gravity
	g.player.vy += gravity
	g.player.y += g.player.vy

	// Ground collision
	groundY := float64(screenHeight - groundHeight - int(g.player.height))
	if g.player.y >= groundY {
		g.player.y = groundY
		g.player.vy = 0
		g.player.onGround = true
	}

	// Screen boundaries
	if g.player.x < 0 {
		g.player.x = 0
	}
	if g.player.x > float64(screenWidth)-float64(g.player.width) {
		g.player.x = float64(screenWidth) - float64(g.player.width)
	}

	// Apple collection
	g.checkAppleCollection()

	return nil
}

func (g *Game) updateLightning() {
	if g.lightning.active {
		g.lightning.timer--
		if g.lightning.timer <= 0 {
			g.lightning.active = false
			g.lightning.branches = []LightningBranch{}
		}
	} else {
		// Random lightning strike (about every 3-8 seconds)
		if g.frameCount%180 == 0 && math.Sin(float64(g.frameCount)*0.01) > 0.3 {
			g.lightning.active = true
			g.lightning.timer = 10 // frames
			g.generateLightning()
		}
	}
}

func (g *Game) generateLightning() {
	// Create lightning bolt from sky
	startX := float32(math.Sin(float64(g.frameCount)*0.1)*screenWidth/2 + screenWidth/2)
	startY := float32(0)
	
	var points []struct{ x, y float32 }
	points = append(points, struct{ x, y float32 }{startX, startY})
	
	currentX := startX
	currentY := startY
	
	for currentY < screenHeight {
		currentY += float32(math.Sin(float64(g.frameCount)*0.2)*20 + 30)
		// Zigzag pattern
		offset := float32(math.Sin(float64(currentY)*0.1) * 40)
		currentX += offset
		points = append(points, struct{ x, y float32 }{currentX, currentY})
	}
	
	g.lightning.branches = []LightningBranch{
		{points: points, width: 3},
	}
}

func (g *Game) checkAppleCollection() {
	playerRect := struct {
		x, y, w, h float32
	}{
		x: float32(g.player.x),
		y: float32(g.player.y),
		w: g.player.width,
		h: g.player.height,
	}

	for i := range g.trees {
		for j := range g.trees[i].apples {
			if g.trees[i].apples[j].collected {
				continue
			}

			apple := &g.trees[i].apples[j]
			// Simple circle-rect collision
			appleCX := apple.x
			appleCY := apple.y

			// Check if apple is within player bounds
			if appleCX > playerRect.x && appleCX < playerRect.x+playerRect.w &&
				appleCY > playerRect.y && appleCY < playerRect.y+playerRect.h {
				apple.collected = true
				g.player.score++
			}
		}
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Draw dark background
	screen.Fill(color.RGBA{20, 20, 40, 255})

	// Title
	title := "GO MARIO"
	titleX := screenWidth/2 - len(title)*12
	ebitenutil.DebugPrintAt(screen, title, titleX, 80)

	// Subtitle
	subtitle := "A 2D Platformer Game"
	subX := screenWidth/2 - len(subtitle)*6
	ebitenutil.DebugPrintAt(screen, subtitle, subX, 115)

	// Time of day selection header
	header := "Select Time of Day"
	headerX := screenWidth/2 - len(header)*8
	ebitenutil.DebugPrintAt(screen, header, headerX, 170)

	// Day option
	dayText := "  [↑] DAY   - Sunny day with blue sky and clouds"
	dayX := screenWidth/2 - len(dayText)*6
	if g.timeOfDay == Day {
		dayText = ">> [↑] DAY   - Sunny day with blue sky and clouds <<"
		dayX = screenWidth/2 - len(dayText)*6
	}
	ebitenutil.DebugPrintAt(screen, dayText, dayX, 210)

	// Night option
	nightText := "  [↓] NIGHT - Starry sky with Milky Way and moon"
	nightX := screenWidth/2 - len(nightText)*6
	if g.timeOfDay == Night {
		nightText = ">> [↓] NIGHT - Starry sky with Milky Way and moon <<"
		nightX = screenWidth/2 - len(nightText)*6
	}
	ebitenutil.DebugPrintAt(screen, nightText, nightX, 245)

	// Weather selection header
	weatherHeader := "Select Weather"
	weatherHeaderX := screenWidth/2 - len(weatherHeader)*8
	ebitenutil.DebugPrintAt(screen, weatherHeader, weatherHeaderX, 295)

	// Clear weather option
	clearText := "  [←] CLEAR  - Clear sunny weather"
	clearX := screenWidth/2 - len(clearText)*6
	if g.weather == Clear {
		clearText = ">> [←] CLEAR  - Clear sunny weather <<"
		clearX = screenWidth/2 - len(clearText)*6
	}
	ebitenutil.DebugPrintAt(screen, clearText, clearX, 335)

	// Stormy weather option
	stormyText := "  [→] STORMY - Rain, thunder, and lightning"
	stormyX := screenWidth/2 - len(stormyText)*6
	if g.weather == Stormy {
		stormyText = ">> [→] STORMY - Rain, thunder, and lightning <<"
		stormyX = screenWidth/2 - len(stormyText)*6
	}
	ebitenutil.DebugPrintAt(screen, stormyText, stormyX, 370)

	// Start prompt
	startText := "Press ENTER or SPACE to Start"
	startX := screenWidth/2 - len(startText)*8
	ebitenutil.DebugPrintAt(screen, startText, startX, 440)

	// Controls info
	controls := []string{
		"",
		"Controls:",
		"Arrow Keys / WASD - Move",
		"Space / W / Up - Jump",
		"Collect apples from trees!",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-100, 490+i*22)
	}
}

func (g *Game) drawDaySky(screen *ebiten.Image) {
	// Blue daytime sky
	screen.Fill(color.RGBA{135, 206, 235, 255})
}

func (g *Game) drawStormySky(screen *ebiten.Image) {
	// Dark stormy sky gradient (gray-blue)
	screen.Fill(color.RGBA{50, 55, 70, 255})
}

func (g *Game) drawStormClouds(screen *ebiten.Image) {
	// Draw dark storm clouds
	for _, cloud := range g.stormClouds {
		g.drawStormCloud(screen, cloud)
	}
}

func (g *Game) drawStormCloud(screen *ebiten.Image, cloud Cloud) {
	xFloat := cloud.x
	yFloat := cloud.y
	sizeFloat := cloud.size

	// Storm cloud color (dark gray)
	cloudColor := color.RGBA{60, 60, 70, 255}

	// Main cloud body (multiple overlapping circles for fluffy look)
	vector.DrawFilledCircle(screen, xFloat-sizeFloat/3, yFloat+sizeFloat/6, sizeFloat/3, cloudColor, false)
	vector.DrawFilledCircle(screen, xFloat, yFloat+sizeFloat/4, sizeFloat/2.5, cloudColor, false)
	vector.DrawFilledCircle(screen, xFloat+sizeFloat/2, yFloat+sizeFloat/6, sizeFloat/3.5, cloudColor, false)
	vector.DrawFilledCircle(screen, xFloat-sizeFloat/6, yFloat-sizeFloat/6, sizeFloat/3, cloudColor, false)
	vector.DrawFilledCircle(screen, xFloat, yFloat-sizeFloat/8, sizeFloat/2, cloudColor, false)
	vector.DrawFilledCircle(screen, xFloat+sizeFloat/3, yFloat-sizeFloat/6, sizeFloat/3.5, cloudColor, false)

	// Lighter gray highlights for depth
	highlightColor := color.RGBA{80, 80, 90, 255}
	vector.DrawFilledCircle(screen, xFloat-10, yFloat-5, sizeFloat/4, highlightColor, false)
}

func (g *Game) drawRain(screen *ebiten.Image) {
	// Draw raindrops
	rainColor := color.RGBA{150, 170, 200, 150}
	for _, drop := range g.raindrops {
		vector.StrokeLine(screen,
			drop.x, drop.y,
			drop.x, drop.y+drop.length,
			1, rainColor, false)
	}
}

func (g *Game) drawLightning(screen *ebiten.Image) {
	if !g.lightning.active {
		return
	}

	// Flash effect - brighten entire screen
	flashAlpha := uint8(100 + g.lightning.timer*10)
	if flashAlpha > 255 {
		flashAlpha = 255
	}
	screen.Fill(color.RGBA{255, 255, 255, flashAlpha})

	// Draw lightning bolt
	for _, branch := range g.lightning.branches {
		if len(branch.points) < 2 {
			continue
		}

		// Outer glow (bright white)
		for i := 0; i < len(branch.points)-1; i++ {
			p1 := branch.points[i]
			p2 := branch.points[i+1]
			vector.StrokeLine(screen, p1.x, p1.y, p2.x, p2.y, branch.width+4, color.RGBA{255, 255, 255, 150}, false)
		}

		// Inner bright core (yellow-white)
		for i := 0; i < len(branch.points)-1; i++ {
			p1 := branch.points[i]
			p2 := branch.points[i+1]
			vector.StrokeLine(screen, p1.x, p1.y, p2.x, p2.y, branch.width, color.RGBA{255, 255, 200, 255}, false)
		}
	}
}

func (g *Game) drawNightSky(screen *ebiten.Image) {
	// Dark night sky gradient
	screen.Fill(color.RGBA{10, 10, 30, 255})

	// Draw stars
	for _, star := range g.stars {
		// Twinkle effect
		twinkle := float32(math.Sin(float64(g.frameCount)*0.1 + float64(star.twinkle))) * 50
		alpha := uint8(150 + twinkle)

		starColor := color.RGBA{255, 255, 255, alpha}
		vector.DrawFilledCircle(screen, star.x, star.y, star.size, starColor, false)
	}

	// Draw Milky Way (diagonal band of stars)
	for i := 0; i < 200; i++ {
		mx := int(float32(i*4+int(math.Sin(float64(i)*0.1)*50))) % screenWidth
		my := float32(i/3) + float32(math.Sin(float64(i)*0.05)*30)
		mAlpha := uint8(50 + math.Sin(float64(g.frameCount)*0.05+float64(i))*20)
		vector.DrawFilledCircle(screen, float32(mx), my, 1, color.RGBA{200, 200, 255, mAlpha}, false)
	}

	// Draw Moon
	g.drawMoon(screen)
}

func (g *Game) drawMoon(screen *ebiten.Image) {
	moonX := g.moonX
	moonY := g.moonY
	moonRadius := float32(35)

	// Moon glow (soft white)
	vector.DrawFilledCircle(screen, moonX, moonY, moonRadius+8, color.RGBA{255, 255, 240, 80}, false)

	// Main moon body (bright white)
	vector.DrawFilledCircle(screen, moonX, moonY, moonRadius, color.RGBA{255, 255, 240, 255}, false)

	// Moon craters (gray circles)
	craterColor := color.RGBA{220, 220, 220, 255}
	vector.DrawFilledCircle(screen, moonX-10, moonY-8, 6, craterColor, false)
	vector.DrawFilledCircle(screen, moonX+15, moonY-5, 8, craterColor, false)
	vector.DrawFilledCircle(screen, moonX-5, moonY+12, 5, craterColor, false)
	vector.DrawFilledCircle(screen, moonX+8, moonY+10, 7, craterColor, false)
	vector.DrawFilledCircle(screen, moonX-12, moonY+5, 4, craterColor, false)

	// Moon shadow (slight gray on one side for 3D effect)
	vector.DrawFilledCircle(screen, moonX+5, moonY-3, moonRadius-5, color.RGBA{240, 240, 240, 150}, false)
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw based on game state
	if g.state == Menu {
		g.drawMenu(screen)
		return
	}

	// Draw sky based on time of day and weather
	if g.weather == Stormy {
		g.drawStormySky(screen)
	} else if g.timeOfDay == Day {
		g.drawDaySky(screen)
	} else {
		g.drawNightSky(screen)
	}

	// Draw sun (day only, clear weather)
	if g.timeOfDay == Day && g.weather == Clear {
		g.drawSun(screen)
	}

	// Draw clouds (day only, clear weather)
	if g.timeOfDay == Day && g.weather == Clear {
		g.drawClouds(screen)
	}

	// Draw storm clouds (stormy weather)
	if g.weather == Stormy {
		g.drawStormClouds(screen)
	}

	// Draw trees
	g.drawTrees(screen)

	// Draw player (bunny)
	g.drawPlayer(screen)

	// Draw ground
	g.drawGround(screen)

	// Draw rain (stormy weather)
	if g.weather == Stormy {
		g.drawRain(screen)
		g.drawLightning(screen)
	}

	// Draw UI (score)
	g.drawUI(screen)
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
		g.drawApple(screen, apple.x, apple.y, apple.offset, apple.collected)
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

func (g *Game) drawApple(screen *ebiten.Image, x, y, offset float32, collected bool) {
	if collected {
		return // Don't draw collected apples
	}
	
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

func (g *Game) drawPlayer(screen *ebiten.Image) {
	x := float32(g.player.x)
	y := float32(g.player.y)
	w := g.player.width
	h := g.player.height

	// Bunny body (light gray/white)
	bodyColor := color.RGBA{240, 240, 240, 255}
	vector.DrawFilledRect(screen, x+5, y+15, w-10, h-15, bodyColor, false)

	// Bunny head (circle)
	headY := y + 10
	headX := x + w/2
	vector.DrawFilledCircle(screen, headX, headY, 12, bodyColor, false)

	// Bunny ears (long, pointing up)
	earColor := color.RGBA{240, 240, 240, 255}
	earInnerColor := color.RGBA{255, 180, 180, 255} // pink inner ear

	// Left ear
	leftEarX := headX - 4
	leftEarY := headY - 8
	vector.DrawFilledRect(screen, leftEarX-3, leftEarY-15, 6, 18, earColor, false)
	vector.DrawFilledRect(screen, leftEarX-1, leftEarY-12, 2, 10, earInnerColor, false)

	// Right ear
	rightEarX := headX + 4
	rightEarY := headY - 8
	vector.DrawFilledRect(screen, rightEarX-3, rightEarY-15, 6, 18, earColor, false)
	vector.DrawFilledRect(screen, rightEarX-1, rightEarY-12, 2, 10, earInnerColor, false)

	// Eyes (black with white highlight)
	eyeOffset := g.player.facing * 3
	leftEyeX := headX - 4 + float32(eyeOffset)
	rightEyeX := headX + 4 + float32(eyeOffset)
	eyeY := headY + 2

	// Eye whites
	vector.DrawFilledCircle(screen, leftEyeX, eyeY, 4, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, eyeY, 4, color.RGBA{255, 255, 255, 255}, false)

	// Pupils (black)
	vector.DrawFilledCircle(screen, leftEyeX+float32(eyeOffset), eyeY, 2, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX+float32(eyeOffset), eyeY, 2, color.RGBA{0, 0, 0, 255}, false)

	// Nose (pink triangle)
	noseX := headX + float32(g.player.facing*2)
	noseY := headY + 8
	vector.DrawFilledCircle(screen, noseX, noseY, 2, color.RGBA{255, 180, 180, 255}, false)

	// Legs (animated based on movement)
	legOffset := float32(math.Sin(float64(g.player.animFrame)*0.5)) * 5
	if !g.player.onGround {
		legOffset = 3 // jumping pose
	}

	// Back leg
	vector.DrawFilledCircle(screen, x+10-legOffset, y+h-5, 5, bodyColor, false)
	// Front leg
	vector.DrawFilledCircle(screen, x+w-10+legOffset, y+h-5, 5, bodyColor, false)

	// Tail (fluffy white ball)
	tailX := x + w - 8
	tailY := y + h/2 + 5
	vector.DrawFilledCircle(screen, tailX, tailY, 5, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Score display
	scoreText := "Apples: " + string(rune('0'+g.player.score))
	ebitenutil.DebugPrintAt(screen, scoreText, 10, 10)

	// Time of day indicator
	timeText := "Time: Day"
	if g.timeOfDay == Night {
		timeText = "Time: Night"
	}
	ebitenutil.DebugPrintAt(screen, timeText, 10, 25)

	// Weather indicator
	weatherText := "Weather: Clear"
	if g.weather == Stormy {
		weatherText = "Weather: Stormy"
	}
	ebitenutil.DebugPrintAt(screen, weatherText, 10, 40)

	// Controls hint
	controlsText := "Arrow Keys/WASD: Move | Space/W/Up: Jump"
	ebitenutil.DebugPrintAt(screen, controlsText, 10, screenHeight-25)
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

package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
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
	InsideHouse
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

type SmokeParticle struct {
	x      float32
	y      float32
	vx     float32
	vy     float32
	size   float32
	life   int
	maxLife int
}

type House struct {
	x         float32
	y         float32
	width     float32
	height    float32
	doorX     float32
	doorY     float32
	doorW     float32
	doorH     float32
	windowX   float32
	windowY   float32
	windowW   float32
	windowH   float32
	chimneyX  float32
	chimneyY  float32
	smoke     []SmokeParticle
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

// AudioSystem manages sound effects and music
type AudioSystem struct {
	jumpSound    *audio.Player
	collectSound *audio.Player
	enterSound   *audio.Player
	thunderSound *audio.Player
}

// NewAudioSystem creates and initializes the audio system
func NewAudioSystem() *AudioSystem {
	audioContext := audio.NewContext(44100)
	
	as := &AudioSystem{}
	
	// Generate simple sound effects programmatically
	as.jumpSound = as.generateJumpSound(audioContext)
	as.collectSound = as.generateCollectSound(audioContext)
	as.enterSound = as.generateEnterSound(audioContext)
	as.thunderSound = as.generateThunderSound(audioContext)
	
	return as
}

// generateJumpSound creates a jump sound effect
func (as *AudioSystem) generateJumpSound(ctx *audio.Context) *audio.Player {
	// Create a simple beep sound - rising frequency
	sampleRate := 44100
	duration := sampleRate / 10 // 0.1 seconds
	data := make([]byte, duration*2)
	
	for i := 0; i < duration; i++ {
		// Rising frequency sweep
		t := float64(i) / float64(sampleRate)
		freq := 300.0 + t*800.0
		value := int16(math.Sin(2*math.Pi*freq*t) * (1.0 - float64(i)/float64(duration)) * 32767 / 2)
		data[i*2] = byte(value)
		data[i*2+1] = byte(value >> 8)
	}
	
	player := ctx.NewPlayerFromBytes(data)
	return player
}

// generateCollectSound creates a coin collection sound effect
func (as *AudioSystem) generateCollectSound(ctx *audio.Context) *audio.Player {
	sampleRate := 44100
	duration := sampleRate / 8 // 0.125 seconds
	data := make([]byte, duration*2)
	
	for i := 0; i < duration; i++ {
		// High pitched "ding" sound
		t := float64(i) / float64(sampleRate)
		freq := 880.0 + math.Sin(t*50)*50 // Slight vibrato
		value := int16(math.Sin(2*math.Pi*freq*t) * (1.0 - float64(i)/float64(duration)) * 32767 / 3)
		data[i*2] = byte(value)
		data[i*2+1] = byte(value >> 8)
	}
	
	player := ctx.NewPlayerFromBytes(data)
	return player
}

// generateEnterSound creates a door entry sound effect
func (as *AudioSystem) generateEnterSound(ctx *audio.Context) *audio.Player {
	sampleRate := 44100
	duration := sampleRate / 5 // 0.2 seconds
	data := make([]byte, duration*2)
	
	for i := 0; i < duration; i++ {
		// Low to mid frequency sweep (whoosh)
		t := float64(i) / float64(sampleRate)
		freq := 200.0 + t*400.0
		value := int16(math.Sin(2*math.Pi*freq*t) * (1.0 - float64(i)/float64(duration)) * 32767 / 4)
		data[i*2] = byte(value)
		data[i*2+1] = byte(value >> 8)
	}
	
	player := ctx.NewPlayerFromBytes(data)
	return player
}

// generateThunderSound creates a thunder sound effect
func (as *AudioSystem) generateThunderSound(ctx *audio.Context) *audio.Player {
	sampleRate := 44100
	duration := sampleRate / 2 // 0.5 seconds
	data := make([]byte, duration*2)
	
	for i := 0; i < duration; i++ {
		// Low rumbling noise
		t := float64(i) / float64(sampleRate)
		// Combine multiple low frequencies for rumble effect
		value := float64(0)
		value += math.Sin(2*math.Pi*60*t) * 0.5
		value += math.Sin(2*math.Pi*80*t) * 0.3
		value += math.Sin(2*math.Pi*100*t) * 0.2
		value *= (1.0 - float64(i)/float64(duration)) * 32767 / 2
		data[i*2] = byte(int16(value))
		data[i*2+1] = byte(int16(value) >> 8)
	}
	
	player := ctx.NewPlayerFromBytes(data)
	return player
}

// PlayJump plays the jump sound effect
func (as *AudioSystem) PlayJump() {
	if as.jumpSound != nil {
		as.jumpSound.Rewind()
		as.jumpSound.Play()
	}
}

// PlayCollect plays the collect sound effect
func (as *AudioSystem) PlayCollect() {
	if as.collectSound != nil {
		as.collectSound.Rewind()
		as.collectSound.Play()
	}
}

// PlayEnter plays the enter sound effect
func (as *AudioSystem) PlayEnter() {
	if as.enterSound != nil {
		as.enterSound.Rewind()
		as.enterSound.Play()
	}
}

// PlayThunder plays the thunder sound effect
func (as *AudioSystem) PlayThunder() {
	if as.thunderSound != nil {
		as.thunderSound.Rewind()
		as.thunderSound.Play()
	}
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
	house      House
	audio      *AudioSystem
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

	// Initialize smoke particles for house chimney
	smoke := make([]SmokeParticle, 20)
	for i := range smoke {
		smoke[i] = SmokeParticle{
			x:      0,
			y:      0,
			vx:     float32(i%5-2) * 0.3,
			vy:     -float32(i%3+1) * 0.5,
			size:   float32(i%5+3),
			life:   i * 20,
			maxLife: 100,
		}
	}

	// Initialize house
	house := House{
		x:        520,
		y:        screenHeight - groundHeight,
		width:    120,
		height:   100,
		doorX:    560,
		doorY:    screenHeight - groundHeight,
		doorW:    40,
		doorH:    50,
		windowX:  590,
		windowY:  screenHeight - groundHeight - 50,
		windowW:  30,
		windowH:  40,
		chimneyX: 590,
		chimneyY: screenHeight - groundHeight - 80,
		smoke:    smoke,
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
		house:      house,
		audio:      NewAudioSystem(),
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

	// Handle inside house state
	if g.state == InsideHouse {
		// Exit house with ESC or inside door
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = Playing
			// Position player outside near door
			g.player.x = 580
			g.player.y = float64(screenHeight - groundHeight - 40)
		}
		
		// Player movement inside house
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
		
		// Check if player is near inside door (for exiting)
		g.checkInsideDoorExit()
		
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
		g.audio.PlayJump()
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

	// House entry detection
	if g.state == Playing {
		g.checkHouseEntry()
	}

	// Apple collection
	g.checkAppleCollection()

	return nil
}

func (g *Game) checkHouseEntry() {
	// Check if player is near the door
	playerRight := g.player.x + float64(g.player.width)
	playerBottom := g.player.y + float64(g.player.height)

	doorLeft := float64(g.house.doorX)
	doorBottom := float64(g.house.doorY)

	// Check if player is in front of door (within 40 pixels horizontally and standing on ground)
	horizontalDist := playerRight - doorLeft
	if horizontalDist < 0 {
		horizontalDist = -horizontalDist
	}

	// Player must be close to door horizontally and near the door vertically
	if horizontalDist < 40 && playerBottom > doorBottom-10 && playerBottom < doorBottom+10 {
		// Player is near door - check for enter key
		if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = InsideHouse
			// Position player inside house
			g.player.x = 400
			g.player.y = 450
			g.audio.PlayEnter()
		}
	}
}

func (g *Game) checkInsideDoorExit() {
	// Check if player is near the inside door (left side of room)
	playerLeft := g.player.x
	playerBottom := g.player.y + float64(g.player.height)

	insideDoorX := float64(100)
	insideDoorBottom := float64(500)

	// Check if player is in front of inside door
	horizontalDist := playerLeft - insideDoorX
	if horizontalDist < 0 {
		horizontalDist = -horizontalDist
	}

	if horizontalDist < 50 && playerBottom > insideDoorBottom-10 && playerBottom < insideDoorBottom+10 {
		// Player is near inside door - check for exit key
		if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = Playing
			// Position player outside near door
			g.player.x = 580
			g.player.y = float64(screenHeight - groundHeight - 40)
		}
	}
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
			g.audio.PlayThunder()
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
				g.audio.PlayCollect()
			}
		}
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

func (g *Game) drawHouse(screen *ebiten.Image) {
	h := g.house

	// Draw house walls (light beige)
	wallColor := color.RGBA{245, 230, 200, 255}
	vector.DrawFilledRect(screen, h.x, h.y-h.height, h.width, h.height, wallColor, false)

	// Draw gabled roof (red-brown) - triangle with peak at top
	roofColor := color.RGBA{139, 69, 50, 255}
	roofBaseY := h.y - h.height  // Top of walls
	roofPeakY := roofBaseY - 50  // Peak is 50 pixels above base
	
	// Draw filled triangle for roof
	for y := roofPeakY; y <= roofBaseY; y++ {
		progress := (y - roofPeakY) / 50.0
		xLeft := h.x + (h.width/2) - (h.width/2+10)*progress
		xRight := h.x + (h.width/2) + (h.width/2+10)*progress
		vector.StrokeLine(screen, xLeft, y, xRight, y, 1, roofColor, false)
	}

	// Draw chimney (dark gray)
	chimneyColor := color.RGBA{80, 80, 80, 255}
	vector.DrawFilledRect(screen, h.chimneyX, h.chimneyY, 20, 40, chimneyColor, false)
	// Chimney top
	vector.DrawFilledRect(screen, h.chimneyX-5, h.chimneyY, 30, 10, chimneyColor, false)

	// Draw smoke particles
	g.updateAndDrawSmoke(screen)

	// Draw door (dark brown)
	doorColor := color.RGBA{101, 67, 33, 255}
	vector.DrawFilledRect(screen, h.doorX, h.doorY-h.doorH, h.doorW, h.doorH, doorColor, false)
	// Door frame
	vector.StrokeRect(screen, h.doorX-2, h.doorY-h.doorH-2, h.doorW+4, h.doorH+4, 2, color.RGBA{60, 40, 20, 255}, false)
	// Doorknob (gold)
	vector.DrawFilledCircle(screen, h.doorX+h.doorW-8, h.doorY-h.doorH/2, 3, color.RGBA{255, 215, 0, 255}, false)

	// Draw window
	g.drawHouseWindow(screen, h.windowX, h.windowY, h.windowW, h.windowH)

	// Draw house foundation (gray stone)
	foundationColor := color.RGBA{120, 120, 120, 255}
	vector.DrawFilledRect(screen, h.x-5, h.y, h.width+10, 10, foundationColor, false)

	// Draw "Press E" hint if player is near door
	g.drawHouseEntryHint(screen)
}

func (g *Game) drawHouseEntryHint(screen *ebiten.Image) {
	playerRight := g.player.x + float64(g.player.width)
	playerBottom := g.player.y + float64(g.player.height)
	doorLeft := float64(g.house.doorX)
	doorBottom := float64(g.house.doorY)

	horizontalDist := playerRight - doorLeft
	if horizontalDist < 0 {
		horizontalDist = -horizontalDist
	}

	if horizontalDist < 60 && playerBottom > doorBottom-10 && playerBottom < doorBottom+10 {
		// Show hint above door
		hintText := "Press E"
		hintX := int(float64(g.house.doorX) + float64(g.house.doorW)/2) - len(hintText)*4
		hintY := int(g.house.doorY - g.house.doorH - 15)
		ebitenutil.DebugPrintAt(screen, hintText, hintX, hintY)
	}
}

func (g *Game) drawHouseWindow(screen *ebiten.Image, x, y, w, h float32) {
	// Window frame (white)
	frameColor := color.RGBA{255, 255, 255, 255}
	vector.DrawFilledRect(screen, x, y, w, h, frameColor, false)

	// Window glass (light blue)
	glassColor := color.RGBA{200, 230, 255, 200}
	vector.DrawFilledRect(screen, x+3, y+3, w-6, h-6, glassColor, false)

	// Window cross (brown)
	vector.StrokeLine(screen, x+w/2, y+3, x+w/2, y+h-3, 2, color.RGBA{139, 69, 19, 255}, false)
	vector.StrokeLine(screen, x+3, y+h/2, x+w-3, y+h/2, 2, color.RGBA{139, 69, 19, 255}, false)

	// Curtains (red with folds)
	curtainColor := color.RGBA{180, 50, 50, 255}
	// Left curtain
	vector.DrawFilledRect(screen, x+2, y+2, w/2-5, h-4, curtainColor, false)
	// Right curtain
	vector.DrawFilledRect(screen, x+w/2+3, y+2, w/2-5, h-4, curtainColor, false)
	// Curtain folds (darker lines)
	for i := 0; i < 4; i++ {
		foldY := y + 5 + float32(i)*8
		vector.StrokeLine(screen, x+5, foldY, x+w/2-8, foldY+3, 1, color.RGBA{150, 30, 30, 255}, false)
		vector.StrokeLine(screen, x+w/2+8, foldY, x+w-5, foldY+3, 1, color.RGBA{150, 30, 30, 255}, false)
	}
}

func (g *Game) updateAndDrawSmoke(screen *ebiten.Image) {
	h := g.house
	for i := range g.house.smoke {
		particle := &g.house.smoke[i]

		// Update particle
		particle.x = h.chimneyX + 10 + particle.vx*float32(g.frameCount%60)
		particle.y = h.chimneyY + particle.vy*float32(g.frameCount%60) - float32(particle.life)/3
		particle.life++

		if particle.life >= particle.maxLife {
			particle.life = 0
			particle.y = h.chimneyY
		}

		// Draw smoke (gray circles with decreasing alpha)
		alpha := uint8(150 - particle.life*150/particle.maxLife)
		smokeColor := color.RGBA{150, 150, 150, alpha}
		vector.DrawFilledCircle(screen, particle.x, particle.y, particle.size, smokeColor, false)
	}
}

func (g *Game) drawInsideHouse(screen *ebiten.Image) {
	// Interior walls (darker beige - bunny doesn't blend in)
	screen.Fill(color.RGBA{220, 210, 180, 255})

	// Floor (wooden planks - brown)
	floorColor := color.RGBA{139, 90, 50, 255}
	vector.DrawFilledRect(screen, 0, 500, screenWidth, screenHeight-500, floorColor, false)
	// Floor plank lines
	for x := 0; x < screenWidth; x += 40 {
		vector.StrokeLine(screen, float32(x), 500, float32(x), screenHeight, 2, color.RGBA{100, 60, 30, 255}, false)
	}

	// Ceiling (white)
	vector.DrawFilledRect(screen, 0, 0, screenWidth, 20, color.RGBA{255, 255, 255, 255}, false)

	// Draw inside door (on left wall)
	g.drawInsideDoor(screen)

	// Draw window with view outside
	g.drawInsideWindow(screen)

	// Draw furniture
	g.drawTableAndChair(screen)
	g.drawBed(screen)

	// Draw decorations
	g.drawChandelier(screen)
	g.drawCactus(screen)
	g.drawPortrait(screen)

	// Draw carrot basket on floor
	g.drawCarrotBasket(screen)

	// Draw player (bunny) inside house
	g.drawPlayer(screen)

	// Draw exit hints
	g.drawInsideDoorHint(screen)
	ebitenutil.DebugPrintAt(screen, "E / ESC - Exit house", 10, screenHeight-25)
}

func (g *Game) drawInsideDoorHint(screen *ebiten.Image) {
	playerLeft := g.player.x
	playerBottom := g.player.y + float64(g.player.height)

	insideDoorX := float64(130)
	insideDoorBottom := float64(500)

	horizontalDist := playerLeft - insideDoorX
	if horizontalDist < 0 {
		horizontalDist = -horizontalDist
	}

	if horizontalDist < 60 && playerBottom > insideDoorBottom-10 && playerBottom < insideDoorBottom+10 {
		hintText := "Press E"
		hintX := int(insideDoorX + 25) - len(hintText)*4
		hintY := int(insideDoorBottom) - 90
		ebitenutil.DebugPrintAt(screen, hintText, hintX, hintY)
	}
}

func (g *Game) drawInsideDoor(screen *ebiten.Image) {
	// Door frame (brown)
	doorX, doorY := float32(80), float32(420)
	doorW, doorH := float32(50), float32(80)

	// Door (dark brown)
	doorColor := color.RGBA{101, 67, 33, 255}
	vector.DrawFilledRect(screen, doorX, doorY, doorW, doorH, doorColor, false)

	// Door frame
	vector.StrokeRect(screen, doorX-3, doorY-3, doorW+6, doorH+6, 3, color.RGBA{60, 40, 20, 255}, false)

	// Doorknob (gold)
	vector.DrawFilledCircle(screen, doorX+doorW-10, doorY+doorH/2, 4, color.RGBA{255, 215, 0, 255}, false)

	// Door panels
	vector.StrokeLine(screen, doorX+10, doorY+10, doorX+10, doorY+doorH-10, 2, color.RGBA{60, 40, 20, 255}, false)
	vector.StrokeLine(screen, doorX+doorW-10, doorY+10, doorX+doorW-10, doorY+doorH-10, 2, color.RGBA{60, 40, 20, 255}, false)
}

func (g *Game) drawCarrotBasket(screen *ebiten.Image) {
	// Basket on floor
	basketX, basketY := float32(700), float32(480)

	// Basket body (woven brown)
	basketColor := color.RGBA{139, 90, 43, 255}
	vector.DrawFilledRect(screen, basketX, basketY, 40, 25, basketColor, false)

	// Basket weave pattern (darker lines)
	for i := 0; i < 5; i++ {
		vector.StrokeLine(screen, basketX+5, basketY+5+float32(i)*4, basketX+35, basketY+5+float32(i)*4, 1, color.RGBA{100, 60, 30, 255}, false)
	}

	// Carrots in basket (orange triangles)
	carrotColor := color.RGBA{255, 140, 0, 255}
	// Carrot 1
	vector.DrawFilledCircle(screen, basketX+10, basketY+8, 5, carrotColor, false)
	// Carrot 2
	vector.DrawFilledCircle(screen, basketX+20, basketY+10, 5, carrotColor, false)
	// Carrot 3
	vector.DrawFilledCircle(screen, basketX+30, basketY+8, 5, carrotColor, false)

	// Carrot greens (green tops)
	greenColor := color.RGBA{50, 150, 50, 255}
	vector.DrawFilledCircle(screen, basketX+10, basketY+4, 3, greenColor, false)
	vector.DrawFilledCircle(screen, basketX+20, basketY+5, 3, greenColor, false)
	vector.DrawFilledCircle(screen, basketX+30, basketY+4, 3, greenColor, false)
}

func (g *Game) drawInsideWindow(screen *ebiten.Image) {
	// Window frame (white)
	windowX, windowY, windowW, windowH := float32(100), float32(150), float32(150), float32(120)
	frameColor := color.RGBA{255, 255, 255, 255}
	vector.DrawFilledRect(screen, windowX, windowY, windowW, windowH, frameColor, false)

	// Window view (outdoor scene)
	// Sky (blue gradient)
	vector.DrawFilledRect(screen, windowX+5, windowY+5, windowW-10, windowH/2-5, color.RGBA{135, 206, 235, 255}, false)

	// Hills (green)
	hillColor := color.RGBA{50, 150, 50, 255}
	vector.DrawFilledCircle(screen, windowX+30, windowY+windowH/2, 40, hillColor, false)
	vector.DrawFilledCircle(screen, windowX+80, windowY+windowH/2, 50, hillColor, false)
	vector.DrawFilledCircle(screen, windowX+130, windowY+windowH/2, 35, hillColor, false)

	// River (blue)
	vector.DrawFilledRect(screen, windowX+5, windowY+windowH/2+20, windowW-10, 30, color.RGBA{70, 130, 180, 255}, false)

	// Apple trees outside (small)
	g.drawSmallTree(screen, windowX+25, windowY+windowH/2+10, 25)
	g.drawSmallTree(screen, windowX+110, windowY+windowH/2+5, 30)

	// Window glass
	glassColor := color.RGBA{200, 230, 255, 150}
	vector.DrawFilledRect(screen, windowX+5, windowY+5, windowW-10, windowH-10, glassColor, false)

	// Window cross (brown)
	vector.StrokeLine(screen, windowX+windowW/2, windowY+5, windowX+windowW/2, windowY+windowH-5, 3, color.RGBA{139, 69, 19, 255}, false)
	vector.StrokeLine(screen, windowX+5, windowY+windowH/2, windowX+windowW-5, windowY+windowH/2, 3, color.RGBA{139, 69, 19, 255}, false)

	// Curtains (red with folds)
	curtainColor := color.RGBA{180, 50, 50, 255}
	vector.DrawFilledRect(screen, windowX+3, windowY+3, windowW/2-8, windowH-6, curtainColor, false)
	vector.DrawFilledRect(screen, windowX+windowW/2+5, windowY+3, windowW/2-8, windowH-6, curtainColor, false)

	// Window sill (wood)
	vector.DrawFilledRect(screen, windowX-5, windowY+windowH-5, windowW+10, 10, color.RGBA{139, 69, 19, 255}, false)
}

func (g *Game) drawSmallTree(screen *ebiten.Image, x, y, size float32) {
	// Trunk
	vector.DrawFilledRect(screen, x-3, y, 6, size/2, color.RGBA{101, 67, 33, 255}, false)
	// Foliage (green circle)
	vector.DrawFilledCircle(screen, x, y, size/2, color.RGBA{34, 139, 34, 255}, false)
	// Apple (red dot)
	vector.DrawFilledCircle(screen, x+5, y+5, 3, color.RGBA{220, 20, 60, 255}, false)
}

func (g *Game) drawChandelier(screen *ebiten.Image) {
	// Chain from ceiling
	vector.StrokeLine(screen, 400, 20, 400, 60, 2, color.RGBA{100, 100, 100, 255}, false)

	// Main body (gold)
	vector.DrawFilledCircle(screen, 400, 70, 15, color.RGBA{255, 215, 0, 255}, false)

	// Hanging crystals (small circles)
	for i := 0; i < 6; i++ {
		angle := float32(i) * 3.14159 / 3
		crystalX := 400 + float32(math.Sin(float64(angle)))*20
		crystalY := 75 + float32(math.Cos(float64(angle)))*10
		vector.DrawFilledCircle(screen, crystalX, crystalY, 4, color.RGBA{255, 255, 255, 200}, false)
	}

	// Glow effect
	vector.DrawFilledCircle(screen, 400, 70, 25, color.RGBA{255, 255, 200, 50}, false)
}

func (g *Game) drawCactus(screen *ebiten.Image) {
	// Pot (brown)
	potX, potY := float32(175), float32(265)
	vector.DrawFilledRect(screen, potX, potY, 30, 20, color.RGBA{139, 69, 19, 255}, false)

	// Cactus body (green)
	vector.DrawFilledCircle(screen, potX+15, potY-10, 12, color.RGBA{34, 139, 34, 255}, false)
	vector.DrawFilledCircle(screen, potX+15, potY-25, 8, color.RGBA{34, 139, 34, 255}, false)
	// Side arm
	vector.DrawFilledCircle(screen, potX+25, potY-15, 6, color.RGBA{34, 139, 34, 255}, false)

	// Spikes (tiny white dots)
	for i := 0; i < 5; i++ {
		vector.DrawFilledCircle(screen, potX+10+float32(i)*3, potY-20, 1, color.RGBA{255, 255, 255, 255}, false)
	}
}

func (g *Game) drawTableAndChair(screen *ebiten.Image) {
	// Table (wooden)
	tableX, tableY := float32(500), float32(400)
	// Table top
	vector.DrawFilledRect(screen, tableX, tableY, 80, 10, color.RGBA{139, 69, 19, 255}, false)
	// Table legs
	vector.DrawFilledRect(screen, tableX+10, tableY+10, 8, 60, color.RGBA{101, 67, 33, 255}, false)
	vector.DrawFilledRect(screen, tableX+62, tableY+10, 8, 60, color.RGBA{101, 67, 33, 255}, false)

	// Chair (wooden)
	chairX, chairY := float32(600), float32(420)
	// Seat
	vector.DrawFilledRect(screen, chairX, chairY, 35, 8, color.RGBA{139, 69, 19, 255}, false)
	// Legs
	vector.DrawFilledRect(screen, chairX+5, chairY+8, 6, 50, color.RGBA{101, 67, 33, 255}, false)
	vector.DrawFilledRect(screen, chairX+24, chairY+8, 6, 50, color.RGBA{101, 67, 33, 255}, false)
	// Backrest
	vector.DrawFilledRect(screen, chairX, chairY-40, 8, 48, color.RGBA{101, 67, 33, 255}, false)
	vector.DrawFilledRect(screen, chairX+27, chairY-40, 8, 48, color.RGBA{101, 67, 33, 255}, false)
	vector.DrawFilledRect(screen, chairX, chairY-35, 35, 8, color.RGBA{139, 69, 19, 255}, false)
}

func (g *Game) drawBed(screen *ebiten.Image) {
	bedX, bedY := float32(250), float32(420)

	// Bed frame (brown wood)
	vector.DrawFilledRect(screen, bedX, bedY+40, 120, 15, color.RGBA{101, 67, 33, 255}, false)
	// Legs
	vector.DrawFilledRect(screen, bedX+10, bedY+55, 10, 15, color.RGBA{101, 67, 33, 255}, false)
	vector.DrawFilledRect(screen, bedX+100, bedY+55, 10, 15, color.RGBA{101, 67, 33, 255}, false)

	// Mattress (white)
	vector.DrawFilledRect(screen, bedX, bedY+25, 120, 20, color.RGBA{255, 255, 255, 255}, false)

	// Blanket (blue)
	vector.DrawFilledRect(screen, bedX+20, bedY+25, 80, 20, color.RGBA{70, 130, 180, 255}, false)

	// Pillow (white)
	vector.DrawFilledRect(screen, bedX+5, bedY+25, 30, 15, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) drawPortrait(screen *ebiten.Image) {
	// Frame (gold, ornate)
	frameX, frameY := float32(650), float32(200)
	frameW, frameH := float32(80), float32(100)

	// Outer frame
	vector.DrawFilledRect(screen, frameX, frameY, frameW, frameH, color.RGBA{255, 215, 0, 255}, false)
	// Inner frame (darker gold)
	vector.DrawFilledRect(screen, frameX+5, frameY+5, frameW-10, frameH-10, color.RGBA{200, 170, 0, 255}, false)

	// Portrait background (dark green)
	vector.DrawFilledRect(screen, frameX+8, frameY+8, frameW-16, frameH-16, color.RGBA{50, 80, 50, 255}, false)

	// Bunny silhouette (gray)
	// Head
	vector.DrawFilledCircle(screen, frameX+40, frameY+35, 15, color.RGBA{150, 150, 150, 255}, false)
	// Ears
	vector.DrawFilledRect(screen, frameX+35, frameY+15, 4, 20, color.RGBA{150, 150, 150, 255}, false)
	vector.DrawFilledRect(screen, frameX+41, frameY+15, 4, 20, color.RGBA{150, 150, 150, 255}, false)
	// Body
	vector.DrawFilledRect(screen, frameX+30, frameY+50, 20, 30, color.RGBA{150, 150, 150, 255}, false)

	// Frame decoration (small circles at corners)
	vector.DrawFilledCircle(screen, frameX+5, frameY+5, 3, color.RGBA{255, 215, 0, 255}, false)
	vector.DrawFilledCircle(screen, frameX+frameW-5, frameY+5, 3, color.RGBA{255, 215, 0, 255}, false)
	vector.DrawFilledCircle(screen, frameX+5, frameY+frameH-5, 3, color.RGBA{255, 215, 0, 255}, false)
	vector.DrawFilledCircle(screen, frameX+frameW-5, frameY+frameH-5, 3, color.RGBA{255, 215, 0, 255}, false)
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
		"E / Enter - Enter house",
		"ESC - Exit house",
		"Collect apples from trees!",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-100, 490+i*22)
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

	if g.state == InsideHouse {
		g.drawInsideHouse(screen)
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

	// Draw house
	g.drawHouse(screen)

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

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 64

	// Game constants
	MaxLives     = 10
	StartGold    = 150
	MaxWave      = 20

	// Tower types
	TowerArcher  = 1
	TowerCannon  = 2
	TowerMagic   = 3
	TowerIce     = 4  // Замедляет врагов
	TowerSniper  = 5  // Очень большая дальность

	// Enemy types
	EnemyKnight  = 1
	EnemyArcher  = 2
	EnemyGiant   = 3
	EnemyFast    = 4
	EnemyBoss    = 5

	// Game states
	StateMenu     = 0
	StatePlaying  = 1
	StateGameOver = 2
	StateWon      = 3

	// Sound types
	SoundShoot    = 0
	SoundHit      = 1
	SoundBuild    = 2
	SoundWave     = 3
	SoundGameOver = 4
	SoundWin      = 5
)

// ============================================================================
// GAME STRUCTURES
// ============================================================================

// Tower - башня защиты
type Tower struct {
	x, y      int
	towerType int
	range_    float64
	damage    int
	fireRate  int
	cooldown  int
	level     int
	color     color.RGBA
}

// Enemy - враг
type Enemy struct {
	x, y      float64
	pathIndex int
	t         int // tile position
	enemyType int
	hp        int
	maxHp     int
	speed     float64
	damage    int
	reward    int
	slowTimer int
}

// Projectile - снаряд
type Projectile struct {
	x, y      float64
	targetX   float64
	targetY   float64
	damage    int
	speed     float64
	alive     bool
	color     color.RGBA
	target    *Enemy
}

// Wave - волна врагов
type Wave struct {
	number    int
	enemies   int
	spawned   int
	spawnTime int
	completed bool
}

// Game - основная игра
type Game struct {
	tiles      [][]int
	towers     []*Tower
	enemies    []*Enemy
	projectiles []*Projectile
	waves      []*Wave
	currentWave int

	gold      int
	lives     int
	score     int
	state     int
	frameCount int

	// Assets
	tilesMap   map[int]*ebiten.Image
	towerImages map[int]*ebiten.Image
	enemyImages map[int]*ebiten.Image
	gameFont   font.Face

	// Audio
	audioCtx   *audio.Context
	sounds     map[int][]byte

	// Selection
	selectedTower int
	hoverX, hoverY int
}

// ============================================================================
// ASSETS
// ============================================================================

func LoadTiles() map[int]*ebiten.Image {
	tilesMap := make(map[int]*ebiten.Image)

	// Load medieval tiles
	tileFiles := []string{
		"medievalTile_001.png", // Grass
		"medievalTile_013.png", // Path
		"medievalTile_026.png", // Castle base
		"medievalTile_045.png", // Build spot
	}

	for i, file := range tileFiles {
		path := filepath.Join("assets", "tiles", file)
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err == nil {
			tilesMap[i] = img
		}
	}

	return tilesMap
}

func LoadTowerImages() map[int]*ebiten.Image {
	images := make(map[int]*ebiten.Image)
	
	// Create placeholder tower images
	for _, t := range []int{TowerArcher, TowerCannon, TowerMagic} {
		img := ebiten.NewImage(48, 48)
		switch t {
		case TowerArcher:
			vector.DrawFilledRect(img, 16, 8, 16, 32, color.RGBA{100, 200, 100, 255}, false)
			vector.DrawFilledCircle(img, 24, 16, 10, color.RGBA{150, 255, 150, 255}, false)
		case TowerCannon:
			vector.DrawFilledRect(img, 12, 16, 24, 24, color.RGBA{100, 100, 100, 255}, false)
			vector.DrawFilledCircle(img, 24, 20, 8, color.RGBA{50, 50, 50, 255}, false)
		case TowerMagic:
			vector.DrawFilledCircle(img, 24, 24, 18, color.RGBA{150, 50, 255, 255}, false)
			vector.DrawFilledCircle(img, 24, 24, 10, color.RGBA{200, 100, 255, 255}, false)
		}
		images[t] = img
	}

	return images
}

func LoadEnemyImages() map[int]*ebiten.Image {
	images := make(map[int]*ebiten.Image)

	// Try to load enemy sprites
	enemyFiles := map[int]string{
		EnemyKnight: "slimeWalk1.png",
		EnemyArcher: "flyFly1.png",
		EnemyGiant:  "blockerBody.png",
		EnemyFast:   "fishSwim1.png",
	}

	for t, file := range enemyFiles {
		path := filepath.Join("assets", "enemies", file)
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err == nil {
			images[t] = img
		}
	}

	// Create fallback images
	for t := range enemyFiles {
		if images[t] == nil {
			img := ebiten.NewImage(32, 32)
			c := color.RGBA{200, 50, 50, 255}
			switch t {
			case EnemyKnight:
				c = color.RGBA{180, 180, 180, 255} // Gray knight
			case EnemyArcher:
				c = color.RGBA{100, 200, 100, 255} // Green archer
			case EnemyGiant:
				c = color.RGBA{150, 100, 50, 255} // Brown giant
			case EnemyFast:
				c = color.RGBA{255, 200, 50, 255} // Yellow fast
			}
			vector.DrawFilledCircle(img, 16, 16, 14, c, false)
			images[t] = img
		}
	}

	return images
}

func LoadFont() font.Face {
	fontPath := filepath.Join("assets", "fonts", "SuperFeel-JpZqa.ttf")
	data, err := os.ReadFile(fontPath)
	if err != nil {
		return nil
	}

	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil
	}

	face, err := opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: 20,
		DPI:  72,
	})
	if err != nil {
		return nil
	}

	return face
}

// ============================================================================
// AUDIO SYSTEM
// ============================================================================

func InitAudio() *audio.Context {
	return audio.NewContext(44100)
}

func GenerateSound(frequency, duration float64, soundType string) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]byte, numSamples*2)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		var envelope float64 = 1.0 - float64(i)/float64(numSamples)
		var value float64

		switch soundType {
		case "shoot":
			value = math.Sin(2*math.Pi*frequency*t) * envelope * 0.3
		case "hit":
			value = math.Sin(2*math.Pi*frequency*2*t) * envelope * envelope * 0.4
		case "build":
			value = (math.Sin(2*math.Pi*frequency*t) + math.Sin(2*math.Pi*frequency*1.5*t)) * envelope * 0.2
		case "wave":
			value = math.Sin(2*math.Pi*440*t) * envelope * 0.3
		case "win":
			value = math.Sin(2*math.Pi*880*t) * envelope * 0.3
		default:
			value = math.Sin(2*math.Pi*frequency*t) * envelope * 0.3
		}

		sample := int16(value * 32767)
		samples[i*2] = byte(sample)
		samples[i*2+1] = byte(sample >> 8)
	}

	return samples
}

func LoadSounds() map[int][]byte {
	sounds := make(map[int][]byte)

	// Generate procedural sounds
	sounds[SoundShoot] = GenerateSound(600, 0.1, "shoot")
	sounds[SoundHit] = GenerateSound(200, 0.15, "hit")
	sounds[SoundBuild] = GenerateSound(800, 0.2, "build")
	sounds[SoundWave] = GenerateSound(440, 0.3, "wave")
	sounds[SoundGameOver] = GenerateSound(150, 0.5, "hit")
	sounds[SoundWin] = GenerateSound(880, 0.4, "win")

	return sounds
}

func PlaySound(g *Game, soundType int) {
	if g.audioCtx == nil {
		return
	}

	samples, ok := g.sounds[soundType]
	if !ok || len(samples) == 0 {
		return
	}

	player := g.audioCtx.NewPlayerFromBytes(samples)
	player.SetVolume(0.4)
	player.Play()
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		tiles:      make([][]int, ScreenWidth/TileSize),
		towers:     make([]*Tower, 0),
		enemies:    make([]*Enemy, 0),
		projectiles: make([]*Projectile, 0),
		waves:      make([]*Wave, 0),
		gold:       StartGold,
		lives:      MaxLives,
		state:      StateMenu,
		selectedTower: TowerArcher,
		audioCtx:   InitAudio(),
		sounds:     LoadSounds(),
	}

	// Initialize tiles
	for x := range g.tiles {
		g.tiles[x] = make([]int, ScreenHeight/TileSize)
	}

	// Load assets
	g.tilesMap = LoadTiles()
	g.towerImages = LoadTowerImages()
	g.enemyImages = LoadEnemyImages()
	g.gameFont = LoadFont()

	// Generate level
	g.GenerateLevel()

	// Create waves
	g.CreateWaves()

	return g
}

func (g *Game) GenerateLevel() {
	width := ScreenWidth / TileSize
	height := ScreenHeight / TileSize

	// Fill with grass
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			g.tiles[x][y] = 0 // Grass
		}
	}

	// Create path (simple winding path)
	path := [][2]int{
		{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2},
		{4, 3}, {4, 4}, {4, 5},
		{5, 5}, {6, 5}, {7, 5}, {8, 5},
		{8, 4}, {8, 3}, {8, 2}, {8, 1},
		{9, 1}, {10, 1}, {11, 1}, {12, 1}, {13, 1},
		{13, 2}, {13, 3}, {13, 4}, {13, 5}, {13, 6}, {13, 7}, {13, 8}, {13, 9},
		{14, 9}, {15, 9}, // Castle at end
	}

	for _, p := range path {
		if p[0] < width && p[1] < height {
			g.tiles[p[0]][p[1]] = 1 // Path
		}
	}

	// Mark buildable spots (areas adjacent to path but not on path)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if g.tiles[x][y] == 0 {
				// Check if adjacent to path
				for _, dx := range []int{-1, 0, 1} {
					for _, dy := range []int{-1, 0, 1} {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < width && ny >= 0 && ny < height {
							if g.tiles[nx][ny] == 1 {
								g.tiles[x][y] = 3 // Buildable
								break
							}
						}
					}
				}
			}
		}
	}

	// Castle at end
	if width > 0 && height > 0 {
		g.tiles[15][9] = 2 // Castle
	}
}

func (g *Game) CreateWaves() {
	for i := 1; i <= MaxWave; i++ {
		wave := &Wave{
			number:  i,
			enemies: 5 + i*2,
			spawnTime: 60 - min(30, i*2),
		}
		g.waves = append(g.waves, wave)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frameCount++

	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StateGameOver, StateWon:
		g.updateEndScreen()
	}

	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StatePlaying
		g.currentWave = 0
		g.gold = StartGold
		g.lives = MaxLives
		g.score = 0
		g.towers = make([]*Tower, 0)
		g.enemies = make([]*Enemy, 0)
		g.projectiles = make([]*Projectile, 0)
	}
}

func (g *Game) updatePlaying() {
	// Handle input
	g.handleInput()

	// Spawn enemies
	g.spawnEnemies()

	// Update enemies
	g.updateEnemies()

	// Update towers
	g.updateTowers()

	// Update projectiles
	g.updateProjectiles()

	// Check wave completion
	g.checkWaveProgress()

	// Check game over
	if g.lives <= 0 {
		g.state = StateGameOver
		PlaySound(g, SoundGameOver)
	}

	// Check win
	if g.currentWave >= len(g.waves) && len(g.enemies) == 0 {
		g.state = StateWon
		PlaySound(g, SoundWin)
	}
}

func (g *Game) handleInput() {
	// Mouse position
	mx, my := ebiten.CursorPosition()
	g.hoverX = mx
	g.hoverY = my

	// Tower selection
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) {
		g.selectedTower = TowerArcher
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) {
		g.selectedTower = TowerCannon
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) {
		g.selectedTower = TowerMagic
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit4) {
		g.selectedTower = TowerIce
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit5) {
		g.selectedTower = TowerSniper
	}

	// Place tower on left click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.tryPlaceTower(mx, my)
	}

	// Cancel on right click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.selectedTower = 0
	}
}

func (g *Game) tryPlaceTower(mx, my int) {
	if g.selectedTower == 0 {
		return
	}

	tx := mx / TileSize
	ty := my / TileSize

	if tx < 0 || tx >= ScreenWidth/TileSize || ty < 0 || ty >= ScreenHeight/TileSize {
		return
	}

	// Check if buildable
	if g.tiles[tx][ty] != 3 {
		return
	}

	// Check if tower already exists
	for _, t := range g.towers {
		if t.x == tx && t.y == ty {
			return
		}
	}

	// Check cost
	cost := g.getTowerCost(g.selectedTower)
	if g.gold < cost {
		return
	}

	// Place tower
	tower := &Tower{
		x: tx,
		y: ty,
		towerType: g.selectedTower,
		range_:    200,
		damage:    10,
		fireRate:  30,
		cooldown:  0,
		level:     1,
	}

	switch g.selectedTower {
	case TowerArcher:
		tower.range_ = 180
		tower.damage = 15
		tower.fireRate = 25
		tower.color = color.RGBA{100, 200, 100, 255}
	case TowerCannon:
		tower.range_ = 150
		tower.damage = 40
		tower.fireRate = 60
		tower.color = color.RGBA{100, 100, 100, 255}
	case TowerMagic:
		tower.range_ = 120
		tower.damage = 8
		tower.fireRate = 10
		tower.color = color.RGBA{150, 50, 255, 255}
	case TowerIce:
		tower.range_ = 140
		tower.damage = 6
		tower.fireRate = 20
		tower.color = color.RGBA{50, 150, 255, 255} // Blue ice
	case TowerSniper:
		tower.range_ = 350 // Very long range!
		tower.damage = 50
		tower.fireRate = 90
		tower.color = color.RGBA{255, 50, 50, 255} // Red sniper
	}

	g.towers = append(g.towers, tower)
	g.gold -= cost
	g.selectedTower = 0
	PlaySound(g, SoundBuild)
}

func (g *Game) getTowerCost(towerType int) int {
	switch towerType {
	case TowerArcher:
		return 50
	case TowerCannon:
		return 100
	case TowerMagic:
		return 150
	case TowerIce:
		return 120
	case TowerSniper:
		return 200
	}
	return 50
}

func (g *Game) spawnEnemies() {
	if g.currentWave >= len(g.waves) {
		return
	}

	wave := g.waves[g.currentWave]
	if wave.spawned >= wave.enemies {
		return
	}

	if g.frameCount%wave.spawnTime == 0 {
		wave.spawned++

		// Determine enemy type based on wave
		enemyType := EnemyKnight
		randVal := rand.Float32()

		if wave.number >= 15 && randVal < 0.1 {
			enemyType = EnemyBoss
		} else if wave.number >= 10 && randVal < 0.2 {
			enemyType = EnemyGiant
		} else if wave.number >= 5 && randVal < 0.3 {
			enemyType = EnemyFast
		} else if wave.number >= 3 && randVal < 0.4 {
			enemyType = EnemyArcher
		}

		hp := 20 + wave.number*5
		speed := 1.0
		reward := 10

		switch enemyType {
		case EnemyArcher:
			hp = hp / 2
			speed = 1.5
			reward = 15
		case EnemyGiant:
			hp = hp * 3
			speed = 0.5
			reward = 30
		case EnemyFast:
			hp = hp / 2
			speed = 2.0
			reward = 15
		case EnemyBoss:
			hp = hp * 5
			speed = 0.7
			reward = 100
		}

		enemy := &Enemy{
			x: 0,
			y: float64(2 * TileSize),
			enemyType: enemyType,
			hp: hp,
			maxHp: hp,
			speed: speed,
			damage: 1,
			reward: reward,
			pathIndex: 0,
		}

		g.enemies = append(g.enemies, enemy)
	}
}

func (g *Game) updateEnemies() {
	// Simple path following
	path := [][2]int{
		{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2},
		{4, 3}, {4, 4}, {4, 5},
		{5, 5}, {6, 5}, {7, 5}, {8, 5},
		{8, 4}, {8, 3}, {8, 2}, {8, 1},
		{9, 1}, {10, 1}, {11, 1}, {12, 1}, {13, 1},
		{13, 2}, {13, 3}, {13, 4}, {13, 5}, {13, 6}, {13, 7}, {13, 8}, {13, 9},
		{14, 9}, {15, 9},
	}

	for i := len(g.enemies) - 1; i >= 0; i-- {
		e := g.enemies[i]

		if e.pathIndex >= len(path) {
			// Reached castle
			g.lives -= e.damage
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			continue
		}

		target := path[e.pathIndex]
		targetX := float64(target[0] * TileSize)
		targetY := float64(target[1] * TileSize)

		dx := targetX - e.x
		dy := targetY - e.y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 5 {
			e.pathIndex++
		} else {
			e.x += (dx / dist) * e.speed
			e.y += (dy / dist) * e.speed
		}
	}
}

func (g *Game) updateTowers() {
	for _, tower := range g.towers {
		if tower.cooldown > 0 {
			tower.cooldown--
			continue
		}

		// Find target
		target := g.findTarget(tower)
		if target != nil {
			g.fireProjectile(tower, target)
			tower.cooldown = tower.fireRate
		}
	}
}

func (g *Game) findTarget(tower *Tower) *Enemy {
	var target *Enemy
	minDist := tower.range_

	for _, e := range g.enemies {
		dx := e.x - float64(tower.x*TileSize)
		dy := e.y - float64(tower.y*TileSize)
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < minDist {
			minDist = dist
			target = e
		}
	}

	return target
}

func (g *Game) fireProjectile(tower *Tower, target *Enemy) {
	proj := &Projectile{
		x: float64(tower.x*TileSize + TileSize/2),
		y: float64(tower.y*TileSize + TileSize/2),
		targetX: target.x,
		targetY: target.y,
		damage: tower.damage,
		speed: 8,
		alive: true,
		color: tower.color,
		target: target,
	}
	g.projectiles = append(g.projectiles, proj)
	PlaySound(g, SoundShoot)
}

func (g *Game) updateProjectiles() {
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		p := g.projectiles[i]

		if !p.alive {
			g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
			continue
		}

		// Move towards target
		dx := p.targetX - p.x
		dy := p.targetY - p.y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 10 {
			// Hit target
			p.alive = false
			if p.target != nil {
				p.target.hp -= p.damage
				if p.target.hp <= 0 {
					g.gold += p.target.reward
					g.score += p.target.reward * 10
					PlaySound(g, SoundHit)
				} else {
					// Hit sound (quieter)
					player := g.audioCtx.NewPlayerFromBytes(g.sounds[SoundHit])
					player.SetVolume(0.2)
					player.Play()
				}
			}
			continue
		}

		p.x += (dx / dist) * p.speed
		p.y += (dy / dist) * p.speed

		// Update target position
		if p.target != nil {
			p.targetX = p.target.x
			p.targetY = p.target.y
		}
	}

	// Remove dead enemies
	for i := len(g.enemies) - 1; i >= 0; i-- {
		if g.enemies[i].hp <= 0 {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
		}
	}
}

func (g *Game) checkWaveProgress() {
	if g.currentWave >= len(g.waves) {
		return
	}

	wave := g.waves[g.currentWave]
	if wave.spawned >= wave.enemies && len(g.enemies) == 0 {
		wave.completed = true
		g.currentWave++
		g.gold += 50 // Wave completion bonus
		PlaySound(g, SoundWave)
	}
}

func (g *Game) updateEndScreen() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StateMenu
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateWon:
		g.drawWon(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Background gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(50 + y/20)
		g := uint8(30 + y/30)
		b := uint8(80 + y/15)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	// Title
	title := "🏰 CASTLE DEFENDER"
	titleX := ScreenWidth/2 - 180
	titleY := ScreenHeight/3

	if g.gameFont != nil {
		text.Draw(screen, title, g.gameFont, titleX+3, titleY+3, color.RGBA{0, 0, 0, 150})
		text.Draw(screen, title, g.gameFont, titleX, titleY, color.RGBA{255, 215, 0, 255})
	} else {
		ebitenutil.DebugPrintAt(screen, title, titleX, titleY)
	}

	// Instructions
	instructions := []string{
		"🎮 Medieval Tower Defense",
		"",
		"Build towers to defend your castle!",
		"",
		"1 - Archer Tower (50g) - Fast, medium range",
		"2 - Cannon Tower (100g) - Slow, high damage",
		"3 - Magic Tower (150g) - Very fast, low damage",
		"4 - Ice Tower (120g) - Slows enemies",
		"5 - Sniper Tower (200g) - Extreme range",
		"",
		"Left Click - Place tower",
		"Right Click - Cancel",
		"",
		"Press ENTER or SPACE to Start",
	}

	y := ScreenHeight/2
	for _, line := range instructions {
		if g.gameFont != nil {
			c := color.RGBA{255, 255, 255, 255}
			if line == "Press ENTER or SPACE to Start" {
				c = color.RGBA{100, 255, 100, 255}
			}
			text.Draw(screen, line, g.gameFont, ScreenWidth/2-200, y, c)
		}
		y += 30
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Draw background
	screen.Fill(color.RGBA{34, 139, 34, 255})

	// Draw tiles
	g.drawTiles(screen)

	// Draw towers
	g.drawTowers(screen)

	// Draw enemies
	g.drawEnemies(screen)

	// Draw projectiles
	g.drawProjectiles(screen)

	// Draw UI
	g.drawUI(screen)

	// Draw placement preview
	if g.selectedTower != 0 {
		g.drawPlacementPreview(screen)
	}
}

func (g *Game) drawTiles(screen *ebiten.Image) {
	width := ScreenWidth / TileSize
	height := ScreenHeight / TileSize

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			tile := g.tiles[x][y]
			drawX := float64(x * TileSize)
			drawY := float64(y * TileSize)

			if img, ok := g.tilesMap[tile]; ok && img != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(drawX, drawY)
				screen.DrawImage(img, op)
			} else {
				// Fallback colors
				c := color.RGBA{34, 139, 34, 255} // Grass
				switch tile {
				case 1:
					c = color.RGBA{139, 119, 101, 255} // Path
				case 2:
					c = color.RGBA{100, 100, 150, 255} // Castle
				case 3:
					c = color.RGBA{50, 100, 50, 255} // Buildable
				}
				vector.DrawFilledRect(screen, float32(drawX), float32(drawY), TileSize, TileSize, c, false)
			}
		}
	}
}

func (g *Game) drawTowers(screen *ebiten.Image) {
	for _, tower := range g.towers {
		drawX := float64(tower.x*TileSize + TileSize/2)
		drawY := float64(tower.y*TileSize + TileSize/2)

		if img, ok := g.towerImages[tower.towerType]; ok && img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX-float64(img.Bounds().Dx())/2, drawY-float64(img.Bounds().Dy())/2)
			screen.DrawImage(img, op)
		} else {
			vector.DrawFilledCircle(screen, float32(drawX), float32(drawY), 20, tower.color, false)
		}

		// Draw range indicator on hover
		mx, my := ebiten.CursorPosition()
		if mx/TileSize == tower.x && my/TileSize == tower.y {
			vector.StrokeCircle(screen, float32(drawX), float32(drawY), float32(tower.range_), 2, color.RGBA{255, 255, 255, 100}, false)
		}
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	for _, e := range g.enemies {
		if img, ok := g.enemyImages[e.enemyType]; ok && img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(e.x-16, e.y-16)
			screen.DrawImage(img, op)
		} else {
			c := color.RGBA{200, 50, 50, 255}
			vector.DrawFilledCircle(screen, float32(e.x), float32(e.y), 14, c, false)
		}

		// Health bar
		hpPercent := float32(e.hp) / float32(e.maxHp)
		vector.DrawFilledRect(screen, float32(e.x)-15, float32(e.y)-25, 30, 4, color.RGBA{100, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(e.x)-15, float32(e.y)-25, 30*hpPercent, 4, color.RGBA{0, 255, 0, 255}, false)
	}
}

func (g *Game) drawProjectiles(screen *ebiten.Image) {
	for _, p := range g.projectiles {
		if !p.alive {
			continue
		}
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), 6, p.color, false)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 50, color.RGBA{0, 0, 0, 200}, false)

	if g.gameFont != nil {
		// Gold
		goldText := fmt.Sprintf("💰 Gold: %d", g.gold)
		text.Draw(screen, goldText, g.gameFont, 20, 32, color.RGBA{255, 215, 0, 255})

		// Lives
		livesText := fmt.Sprintf("❤️ Lives: %d", g.lives)
		text.Draw(screen, livesText, g.gameFont, 200, 32, color.RGBA{255, 100, 100, 255})

		// Wave
		waveText := fmt.Sprintf("🌊 Wave: %d/%d", g.currentWave+1, len(g.waves))
		text.Draw(screen, waveText, g.gameFont, 400, 32, color.RGBA{100, 200, 255, 255})

		// Score
		scoreText := fmt.Sprintf("⭐ Score: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, 650, 32, color.RGBA{255, 255, 255, 255})

		// Selected tower
		towerNames := map[int]string{
			TowerArcher: "🏹 Archer (50g)",
			TowerCannon: "💣 Cannon (100g)",
			TowerMagic:  "🔮 Magic (150g)",
		}
		if name, ok := towerNames[g.selectedTower]; ok {
			text.Draw(screen, name, g.gameFont, 850, 32, color.RGBA{150, 255, 150, 255})
		} else {
			text.Draw(screen, "Select tower: 1/2/3", g.gameFont, 850, 32, color.RGBA{200, 200, 200, 255})
		}
	} else {
		uiText := fmt.Sprintf("Gold: %d | Lives: %d | Wave: %d/%d | Score: %d",
			g.gold, g.lives, g.currentWave+1, len(g.waves), g.score)
		ebitenutil.DebugPrintAt(screen, uiText, 20, 20)
	}

	// Tower selection panel
	panelY := ScreenHeight - 80
	vector.DrawFilledRect(screen, 0, float32(panelY), ScreenWidth, 80, color.RGBA{0, 0, 0, 180}, false)

	towerInfo := []string{
		"[1] Archer: 50g - Fast",
		"[2] Cannon: 100g - Heavy",
		"[3] Magic: 150g - Rapid",
		"[4] Ice: 120g - Slow",
		"[5] Sniper: 200g - Range",
	}
	for i, info := range towerInfo {
		if g.gameFont != nil {
			c := color.RGBA{200, 200, 200, 255}
			if i+1 == g.selectedTower {
				c = color.RGBA{100, 255, 100, 255}
			}
			text.Draw(screen, info, g.gameFont, 20+i*200, panelY+25, c)
		}
	}
}

func (g *Game) drawPlacementPreview(screen *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	tx := mx / TileSize
	ty := my / TileSize

	if tx < 0 || tx >= ScreenWidth/TileSize || ty < 0 || ty >= ScreenHeight/TileSize {
		return
	}

	x := float32(tx * TileSize)
	y := float32(ty * TileSize)

	// Check if valid placement
	valid := g.tiles[tx][ty] == 3
	hasTower := false
	for _, t := range g.towers {
		if t.x == tx && t.y == ty {
			hasTower = true
			break
		}
	}

	c := color.RGBA{100, 255, 100, 150}
	if !valid || hasTower {
		c = color.RGBA{255, 100, 100, 150}
	}

	vector.DrawFilledRect(screen, x+4, y+4, TileSize-8, TileSize-8, c, false)
	vector.StrokeRect(screen, x, y, TileSize, TileSize, 2, c, false)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255})

	title := "💀 GAME OVER"
	if g.gameFont != nil {
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-120, ScreenHeight/2-30, color.RGBA{255, 50, 50, 255})

		scoreText := fmt.Sprintf("Final Score: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, ScreenWidth/2-100, ScreenHeight/2+20, color.RGBA{255, 255, 255, 255})

		restartText := "Press ENTER to return to menu"
		text.Draw(screen, restartText, g.gameFont, ScreenWidth/2-130, ScreenHeight/2+60, color.RGBA{200, 200, 200, 255})
	} else {
		ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-120, ScreenHeight/2-30)
	}
}

func (g *Game) drawWon(screen *ebiten.Image) {
	// Victory gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(50 + y/20)
		g := uint8(100 + y/15)
		b := uint8(50 + y/20)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	title := "🏆 VICTORY!"
	if g.gameFont != nil {
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-100, ScreenHeight/2-50, color.RGBA{255, 215, 0, 255})

		scoreText := fmt.Sprintf("Final Score: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, ScreenWidth/2-80, ScreenHeight/2, color.RGBA{255, 255, 255, 255})

		livesText := fmt.Sprintf("Lives Remaining: %d", g.lives)
		text.Draw(screen, livesText, g.gameFont, ScreenWidth/2-100, ScreenHeight/2+40, color.RGBA{100, 255, 100, 255})

		restartText := "Press ENTER to return to menu"
		text.Draw(screen, restartText, g.gameFont, ScreenWidth/2-130, ScreenHeight/2+80, color.RGBA{255, 255, 255, 255})
	} else {
		ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-100, ScreenHeight/2-50)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🏰 Castle Defender - Medieval Tower Defense | Go365 Day 84")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

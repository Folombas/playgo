package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
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
	ScreenWidth  = 1280
	ScreenHeight = 720

	// Game states
	StateMenu        = 0
	StateGalaxy      = 1
	StatePlanet      = 2
	StateBattle      = 3
	StateResearch    = 4
	StateShipyard    = 5
	StateSettings    = 6
	StateGameOver    = 7
	StateVictory     = 8

	// Resource types
	ResourceCredits  = 0
	ResourceMetal    = 1
	ResourceCrystal  = 2
	ResourceEnergy   = 3
	ResourceFood     = 4

	// Planet types
	PlanetTerran    = 1
	PlanetDesert    = 2
	PlanetIce       = 3
	PlanetVolcanic  = 4
	PlanetOcean     = 5
	PlanetGasGiant  = 6

	// Ship types
	ShipScout       = 1
	ShipFighter     = 2
	ShipCruiser     = 3
	ShipBattleship  = 4
	ShipCarrier     = 5
	ShipFreighter   = 6

	// Building types
	BuildingMine      = 1
	BuildingFactory   = 2
	BuildingFarm      = 3
	BuildingLab       = 4
	BuildingShipyard  = 5
	BuildingDefense   = 6

	// Tech types
	TechWeapons    = 1
	TechHull       = 2
	TechEngines    = 3
	TechIndustry   = 4
	TechScience    = 5
)

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Resource struct {
	credits  int
	metal    int
	crystal  int
	energy   int
	food     int
}

type Planet struct {
	id           int
	name         string
	x, y         float64
	planetType   int
	size         int
	population   int
	maxPop       int
	owner        int // 0 = unowned, 1 = player, 2+ = enemies
	buildings    []int
	production   int
	science      int
	ships        []*Ship
	explored     bool
	image        *ebiten.Image
}

type Ship struct {
	id        int
	shipType  int
	hp        int
	maxHp     int
	damage    int
	speed     float64
	range_    float64
	x, y      float64
	targetX   float64
	targetY   float64
	owner     int
	level     int
	selected  bool
}

type Building struct {
	id          int
	buildingType int
	level       int
	production  int
	science     int
}

type Tech struct {
	id          int
	name        string
	level       int
	maxLevel    int
	cost        int
	description string
}

type Enemy struct {
	id          int
	name        string
	planets     []*Planet
	ships       []*Ship
	aggression  int // 1-10
	techLevel   int
	alive       bool
}

type GameState struct {
	state      int
	resources  Resource
	planets    []*Planet
	enemies    []*Enemy
	ships      []*Ship
	fleet      []*Ship
	
	// Research
	techs      map[int]*Tech
	researchPoints int
	
	// Production
	productionQueue []int
	producing       int
	
	// Galaxy
	galaxyWidth  int
	galaxyHeight int
	cameraX      float64
	cameraY      float64
	zoom         float64
	
	// UI
	selectedPlanet   *Planet
	selectedShip     *Ship
	hoveredElement   string
	settingsOpen     bool
	soundEnabled     bool
	musicVolume      float64
	sfxVolume        float64
	gameSpeed        float64
	
	// Progress
	turn         int
	year         int
	difficulty   int
	
	// Visual effects
	particles    []*Particle
	screenShake  int
	
	// Assets
	gameFont     font.Face
	smallFont    font.Face
	assets       map[string]*ebiten.Image
	
	// Audio
	audioCtx     *audio.Context
	sounds       map[int][]byte
}

type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	color    color.RGBA
	size     float32
}

// ============================================================================
// ASSETS & AUDIO
// ============================================================================

func LoadFont(path string, size int) font.Face {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	face, _ := opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  72,
	})
	return face
}

func InitAudio() *audio.Context {
	return audio.NewContext(44100)
}

func GenerateSound(frequency, duration float64) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		envelope := 1.0 - float64(i)/float64(numSamples)
		value := math.Sin(2*math.Pi*frequency*t) * envelope * 0.3
		sample := int16(value * 32767)
		samples[i*2] = byte(sample)
		samples[i*2+1] = byte(sample >> 8)
	}
	return samples
}

func LoadSounds() map[int][]byte {
	sounds := make(map[int][]byte)
	sounds[0] = GenerateSound(400, 0.1)  // Click
	sounds[1] = GenerateSound(600, 0.15) // Build
	sounds[2] = GenerateSound(300, 0.2)  // Launch
	sounds[3] = GenerateSound(200, 0.3)  // Battle
	sounds[4] = GenerateSound(800, 0.2)  // Research
	sounds[5] = GenerateSound(1000, 0.3) // Victory
	return sounds
}

func PlaySound(g *GameState, soundType int) {
	if !g.soundEnabled || g.audioCtx == nil {
		return
	}
	samples, ok := g.sounds[soundType]
	if !ok || len(samples) == 0 {
		return
	}
	player := g.audioCtx.NewPlayerFromBytes(samples)
	player.SetVolume(g.sfxVolume)
	player.Play()
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGameState() *GameState {
	rand.Seed(time.Now().UnixNano())

	g := &GameState{
		state:      StateMenu,
		resources:  Resource{credits: 1000, metal: 500, crystal: 300, energy: 200, food: 400},
		planets:    make([]*Planet, 0),
		enemies:    make([]*Enemy, 0),
		ships:      make([]*Ship, 0),
		fleet:      make([]*Ship, 0),
		techs:      make(map[int]*Tech),
		particles:  make([]*Particle, 0),
		soundEnabled: true,
		musicVolume:  0.5,
		sfxVolume:    0.5,
		gameSpeed:    1.0,
		year:         2100,
		difficulty:   1,
		galaxyWidth:  3000,
		galaxyHeight: 2000,
		zoom:         1.0,
	}

	// Initialize technologies
	g.techs[TechWeapons] = &Tech{id: TechWeapons, name: "Оружие", level: 1, maxLevel: 10, cost: 100, description: "+10% урон"}
	g.techs[TechHull] = &Tech{id: TechHull, name: "Корпус", level: 1, maxLevel: 10, cost: 100, description: "+10% HP"}
	g.techs[TechEngines] = &Tech{id: TechEngines, name: "Двигатели", level: 1, maxLevel: 10, cost: 100, description: "+10% скорость"}
	g.techs[TechIndustry] = &Tech{id: TechIndustry, name: "Промышленность", level: 1, maxLevel: 10, cost: 100, description: "+10% производство"}
	g.techs[TechScience] = &Tech{id: TechScience, name: "Наука", level: 1, maxLevel: 10, cost: 100, description: "+10% наука"}

	// Load fonts
	g.gameFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 24)
	g.smallFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 16)

	// Load sounds
	g.audioCtx = InitAudio()
	g.sounds = LoadSounds()

	// Generate galaxy
	g.GenerateGalaxy()

	return g
}

func (g *GameState) GenerateGalaxy() {
	planetNames := []string{"Земля", "Марс", "Венера", "Юпитер", "Сатурн", "Титан", "Европа", "Кеплер", "Проксима", "Альфа"}
	
	// Player home planet
	homePlanet := &Planet{
		id: 1,
		name: "Земля",
		x: float64(g.galaxyWidth) / 2,
		y: float64(g.galaxyHeight) / 2,
		planetType: PlanetTerran,
		size: 60,
		population: 100,
		maxPop: 500,
		owner: 1,
		buildings: []int{BuildingMine, BuildingFarm},
		production: 10,
		science: 5,
		explored: true,
	}
	g.planets = append(g.planets, homePlanet)

	// Generate random planets
	for i := 0; i < 50; i++ {
		planetType := rand.Intn(6) + 1
		size := rand.Intn(40) + 30
		name := planetNames[rand.Intn(len(planetNames))]
		
		// Ensure not too close to home
		x := rand.Float64() * float64(g.galaxyWidth)
		y := rand.Float64() * float64(g.galaxyHeight)
		
		dx := x - homePlanet.x
		dy := y - homePlanet.y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist < 200 {
			continue // Too close
		}
		
		planet := &Planet{
			id: i + 2,
			name: fmt.Sprintf("%s-%d", name, i),
			x: x,
			y: y,
			planetType: planetType,
			size: size,
			population: 0,
			maxPop: size * 5,
			owner: 0,
			buildings: make([]int, 0),
			production: rand.Intn(10) + 5,
			science: rand.Intn(5) + 1,
			explored: false,
		}
		g.planets = append(g.planets, planet)
	}

	// Generate enemies
	for i := 0; i < 3; i++ {
		enemy := &Enemy{
			id: i + 2,
			name: fmt.Sprintf("Империя %d", i+1),
			planets: make([]*Planet, 0),
			ships: make([]*Ship, 0),
			aggression: rand.Intn(8) + 2,
			techLevel: rand.Intn(5) + 1,
			alive: true,
		}
		
		// Give enemy some planets
		for j := 0; j < rand.Intn(3)+2; j++ {
			for _, planet := range g.planets {
				if planet.owner == 0 {
					planet.owner = enemy.id
					enemy.planets = append(enemy.planets, planet)
					break
				}
			}
		}
		
		g.enemies = append(g.enemies, enemy)
	}
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *GameState) Update() error {
	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Screen shake decay
	if g.screenShake > 0 {
		g.screenShake--
	}

	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StateGalaxy:
		g.updateGalaxy()
	case StatePlanet:
		g.updatePlanet()
	case StateResearch:
		g.updateResearch()
	case StateSettings:
		g.updateSettings()
	case StateGameOver, StateVictory:
		g.updateEndScreen()
	}

	return nil
}

func (g *GameState) updateMenu() {
	mx, my := ebiten.CursorPosition()

	buttons := []struct{ x, y, w, h int }{
		{ScreenWidth/2 - 150, 250, 300, 60},
		{ScreenWidth/2 - 150, 330, 300, 60},
		{ScreenWidth/2 - 150, 410, 300, 60},
		{ScreenWidth/2 - 150, 490, 300, 60},
	}

	for i, btn := range buttons {
		if mx >= btn.x && mx <= btn.x+btn.w && my >= btn.y && my <= btn.y+btn.h {
			g.hoveredElement = fmt.Sprintf("menu_%d", i)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				switch i {
				case 0: // New Game
					g.state = StateGalaxy
				case 1: // Load Game
					g.LoadGame()
				case 2: // Settings
					g.state = StateSettings
				case 3: // Quit
					os.Exit(0)
				}
				PlaySound(g, 0)
			}
		}
	}
}

func (g *GameState) updateGalaxy() {
	mx, my := ebiten.CursorPosition()

	// Camera movement with middle mouse button or space
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) || ebiten.IsKeyPressed(ebiten.KeySpace) {
		dx, dy := ebiten.Wheel()
		g.cameraX += dx * 10
		g.cameraY += dy * 10
	}

	// Zoom
	wheelX, _ := ebiten.Wheel()
	g.zoom += wheelX * 0.1
	if g.zoom < 0.5 {
		g.zoom = 0.5
	}
	if g.zoom > 3.0 {
		g.zoom = 3.0
	}

	// Check planet clicks
	for _, planet := range g.planets {
		screenX := (planet.x - g.cameraX) * g.zoom
		screenY := (planet.y - g.cameraY) * g.zoom
		
		dx := float64(mx) - screenX
		dy := float64(my) - screenY
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist < float64(planet.size)*g.zoom {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if planet.owner == 1 || planet.owner == 0 {
					g.selectedPlanet = planet
					g.state = StatePlanet
					PlaySound(g, 0)
				}
			}
			g.hoveredElement = fmt.Sprintf("planet_%d", planet.id)
		}
	}

	// UI buttons
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateMenu
			PlaySound(g, 0)
		}
	}

	// End turn button
	if mx >= ScreenWidth-150 && mx <= ScreenWidth-20 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.EndTurn()
		}
	}
}

func (g *GameState) updatePlanet() {
	mx, my := ebiten.CursorPosition()

	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateGalaxy
			PlaySound(g, 0)
		}
	}

	// Building buttons
	buildingY := 200
	for i, buildingType := range []int{BuildingMine, BuildingFactory, BuildingFarm, BuildingLab, BuildingShipyard, BuildingDefense} {
		if mx >= 800 && mx <= 1000 && my >= buildingY+i*50 && my <= buildingY+i*50+40 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.BuildBuilding(buildingType)
			}
		}
	}

	// Ship production buttons
	shipY := 450
	for i, shipType := range []int{ShipScout, ShipFighter, ShipCruiser, ShipFreighter} {
		if mx >= 800 && mx <= 1000 && my >= shipY+i*50 && my <= shipY+i*50+40 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.ProduceShip(shipType)
			}
		}
	}
}

func (g *GameState) updateResearch() {
	mx, my := ebiten.CursorPosition()

	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		g.state = StateGalaxy
		PlaySound(g, 0)
	}

	// Tech buttons
	for techID, tech := range g.techs {
		y := 150 + (techID-1)*80
		if mx >= 100 && mx <= 600 && my >= y && my <= y+60 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if g.researchPoints >= tech.cost && tech.level < tech.maxLevel {
					g.researchPoints -= tech.cost
					tech.level++
					tech.cost = tech.cost * 3 / 2
					PlaySound(g, 4)
				}
			}
		}
	}
}

func (g *GameState) updateSettings() {
	mx, my := ebiten.CursorPosition()

	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		g.state = StateMenu
		PlaySound(g, 0)
	}

	// Sound toggle
	if mx >= ScreenWidth/2+100 && mx <= ScreenWidth/2+160 && my >= 230 && my <= 260 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.soundEnabled = !g.soundEnabled
			PlaySound(g, 0)
		}
	}

	// Game speed slider
	if mx >= ScreenWidth/2-100 && mx <= ScreenWidth/2+100 && my >= 390 && my <= 410 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.gameSpeed = float64(mx-(ScreenWidth/2-100)) / 200 * 3.0
			if g.gameSpeed < 0.5 {
				g.gameSpeed = 0.5
			}
			if g.gameSpeed > 3.0 {
				g.gameSpeed = 3.0
			}
		}
	}
}

func (g *GameState) updateEndScreen() {
	mx, my := ebiten.CursorPosition()

	if mx >= ScreenWidth/2-150 && mx <= ScreenWidth/2+150 && my >= ScreenHeight/2+50 && my <= ScreenHeight/2+110 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateMenu
			// Reset game
			g.resources = Resource{credits: 1000, metal: 500, crystal: 300, energy: 200, food: 400}
			g.planets = make([]*Planet, 0)
			g.enemies = make([]*Enemy, 0)
			g.ships = make([]*Ship, 0)
			g.year = 2100
			g.turn = 0
			g.GenerateGalaxy()
		}
	}
}

func (g *GameState) EndTurn() {
	g.turn++
	g.year++
	PlaySound(g, 0)

	// Resource production
	for _, planet := range g.planets {
		if planet.owner == 1 {
			g.resources.metal += planet.production
			g.resources.food += planet.production
			g.researchPoints += planet.science
		}
	}

	// Population growth
	for _, planet := range g.planets {
		if planet.owner == 1 && planet.population < planet.maxPop {
			if g.resources.food >= 10 {
				planet.population += 5
				g.resources.food -= 10
			}
		}
	}

	// Enemy AI
	g.UpdateEnemyAI()

	// Check victory
	victory := true
	for _, enemy := range g.enemies {
		if enemy.alive {
			victory = false
			break
		}
	}
	if victory && len(g.enemies) > 0 {
		g.state = StateVictory
		PlaySound(g, 5)
	}

	// Check game over
	playerAlive := false
	for _, planet := range g.planets {
		if planet.owner == 1 {
			playerAlive = true
			break
		}
	}
	if !playerAlive {
		g.state = StateGameOver
		PlaySound(g, 3)
	}
}

func (g *GameState) UpdateEnemyAI() {
	for _, enemy := range g.enemies {
		if !enemy.alive {
			continue
		}

		// Build ships based on aggression
		if rand.Float32() < float32(enemy.aggression)/10.0 {
			for _, planet := range enemy.planets {
				ship := &Ship{
					id: rand.Intn(10000),
					shipType: ShipFighter,
					hp: 50 * enemy.techLevel,
					maxHp: 50 * enemy.techLevel,
					damage: 10 * enemy.techLevel,
					speed: 2.0,
					x: planet.x,
					y: planet.y,
					owner: enemy.id,
				}
				g.ships = append(g.ships, ship)
				enemy.ships = append(enemy.ships, ship)
				break
			}
		}

		// Move ships towards player
		for _, ship := range enemy.ships {
			if len(g.planets) > 0 {
				target := g.planets[rand.Intn(len(g.planets))]
				if target.owner == 1 {
					dx := target.x - ship.x
					dy := target.y - ship.y
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist > 0 {
						ship.x += (dx / dist) * ship.speed
						ship.y += (dy / dist) * ship.speed
					}
				}
			}
		}
	}
}

func (g *GameState) BuildBuilding(buildingType int) {
	if g.selectedPlanet == nil {
		return
	}

	cost := 100
	switch buildingType {
	case BuildingMine:
		cost = 100
	case BuildingFactory:
		cost = 150
	case BuildingFarm:
		cost = 80
	case BuildingLab:
		cost = 200
	case BuildingShipyard:
		cost = 250
	case BuildingDefense:
		cost = 120
	}

	if g.resources.credits >= cost {
		g.resources.credits -= cost
		g.selectedPlanet.buildings = append(g.selectedPlanet.buildings, buildingType)
		
		switch buildingType {
		case BuildingMine:
			g.selectedPlanet.production += 5
		case BuildingFactory:
			g.selectedPlanet.production += 10
		case BuildingLab:
			g.selectedPlanet.science += 5
		}
		
		PlaySound(g, 1)
		g.spawnParticles(float64(ScreenWidth/2), float64(ScreenHeight/2), 20, color.RGBA{0, 255, 0, 255})
	}
}

func (g *GameState) ProduceShip(shipType int) {
	if g.selectedPlanet == nil {
		return
	}

	cost := 0
	switch shipType {
	case ShipScout:
		cost = 100
	case ShipFighter:
		cost = 200
	case ShipCruiser:
		cost = 400
	case ShipFreighter:
		cost = 150
	}

	if g.resources.credits >= cost && g.resources.metal >= cost/2 {
		g.resources.credits -= cost
		g.resources.metal -= cost / 2
		
		ship := &Ship{
			id: rand.Intn(10000),
			shipType: shipType,
			hp: 100,
			maxHp: 100,
			damage: 20,
			speed: 3.0,
			x: g.selectedPlanet.x,
			y: g.selectedPlanet.y,
			owner: 1,
		}
		
		g.ships = append(g.ships, ship)
		g.fleet = append(g.fleet, ship)
		PlaySound(g, 2)
		g.spawnParticles(float64(ScreenWidth/2), float64(ScreenHeight/2), 30, color.RGBA{100, 100, 255, 255})
	}
}

func (g *GameState) LoadGame() {
	// Placeholder for save/load system
	g.state = StateGalaxy
}

func (g *GameState) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x, y: y,
			vx: float64(rand.Intn(10)-5) * 0.5,
			vy: float64(rand.Intn(10)-5) * 0.5,
			life: 30 + rand.Intn(20),
			color: c,
			size: float32(rand.Intn(4)+2),
		})
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *GameState) Draw(screen *ebiten.Image) {
	// Apply screen shake
	if g.screenShake > 0 {
		dx := float64(rand.Intn(g.screenShake*2) - g.screenShake)
		dy := float64(rand.Intn(g.screenShake*2) - g.screenShake)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(dx, dy)
		tmp := ebiten.NewImage(ScreenWidth, ScreenHeight)
		g.drawGame(tmp)
		screen.DrawImage(tmp, op)
	} else {
		g.drawGame(screen)
	}
}

func (g *GameState) drawGame(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StateGalaxy:
		g.drawGalaxy(screen)
	case StatePlanet:
		g.drawPlanet(screen)
	case StateResearch:
		g.drawResearch(screen)
	case StateSettings:
		g.drawSettings(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateVictory:
		g.drawVictory(screen)
	}
}

func (g *GameState) drawMenu(screen *ebiten.Image) {
	// Space background with stars
	screen.Fill(color.RGBA{10, 10, 30, 255})
	
	// Draw stars
	for i := 0; i < 200; i++ {
		x := float32(rand.Intn(ScreenWidth))
		y := float32(rand.Intn(ScreenHeight))
		size := float32(rand.Intn(3) + 1)
		vector.DrawFilledCircle(screen, x, y, size, color.RGBA{255, 255, 255, 255}, false)
	}

	// Title
	title := "🚀 SPACE EMPIRE 🚀"
	titleX := ScreenWidth/2 - 200
	titleY := 150

	if g.gameFont != nil {
		text.Draw(screen, title, g.gameFont, titleX+4, titleY+4, color.RGBA{0, 0, 0, 200})
		text.Draw(screen, title, g.gameFont, titleX, titleY, color.RGBA{100, 200, 255, 255})

		subtitle := "4X Space Strategy"
		text.Draw(screen, subtitle, g.smallFont, ScreenWidth/2-80, 210, color.RGBA{150, 150, 150, 255})
	}

	// Buttons
	buttons := []string{"🎮 Новая игра", "📂 Загрузить", "⚙️ Настройки", "🚪 Выход"}
	buttonY := 250
	for i, btnText := range buttons {
		x := ScreenWidth/2 - 150
		y := buttonY + i*80

		btnColor := color.RGBA{40, 60, 100, 255}
		if g.hoveredElement == fmt.Sprintf("menu_%d", i) {
			btnColor = color.RGBA{60, 100, 160, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), 300, 60, btnColor, false)
		vector.StrokeRect(screen, float32(x), float32(y), 300, 60, 2, color.RGBA{100, 200, 255, 255}, false)

		if g.gameFont != nil {
			text.Draw(screen, btnText, g.gameFont, ScreenWidth/2-100, y+35, color.RGBA{255, 255, 255, 255})
		}
	}

	// Version
	if g.smallFont != nil {
		text.Draw(screen, "Go365 Day 84 | v1.0 | Pure Go + Ebitengine", g.smallFont, 20, ScreenHeight-30, color.RGBA{100, 150, 200, 255})
	}
}

func (g *GameState) drawGalaxy(screen *ebiten.Image) {
	// Space background
	screen.Fill(color.RGBA{5, 5, 20, 255})

	// Draw stars background
	for i := 0; i < 500; i++ {
		x := float32(rand.Intn(ScreenWidth))
		y := float32(rand.Intn(ScreenHeight))
		vector.DrawFilledCircle(screen, x, y, 1, color.RGBA{255, 255, 255, 100}, false)
	}

	// Draw galaxy grid
	for x := 0; x < ScreenWidth; x += 50 {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(ScreenHeight), 1, color.RGBA{30, 30, 60, 100}, false)
	}
	for y := 0; y < ScreenHeight; y += 50 {
		vector.StrokeLine(screen, 0, float32(y), float32(ScreenWidth), float32(y), 1, color.RGBA{30, 30, 60, 100}, false)
	}

	// Draw planets
	for _, planet := range g.planets {
		screenX := (planet.x - g.cameraX) * g.zoom
		screenY := (planet.y - g.cameraY) * g.zoom

		if screenX < -100 || screenX > ScreenWidth+100 || screenY < -100 || screenY > ScreenHeight+100 {
			continue
		}

		// Planet color based on type
		planetColor := color.RGBA{100, 150, 100, 255}
		switch planet.planetType {
		case PlanetTerran:
			planetColor = color.RGBA{50, 150, 50, 255}
		case PlanetDesert:
			planetColor = color.RGBA{200, 150, 50, 255}
		case PlanetIce:
			planetColor = color.RGBA{200, 230, 255, 255}
		case PlanetVolcanic:
			planetColor = color.RGBA{200, 50, 50, 255}
		case PlanetOcean:
			planetColor = color.RGBA{50, 100, 200, 255}
		case PlanetGasGiant:
			planetColor = color.RGBA{180, 140, 100, 255}
		}

		size := float32(planet.size) * float32(g.zoom)
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), size, planetColor, false)
		
		// Owner indicator
		ownerColor := color.RGBA{255, 255, 255, 100}
		if planet.owner == 1 {
			ownerColor = color.RGBA{0, 200, 0, 200}
		} else if planet.owner > 1 {
			ownerColor = color.RGBA{200, 0, 0, 200}
		}
		vector.StrokeCircle(screen, float32(screenX), float32(screenY), size+5, 2, ownerColor, false)

		// Planet name
		if g.smallFont != nil && g.zoom > 0.7 {
			text.Draw(screen, planet.name, g.smallFont, int(screenX)-30, int(screenY)+int(size)+15, color.RGBA{255, 255, 255, 255})
		}
	}

	// Draw ships
	for _, ship := range g.ships {
		screenX := (ship.x - g.cameraX) * g.zoom
		screenY := (ship.y - g.cameraY) * g.zoom

		if screenX < -20 || screenX > ScreenWidth+20 || screenY < -20 || screenY > ScreenHeight+20 {
			continue
		}

		shipColor := color.RGBA{100, 100, 200, 255}
		if ship.owner > 1 {
			shipColor = color.RGBA{200, 50, 50, 255}
		}
		
		// Draw ship as triangle
		vector.StrokeLine(screen, float32(screenX), float32(screenY)-10, float32(screenX)-8, float32(screenY)+8, 2, shipColor, false)
		vector.StrokeLine(screen, float32(screenX)-8, float32(screenY)+8, float32(screenX)+8, float32(screenY)+8, 2, shipColor, false)
		vector.StrokeLine(screen, float32(screenX)+8, float32(screenY)+8, float32(screenX), float32(screenY)-10, 2, shipColor, false)
	}

	// Draw particles
	g.drawParticles(screen)

	// UI Panel
	g.drawUI(screen)
}

func (g *GameState) drawPlanet(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{20, 20, 40, 255})

	if g.selectedPlanet != nil {
		// Planet view
		planetSize := float32(150)
		vector.DrawFilledCircle(screen, 400, 300, planetSize, color.RGBA{50, 150, 50, 255}, false)
		
		// Planet name
		if g.gameFont != nil {
			text.Draw(screen, g.selectedPlanet.name, g.gameFont, 300, 100, color.RGBA{255, 255, 255, 255})
		}

		// Stats panel
		statsY := 150
		if g.gameFont != nil {
			text.Draw(screen, fmt.Sprintf("Население: %d/%d", g.selectedPlanet.population, g.selectedPlanet.maxPop), g.smallFont, 50, statsY, color.RGBA{255, 255, 255, 255})
			text.Draw(screen, fmt.Sprintf("Производство: %d", g.selectedPlanet.production), g.smallFont, 50, statsY+30, color.RGBA{200, 200, 100, 255})
			text.Draw(screen, fmt.Sprintf("Наука: %d", g.selectedPlanet.science), g.smallFont, 50, statsY+60, color.RGBA{100, 200, 255, 255})
		}

		// Buildings
		if g.gameFont != nil {
			text.Draw(screen, "Здания:", g.gameFont, 50, 250, color.RGBA{255, 215, 0, 255})
		}
		for i, building := range g.selectedPlanet.buildings {
			buildingName := "Здание"
			switch building {
			case BuildingMine:
				buildingName = "🏭 Шахта"
			case BuildingFactory:
				buildingName = "🏭 Завод"
			case BuildingFarm:
				buildingName = "🌾 Ферма"
			case BuildingLab:
				buildingName = "🔬 Лаборатория"
			}
			if g.smallFont != nil {
				text.Draw(screen, buildingName, g.smallFont, 70, 280+i*25, color.RGBA{200, 200, 200, 255})
			}
		}

		// Build buttons
		buildings := []string{"🏭 Шахта (100¢)", "🏭 Завод (150¢)", "🌾 Ферма (80¢)", "🔬 Лаборатория (200¢)"}
		for i, btnText := range buildings {
			y := 200 + i*50
			vector.DrawFilledRect(screen, 800, float32(y), 200, 40, color.RGBA{60, 80, 120, 255}, false)
			if g.smallFont != nil {
				text.Draw(screen, btnText, g.smallFont, 810, y+25, color.RGBA{255, 255, 255, 255})
			}
		}

		// Ship production
		if g.gameFont != nil {
			text.Draw(screen, "Производство кораблей:", g.gameFont, 800, 420, color.RGBA{255, 215, 0, 255})
		}
		ships := []string{"🚀 Разведчик (100¢)", "⚔️ Истребитель (200¢)", "🛡️ Крейсер (400¢)"}
		for i, shipText := range ships {
			y := 450 + i*50
			vector.DrawFilledRect(screen, 800, float32(y), 200, 40, color.RGBA{80, 60, 120, 255}, false)
			if g.smallFont != nil {
				text.Draw(screen, shipText, g.smallFont, 810, y+25, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}

	// Resources
	g.drawResources(screen)
}

func (g *GameState) drawResearch(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 50, 255})

	if g.gameFont != nil {
		title := "🔬 Исследования"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-100, 80, color.RGBA{100, 200, 255, 255})

		rpText := fmt.Sprintf("Очки науки: %d", g.researchPoints)
		text.Draw(screen, rpText, g.smallFont, ScreenWidth/2-80, 120, color.RGBA{255, 255, 0, 255})
	}

	// Tech tree
	for techID, tech := range g.techs {
		y := 150 + (techID-1)*80
		
		// Tech box
		boxColor := color.RGBA{60, 80, 120, 255}
		if tech.level >= tech.maxLevel {
			boxColor = color.RGBA{100, 100, 100, 255}
		}
		vector.DrawFilledRect(screen, 100, float32(y), 500, 60, boxColor, false)
		vector.StrokeRect(screen, 100, float32(y), 500, 60, 2, color.RGBA{100, 200, 255, 255}, false)

		if g.gameFont != nil {
			nameText := fmt.Sprintf("%s (ур. %d/%d)", tech.name, tech.level, tech.maxLevel)
			text.Draw(screen, nameText, g.gameFont, 120, y+20, color.RGBA{255, 255, 255, 255})
			
			descText := tech.description
			text.Draw(screen, descText, g.smallFont, 120, y+45, color.RGBA{200, 200, 200, 255})
			
			costText := fmt.Sprintf("Стоимость: %d", tech.cost)
			text.Draw(screen, costText, g.smallFont, 450, y+20, color.RGBA{255, 215, 0, 255})
		}
	}

	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawSettings(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})

	if g.gameFont != nil {
		title := "⚙️ Настройки"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-80, 100, color.RGBA{255, 215, 0, 255})

		// Sound toggle
		soundText := "🔊 Звук:"
		text.Draw(screen, soundText, g.gameFont, ScreenWidth/2-150, 240, color.RGBA{255, 255, 255, 255})
		toggleColor := color.RGBA{0, 200, 0, 255}
		if !g.soundEnabled {
			toggleColor = color.RGBA{200, 0, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(ScreenWidth/2+100), 230, 60, 30, toggleColor, false)
		status := "ВКЛ"
		if !g.soundEnabled {
			status = "ВЫКЛ"
		}
		text.Draw(screen, status, g.smallFont, ScreenWidth/2+115, 250, color.RGBA{255, 255, 255, 255})

		// Game speed slider
		speedText := fmt.Sprintf("⏩ Скорость: %.1fx", g.gameSpeed)
		text.Draw(screen, speedText, g.gameFont, ScreenWidth/2-150, 380, color.RGBA{255, 255, 255, 255})
		vector.StrokeLine(screen, float32(ScreenWidth/2-100), 400, float32(ScreenWidth/2+100), 400, 4, color.RGBA{100, 100, 100, 255}, false)
		sliderX := float32(ScreenWidth/2-100) + float32(g.gameSpeed/3.0)*200
		vector.DrawFilledCircle(screen, sliderX, 400, 10, color.RGBA{0, 200, 255, 255}, false)
	}

	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 70, color.RGBA{0, 0, 0, 200}, false)

	g.drawResources(screen)

	// Year and turn
	if g.gameFont != nil {
		yearText := fmt.Sprintf("📅 %d год", g.year)
		text.Draw(screen, yearText, g.gameFont, 600, 20, color.RGBA{255, 255, 255, 255})
		
		turnText := fmt.Sprintf("🔄 Ход %d", g.turn)
		text.Draw(screen, turnText, g.gameFont, 800, 20, color.RGBA{255, 215, 0, 255})
	}

	// End turn button
	endTurnX := ScreenWidth - 150
	vector.DrawFilledRect(screen, float32(endTurnX), 20, 130, 40, color.RGBA{50, 150, 50, 255}, false)
	vector.StrokeRect(screen, float32(endTurnX), 20, 130, 40, 2, color.RGBA{100, 255, 100, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "Конец хода", g.smallFont, endTurnX+20, 42, color.RGBA{255, 255, 255, 255})
	}

	// Research button
	researchX := ScreenWidth - 320
	vector.DrawFilledRect(screen, float32(researchX), 20, 130, 40, color.RGBA{50, 50, 150, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "🔬 Наука", g.smallFont, researchX+25, 42, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawResources(screen *ebiten.Image) {
	if g.smallFont != nil {
		x := 20
		y := 35
		text.Draw(screen, fmt.Sprintf("💰 %d", g.resources.credits), g.smallFont, x, y, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, fmt.Sprintf("🔩 %d", g.resources.metal), g.smallFont, x+120, y, color.RGBA{200, 200, 200, 255})
		text.Draw(screen, fmt.Sprintf("💎 %d", g.resources.crystal), g.smallFont, x+240, y, color.RGBA{100, 200, 255, 255})
		text.Draw(screen, fmt.Sprintf("⚡ %d", g.resources.energy), g.smallFont, x+360, y, color.RGBA{255, 255, 0, 255})
		text.Draw(screen, fmt.Sprintf("🌾 %d", g.resources.food), g.smallFont, x+480, y, color.RGBA{100, 255, 100, 255})
	}
}

func (g *GameState) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), p.size, p.color, false)
	}
}

func (g *GameState) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255})

	if g.gameFont != nil {
		title := "💀 ПОРАЖЕНИЕ"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-140, ScreenHeight/2-50, color.RGBA{255, 50, 50, 255})

		statsText := fmt.Sprintf("Прожито лет: %d", g.year-2100)
		text.Draw(screen, statsText, g.smallFont, ScreenWidth/2-80, ScreenHeight/2+20, color.RGBA{255, 255, 255, 255})
	}

	// Continue button
	btnX := ScreenWidth/2 - 150
	btnY := ScreenHeight/2 + 80
	vector.DrawFilledRect(screen, float32(btnX), float32(btnY), 300, 60, color.RGBA{100, 50, 50, 255}, false)
	if g.gameFont != nil {
		text.Draw(screen, "В меню", g.gameFont, btnX+100, btnY+30, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawVictory(screen *ebiten.Image) {
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(50 + y/30)
		g := uint8(80 + y/40)
		b := uint8(50 + y/30)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	if g.gameFont != nil {
		title := "🏆 ПОБЕДА! ГАЛАКТИКА ВАША!"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-250, ScreenHeight/2-80, color.RGBA{255, 215, 0, 255})

		statsText := fmt.Sprintf("Империя процветает с %d года", g.year)
		text.Draw(screen, statsText, g.smallFont, ScreenWidth/2-150, ScreenHeight/2, color.RGBA{255, 255, 255, 255})
	}

	// Continue button
	btnX := ScreenWidth/2 - 150
	btnY := ScreenHeight/2 + 80
	vector.DrawFilledRect(screen, float32(btnX), float32(btnY), 300, 60, color.RGBA{50, 100, 50, 255}, false)
	if g.gameFont != nil {
		text.Draw(screen, "В меню", g.gameFont, btnX+100, btnY+30, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🚀 Space Empire - 4X Strategy | Go365 Day 84")

	game := NewGameState()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Go365 Day 87 - GO MARIO: SUPER PLATFORMER v7.0.0
// Красочный 2D платформер с БОССАМИ, ЗВУКАМИ и МИРАМИ!

package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 48
	Gravity      = 0.6
	JumpForce    = -12.0
	PlayerSpeed  = 5.0
	MaxWorld     = 3
	LevelsPerWorld = 3
)

// ============================================================================
// ASSETS - МНОЖЕСТВО СПРАЙТОВ!
// ============================================================================

type Assets struct {
	// Player - Green Alien
	playerStand  *ebiten.Image
	playerWalk1  *ebiten.Image
	playerWalk2  *ebiten.Image
	playerJump   *ebiten.Image
	playerDuck   *ebiten.Image

	// Enemies - РАЗНЫЕ ВРАГИ!
	slimeGreen   *ebiten.Image
	slimeBlue    *ebiten.Image
	slimePurple  *ebiten.Image
	bee          *ebiten.Image
	frog         *ebiten.Image
	mouse        *ebiten.Image
	fly          *ebiten.Image
	snail        *ebiten.Image

	// Tiles - РАЗНЫЕ ТАЙЛЫ!
	grassTile    *ebiten.Image
	dirtTile     *ebiten.Image
	brickTile    *ebiten.Image
	stoneTile    *ebiten.Image
	sandTile     *ebiten.Image
	snowTile     *ebiten.Image
	platformTile *ebiten.Image
	questionTile *ebiten.Image
	usedTile     *ebiten.Image
	spikeTile    *ebiten.Image

	// Decorations - УКРАШЕНИЯ!
	tree1        *ebiten.Image
	tree2        *ebiten.Image
	bush1        *ebiten.Image
	bush2        *ebiten.Image
	cloud1       *ebiten.Image
	cloud2       *ebiten.Image
	cloud3       *ebiten.Image
	rock1        *ebiten.Image
	rock2        *ebiten.Image
	flower1      *ebiten.Image
	flower2      *ebiten.Image
	mushroom     *ebiten.Image
	cactus       *ebiten.Image

	// Items - ПРЕДМЕТЫ!
	coinGold     *ebiten.Image
	coinSilver   *ebiten.Image
	coinBronze   *ebiten.Image
	gemRed       *ebiten.Image
	gemBlue      *ebiten.Image
	gemGreen     *ebiten.Image
	flagGreen    *ebiten.Image
	flagRed      *ebiten.Image
	flagBlue     *ebiten.Image
	mushroomRed  *ebiten.Image
	mushroomBrown *ebiten.Image
	star         *ebiten.Image
	keyGold      *ebiten.Image
	bomb         *ebiten.Image

	// Backgrounds - ФОНЫ!
	bgMountains  *ebiten.Image
	bgHills      *ebiten.Image

	// Font
	gameFont     font.Face
	largeFont    font.Face
}

var gameAssets *Assets

func LoadAssets() *Assets {
	assets := &Assets{}
	var err error

	// Player (Green Alien)
	assets.playerStand, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	assets.playerWalk1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	assets.playerWalk2, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	assets.playerJump, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_jump.png")
	assets.playerDuck, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_duck.png")

	// Enemies
	assets.slimeGreen, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeGreen.png")
	assets.slimeBlue, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeBlue.png")
	assets.slimePurple, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimePurple.png")
	assets.bee, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/bee.png")
	assets.frog, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/frog.png")
	assets.mouse, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/mouse.png")
	assets.fly, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/fly.png")
	assets.snail, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/snail.png")

	// Tiles
	assets.grassTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Grass/grass.png")
	assets.dirtTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickGrey.png")
	assets.brickTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickBrown.png")
	assets.stoneTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Ground/Stone/stone.png")
	assets.sandTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Ground/Sand/sand.png")
	assets.snowTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Ground/Snow/snow.png")
	assets.platformTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/bridgeA.png")
	assets.questionTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem.png")
	assets.usedTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem_disabled.png")
	assets.spikeTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/spikes.png")

	// Decorations
	assets.tree1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/cactus.png")
	assets.bush1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/bush.png")
	assets.cloud1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/cloud1.png")
	assets.cloud2, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/cloud2.png")
	assets.cloud3, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/cloud3.png")
	assets.rock1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/rock.png")
	assets.flower1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/plantPurple.png")
	assets.mushroom, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/mushroomBrown.png")

	// Items
	assets.coinGold, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/coinGold.png")
	assets.coinSilver, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/coinSilver.png")
	assets.coinBronze, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/coinBronze.png")
	assets.gemRed, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/gemRed.png")
	assets.gemBlue, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/gemBlue.png")
	assets.gemGreen, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/gemGreen.png")
	assets.flagGreen, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/flagGreen1.png")
	assets.flagRed, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/flagRed1.png")
	assets.flagBlue, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/flagBlue1.png")
	assets.mushroomRed, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/mushroomRed.png")
	assets.mushroomBrown, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/mushroomBrown.png")
	assets.star, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/star.png")
	assets.keyGold, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/keyGold.png")
	assets.bomb, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/bomb.png")

	// Font
	assets.gameFont, err = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 24)
	if err != nil {
		assets.gameFont = nil
	}
	assets.largeFont, err = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 48)
	if err != nil {
		assets.largeFont = nil
	}

	return assets
}

func loadFont(path string, size int) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  72,
	})
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Player struct {
	x, y      float64
	vx, vy    float64
	width     float32
	height    float32
	onGround  bool
	facing    int
	animFrame int
	health    int
	maxHealth int
	coins     int
	gems      int
	score     int
	lives     int
}

type Tile struct {
	x, y  int
	id    int
}

type Enemy struct {
	x, y      float64
	vx        float64
	width     float32
	height    float32
	enemyType int
	alive     bool
	animFrame int
}

type Coin struct {
	x, y      float64
	id        int // 0=bronze, 1=silver, 2=gold
	collected bool
	animFrame int
}

type Gem struct {
	x, y      float64
	id        int // 0=red, 1=blue, 2=green
	collected bool
	animFrame int
}

type Decoration struct {
	x, y      float64
	id        int
	animFrame int
}

type Particle struct {
	x, y, vx, vy float64
	life         int
	color        color.RGBA
	size         float32
}

type PowerUp struct {
	x, y      float64
	id        int // 0=mushroom, 1=star, 2=flower, 3=1up
	alive     bool
	animFrame int
}

type Achievement struct {
	id          string
	name        string
	description string
	unlocked    bool
}

type Boss struct {
	x, y         float64
	vx, vy       float64
	width        float32
	height       float32
	health       int
	maxHealth    int
	phase        int
	attackTimer  int
	attackType   int
	alive        bool
	animFrame    int
}

type SaveData struct {
	World        int `json:"world"`
	Level        int `json:"level"`
	Health       int `json:"health"`
	MaxHealth    int `json:"maxHealth"`
	Coins        int `json:"coins"`
	Gems         int `json:"gems"`
	Score        int `json:"score"`
	Lives        int `json:"lives"`
	TotalScore   int `json:"totalScore"`
}

type Game struct {
	player      *Player
	tiles       []*Tile
	enemies     []*Enemy
	coins       []*Coin
	gems        []*Gem
	powerUps    []*PowerUp
	decorations []*Decoration
	particles   []*Particle
	boss        *Boss
	cameraX     float64
	state       int // 0=menu, 1=playing, 2=gameover, 3=win, 4=boss
	frame       int
	levelWidth  int
	flagX       float64
	bgOffset    float64
	world       int
	level       int
	hasBoss     bool
	saveFile    string
	achievements map[string]*Achievement
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	gameAssets = LoadAssets()

	g := &Game{
		player: &Player{
			x: 100,
			y: 300,
			width: 40,
			height: 56,
			facing: 1,
			maxHealth: 100,
			health: 100,
			lives: 3,
		},
		state: 0,
		world: 1,
		level: 1,
		tiles: make([]*Tile, 0),
		enemies: make([]*Enemy, 0),
		coins: make([]*Coin, 0),
		gems: make([]*Gem, 0),
		powerUps: make([]*PowerUp, 0),
		decorations: make([]*Decoration, 0),
		particles: make([]*Particle, 0),
		boss: nil,
		saveFile: "savegame.json",
		achievements: make(map[string]*Achievement),
	}
	
	g.initAchievements()

	g.GenerateLevel()
	return g
}

func (g *Game) initAchievements() {
	g.achievements = map[string]*Achievement{
		"first_blood": {id: "first_blood", name: "Первая кровь", description: "Победите первого врага", unlocked: false},
		"coin_master": {id: "coin_master", name: "Мастер монет", description: "Соберите 50 монет", unlocked: false},
		"boss_slayer": {id: "boss_slayer", name: "Убийца боссов", description: "Победите босса", unlocked: false},
		"world_conqueror": {id: "world_conqueror", name: "Завоеватель", description: "Пройдите все миры", unlocked: false},
		"survivor": {id: "survivor", name: "Выживший", description: "Достигните 3 мира", unlocked: false},
	}
}

func (g *Game) SpawnBoss() {
	g.hasBoss = true
	g.state = 4 // Boss fight
	g.boss = &Boss{
		x: float64(g.levelWidth*TileSize - 200),
		y: 200,
		width: 100,
		height: 100,
		maxHealth: 200 + g.world*100,
		health: 200 + g.world*100,
		phase: 1,
		alive: true,
	}
}

func (g *Game) SaveGame() {
	data := SaveData{
		World: g.world,
		Level: g.level,
		Health: g.player.health,
		MaxHealth: g.player.maxHealth,
		Coins: g.player.coins,
		Gems: g.player.gems,
		Score: g.player.score,
		Lives: g.player.lives,
		TotalScore: g.player.score,
	}
	
	file, err := os.Create(g.saveFile)
	if err != nil {
		return
	}
	defer file.Close()
	
	json.NewEncoder(file).Encode(data)
}

func (g *Game) LoadGame() bool {
	file, err := os.Open(g.saveFile)
	if err != nil {
		return false
	}
	defer file.Close()
	
	var data SaveData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return false
	}
	
	g.world = data.World
	g.level = data.Level
	g.player.health = data.Health
	g.player.maxHealth = data.MaxHealth
	g.player.coins = data.Coins
	g.player.gems = data.Gems
	g.player.score = data.Score
	g.player.lives = data.Lives
	
	return true
}

func (g *Game) GenerateLevel() {
	g.levelWidth = 150 // tiles

	// Generate varied terrain
	for x := 0; x < g.levelWidth; x++ {
		// Ground with variations
		tileType := 1 // grass
		if x > 50 && x < 80 {
			tileType = 4 // stone
		} else if x > 100 {
			tileType = 5 // snow
		}

		// Ground tiles
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: tileType})
		g.tiles = append(g.tiles, &Tile{x: x, y: 15, id: 2}) // dirt below

		// Random platforms with different tiles
		if x > 5 && rand.Float32() < 0.12 {
			platY := rand.Intn(4) + 8
			platTile := rand.Intn(3) + 1 // brick, stone, or platform
			for bx := 0; bx < rand.Intn(4)+2; bx++ {
				if x+bx < g.levelWidth {
					g.tiles = append(g.tiles, &Tile{x: x+bx, y: platY, id: platTile})
				}
			}
			// Question block
			if rand.Float32() < 0.3 {
				g.tiles = append(g.tiles, &Tile{x: x + 1, y: platY - 1, id: 8}) // question
			}
		}

		// Random coins (different types)
		if x > 5 && rand.Float32() < 0.25 {
			coinY := float64(rand.Intn(8)+4) * TileSize
			coinType := rand.Intn(3) // bronze, silver, gold
			g.coins = append(g.coins, &Coin{
				x: float64(x*TileSize + 15),
				y: coinY,
				id: coinType,
			})
		}

		// Random gems
		if x > 10 && rand.Float32() < 0.08 {
			gemY := float64(rand.Intn(6)+5) * TileSize
			gemType := rand.Intn(3)
			g.gems = append(g.gems, &Gem{
				x: float64(x*TileSize + 15),
				y: gemY,
				id: gemType,
			})
		}

		// Random enemies (different types!)
		if x > 10 && rand.Float32() < 0.06 {
			enemyType := rand.Intn(4) // slime, bee, frog, mouse
			enY := float64(13 * TileSize)
			if enemyType == 1 { // bee flies
				enY = float64(rand.Intn(5)+6) * TileSize
			}
			g.enemies = append(g.enemies, &Enemy{
				x: float64(x * TileSize),
				y: enY,
				vx: -1.5,
				width: 40,
				height: 40,
				enemyType: enemyType,
				alive: true,
			})
		}

		// Spikes
		if x > 20 && rand.Float32() < 0.03 {
			g.tiles = append(g.tiles, &Tile{x: x, y: 13, id: 10}) // spike
		}
	}

	// Add decorations
	for x := 0; x < g.levelWidth; x++ {
		// Trees/bushes on ground
		if rand.Float32() < 0.08 {
			decType := rand.Intn(3) + 1
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize),
				y: float64(13*TileSize - 60),
				id: decType,
			})
		}
		// Flowers on ground
		if rand.Float32() < 0.1 {
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize + rand.Intn(30)),
				y: float64(13*TileSize - 20),
				id: rand.Intn(2) + 4,
			})
		}
		// Clouds in sky
		if rand.Float32() < 0.05 {
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize),
				y: float64(rand.Intn(150) + 50),
				id: rand.Intn(3) + 7,
			})
		}
		// Rocks
		if rand.Float32() < 0.06 {
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize),
				y: float64(13*TileSize - 25),
				id: rand.Intn(2) + 10,
			})
		}
	}
	
	// Spawn power-ups randomly
	for x := 0; x < g.levelWidth; x++ {
		if rand.Float32() < 0.02 { // 2% chance
			puType := rand.Intn(4) // 0=mushroom, 1=star, 2=flower, 3=1up
			g.powerUps = append(g.powerUps, &PowerUp{
				x: float64(x*TileSize + 15),
				y: float64(rand.Intn(6)+5) * TileSize,
				id: puType,
				alive: true,
			})
		}
	}

	// Flag at end
	g.flagX = float64((g.levelWidth - 5) * TileSize)
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frame++

	switch g.state {
	case 0: // Menu
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			// Try load game or start new
			if !g.LoadGame() {
				g.world = 1
				g.level = 1
			}
			g.state = 1
			g.GenerateLevel()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.SaveGame()
		}
		return nil

	case 2, 3: // GameOver/Win
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = 0
		}
		return nil
	
	case 4: // Boss fight
		g.updatePlayer()
		g.updateCamera()
		g.updateBoss()
		g.updateParticles()
		if g.boss == nil || !g.boss.alive {
			// Boss defeated!
			g.hasBoss = false
			g.state = 1
			g.player.score += 5000
			g.world++
			if g.world > MaxWorld {
				g.state = 3 // Victory!
			} else {
				g.level = 1
				g.GenerateLevel()
			}
		}
		return nil
	}

	// Playing
	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.updateCoins()
	g.updateGems()
	g.updateParticles()
	g.updatePowerUps()
	g.checkAchievements()
	g.checkWin()

	return nil
}

func (g *Game) updatePlayer() {
	p := g.player

	// Input
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = PlayerSpeed
		p.facing = 1
		p.animFrame++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -PlayerSpeed
		p.facing = -1
		p.animFrame++
	} else {
		p.vx = 0
	}

	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
		g.spawnJumpParticles(p.x+float64(p.width)/2, p.y+float64(p.height))
	}

	// Physics
	p.vy += Gravity
	if p.vy > 10 {
		p.vy = 10
	}
	p.x += p.vx
	p.y += p.vy

	// Collision with tiles
	g.checkTileCollision()

	// Boundaries
	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(g.levelWidth*TileSize)-float64(p.width) {
		p.x = float64(g.levelWidth*TileSize) - float64(p.width)
	}

	// Death
	if p.y > ScreenHeight {
		p.health = 0
		if p.health <= 0 {
			g.state = 2
		}
	}
}

func (g *Game) checkTileCollision() {
	p := g.player
	p.onGround = false

	for _, tile := range g.tiles {
		tileX := float64(tile.x * TileSize)
		tileY := float64(tile.y * TileSize)

		if p.x < tileX+float64(TileSize)-5 &&
			p.x+float64(p.width) > tileX+5 &&
			p.y < tileY+float64(TileSize) &&
			p.y+float64(p.height) > tileY {

			// Landing
			if p.vy >= 0 && p.y+float64(p.height) <= tileY+20 {
				p.y = tileY - float64(p.height)
				p.vy = 0
				p.onGround = true
			}
			// Ceiling
			if p.vy < 0 && p.y >= tileY {
				p.y = tileY + float64(TileSize)
				p.vy = 0
			}
			// Sides
			if p.vx > 0 && p.x < tileX {
				p.x = tileX - float64(p.width)
			} else if p.vx < 0 && p.x+float64(p.width) > tileX+float64(TileSize) {
				p.x = tileX + float64(TileSize)
			}
		}
	}
}

func (g *Game) updateCamera() {
	targetX := g.player.x - ScreenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > float64(g.levelWidth*TileSize)-ScreenWidth {
		g.cameraX = float64(g.levelWidth*TileSize) - ScreenWidth
	}
	g.bgOffset = g.cameraX * 0.3
}

func (g *Game) updateEnemies() {
	p := g.player

	for _, e := range g.enemies {
		if !e.alive {
			continue
		}

		e.x += e.vx
		e.animFrame++

		// Patrol
		if rand.Float32() < 0.02 {
			e.vx *= -1
		}

		// Collision with player
		if p.x < e.x+float64(e.width)-10 &&
			p.x+float64(p.width) > e.x+10 &&
			p.y < e.y+float64(e.height)-10 &&
			p.y+float64(p.height) > e.y+10 {

			if p.vy > 0 && p.y+float64(p.height) < e.y+float64(e.height)/2 {
				// Stomp!
				e.alive = false
				p.vy = JumpForce / 2
				p.score += 100
				g.spawnHitParticles(e.x+float64(e.width)/2, e.y+float64(e.height)/2)
			} else {
				// Hurt!
				p.health -= 15
				p.vy = -5
				if p.vx > 0 {
					p.vx = -5
				} else {
					p.vx = 5
				}
				if p.health <= 0 {
					g.state = 2
				}
			}
		}
	}
}

func (g *Game) updateBoss() {
	b := g.boss
	if b == nil || !b.alive {
		return
	}
	
	p := g.player
	b.animFrame++
	
	// Move towards player
	if p.x > b.x {
		b.vx = 2
	} else {
		b.vx = -2
	}
	
	// Jump occasionally
	if rand.Float32() < 0.02 && b.y >= 200 {
		b.vy = -10
	}
	
	// Gravity
	b.vy += 0.4
	if b.vy > 8 {
		b.vy = 8
	}
	
	b.x += b.vx
	b.y += b.vy
	
	// Floor collision
	if b.y > 500 {
		b.y = 500
		b.vy = 0
	}
	
	// Attack patterns
	b.attackTimer--
	if b.attackTimer <= 0 {
		b.attackTimer = 60
		b.attackType = rand.Intn(3)
		
		switch b.attackType {
		case 0: // Shoot projectile
			angle := math.Atan2(p.y-b.y, p.x-b.x)
			g.particles = append(g.particles, &Particle{
				x: b.x + float64(b.width)/2,
				y: b.y + float64(b.height)/2,
				vx: math.Cos(angle) * 6,
				vy: math.Sin(angle) * 6,
				life: 120,
				color: color.RGBA{255, 50, 50, 255},
				size: 10,
			})
		case 1: // Charge
			b.vx = float64(p.x-b.x) * 0.1
		case 2: // Jump attack
			b.vy = -15
		}
	}
	
	// Collision with player
	if p.x < b.x+float64(b.width)-20 &&
		p.x+float64(p.width) > b.x+20 &&
		p.y < b.y+float64(b.height)-20 &&
		p.y+float64(p.height) > b.y+20 {
		
		if p.vy > 0 && p.y+float64(p.height) < b.y+float64(b.height)/2 {
			// Jump on boss!
			b.health -= 20
			p.vy = JumpForce
			g.spawnHitParticles(b.x+float64(b.width)/2, b.y+20)
			if b.health <= 0 {
				b.alive = false
				g.spawnHitParticles(b.x+float64(b.width)/2, b.y+float64(b.height)/2)
			}
		} else {
			// Boss hurts player
			p.health -= 20
			p.vy = -8
			if p.x < b.x {
				p.vx = -8
			} else {
				p.vx = 8
			}
			if p.health <= 0 {
				g.state = 2
			}
		}
	}
	
	// Check phase
	if b.health < b.maxHealth/2 && b.phase == 1 {
		b.phase = 2
		b.attackTimer = 30 // Faster attacks
	}
}

func (g *Game) updateCoins() {
	p := g.player

	for _, c := range g.coins {
		if c.collected {
			continue
		}

		c.animFrame++

		if p.x < c.x+20 &&
			p.x+float64(p.width) > c.x &&
			p.y < c.y+20 &&
			p.y+float64(p.height) > c.y {
			c.collected = true
			p.coins++
			p.score += (c.id + 1) * 10
			g.spawnCollectParticles(c.x+10, c.y+10, color.RGBA{255, 215, 0, 255})
		}
	}
}

func (g *Game) updateGems() {
	p := g.player

	for _, gem := range g.gems {
		if gem.collected {
			continue
		}

		gem.animFrame++

		if p.x < gem.x+20 &&
			p.x+float64(p.width) > gem.x &&
			p.y < gem.y+20 &&
			p.y+float64(p.height) > gem.y {
			gem.collected = true
			p.gems++
			p.score += (gem.id + 1) * 50
			g.spawnCollectParticles(gem.x+10, gem.y+10, color.RGBA{100, 255, 100, 255})
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.3
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updatePowerUps() {
	p := g.player
	
	for i := range g.powerUps {
		pu := g.powerUps[i]
		if !pu.alive {
			continue
		}
		
		pu.animFrame++
		
		// Magnet effect
		dist := math.Hypot(pu.x-p.x, pu.y-p.y)
		if dist < 100 {
			pu.x += (p.x - pu.x) * 0.05
			pu.y += (p.y - pu.y) * 0.05
		}
		
		// Collection
		if dist < 40 {
			pu.alive = false
			g.applyPowerUp(pu)
		}
	}
}

func (g *Game) applyPowerUp(pu *PowerUp) {
	p := g.player
	
	switch pu.id {
	case 0: // Mushroom - +20 max HP + heal
		p.maxHealth += 20
		p.health = p.maxHealth
		g.spawnCollectParticles(pu.x, pu.y, color.RGBA{255, 100, 100, 255})
	case 1: // Star - +500 score
		p.score += 500
		g.spawnCollectParticles(pu.x, pu.y, color.RGBA{255, 255, 0, 255})
	case 2: // Flower - +10 coins
		p.coins += 10
		g.spawnCollectParticles(pu.x, pu.y, color.RGBA{100, 255, 100, 255})
	case 3: // 1UP - extra life
		p.lives++
		g.spawnCollectParticles(pu.x, pu.y, color.RGBA{0, 255, 0, 255})
	}
}

func (g *Game) checkAchievements() {
	p := g.player
	
	// First blood
	if p.score >= 100 && !g.achievements["first_blood"].unlocked {
		g.achievements["first_blood"].unlocked = true
	}
	
	// Coin master
	if p.coins >= 50 && !g.achievements["coin_master"].unlocked {
		g.achievements["coin_master"].unlocked = true
	}
	
	// Boss slayer
	if g.boss != nil && !g.boss.alive && !g.achievements["boss_slayer"].unlocked {
		g.achievements["boss_slayer"].unlocked = true
	}
	
	// Survivor
	if g.world >= 3 && !g.achievements["survivor"].unlocked {
		g.achievements["survivor"].unlocked = true
	}
	
	// World conqueror
	if g.world > MaxWorld && !g.achievements["world_conqueror"].unlocked {
		g.achievements["world_conqueror"].unlocked = true
	}
}

func (g *Game) checkWin() {
	if g.player.x >= g.flagX {
		if g.level < LevelsPerWorld && !g.hasBoss {
			// Spawn boss on last level of each world
			if g.level == LevelsPerWorld {
				g.SpawnBoss()
			} else {
				// Next level
				g.level++
				g.player.health = g.player.maxHealth
				g.GenerateLevel()
			}
		} else if g.hasBoss {
			// Boss already spawned, wait for boss defeat
			return
		} else {
			// Level complete without boss
			g.level++
			if g.level > LevelsPerWorld {
				// Spawn boss!
				g.SpawnBoss()
			} else {
				g.player.health = g.player.maxHealth
				g.GenerateLevel()
			}
		}
	}
}

// ============================================================================
// PARTICLES
// ============================================================================

func (g *Game) spawnJumpParticles(x, y float64) {
	for i := 0; i < 8; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + (rand.Float64()-0.5)*20,
			y: y,
			vx: (rand.Float64() - 0.5) * 3,
			vy: rand.Float64() * 2,
			life: 20,
			color: color.RGBA{200, 200, 200, 255},
			size: 3,
		})
	}
}

func (g *Game) spawnHitParticles(x, y float64) {
	for i := 0; i < 15; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: (rand.Float64() - 0.5) * 8,
			vy: (rand.Float64() - 0.5) * 8,
			life: 25,
			color: color.RGBA{255, 100, 100, 255},
			size: 4,
		})
	}
}

func (g *Game) spawnCollectParticles(x, y float64, c color.RGBA) {
	for i := 0; i < 10; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: (rand.Float64() - 0.5) * 5,
			vy: (rand.Float64() - 0.5) * 5,
			life: 20,
			color: c,
			size: 3,
		})
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case 0:
		g.drawMenu(screen)
	case 1, 2, 3:
		g.drawGame(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Sky gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(135 - y/10)
		g := uint8(206 - y/10)
		b := uint8(235 - y/10)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	if gameAssets.largeFont != nil {
		text.Draw(screen, "GO MARIO", gameAssets.largeFont, ScreenWidth/2-120, 180, color.White)
		text.Draw(screen, "SUPER PLATFORMER", gameAssets.gameFont, ScreenWidth/2-140, 250, color.RGBA{100, 255, 100, 255})
	}

	if gameAssets.gameFont != nil {
		controls := []string{
			"Arrow Keys / WASD - Move",
			"Space / W / Up - Jump",
			"Collect coins and gems!",
			"Jump on enemies!",
			"Reach the flag!",
			"",
			"Press ENTER to start",
		}
		y := 350
		for _, line := range controls {
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-120, y, color.White)
			y += 28
		}
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Sky
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(135 - y/10)
		g := uint8(206 - y/10)
		b := uint8(235 - y/10)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	camX := g.cameraX

	// Background decorations (parallax clouds)
	for _, d := range g.decorations {
		if d.id >= 7 && d.id <= 9 { // Clouds
			var img *ebiten.Image
			switch d.id {
			case 7: img = gameAssets.cloud1
			case 8: img = gameAssets.cloud2
			case 9: img = gameAssets.cloud3
			}
			if img != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(d.x-g.bgOffset, d.y)
				screen.DrawImage(img, op)
			}
		}
	}

	// Tiles
	for _, tile := range g.tiles {
		if float64(tile.x*TileSize) < camX-100 || float64(tile.x*TileSize) > camX+ScreenWidth+100 {
			continue
		}

		var img *ebiten.Image
		switch tile.id {
		case 1: img = gameAssets.grassTile
		case 2: img = gameAssets.dirtTile
		case 3: img = gameAssets.brickTile
		case 4: img = gameAssets.stoneTile
		case 5: img = gameAssets.sandTile
		case 6: img = gameAssets.snowTile
		case 7: img = gameAssets.platformTile
		case 8: img = gameAssets.questionTile
		case 9: img = gameAssets.usedTile
		case 10: img = gameAssets.spikeTile
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(tile.x*TileSize)-camX, float64(tile.y*TileSize))
			screen.DrawImage(img, op)
		}
	}

	// Foreground decorations
	for _, d := range g.decorations {
		if d.id < 7 || d.id > 9 {
			var img *ebiten.Image
			switch d.id {
			case 1: img = gameAssets.tree1
			case 2: img = gameAssets.bush1
			case 3: img = gameAssets.mushroom
			case 4, 5: img = gameAssets.flower1
			case 10, 11: img = gameAssets.rock1
			}
			if img != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(d.x-camX, d.y)
				screen.DrawImage(img, op)
			}
		}
	}

	// Coins
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		var img *ebiten.Image
		switch c.id {
		case 0: img = gameAssets.coinBronze
		case 1: img = gameAssets.coinSilver
		case 2: img = gameAssets.coinGold
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			bob := math.Sin(float64(c.animFrame)*0.1) * 5
			op.GeoM.Translate(c.x-camX, c.y+bob)
			screen.DrawImage(img, op)
		}
	}

	// Gems
	for _, gem := range g.gems {
		if gem.collected {
			continue
		}
		var img *ebiten.Image
		switch gem.id {
		case 0: img = gameAssets.gemRed
		case 1: img = gameAssets.gemBlue
		case 2: img = gameAssets.gemGreen
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			bob := math.Sin(float64(gem.animFrame)*0.15) * 5
			op.GeoM.Translate(gem.x-camX, gem.y+bob)
			screen.DrawImage(img, op)
		}
	}

	// PowerUps
	for _, pu := range g.powerUps {
		if !pu.alive {
			continue
		}
		var img *ebiten.Image
		switch pu.id {
		case 0: img = gameAssets.mushroomRed
		case 1: img = gameAssets.star
		case 2: img = gameAssets.flower1
		case 3: img = gameAssets.mushroomBrown
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			bob := math.Sin(float64(pu.animFrame)*0.1) * 5
			op.GeoM.Translate(pu.x-camX, pu.y+bob)
			screen.DrawImage(img, op)
		}
	}

	// Enemies
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		var img *ebiten.Image
		switch e.enemyType {
		case 0: img = gameAssets.slimeGreen
		case 1: img = gameAssets.bee
		case 2: img = gameAssets.frog
		case 3: img = gameAssets.mouse
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(e.x-camX, e.y)
			screen.DrawImage(img, op)
		} else {
			vector.DrawFilledCircle(screen, float32(e.x-camX+20), float32(e.y+20), 20, color.RGBA{255, 50, 50, 255}, true)
		}
	}

	// Flag
	var flagImg *ebiten.Image
	if g.frame%180 < 90 {
		flagImg = gameAssets.flagGreen
	} else {
		flagImg = gameAssets.flagRed
	}
	if flagImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.flagX-camX, g.flagX)
		screen.DrawImage(flagImg, op)
	}

	// Boss
	if g.boss != nil && g.boss.alive {
		vector.DrawFilledRect(screen, float32(g.boss.x-camX), float32(g.boss.y), float32(g.boss.width), float32(g.boss.height), color.RGBA{255, 50, 50, 255}, true)
		vector.StrokeRect(screen, float32(g.boss.x-camX), float32(g.boss.y), float32(g.boss.width), float32(g.boss.height), 3, color.RGBA{255, 200, 50, 255}, true)
		// Boss eyes
		vector.DrawFilledCircle(screen, float32(g.boss.x-camX+30), float32(g.boss.y+30), 15, color.RGBA{255, 255, 0, 255}, true)
		vector.DrawFilledCircle(screen, float32(g.boss.x-camX+70), float32(g.boss.y+30), 15, color.RGBA{255, 255, 0, 255}, true)
	}

	// Player
	g.drawPlayer(screen, camX)

	// Particles
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x-camX), float32(p.y), p.size, p.color, true)
	}

	// HUD
	g.drawHUD(screen)

	// Overlays
	if g.state == 2 {
		g.drawGameOver(screen)
	} else if g.state == 3 {
		g.drawWin(screen)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, camX float64) {
	p := g.player

	var img *ebiten.Image
	if !p.onGround {
		img = gameAssets.playerJump
	} else if math.Abs(p.vx) > 0.5 {
		if (p.animFrame/8)%2 == 0 {
			img = gameAssets.playerWalk1
		} else {
			img = gameAssets.playerWalk2
		}
	} else {
		img = gameAssets.playerStand
	}

	if img != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-camX, p.y)
		if p.facing == -1 {
			op.GeoM.Scale(-0.5, 0.5)
			op.GeoM.Translate(float64(p.width), 0)
		} else {
			op.GeoM.Scale(0.5, 0.5)
		}
		screen.DrawImage(img, op)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 55, color.RGBA{0, 0, 0, 180}, true)

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("World %d-%d", g.world, g.level), gameAssets.gameFont, 20, 18, color.RGBA{100, 200, 255, 255})
		text.Draw(screen, fmt.Sprintf("HP: %d", g.player.health), gameAssets.gameFont, 180, 18, color.RGBA{255, 100, 100, 255})
		text.Draw(screen, fmt.Sprintf("Lives: %d", g.player.lives), gameAssets.gameFont, 320, 18, color.RGBA{100, 255, 100, 255})
		text.Draw(screen, fmt.Sprintf("Coins: %d", g.player.coins), gameAssets.gameFont, 480, 18, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), gameAssets.gameFont, 650, 18, color.White)

		// Save hint
		text.Draw(screen, "[S] Save", gameAssets.gameFont, 850, 18, color.RGBA{150, 150, 150, 255})
	}

	// Boss health bar
	if g.boss != nil && g.boss.alive {
		vector.DrawFilledRect(screen, ScreenWidth/2-200, 70, 400, 20, color.RGBA{80, 0, 0, 255}, true)
		if g.boss.maxHealth > 0 {
			vector.DrawFilledRect(screen, ScreenWidth/2-200, 70, 400*float32(g.boss.health)/float32(g.boss.maxHealth), 20, color.RGBA{255, 50, 50, 255}, true)
		}
		if gameAssets.gameFont != nil {
			text.Draw(screen, "BOSS", gameAssets.gameFont, ScreenWidth/2-30, 85, color.White)
		}
	}
	
	// Show unlocked achievements (bottom right)
	y := ScreenHeight - 100
	for _, ach := range g.achievements {
		if ach.unlocked {
			if gameAssets.gameFont != nil {
				vector.DrawFilledRect(screen, ScreenWidth-300, float32(y), 290, 20, color.RGBA{0, 0, 0, 150}, true)
				text.Draw(screen, fmt.Sprintf("🏆 %s", ach.name), gameAssets.gameFont, ScreenWidth-290, y+15, color.RGBA{255, 215, 0, 255})
			}
			y -= 25
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 200}, true)
	if gameAssets.largeFont != nil {
		text.Draw(screen, "GAME OVER", gameAssets.largeFont, ScreenWidth/2-120, ScreenHeight/2-30, color.RGBA{255, 50, 50, 255})
		text.Draw(screen, "Press ENTER", gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+50, color.White)
	}
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 100, 0, 200}, true)
	if gameAssets.largeFont != nil {
		text.Draw(screen, "YOU WIN!", gameAssets.largeFont, ScreenWidth/2-100, ScreenHeight/2-30, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+50, color.White)
		text.Draw(screen, "Press ENTER", gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+100, color.White)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GO MARIO: SUPER PLATFORMER - Go365 Day 86")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

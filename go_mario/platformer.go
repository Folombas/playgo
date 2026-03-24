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
	ScreenWidth  = 800
	ScreenHeight = 600
	TileSize     = 40

	// Physics
	Gravity       = 0.5
	JumpForce     = -11.0
	RunSpeed      = 5.0
	WalkSpeed     = 2.5
	MaxFallSpeed  = 12.0
	Friction      = 0.8
	Acceleration  = 0.5

	// Tile types
	TileAir         = 0
	TileGround      = 1
	TileBrick       = 2
	TileQuestion    = 3
	TileHard        = 4
	TilePipeL       = 5
	TilePipeR       = 6
	TilePipeTopL    = 7
	TilePipeTopR    = 8
	TileCoin        = 9
	TileUsed        = 10

	// Enemy types
	EnemyGoomba  = 1
	EnemyKoopa   = 2
	EnemyPiranha = 3
	EnemyFly     = 4
	EnemyFrog    = 5
	EnemyMouse   = 6
	EnemySaw     = 7

	// New enemy types (Day 86)
	EnemyBarnacle = 8
	EnemyLadybug  = 9
	EnemySnail    = 10
	EnemyWorm     = 11
	EnemyFish     = 12

	// Powerup types
	PowerupMushroom = 1
	PowerupFlower   = 2
	PowerupStar     = 3
	Powerup1UP      = 4

	// Game states
	StateMenu     = 0
	StatePlaying  = 1
	StateGameOver = 2
	StateWon      = 3

	// Tile types extended
	TileSpike   = 11
	TileChest   = 12
	TileKey     = 13
	TileSpring  = 14
	TilePortal  = 15

	// Combo system
	MaxCombo    = 10
)

// ============================================================================
// ACHIEVEMENTS
// ============================================================================

type Achievement struct {
	id          string
	name        string
	description string
	unlocked    bool
}

var achievements = map[string]*Achievement{
	"first_blood": {
		id:          "first_blood",
		name:        "Первая кровь",
		description: "Победите первого врага",
		unlocked:    false,
	},
	"coin_master": {
		id:          "coin_master",
		name:        "Мастер монет",
		description: "Соберите 100 монет",
		unlocked:    false,
	},
	"enemy_slayer": {
		id:          "enemy_slayer",
		name:        "Убийца врагов",
		description: "Победите 10 врагов",
		unlocked:    false,
	},
	"treasure_hunter": {
		id:          "treasure_hunter",
		name:        "Охотник за сокровищами",
		description: "Найдите секретный сундук",
		unlocked:    false,
	},
	"speedrunner": {
		id:          "speedrunner",
		name:        "Спидраннер",
		description: "Пройдите уровень за 60 секунд",
		unlocked:    false,
	},
	"survivor": {
		id:          "survivor",
		name:        "Выживший",
		description: "Пройдите уровень без урона",
		unlocked:    false,
	},
}

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	// Player sprites (Green alien - main character)
	playerStand  *ebiten.Image
	playerWalk1  *ebiten.Image
	playerWalk2  *ebiten.Image
	playerJump   *ebiten.Image
	playerDuck   *ebiten.Image

	// Enemy sprites
	slimeGreen1  *ebiten.Image
	slimeGreen2  *ebiten.Image
	slimeBlue1   *ebiten.Image
	slimeBlue2   *ebiten.Image
	bee1         *ebiten.Image
	bee2         *ebiten.Image
	
	// New enemy sprites
	fly1         *ebiten.Image
	fly2         *ebiten.Image
	frog1        *ebiten.Image
	frog2        *ebiten.Image
	mouse1       *ebiten.Image
	mouse2       *ebiten.Image
	saw1         *ebiten.Image
	saw2         *ebiten.Image

	// Additional enemy sprites (Day 86)
	barnacle1    *ebiten.Image
	barnacle2    *ebiten.Image
	ladybug1     *ebiten.Image
	ladybug2     *ebiten.Image
	snail1       *ebiten.Image
	snail2       *ebiten.Image
	worm1        *ebiten.Image
	worm2        *ebiten.Image
	fish1        *ebiten.Image
	fish2        *ebiten.Image

	// Tile sprites
	grassTile    *ebiten.Image
	brickTile    *ebiten.Image
	questionTile *ebiten.Image
	hardTile     *ebiten.Image
	pipeTile     *ebiten.Image
	usedTile     *ebiten.Image

	// Items
	coinSprite   *ebiten.Image
	flagSprite   *ebiten.Image

	// Font
	gameFont     font.Face
}

var gameAssets *Assets
var tileImages map[int]*ebiten.Image

func LoadAssets() (*Assets, error) {
	assets := &Assets{}
	tileImages = make(map[int]*ebiten.Image)

	// Load player sprites (Green alien)
	var err error
	assets.playerStand, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	if err != nil {
		assets.playerStand = nil
	}

	assets.playerWalk1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	if err != nil {
		assets.playerWalk1 = nil
	}

	assets.playerWalk2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	if err != nil {
		assets.playerWalk2 = nil
	}

	assets.playerJump, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_jump.png")
	if err != nil {
		assets.playerJump = nil
	}

	// Load enemy sprites (2 frames for animation)
	assets.slimeGreen1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeGreen.png")
	if err != nil {
		assets.slimeGreen1 = nil
	}
	// Use same sprite for frame 2 if no second frame
	assets.slimeGreen2 = assets.slimeGreen1

	assets.slimeBlue1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeBlue.png")
	if err != nil {
		assets.slimeBlue1 = nil
	}
	assets.slimeBlue2 = assets.slimeBlue1

	assets.bee1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/bee.png")
	if err != nil {
		assets.bee1 = nil
	}
	assets.bee2 = assets.bee1

	// Load new enemy sprites
	assets.fly1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/fly_move.png")
	if err != nil {
		assets.fly1 = nil
	}
	assets.fly2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/fly.png")
	if err != nil {
		assets.fly2 = assets.fly1
	}

	assets.frog1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/frog_move.png")
	if err != nil {
		assets.frog1 = nil
	}
	assets.frog2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/frog.png")
	if err != nil {
		assets.frog2 = assets.frog1
	}

	assets.mouse1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/mouse_move.png")
	if err != nil {
		assets.mouse1 = nil
	}
	assets.mouse2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/mouse.png")
	if err != nil {
		assets.mouse2 = assets.mouse1
	}

	assets.saw1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/saw_move.png")
	if err != nil {
		assets.saw1 = nil
	}
	assets.saw2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/saw.png")
	if err != nil {
		assets.saw2 = assets.saw1
	}

	// Load additional enemy sprites (Day 86)
	assets.barnacle1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/barnacle.png")
	if err != nil {
		assets.barnacle1 = nil
	}
	assets.barnacle2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/barnacle_attack.png")
	if err != nil {
		assets.barnacle2 = assets.barnacle1
	}

	assets.ladybug1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/ladybug_move.png")
	if err != nil {
		assets.ladybug1 = nil
	}
	assets.ladybug2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/ladybug_fly.png")
	if err != nil {
		assets.ladybug2 = assets.ladybug1
	}

	assets.snail1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/snail_move.png")
	if err != nil {
		assets.snail1 = nil
	}
	assets.snail2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/snail.png")
	if err != nil {
		assets.snail2 = assets.snail1
	}

	assets.worm1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/wormGreen_move.png")
	if err != nil {
		assets.worm1 = nil
	}
	assets.worm2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/wormGreen.png")
	if err != nil {
		assets.worm2 = assets.worm1
	}

	assets.fish1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/fishGreen_move.png")
	if err != nil {
		assets.fish1 = nil
	}
	assets.fish2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/fishGreen.png")
	if err != nil {
		assets.fish2 = assets.fish1
	}

	// Load tile sprites
	assets.grassTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Ground/Grass/grass.png")
	if err != nil {
		assets.grassTile = nil
	}
	tileImages[TileGround] = assets.grassTile

	assets.brickTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickGrey.png")
	if err != nil {
		assets.brickTile = nil
	}
	tileImages[TileBrick] = assets.brickTile

	assets.questionTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem.png")
	if err != nil {
		assets.questionTile = nil
	}
	tileImages[TileQuestion] = assets.questionTile

	assets.hardTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickBrown.png")
	if err != nil {
		assets.hardTile = nil
	}
	tileImages[TileHard] = assets.hardTile

	assets.pipeTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/lockGreen.png")
	if err != nil {
		assets.pipeTile = nil
	}

	assets.usedTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem_disabled.png")
	if err != nil {
		assets.usedTile = nil
	}
	tileImages[TileUsed] = assets.usedTile

	// Load coin sprite
	assets.coinSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/coinGold.png")
	if err != nil {
		assets.coinSprite = nil
	}

	// Load font
	assets.gameFont, err = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 24)
	if err != nil {
		// Fallback to default font
		assets.gameFont = nil
	}

	return assets, nil
}

// loadFont загружает шрифт из файла
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

func (a *Assets) HasSprites() bool {
	return a.playerStand != nil
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

// Player - наш герой (Mario-style)
type Player struct {
	x, y        float64
	vx, vy      float64
	width       float32
	height      float32
	onGround    bool
	facing      int
	animFrame   int
	animTimer   int

	// Stats
	coins       int
	keys        int
	score       int
	lives       int
	world       int
	enemiesDefeated int
	damageTaken     int

	// Combo system
	combo       int
	comboTimer  int
	lastStompTime int

	// Power state
	isBig       bool
	isFire      bool
	isInvincible bool
	powerTimer  int
}

// Enemy - враг
type Enemy struct {
	x, y      float64
	vx, vy    float64
	width     float32
	height    float32
	enemyType int
	alive     bool
	squashed  bool
	animFrame int
	facing    int
}

// Powerup - бонус
type Powerup struct {
	x, y      float64
	vy        float64
	width     float32
	height    float32
	powerType int
	alive     bool
	animFrame int
}

// Particle - частица
type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	color   color.RGBA
	size    float32
}

// Level - уровень
type Level struct {
	width       int
	height      int
	tiles       [][]int
	coins       []Coin
	enemies     []*Enemy
	powerups    []*Powerup
	keys        []*Key
	chests      []*Chest
	spikes      []*Spike
	springs     []*Spring
	portals     []*Portal
	flagX       int
	flagY       int
	timeLimit   int
	timeElapsed int
}

// Spring - пружина
type Spring struct {
	x, y      float64
	width     float32
	height    float32
	compressed bool
	timer     int
	color     color.RGBA
}

// Portal - телепорт
type Portal struct {
	x, y      float64
	width     float32
	height    float32
	linkedTo  *Portal
	color     color.RGBA
	animFrame int
}

// Key - ключ
type Key struct {
	x, y      float64
	collected bool
	animFrame int
}

// Chest - сундук
type Chest struct {
	x, y      float64
	width     float32
	height    float32
	opened    bool
	locked    bool
	animFrame int
	contents  string // "coins", "star", "1up"
	value     int
}

// Spike - шипы
type Spike struct {
	x, y     float64
	width    float32
	height   float32
	damage   int
}

// Coin - монета
type Coin struct {
	x, y      float64
	collected bool
	animFrame int
}

// Camera - камера
type Camera struct {
	x, y float64
}

// Game - основная игра
type Game struct {
	player     *Player
	level      *Level
	camera     *Camera
	particles  []*Particle
	state      int
	frameCount int
	levelStartTime int

	// Achievements
	achievements   map[string]*Achievement
	newAchievements []*Achievement

	// Audio
	audioEnabled bool
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	// Load assets
	var err error
	gameAssets, err = LoadAssets()
	if err != nil {
		fmt.Println("Warning: Some assets failed to load, using fallback rendering")
	}

	g := &Game{
		player: &Player{
			x:      100,
			y:      100,
			width:  30,
			height: 40,
			facing: 1,
			lives:  3,
		},
		camera:     &Camera{},
		state:      StateMenu,
		frameCount: 0,
		levelStartTime: 0,
		achievements:   achievements,
		newAchievements: make([]*Achievement, 0),
		audioEnabled: true,
	}

	g.LoadLevel(1)
	return g
}

// LoadLevel загружает уровень
func (g *Game) LoadLevel(world int) {
	g.player.world = world
	g.level = GenerateLevel(world)
	g.player.x = 100
	g.player.y = 100
	g.player.vx = 0
	g.player.vy = 0
	g.camera.x = 0
	g.particles = make([]*Particle, 0)
	g.level.timeElapsed = 0
	g.levelStartTime = g.frameCount
	g.player.damageTaken = 0
	g.player.combo = 0 // Reset combo on level load
	g.player.comboTimer = 0
}

// GenerateLevel генерирует уровень
func GenerateLevel(world int) *Level {
	width := 200  // tiles
	height := 15  // tiles

	level := &Level{
		width:  width,
		height: height,
		tiles:  make([][]int, width),
		coins:  make([]Coin, 0),
		enemies: make([]*Enemy, 0),
		powerups: make([]*Powerup, 0),
		keys:   make([]*Key, 0),
		chests: make([]*Chest, 0),
		spikes: make([]*Spike, 0),
		springs: make([]*Spring, 0),
		portals: make([]*Portal, 0),
		timeLimit: 300 + world*30, // 5 секунд на уровень + бонус
	}

	// Initialize tiles
	for x := range level.tiles {
		level.tiles[x] = make([]int, height)
	}

	// Generate terrain
	for x := 0; x < width; x++ {
		// Ground
		for y := 10; y < height; y++ {
			level.tiles[x][y] = TileGround
		}

		// Gaps (pipes over pits)
		if x%50 == 45 && x > 50 {
			for y := 10; y < height; y++ {
				level.tiles[x][y] = TileAir
				if x+1 < width {
					level.tiles[x+1][y] = TileAir
				}
			}
		}

		// Random structures
		if x > 10 && rand.Float32() < 0.1 {
			structureType := rand.Intn(5)

			switch structureType {
			case 0: // Brick platform
				platY := rand.Intn(3) + 5
				for bx := 0; bx < 5; bx++ {
					if x+bx < width {
						level.tiles[x+bx][platY] = TileBrick
					}
				}
				// Add coin above
				if x+2 < width {
					level.coins = append(level.coins, Coin{
						x: float64((x+2)*TileSize),
						y: float64((platY-1)*TileSize),
					})
				}

			case 1: // Question block
				if x < width && rand.Intn(8) > 3 {
					level.tiles[x][rand.Intn(3)+5] = TileQuestion
				}

			case 2: // Pipe
				pipeHeight := rand.Intn(3) + 2
				pipeY := 10 - pipeHeight
				for py := pipeY; py < 10; py++ {
					if x < width {
						level.tiles[x][py] = TilePipeL
						if x+1 < width {
							level.tiles[x+1][py] = TilePipeR
						}
					}
				}
				// Pipe top
				if pipeY > 0 && x < width {
					level.tiles[x][pipeY-1] = TilePipeTopL
					if x+1 < width {
						level.tiles[x+1][pipeY-1] = TilePipeTopR
					}
				}

				// Piranha plant chance
				if rand.Float32() < 0.3 {
					level.enemies = append(level.enemies, &Enemy{
						x: float64(x * TileSize),
						y: float64((pipeY - 2) * TileSize),
						width: 30,
						height: 30,
						enemyType: EnemyPiranha,
						alive: true,
					})
				}

			case 3: // Enemy spawn
				// Select enemy type based on world and random chance
				enemyType := EnemyGoomba
				randVal := rand.Float32()

				if world >= 5 && randVal < 0.12 {
					enemyType = EnemyFish      // Fish in world 5+
				} else if world >= 4 && randVal < 0.18 {
					enemyType = EnemyWorm      // Worm in world 4+
				} else if world >= 4 && randVal < 0.24 {
					enemyType = EnemySaw       // Saw in later worlds
				} else if world >= 3 && randVal < 0.32 {
					enemyType = EnemySnail     // Snail in world 3+
				} else if world >= 3 && randVal < 0.40 {
					enemyType = EnemyMouse     // Mouse in world 3+
				} else if world >= 2 && randVal < 0.50 {
					enemyType = EnemyLadybug   // Ladybug in world 2+
				} else if world >= 2 && randVal < 0.60 {
					enemyType = EnemyFrog      // Frog in world 2+
				} else if world >= 2 && randVal < 0.70 {
					enemyType = EnemyFly       // Fly in world 2+
				} else if world >= 1 && randVal < 0.85 {
					enemyType = EnemyKoopa     // Koopa common
				} else if world >= 1 {
					enemyType = EnemyBarnacle  // Barnacle on structures
				}

				// Flying enemies spawn in air
				spawnY := 8 * TileSize
				if enemyType == EnemyFly || enemyType == EnemyLadybug {
					spawnY = 3 * TileSize // Spawn higher
				} else if enemyType == EnemyFish {
					spawnY = 6 * TileSize // Spawn in mid-air (water zones later)
				}

				level.enemies = append(level.enemies, &Enemy{
					x: float64(x * TileSize),
					y: float64(spawnY),
					width: 32,
					height: 32,
					enemyType: enemyType,
					alive: true,
					facing: -1,
				})

			case 4: // Stairs
				stairHeight := rand.Intn(4) + 2
				for sy := 0; sy < stairHeight; sy++ {
					for sx := 0; sx <= sy; sx++ {
						if x+sx < width {
							level.tiles[x+sx][9-sy] = TileHard
						}
					}
				}

			case 5: // Key spawn
				keyY := rand.Intn(4) + 4
				level.keys = append(level.keys, &Key{
					x: float64(x*TileSize + TileSize/2),
					y: float64(keyY*TileSize + TileSize/2),
				})

			case 6: // Chest spawn
				chestY := 9
				locked := rand.Float32() < 0.5
				contents := []string{"coins", "star", "1up"}
				content := contents[rand.Intn(len(contents))]
				value := rand.Intn(50) + 50
				if content == "star" {
					value = 1
				}
				level.chests = append(level.chests, &Chest{
					x: float64(x * TileSize),
					y: float64(chestY * TileSize),
					width: 40,
					height: 32,
					locked: locked,
					contents: content,
					value: value,
				})
			}
		}

		// Add spikes on ground
		if x > 20 && rand.Float32() < 0.03 {
			level.spikes = append(level.spikes, &Spike{
				x: float64(x * TileSize),
				y: float64(9*TileSize + 24),
				width: 40,
				height: 16,
				damage: 1,
			})
		}

		// Add springs (Day 84)
		if x > 30 && rand.Float32() < 0.015 {
			springColor := color.RGBA{255, 100, 100, 255} // Red spring
			if world >= 2 {
				springColor = color.RGBA{100, 255, 100, 255} // Green in world 2+
			}
			if world >= 4 {
				springColor = color.RGBA{100, 100, 255, 255} // Blue in world 4+
			}
			level.springs = append(level.springs, &Spring{
				x: float64(x * TileSize),
				y: float64(9*TileSize + 8),
				width: 40,
				height: 12,
				compressed: false,
				color: springColor,
			})
		}

		// Add portal pairs (Day 84) - spawn at specific positions
		if x == 50 && world >= 2 && len(level.portals) == 0 {
			portal1 := &Portal{
				x: float64(x * TileSize),
				y: float64(5 * TileSize),
				width: 50,
				height: 70,
				color: color.RGBA{150, 50, 255, 255}, // Purple
			}
			portal2 := &Portal{
				x: float64((x + 30) * TileSize),
				y: float64(5 * TileSize),
				width: 50,
				height: 70,
				color: color.RGBA{50, 150, 255, 255}, // Blue
			}
			portal1.linkedTo = portal2
			portal2.linkedTo = portal1
			level.portals = append(level.portals, portal1, portal2)
		}
	}

	// Add flag at end
	level.flagX = (width - 5) * TileSize
	level.flagY = 6 * TileSize

	// Add coins along the level
	for i := 0; i < 100; i++ {
		cx := rand.Intn(width-10) + 5
		cy := rand.Intn(8) + 2
		if level.tiles[cx][cy+1] != TileAir {
			level.coins = append(level.coins, Coin{
				x: float64(cx * TileSize),
				y: float64(cy * TileSize),
			})
		}
	}

	return level
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
		g.player.lives = 3
		g.player.score = 0
		g.player.coins = 0
		g.LoadLevel(1)
		playSound(SoundStart)
	}
}

func (g *Game) updatePlaying() {
	g.level.timeElapsed++

	// Check time limit for speedrunner achievement
	if g.level.timeElapsed < 60*60 && g.player.x >= float64(g.level.flagX) {
		g.unlockAchievement("speedrunner")
	}

	// Player input
	g.updatePlayer()

	// Update camera
	g.camera.x = g.player.x - ScreenWidth/2
	if g.camera.x < 0 {
		g.camera.x = 0
	}
	if g.camera.x > float64(g.level.width*TileSize-ScreenWidth) {
		g.camera.x = float64(g.level.width*TileSize - ScreenWidth)
	}

	// Update enemies
	g.updateEnemies()

	// Update powerups
	g.updatePowerups()

	// Update keys
	g.updateKeys()

	// Update chests
	g.updateChests()

	// Check spikes
	g.checkSpikes()

	// Update springs (Day 84)
	g.updateSprings()

	// Update portals (Day 84)
	g.updatePortals()

	// Update particles
	g.updateParticles()

	// Update achievements
	g.updateAchievements()

	// Update combo (Day 84)
	g.updateCombo()

	// Check win condition
	if g.player.x >= float64(g.level.flagX) {
		g.state = StateWon
		playSound(SoundWin)

		// Check survivor achievement
		if g.player.damageTaken == 0 {
			g.unlockAchievement("survivor")
		}
	}

	// Check death
	if g.player.y > ScreenHeight {
		g.playerDie()
	}
}

func (g *Game) updatePlayer() {
	p := g.player

	// Horizontal movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		if p.vx < RunSpeed {
			p.vx += Acceleration
		}
		p.facing = 1
		p.animFrame++
		// Play footstep sound periodically while walking
		if p.onGround && p.animFrame%20 == 0 {
			playSound(SoundFootstep)
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		if p.vx > -RunSpeed {
			p.vx -= Acceleration
		}
		p.facing = -1
		p.animFrame++
		// Play footstep sound periodically while walking
		if p.onGround && p.animFrame%20 == 0 {
			playSound(SoundFootstep)
		}
	} else {
		// Friction
		p.vx *= Friction
		if math.Abs(p.vx) < 0.1 {
			p.vx = 0
		}
	}

	// Jump
	if (ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeySpace)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
		playSound(SoundJump)
		g.spawnJumpParticles(p.x+float64(p.width/2), p.y+float64(p.height))
	}

	// Variable jump height
	if !ebiten.IsKeyPressed(ebiten.KeyArrowUp) && !ebiten.IsKeyPressed(ebiten.KeyW) && !ebiten.IsKeyPressed(ebiten.KeySpace) && p.vy < JumpForce/2 {
		p.vy *= 0.5
	}

	// Apply gravity
	p.vy += Gravity
	if p.vy > MaxFallSpeed {
		p.vy = MaxFallSpeed
	}

	// Move and collide
	p.x += p.vx
	g.collideHorizontal(p)

	p.y += p.vy
	g.collideVertical(p)

	// World bounds
	if p.x < 0 {
		p.x = 0
		p.vx = 0
	}

	// Animation timer
	if p.animFrame > 1000 {
		p.animFrame = 0
	}

	// Invincibility timer
	if p.isInvincible {
		p.powerTimer--
		if p.powerTimer <= 0 {
			p.isInvincible = false
		}
	}
}

func (g *Game) collideHorizontal(p *Player) {
	leftTile := int(p.x) / TileSize
	rightTile := int(p.x+float64(p.width)) / TileSize
	topTile := int(p.y) / TileSize
	bottomTile := int(p.y+float64(p.height)-1) / TileSize

	// Check left
	if p.vx < 0 {
		if g.isSolid(leftTile, topTile) || g.isSolid(leftTile, bottomTile) {
			p.x = float64((leftTile+1)*TileSize)
			p.vx = 0
		}
	}

	// Check right
	if p.vx > 0 {
		if g.isSolid(rightTile, topTile) || g.isSolid(rightTile, bottomTile) {
			p.x = float64(rightTile*TileSize - int(p.width))
			p.vx = 0
		}
	}
}

func (g *Game) collideVertical(p *Player) {
	p.onGround = false

	leftTile := int(p.x) / TileSize
	rightTile := int(p.x+float64(p.width)) / TileSize

	// Check falling
	if p.vy > 0 {
		bottomTile := int(p.y+float64(p.height)) / TileSize

		if g.isSolid(leftTile, bottomTile) || g.isSolid(rightTile, bottomTile) {
			p.y = float64(bottomTile*TileSize - int(p.height))
			p.vy = 0
			p.onGround = true
		}
	}

	// Check jumping
	if p.vy < 0 {
		topTile := int(p.y) / TileSize

		if g.isSolid(leftTile, topTile) || g.isSolid(rightTile, topTile) {
			p.y = float64((topTile+1)*TileSize)
			p.vy = 0

			// Hit block
			g.hitBlock(leftTile, topTile)
		}
	}
}

func (g *Game) isSolid(x, y int) bool {
	if x < 0 || x >= g.level.width || y < 0 || y >= g.level.height {
		return false
	}
	tile := g.level.tiles[x][y]
	return tile != TileAir && tile != TileCoin && tile != TileQuestion && tile != TileUsed
}

func (g *Game) hitBlock(x, y int) {
	if x < 0 || x >= g.level.width || y < 0 || y >= g.level.height {
		return
	}

	tile := g.level.tiles[x][y]

	if tile == TileQuestion {
		g.level.tiles[x][y] = TileUsed
		g.player.coins++
		g.player.score += 200
		playSound(SoundCoin)
		g.spawnCoinParticles(float64(x*TileSize), float64(y*TileSize))
		g.spawnPowerupParticles(float64(x*TileSize+TileSize/2), float64(y*TileSize))

		// Chance for powerup
		if rand.Float32() < 0.1 {
			powerType := PowerupMushroom
			if g.player.isBig {
				powerType = PowerupFlower
			}
			g.level.powerups = append(g.level.powerups, &Powerup{
				x: float64(x * TileSize),
				y: float64((y - 1) * TileSize),
				powerType: powerType,
				alive: true,
				width: 30,
				height: 30,
			})
		}
	} else if tile == TileBrick {
		if g.player.isBig {
			g.level.tiles[x][y] = TileAir
			g.player.score += 50
			playSound(SoundBreak)
			g.spawnParticles(float64(x*TileSize+TileSize/2), float64(y*TileSize+TileSize/2), 8, color.RGBA{139, 69, 19, 255})
		} else {
			playSound(SoundBump)
		}
	}
}

func (g *Game) updateEnemies() {
	for _, enemy := range g.level.enemies {
		if !enemy.alive || enemy.squashed {
			continue
		}

		enemy.animFrame++

		// Enemy AI based on type
		switch enemy.enemyType {
		case EnemyPiranha:
			// Move up and down in pipe
			enemy.y += math.Sin(float64(enemy.animFrame)*0.05) * 0.5

		case EnemyFly:
			// Flying enemy - moves in sine wave pattern
			enemy.x += float64(enemy.facing) * 0.8
			enemy.y += math.Sin(float64(enemy.animFrame)*0.1) * 1.5
			// Change direction periodically
			if enemy.animFrame%200 == 0 {
				enemy.facing = -enemy.facing
			}

		case EnemyFrog:
			// Frog - jumps around
			enemy.x += float64(enemy.facing) * 0.3
			enemy.vy += Gravity * 0.3
			enemy.y += enemy.vy
			// Ground check for frog
			bottomTile := int(enemy.y+float64(enemy.height)+1) / TileSize
			leftTile := int(enemy.x) / TileSize
			rightTile := int(enemy.x+float64(enemy.width)) / TileSize
			if g.isSolid(leftTile, bottomTile) || g.isSolid(rightTile, bottomTile) {
				enemy.y = float64(bottomTile*TileSize - int(enemy.height))
				enemy.vy = -8 // Jump!
			}
			// Change direction at edges
			if enemy.animFrame%150 == 0 {
				enemy.facing = -enemy.facing
			}

		case EnemyMouse:
			// Mouse - fast ground enemy
			enemy.x += float64(enemy.facing) * 1.2
			// Turn at edges or walls
			leftTile := int(enemy.x) / TileSize
			rightTile := int(enemy.x+float64(enemy.width)) / TileSize
			bottomTile := int(enemy.y+float64(enemy.height)+1) / TileSize
			if enemy.facing < 0 && (!g.isSolid(leftTile, bottomTile) || g.isSolid(leftTile, int(enemy.y)/TileSize)) {
				enemy.facing = 1
			} else if enemy.facing > 0 && (!g.isSolid(rightTile, bottomTile) || g.isSolid(rightTile, int(enemy.y)/TileSize)) {
				enemy.facing = -1
			}

		case EnemySaw:
			// Saw - moves horizontally, damages on touch
			enemy.x += math.Sin(float64(enemy.animFrame)*0.08) * 2
			// Saws can't be stomped

		// New enemy behaviors (Day 86)
		case EnemyBarnacle:
			// Barnacle - stationary enemy that attacks when player is near
			distToPlayer := math.Abs(g.player.x - enemy.x)
			if distToPlayer < 100 {
				// Attack mode - extend
				enemy.y += math.Sin(float64(enemy.animFrame)*0.2) * 0.3
			}

		case EnemyLadybug:
			// Ladybug - flies in figure-8 pattern
			enemy.x += float64(enemy.facing) * 1.0
			enemy.y += math.Sin(float64(enemy.animFrame)*0.15) * 2
			if enemy.animFrame%180 == 0 {
				enemy.facing = -enemy.facing
			}

		case EnemySnail:
			// Snail - very slow but armored
			enemy.x += float64(enemy.facing) * 0.2
			// Turn at edges
			leftTile := int(enemy.x) / TileSize
			rightTile := int(enemy.x+float64(enemy.width)) / TileSize
			bottomTile := int(enemy.y+float64(enemy.height)+1) / TileSize
			if enemy.facing < 0 && !g.isSolid(leftTile, bottomTile) {
				enemy.facing = 1
			} else if enemy.facing > 0 && !g.isSolid(rightTile, bottomTile) {
				enemy.facing = -1
			}

		case EnemyWorm:
			// Worm - emerges from ground periodically
			baseY := enemy.y
			if enemy.animFrame%200 < 100 {
				// Emerged - can be stomped
				enemy.y = baseY - 10
			} else {
				// Hidden in ground - invulnerable
				enemy.y = baseY
			}

		case EnemyFish:
			// Fish - swims in water/lava areas
			enemy.x += float64(enemy.facing) * 1.5
			enemy.y += math.Sin(float64(enemy.animFrame)*0.1) * 1.0
			if enemy.animFrame%150 == 0 {
				enemy.facing = -enemy.facing
			}

		default:
			// Goomba/Koopa - simple walk
			enemy.x += float64(enemy.facing) * 0.5

			// Turn at edges or walls
			leftTile := int(enemy.x) / TileSize
			rightTile := int(enemy.x+float64(enemy.width)) / TileSize
			bottomTile := int(enemy.y+float64(enemy.height)+1) / TileSize

			if enemy.facing < 0 && (!g.isSolid(leftTile, bottomTile) || g.isSolid(leftTile, int(enemy.y)/TileSize)) {
				enemy.facing = 1
			} else if enemy.facing > 0 && (!g.isSolid(rightTile, bottomTile) || g.isSolid(rightTile, int(enemy.y)/TileSize)) {
				enemy.facing = -1
			}
		}

		// Collision with player
		if g.checkCollision(g.player, enemy) {
			// Can't stomp certain enemies
			canStomp := enemy.enemyType != EnemyFly && enemy.enemyType != EnemySaw

			// Special enemy behaviors
			switch enemy.enemyType {
			case EnemyBarnacle:
				// Can only stomp when retracted
				if math.Sin(float64(enemy.animFrame)*0.2) > 0.5 {
					canStomp = false
				}
			case EnemySnail:
				// Snail requires 2 stomps or star power
				if !g.player.isInvincible && !g.player.isBig {
					canStomp = false // Too armored
				}
			case EnemyWorm:
				// Can only stomp when emerged
				if enemy.animFrame%200 >= 100 {
					canStomp = false
				}
			}

			if enemy.enemyType == EnemyPiranha || enemy.enemyType == EnemyBarnacle {
				g.playerHit()
			} else if canStomp && g.player.vy > 0 && g.player.y+float64(g.player.height) < enemy.y+float64(enemy.height)/2 {
				// Stomp enemy
				enemy.squashed = true
				g.player.vy = -6
				g.player.score += 100
				playSound(SoundStomp)
				g.spawnEnemyDefeatParticles(enemy.x+float64(enemy.width/2), enemy.y+float64(enemy.height/2), enemy.enemyType)
				
				// Combo system (Day 84)
				g.player.combo++
				if g.player.combo > MaxCombo {
					g.player.combo = MaxCombo
				}
				g.player.comboTimer = 120 // 2 seconds at 60 FPS
				g.player.lastStompTime = g.frameCount
				
				// Combo multiplier
				comboMultiplier := 1.0 + float64(g.player.combo)*0.5
				bonusScore := int(float64(100) * comboMultiplier)
				g.player.score += bonusScore
			} else if !g.player.isInvincible {
				g.playerHit()
			}
		}
	}

	// Remove squashed enemies
	activeEnemies := make([]*Enemy, 0)
	for _, e := range g.level.enemies {
		if e.alive && !e.squashed {
			activeEnemies = append(activeEnemies, e)
		}
	}
	g.level.enemies = activeEnemies
}

func (g *Game) updatePowerups() {
	for _, p := range g.level.powerups {
		if !p.alive {
			continue
		}

		p.animFrame++
		p.vy += Gravity
		p.y += p.vy

		// Ground collision
		bottomTile := int(p.y+float64(p.height)) / TileSize
		leftTile := int(p.x) / TileSize
		rightTile := int(p.x+float64(p.width)) / TileSize

		if g.isSolid(leftTile, bottomTile) || g.isSolid(rightTile, bottomTile) {
			p.y = float64(bottomTile*TileSize - int(p.height))
			p.vy = 0
		}

		// Collision with player
		if g.checkPlayerPowerup(p) {
			p.alive = false
			g.applyPowerup(p.powerType)
		}
	}
}

func (g *Game) checkPlayerPowerup(p *Powerup) bool {
	return g.player.x < p.x+float64(p.width) &&
		g.player.x+float64(g.player.width) > p.x &&
		g.player.y < p.y+float64(p.height) &&
		g.player.y+float64(g.player.height) > p.y
}

func (g *Game) applyPowerup(powerType int) {
	playSound(SoundPowerup)

	switch powerType {
	case PowerupMushroom:
		g.player.isBig = true
		g.player.height = 50
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{220, 20, 60, 255})

	case PowerupFlower:
		g.player.isFire = true
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{255, 100, 0, 255})

	case PowerupStar:
		g.player.isInvincible = true
		g.player.powerTimer = 600 // 10 seconds
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 30, color.RGBA{255, 215, 0, 255})

	case Powerup1UP:
		g.player.lives++
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{0, 255, 0, 255})
	}
}

// ============================================================================
// KEYS & CHESTS
// ============================================================================

func (g *Game) updateKeys() {
	for _, key := range g.level.keys {
		if key.collected {
			continue
		}
		key.animFrame++

		// Check collision
		if g.checkKeyCollision(key) {
			key.collected = true
			g.player.keys++
			g.player.score += 50
			playSound(SoundItem)
			g.spawnKeyParticles(key.x, key.y)
		}
	}
}

func (g *Game) checkKeyCollision(key *Key) bool {
	dx := (g.player.x + float64(g.player.width)/2) - key.x
	dy := (g.player.y + float64(g.player.height)/2) - key.y
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist < float64(g.player.width/2+g.player.height/3)
}

func (g *Game) updateChests() {
	for _, chest := range g.level.chests {
		if chest.opened {
			continue
		}
		chest.animFrame++

		// Check collision
		if g.player.x < chest.x+float64(chest.width) &&
			g.player.x+float64(g.player.width) > chest.x &&
			g.player.y < chest.y+float64(chest.height) &&
			g.player.y+float64(g.player.height) > chest.y {

			if chest.locked {
				if g.player.keys > 0 {
					g.player.keys--
					chest.locked = false
					playSound(SoundItem)
				}
			} else {
				// Open chest
				chest.opened = true
				g.player.score += chest.value

				switch chest.contents {
				case "coins":
					g.player.coins += chest.value
				case "star":
					g.player.isInvincible = true
					g.player.powerTimer = 600
				case "1up":
					g.player.lives++
				}

				playSound(SoundPowerup)
				g.spawnChestParticles(chest.x+float64(chest.width)/2, chest.y+float64(chest.height)/2)
				g.unlockAchievement("treasure_hunter")
			}
		}
	}
}

func (g *Game) checkSpikes() {
	for _, spike := range g.level.spikes {
		if g.player.x < spike.x+float64(spike.width) &&
			g.player.x+float64(g.player.width) > spike.x &&
			g.player.y < spike.y+float64(spike.height) &&
			g.player.y+float64(g.player.height) > spike.y {

			if !g.player.isInvincible {
				g.player.damageTaken++
				g.playerHit()
			}
		}
	}
}

// ============================================================================
// ACHIEVEMENTS
// ============================================================================

func (g *Game) unlockAchievement(id string) {
	if ach, ok := g.achievements[id]; ok {
		if !ach.unlocked {
			ach.unlocked = true
			g.newAchievements = append(g.newAchievements, ach)
			playSound(SoundPowerup)
		}
	}
}

func (g *Game) updateAchievements() {
	// Remove old notifications after 3 seconds (180 frames)
	if len(g.newAchievements) > 0 && g.frameCount%180 == 0 {
		if len(g.newAchievements) > 1 {
			g.newAchievements = g.newAchievements[1:]
		} else {
			g.newAchievements = make([]*Achievement, 0)
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.2
		p.life--

		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

// updateCombo обновляет таймер комбо (Day 84)
func (g *Game) updateCombo() {
	if g.player.combo > 0 {
		g.player.comboTimer--
		if g.player.comboTimer <= 0 {
			g.player.combo = 0
		}
	}
}

// updateSprings обновляет пружины (Day 84)
func (g *Game) updateSprings() {
	for _, spring := range g.level.springs {
		if spring.compressed {
			spring.timer--
			if spring.timer <= 0 {
				spring.compressed = false
			}
		}
		
		// Check collision with player
		if g.player.x < spring.x+float64(spring.width) &&
			g.player.x+float64(g.player.width) > spring.x &&
			g.player.y+float64(g.player.height) > spring.y &&
			g.player.y < spring.y+float64(spring.height) {
			
			if !spring.compressed && g.player.vy > 0 {
				// Compress spring and bounce
				spring.compressed = true
				spring.timer = 10
				g.player.vy = -16 // High bounce!
				g.player.onGround = false
				playSound(SoundPowerup)
				g.spawnParticles(spring.x+float64(spring.width/2), spring.y, 10, spring.color)
			}
		}
	}
}

// updatePortals обновляет телепорты (Day 84)
func (g *Game) updatePortals() {
	for _, portal := range g.level.portals {
		portal.animFrame++
		
		// Check collision with player
		if g.player.x < portal.x+float64(portal.width) &&
			g.player.x+float64(g.player.width) > portal.x &&
			g.player.y < portal.y+float64(portal.height) &&
			g.player.y+float64(g.player.height) > portal.y {
			
			// Teleport player
			if portal.linkedTo != nil {
				g.player.x = portal.linkedTo.x - float64(g.player.width)
				g.player.y = portal.linkedTo.y
				g.player.vx = 0
				g.player.vy = 0
				playSound(SoundDoor)
				g.spawnParticles(portal.x+float64(portal.width/2), portal.y+float64(portal.height/2), 20, portal.color)
				g.spawnParticles(portal.linkedTo.x+float64(portal.linkedTo.width/2), portal.linkedTo.y+float64(portal.linkedTo.height/2), 20, portal.linkedTo.color)
			}
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: float64(rand.Intn(10)-5) * 0.5,
			vy: float64(rand.Intn(10)-5) * 0.5,
			life: 30 + rand.Intn(20),
			color: c,
			size: float32(rand.Intn(4)+2),
		})
	}
}

// spawnJumpParticles создаёт частицы при прыжке
func (g *Game) spawnJumpParticles(x, y float64) {
	g.spawnParticles(x, y, 8, color.RGBA{200, 200, 200, 255}) // Dust
}

// spawnCoinParticles создаёт частицы при сборе монеты
func (g *Game) spawnCoinParticles(x, y float64) {
	for i := 0; i < 10; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + 10,
			y: y + 10,
			vx: float64(rand.Intn(8)-4) * 0.8,
			vy: float64(-rand.Intn(10)-5) * 0.5,
			life: 40 + rand.Intn(20),
			color: color.RGBA{255, 215, 0, 255}, // Gold
			size: float32(rand.Intn(6)+3),
		})
	}
}

// spawnStompParticles создаёт частицы при уничтожении врага
func (g *Game) spawnStompParticles(x, y float64) {
	for i := 0; i < 12; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + 16,
			y: y + 16,
			vx: float64(rand.Intn(12)-6) * 0.7,
			vy: float64(rand.Intn(12)-6) * 0.7,
			life: 25 + rand.Intn(15),
			color: color.RGBA{139, 69, 19, 255}, // Brown
			size: float32(rand.Intn(5)+2),
		})
	}
}

// spawnHitParticles создаёт частицы при получении урона
func (g *Game) spawnHitParticles(x, y float64) {
	for i := 0; i < 15; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + 15,
			y: y + 20,
			vx: float64(rand.Intn(14)-7) * 0.6,
			vy: float64(rand.Intn(14)-7) * 0.6,
			life: 35 + rand.Intn(20),
			color: color.RGBA{255, 50, 50, 255}, // Red
			size: float32(rand.Intn(5)+3),
		})
	}
}

// spawnPowerupParticles создаёт частицы при получении бонуса
func (g *Game) spawnPowerupParticles(x, y float64) {
	for i := 0; i < 20; i++ {
		angle := float64(i) * 2 * math.Pi / 20
		speed := 2.0
		g.particles = append(g.particles, &Particle{
			x: x + 16,
			y: y + 16,
			vx: math.Cos(angle) * speed,
			vy: math.Sin(angle) * speed,
			life: 50 + rand.Intn(20),
			color: color.RGBA{255, 100, 100, 255}, // Pink/Red
			size: float32(rand.Intn(6)+4),
		})
	}
}

// spawnKeyParticles создаёт частицы при сборе ключа
func (g *Game) spawnKeyParticles(x, y float64) {
	for i := 0; i < 12; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: float64(rand.Intn(8)-4) * 0.7,
			vy: float64(-rand.Intn(8)-4) * 0.5,
			life: 35 + rand.Intn(15),
			color: color.RGBA{255, 215, 0, 255}, // Gold
			size: float32(rand.Intn(5) + 3),
		})
	}
}

// spawnChestParticles создаёт частицы при открытии сундука
func (g *Game) spawnChestParticles(x, y float64) {
	for i := 0; i < 30; i++ {
		angle := float64(i) * 2 * math.Pi / 30
		speed := float64(rand.Intn(10)+5) * 0.7
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: math.Cos(angle) * speed,
			vy: math.Sin(angle) * speed,
			life: 50 + rand.Intn(20),
			color: color.RGBA{255, 215, 0, 255}, // Gold treasure
			size: float32(rand.Intn(6) + 4),
		})
	}
}

// spawnEnemyDefeatParticles создаёт частицы при уничтожении врага (тип зависит от врага)
func (g *Game) spawnEnemyDefeatParticles(x, y float64, enemyType int) {
	switch enemyType {
	case EnemyFly:
		// Green sparkles for fly
		for i := 0; i < 15; i++ {
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: float64(rand.Intn(16)-8) * 0.8,
				vy: float64(rand.Intn(16)-8) * 0.8,
				life: 30 + rand.Intn(20),
				color: color.RGBA{100, 255, 100, 255}, // Green
				size: float32(rand.Intn(5)+2),
			})
		}
	case EnemyFrog:
		// Green blobs for frog
		for i := 0; i < 18; i++ {
			angle := float64(i) * 2 * math.Pi / 18
			speed := float64(rand.Intn(5)+3) * 0.5
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: math.Cos(angle) * speed,
				vy: math.Sin(angle) * speed,
				life: 35 + rand.Intn(15),
				color: color.RGBA{0, 200, 100, 255}, // Frog green
				size: float32(rand.Intn(6)+3),
			})
		}
	case EnemyMouse:
		// Brown/orange for mouse
		for i := 0; i < 14; i++ {
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: float64(rand.Intn(14)-7) * 0.9,
				vy: float64(-rand.Intn(10)-3) * 0.6,
				life: 28 + rand.Intn(12),
				color: color.RGBA{180, 120, 60, 255}, // Brown/orange
				size: float32(rand.Intn(5)+2),
			})
		}
	case EnemySaw:
		// Metal sparks for saw (can't be stomped, but destroyed by star)
		for i := 0; i < 25; i++ {
			angle := float64(i) * 2 * math.Pi / 25
			speed := float64(rand.Intn(8)+4) * 0.7
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: math.Cos(angle) * speed,
				vy: math.Sin(angle) * speed,
				life: 40 + rand.Intn(20),
				color: color.RGBA{200, 200, 200, 255}, // Silver
				size: float32(rand.Intn(4)+2),
			})
		}

	// New enemy particles (Day 86)
	case EnemyBarnacle:
		// Red/orange for barnacle
		for i := 0; i < 16; i++ {
			angle := float64(i) * 2 * math.Pi / 16
			speed := float64(rand.Intn(6)+3) * 0.6
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: math.Cos(angle) * speed,
				vy: math.Sin(angle) * speed,
				life: 30 + rand.Intn(15),
				color: color.RGBA{255, 100, 50, 255}, // Red-orange
				size: float32(rand.Intn(5)+2),
			})
		}
	case EnemyLadybug:
		// Red with black spots for ladybug
		for i := 0; i < 18; i++ {
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: float64(rand.Intn(16)-8) * 0.7,
				vy: float64(-rand.Intn(12)-4) * 0.5,
				life: 32 + rand.Intn(18),
				color: color.RGBA{220, 20, 20, 255}, // Red
				size: float32(rand.Intn(5)+2),
			})
		}
	case EnemySnail:
		// Green/brown for snail
		for i := 0; i < 14; i++ {
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: float64(rand.Intn(12)-6) * 0.5,
				vy: float64(-rand.Intn(8)-2) * 0.4,
				life: 35 + rand.Intn(15),
				color: color.RGBA{100, 150, 50, 255}, // Green-brown
				size: float32(rand.Intn(5)+3),
			})
		}
	case EnemyWorm:
		// Pink for worm
		for i := 0; i < 16; i++ {
			angle := float64(i) * 2 * math.Pi / 16
			speed := float64(rand.Intn(5)+2) * 0.6
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: math.Cos(angle) * speed,
				vy: math.Sin(angle) * speed,
				life: 28 + rand.Intn(12),
				color: color.RGBA{255, 150, 150, 255}, // Pink
				size: float32(rand.Intn(4)+2),
			})
		}
	case EnemyFish:
		// Blue/cyan for fish
		for i := 0; i < 20; i++ {
			angle := float64(i) * 2 * math.Pi / 20
			speed := float64(rand.Intn(7)+4) * 0.6
			g.particles = append(g.particles, &Particle{
				x: x + 16,
				y: y + 16,
				vx: math.Cos(angle) * speed,
				vy: math.Sin(angle) * speed,
				life: 35 + rand.Intn(15),
				color: color.RGBA{50, 150, 255, 255}, // Blue
				size: float32(rand.Intn(5)+2),
			})
		}

	default:
		// Default brown for Goomba/Koopa
		g.spawnStompParticles(x, y)
	}
}

func (g *Game) checkCollision(p *Player, e *Enemy) bool {
	return p.x < e.x+float64(e.width) &&
		p.x+float64(p.width) > e.x &&
		p.y < e.y+float64(e.height) &&
		p.y+float64(p.height) > e.y
}

func (g *Game) playerHit() {
	if g.player.isInvincible {
		return
	}

	if g.player.isBig {
		g.player.isBig = false
		g.player.height = 40
		g.player.isInvincible = true
		g.player.powerTimer = 120
		playSound(SoundHit)
		g.spawnHitParticles(g.player.x, g.player.y)
	} else {
		g.playerDie()
	}
}

func (g *Game) playerDie() {
	g.player.lives--
	g.player.combo = 0 // Reset combo on death
	g.player.comboTimer = 0
	playSound(SoundDie)

	if g.player.lives <= 0 {
		g.state = StateGameOver
	} else {
		g.LoadLevel(g.player.world)
	}
}

func (g *Game) updateEndScreen() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StatePlaying
		g.player.lives = 3
		g.player.score = 0
		g.player.coins = 0
		g.LoadLevel(1)
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
	// Sky gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(100 + y/10)
		g := uint8(150 + y/15)
		b := uint8(200)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	// Title with shadow
	title := "SUPER GO MARIO"
	titleX := ScreenWidth/2 - 140
	
	// Shadow
	text.Draw(screen, title, gameAssets.gameFont, titleX+3, 153, color.RGBA{0, 0, 0, 100})
	// Main title
	text.Draw(screen, title, gameAssets.gameFont, titleX, 150, color.RGBA{255, 215, 0, 255})

	// Subtitle
	subtitle := "A Classic 2D Platformer"
	subX := ScreenWidth/2 - 100
	text.Draw(screen, subtitle, gameAssets.gameFont, subX, 200, color.RGBA{255, 255, 255, 255})

	// Decorative line
	vector.DrawFilledRect(screen, 200, 230, 400, 3, color.RGBA{255, 255, 255, 200}, false)

	// Instructions
	instructions := []string{
		"Arrow Keys / WASD - Move",
		"Space / W / Up - Jump",
		"Stomp enemies, collect coins & keys!",
		"Use springs to bounce high!",
		"Portals teleport you away!",
		"Build combos for bonus points!",
		"Press ENTER or SPACE to Start",
	}

	for i, line := range instructions {
		textColor := color.RGBA{255, 255, 255, 255}
		if i == len(instructions)-1 {
			textColor = color.RGBA{100, 255, 100, 255} // Green for start prompt
		}
		text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-130, 270+i*32, textColor)
	}

	// Features
	features := "🗝️ Keys  🎁 Chests  ⚠️ Spikes  🏆 Achievements"
	text.Draw(screen, features, gameAssets.gameFont, ScreenWidth/2-180, ScreenHeight-90, color.RGBA{255, 215, 0, 255})
	
	newFeatures := "🔴 Springs  🌀 Portals  💥 Combo System"
	text.Draw(screen, newFeatures, gameAssets.gameFont, ScreenWidth/2-160, ScreenHeight-60, color.RGBA{100, 255, 100, 255})

	// Version info
	versionText := "Go365 - Day 84 | March 24, 2026 | Combo + New Blocks!"
	text.Draw(screen, versionText, gameAssets.gameFont, 20, ScreenHeight-30, color.RGBA{200, 200, 200, 255})
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Sky
	screen.Fill(color.RGBA{100, 150, 200, 255})

	// Draw level
	g.drawLevel(screen)

	// Draw player
	g.drawPlayer(screen)

	// Draw particles
	g.drawParticles(screen)

	// Draw UI
	g.drawUI(screen)

	// Draw achievement notifications
	g.drawAchievements(screen)
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), p.size, p.color, true)
	}
}

func (g *Game) drawAchievements(screen *ebiten.Image) {
	if len(g.newAchievements) == 0 {
		return
	}

	// Draw achievement notification
	y := float32(80)
	for _, ach := range g.newAchievements {
		vector.DrawFilledRect(screen, 20, y-30, float32(ScreenWidth-40), 70, color.RGBA{0, 0, 0, 200}, false)
		vector.DrawFilledRect(screen, 20, y-30, float32(ScreenWidth-40), 3, color.RGBA{255, 215, 0, 255}, false)

		if gameAssets != nil && gameAssets.gameFont != nil {
			text.Draw(screen, "🏆 ДОСТИЖЕНИЕ!", gameAssets.gameFont, 40, int(y)-10, color.RGBA{255, 215, 0, 255})
			text.Draw(screen, ach.name, gameAssets.gameFont, 40, int(y)+15, color.RGBA{255, 255, 255, 255})
		}
		y += 80
	}
}

func (g *Game) drawLevel(screen *ebiten.Image) {
	startX := int(g.camera.x) / TileSize
	endX := startX + ScreenWidth/TileSize + 2

	for x := startX; x < endX && x < g.level.width; x++ {
		for y := 0; y < g.level.height; y++ {
			tile := g.level.tiles[x][y]
			if tile != TileAir {
				drawX := float32(x*TileSize) - float32(g.camera.x)
				drawY := float32(y * TileSize)

				g.drawTile(screen, tile, drawX, drawY)
			}
		}
	}

	// Draw coins
	for _, coin := range g.level.coins {
		if coin.collected {
			continue
		}

		drawX := float32(coin.x) - float32(g.camera.x)
		drawY := float32(coin.y)

		if drawX > -20 && drawX < ScreenWidth+20 {
			coin.animFrame++
			
			// Use coin sprite if available
			if gameAssets != nil && gameAssets.coinSprite != nil {
				op := &ebiten.DrawImageOptions{}
				offset := float32(math.Sin(float64(coin.animFrame)*0.1) * 3)
				op.GeoM.Translate(float64(drawX), float64(drawY+offset))
				screen.DrawImage(gameAssets.coinSprite, op)
			} else {
				// Fallback: Vector coin
				vector.DrawFilledCircle(screen, drawX+10, drawY+10, 8, color.RGBA{255, 215, 0, 255}, false)
				vector.DrawFilledCircle(screen, drawX+10, drawY+10, 5, color.RGBA{255, 235, 100, 255}, false)
			}
		}
	}

	// Draw flag with animation
	flagX := float32(g.level.flagX) - float32(g.camera.x)
	flagFrame := g.frameCount / 15 % 3
	
	// Flag pole
	vector.StrokeLine(screen, flagX+10, float32(g.level.flagY), flagX+10, float32(g.level.flagY+TileSize*4), 3, color.RGBA{100, 100, 100, 255}, false)
	
	// Animated flag sprite or fallback
	if gameAssets != nil {
		var flagSprite *ebiten.Image
		var err error
		switch flagFrame {
		case 0:
			flagSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/flagRed1.png")
		case 1:
			flagSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/flagRed2.png")
		default:
			flagSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/flagRed_down.png")
		}
		if err == nil && flagSprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(flagX+10), float64(g.level.flagY))
			screen.DrawImage(flagSprite, op)
		} else {
			// Fallback: Vector flag
			vector.DrawFilledRect(screen, flagX+10, float32(g.level.flagY), 40, 30, color.RGBA{0, 200, 0, 255}, false)
		}
	} else {
		// Fallback: Vector flag
		vector.DrawFilledRect(screen, flagX+10, float32(g.level.flagY), 40, 30, color.RGBA{0, 200, 0, 255}, false)
	}

	// Draw keys
	g.drawKeys(screen)

	// Draw chests
	g.drawChests(screen)

	// Draw spikes
	g.drawSpikes(screen)

	// Draw springs (Day 84)
	g.drawSprings(screen)

	// Draw portals (Day 84)
	g.drawPortals(screen)

	// Draw enemies
	g.drawEnemies(screen)
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	for _, enemy := range g.level.enemies {
		if !enemy.alive {
			continue
		}

		drawX := float32(enemy.x) - float32(g.camera.x)
		drawY := float32(enemy.y)

		// Skip if off-screen
		if drawX < -40 || drawX > ScreenWidth+40 {
			continue
		}

		// Use enemy sprite with animation if available
		var sprite *ebiten.Image
		if gameAssets != nil {
			// Animation frame based on enemy position/time
			animFrame := (enemy.animFrame / 10) % 2

			switch enemy.enemyType {
			case EnemyGoomba:
				if animFrame == 0 {
					sprite = gameAssets.slimeGreen1
				} else {
					sprite = gameAssets.slimeGreen2
				}
			case EnemyKoopa:
				if animFrame == 0 {
					sprite = gameAssets.slimeBlue1
				} else {
					sprite = gameAssets.slimeBlue2
				}
			case EnemyFly:
				if animFrame == 0 {
					sprite = gameAssets.fly1
				} else {
					sprite = gameAssets.fly2
				}
			case EnemyFrog:
				if animFrame == 0 {
					sprite = gameAssets.frog1
				} else {
					sprite = gameAssets.frog2
				}
			case EnemyMouse:
				if animFrame == 0 {
					sprite = gameAssets.mouse1
				} else {
					sprite = gameAssets.mouse2
				}
			case EnemySaw:
				if animFrame == 0 {
					sprite = gameAssets.saw1
				} else {
					sprite = gameAssets.saw2
				}

			// New enemy sprites (Day 86)
			case EnemyBarnacle:
				if animFrame == 0 {
					sprite = gameAssets.barnacle1
				} else {
					sprite = gameAssets.barnacle2
				}
			case EnemyLadybug:
				if animFrame == 0 {
					sprite = gameAssets.ladybug1
				} else {
					sprite = gameAssets.ladybug2
				}
			case EnemySnail:
				if animFrame == 0 {
					sprite = gameAssets.snail1
				} else {
					sprite = gameAssets.snail2
				}
			case EnemyWorm:
				if animFrame == 0 {
					sprite = gameAssets.worm1
				} else {
					sprite = gameAssets.worm2
				}
			case EnemyFish:
				if animFrame == 0 {
					sprite = gameAssets.fish1
				} else {
					sprite = gameAssets.fish2
				}

			default:
				sprite = gameAssets.slimeGreen1
			}
		}

		if sprite != nil {
			scale := float64(float32(enemy.height) / float32(sprite.Bounds().Dy()))
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(drawX), float64(drawY))

			// Flip if facing left
			if enemy.facing == -1 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(float64(float32(sprite.Bounds().Dx())*float32(enemy.height)/float32(sprite.Bounds().Dy())), 0)
			}

			screen.DrawImage(sprite, op)
		} else {
			// Fallback: Vector enemy with unique colors per type
			enemyColor := color.RGBA{139, 69, 19, 255} // Brown (Goomba)
			switch enemy.enemyType {
			case EnemyKoopa:
				enemyColor = color.RGBA{0, 100, 200, 255} // Blue
			case EnemyFly:
				enemyColor = color.RGBA{100, 255, 100, 255} // Green
			case EnemyFrog:
				enemyColor = color.RGBA{0, 200, 100, 255} // Frog green
			case EnemyMouse:
				enemyColor = color.RGBA{150, 100, 50, 255} // Brown mouse
			case EnemySaw:
				enemyColor = color.RGBA{150, 150, 150, 255} // Gray saw
			}
			vector.DrawFilledRect(screen, drawX, drawY, enemy.width, enemy.height, enemyColor, false)

			// Eyes for organic enemies
			if enemy.enemyType != EnemySaw {
				eyeY := drawY + 8
				eyeOffset := enemy.facing * 3
				vector.DrawFilledCircle(screen, drawX+8+float32(eyeOffset), eyeY, 4, color.RGBA{255, 255, 255, 255}, false)
				vector.DrawFilledCircle(screen, drawX+20+float32(eyeOffset), eyeY, 4, color.RGBA{255, 255, 255, 255}, false)
			}
		}
	}
}

// ============================================================================
// DRAW KEYS, CHESTS, SPIKES
// ============================================================================

func (g *Game) drawKeys(screen *ebiten.Image) {
	for _, key := range g.level.keys {
		if key.collected {
			continue
		}
		drawX := float32(key.x) - float32(g.camera.x)
		drawY := float32(key.y)

		if drawX > -20 && drawX < ScreenWidth+20 {
			key.animFrame++
			offset := float32(math.Sin(float64(key.animFrame)*0.1) * 5)

			// Draw key
			vector.DrawFilledCircle(screen, drawX, drawY+offset, 8, color.RGBA{255, 215, 0, 255}, false)
			vector.DrawFilledRect(screen, drawX-2, drawY+offset+8, 4, 10, color.RGBA{255, 215, 0, 255}, false)
		}
	}
}

func (g *Game) drawChests(screen *ebiten.Image) {
	for _, chest := range g.level.chests {
		if chest.opened {
			continue
		}
		drawX := float32(chest.x) - float32(g.camera.x)
		drawY := float32(chest.y)

		if chest.locked {
			// Locked chest - red
			vector.DrawFilledRect(screen, drawX, drawY+10, chest.width, chest.height-10, color.RGBA{139, 90, 43, 255}, false)
			vector.DrawFilledRect(screen, drawX+5, drawY, chest.width-10, 15, color.RGBA{100, 80, 40, 255}, false)
			// Lock symbol
			vector.DrawFilledCircle(screen, drawX+float32(chest.width)/2, drawY+float32(chest.height)/2, 6, color.RGBA{255, 0, 0, 255}, false)
		} else {
			// Unlocked chest - gold
			vector.DrawFilledRect(screen, drawX, drawY+10, chest.width, chest.height-10, color.RGBA{180, 120, 60, 255}, false)
			vector.DrawFilledRect(screen, drawX+5, drawY, chest.width-10, 15, color.RGBA{150, 100, 50, 255}, false)
			// Gold trim
			vector.DrawFilledRect(screen, drawX+float32(chest.width)/2-5, drawY+float32(chest.height)/2-3, 10, 6, color.RGBA{255, 215, 0, 255}, false)
		}
	}
}

func (g *Game) drawSpikes(screen *ebiten.Image) {
	for _, spike := range g.level.spikes {
		drawX := float32(spike.x) - float32(g.camera.x)
		drawY := float32(spike.y)

		// Draw spike as triangle
		vector.StrokeLine(screen, drawX, drawY+spike.height, drawX+spike.width/2, drawY, 3, color.RGBA{150, 150, 150, 255}, false)
		vector.StrokeLine(screen, drawX+spike.width/2, drawY, drawX+spike.width, drawY+spike.height, 3, color.RGBA{150, 150, 150, 255}, false)
		// Fill
		for i := int(drawX); i < int(drawX+spike.width); i++ {
			height := spike.height * (1 - float32(math.Abs(float64(i-int(drawX)-int(spike.width/2))))/float32(spike.width/2))
			vector.DrawFilledRect(screen, float32(i), drawY+spike.height-float32(height), 1, float32(height), color.RGBA{150, 150, 150, 180}, false)
		}
	}
}

// drawSprings отрисовывает пружины (Day 84)
func (g *Game) drawSprings(screen *ebiten.Image) {
	for _, spring := range g.level.springs {
		drawX := float32(spring.x) - float32(g.camera.x)
		drawY := float32(spring.y)

		if spring.compressed {
			// Compressed spring
			vector.DrawFilledRect(screen, drawX, drawY+4, spring.width, spring.height/2, spring.color, false)
		} else {
			// Extended spring
			vector.DrawFilledRect(screen, drawX, drawY, spring.width, spring.height, spring.color, false)
			// Spring coils
			for i := 0; i < 4; i++ {
				y := drawY + float32(i*3) + 2
				vector.StrokeLine(screen, drawX+2, y, drawX+spring.width-2, y, 2, color.RGBA{255, 255, 255, 200}, false)
			}
		}
	}
}

// drawPortals отрисовывает телепорты (Day 84)
func (g *Game) drawPortals(screen *ebiten.Image) {
	for _, portal := range g.level.portals {
		drawX := float32(portal.x) - float32(g.camera.x)
		drawY := float32(portal.y)

		// Animated swirling portal
		animOffset := float32(math.Sin(float64(portal.animFrame)*0.15) * 5)
		
		// Portal frame
		vector.DrawFilledRect(screen, drawX-3, drawY, 6, portal.height, color.RGBA{100, 100, 100, 255}, false)
		vector.DrawFilledRect(screen, drawX-3, drawY, portal.width+6, 6, color.RGBA{100, 100, 100, 255}, false)
		
		// Portal swirl effect
		for i := 0; i < 5; i++ {
			y := drawY + float32(i*14) + animOffset
			vector.StrokeLine(screen, drawX, y, drawX+portal.width, y, 3, portal.color, false)
		}
		
		// Portal glow
		vector.DrawFilledRect(screen, drawX+5, drawY+5, portal.width-10, portal.height-10, color.RGBA{200, 100, 255, 100}, false)
	}
}

func (g *Game) drawTile(screen *ebiten.Image, tile int, x, y float32) {
	// Try to use sprite first
	if gameAssets != nil {
		var sprite *ebiten.Image
		switch tile {
		case TileGround:
			sprite = gameAssets.grassTile
		case TileBrick:
			sprite = gameAssets.brickTile
		case TileQuestion:
			sprite = gameAssets.questionTile
		case TileHard:
			sprite = gameAssets.hardTile
		case TileUsed:
			sprite = gameAssets.usedTile
		case TilePipeL, TilePipeTopL, TilePipeR, TilePipeTopR:
			sprite = gameAssets.pipeTile
		}

		if sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(sprite, op)
			return
		}
	}

	// Fallback: Vector-based rendering
	switch tile {
	case TileGround:
		// Ground with grass top
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{139, 69, 19, 255}, false)
		vector.DrawFilledRect(screen, x, y, TileSize, 8, color.RGBA{34, 139, 34, 255}, false)

	case TileBrick:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{178, 34, 34, 255}, false)
		// Brick pattern
		vector.StrokeLine(screen, x, y+TileSize/2, x+TileSize, y+TileSize/2, 2, color.RGBA{100, 20, 20, 255}, false)
		vector.StrokeLine(screen, x+TileSize/2, y, x+TileSize/2, y+TileSize/2, 2, color.RGBA{100, 20, 20, 255}, false)

	case TileQuestion:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{255, 215, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "?", int(x)+12, int(y)+10)

	case TileHard:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{128, 128, 128, 255}, false)

	case TileUsed:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{100, 80, 60, 255}, false)

	case TilePipeL, TilePipeTopL:
		vector.DrawFilledRect(screen, x, y, TileSize/2, TileSize, color.RGBA{0, 180, 0, 255}, false)
		vector.DrawFilledRect(screen, x+2, y, 4, TileSize, color.RGBA{0, 220, 0, 255}, false)

	case TilePipeR, TilePipeTopR:
		vector.DrawFilledRect(screen, x, y, TileSize/2, TileSize, color.RGBA{0, 160, 0, 255}, false)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := g.player
	drawX := float32(p.x) - float32(g.camera.x)
	drawY := float32(p.y)

	// Flicker when invincible
	if p.isInvincible && g.frameCount%4 < 2 {
		return
	}

	// Use sprite if available
	if gameAssets != nil && gameAssets.HasSprites() {
		var sprite *ebiten.Image
		
		// Choose sprite based on state
		if !p.onGround {
			sprite = gameAssets.playerJump
		} else if p.vy != 0 || p.vx != 0 {
			// Walking animation
			if (p.animFrame / 8) % 2 == 0 {
				sprite = gameAssets.playerWalk1
			} else {
				sprite = gameAssets.playerWalk2
			}
		} else {
			sprite = gameAssets.playerStand
		}
		
		if sprite != nil {
			// Scale sprite to player size
			scale := float64(float32(p.height) / 256.0 * 2.5) // Adjust scale
			width := float64(float32(sprite.Bounds().Dx()) * float32(p.height) / 256.0 * 2.5)
			
			// Flip if facing left
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(drawX), float64(drawY))
			
			if p.facing == -1 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(width, 0)
			}
			
			screen.DrawImage(sprite, op)
			return
		}
	}

	// Fallback: Vector-based rendering (original code)
	// Body color based on power state
	bodyColor := color.RGBA{220, 20, 20, 255} // Red (Mario)
	if p.isFire {
		bodyColor = color.RGBA{255, 200, 0, 255} // White/Fire
	}

	// Body
	vector.DrawFilledRect(screen, drawX, drawY, p.width, p.height, bodyColor, false)

	// Overalls (blue)
	vector.DrawFilledRect(screen, drawX+5, drawY+p.height-15, p.width-10, 10, color.RGBA{0, 0, 180, 255}, false)

	// Face
	faceX := drawX + p.width/2
	faceY := drawY + 10
	vector.DrawFilledCircle(screen, faceX, faceY, 10, color.RGBA{255, 220, 180, 255}, false)

	// Hat
	vector.DrawFilledRect(screen, drawX+2, drawY, p.width-4, 8, bodyColor, false)

	// Eyes
	eyeOffset := p.facing * 2
	vector.DrawFilledCircle(screen, faceX+float32(eyeOffset)-2, faceY+2, 3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, faceX+float32(eyeOffset)+2, faceY+2, 3, color.RGBA{0, 0, 0, 255}, false)

	// Mustache
	vector.DrawFilledRect(screen, faceX-5+float32(eyeOffset), faceY+6, 10, 3, color.RGBA{50, 30, 20, 255}, false)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 180}, false)
	vector.DrawFilledRect(screen, 0, 39, ScreenWidth, 2, color.RGBA{100, 100, 100, 255}, false)

	// Use custom font if available, otherwise fallback to DebugPrint
	if gameAssets != nil && gameAssets.gameFont != nil {
		// Score
		scoreText := fmt.Sprintf("MARIO\n%06d", g.player.score)
		text.Draw(screen, scoreText, gameAssets.gameFont, 20, 28, color.RGBA{255, 255, 255, 255})

		// Coins
		coinText := fmt.Sprintf("COINS\nx%02d", g.player.coins)
		text.Draw(screen, coinText, gameAssets.gameFont, 180, 28, color.RGBA{255, 215, 0, 255})

		// Keys
		keyText := fmt.Sprintf("KEYS\nx%d", g.player.keys)
		text.Draw(screen, keyText, gameAssets.gameFont, 280, 28, color.RGBA{255, 100, 100, 255})

		// World
		worldText := fmt.Sprintf("WORLD\n%d-1", g.player.world)
		text.Draw(screen, worldText, gameAssets.gameFont, 380, 28, color.RGBA{100, 200, 100, 255})

		// Lives
		livesText := fmt.Sprintf("LIVES\nx%d", g.player.lives)
		text.Draw(screen, livesText, gameAssets.gameFont, 540, 28, color.RGBA{255, 100, 100, 255})

		// Time
		timeLeft := g.level.timeLimit - g.level.timeElapsed
		timeText := fmt.Sprintf("TIME\n%d", timeLeft/60)
		timeColor := color.RGBA{255, 255, 255, 255}
		if timeLeft < 100 {
			timeColor = color.RGBA{255, 100, 100, 255}
		}
		text.Draw(screen, timeText, gameAssets.gameFont, 680, 28, timeColor)
		
		// Combo indicator (Day 84)
		if g.player.combo > 1 {
			comboText := fmt.Sprintf("COMBO\nx%d", g.player.combo)
			comboColor := color.RGBA{255, 215, 0, 255} // Gold
			if g.player.combo >= 5 {
				comboColor = color.RGBA{255, 100, 100, 255} // Red for high combo
			}
			if g.player.combo >= MaxCombo {
				comboColor = color.RGBA{150, 50, 255, 255} // Purple for MAX combo
			}
			text.Draw(screen, comboText, gameAssets.gameFont, 420, 28, comboColor)
		}
	} else {
		// Fallback to DebugPrint
		scoreText := fmt.Sprintf("MARIO\n%06d", g.player.score)
		ebitenutil.DebugPrintAt(screen, scoreText, 20, 5)

		coinText := fmt.Sprintf("COINS\nx%02d", g.player.coins)
		ebitenutil.DebugPrintAt(screen, coinText, 150, 5)

		worldText := fmt.Sprintf("WORLD\n%d-1", g.player.world)
		ebitenutil.DebugPrintAt(screen, worldText, 280, 5)

		livesText := fmt.Sprintf("LIVES\nx%d", g.player.lives)
		ebitenutil.DebugPrintAt(screen, livesText, 410, 5)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Dark red background
	screen.Fill(color.RGBA{50, 0, 0, 255})

	title := "GAME OVER"
	text.Draw(screen, title, gameAssets.gameFont, ScreenWidth/2-90, ScreenHeight/2-20, color.RGBA{255, 50, 50, 255})

	scoreText := fmt.Sprintf("Final Score: %06d", g.player.score)
	text.Draw(screen, scoreText, gameAssets.gameFont, ScreenWidth/2-110, ScreenHeight/2+30, color.RGBA{255, 255, 255, 255})

	restartText := "Press ENTER to restart"
	text.Draw(screen, restartText, gameAssets.gameFont, ScreenWidth/2-110, ScreenHeight/2+80, color.RGBA{200, 200, 200, 255})
}

func (g *Game) drawWon(screen *ebiten.Image) {
	// Victory gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(100 + y/10)
		g := uint8(180 + y/15)
		b := uint8(150)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	title := "COURSE CLEAR!"
	text.Draw(screen, title, gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2-50, color.RGBA{255, 215, 0, 255})

	scoreText := fmt.Sprintf("Score: %06d", g.player.score)
	text.Draw(screen, scoreText, gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+10, color.RGBA{255, 255, 255, 255})

	coinsText := fmt.Sprintf("Coins: %d", g.player.coins)
	text.Draw(screen, coinsText, gameAssets.gameFont, ScreenWidth/2-110, ScreenHeight/2+50, color.RGBA{255, 215, 0, 255})

	// Count unlocked achievements
	unlockedCount := 0
	for _, ach := range g.achievements {
		if ach.unlocked {
			unlockedCount++
		}
	}
	achText := fmt.Sprintf("Achievements: %d/%d", unlockedCount, len(g.achievements))
	text.Draw(screen, achText, gameAssets.gameFont, ScreenWidth/2-110, ScreenHeight/2+90, color.RGBA{255, 100, 255, 255})

	restartText := "Press ENTER to continue"
	text.Draw(screen, restartText, gameAssets.gameFont, ScreenWidth/2-110, ScreenHeight/2+130, color.RGBA{255, 255, 255, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// AUDIO (Enhanced with Kenney RPG Audio)
// ============================================================================

type SoundType int

const (
	SoundJump SoundType = iota
	SoundCoin
	SoundStomp
	SoundHit
	SoundDie
	SoundPowerup
	SoundBump
	SoundBreak
	SoundStart
	SoundWin
	SoundFootstep
	SoundDoor
	SoundItem
)

var audioCtx *audio.Context

// soundFiles maps sound types to Kenney RPG audio files
var soundFiles = map[SoundType]string{
	SoundJump:     "Audio/footstep00.ogg",
	SoundFootstep: "Audio/footstep01.ogg",
	SoundCoin:     "Audio/handleCoins.ogg",
	SoundStomp:    "Audio/cloth1.ogg",
	SoundHit:      "Audio/creak1.ogg",
	SoundDie:      "Audio/doorClose_1.ogg",
	SoundPowerup:  "Audio/bookOpen.ogg",
	SoundBump:     "Audio/bookPlace1.ogg",
	SoundBreak:    "Audio/knifeSlice.ogg",
	SoundStart:    "Audio/metalClick.ogg",
	SoundWin:      "Audio/handleCoins2.ogg",
	SoundDoor:     "Audio/doorOpen_1.ogg",
	SoundItem:     "Audio/beltHandle1.ogg",
}

func initAudio() {
	audioCtx = audio.NewContext(44100)
}

// loadSoundFromFile загружает звук из файла
func loadSoundFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// generateBeep creates a simple beep sound (fallback)
func generateBeep(frequency, duration float64) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]byte, numSamples*2)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		// Simple sine wave with envelope
		envelope := 1.0 - float64(i)/float64(numSamples)
		value := math.Sin(2*math.Pi*frequency*t) * envelope * 0.3

		// Convert to 16-bit
		sample := int16(value * 32767)
		samples[i*2] = byte(sample)
		samples[i*2+1] = byte(sample >> 8)
	}

	return samples
}

func playSound(sound SoundType) {
	if audioCtx == nil {
		return
	}

	// Try to load from file first
	filePath, ok := soundFiles[sound]
	if ok {
		fullPath := "assets/sounds/" + filePath
		data, err := loadSoundFromFile(fullPath)
		if err == nil && len(data) > 0 {
			player := audioCtx.NewPlayerFromBytes(data)
			player.SetVolume(0.5)
			player.Play()
			return
		}
	}

	// Fallback to generated beep
	var samples []byte
	switch sound {
	case SoundJump:
		samples = generateBeep(400, 0.1)
	case SoundCoin:
		samples = generateBeep(1200, 0.15)
	case SoundStomp:
		samples = generateBeep(200, 0.08)
	case SoundHit:
		samples = generateBeep(150, 0.2)
	case SoundDie:
		samples = generateBeep(100, 0.5)
	case SoundPowerup:
		samples = generateBeep(800, 0.3)
	case SoundBump:
		samples = generateBeep(100, 0.05)
	case SoundBreak:
		samples = generateBeep(80, 0.1)
	case SoundStart:
		samples = generateBeep(600, 0.2)
	case SoundWin:
		samples = generateBeep(800, 0.4)
	case SoundFootstep:
		samples = generateBeep(300, 0.05)
	case SoundItem:
		samples = generateBeep(1000, 0.1)
	}

	if len(samples) > 0 {
		player := audioCtx.NewPlayerFromBytes(samples)
		player.Play()
	}
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	initAudio()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🍄 Super Go Mario - Classic 2D Platformer")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

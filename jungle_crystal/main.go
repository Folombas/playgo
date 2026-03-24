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
	TileSpike       = 11
	TileChest       = 12
	TileKey         = 13
	TileSpring      = 14
	TilePortal      = 15
	TileCrystal     = 16 // New crystal tile

	// Enemy types
	EnemyGoomba   = 1
	EnemyKoopa    = 2
	EnemyPiranha  = 3
	EnemyFly      = 4
	EnemyFrog     = 5
	EnemyMouse    = 6
	EnemySaw      = 7
	EnemyBarnacle = 8
	EnemyLadybug  = 9
	EnemySnail    = 10
	EnemyWorm     = 11
	EnemyFish     = 12
	EnemyBeetle   = 13 // New jungle enemy

	// Powerup types
	PowerupMushroom = 1
	PowerupFlower   = 2
	PowerupStar     = 3
	Powerup1UP      = 4
	PowerupCrystal  = 5 // New crystal powerup

	// Game states
	StateMenu     = 0
	StatePlaying  = 1
	StateGameOver = 2
	StateWon      = 3

	// Combo system
	MaxCombo = 10
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
	"crystal_collector": {
		id:          "crystal_collector",
		name:        "Коллекционер кристаллов",
		description: "Соберите 50 кристаллов",
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

	// Enemy sprites
	fly1         *ebiten.Image
	fly2         *ebiten.Image
	frog1        *ebiten.Image
	frog2        *ebiten.Image
	mouse1       *ebiten.Image
	mouse2       *ebiten.Image
	saw1         *ebiten.Image
	saw2         *ebiten.Image

	// Additional enemy sprites
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
	crystalTile  *ebiten.Image // New crystal sprite

	// Items
	coinSprite   *ebiten.Image
	flagSprite   *ebiten.Image
	crystalItem  *ebiten.Image // New crystal item sprite

	// Font
	gameFont     font.Face
}

var gameAssets *Assets
var tileImages map[int]*ebiten.Image

func LoadAssets() (*Assets, error) {
	assets := &Assets{}
	tileImages = make(map[int]*ebiten.Image)

	var err error

	// Load player sprites (Green alien)
	assets.playerStand, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	if err != nil {
		assets.playerStand = nil
	}

	assets.playerWalk1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	if err != nil {
		assets.playerWalk1 = nil
	}

	assets.playerWalk2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	if err != nil {
		assets.playerWalk2 = nil
	}

	assets.playerJump, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Players/128x256/Green/alienGreen_jump.png")
	if err != nil {
		assets.playerJump = nil
	}

	// Load enemy sprites
	assets.slimeGreen1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/slimeGreen.png")
	if err != nil {
		assets.slimeGreen1 = nil
	}
	assets.slimeGreen2 = assets.slimeGreen1

	assets.slimeBlue1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/slimeBlue.png")
	if err != nil {
		assets.slimeBlue1 = nil
	}
	assets.slimeBlue2 = assets.slimeBlue1

	assets.bee1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/bee.png")
	if err != nil {
		assets.bee1 = nil
	}
	assets.bee2 = assets.bee1

	// Load new enemy sprites
	assets.fly1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/fly_move.png")
	if err != nil {
		assets.fly1 = nil
	}
	assets.fly2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/fly.png")
	if err != nil {
		assets.fly2 = assets.fly1
	}

	assets.frog1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/frog_move.png")
	if err != nil {
		assets.frog1 = nil
	}
	assets.frog2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/frog.png")
	if err != nil {
		assets.frog2 = assets.frog1
	}

	assets.mouse1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/mouse_move.png")
	if err != nil {
		assets.mouse1 = nil
	}
	assets.mouse2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/mouse.png")
	if err != nil {
		assets.mouse2 = assets.mouse1
	}

	assets.saw1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/saw_move.png")
	if err != nil {
		assets.saw1 = nil
	}
	assets.saw2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/saw.png")
	if err != nil {
		assets.saw2 = assets.saw1
	}

	// Load additional enemy sprites
	assets.barnacle1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/barnacle.png")
	if err != nil {
		assets.barnacle1 = nil
	}
	assets.barnacle2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/barnacle_attack.png")
	if err != nil {
		assets.barnacle2 = assets.barnacle1
	}

	assets.ladybug1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/ladybug_move.png")
	if err != nil {
		assets.ladybug1 = nil
	}
	assets.ladybug2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/ladybug_fly.png")
	if err != nil {
		assets.ladybug2 = assets.ladybug1
	}

	assets.snail1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/snail_move.png")
	if err != nil {
		assets.snail1 = nil
	}
	assets.snail2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/snail.png")
	if err != nil {
		assets.snail2 = assets.snail1
	}

	assets.worm1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/wormGreen_move.png")
	if err != nil {
		assets.worm1 = nil
	}
	assets.worm2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/wormGreen.png")
	if err != nil {
		assets.worm2 = assets.worm1
	}

	assets.fish1, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/fishGreen_move.png")
	if err != nil {
		assets.fish1 = nil
	}
	assets.fish2, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Enemies/fishGreen.png")
	if err != nil {
		assets.fish2 = assets.fish1
	}

	// Load tile sprites
	assets.grassTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Ground/Grass/grass.png")
	if err != nil {
		assets.grassTile = nil
	}
	tileImages[TileGround] = assets.grassTile

	assets.brickTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Tiles/brickGrey.png")
	if err != nil {
		assets.brickTile = nil
	}
	tileImages[TileBrick] = assets.brickTile

	assets.questionTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Tiles/boxItem.png")
	if err != nil {
		assets.questionTile = nil
	}
	tileImages[TileQuestion] = assets.questionTile

	assets.hardTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Tiles/brickBrown.png")
	if err != nil {
		assets.hardTile = nil
	}
	tileImages[TileHard] = assets.hardTile

	assets.usedTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Tiles/boxItem_disabled.png")
	if err != nil {
		assets.usedTile = nil
	}
	tileImages[TileUsed] = assets.usedTile

	// Load crystal sprite (new!)
	assets.crystalTile, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Items/crystal.png")
	if err != nil {
		// Fallback: use coin sprite tinted
		assets.crystalTile = nil
	}

	// Load coin sprite
	assets.coinSprite, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Items/coinGold.png")
	if err != nil {
		assets.coinSprite = nil
	}

	// Load crystal item sprite
	assets.crystalItem, _, err = ebitenutil.NewImageFromFile("../go_mario/assets/PNG/Items/crystal.png")
	if err != nil {
		assets.crystalItem = nil
	}

	// Load font from playgo/fonts
	assets.gameFont, err = loadFont("../../fonts/super-adorable-font.zip", 24)
	if err != nil {
		// Fallback to go_mario font
		assets.gameFont, err = loadFont("../go_mario/assets/fonts/SuperAdorable-MAvyp.ttf", 24)
		if err != nil {
			assets.gameFont = nil
		}
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
	crystals    int // New crystal currency

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

type Powerup struct {
	x, y      float64
	vy        float64
	width     float32
	height    float32
	powerType int
	alive     bool
	animFrame int
}

type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	color   color.RGBA
	size    float32
}

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
	crystals    []Crystal // New crystal collectibles
	flagX       int
	flagY       int
	timeLimit   int
	timeElapsed int
}

type Crystal struct {
	x, y      float64
	collected bool
	animFrame int
	color     color.RGBA // Different crystal colors
}

type Spring struct {
	x, y      float64
	width     float32
	height    float32
	compressed bool
	timer     int
	color     color.RGBA
}

type Portal struct {
	x, y      float64
	width     float32
	height    float32
	linkedTo  *Portal
	color     color.RGBA
	animFrame int
}

type Key struct {
	x, y      float64
	collected bool
	animFrame int
}

type Chest struct {
	x, y      float64
	width     float32
	height    float32
	opened    bool
	locked    bool
	animFrame int
	contents  string
	value     int
}

type Spike struct {
	x, y     float64
	width    float32
	height   float32
	damage   int
}

type Coin struct {
	x, y      float64
	collected bool
	animFrame int
}

type Camera struct {
	x, y float64
}

// Audio system
type AudioPlayer struct {
	audioCtx *audio.Context
	jumpSound *audio.Player
	coinSound *audio.Player
	stompSound *audio.Player
	powerupSound *audio.Player
	dieSound *audio.Player
	winSound *audio.Player
	crystalSound *audio.Player // New crystal sound
	enabled    bool
}

type SoundType int

const (
	SoundJump SoundType = iota
	SoundCoin
	SoundStomp
	SoundPowerup
	SoundDie
	SoundWin
	SoundCrystal // New crystal sound
	SoundFootstep
	SoundSpring
	SoundPortal
	SoundStart
)

var audioPlayer *AudioPlayer

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

	// Initialize audio
	audioPlayer = &AudioPlayer{
		enabled: true,
	}

	g := &Game{
		player: &Player{
			x:      100,
			y:      100,
			width:  30,
			height: 40,
			facing: 1,
			lives:  3,
			crystals: 0,
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
	g.player.combo = 0
	g.player.comboTimer = 0
	g.player.crystals = 0
}

func GenerateLevel(world int) *Level {
	width := 200
	height := 15

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
		crystals: make([]Crystal, 0),
		timeLimit: 300 + world*30,
	}

	// Initialize tiles
	for x := range level.tiles {
		level.tiles[x] = make([]int, height)
	}

	// Generate terrain with jungle theme
	for x := 0; x < width; x++ {
		// Ground
		for y := 10; y < height; y++ {
			level.tiles[x][y] = TileGround
		}

		// Gaps
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
				enemyType := EnemyGoomba
				randVal := rand.Float32()

				if world >= 5 && randVal < 0.12 {
					enemyType = EnemyFish
				} else if world >= 4 && randVal < 0.18 {
					enemyType = EnemyWorm
				} else if world >= 4 && randVal < 0.24 {
					enemyType = EnemySaw
				} else if world >= 3 && randVal < 0.32 {
					enemyType = EnemySnail
				} else if world >= 3 && randVal < 0.40 {
					enemyType = EnemyMouse
				} else if world >= 2 && randVal < 0.50 {
					enemyType = EnemyLadybug
				} else if world >= 2 && randVal < 0.60 {
					enemyType = EnemyFrog
				} else if world >= 2 && randVal < 0.70 {
					enemyType = EnemyFly
				} else if world >= 1 && randVal < 0.85 {
					enemyType = EnemyKoopa
				} else if world >= 1 {
					enemyType = EnemyBarnacle
				}

				spawnY := 8 * TileSize
				if enemyType == EnemyFly || enemyType == EnemyLadybug {
					spawnY = 3 * TileSize
				} else if enemyType == EnemyFish {
					spawnY = 6 * TileSize
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

		// Add spikes
		if x > 20 && rand.Float32() < 0.03 {
			level.spikes = append(level.spikes, &Spike{
				x: float64(x * TileSize),
				y: float64(9*TileSize + 24),
				width: 40,
				height: 16,
				damage: 1,
			})
		}

		// Add springs
		if x > 30 && rand.Float32() < 0.015 {
			springColor := color.RGBA{255, 100, 100, 255}
			if world >= 2 {
				springColor = color.RGBA{100, 255, 100, 255}
			}
			if world >= 4 {
				springColor = color.RGBA{100, 100, 255, 255}
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

		// Add portal pairs
		if x == 50 && world >= 2 && len(level.portals) == 0 {
			portal1 := &Portal{
				x: float64(x * TileSize),
				y: float64(5 * TileSize),
				width: 50,
				height: 70,
				color: color.RGBA{150, 50, 255, 255},
			}
			portal2 := &Portal{
				x: float64((x + 30) * TileSize),
				y: float64(5 * TileSize),
				width: 50,
				height: 70,
				color: color.RGBA{50, 150, 255, 255},
			}
			portal1.linkedTo = portal2
			portal2.linkedTo = portal1
			level.portals = append(level.portals, portal1, portal2)
		}

		// Add crystals (new feature!)
		if x > 15 && rand.Float32() < 0.04 {
			crystalY := rand.Intn(6) + 3
			crystalColors := []color.RGBA{
				{255, 50, 50, 255},   // Red
				{50, 255, 50, 255},   // Green
				{50, 50, 255, 255},   // Blue
				{255, 255, 50, 255},  // Yellow
				{255, 50, 255, 255},  // Purple
			}
			level.crystals = append(level.crystals, Crystal{
				x: float64(x*TileSize + TileSize/2),
				y: float64(crystalY*TileSize + TileSize/2),
				color: crystalColors[rand.Intn(len(crystalColors))],
			})
		}
	}

	// Add flag at end
	level.flagX = (width - 5) * TileSize
	level.flagY = 6 * TileSize

	// Add coins
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
		g.player.crystals = 0
		g.LoadLevel(1)
		playSound(SoundStart)
	}
}

func (g *Game) updatePlaying() {
	g.level.timeElapsed++

	// Check time limit
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

	// Update crystals
	g.updateCrystals()

	// Check spikes
	g.checkSpikes()

	// Update springs
	g.updateSprings()

	// Update portals
	g.updatePortals()

	// Update particles
	g.updateParticles()

	// Update achievements
	g.updateAchievements()

	// Update combo
	g.updateCombo()

	// Check win condition
	if g.player.x >= float64(g.level.flagX) {
		g.state = StateWon
		playSound(SoundWin)

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
		if p.onGround && p.animFrame%20 == 0 {
			playSound(SoundFootstep)
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		if p.vx > -RunSpeed {
			p.vx -= Acceleration
		}
		p.facing = -1
		p.animFrame++
		if p.onGround && p.animFrame%20 == 0 {
			playSound(SoundFootstep)
		}
	} else {
		p.vx *= Friction
		if math.Abs(p.vx) < 0.1 {
			p.vx = 0
		}
		p.animFrame = 0
	}

	// Jumping
	if (ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeySpace)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
		playSound(SoundJump)
		
		// Spawn jump particles
		for i := 0; i < 5; i++ {
			g.particles = append(g.particles, &Particle{
				x: p.x + float64(p.width)/2,
				y: p.y + float64(p.height),
				vx: float64(rand.Float32()*2 - 1),
				vy: float64(rand.Float32() * 2),
				life: 20,
				color: color.RGBA{100, 200, 100, 255},
				size: float32(rand.Intn(4) + 2),
			})
		}
	}

	// Apply gravity
	p.vy += Gravity
	if p.vy > MaxFallSpeed {
		p.vy = MaxFallSpeed
	}

	// Update position
	p.x += p.vx
	p.y += p.vy

	// Collision detection
	g.checkPlayerCollisions()

	// Update powerup timer
	if p.isInvincible {
		p.powerTimer--
		if p.powerTimer <= 0 {
			p.isInvincible = false
		}
	}

	// Update combo timer
	if p.combo > 0 {
		p.comboTimer--
		if p.comboTimer <= 0 {
			p.combo = 0
		}
	}
}

func (g *Game) checkPlayerCollisions() {
	p := g.player

	// Tile collisions
	tileX := int((p.x + float64(p.width)/2) / TileSize)
	tileY := int((p.y + float64(p.height)/2) / TileSize)

	if tileX >= 0 && tileX < g.level.width && tileY >= 0 && tileY < g.level.height {
		tile := g.level.tiles[tileX][tileY]
		if tile != TileAir {
			// Simple collision response
			if p.vy > 0 && p.y+float64(p.height) > float64(tileY*TileSize) {
				p.y = float64(tileY*TileSize) - float64(p.height)
				p.vy = 0
				p.onGround = true
			}
		}
	}

	// Screen boundaries
	if p.x < 0 {
		p.x = 0
		p.vx = 0
	}
	if p.x > float64(g.level.width*TileSize) {
		p.x = float64(g.level.width * TileSize)
		p.vx = 0
	}
}

func (g *Game) updateEnemies() {
	for _, enemy := range g.level.enemies {
		if !enemy.alive || enemy.squashed {
			continue
		}

		// Simple AI: move back and forth
		enemy.x += float64(enemy.facing) * 1.5
		enemy.animFrame++

		// Check collision with player
		if g.checkEnemyPlayerCollision(enemy) {
			// Player jumps on enemy
			if g.player.vy > 0 && g.player.y+float64(g.player.height) < enemy.y+float64(enemy.height)/2 {
				enemy.alive = false
				g.player.vy = JumpForce / 2
				g.player.score += 100 * (g.player.combo + 1)
				g.player.enemiesDefeated++
				g.player.combo++
				g.player.comboTimer = 120
				g.player.lastStompTime = g.frameCount
				playSound(SoundStomp)

				// Spawn particles
				for i := 0; i < 8; i++ {
					g.particles = append(g.particles, &Particle{
						x: enemy.x + float64(enemy.width)/2,
						y: enemy.y + float64(enemy.height)/2,
						vx: float64(rand.Float32()*4 - 2),
						vy: float64(rand.Float32()*4 - 2),
						life: 30,
						color: color.RGBA{100, 255, 100, 255},
						size: float32(rand.Intn(5) + 3),
					})
				}

				g.unlockAchievement("first_blood")
				if g.player.enemiesDefeated >= 10 {
					g.unlockAchievement("enemy_slayer")
				}
			} else if !g.player.isInvincible {
				// Enemy hits player
				g.playerHit()
			}
		}
	}
}

func (g *Game) checkEnemyPlayerCollision(enemy *Enemy) bool {
	return g.player.x < enemy.x+float64(enemy.width) &&
		g.player.x+float64(g.player.width) > enemy.x &&
		g.player.y < enemy.y+float64(enemy.height) &&
		g.player.y+float64(g.player.height) > enemy.y
}

func (g *Game) updateCrystals() {
	for i := range g.level.crystals {
		crystal := &g.level.crystals[i]
		if crystal.collected {
			continue
		}

		crystal.animFrame++

		// Check collision with player
		if g.player.x < crystal.x+20 &&
			g.player.x+float64(g.player.width) > crystal.x-20 &&
			g.player.y < crystal.y+20 &&
			g.player.y+float64(g.player.height) > crystal.y-20 {
			
			crystal.collected = true
			g.player.crystals++
			g.player.score += 50
			playSound(SoundCrystal)

			// Spawn crystal particles
			for i := 0; i < 6; i++ {
				g.particles = append(g.particles, &Particle{
					x: crystal.x,
					y: crystal.y,
					vx: float64(rand.Float32()*3 - 1.5),
					vy: float64(rand.Float32()*3 - 1.5),
					life: 25,
					color: crystal.color,
					size: float32(rand.Intn(4) + 3),
				})
			}

			// Check achievement
			if g.player.crystals >= 50 {
				g.unlockAchievement("crystal_collector")
			}
		}
	}
}

func (g *Game) updatePowerups() {
	for _, powerup := range g.level.powerups {
		if !powerup.alive {
			continue
		}

		powerup.vy += Gravity
		powerup.y += powerup.vy
		powerup.animFrame++

		// Check collision with player
		if g.player.x < powerup.x+float64(powerup.width) &&
			g.player.x+float64(g.player.width) > powerup.x &&
			g.player.y < powerup.y+float64(powerup.height) &&
			g.player.y+float64(g.player.height) > powerup.y {
			
			powerup.alive = false
			g.applyPowerup(powerup.powerType)
		}
	}
}

func (g *Game) applyPowerup(powerType int) {
	playSound(SoundPowerup)
	switch powerType {
	case PowerupMushroom:
		g.player.isBig = true
		g.player.score += 1000
	case PowerupFlower:
		g.player.isFire = true
		g.player.score += 1000
	case PowerupStar:
		g.player.isInvincible = true
		g.player.powerTimer = 600 // 10 seconds
		g.player.score += 1000
	case Powerup1UP:
		g.player.lives++
		g.player.score += 5000
	case PowerupCrystal:
		g.player.crystals += 5
		g.player.score += 500
	}
}

func (g *Game) updateKeys() {
	for _, key := range g.level.keys {
		if key.collected {
			continue
		}

		key.animFrame++

		if g.player.x < key.x+15 &&
			g.player.x+float64(g.player.width) > key.x-15 &&
			g.player.y < key.y+15 &&
			g.player.y+float64(g.player.height) > key.y-15 {
			
			key.collected = true
			g.player.keys++
			g.player.score += 200
			playSound(SoundCoin)
		}
	}
}

func (g *Game) updateChests() {
	for _, chest := range g.level.chests {
		if chest.opened {
			continue
		}

		chest.animFrame++

		if g.player.x < chest.x+float64(chest.width) &&
			g.player.x+float64(g.player.width) > chest.x &&
			g.player.y < chest.y+float64(chest.height) &&
			g.player.y+float64(g.player.height) > chest.y {
			
			if chest.locked && g.player.keys == 0 {
				continue
			}

			chest.opened = true
			chest.locked = false
			g.player.score += chest.value

			switch chest.contents {
			case "coins":
				g.player.coins += chest.value / 10
			case "star":
				g.player.isInvincible = true
				g.player.powerTimer = 600
			case "1up":
				g.player.lives++
			}

			playSound(SoundPowerup)
			g.unlockAchievement("treasure_hunter")
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
				g.playerHit()
			}
		}
	}
}

func (g *Game) updateSprings() {
	for _, spring := range g.level.springs {
		if spring.compressed {
			spring.timer--
			if spring.timer <= 0 {
				spring.compressed = false
			}
		}

		// Check collision with player
		if g.player.vy > 0 &&
			g.player.x < spring.x+float64(spring.width) &&
			g.player.x+float64(g.player.width) > spring.x &&
			g.player.y+float64(g.player.height) > spring.y &&
			g.player.y+float64(g.player.height) < spring.y+float64(spring.height)+10 {
			
			spring.compressed = true
			spring.timer = 10
			g.player.vy = JumpForce * 1.5
			g.player.onGround = false
			playSound(SoundSpring)

			// Spawn particles
			for i := 0; i < 10; i++ {
				g.particles = append(g.particles, &Particle{
					x: spring.x + float64(spring.width)/2,
					y: spring.y,
					vx: float64(rand.Float32()*4 - 2),
					vy: float64(rand.Float32() * -3),
					life: 20,
					color: spring.color,
					size: float32(rand.Intn(4) + 2),
				})
			}
		}
	}
}

func (g *Game) updatePortals() {
	for _, portal := range g.level.portals {
		portal.animFrame++

		if g.player.x < portal.x+float64(portal.width) &&
			g.player.x+float64(g.player.width) > portal.x &&
			g.player.y < portal.y+float64(portal.height) &&
			g.player.y+float64(g.player.height) > portal.y {
			
			if portal.linkedTo != nil {
				g.player.x = portal.linkedTo.x - 20
				g.player.y = portal.linkedTo.y
				playSound(SoundPortal)
			}
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1 // gravity
		p.life--

		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) playerHit() {
	if g.player.isInvincible {
		return
	}

	g.player.damageTaken++
	playSound(SoundDie)

	if g.player.isBig {
		g.player.isBig = false
		g.player.isFire = false
		g.player.isInvincible = true
		g.player.powerTimer = 120
	} else {
		g.playerDie()
	}
}

func (g *Game) playerDie() {
	g.player.lives--
	playSound(SoundDie)

	if g.player.lives <= 0 {
		g.state = StateGameOver
	} else {
		g.LoadLevel(g.player.world)
	}
}

func (g *Game) updateCombo() {
	if g.player.combo > 0 {
		g.player.comboTimer--
		if g.player.comboTimer <= 0 {
			g.player.combo = 0
		}
	}
}

func (g *Game) updateAchievements() {
	// Check coin master
	if g.player.coins >= 100 {
		g.unlockAchievement("coin_master")
	}
}

func (g *Game) unlockAchievement(id string) {
	if achievement, ok := g.achievements[id]; ok {
		if !achievement.unlocked {
			achievement.unlocked = true
			g.newAchievements = append(g.newAchievements, achievement)
		}
	}
}

func (g *Game) updateEndScreen() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StateMenu
		g.LoadLevel(1)
	}
}

// ============================================================================
// RENDERING
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StateGameOver, StateWon:
		g.drawEndScreen(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Clear screen with jungle background
	screen.Fill(color.RGBA{34, 139, 34, 255})

	// Draw title
	if gameAssets != nil && gameAssets.gameFont != nil {
		title := "JUNGLE CRYSTAL ADVENTURE"
		bounds := text.BoundString(gameAssets.gameFont, title)
		x := (ScreenWidth - bounds.Dx()) / 2
		y := 150
		text.Draw(screen, title, gameAssets.gameFont, x, y, color.White)

		instruction := "Press ENTER or SPACE to Start"
		bounds = text.BoundString(gameAssets.gameFont, instruction)
		x = (ScreenWidth - bounds.Dx()) / 2
		y = 300
		text.Draw(screen, instruction, gameAssets.gameFont, x, y, color.RGBA{255, 255, 0, 255})
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Clear screen with jungle sky color
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw background
	g.drawBackground(screen)

	// Apply camera transform
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.camera.x, -g.camera.y)

	// Draw tiles
	g.drawTiles(screen, op)

	// Draw crystals
	g.drawCrystals(screen, op)

	// Draw coins
	g.drawCoins(screen, op)

	// Draw keys
	g.drawKeys(screen, op)

	// Draw chests
	g.drawChests(screen, op)

	// Draw springs
	g.drawSprings(screen, op)

	// Draw portals
	g.drawPortals(screen, op)

	// Draw spikes
	g.drawSpikes(screen, op)

	// Draw enemies
	g.drawEnemies(screen, op)

	// Draw powerups
	g.drawPowerups(screen, op)

	// Draw player
	g.drawPlayer(screen, op)

	// Draw particles
	g.drawParticles(screen, op)

	// Draw flag
	g.drawFlag(screen, op)

	// Draw HUD
	g.drawHUD(screen)

	// Draw new achievements
	g.drawNewAchievements(screen)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	// Draw jungle sky gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(135 - y/10)
		g := uint8(206 - y/15)
		b := uint8(235)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	// Draw sun
	vector.DrawFilledCircle(screen, 700, 80, 40, color.RGBA{255, 255, 0, 255}, true)

	// Draw clouds
	for i := 0; i < 5; i++ {
		x := (i*200 + g.frameCount/60) % (g.level.width*TileSize)
		y := 50 + i*30
		vector.DrawFilledCircle(screen, float32(x), float32(y), 30, color.RGBA{255, 255, 255, 200}, true)
		vector.DrawFilledCircle(screen, float32(x)+25, float32(y)+10, 25, color.RGBA{255, 255, 255, 200}, true)
		vector.DrawFilledCircle(screen, float32(x)-25, float32(y)+10, 25, color.RGBA{255, 255, 255, 200}, true)
	}
}

func (g *Game) drawTiles(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	startX := int(g.camera.x / TileSize)
	endX := startX + ScreenWidth/TileSize + 1

	for x := startX; x < endX && x < g.level.width; x++ {
		for y := 0; y < g.level.height; y++ {
			tile := g.level.tiles[x][y]
			if tile != TileAir {
				if img, ok := tileImages[tile]; ok && img != nil {
					tileOp := &ebiten.DrawImageOptions{}
					tileOp.GeoM.Translate(float64(x*TileSize), float64(y*TileSize))
					screen.DrawImage(img, tileOp)
				} else {
					// Fallback rendering
					tileColor := color.RGBA{139, 69, 19, 255}
					if tile == TileBrick {
						tileColor = color.RGBA{150, 100, 50, 255}
					} else if tile == TileQuestion {
						tileColor = color.RGBA{255, 200, 50, 255}
					} else if tile == TileHard {
						tileColor = color.RGBA{100, 100, 100, 255}
					}
					vector.DrawFilledRect(screen, float32(x*TileSize), float32(y*TileSize), TileSize, TileSize, tileColor, true)
				}
			}
		}
	}
}

func (g *Game) drawCrystals(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, crystal := range g.level.crystals {
		if crystal.collected {
			continue
		}

		// Draw crystal as rotating diamond
		offset := math.Sin(float64(crystal.animFrame)*0.1) * 5
		vector.DrawFilledRect(screen, 
			float32(crystal.x)-10, 
			float32(crystal.y)+float32(offset)-10, 
			20, 20, 
			crystal.color, true)
	}
}

func (g *Game) drawCoins(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, coin := range g.level.coins {
		if coin.collected {
			continue
		}

		coin.animFrame++
		offset := math.Sin(float64(coin.animFrame)*0.1) * 3

		if gameAssets != nil && gameAssets.coinSprite != nil {
			coinOp := &ebiten.DrawImageOptions{}
			coinOp.GeoM.Translate(coin.x, coin.y+offset)
			screen.DrawImage(gameAssets.coinSprite, coinOp)
		} else {
			vector.DrawFilledCircle(screen, float32(coin.x)+20, float32(coin.y)+20+float32(offset), 10, color.RGBA{255, 215, 0, 255}, true)
		}
	}
}

func (g *Game) drawKeys(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, key := range g.level.keys {
		if key.collected {
			continue
		}

		key.animFrame++
		offset := math.Sin(float64(key.animFrame)*0.1) * 3

		vector.DrawFilledRect(screen, float32(key.x)-8, float32(key.y)+float32(offset)-4, 16, 8, color.RGBA{255, 215, 0, 255}, true)
		vector.DrawFilledCircle(screen, float32(key.x), float32(key.y)+float32(offset), 6, color.RGBA{255, 215, 0, 255}, true)
	}
}

func (g *Game) drawChests(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, chest := range g.level.chests {
		if chest.opened {
			// Draw opened chest
			vector.DrawFilledRect(screen, float32(chest.x), float32(chest.y), chest.width, chest.height/2, color.RGBA{139, 90, 43, 255}, true)
		} else {
			// Draw closed chest
			vector.DrawFilledRect(screen, float32(chest.x), float32(chest.y), chest.width, chest.height, color.RGBA{139, 90, 43, 255}, true)
			if chest.locked {
				vector.DrawFilledCircle(screen, float32(chest.x)+float32(chest.width)/2, float32(chest.y)+float32(chest.height)/2, 5, color.RGBA{255, 215, 0, 255}, true)
			}
		}
	}
}

func (g *Game) drawSprings(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, spring := range g.level.springs {
		height := spring.height
		if spring.compressed {
			height = spring.height / 2
		}
		vector.DrawFilledRect(screen, float32(spring.x), float32(spring.y)+spring.height-height, spring.width, height, spring.color, true)
	}
}

func (g *Game) drawPortals(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, portal := range g.level.portals {
		portal.animFrame++
		swirl := math.Sin(float64(portal.animFrame)*0.2) * 10

		vector.DrawFilledRect(screen,
			float32(portal.x)+float32(swirl),
			float32(portal.y),
			portal.width,
			portal.height,
			color.RGBA{uint8(150+swirl*5), uint8(50-swirl*5), 255, 180}, true)
	}
}

func (g *Game) drawSpikes(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, spike := range g.level.spikes {
		// Draw triangle spike
		vector.DrawFilledRect(screen, float32(spike.x), float32(spike.y), spike.width, spike.height, color.RGBA{150, 150, 150, 255}, true)
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, enemy := range g.level.enemies {
		if !enemy.alive || enemy.squashed {
			continue
		}

		var sprite *ebiten.Image
		switch enemy.enemyType {
		case EnemyGoomba:
			sprite = gameAssets.slimeGreen1
		case EnemyKoopa:
			sprite = gameAssets.slimeBlue1
		case EnemyFly:
			sprite = gameAssets.fly1
		case EnemyFrog:
			sprite = gameAssets.frog1
		case EnemyMouse:
			sprite = gameAssets.mouse1
		case EnemySaw:
			sprite = gameAssets.saw1
		default:
			sprite = gameAssets.slimeGreen1
		}

		if sprite != nil {
			enemyOp := &ebiten.DrawImageOptions{}
			enemyOp.GeoM.Translate(enemy.x, enemy.y)
			if enemy.facing == -1 {
				enemyOp.GeoM.Scale(-1, 1)
				enemyOp.GeoM.Translate(-float64(enemy.width), 0)
			}
			screen.DrawImage(sprite, enemyOp)
		} else {
			// Fallback rendering
			enemyColor := color.RGBA{100, 200, 100, 255}
			vector.DrawFilledRect(screen, float32(enemy.x), float32(enemy.y), enemy.width, enemy.height, enemyColor, true)
		}
	}
}

func (g *Game) drawPowerups(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, powerup := range g.level.powerups {
		if !powerup.alive {
			continue
		}

		var puColor color.RGBA
		switch powerup.powerType {
		case PowerupMushroom:
			puColor = color.RGBA{255, 100, 100, 255}
		case PowerupFlower:
			puColor = color.RGBA{255, 200, 50, 255}
		case PowerupStar:
			puColor = color.RGBA{255, 255, 0, 255}
		case Powerup1UP:
			puColor = color.RGBA{50, 255, 50, 255}
		case PowerupCrystal:
			puColor = color.RGBA{100, 100, 255, 255}
		}

		vector.DrawFilledCircle(screen, float32(powerup.x)+powerup.width/2, float32(powerup.y)+powerup.height/2, 15, puColor, true)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	p := g.player

	var sprite *ebiten.Image
	if !p.onGround {
		sprite = gameAssets.playerJump
	} else if math.Abs(p.vx) > 0.5 {
		if (p.animFrame / 10) % 2 == 0 {
			sprite = gameAssets.playerWalk1
		} else {
			sprite = gameAssets.playerWalk2
		}
	} else {
		sprite = gameAssets.playerStand
	}

	if sprite != nil {
		playerOp := &ebiten.DrawImageOptions{}
		playerOp.GeoM.Translate(p.x, p.y)
		if p.facing == -1 {
			playerOp.GeoM.Scale(-1, 1)
			playerOp.GeoM.Translate(-float64(p.width), 0)
		}

		if p.isInvincible && (p.powerTimer/5)%2 == 0 {
			playerOp.ColorM.Scale(1, 1, 1, 0.5)
		}

		screen.DrawImage(sprite, playerOp)
	} else {
		// Fallback rendering
		playerColor := color.RGBA{50, 200, 50, 255}
		if p.isInvincible {
			playerColor = color.RGBA{255, 255, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(p.x), float32(p.y), p.width, p.height, playerColor, true)
	}
}

func (g *Game) drawParticles(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, particle := range g.particles {
		alpha := uint8(255 * particle.life / 30)
		particleColor := color.RGBA{particle.color.R, particle.color.G, particle.color.B, alpha}
		vector.DrawFilledCircle(screen, float32(particle.x), float32(particle.y), particle.size, particleColor, true)
	}
}

func (g *Game) drawFlag(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	// Draw flag pole
	vector.DrawFilledRect(screen, float32(g.level.flagX), float32(g.level.flagY), 5, 150, color.RGBA{100, 100, 100, 255}, true)
	// Draw flag
	vector.DrawFilledRect(screen, float32(g.level.flagX)+5, float32(g.level.flagY), 60, 40, color.RGBA{255, 0, 0, 255}, true)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	if gameAssets == nil || gameAssets.gameFont == nil {
		return
	}

	// Draw score
	scoreText := fmt.Sprintf("SCORE: %d", g.player.score)
	text.Draw(screen, scoreText, gameAssets.gameFont, 10, 30, color.White)

	// Draw coins
	coinText := fmt.Sprintf("COINS: %d", g.player.coins)
	text.Draw(screen, coinText, gameAssets.gameFont, 10, 60, color.RGBA{255, 255, 0, 255})

	// Draw crystals (new!)
	crystalText := fmt.Sprintf("CRYSTALS: %d", g.player.crystals)
	text.Draw(screen, crystalText, gameAssets.gameFont, 10, 90, color.RGBA{100, 100, 255, 255})

	// Draw lives
	livesText := fmt.Sprintf("LIVES: %d", g.player.lives)
	text.Draw(screen, livesText, gameAssets.gameFont, 10, 120, color.RGBA{255, 0, 0, 255})

	// Draw world
	worldText := fmt.Sprintf("WORLD: %d", g.player.world)
	text.Draw(screen, worldText, gameAssets.gameFont, ScreenWidth-150, 30, color.White)

	// Draw time
	timeLeft := g.level.timeLimit - g.level.timeElapsed/60
	timeText := fmt.Sprintf("TIME: %d", timeLeft)
	text.Draw(screen, timeText, gameAssets.gameFont, ScreenWidth-150, 60, color.White)

	// Draw combo
	if g.player.combo > 0 {
		comboText := fmt.Sprintf("COMBO x%d!", g.player.combo)
		text.Draw(screen, comboText, gameAssets.gameFont, ScreenWidth/2-50, 50, color.RGBA{255, 255, 0, 255})
	}
}

func (g *Game) drawNewAchievements(screen *ebiten.Image) {
	if len(g.newAchievements) == 0 {
		return
	}

	if gameAssets == nil || gameAssets.gameFont == nil {
		return
	}

	// Draw latest achievement
	achievement := g.newAchievements[len(g.newAchievements)-1]
	text.Draw(screen, "ACHIEVEMENT UNLOCKED!", gameAssets.gameFont, ScreenWidth/2-150, 200, color.RGBA{255, 255, 0, 255})
	text.Draw(screen, achievement.name, gameAssets.gameFont, ScreenWidth/2-100, 240, color.White)
	text.Draw(screen, achievement.description, gameAssets.gameFont, ScreenWidth/2-120, 270, color.RGBA{200, 200, 200, 255})
}


func (g *Game) drawEndScreen(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if gameAssets == nil || gameAssets.gameFont == nil {
		return
	}

	var message string
	if g.state == StateWon {
		message = "YOU WIN!"
	} else {
		message = "GAME OVER"
	}

	bounds := text.BoundString(gameAssets.gameFont, message)
	x := (ScreenWidth - bounds.Dx()) / 2
	y := 200
	text.Draw(screen, message, gameAssets.gameFont, x, y, color.RGBA{255, 255, 0, 255})

	scoreText := fmt.Sprintf("Final Score: %d", g.player.score)
	bounds = text.BoundString(gameAssets.gameFont, scoreText)
	x = (ScreenWidth - bounds.Dx()) / 2
	y = 300
	text.Draw(screen, scoreText, gameAssets.gameFont, x, y, color.White)

	instruction := "Press ENTER to Continue"
	bounds = text.BoundString(gameAssets.gameFont, instruction)
	x = (ScreenWidth - bounds.Dx()) / 2
	y = 400
	text.Draw(screen, instruction, gameAssets.gameFont, x, y, color.RGBA{200, 200, 200, 255})
}

// ============================================================================
// AUDIO SYSTEM
// ============================================================================

func playSound(soundType SoundType) {
	if !audioPlayer.enabled {
		return
	}

// Simple placeholder - actual audio loading would go here
	// For now, just log the sound
	// fmt.Printf("Playing sound: %v\n", soundType)
}

// Layout implements ebiten.Game.Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Jungle Crystal Adventure - Go365 Day 87")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

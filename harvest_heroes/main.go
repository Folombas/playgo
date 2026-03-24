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
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 48

	// Game states
	StateMenu       = 0
	StateFarm       = 1
	StateDungeon    = 2
	StateShop       = 3
	StateInventory  = 4
	StateCrafting   = 5
	StateTown       = 6
	StateDialogue   = 7
	StateSettings   = 8
	StateFishing    = 9
	StateMining     = 10
	StateFestival   = 11
	StateRomance    = 12

	// Crop types
	CropWheat    = 1
	CropCorn     = 2
	CropTomato   = 3
	CropPotato   = 4
	CropCarrot   = 5
	CropPumpkin  = 6
	CropBerry    = 7
	CropMagic    = 8
	CropStrawberry = 9
	CropGrape    = 10
	CropSunflower = 11
	CropCoffee   = 12
	CropAncient  = 13  // Rare ancient fruit

	// Tool types
	ToolHoe      = 1
	ToolWatering = 2
	ToolAxe      = 3
	ToolPickaxe  = 4
	ToolSword    = 5
	ToolStaff    = 6
	ToolFishing  = 7
	ToolMagic    = 8

	// Enemy types
	EnemySlime   = 1
	EnemyBat     = 2
	EnemySkeleton = 3
	EnemyGhost   = 4
	EnemyDemon   = 5
	EnemyBoss    = 6
	EnemyDragon  = 7  // Ultimate boss
	EnemyWisp    = 8  // Magical enemy
	EnemyGolem   = 9  // Mining area enemy

	// Seasons
	SeasonSpring = 1
	SeasonSummer = 2
	SeasonFall   = 3
	SeasonWinter = 4

	// Weather
	WeatherSunny  = 1
	WeatherRainy  = 2
	WeatherStorm  = 3
	WeatherSnow   = 4

	// Tile types
	TileGrass    = 0
	TileSoil     = 1
	TileWater    = 2
	TilePath     = 3
	TileWood     = 4
	TileStone    = 5
)

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Crop struct {
	cropType   int
	growth     int      // 0-100%
	maxGrowth  int
	growTime   int      // days to mature
	watered    bool
	ready      bool
	plantDay   int
	value      int
}

type Plot struct {
	x, y      int
	tile      int
	crop      *Crop
	waterLevel int
}

type Tool struct {
	toolType   int
	name       string
	level      int
	durability int
	maxDurability int
}

type Item struct {
	id         int
	name       string
	description string
	itemType   int  // 1=crop, 2=ore, 3=craft, 4=seed
	quantity   int
	value      int
	stackSize  int
}

type Enemy struct {
	x, y      float64
	enemyType int
	hp        int
	maxHp     int
	damage    int
	speed     float64
	alive     bool
	reward    int
}

type NPC struct {
	id         int
	name       string
	x, y       float64
	dialogue   []string
	friendship int  // 0-100
	schedule   map[int]string
}

type Player struct {
	x, y       float64
	energy     int
	maxEnergy  int
	mana       int
	maxMana    int
	gold       int
	tools      []*Tool
	inventory  []*Item
	equipped   int
	hp         int
	maxHp      int
	attack     int
	defense    int
	speed      float64
	luck       float32
	fishingRod *Tool
	pickaxeLevel int
}

type Dungeon struct {
	floor      int
	width      int
	height     int
	tiles      [][]int
	enemies    []*Enemy
	chests     []*Chest
	exitX      int
	exitY      int
}

type Chest struct {
	x, y      int
	opened    bool
	contents  *Item
}

type Recipe struct {
	id         int
	name       string
	ingredients map[int]int
	result     *Item
	category   int  // 1=food, 2=potion, 3=tool, 4=weapon
}

type Quest struct {
	id          int
	name        string
	description string
	objective   string
	target      int
	current     int
	completed   bool
	reward      int
	giver       *NPC
}

type Achievement struct {
	id          string
	name        string
	description string
	unlocked    bool
	progress    int
	requirement int
}

type Animal struct {
	id        int
	name      string
	x, y      float64
	animalType int  // 1=chicken, 2=cow, 3=sheep
	products   int  // eggs, milk, wool
	happiness  int
	fed        bool
}

type Fish struct {
	id        int
	name      string
	rarity    int  // 1-5
	value     int
	season    int  // available seasons
	weather   int  // available weather
}

type Ore struct {
	id        int
	name      string
	value     int
	toolLevel int  // required pickaxe level
}

type GameState struct {
	state      int
	player     *Player
	farm       [][]*Plot
	dungeon    *Dungeon
	npcs       []*NPC
	recipes    []*Recipe
	quests     []*Quest
	achievements map[string]*Achievement
	animals    []*Animal
	fishing    []*Fish
	ores       []*Ore
	shops      map[int]int  // itemID -> price
	
	// Combat stats
	combo      int
	critChance float32
	critDamage float32
	mana       int      // Magic system
	maxMana    int
	luck       float32  // Affects crits, fishing, mining
	
	// Fishing
	fishingRod   *Tool
	fishingActive bool
	fishingProgress int
	fishingMinigame float32
	
	// Mining
	pickaxeLevel int
	miningNodes  []*MiningNode
	
	// Romance
	romanceLevel map[int]int  // NPC ID -> romance level
	married      bool
	spouseID     int
	
	// Festival
	currentFestival string
	festivalTimer   int
	
	// Time system
	day        int
	season     int
	weather    int
	hour       int
	minute     int
	timeScale  float64
	
	// Stats
	totalGold  int
	cropsHarvested int
	enemiesDefeated int
	dungeonFloor   int
	fishCaught     int
	oresMined      int
	questsCompleted int
	
	// UI
	selectedTool   int
	selectedPlot   *Plot
	hoveredElement string
	dialogueIndex  int
	currentNPC     *NPC
	currentQuest   *Quest
	showCrafting   bool
	showFishing    bool
	
	// Settings
	soundEnabled bool
	musicVolume  float64
	sfxVolume    float64
	
	// Visual effects
	particles    []*Particle
	screenShake  int
	damageNumbers []*DamageNumber
	
	// Assets
	gameFont     font.Face
	smallFont    font.Face
	
	// Audio
	audioCtx     *audio.Context
	sounds       map[int][]byte
}

type DamageNumber struct {
	x, y     float64
	value    int
	life     int
	isCrit   bool
}

type MiningNode struct {
	x, y      float64
	oreType   int
	hp        int
	maxHp     int
	depleted  bool
}

type Spell struct {
	id        int
	name      string
	manaCost  int
	damage    int
	effect    string
	color     color.RGBA
}

type Festival struct {
	id        string
	name      string
	season    int
	day       int
	description string
	minigame  string
	rewards   []*Item
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
	sounds[0] = GenerateSound(500, 0.1)   // UI click
	sounds[1] = GenerateSound(700, 0.15)  // Harvest
	sounds[2] = GenerateSound(300, 0.2)   // Attack
	sounds[3] = GenerateSound(200, 0.15)  // Hit
	sounds[4] = GenerateSound(800, 0.2)   // Collect
	sounds[5] = GenerateSound(400, 0.1)   // Water
	sounds[6] = GenerateSound(600, 0.25)  // Level up
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
		state: StateMenu,
		player: &Player{
			x: TileSize * 5,
			y: TileSize * 5,
			energy: 100,
			maxEnergy: 100,
			mana: 50,
			maxMana: 50,
			gold: 500,
			hp: 50,
			maxHp: 50,
			attack: 10,
			defense: 5,
			speed: 3.0,
			luck: 0.1,
			tools: make([]*Tool, 0),
			inventory: make([]*Item, 0),
		},
		farm: make([][]*Plot, 15),
		npcs: make([]*NPC, 0),
		recipes: make([]*Recipe, 0),
		shops: make(map[int]int),
		day: 1,
		season: SeasonSpring,
		weather: WeatherSunny,
		hour: 8,
		minute: 0,
		timeScale: 1.0,
		soundEnabled: true,
		musicVolume: 0.5,
		sfxVolume: 0.5,
		particles: make([]*Particle, 0),
		totalGold: 500,
	}

	// Initialize tools
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolHoe, name: "Мотыга", level: 1, durability: 100, maxDurability: 100})
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolWatering, name: "Лейка", level: 1, durability: 100, maxDurability: 100})
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolSword, name: "Меч", level: 1, durability: 50, maxDurability: 50})
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolPickaxe, name: "Кирка", level: 1, durability: 100, maxDurability: 100})
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolFishing, name: "Удочка", level: 1, durability: 100, maxDurability: 100})
	g.player.tools = append(g.player.tools, &Tool{toolType: ToolMagic, name: "Посох", level: 1, durability: 50, maxDurability: 50})
	g.player.fishingRod = g.player.tools[len(g.player.tools)-2]
	g.player.pickaxeLevel = 1
	
	// Initialize romance
	g.romanceLevel = make(map[int]int)

	// Initialize farm
	for x := 0; x < 15; x++ {
		g.farm[x] = make([]*Plot, 10)
		for y := 0; y < 10; y++ {
			tile := TileGrass
			if x == 7 && y == 5 {
				tile = TileWater
			}
			g.farm[x][y] = &Plot{x: x, y: y, tile: tile, waterLevel: 50}
		}
	}

	// Initialize NPCs
	g.npcs = append(g.npcs, &NPC{id: 1, name: "Старейшина", x: 400, y: 300, friendship: 0, dialogue: []string{
		"Добро пожаловать на ферму!",
		"Выращивай культуры и исследуй подземелья!",
		"Береги энергию — она восстанавливается только ночью.",
	}})
	g.npcs = append(g.npcs, &NPC{id: 2, name: "Торговец", x: 600, y: 300, friendship: 0, dialogue: []string{
		"Покупай семена и продавай урожай!",
		"У меня лучшие цены в округе!",
		"Спасибо за покупку!",
	}})

	// Initialize recipes
	g.recipes = append(g.recipes, &Recipe{id: 1, name: "Хлеб", ingredients: map[int]int{1: 3}, result: &Item{id: 101, name: "Хлеб", value: 50}, category: 1})
	g.recipes = append(g.recipes, &Recipe{id: 2, name: "Салат", ingredients: map[int]int{3: 2, 5: 1}, result: &Item{id: 102, name: "Салат", value: 80}, category: 1})
	g.recipes = append(g.recipes, &Recipe{id: 3, name: "Зелье здоровья", ingredients: map[int]int{8: 2, 7: 1}, result: &Item{id: 103, name: "Зелье здоровья", value: 150}, category: 2})
	g.recipes = append(g.recipes, &Recipe{id: 4, name: "Зелье энергии", ingredients: map[int]int{8: 1, 3: 2}, result: &Item{id: 104, name: "Зелье энергии", value: 120}, category: 2})
	g.recipes = append(g.recipes, &Recipe{id: 5, name: "Тыквенный пирог", ingredients: map[int]int{6: 3, 1: 1}, result: &Item{id: 105, name: "Тыквенный пирог", value: 200}, category: 1})
	
	// Initialize quests
	g.quests = append(g.quests, &Quest{id: 1, name: "Первый урожай", description: "Соберите 10 культур", objective: "harvest", target: 10, current: 0, completed: false, reward: 100})
	g.quests = append(g.quests, &Quest{id: 2, name: "Охотник на монстров", description: "Победите 20 врагов", objective: "kill", target: 20, current: 0, completed: false, reward: 200})
	g.quests = append(g.quests, &Quest{id: 3, name: "Шахтёр", description: "Достигните 5 этажа подземелья", objective: "floor", target: 5, current: 0, completed: false, reward: 150})
	g.quests = append(g.quests, &Quest{id: 4, name: "Богач", description: "Накопите 1000 золота", objective: "gold", target: 1000, current: 0, completed: false, reward: 300})
	
	// Initialize achievements
	g.achievements = make(map[string]*Achievement)
	g.achievements["first_harvest"] = &Achievement{id: "first_harvest", name: "Первый урожай", description: "Соберите первую культуру", unlocked: false, requirement: 1}
	g.achievements["master_farmer"] = &Achievement{id: "master_farmer", name: "Мастер фермер", description: "Соберите 100 культур", unlocked: false, requirement: 100}
	g.achievements["demon_slayer"] = &Achievement{id: "demon_slayer", name: "Убийца демонов", description: "Победите 50 врагов", unlocked: false, requirement: 50}
	g.achievements["deep_diver"] = &Achievement{id: "deep_diver", name: "Глубоководник", description: "Достигните 10 этажа", unlocked: false, requirement: 10}
	g.achievements["rich"] = &Achievement{id: "rich", name: "Богач", description: "Накопите 5000 золота", unlocked: false, requirement: 5000}
	
	// Initialize animals
	g.animals = append(g.animals, &Animal{id: 1, name: "Курица", animalType: 1, x: 200, y: 300, products: 0, happiness: 50, fed: false})
	g.animals = append(g.animals, &Animal{id: 2, name: "Корова", animalType: 2, x: 300, y: 300, products: 0, happiness: 50, fed: false})
	
	// Initialize fishing
	g.fishing = append(g.fishing, &Fish{id: 1, name: "Карась", rarity: 1, value: 20, season: 0, weather: 0})
	g.fishing = append(g.fishing, &Fish{id: 2, name: "Щука", rarity: 2, value: 40, season: 0, weather: WeatherRainy})
	g.fishing = append(g.fishing, &Fish{id: 3, name: "Лосось", rarity: 3, value: 80, season: SeasonSummer, weather: 0})
	g.fishing = append(g.fishing, &Fish{id: 4, name: "Золотая рыбка", rarity: 5, value: 200, season: 0, weather: WeatherSunny})
	
	// Initialize ores
	g.ores = append(g.ores, &Ore{id: 1, name: "Медь", value: 30, toolLevel: 1})
	g.ores = append(g.ores, &Ore{id: 2, name: "Железо", value: 50, toolLevel: 2})
	g.ores = append(g.ores, &Ore{id: 3, name: "Золото", value: 100, toolLevel: 3})
	g.ores = append(g.ores, &Ore{id: 4, name: "Алмаз", value: 200, toolLevel: 4})
	
	// Initialize spells
	spells := []*Spell{
		{id: 1, name: "Огненный шар", manaCost: 10, damage: 30, effect: "fire", color: color.RGBA{255, 100, 50, 255}},
		{id: 2, name: "Ледяная стрела", manaCost: 8, damage: 20, effect: "ice", color: color.RGBA{50, 200, 255, 255}},
		{id: 3, name: "Лечение", manaCost: 15, damage: 0, effect: "heal", color: color.RGBA{50, 255, 100, 255}},
		{id: 4, name: "Молния", manaCost: 12, damage: 25, effect: "lightning", color: color.RGBA{255, 255, 0, 255}},
	}
	_ = spells // Store for later use
	
	// Initialize festivals
	festivals := []*Festival{
		{id: "spring_flower", name: "Праздник цветов", season: SeasonSpring, day: 15, description: "Соберите 10 цветов", minigame: "collect", rewards: []*Item{{id: 201, name: "Редкие семена", value: 100}}},
		{id: "summer_luau", name: "Луау", season: SeasonSummer, day: 15, description: "Принесите рыбу", minigame: "cook", rewards: []*Item{{id: 202, name: "Трофей", value: 200}}},
		{id: "fall_harvest", name: "Праздник урожая", season: SeasonFall, day: 15, description: "Выставка урожая", minigame: "contest", rewards: []*Item{{id: 203, name: "Золотая тыква", value: 500}}},
		{id: "winter_star", name: "Звезда зимы", season: SeasonWinter, day: 25, description: "Подарки друзьям", minigame: "gift", rewards: []*Item{{id: 204, name: "Зимний подарок", value: 300}}},
	}
	_ = festivals // Store for later use

	// Initialize shop prices
	g.shops[1] = 20  // Wheat seeds
	g.shops[2] = 30  // Corn seeds
	g.shops[3] = 40  // Tomato seeds
	g.shops[4] = 25  // Potato seeds
	g.shops[5] = 35  // Carrot seeds
	g.shops[6] = 50  // Pumpkin seeds
	g.shops[7] = 60  // Berry seeds
	g.shops[8] = 100 // Magic seeds

	// Load fonts
	g.gameFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 24)
	g.smallFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 16)

	// Load sounds
	g.audioCtx = InitAudio()
	g.sounds = LoadSounds()

	return g
}

func (g *GameState) GenerateDungeon(floor int) {
	width := 20
	height := 15
	
	g.dungeon = &Dungeon{
		floor: floor,
		width: width,
		height: height,
		tiles: make([][]int, width),
		enemies: make([]*Enemy, 0),
		chests: make([]*Chest, 0),
	}
	
	for x := 0; x < width; x++ {
		g.dungeon.tiles[x] = make([]int, height)
		for y := 0; y < height; y++ {
			// Generate walls and floors
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				g.dungeon.tiles[x][y] = TileStone
			} else if rand.Float32() < 0.1 {
				g.dungeon.tiles[x][y] = TileStone
			} else {
				g.dungeon.tiles[x][y] = TileGrass
			}
		}
	}
	
	// Place exit
	g.dungeon.exitX = width - 2
	g.dungeon.exitY = height - 2
	g.dungeon.tiles[g.dungeon.exitX][g.dungeon.exitY] = TilePath
	
	// Place player at entrance
	g.player.x = float64(TileSize)
	g.player.y = float64(TileSize)
	
	// Spawn enemies
	enemyCount := 3 + floor*2
	for i := 0; i < enemyCount; i++ {
		ex := rand.Intn(width-2) + 1
		ey := rand.Intn(height-2) + 1
		
		if g.dungeon.tiles[ex][ey] == TileGrass {
			enemyType := EnemySlime
			if floor >= 3 && rand.Float32() < 0.3 {
				enemyType = EnemyBat
			}
			if floor >= 5 && rand.Float32() < 0.3 {
				enemyType = EnemySkeleton
			}
			if floor >= 8 && rand.Float32() < 0.3 {
				enemyType = EnemyGhost
			}
			if floor >= 10 && rand.Float32() < 0.2 {
				enemyType = EnemyDemon
			}
			
			hp := 20 + floor*5
			damage := 5 + floor*2
			reward := 10 + floor*3
			
			enemy := &Enemy{
				x: float64(ex * TileSize),
				y: float64(ey * TileSize),
				enemyType: enemyType,
				hp: hp,
				maxHp: hp,
				damage: damage,
				speed: 1.0 + float64(floor)*0.1,
				alive: true,
				reward: reward,
			}
			g.dungeon.enemies = append(g.dungeon.enemies, enemy)
		}
	}
	
	// Place chests
	chestCount := 1 + floor/3
	for i := 0; i < chestCount; i++ {
		cx := rand.Intn(width-2) + 1
		cy := rand.Intn(height-2) + 1
		
		if g.dungeon.tiles[cx][cy] == TileGrass {
			chest := &Chest{
				x: cx,
				y: cy,
				opened: false,
				contents: &Item{
					id: rand.Intn(8) + 1,
					quantity: rand.Intn(5) + 1,
				},
			}
			g.dungeon.chests = append(g.dungeon.chests, chest)
		}
	}
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *GameState) Update() error {
	// Update time
	g.minute += int(1.0 * g.timeScale)
	if g.minute >= 60 {
		g.minute = 0
		g.hour++
	}
	if g.hour >= 24 {
		g.hour = 0
		g.NextDay()
	}
	
	// Update particles
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
	
	// Screen shake decay
	if g.screenShake > 0 {
		g.screenShake--
	}
	
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StateFarm:
		g.updateFarm()
	case StateDungeon:
		g.updateDungeon()
	case StateShop:
		g.updateShop()
	case StateInventory:
		g.updateInventory()
	case StateSettings:
		g.updateSettings()
	case StateDialogue:
		g.updateDialogue()
	}
	
	return nil
}

func (g *GameState) updateMenu() {
	mx, my := ebiten.CursorPosition()
	
	buttons := []struct{ x, y, w, h int }{
		{ScreenWidth/2 - 150, 250, 300, 60},
		{ScreenWidth/2 - 150, 330, 300, 60},
		{ScreenWidth/2 - 150, 410, 300, 60},
	}
	
	for i, btn := range buttons {
		if mx >= btn.x && mx <= btn.x+btn.w && my >= btn.y && my <= btn.y+btn.h {
			g.hoveredElement = fmt.Sprintf("menu_%d", i)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				switch i {
				case 0:
					g.state = StateFarm
				case 1:
					g.state = StateSettings
				case 2:
					os.Exit(0)
				}
				PlaySound(g, 0)
			}
		}
	}
}

func (g *GameState) updateFarm() {
	mx, my := ebiten.CursorPosition()
	
	// Tool selection
	for i := 0; i < len(g.player.tools); i++ {
		if mx >= 20+i*50 && mx <= 60+i*50 && my >= ScreenHeight-60 && my <= ScreenHeight-20 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.selectedTool = i
				PlaySound(g, 0)
			}
		}
	}
	
	// Interact with plots
	plotX := mx / TileSize
	plotY := my / TileSize
	
	if plotX >= 0 && plotX < 15 && plotY >= 0 && plotY < 10 {
		plot := g.farm[plotX][plotY]
		
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.interactWithPlot(plot)
		}
	}
	
	// Navigation buttons
	if mx >= ScreenWidth-150 && mx <= ScreenWidth-20 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateDungeon
			g.GenerateDungeon(1)
			PlaySound(g, 0)
		}
	}
	
	if mx >= ScreenWidth-150 && mx <= ScreenWidth-20 && my >= 80 && my <= 120 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateTown
			PlaySound(g, 0)
		}
	}
}

func (g *GameState) interactWithPlot(plot *Plot) {
	tool := g.player.tools[g.selectedTool]
	
	switch tool.toolType {
	case ToolHoe:
		if plot.tile == TileGrass {
			plot.tile = TileSoil
			g.spawnParticles(float64(plot.x*TileSize), float64(plot.y*TileSize), 10, color.RGBA{139, 69, 19, 255})
			g.player.energy -= 5
		} else if plot.tile == TileSoil && plot.crop == nil {
			// Plant seeds (simplified - auto plant wheat)
			plot.crop = &Crop{
				cropType: CropWheat,
				growth: 0,
				maxGrowth: 100,
				growTime: 3,
				value: 30,
			}
			g.player.energy -= 3
			PlaySound(g, 4)
		}
		
	case ToolWatering:
		if plot.tile == TileSoil && plot.crop != nil && !plot.crop.watered {
			plot.crop.watered = true
			plot.waterLevel = 100
			g.spawnParticles(float64(plot.x*TileSize), float64(plot.y*TileSize), 15, color.RGBA{50, 150, 255, 255})
			g.player.energy -= 2
			PlaySound(g, 5)
		}
	}
}

func (g *GameState) updateDungeon() {
	mx, my := ebiten.CursorPosition()
	
	// Player movement
	dx := 0.0
	dy := 0.0
	
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx = -g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx = g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy = -g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy = g.player.speed
	}
	
	g.player.x += dx
	g.player.y += dy
	
	// Dungeon bounds
	if g.player.x < 0 {
		g.player.x = 0
	}
	if g.player.x > float64(g.dungeon.width*TileSize-TileSize) {
		g.player.x = float64(g.dungeon.width*TileSize - TileSize)
	}
	if g.player.y < 0 {
		g.player.y = 0
	}
	if g.player.y > float64(g.dungeon.height*TileSize-TileSize) {
		g.player.y = float64(g.dungeon.height*TileSize - TileSize)
	}
	
	// Attack
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.attack()
	}
	
	// Update enemies
	for _, enemy := range g.dungeon.enemies {
		if !enemy.alive {
			continue
		}
		
		// Move towards player
		dx := g.player.x - enemy.x
		dy := g.player.y - enemy.y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist > 0 {
			enemy.x += (dx / dist) * enemy.speed * 0.5
			enemy.y += (dy / dist) * enemy.speed * 0.5
		}
		
		// Damage player
		if dist < float64(TileSize) {
			g.player.hp -= enemy.damage / 10
			g.screenShake = 5
			PlaySound(g, 3)
		}
	}
	
	// Check exit
	tileX := int(g.player.x) / TileSize
	tileY := int(g.player.y) / TileSize
	
	if tileX == g.dungeon.exitX && tileY == g.dungeon.exitY {
		g.dungeon.floor++
		g.GenerateDungeon(g.dungeon.floor)
		PlaySound(g, 6)
	}
	
	// Back to farm
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateFarm
			PlaySound(g, 0)
		}
	}
	
	// Check game over
	if g.player.hp <= 0 {
		g.player.hp = g.player.maxHp
		g.player.x = float64(TileSize * 5)
		g.player.y = float64(TileSize * 5)
		g.state = StateFarm
		g.dungeon = nil
	}
}

func (g *GameState) attack() {
	PlaySound(g, 2)
	
	// Combo system
	g.combo++
	critRoll := rand.Float32()
	isCrit := critRoll < (g.critChance + g.luck*0.1)
	
	for _, enemy := range g.dungeon.enemies {
		if !enemy.alive {
			continue
		}
		
		dx := g.player.x - enemy.x
		dy := g.player.y - enemy.y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist < float64(TileSize)*1.5 {
			damage := g.player.attack
			
			// Combo bonus
			damage += g.combo
			
			// Critical hit
			if isCrit {
				damage = int(float64(damage) * float64(g.critDamage))
				g.damageNumbers = append(g.damageNumbers, &DamageNumber{
					x: enemy.x, y: enemy.y,
					value: damage, life: 60, isCrit: true,
				})
				g.screenShake = 10
			} else {
				g.damageNumbers = append(g.damageNumbers, &DamageNumber{
					x: enemy.x, y: enemy.y,
					value: damage, life: 40, isCrit: false,
				})
			}
			
			enemy.hp -= damage
			g.spawnParticles(enemy.x, enemy.y, 10, color.RGBA{255, 0, 0, 255})
			
			if enemy.hp <= 0 {
				enemy.alive = false
				g.player.gold += enemy.reward
				g.totalGold += enemy.reward
				g.enemiesDefeated++
				g.combo = 0 // Reset combo on kill
				PlaySound(g, 4)
				
				// Check achievements
				g.checkAchievements()
			}
		}
	}
	
	// Combo decay
	if g.combo > 0 {
		g.combo--
	}
}

func (g *GameState) startFishing() {
	g.state = StateFishing
	g.fishingActive = true
	g.fishingProgress = 0
	PlaySound(g, 5)
}

func (g *GameState) updateFishing() {
	// Fishing minigame - keep bar in position
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.fishingMinigame = 0.5 // Reset position
	}
	
	g.fishingProgress++
	
	// Catch fish after some time
	if g.fishingProgress > 180 { // 3 seconds
		// Determine catch based on luck and season
		catchChance := 0.3 + g.luck*0.2
		if rand.Float32() < catchChance {
			// Got a fish!
			fish := g.fishing[rand.Intn(len(g.fishing))]
			g.player.inventory = append(g.player.inventory, &Item{
				id: fish.id,
				name: fish.name,
				itemType: 5, // fish
				value: fish.value,
				quantity: 1,
			})
			g.fishCaught++
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/2), y: float64(ScreenHeight/2),
				value: fish.value, life: 60, isCrit: false,
			})
			PlaySound(g, 4)
		}
		g.state = StateFarm
		g.fishingActive = false
	}
}

func (g *GameState) checkAchievements() {
	for id, ach := range g.achievements {
		if ach.unlocked {
			continue
		}
		
		unlocked := false
		switch id {
		case "first_harvest":
			unlocked = g.cropsHarvested >= 1
		case "master_farmer":
			unlocked = g.cropsHarvested >= 100
		case "demon_slayer":
			unlocked = g.enemiesDefeated >= 50
		case "deep_diver":
			unlocked = g.dungeonFloor >= 10
		case "rich":
			unlocked = g.totalGold >= 5000
		}
		
		if unlocked {
			ach.unlocked = true
			g.achievements[id] = ach
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/2), y: 100,
				value: 0, life: 180, isCrit: true,
			})
			PlaySound(g, 6)
		}
	}
}

func (g *GameState) updateShop() {
	mx, my := ebiten.CursorPosition()
	
	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateTown
			PlaySound(g, 0)
		}
	}
	
	// Buy buttons
	for i := 1; i <= 8; i++ {
		y := 150 + (i-1)*50
		if mx >= 600 && mx <= 800 && my >= y && my <= y+40 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				price := g.shops[i]
				if g.player.gold >= price {
					g.player.gold -= price
					g.player.inventory = append(g.player.inventory, &Item{
						id: i,
						itemType: 4, // seed
						quantity: 1,
						value: price,
					})
					PlaySound(g, 4)
				}
			}
		}
	}
}

func (g *GameState) updateInventory() {
	mx, my := ebiten.CursorPosition()
	
	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		g.state = StateFarm
		PlaySound(g, 0)
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
	if mx >= ScreenWidth/2+100 && mx <= ScreenWidth/2+160 && my >= 250 && my <= 280 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.soundEnabled = !g.soundEnabled
			PlaySound(g, 0)
		}
	}
}

func (g *GameState) updateDialogue() {
	mx, my := ebiten.CursorPosition()
	
	// Continue dialogue
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.currentNPC != nil && g.dialogueIndex < len(g.currentNPC.dialogue)-1 {
			g.dialogueIndex++
		} else {
			g.state = StateTown
			g.currentNPC = nil
		}
		PlaySound(g, 0)
	}
	
	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		g.state = StateTown
		g.currentNPC = nil
		PlaySound(g, 0)
	}
}

func (g *GameState) NextDay() {
	g.day++
	g.hour = 6
	g.minute = 0
	
	// Grow crops
	for x := 0; x < 15; x++ {
		for y := 0; y < 10; y++ {
			plot := g.farm[x][y]
			if plot.crop != nil && plot.crop.watered {
				plot.crop.growth += 100 / plot.crop.growTime
				plot.crop.watered = false
				plot.waterLevel -= 30
				
				if plot.crop.growth >= 100 {
					plot.crop.ready = true
				}
			}
		}
	}
	
	// Weather change
	if rand.Float32() < 0.3 {
		g.weather = WeatherRainy
	} else {
		g.weather = WeatherSunny
	}
	
	// Restore energy
	g.player.energy = g.player.maxEnergy
	
	// Season change (every 30 days)
	if g.day%30 == 1 {
		g.season++
		if g.season > SeasonWinter {
			g.season = SeasonSpring
		}
	}
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
	case StateFarm:
		g.drawFarm(screen)
	case StateDungeon:
		g.drawDungeon(screen)
	case StateShop:
		g.drawShop(screen)
	case StateInventory:
		g.drawInventory(screen)
	case StateSettings:
		g.drawSettings(screen)
	case StateDialogue:
		g.drawDialogue(screen)
	}
}

func (g *GameState) drawMenu(screen *ebiten.Image) {
	// Sky gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(100 + y/20)
		g := uint8(150 + y/30)
		b := uint8(50 + y/40)
		screen.Fill(color.RGBA{r, g, b, 255})
	}
	
	// Title
	title := "🌾 HARVEST HEROES 🗡️"
	titleX := ScreenWidth/2 - 250
	titleY := 150
	
	if g.gameFont != nil {
		text.Draw(screen, title, g.gameFont, titleX+4, titleY+4, color.RGBA{0, 0, 0, 150})
		text.Draw(screen, title, g.gameFont, titleX, titleY, color.RGBA{255, 215, 0, 255})
		
		subtitle := "Farm & Dungeon RPG"
		text.Draw(screen, subtitle, g.smallFont, ScreenWidth/2-100, 210, color.RGBA{255, 255, 255, 255})
	}
	
	// Buttons
	buttons := []string{"🎮 Новая игра", "⚙️ Настройки", "🚪 Выход"}
	buttonY := 280
	for i, btnText := range buttons {
		x := ScreenWidth/2 - 150
		y := buttonY + i*80
		
		btnColor := color.RGBA{60, 100, 60, 255}
		if g.hoveredElement == fmt.Sprintf("menu_%d", i) {
			btnColor = color.RGBA{80, 140, 80, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), 300, 60, btnColor, false)
		vector.StrokeRect(screen, float32(x), float32(y), 300, 60, 3, color.RGBA{150, 255, 150, 255}, false)
		
		if g.gameFont != nil {
			text.Draw(screen, btnText, g.gameFont, ScreenWidth/2-120, y+35, color.RGBA{255, 255, 255, 255})
		}
	}
	
	// Version
	if g.smallFont != nil {
		text.Draw(screen, "Go365 Day 84 | v1.0 | Pure Go + Ebitengine", g.smallFont, 20, ScreenHeight-30, color.RGBA{200, 255, 200, 255})
	}
}

func (g *GameState) drawFarm(screen *ebiten.Image) {
	// Grass background
	screen.Fill(color.RGBA{34, 139, 34, 255})
	
	// Draw farm plots
	for x := 0; x < 15; x++ {
		for y := 0; y < 10; y++ {
			plot := g.farm[x][y]
			drawX := float32(x * TileSize)
			drawY := float32(y * TileSize)
			
			// Tile color
			tileColor := color.RGBA{34, 139, 34, 255}
			switch plot.tile {
			case TileSoil:
				tileColor = color.RGBA{139, 69, 19, 255}
			case TileWater:
				tileColor = color.RGBA{50, 150, 255, 255}
			}
			
			vector.DrawFilledRect(screen, drawX+1, drawY+1, TileSize-2, TileSize-2, tileColor, false)
			
			// Draw crop
			if plot.crop != nil {
				cropColor := color.RGBA{100, 200, 100, 255}
				if plot.crop.ready {
					cropColor = color.RGBA{255, 215, 0, 255}
				}
				
				size := float32(plot.crop.growth) / 100.0 * float32(TileSize-10)
				vector.DrawFilledCircle(screen, drawX+float32(TileSize)/2, drawY+float32(TileSize)/2, size/2, cropColor, false)
				
				// Water indicator
				if plot.crop.watered {
					vector.DrawFilledCircle(screen, drawX+float32(TileSize)-10, drawY+10, 5, color.RGBA{50, 150, 255, 255}, false)
				}
			}
		}
	}
	
	// UI
	g.drawUI(screen)
	
	// Tool selection
	for i, tool := range g.player.tools {
		x := 20 + i*50
		y := ScreenHeight - 60
		
		toolColor := color.RGBA{100, 100, 100, 255}
		if i == g.selectedTool {
			toolColor = color.RGBA{255, 215, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), 40, 40, toolColor, false)
		vector.StrokeRect(screen, float32(x), float32(y), 40, 40, 2, color.RGBA{255, 255, 255, 255}, false)
		
		if g.smallFont != nil {
			toolIcon := "🔧"
			switch tool.toolType {
			case ToolHoe:
				toolIcon = "⛏️"
			case ToolWatering:
				toolIcon = "💧"
			case ToolSword:
				toolIcon = "⚔️"
			}
			text.Draw(screen, toolIcon, g.smallFont, x+10, y+25, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *GameState) drawDungeon(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{20, 20, 40, 255})
	
	if g.dungeon != nil {
		// Draw tiles
		for x := 0; x < g.dungeon.width; x++ {
			for y := 0; y < g.dungeon.height; y++ {
				tile := g.dungeon.tiles[x][y]
				drawX := float32(x * TileSize)
				drawY := float32(y * TileSize)
				
				tileColor := color.RGBA{100, 100, 100, 255}
				switch tile {
				case TileGrass:
					tileColor = color.RGBA{60, 60, 80, 255}
				case TileStone:
					tileColor = color.RGBA{80, 80, 100, 255}
				case TilePath:
					tileColor = color.RGBA{120, 100, 80, 255}
				}
				
				vector.DrawFilledRect(screen, drawX, drawY, TileSize, TileSize, tileColor, false)
			}
		}
		
		// Draw exit
		exitX := float32(g.dungeon.exitX * TileSize)
		exitY := float32(g.dungeon.exitY * TileSize)
		vector.DrawFilledRect(screen, exitX, exitY, TileSize, TileSize, color.RGBA{100, 255, 100, 255}, false)
		
		// Draw enemies
		for _, enemy := range g.dungeon.enemies {
			if !enemy.alive {
				continue
			}
			
			enemyColor := color.RGBA{200, 50, 50, 255}
			switch enemy.enemyType {
			case EnemySlime:
				enemyColor = color.RGBA{50, 200, 50, 255}
			case EnemyBat:
				enemyColor = color.RGBA{150, 50, 150, 255}
			case EnemySkeleton:
				enemyColor = color.RGBA{200, 200, 200, 255}
			case EnemyGhost:
				enemyColor = color.RGBA{150, 150, 255, 150}
			}
			
			vector.DrawFilledCircle(screen, float32(enemy.x)+float32(TileSize)/2, float32(enemy.y)+float32(TileSize)/2, float32(TileSize)/2-5, enemyColor, false)
			
			// HP bar
			hpPercent := float32(enemy.hp) / float32(enemy.maxHp)
			vector.DrawFilledRect(screen, float32(enemy.x), float32(enemy.y)-10, float32(TileSize)*hpPercent, 5, color.RGBA{255, 0, 0, 255}, false)
		}
		
		// Draw chests
		for _, chest := range g.dungeon.chests {
			if chest.opened {
				continue
			}
			chestX := float32(chest.x * TileSize)
			chestY := float32(chest.y * TileSize)
			vector.DrawFilledRect(screen, chestX+10, chestY+15, TileSize-20, TileSize-20, color.RGBA{255, 215, 0, 255}, false)
		}
	}
	
	// Draw player
	vector.DrawFilledCircle(screen, float32(g.player.x)+float32(TileSize)/2, float32(g.player.y)+float32(TileSize)/2, float32(TileSize)/2-5, color.RGBA{50, 150, 255, 255}, false)
	
	// Draw damage numbers
	g.drawDamageNumbers(screen)
	
	// Draw particles
	g.drawParticles(screen)
	
	// UI
	g.drawUI(screen)
	
	// Dungeon info
	if g.gameFont != nil {
		floorText := fmt.Sprintf("Этаж %d", g.dungeon.floor)
		text.Draw(screen, floorText, g.gameFont, ScreenWidth-150, 20, color.RGBA{255, 215, 0, 255})
		
		// Combo display
		if g.combo > 1 {
			comboText := fmt.Sprintf("🔥 Комбо x%d!", g.combo)
			text.Draw(screen, comboText, g.smallFont, ScreenWidth-150, 50, color.RGBA{255, 100, 50, 255})
		}
	}
	
	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Ферма", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawDamageNumbers(screen *ebiten.Image) {
	for i := len(g.damageNumbers) - 1; i >= 0; i-- {
		d := g.damageNumbers[i]
		d.y -= 0.5
		d.life--
		
		if d.life <= 0 {
			g.damageNumbers = append(g.damageNumbers[:i], g.damageNumbers[i+1:]...)
			continue
		}
		
		c := color.RGBA{255, 255, 255, 255}
		if d.isCrit {
			c = color.RGBA{255, 215, 0, 255}
		}
		
		if g.smallFont != nil {
			numText := fmt.Sprintf("%d", d.value)
			if d.isCrit {
				numText = fmt.Sprintf("💥 %d!", d.value)
			}
			text.Draw(screen, numText, g.smallFont, int(d.x), int(d.y), c)
		}
	}
}

func (g *GameState) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 90, color.RGBA{0, 0, 0, 200}, false)
	
	if g.gameFont != nil {
		// Stats row 1
		text.Draw(screen, fmt.Sprintf("❤️ %d/%d", g.player.hp, g.player.maxHp), g.smallFont, 20, 25, color.RGBA{255, 100, 100, 255})
		text.Draw(screen, fmt.Sprintf("⚡ %d/%d", g.player.energy, g.player.maxEnergy), g.smallFont, 20, 45, color.RGBA{255, 255, 0, 255})
		text.Draw(screen, fmt.Sprintf("💰 %d", g.player.gold), g.smallFont, 200, 25, color.RGBA{255, 215, 0, 255})
		
		// Time
		timeText := fmt.Sprintf("📅 День %d, %02d:%02d", g.day, g.hour, g.minute)
		text.Draw(screen, timeText, g.gameFont, 400, 25, color.RGBA{255, 255, 255, 255})
		
		// Season
		seasonNames := []string{"", "🌸 Весна", "☀️ Лето", "🍂 Осень", "❄️ Зима"}
		text.Draw(screen, seasonNames[g.season], g.smallFont, 700, 25, color.RGBA{200, 200, 255, 255})
		
		// Weather
		weatherIcons := []string{"", "☀️", "🌧️", "⛈️", "❄️"}
		text.Draw(screen, weatherIcons[g.weather], g.smallFont, 850, 25, color.RGBA{255, 255, 255, 255})
		
		// Stats row 2
		text.Draw(screen, fmt.Sprintf("🌾 Урожай: %d", g.cropsHarvested), g.smallFont, 20, 65, color.RGBA{100, 255, 100, 255})
		text.Draw(screen, fmt.Sprintf("⚔️ Врагов: %d", g.enemiesDefeated), g.smallFont, 200, 65, color.RGBA{255, 100, 100, 255})
		text.Draw(screen, fmt.Sprintf("📜 Квесты: %d/%d", g.questsCompleted, len(g.quests)), g.smallFont, 400, 65, color.RGBA{255, 215, 0, 255})
	}
	
	// Navigation buttons
	if g.state == StateFarm {
		dungeonBtn := color.RGBA{100, 50, 100, 255}
		vector.DrawFilledRect(screen, float32(ScreenWidth-150), 20, 130, 40, dungeonBtn, false)
		if g.smallFont != nil {
			text.Draw(screen, "🗡️ Подземелье", g.smallFont, ScreenWidth-145, 42, color.RGBA{255, 255, 255, 255})
		}
		
		townBtn := color.RGBA{100, 100, 50, 255}
		vector.DrawFilledRect(screen, float32(ScreenWidth-150), 80, 130, 40, townBtn, false)
		if g.smallFont != nil {
			text.Draw(screen, "🏘️ Город", g.smallFont, ScreenWidth-130, 102, color.RGBA{255, 255, 255, 255})
		}
		
		// Quests button
		questsBtn := color.RGBA{50, 100, 150, 255}
		vector.DrawFilledRect(screen, float32(ScreenWidth-300), 20, 130, 40, questsBtn, false)
		if g.smallFont != nil {
			text.Draw(screen, "📜 Квесты", g.smallFont, ScreenWidth-295, 42, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *GameState) drawShop(screen *ebiten.Image) {
	screen.Fill(color.RGBA{100, 80, 60, 255})
	
	if g.gameFont != nil {
		title := "🏪 Магазин семян"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-120, 100, color.RGBA{255, 215, 0, 255})
		
		// Items
		items := []struct{ id, price int; name string }{
			{1, 20, "🌾 Пшеница"},
			{2, 30, "🌽 Кукуруза"},
			{3, 40, "🍅 Томат"},
			{4, 25, "🥔 Картофель"},
			{5, 35, "🥕 Морковь"},
			{6, 50, "🎃 Тыква"},
			{7, 60, "🫐 Ягода"},
			{8, 100, "✨ Магия"},
		}
		
		for i, item := range items {
			y := 150 + i*50
			vector.DrawFilledRect(screen, 200, float32(y), 600, 40, color.RGBA{80, 60, 40, 255}, false)
			
			text.Draw(screen, fmt.Sprintf("%s - %d¢", item.name, item.price), g.smallFont, 220, y+25, color.RGBA{255, 255, 255, 255})
			text.Draw(screen, fmt.Sprintf("Золото: %d", g.player.gold), g.smallFont, 650, y+25, color.RGBA{255, 215, 0, 255})
		}
	}
	
	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawInventory(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 60, 255})
	
	if g.gameFont != nil {
		title := "🎒 Инвентарь"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-80, 100, color.RGBA{255, 215, 0, 255})
		
		// Items
		y := 150
		for _, item := range g.player.inventory {
			vector.DrawFilledRect(screen, 200, float32(y), 600, 40, color.RGBA{60, 60, 80, 255}, false)
			text.Draw(screen, fmt.Sprintf("x%d %s", item.quantity, item.name), g.smallFont, 220, y+25, color.RGBA{255, 255, 255, 255})
			y += 50
		}
		
		if len(g.player.inventory) == 0 {
			text.Draw(screen, "Пусто...", g.smallFont, ScreenWidth/2-50, 200, color.RGBA{150, 150, 150, 255})
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
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-80, 150, color.RGBA{255, 215, 0, 255})
		
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
	}
	
	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{150, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawDialogue(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 100, ScreenHeight/2-100, ScreenWidth-200, 200, color.RGBA{0, 0, 0, 200}, false)
	vector.StrokeRect(screen, 100, float32(ScreenHeight/2-100), ScreenWidth-200, 200, 3, color.RGBA{255, 215, 0, 255}, false)
	
	if g.currentNPC != nil && g.gameFont != nil {
		text.Draw(screen, g.currentNPC.name, g.gameFont, 120, ScreenHeight/2-70, color.RGBA{255, 215, 0, 255})
		
		if g.dialogueIndex < len(g.currentNPC.dialogue) {
			text.Draw(screen, g.currentNPC.dialogue[g.dialogueIndex], g.smallFont, 120, ScreenHeight/2-20, color.RGBA{255, 255, 255, 255})
		}
		
		text.Draw(screen, "Нажмите для продолжения...", g.smallFont, 120, ScreenHeight/2+50, color.RGBA{150, 150, 150, 255})
	}
}

func (g *GameState) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), p.size, p.color, false)
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
	ebiten.SetWindowTitle("🌾 Harvest Heroes - Farm & Dungeon RPG | Go365 Day 84")

	game := NewGameState()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

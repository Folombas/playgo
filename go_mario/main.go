package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"

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

	// World dimensions
	worldWidth  = 4000
	worldHeight = 800

	// Block size for terrain
	blockSize = 40

	// Inventory slots
	inventorySize = 9
)

// BlockType - тип блока
type BlockType int

const (
	Air BlockType = iota
	Dirt
	Grass
	Stone
	Wood
	Leaves
	Sand
	Coal_Ore
	Iron_Ore
	Gold_Ore
	Diamond_Ore
	Bricks
	Plank
	Crafting_Table
)

// Block - блок мира
type Block struct {
	x, y    int
	typ     BlockType
	solid   bool
	minable bool
}

// Inventory - инвентарь игрока
type Inventory struct {
	slots []InventorySlot
	selected int
}

// InventorySlot - слот инвентаря
type InventorySlot struct {
	item     BlockType
	count    int
	maxStack int
}

// Recipe - рецепт крафта
type Recipe struct {
	result   BlockType
	count    int
	ingredients map[BlockType]int
}

// Camera - камера для слежения за игроком
type Camera struct {
	x, y float64
}

// World - игровой мир
type World struct {
	blocks  [][]Block
	width   int
	height  int
	seed    int64
}

// Coin - монета для сбора
type Coin struct {
	x        float32
	y        float32
	collected bool
	animFrame int
}

// Enemy - враг (Goomba-style)
type Enemy struct {
	x         float32
	y         float32
	width     float32
	height    float32
	vx        float32
	onGround  bool
	animFrame int
	alive     bool
}

// Platform - платформа для прыжков
type Platform struct {
	x      float32
	y      float32
	width  float32
	height float32
}

// JumpParticle - частица прыжка
type JumpParticle struct {
	x       float32
	y       float32
	vx      float32
	vy      float32
	size    float32
	life    int
	maxLife int
}

// PowerupType - тип бонуса
type PowerupType int

const (
	MushroomPower PowerupType = iota // Гриб - увеличивает размер
	StarPower                        // Звезда - неуязвимость
	ExtraLifePower                   // Дополнительная жизнь
)

// Powerup - бонусный предмет
type Powerup struct {
	x         float32
	y         float32
	vy        float32
	px        float32 // для движения гриба
	powerType PowerupType
	collected bool
	animFrame int
	onGround  bool
}

// Player - расширенная структура игрока
type Player struct {
	x         float64
	y         float64
	vy        float64
	width     float32
	height    float32
	onGround  bool
	score     int
	coins     int
	lives     int
	facing    int
	animFrame int
	invincible int // кадры неуязвимости
}

type GameState int

const (
	Menu GameState = iota
	Playing
	InsideHouse
	GameWon
	Crafting  // Режим крафта
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

// TutorialStep - шаг обучения
type TutorialStep struct {
	id          int
	title       string
	description string
	completed   bool
	triggerFunc func(*Game) bool // Условие выполнения
}

// Tutorial - система обучения
type Tutorial struct {
	steps       []TutorialStep
	currentStep int
	visible     bool
	showHint    bool
	hintText    string
	hintTimer   int
}

// Quest - задание/квест
type Quest struct {
	id          int
	title       string
	description string
	objective   string
	completed   bool
	reward      int
}

// AchievementType - тип достижения
type AchievementType int

const (
	FirstSteps AchievementType = iota
	BlockMiner
	CoinCollector
	EnemySlayer
	DiamondFinder
	WorldExplorer
	Builder
	SpeedRunner
	Survivor
	Champion
)

// MedalTier - уровень медали
type MedalTier int

const (
	Bronze MedalTier = iota
	Silver
	Gold
	Platinum
	Diamond
)

// Achievement - достижение/ачивка
type Achievement struct {
	id            int
	achType       AchievementType
	title         string
	description   string
	medalTier     MedalTier
	completed     bool
	unlockedAt    int
	progress      int
	maxProgress   int
	icon          string
}

// AchievementAlbum - альбом достижений
type AchievementAlbum struct {
	achievements []Achievement
	totalUnlocked int
	showAlbum    bool
	pendingNotifications []Achievement
}

// AchievementNotification - активное уведомление ачивки
type AchievementNotification struct {
	achievement Achievement
	life        int
	maxLife     int
	scale       float32
	y           float32
	targetY     float32
}

// Checkpoint - контрольная точка возрождения
type Checkpoint struct {
	x        float32
	y        float32
	activated bool
}

// HealthPack - аптечка для лечения
type HealthPack struct {
	x        float32
	y        float32
	vy       float32
	healAmount int
	collected bool
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

// SparkParticle - частица-искра для эффектов
type SparkParticle struct {
	x       float32
	y       float32
	vx      float32
	vy      float32
	size    float32
	life    int
	maxLife int
	color   color.RGBA
}

// FloatingText - всплывающий текст
type FloatingText struct {
	x       float32
	y       float32
	text    string
	life    int
	maxLife int
	vy      float32
	color   color.RGBA
	scale   float32
}

// ScreenShake - тряска экрана
type ScreenShake struct {
	active  bool
	intensity float32
	timer   int
}

type Game struct {
	playerX       float64
	playerY       float64
	frameCount    int
	clouds        []Cloud
	trees         []Tree
	player        Player
	state         GameState
	timeOfDay     TimeOfDay
	weather       Weather
	stars         []Star
	moonX         float32
	moonY         float32
	raindrops     []Raindrop
	lightning     Lightning
	stormClouds   []Cloud
	house         House
	audio         *AudioSystem
	coins         []Coin
	enemies       []Enemy
	platforms     []Platform
	jumpParticles []JumpParticle
	powerups      []Powerup
	sparkParticles []SparkParticle
	floatingTexts []FloatingText
	screenShake   ScreenShake
	
	// New fields for expanded world and crafting
	world    *World
	camera   *Camera
	inventory *Inventory
	recipes   []Recipe
	
	// Tutorial and quests
	tutorial    *Tutorial
	quests      []Quest
	checkpoints []Checkpoint
	healthPacks []HealthPack
	
	// Achievement album
	album       *AchievementAlbum
	activeNotifications []AchievementNotification
	blocksMined int
	enemiesDefeated int
	
	// Tutorial hints
	showControls    bool
	controlsTimer   int
}

// SoundEffect - структура звукового эффекта
type SoundEffect struct {
	frequency float64
	duration  int
	volume    float64
	waveform  int // 0=sine, 1=square, 2=sawtooth, 3=noise
	sliding   bool
}

// AudioStream - поток для воспроизведения звука
type AudioStream struct {
	buffer []float64
	pos    int
}

func (a *AudioStream) Read(b []byte) (int, error) {
	if a.pos >= len(a.buffer) {
		return 0, io.EOF
	}

	n := 0
	for i := 0; i < len(b); i += 4 {
		if a.pos >= len(a.buffer) {
			break
		}
		sample := a.buffer[a.pos]
		a.pos++

		// Convert float64 [-1,1] to int16
		val := int16(sample * 32767)
		b[i] = byte(val)
		b[i+1] = byte(val >> 8)
		b[i+2] = 0
		b[i+3] = 0
		n += 4
	}

	return n, nil
}

func (a *AudioStream) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		a.pos = int(offset)
	case io.SeekCurrent:
		a.pos += int(offset)
	case io.SeekEnd:
		a.pos = len(a.buffer) + int(offset)
	}
	if a.pos < 0 {
		a.pos = 0
	}
	if a.pos > len(a.buffer) {
		a.pos = len(a.buffer)
	}
	return int64(a.pos), nil
}

// AudioSystem - система воспроизведения звуков
type AudioSystem struct {
	enabled    bool
	audioCtx   *audio.Context
	soundQueue []SoundEffect
}

// NewAudioSystem создаёт аудиосистему
func NewAudioSystem() *AudioSystem {
	audioCtx := audio.NewContext(44100)
	return &AudioSystem{
		enabled:    true,
		audioCtx:   audioCtx,
		soundQueue: make([]SoundEffect, 0),
	}
}

// generateSamples генерирует сэмплы для звука
func (as *AudioSystem) generateSamples(effect SoundEffect) []float64 {
	samples := make([]float64, effect.duration)
	phase := 0.0
	freq := effect.frequency

	for i := range samples {
		if effect.sliding {
			// Скользящая частота (для прыжков)
			freq = effect.frequency * (1.0 - float64(i)/float64(effect.duration)*0.5)
		}

		var sample float64
		switch effect.waveform {
		case 0: // Sine
			sample = math.Sin(phase)
		case 1: // Square
			if math.Sin(phase) >= 0 {
				sample = 1.0
			} else {
				sample = -1.0
			}
		case 2: // Sawtooth
			sample = 2*(phase/(2*math.Pi)-math.Floor(phase/(2*math.Pi)+0.5))
		case 3: // Noise
			sample = (rand.Float64() - 0.5) * 2
		default:
			sample = math.Sin(phase)
		}

		// Envelope (ADSR-like)
		envelope := 1.0
		if i < effect.duration/10 {
			envelope = float64(i) / float64(effect.duration/10)
		} else if i > effect.duration*7/10 {
			envelope = float64(effect.duration-i) / float64(effect.duration*3/10)
		}

		samples[i] = sample * envelope * effect.volume
		phase += 2 * math.Pi * freq / 44100
	}

	return samples
}

// playSound воспроизводит звук
func (as *AudioSystem) playSound(effect SoundEffect) {
	if !as.enabled {
		return
	}

	// Generate samples
	samples := as.generateSamples(effect)

	// Create audio stream
	stream := &AudioStream{buffer: samples, pos: 0}

	// Create player and play
	player, err := as.audioCtx.NewPlayer(stream)
	if err != nil {
		log.Printf("Error creating player: %v", err)
		return
	}

	player.Play()
}

// PlayJump - звук прыжка
func (as *AudioSystem) PlayJump() {
	as.playSound(SoundEffect{
		frequency: 400,
		duration:  200,
		volume:    0.3,
		waveform:  0,
		sliding:   true,
	})
}

// PlayCollect - звук сбора предмета
func (as *AudioSystem) PlayCollect() {
	as.playSound(SoundEffect{
		frequency: 880,
		duration:  150,
		volume:    0.25,
		waveform:  0,
		sliding:   false,
	})
}

// PlayEnter - звук входа в дом
func (as *AudioSystem) PlayEnter() {
	as.playSound(SoundEffect{
		frequency: 330,
		duration:  300,
		volume:    0.2,
		waveform:  1,
		sliding:   false,
	})
}

// PlayThunder - звук грома
func (as *AudioSystem) PlayThunder() {
	as.playSound(SoundEffect{
		frequency: 80,
		duration:  500,
		volume:    0.4,
		waveform:  3,
		sliding:   true,
	})
}

// PlayHit - звук получения урона
func (as *AudioSystem) PlayHit() {
	as.playSound(SoundEffect{
		frequency: 200,
		duration:  250,
		volume:    0.35,
		waveform:  2,
		sliding:   true,
	})
}

// PlayCoin - звук сбора монеты (высокий звонкий)
func (as *AudioSystem) PlayCoin() {
	as.playSound(SoundEffect{
		frequency: 1200,
		duration:  180,
		volume:    0.3,
		waveform:  0,
		sliding:   false,
	})
	// Второй тон для аккорда
	as.playSound(SoundEffect{
		frequency: 1600,
		duration:  180,
		volume:    0.2,
		waveform:  0,
		sliding:   false,
	})
}

// PlayPowerup - звук бонуса
func (as *AudioSystem) PlayPowerup() {
	// Восходящая арпеджио
	as.playSound(SoundEffect{frequency: 523, duration: 100, volume: 0.3, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 659, duration: 100, volume: 0.3, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 784, duration: 100, volume: 0.3, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 1047, duration: 150, volume: 0.3, waveform: 0, sliding: false})
}

// PlayStar - звук звезды (неуязвимость)
func (as *AudioSystem) PlayStar() {
	for i := 0; i < 5; i++ {
		as.playSound(SoundEffect{
			frequency: 880 + float64(i)*100,
			duration:  80,
			volume:    0.25,
			waveform:  0,
			sliding:   false,
		})
	}
}

// PlayExtraLife - звук дополнительной жизни
func (as *AudioSystem) PlayExtraLife() {
	as.playSound(SoundEffect{frequency: 659, duration: 120, volume: 0.3, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 988, duration: 120, volume: 0.3, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 1319, duration: 200, volume: 0.3, waveform: 0, sliding: false})
}

// PlayAchievementBronze - звук для бронзовой ачивки
func (as *AudioSystem) PlayAchievementBronze() {
	// Short triumphant fanfare
	as.playSound(SoundEffect{frequency: 523, duration: 100, volume: 0.35, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 659, duration: 100, volume: 0.35, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 784, duration: 150, volume: 0.35, waveform: 0, sliding: false})
}

// PlayAchievementSilver - звук для серебряной ачивки
func (as *AudioSystem) PlayAchievementSilver() {
	// Medium fanfare with more notes
	as.playSound(SoundEffect{frequency: 523, duration: 80, volume: 0.35, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 659, duration: 80, volume: 0.35, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 784, duration: 80, volume: 0.35, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 1047, duration: 150, volume: 0.35, waveform: 0, sliding: false})
}

// PlayAchievementGold - звук для золотой ачивки
func (as *AudioSystem) PlayAchievementGold() {
	// Grand fanfare
	as.playSound(SoundEffect{frequency: 523, duration: 80, volume: 0.4, waveform: 1, sliding: false})
	as.playSound(SoundEffect{frequency: 659, duration: 80, volume: 0.4, waveform: 1, sliding: false})
	as.playSound(SoundEffect{frequency: 784, duration: 80, volume: 0.4, waveform: 1, sliding: false})
	as.playSound(SoundEffect{frequency: 1047, duration: 80, volume: 0.4, waveform: 1, sliding: false})
	as.playSound(SoundEffect{frequency: 1319, duration: 100, volume: 0.4, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 1568, duration: 200, volume: 0.4, waveform: 0, sliding: false})
}

// PlayAchievementPlatinum - звук для платиновой ачивки
func (as *AudioSystem) PlayAchievementPlatinum() {
	// Epic ascending arpeggio
	for i := 0; i < 8; i++ {
		as.playSound(SoundEffect{
			frequency: 523 * math.Pow(1.05946, float64(i)),
			duration:  60,
			volume:    0.4,
			waveform:  0,
			sliding:   false,
		})
	}
	as.playSound(SoundEffect{frequency: 1047, duration: 300, volume: 0.4, waveform: 0, sliding: false})
}

// PlayAchievementDiamond - звук для алмазной ачивки (самый эпичный)
func (as *AudioSystem) PlayAchievementDiamond() {
	// Ultra epic multi-layered fanfare
	// First arpeggio up
	for i := 0; i < 12; i++ {
		as.playSound(SoundEffect{
			frequency: 523 * math.Pow(1.05946, float64(i)),
			duration:  50,
			volume:    0.4,
			waveform:  0,
			sliding:   false,
		})
	}
	// Second arpeggio even higher
	for i := 0; i < 8; i++ {
		as.playSound(SoundEffect{
			frequency: 1047 * math.Pow(1.05946, float64(i)),
			duration:  40,
			volume:    0.4,
			waveform:  0,
			sliding:   false,
		})
	}
	// Final triumphant chord
	as.playSound(SoundEffect{frequency: 1047, duration: 400, volume: 0.45, waveform: 1, sliding: false})
	as.playSound(SoundEffect{frequency: 1319, duration: 400, volume: 0.45, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 1568, duration: 400, volume: 0.45, waveform: 0, sliding: false})
	as.playSound(SoundEffect{frequency: 2093, duration: 500, volume: 0.5, waveform: 0, sliding: true})
}

// PlayJumpBump - звук при приземлении
func (as *AudioSystem) PlayJumpBump() {
	as.playSound(SoundEffect{
		frequency: 150,
		duration:  80,
		volume:    0.2,
		waveform:  3,
		sliding:   false,
	})
}

// PlayEnemyDefeat - звук победы над врагом
func (as *AudioSystem) PlayEnemyDefeat() {
	as.playSound(SoundEffect{
		frequency: 440,
		duration:  100,
		volume:    0.3,
		waveform:  1,
		sliding:   true,
	})
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
		chimneyX: 600,
		chimneyY: screenHeight - groundHeight - 100 - 40, // На крыше дома
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
		x:          50,
		y:          float64(screenHeight - groundHeight - 40),
		vy:         0,
		width:      30,
		height:     40,
		onGround:   true,
		score:      0,
		coins:      0,
		lives:      3,
		facing:     1,
		animFrame:  0,
		invincible: 0,
	}

	// Initialize coins (scattered around the level)
	coins := []Coin{
		{x: 200, y: screenHeight - groundHeight - 80, collected: false},
		{x: 350, y: screenHeight - groundHeight - 120, collected: false},
		{x: 500, y: screenHeight - groundHeight - 60, collected: false},
		{x: 100, y: screenHeight - groundHeight - 150, collected: false},
		{x: 700, y: screenHeight - groundHeight - 100, collected: false},
		{x: 250, y: screenHeight - groundHeight - 200, collected: false},
		{x: 450, y: screenHeight - groundHeight - 180, collected: false},
		{x: 600, y: screenHeight - groundHeight - 140, collected: false},
	}

	// Initialize enemies (patrolling Goomba-style)
	enemies := []Enemy{
		{x: 300, y: screenHeight - groundHeight - 30, width: 30, height: 30, vx: 1, alive: true},
		{x: 550, y: screenHeight - groundHeight - 30, width: 30, height: 30, vx: -1, alive: true},
		{x: 150, y: screenHeight - groundHeight - 30, width: 30, height: 30, vx: 1.5, alive: true},
	}

	// Initialize platforms (floating platforms for jumping)
	platforms := []Platform{
		{x: 180, y: screenHeight - groundHeight - 120, width: 80, height: 15},
		{x: 350, y: screenHeight - groundHeight - 180, width: 100, height: 15},
		{x: 520, y: screenHeight - groundHeight - 240, width: 80, height: 15},
		{x: 100, y: screenHeight - groundHeight - 280, width: 120, height: 15},
		{x: 650, y: screenHeight - groundHeight - 200, width: 90, height: 15},
	}

	// Initialize jump particles
	jumpParticles := make([]JumpParticle, 20)
	for i := range jumpParticles {
		jumpParticles[i] = JumpParticle{
			x:       0,
			y:       0,
			vx:      float32(i%5-2) * 0.5,
			vy:      float32(i%3) * 0.5,
			size:    float32(i%4+2),
			life:    0,
			maxLife: 20,
		}
	}

	// Initialize powerups (bonus items)
	powerups := []Powerup{
		{x: 280, y: screenHeight - groundHeight - 150, powerType: MushroomPower, collected: false},
		{x: 620, y: screenHeight - groundHeight - 200, powerType: StarPower, collected: false},
		{x: 150, y: screenHeight - groundHeight - 250, powerType: ExtraLifePower, collected: false},
	}

	// Initialize spark particles
	sparkParticles := make([]SparkParticle, 50)
	for i := range sparkParticles {
		sparkParticles[i] = SparkParticle{
			x:       0,
			y:       0,
			vx:      float32(rand.Intn(10)-5) * 0.8,
			vy:      float32(rand.Intn(10)-5) * 0.8,
			size:    float32(rand.Intn(4)+2),
			life:    0,
			maxLife: 30,
			color:   color.RGBA{255, 215, 0, 255},
		}
	}

	// Initialize floating texts
	floatingTexts := make([]FloatingText, 20)
	for i := range floatingTexts {
		floatingTexts[i] = FloatingText{
			x:       0,
			y:       0,
			text:    "",
			life:    0,
			maxLife: 60,
			vy:      -1,
			color:   color.RGBA{255, 255, 255, 255},
			scale:   1.0,
		}
	}

	return &Game{
		playerX:       100,
		playerY:       screenHeight - groundHeight - 50,
		frameCount:    0,
		clouds:        clouds,
		stormClouds:   stormClouds,
		trees:         trees,
		player:        player,
		state:         Menu,
		timeOfDay:     Day,
		weather:       Clear,
		stars:         stars,
		moonX:         100,
		moonY:         80,
		raindrops:     raindrops,
		lightning:     Lightning{active: false, timer: 0, branches: []LightningBranch{}},
		house:         house,
		audio:         NewAudioSystem(),
		coins:         coins,
		enemies:       enemies,
		platforms:     platforms,
		jumpParticles: jumpParticles,
		powerups:      powerups,
		sparkParticles: sparkParticles,
		floatingTexts: floatingTexts,
		screenShake:   ScreenShake{active: false, intensity: 0, timer: 0},
		
		// Initialize world and crafting
		world:     NewWorld(rand.Int63()),
		camera:    NewCamera(),
		inventory: NewInventory(),
		recipes:   NewRecipes(),
		
		// Tutorial and quests
		tutorial:    NewTutorial(),
		quests:      NewQuests(),
		checkpoints: NewCheckpoints(),
		healthPacks: NewHealthPacks(),
		
		// Achievement album
		album:       NewAchievementAlbum(),
		activeNotifications: make([]AchievementNotification, 0),
		blocksMined: 0,
		enemiesDefeated: 0,
		
		// Tutorial hints
		showControls:    false,
		controlsTimer:   0,
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

	// Handle game won state
	if g.state == GameWon {
		// Play again with Enter
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.resetGame()
			g.state = Playing
		}
		// Return to menu with ESC
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = Menu
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
		g.spawnJumpParticles(float32(g.player.x)+g.player.width/2, float32(g.player.y)+g.player.height)
		g.tutorial.CompleteStep(1) // Complete jump step
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

	// Screen boundaries (world bounds)
	if g.player.x < 0 {
		g.player.x = 0
	}
	if g.player.x > float64(worldWidth)-float64(g.player.width) {
		g.player.x = float64(worldWidth) - float64(g.player.width)
	}

	// House entry detection
	if g.state == Playing {
		g.checkHouseEntry()
	}

	// Apple collection
	g.checkAppleCollection()

	// Update platforms collision
	g.updatePlatforms()

	// Update enemies
	g.updateEnemies()

	// Update coins
	g.updateCoins()

	// Update powerups
	g.updatePowerups()

	// Update jump particles
	g.updateJumpParticles()

	// Update spark particles
	g.updateSparkParticles()

	// Update floating texts
	g.updateFloatingTexts()

	// Update screen shake
	g.updateScreenShake()

	// Check win condition
	g.checkWinCondition()

	// Update camera
	if g.camera != nil && g.state == Playing {
		g.camera.Update(g.player.x, g.player.y)
	}

	// Handle mining and placing blocks
	g.handleBlockInteraction()

	// Handle crafting mode
	g.handleCrafting()

	// Update tutorial
	if g.tutorial != nil {
		g.tutorial.Update(g)
	}
	
	// Update achievement notifications
	g.updateAchievementNotifications()
	
	// Update checkpoints
	g.updateCheckpoints()
	
	// Update health packs
	g.updateHealthPacks()

	// Update invincibility
	if g.player.invincible > 0 {
		g.player.invincible--
	}

	return nil
}

// handleBlockInteraction - обработка добычи и размещения блоков
func (g *Game) handleBlockInteraction() {
	if g.state != Playing {
		return
	}

	// Get mouse position
	mx, my := ebiten.CursorPosition()

	// Left click - mine block
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.mineBlock(mx, my)
		g.tutorial.CompleteStep(2) // Complete mining step
		g.blocksMined++
		g.album.UpdateAchievement(BlockMiner, g.blocksMined)
	}

	// Right click - place block
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.placeBlock(mx, my)
		g.tutorial.CompleteStep(3) // Complete placing step
		g.album.UpdateAchievement(Builder, g.blocksMined) // Reuse blocksMined for builder
	}

	// Select slot with number keys
	for i := 0; i < inventorySize; i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(i + 49)) { // Key1=49, Key2=50, etc.
			g.inventory.selected = i
			g.tutorial.CompleteStep(4) // Complete inventory step
		}
	}
	
	// Select slot with mouse wheel
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		g.inventory.selected--
		if g.inventory.selected < 0 {
			g.inventory.selected = inventorySize - 1
		}
		g.tutorial.CompleteStep(4)
	} else if wheelY < 0 {
		g.inventory.selected++
		if g.inventory.selected >= inventorySize {
			g.inventory.selected = 0
		}
		g.tutorial.CompleteStep(4)
	}
	
	// Toggle controls hint with H key
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.showControls = !g.showControls
	}
	
	// Toggle achievement album with B key
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		if g.album != nil {
			g.album.showAlbum = !g.album.showAlbum
		}
	}
}

// handleCrafting - обработка режима крафта
func (g *Game) handleCrafting() {
	if g.state != Crafting {
		return
	}

	// Exit crafting mode with ESC
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = Playing
		return
	}

	// Craft with number keys (1-5)
	recipes := []BlockType{Plank, Bricks, Crafting_Table}
	for i, recipe := range recipes {
		if inpututil.IsKeyJustPressed(ebiten.Key(i + 49)) {
			// Find and craft recipe
			for _, r := range g.recipes {
				if r.result == recipe && CanCraft(r, g.inventory) {
					Craft(r, g.inventory)
					g.audio.PlayPowerup()
					break
				}
			}
		}
	}
}

// updateAchievementNotifications - обновляет активные уведомления ачивок
func (g *Game) updateAchievementNotifications() {
	// Check for new notifications
	if g.album != nil {
		ach := g.album.GetNextNotification()
		if ach != nil {
			// Play sound based on tier
			switch ach.medalTier {
			case Bronze:
				g.audio.PlayAchievementBronze()
			case Silver:
				g.audio.PlayAchievementSilver()
			case Gold:
				g.audio.PlayAchievementGold()
			case Platinum:
				g.audio.PlayAchievementPlatinum()
			case Diamond:
				g.audio.PlayAchievementDiamond()
			}
			
			// Create notification
			notification := AchievementNotification{
				achievement: *ach,
				life:        0,
				maxLife:     300, // 5 seconds at 60 FPS
				scale:       0,
				y:           screenHeight + 100,
				targetY:     screenHeight - 120,
			}
			
			// Add to active notifications (max 3 at a time)
			if len(g.activeNotifications) < 3 {
				g.activeNotifications = append(g.activeNotifications, notification)
			} else {
				// Queue for later
				g.album.pendingNotifications = append(g.album.pendingNotifications, *ach)
			}
			
			// Trigger screen shake and particles
			g.triggerScreenShake(8, 20)
			g.spawnSparkParticles(screenWidth/2, screenHeight/2, 50, GetMedalColor(ach.medalTier))
		}
	}
	
	// Update active notifications
	for i := range g.activeNotifications {
		notif := &g.activeNotifications[i]
		if notif.life <= 0 {
			continue
		}
		
		notif.life++
		
		// Animate in
		if notif.life < 30 {
			notif.scale += 0.03
			notif.y += (notif.targetY - notif.y) * 0.1
		}
		
		// Animate out
		if notif.life > notif.maxLife-30 {
			notif.scale -= 0.03
			notif.y += 3
		}
		
		if notif.life >= notif.maxLife {
			notif.life = 0
		}
	}
	
	// Remove finished notifications
	cleaned := make([]AchievementNotification, 0)
	for _, notif := range g.activeNotifications {
		if notif.life > 0 && notif.life < notif.maxLife {
			cleaned = append(cleaned, notif)
		}
	}
	g.activeNotifications = cleaned
}

// drawAchievementNotifications - отрисовка активных уведомлений ачивок
func (g *Game) drawAchievementNotifications(screen *ebiten.Image) {
	for _, notif := range g.activeNotifications {
		if notif.life <= 0 {
			continue
		}
		
		ach := notif.achievement
		tierColor := GetMedalColor(ach.medalTier)
		
		// Background panel with gradient
		panelW := 500
		panelH := 100
		panelX := screenWidth/2 - panelW/2
		panelY := int(notif.y)
		
		// Outer glow
		vector.DrawFilledRect(screen, float32(panelX-5), float32(panelY-5), float32(panelW+10), float32(panelH+10), color.RGBA{tierColor.R, tierColor.G, tierColor.B, 100}, false)
		
		// Main background
		vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{0, 0, 0, 200}, false)
		
		// Border
		vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 4, tierColor, false)
		
		// Icon/Medal (large, animated)
		iconScale := notif.scale
		if iconScale > 1 {
			iconScale = 1
		}
		iconSize := int(60 * iconScale)
		iconX := panelX + 35
		iconY := panelY + 50 - iconSize/2

		// Medal glow
		vector.DrawFilledCircle(screen, float32(iconX), float32(iconY), float32(iconSize)/2+10, color.RGBA{tierColor.R, tierColor.G, tierColor.B, 100}, false)

		// Medal background
		vector.DrawFilledCircle(screen, float32(iconX), float32(iconY), float32(iconSize)/2, tierColor, false)

		// Icon emoji
		ebitenutil.DebugPrintAt(screen, ach.icon, iconX-15, iconY-15)
		
		// Text
		textX := panelX + 90
		
		// "ACHIEVEMENT UNLOCKED!" title
		titleText := "🏆 ACHIEVEMENT UNLOCKED!"
		titleY := panelY + 20
		ebitenutil.DebugPrintAt(screen, titleText, textX, titleY)
		
		// Achievement name
		nameY := panelY + 45
		ebitenutil.DebugPrintAt(screen, ach.title, textX, nameY)
		
		// Tier name
		tierY := panelY + 70
		tierText := GetTierName(ach.medalTier)
		ebitenutil.DebugPrintAt(screen, tierText, textX, tierY)
		
		// Animated sparkles around notification
		if notif.life < 60 {
			for i := 0; i < 5; i++ {
				sparkleX := float32(panelX + (i % 3) * 170 + (notif.life % 20))
				sparkleY := float32(panelY + (i / 3) * 50 + (notif.life % 15))
				vector.DrawFilledCircle(screen, sparkleX, sparkleY, 3, color.RGBA{255, 255, 255, 200}, false)
			}
		}
	}
}

// updateCheckpoints - обновляет чекпоинты
func (g *Game) updateCheckpoints() {
	for i := range g.checkpoints {
		cp := &g.checkpoints[i]
		
		// Check collision with player
		if checkRectCollision(
			float32(g.player.x), float32(g.player.y), g.player.width, g.player.height,
			cp.x, cp.y-40, 40, 40,
		) {
			if !cp.activated {
				cp.activated = true
				g.audio.PlayCollect()
				g.spawnFloatingText(cp.x, cp.y-60, "CHECKPOINT!", color.RGBA{0, 255, 255, 255})
				g.spawnSparkParticles(cp.x+20, cp.y-20, 20, color.RGBA{0, 255, 255, 255})
			}
		}
	}
}

// updateHealthPacks - обновляет аптечки
func (g *Game) updateHealthPacks() {
	for i := range g.healthPacks {
		pack := &g.healthPacks[i]
		if pack.collected {
			continue
		}
		
		// Apply gravity
		pack.vy += gravity
		pack.y += pack.vy
		
		// Ground collision
		groundY := float32(screenHeight - groundHeight - 20)
		if pack.y >= groundY {
			pack.y = groundY
			pack.vy = 0
		}
		
		// Check collision with player
		if checkRectCollision(
			float32(g.player.x), float32(g.player.y), g.player.width, g.player.height,
			pack.x, pack.y-15, 20, 20,
		) {
			pack.collected = true
			g.player.lives++
			g.audio.PlayExtraLife()
			g.spawnFloatingText(pack.x, pack.y, "+1 LIFE", color.RGBA{255, 100, 100, 255})
			g.spawnSparkParticles(pack.x+10, pack.y-10, 15, color.RGBA{255, 100, 100, 255})
		}
	}
}

// updateQuests - обновляет квесты
func (g *Game) updateQuests(blocksMined int, enemiesDefeated int) {
	for i := range g.quests {
		quest := &g.quests[i]
		if quest.completed {
			continue
		}
		
		switch quest.id {
		case 0: // First steps - mine 5 blocks
			// This would need a counter, simplified for now
		case 1: // Collector - 10 coins
			if g.player.coins >= 10 {
				quest.completed = true
				quest.objective = "✓ Выполнено"
				g.player.score += quest.reward
				g.audio.PlayPowerup()
				g.spawnFloatingText(float32(g.player.x), float32(g.player.y), "QUEST COMPLETE!", color.RGBA{255, 215, 0, 255})
			} else {
				quest.objective = fmt.Sprintf("%d/10 монет", g.player.coins)
			}
		case 2: // Enemy hunter - 3 enemies
			// This would need a counter
		case 3: // Miner - diamond ore
			// This would need tracking
		}
	}
}

// checkWinCondition - проверяет условие победы
func (g *Game) checkWinCondition() {
	// Check if all coins collected
	allCoinsCollected := true
	for _, coin := range g.coins {
		if !coin.collected {
			allCoinsCollected = false
			break
		}
	}

	allEnemiesDefeated := true
	for _, enemy := range g.enemies {
		if enemy.alive {
			allEnemiesDefeated = false
			break
		}
	}

	// Win if all coins collected and all enemies defeated
	// (powerups and apples are optional bonuses)
	if allCoinsCollected && allEnemiesDefeated {
		g.state = GameWon
		g.audio.PlayExtraLife() // Play victory sound
		g.triggerScreenShake(10, 30)
		g.spawnSparkParticles(screenWidth/2, screenHeight/2, 100, color.RGBA{255, 215, 0, 255})
	}
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
				// Small sparkle effect
				g.spawnSparkParticles(appleCX, appleCY, 5, color.RGBA{220, 20, 60, 255})
				g.spawnFloatingText(appleCX, appleCY, "+1", color.RGBA{220, 20, 60, 255})
			}
		}
	}
}

// updatePlatforms - проверяет коллизию игрока с платформами
func (g *Game) updatePlatforms() {
	playerRect := struct {
		x, y, w, h float32
	}{
		x: float32(g.player.x),
		y: float32(g.player.y),
		w: g.player.width,
		h: g.player.height,
	}

	for _, platform := range g.platforms {
		// Check if player is falling onto platform
		if g.player.vy >= 0 { // Only when falling
			// Check horizontal overlap
			if playerRect.x < platform.x+platform.width &&
				playerRect.x+playerRect.w > platform.x {
				// Check if player feet are near platform top
				platformTop := platform.y
				feetY := playerRect.y + playerRect.h

				// Allow some tolerance for collision
				if feetY >= platformTop-5 && feetY <= platformTop+10 {
					g.player.y = float64(platformTop - playerRect.h)
					g.player.vy = 0
					g.player.onGround = true
				}
			}
		}
	}
}

// updateEnemies - обновляет врагов (патрулирование, урон игроку)
func (g *Game) updateEnemies() {
	for i := range g.enemies {
		enemy := &g.enemies[i]
		if !enemy.alive {
			continue
		}

		// Move enemy
		enemy.x += enemy.vx
		enemy.animFrame++

		// Patrol boundaries (reverse direction at edges)
		if enemy.x < 100 || enemy.x > 700 {
			enemy.vx = -enemy.vx
		}

		// Check collision with player
		if g.player.invincible <= 0 {
			if checkRectCollision(
				float32(g.player.x), float32(g.player.y), g.player.width, g.player.height,
				enemy.x, enemy.y, enemy.width, enemy.height,
			) {
				// Player takes damage
				g.player.lives--
				g.player.invincible = 120 // 2 seconds at 60 FPS
				g.audio.PlayHit()
				g.triggerScreenShake(5, 15)
				// Red flash effect
				g.spawnSparkParticles(float32(g.player.x)+g.player.width/2, float32(g.player.y)+g.player.height/2, 15, color.RGBA{255, 0, 0, 255})
				g.spawnFloatingText(float32(g.player.x), float32(g.player.y), "OUCH!", color.RGBA{255, 0, 0, 255})

				// Knockback player
				if enemy.x < float32(g.player.x) {
					g.player.x += 50
				} else {
					g.player.x -= 50
				}
				g.player.vy = -5

				// Check if player died
				if g.player.lives <= 0 {
					g.resetGame()
				}
			}
		}

		// Check if player jumped on enemy (Mario-style)
		if g.player.vy > 0 {
			playerBottom := float32(g.player.y) + g.player.height
			if playerBottom > enemy.y && playerBottom < enemy.y+20 &&
				float32(g.player.x)+g.player.width/2 > enemy.x &&
				float32(g.player.x)+g.player.width/2 < enemy.x+enemy.width {
				// Enemy defeated
				enemy.alive = false
				g.player.vy = -8 // Bounce
				g.player.score += 100
				g.audio.PlayEnemyDefeat()
				g.tutorial.CompleteStep(6) // Complete enemy step
				g.enemiesDefeated++
				g.album.UpdateAchievement(EnemySlayer, g.enemiesDefeated)
				// Spark effect
				g.spawnSparkParticles(enemy.x+enemy.width/2, enemy.y+enemy.height/2, 20, color.RGBA{139, 69, 19, 255})
				g.spawnFloatingText(enemy.x, enemy.y, "+100", color.RGBA{255, 255, 0, 255})
			}
		}
	}
}

// updateCoins - обновляет монеты (сбор)
func (g *Game) updateCoins() {
	playerRect := struct {
		x, y, w, h float32
	}{
		x: float32(g.player.x),
		y: float32(g.player.y),
		w: g.player.width,
		h: g.player.height,
	}

	for i := range g.coins {
		coin := &g.coins[i]
		if coin.collected {
			continue
		}

		coin.animFrame++

		// Check collision with player
		if checkRectCollision(
			playerRect.x, playerRect.y, playerRect.w, playerRect.h,
			coin.x, coin.y, 20, 20,
		) {
			coin.collected = true
			g.player.coins++
			g.player.score += 50
			g.audio.PlayCoin()
			g.album.UpdateAchievement(CoinCollector, g.player.coins)
			// Spark effect
			g.spawnSparkParticles(coin.x+10, coin.y+10, 10, color.RGBA{255, 215, 0, 255})
			// Floating text
			g.spawnFloatingText(coin.x, coin.y, "+50", color.RGBA{255, 215, 0, 255})
		}
	}
}

// updateJumpParticles - обновляет частицы прыжка
func (g *Game) updateJumpParticles() {
	for i := range g.jumpParticles {
		particle := &g.jumpParticles[i]
		if particle.life <= 0 {
			continue
		}

		particle.x += particle.vx
		particle.y += particle.vy
		particle.vy += 0.1 // gravity
		particle.life--

		if particle.life <= 0 {
			// Reset particle for reuse
			particle.size = 0
		}
	}
}

// updateSparkParticles - обновляет частицы искр
func (g *Game) updateSparkParticles() {
	for i := range g.sparkParticles {
		particle := &g.sparkParticles[i]
		if particle.life <= 0 {
			continue
		}

		particle.x += particle.vx
		particle.y += particle.vy
		particle.vy += 0.05 // gravity
		particle.life--
		particle.size *= 0.95 // shrink

		if particle.life <= 0 {
			particle.size = 0
		}
	}
}

// updateFloatingTexts - обновляет всплывающий текст
func (g *Game) updateFloatingTexts() {
	for i := range g.floatingTexts {
		text := &g.floatingTexts[i]
		if text.life <= 0 {
			continue
		}

		text.y += text.vy
		text.life--
		text.scale *= 0.98

		if text.life <= 0 {
			text.text = ""
		}
	}
}

// updateScreenShake - обновляет тряску экрана
func (g *Game) updateScreenShake() {
	if g.screenShake.active {
		g.screenShake.timer--
		if g.screenShake.timer <= 0 {
			g.screenShake.active = false
			g.screenShake.intensity = 0
		}
	}
}

// updatePowerups - обновляет бонусы (сбор, движение грибов)
func (g *Game) updatePowerups() {
	playerRect := struct {
		x, y, w, h float32
	}{
		x: float32(g.player.x),
		y: float32(g.player.y),
		w: g.player.width,
		h: g.player.height,
	}

	for i := range g.powerups {
		powerup := &g.powerups[i]
		if powerup.collected {
			continue
		}

		powerup.animFrame++

		// Mushroom moves left and right
		if powerup.powerType == MushroomPower && powerup.onGround {
			powerup.x += powerup.px
			// Change direction at boundaries
			if powerup.x < 50 || powerup.x > 750 {
				powerup.px = -powerup.px
			}
		}

		// Apply gravity to powerups
		if !powerup.onGround {
			powerup.vy += gravity
			powerup.y += powerup.vy
		}

		// Ground collision for powerups
		groundY := float32(screenHeight - groundHeight - 20)
		if powerup.y >= groundY {
			powerup.y = groundY
			powerup.vy = 0
			powerup.onGround = true
			if powerup.powerType == MushroomPower {
				powerup.px = 2 // Start moving when hits ground
			}
		}

		// Check collision with platforms
		for _, platform := range g.platforms {
			if powerup.vy >= 0 {
				if powerup.x+15 > platform.x && powerup.x-15 < platform.x+platform.width {
					platformTop := platform.y
					if powerup.y+20 >= platformTop-5 && powerup.y+20 <= platformTop+10 {
						powerup.y = platformTop - 20
						powerup.vy = 0
						powerup.onGround = true
						if powerup.powerType == MushroomPower {
							powerup.px = 2
						}
					}
				}
			}
		}

		// Check collision with player
		if checkRectCollision(
			playerRect.x, playerRect.y, playerRect.w, playerRect.h,
			powerup.x-15, powerup.y-15, 30, 30,
		) {
			powerup.collected = true
			g.applyPowerup(powerup)
		}
	}
}

// applyPowerup - применяет эффект бонуса
func (g *Game) applyPowerup(powerup *Powerup) {
	switch powerup.powerType {
	case MushroomPower:
		// Increase player size and score
		g.player.width = 40
		g.player.height = 50
		g.player.score += 200
		g.audio.PlayPowerup()
		g.spawnSparkParticles(powerup.x, powerup.y, 25, color.RGBA{220, 20, 60, 255})
		g.spawnFloatingText(powerup.x, powerup.y, "MUSHROOM!", color.RGBA{220, 20, 60, 255})
		g.triggerScreenShake(3, 10)
	case StarPower:
		// Make player invincible for 10 seconds (600 frames)
		g.player.invincible = 600
		g.player.score += 500
		g.audio.PlayStar()
		g.spawnSparkParticles(powerup.x, powerup.y, 30, color.RGBA{255, 215, 0, 255})
		g.spawnFloatingText(powerup.x, powerup.y, "STAR POWER!", color.RGBA{255, 215, 0, 255})
		g.triggerScreenShake(4, 15)
	case ExtraLifePower:
		// Add extra life
		g.player.lives++
		g.player.score += 1000
		g.audio.PlayExtraLife()
		g.spawnSparkParticles(powerup.x, powerup.y, 25, color.RGBA{255, 20, 60, 255})
		g.spawnFloatingText(powerup.x, powerup.y, "1UP!", color.RGBA{255, 100, 100, 255})
	}
}

// spawnJumpParticles - создаёт частицы при прыжке
func (g *Game) spawnJumpParticles(x, y float32) {
	for i := 0; i < 5; i++ {
		// Find inactive particle
		for j := range g.jumpParticles {
			if g.jumpParticles[j].life <= 0 {
				g.jumpParticles[j] = JumpParticle{
					x:       x + float32(i)*3 - 7,
					y:       y,
					vx:      float32(i-2) * 0.8,
					vy:      -float32(i%3+1) * 0.5,
					size:    float32(i%4+2),
					life:    20,
					maxLife: 20,
				}
				break
			}
		}
	}
}

// spawnSparkParticles - создаёт искры в точке
func (g *Game) spawnSparkParticles(x, y float32, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		for j := range g.sparkParticles {
			if g.sparkParticles[j].life <= 0 {
				g.sparkParticles[j] = SparkParticle{
					x:       x,
					y:       y,
					vx:      float32(rand.Intn(10)-5) * 1.5,
					vy:      float32(rand.Intn(10)-5) * 1.5,
					size:    float32(rand.Intn(4)+3),
					life:    30,
					maxLife: 30,
					color:   c,
				}
				break
			}
		}
	}
}

// spawnFloatingText - создаёт всплывающий текст
func (g *Game) spawnFloatingText(x, y float32, text string, c color.RGBA) {
	for i := range g.floatingTexts {
		if g.floatingTexts[i].life <= 0 {
			g.floatingTexts[i] = FloatingText{
				x:       x,
				y:       y,
				text:    text,
				life:    60,
				maxLife: 60,
				vy:      -1.5,
				color:   c,
				scale:   1.2,
			}
			break
		}
	}
}

// triggerScreenShake - запускает тряску экрана
func (g *Game) triggerScreenShake(intensity float32, duration int) {
	g.screenShake = ScreenShake{
		active:    true,
		intensity: intensity,
		timer:     duration,
	}
}

// checkRectCollision - проверяет столкновение двух прямоугольников
func checkRectCollision(x1, y1, w1, h1, x2, y2, w2, h2 float32) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// resetGame - сбрасывает игру при смерти
func (g *Game) resetGame() {
	g.player.x = 50
	g.player.y = float64(screenHeight - groundHeight - 40)
	g.player.vy = 0
	g.player.lives = 3
	g.player.score = 0
	g.player.coins = 0
	g.player.invincible = 60
	g.player.width = 30
	g.player.height = 40

	// Reset coins
	for i := range g.coins {
		g.coins[i].collected = false
	}

	// Reset enemies
	for i := range g.enemies {
		g.enemies[i].alive = true
	}

	// Reset apples
	for i := range g.trees {
		for j := range g.trees[i].apples {
			g.trees[i].apples[j].collected = false
		}
	}

	// Reset powerups
	for i := range g.powerups {
		g.powerups[i].collected = false
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

	// Draw chimney (dark gray) - on the roof
	chimneyColor := color.RGBA{80, 80, 80, 255}
	chimneyBaseY := h.y - h.height - 20 // На крыше, чуть ниже пика
	vector.DrawFilledRect(screen, h.chimneyX, chimneyBaseY-40, 20, 40, chimneyColor, false)
	// Chimney top cap (wider)
	vector.DrawFilledRect(screen, h.chimneyX-5, chimneyBaseY-40, 30, 10, chimneyColor, false)
	// Chimney base (where it meets roof)
	vector.DrawFilledRect(screen, h.chimneyX-8, chimneyBaseY-5, 36, 10, chimneyColor, false)

	// Update chimneyY for smoke particles
	h.chimneyY = chimneyBaseY - 40

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
	chimneyTopX := h.chimneyX + 10
	chimneyTopY := h.chimneyY // Верх трубы
	
	for i := range g.house.smoke {
		particle := &g.house.smoke[i]

		// Update particle position - smoke rises from chimney top
		timeOffset := float32(g.frameCount+i*5) * 0.5
		particle.x = chimneyTopX + float32(math.Sin(float64(timeOffset)))*5 + particle.vx*2
		particle.y = chimneyTopY - float32(particle.life)*0.8 - particle.size/2
		particle.life++

		if particle.life >= particle.maxLife {
			particle.life = 0
		}

		// Draw smoke (gray circles with decreasing alpha and increasing size)
		alpha := uint8(150 - particle.life*150/particle.maxLife)
		size := particle.size + float32(particle.life)/5
		smokeColor := color.RGBA{180, 180, 180, alpha}
		vector.DrawFilledCircle(screen, particle.x, particle.y, size, smokeColor, false)
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

// drawGameWon - отрисовка экрана победы
func (g *Game) drawGameWon(screen *ebiten.Image) {
	// Animated background (gradient cycling through colors)
	r := uint8(128 + 127*math.Sin(float64(g.frameCount)*0.02))
	green := uint8(128 + 127*math.Sin(float64(g.frameCount)*0.02+2))
	blue := uint8(128 + 127*math.Sin(float64(g.frameCount)*0.02+4))
	screen.Fill(color.RGBA{r, green, blue, 255})

	// Draw victory stars
	for i := 0; i < 50; i++ {
		x := float32((i*73 + g.frameCount) % screenWidth)
		y := float32((i*47) % (screenHeight - 200))
		size := float32((i % 3) + 1)
		alpha := uint8(150 + 105*math.Sin(float64(g.frameCount)*0.1+float64(i)))
		vector.DrawFilledCircle(screen, x, y, size, color.RGBA{255, 255, 200, alpha}, false)
	}

	// Victory message with shadow
	victoryTitle := "🎉 VICTORY! 🎉"
	titleX := screenWidth/2 - len(victoryTitle)*12
	titleY := 100
	
	// Shadow
	ebitenutil.DebugPrintAt(screen, victoryTitle, titleX+3, titleY+3)
	// Main text (yellow)
	ebitenutil.DebugPrintAt(screen, victoryTitle, titleX, titleY)

	// Congratulations message
	congrats := "Congratulations, Hero!"
	congratsX := screenWidth/2 - len(congrats)*8
	ebitenutil.DebugPrintAt(screen, congrats, congratsX, 160)

	// Final stats box
	statsBoxX := screenWidth/2 - 150
	statsBoxY := 220
	statsBoxW := 300
	statsBoxH := 200
	
	// Box background (semi-transparent dark)
	vector.DrawFilledRect(screen, float32(statsBoxX), float32(statsBoxY), float32(statsBoxW), float32(statsBoxH), color.RGBA{0, 0, 0, 200}, false)
	// Box border
	vector.StrokeRect(screen, float32(statsBoxX), float32(statsBoxY), float32(statsBoxW), float32(statsBoxH), 3, color.RGBA{255, 215, 0, 255}, false)

	// Stats
	stats := []string{
		fmt.Sprintf("Final Score: %d", g.player.score),
		fmt.Sprintf("Coins: %d/%d", g.player.coins, len(g.coins)),
		fmt.Sprintf("Lives Remaining: %d", g.player.lives),
		"",
		"🌟 All objectives completed! 🌟",
	}
	
	for i, line := range stats {
		ebitenutil.DebugPrintAt(screen, line, statsBoxX+20, statsBoxY+30+i*30)
	}

	// Play again prompt (blinking)
	if g.frameCount%60 < 30 {
		playAgain := "Press ENTER to Play Again"
		playAgainX := screenWidth/2 - len(playAgain)*8
		ebitenutil.DebugPrintAt(screen, playAgain, playAgainX, 500)
	}

	// Or return to menu
	menuHint := "Press ESC for Main Menu"
	menuHintX := screenWidth/2 - len(menuHint)*6
	ebitenutil.DebugPrintAt(screen, menuHint, menuHintX, 540)
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

	if g.state == GameWon {
		g.drawGameWon(screen)
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

	// Draw world (terrain blocks)
	g.drawWorld(screen)

	// Draw house
	g.drawHouse(screen)

	// Draw trees
	g.drawTrees(screen)

	// Draw platforms
	g.drawPlatforms(screen)

	// Draw coins
	g.drawCoins(screen)

	// Draw powerups
	g.drawPowerups(screen)

	// Draw enemies
	g.drawEnemies(screen)

	// Draw player (bunny)
	g.drawPlayer(screen)

	// Draw jump particles
	g.drawJumpParticles(screen)

	// Draw spark particles
	g.drawSparkParticles(screen)

	// Draw floating texts
	g.drawFloatingTexts(screen)

	// Draw rain (stormy weather)
	if g.weather == Stormy {
		g.drawRain(screen)
		g.drawLightning(screen)
	}

	// Draw UI (score, lives, coins, inventory)
	g.drawUI(screen)

	// Draw tutorial
	g.drawTutorial(screen)
	
	// Draw quests
	g.drawQuests(screen)
	
	// Draw checkpoints
	g.drawCheckpoints(screen)
	
	// Draw health packs
	g.drawHealthPacks(screen)
	
	// Draw current hint
	g.drawCurrentHint(screen)
	
	// Draw controls hint
	g.drawControlsHint(screen)
	
	// Draw achievement album (overlay)
	g.drawAchievementAlbum(screen)
	
	// Draw achievement notifications (on top)
	g.drawAchievementNotifications(screen)

	// Draw screen shake flash effect
	if g.screenShake.active {
		// Draw red flash overlay
		alpha := uint8(g.screenShake.intensity * 10)
		if alpha > 100 {
			alpha = 100
		}
		flashColor := color.RGBA{255, 0, 0, alpha}
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, flashColor, false)
	}
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

// drawPlatforms - отрисовка платформ
func (g *Game) drawPlatforms(screen *ebiten.Image) {
	for _, platform := range g.platforms {
		// Platform base (brown/gray stone)
		platformColor := color.RGBA{139, 119, 101, 255}
		vector.DrawFilledRect(screen, platform.x, platform.y, platform.width, platform.height, platformColor, false)

		// Grass top on platform
		grassColor := color.RGBA{34, 139, 34, 255}
		vector.DrawFilledRect(screen, platform.x, platform.y, platform.width, 5, grassColor, false)

		// Brick pattern
		brickColor := color.RGBA{100, 80, 60, 255}
		for bx := platform.x + 5; bx < platform.x+platform.width-5; bx += 20 {
			vector.StrokeLine(screen, bx, platform.y+5, bx, platform.y+platform.height-2, 1, brickColor, false)
		}
		for by := platform.y + 10; by < platform.y+platform.height-2; by += 8 {
			vector.StrokeLine(screen, platform.x+5, by, platform.x+platform.width-5, by, 1, brickColor, false)
		}
	}
}

// drawCoins - отрисовка монет
func (g *Game) drawCoins(screen *ebiten.Image) {
	for _, coin := range g.coins {
		if coin.collected {
			continue
		}

		// Animate coin (spinning effect)
		animOffset := float32(math.Sin(float64(coin.animFrame)*0.1)) * 5
		coinY := coin.y + animOffset

		// Coin outer ring (gold)
		coinColor := color.RGBA{255, 215, 0, 255}
		vector.DrawFilledCircle(screen, coin.x+10, coinY+10, 10, coinColor, false)

		// Coin inner circle (lighter gold)
		innerColor := color.RGBA{255, 235, 100, 255}
		vector.DrawFilledCircle(screen, coin.x+10, coinY+10, 7, innerColor, false)

		// Dollar sign / star in center
		centerColor := color.RGBA{200, 170, 0, 255}
		vector.DrawFilledCircle(screen, coin.x+10, coinY+10, 3, centerColor, false)

		// Glow effect
		glowColor := color.RGBA{255, 215, 0, 100}
		vector.DrawFilledCircle(screen, coin.x+10, coinY+10, 12, glowColor, false)
	}
}

// drawEnemies - отрисовка врагов
func (g *Game) drawEnemies(screen *ebiten.Image) {
	for _, enemy := range g.enemies {
		if !enemy.alive {
			// Draw defeated enemy (squashed)
			squashedColor := color.RGBA{139, 69, 19, 255}
			vector.DrawFilledRect(screen, enemy.x, enemy.y+enemy.height-5, enemy.width, 5, squashedColor, false)
			continue
		}

		// Enemy body (Goomba-style - brown mushroom)
		bodyColor := color.RGBA{139, 69, 19, 255}
		vector.DrawFilledCircle(screen, enemy.x+enemy.width/2, enemy.y+enemy.height/2, enemy.width/2, bodyColor, false)

		// Enemy head (darker brown)
		headColor := color.RGBA{100, 50, 20, 255}
		vector.DrawFilledCircle(screen, enemy.x+enemy.width/2, enemy.y+enemy.height/3, enemy.width/2-3, headColor, false)

		// Eyes (white with black pupils)
		eyeOffset := enemy.vx * 2 // Look in movement direction
		leftEyeX := enemy.x + enemy.width/3 + float32(eyeOffset)
		rightEyeX := enemy.x + 2*enemy.width/3 + float32(eyeOffset)
		eyeY := enemy.y + enemy.height/3

		vector.DrawFilledCircle(screen, leftEyeX, eyeY, 5, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledCircle(screen, rightEyeX, eyeY, 5, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledCircle(screen, leftEyeX+float32(eyeOffset), eyeY, 2, color.RGBA{0, 0, 0, 255}, false)
		vector.DrawFilledCircle(screen, rightEyeX+float32(eyeOffset), eyeY, 2, color.RGBA{0, 0, 0, 255}, false)

		// Feet (animated)
		footOffset := float32(math.Sin(float64(enemy.animFrame)*0.3)) * 3
		footColor := color.RGBA{50, 30, 10, 255}
		vector.DrawFilledCircle(screen, enemy.x+8-footOffset, enemy.y+enemy.height-5, 6, footColor, false)
		vector.DrawFilledCircle(screen, enemy.x+enemy.width-8+footOffset, enemy.y+enemy.height-5, 6, footColor, false)
	}
}

// drawPowerups - отрисовка бонусов
func (g *Game) drawPowerups(screen *ebiten.Image) {
	for _, powerup := range g.powerups {
		if powerup.collected {
			continue
		}

		// Animate powerup (bobbing effect)
		bobOffset := float32(math.Sin(float64(powerup.animFrame)*0.1)) * 3

		switch powerup.powerType {
		case MushroomPower:
			// Draw mushroom (red cap with white spots)
			capY := powerup.y - 10 + bobOffset

			// Mushroom cap (red semicircle)
			capColor := color.RGBA{220, 20, 60, 255}
			vector.DrawFilledCircle(screen, powerup.x, capY, 15, capColor, false)

			// White spots on cap
			spotColor := color.RGBA{255, 255, 255, 255}
			vector.DrawFilledCircle(screen, powerup.x-5, capY-3, 4, spotColor, false)
			vector.DrawFilledCircle(screen, powerup.x+5, capY-3, 4, spotColor, false)

			// Mushroom stem (white)
			stemColor := color.RGBA{255, 250, 240, 255}
			vector.DrawFilledRect(screen, powerup.x-6, powerup.y+5, 12, 12, stemColor, false)

		case StarPower:
			// Draw star (yellow rotating star)
			starY := powerup.y + bobOffset
			starColor := color.RGBA{255, 215, 0, 255}

			// Draw star as a circle with glow for simplicity
			vector.DrawFilledCircle(screen, powerup.x, starY, 12, starColor, false)

			// Star glow effect
			glowColor := color.RGBA{255, 215, 0, 100}
			vector.DrawFilledCircle(screen, powerup.x, starY, 18, glowColor, false)

		case ExtraLifePower:
			// Draw extra life (heart)
			heartY := powerup.y + bobOffset
			heartColor := color.RGBA{255, 20, 60, 255}

			// Heart shape (two circles + triangle)
			vector.DrawFilledCircle(screen, powerup.x-6, heartY-5, 8, heartColor, false)
			vector.DrawFilledCircle(screen, powerup.x+6, heartY-5, 8, heartColor, false)
			vector.DrawFilledRect(screen, powerup.x-8, heartY-5, 16, 12, heartColor, false)
			vector.DrawFilledCircle(screen, powerup.x, heartY+2, 7, heartColor, false)

			// Heart shine
			shineColor := color.RGBA{255, 100, 150, 200}
			vector.DrawFilledCircle(screen, powerup.x-4, heartY-6, 3, shineColor, false)
		}
	}
}

// drawJumpParticles - отрисовка частиц прыжка
func (g *Game) drawJumpParticles(screen *ebiten.Image) {
	for _, particle := range g.jumpParticles {
		if particle.life <= 0 {
			continue
		}

		// Calculate alpha based on life
		alpha := uint8(255 * particle.life / particle.maxLife)
		particleColor := color.RGBA{200, 200, 200, alpha}

		vector.DrawFilledCircle(screen, particle.x, particle.y, particle.size, particleColor, false)
	}
}

// drawSparkParticles - отрисовка частиц искр
func (g *Game) drawSparkParticles(screen *ebiten.Image) {
	for _, particle := range g.sparkParticles {
		if particle.life <= 0 {
			continue
		}

		// Calculate alpha based on life
		alpha := uint8(255 * particle.life / particle.maxLife)
		particleColor := color.RGBA{particle.color.R, particle.color.G, particle.color.B, alpha}

		vector.DrawFilledCircle(screen, particle.x, particle.y, particle.size, particleColor, false)
	}
}

// drawFloatingTexts - отрисовка всплывающего текста
func (g *Game) drawFloatingTexts(screen *ebiten.Image) {
	for _, textData := range g.floatingTexts {
		if textData.life <= 0 || textData.text == "" {
			continue
		}

		// Draw shadow
		ebitenutil.DebugPrintAt(screen, textData.text, int(textData.x)+1, int(textData.y)+1)
		
		// Draw main text
		ebitenutil.DebugPrintAt(screen, textData.text, int(textData.x), int(textData.y))
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Draw inventory bar
	g.drawInventory(screen)
	
	// Draw UI background panel (semi-transparent dark)
	uiBgColor := color.RGBA{0, 0, 0, 180}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, 50, uiBgColor, false)

	// Score display with icon
	scoreText := fmt.Sprintf("🍎 %d", g.player.score)
	ebitenutil.DebugPrintAt(screen, scoreText, 15, 12)

	// Coins display
	coinText := fmt.Sprintf("🪙 %d", g.player.coins)
	ebitenutil.DebugPrintAt(screen, coinText, 120, 12)

	// Lives display (hearts)
	livesText := "❤️ "
	for i := 0; i < g.player.lives; i++ {
		livesText += "❤️ "
	}
	ebitenutil.DebugPrintAt(screen, livesText, 220, 12)

	// Time of day indicator
	timeIcon := "☀️"
	if g.timeOfDay == Night {
		timeIcon = "🌙"
	}
	timeText := fmt.Sprintf("%s", timeIcon)
	ebitenutil.DebugPrintAt(screen, timeText, 350, 12)

	// Weather indicator
	weatherIcon := "☀️"
	if g.weather == Stormy {
		weatherIcon = "⛈️"
	}
	weatherText := fmt.Sprintf("%s", weatherIcon)
	ebitenutil.DebugPrintAt(screen, weatherText, 390, 12)

	// Invincibility indicator (if player is invincible)
	if g.player.invincible > 0 {
		invText := "✨ PROTECTED ✨"
		ebitenutil.DebugPrintAt(screen, invText, 500, 12)
	}

	// Controls hint (bottom of screen)
	controlsBgColor := color.RGBA{0, 0, 0, 120}
	vector.DrawFilledRect(screen, 0, screenHeight-30, screenWidth, 30, controlsBgColor, false)

	controlsText := "⬅️➡️/WASD: Move | ⬆️/Space: Jump | E: Enter/Use | ESC: Menu | 1-9: Select Block | LMB: Mine/Place"
	ebitenutil.DebugPrintAt(screen, controlsText, 10, screenHeight-25)

	// Draw level progress (coins collected / total)
	totalCoins := len(g.coins)
	collectedCoins := 0
	for _, coin := range g.coins {
		if coin.collected {
			collectedCoins++
		}
	}
	progressText := fmt.Sprintf("🪙 %d/%d", collectedCoins, totalCoins)
	ebitenutil.DebugPrintAt(screen, progressText, screenWidth-100, 12)
}

// drawInventory - отрисовка инвентаря
func (g *Game) drawInventory(screen *ebiten.Image) {
	// Draw inventory bar at bottom
	barY := screenHeight - 60
	barHeight := 50
	
	// Background
	vector.DrawFilledRect(screen, 10, float32(barY), float32(inventorySize*52+20), float32(barHeight), color.RGBA{0, 0, 0, 200}, false)
	
	// Draw slots
	for i := 0; i < inventorySize; i++ {
		slotX := 20 + i*52
		slotY := barY + 5
		
		// Slot background
		slotColor := color.RGBA{80, 80, 80, 255}
		if i == g.inventory.selected {
			slotColor = color.RGBA{255, 215, 0, 255} // Gold for selected slot
		}
		vector.DrawFilledRect(screen, float32(slotX), float32(slotY), 48, 48, slotColor, false)
		vector.StrokeRect(screen, float32(slotX), float32(slotY), 48, 48, 2, color.RGBA{150, 150, 150, 255}, false)
		
		// Draw item
		slot := g.inventory.slots[i]
		if slot.count > 0 {
			itemColor := getBlockColor(slot.item)
			vector.DrawFilledRect(screen, float32(slotX+8), float32(slotY+8), 32, 32, itemColor, false)
			
			// Draw count
			countText := fmt.Sprintf("%d", slot.count)
			countX := slotX + 32
			countY := slotY + 32
			ebitenutil.DebugPrintAt(screen, countText, countX, countY)
		}
	}
	
	// Crafting hint
	if g.state == Crafting {
		craftHint := "Press 1-5 to craft | ESC to exit"
		ebitenutil.DebugPrintAt(screen, craftHint, 20, barY-25)
	}
}

// drawTutorial - отрисовка туториала
func (g *Game) drawTutorial(screen *ebiten.Image) {
	if g.tutorial == nil || !g.tutorial.visible {
		return
	}
	
	// Draw tutorial panel on left side
	panelX := 10
	panelY := 60
	panelW := 280
	panelH := 200
	
	// Background
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 2, color.RGBA{255, 215, 0, 255}, false)
	
	// Title
	title := "📖 TUTORIAL"
	ebitenutil.DebugPrintAt(screen, title, panelX+10, panelY+10)
	
	// Draw steps
	for i, step := range g.tutorial.steps {
		y := panelY + 35 + i*25
		marker := "⬜"

		if step.completed {
			marker = "✅"
		} else if i == g.tutorial.currentStep {
			marker = "➡️"
		}

		text := fmt.Sprintf("%s %s", marker, step.title)
		ebitenutil.DebugPrintAt(screen, text, panelX+10, y)
	}
}

// drawQuests - отрисовка квестов
func (g *Game) drawQuests(screen *ebiten.Image) {
	// Draw quests panel on right side
	panelX := screenWidth - 250
	panelY := 60
	panelW := 240
	panelH := 180
	
	// Background
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 2, color.RGBA{255, 100, 100, 255}, false)
	
	// Title
	title := "📜 QUESTS"
	ebitenutil.DebugPrintAt(screen, title, panelX+10, panelY+10)
	
	// Draw quests
	for i, quest := range g.quests {
		y := panelY + 35 + i*30
		marker := "⬜"
		
		if quest.completed {
			marker = "✅"
		}
		
		// Quest title
		text := fmt.Sprintf("%s %s", marker, quest.title)
		ebitenutil.DebugPrintAt(screen, text, panelX+10, y)
		
		// Quest objective
		objText := fmt.Sprintf("   %s", quest.objective)
		ebitenutil.DebugPrintAt(screen, objText, panelX+10, y+15)
	}
}

// drawCheckpoints - отрисовка чекпоинтов
func (g *Game) drawCheckpoints(screen *ebiten.Image) {
	for _, cp := range g.checkpoints {
		// Only draw if on screen
		drawX := float32(cp.x) - float32(g.camera.x)
		drawY := float32(cp.y) - float32(g.camera.y)
		
		if drawX < -50 || drawX > screenWidth+50 || drawY < -50 || drawY > screenHeight+50 {
			continue
		}
		
		if cp.activated {
			// Activated checkpoint (blue flag)
			vector.StrokeLine(screen, drawX+20, drawY-40, drawX+20, drawY, 3, color.RGBA{0, 255, 255, 255}, false)
			vector.DrawFilledRect(screen, drawX+20, drawY-40, 25, 15, color.RGBA{0, 255, 255, 200}, false)
		} else {
			// Inactive checkpoint (gray flag)
			vector.StrokeLine(screen, drawX+20, drawY-40, drawX+20, drawY, 3, color.RGBA{128, 128, 128, 255}, false)
			vector.DrawFilledRect(screen, drawX+20, drawY-40, 25, 15, color.RGBA{128, 128, 128, 200}, false)
		}
	}
}

// drawHealthPacks - отрисовка аптечек
func (g *Game) drawHealthPacks(screen *ebiten.Image) {
	for _, pack := range g.healthPacks {
		if pack.collected {
			continue
		}
		
		// Only draw if on screen
		drawX := float32(pack.x) - float32(g.camera.x)
		drawY := float32(pack.y) - float32(g.camera.y)
		
		if drawX < -50 || drawX > screenWidth+50 || drawY < -50 || drawY > screenHeight+50 {
			continue
		}
		
		// Draw health pack (red cross)
		vector.DrawFilledRect(screen, drawX+5, drawY-15, 10, 30, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledRect(screen, drawX, drawY-5, 20, 10, color.RGBA{255, 0, 0, 255}, false)
		
		// Glow effect
		vector.DrawFilledCircle(screen, drawX+10, drawY, 20, color.RGBA{255, 0, 0, 50}, false)
	}
}

// drawControlsHint - отрисовка подсказок управления
func (g *Game) drawControlsHint(screen *ebiten.Image) {
	if !g.showControls {
		return
	}
	
	// Draw controls panel
	panelX := screenWidth/2 - 200
	panelY := screenHeight/2 - 150
	panelW := 400
	panelH := 300
	
	// Background
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{0, 0, 0, 220}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 3, color.RGBA{255, 215, 0, 255}, false)
	
	// Title
	title := "🎮 CONTROLS"
	ebitenutil.DebugPrintAt(screen, title, panelX+150, panelY+15)
	
	// Controls list
	controls := []string{
		"",
		"⬅️➡️ / AD - Движение",
		"⬆️ / W / SPACE - Прыжок",
		"",
		"🖱️ ЛКМ - Добыча блока",
		"🖱️ ПКМ - Размещение блока",
		"",
		"1-9 - Выбор слота",
		"Колёсико - Прокрутка",
		"",
		"H - Скрыть/Показать подсказки",
		"B - Открыть альбом достижений",
		"ESC - Меню",
	}
	
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, panelX+20, panelY+45+i*22)
	}
}

// drawAchievementAlbum - отрисовка альбома достижений
func (g *Game) drawAchievementAlbum(screen *ebiten.Image) {
	if g.album == nil || !g.album.showAlbum {
		return
	}
	
	// Draw album background (full screen overlay)
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 230}, false)
	
	// Album title
	title := "🏆 ACHIEVEMENT ALBUM"
	titleX := screenWidth/2 - len(title)*8
	ebitenutil.DebugPrintAt(screen, title, titleX, 20)
	
	// Stats
	statsText := fmt.Sprintf("Unlocked: %d / %d", g.album.totalUnlocked, len(g.album.achievements))
	ebitenutil.DebugPrintAt(screen, statsText, 20, 55)
	
	// Draw achievements grid (3 columns)
	startX := 50
	startY := 90
	cellW := 220
	cellH := 100
	
	for i, ach := range g.album.achievements {
		x := startX + (i%3)*cellW
		y := startY + (i/3)*cellH
		
		// Card background
		cardColor := color.RGBA{50, 50, 50, 200}
		if ach.completed {
			cardColor = GetMedalColor(ach.medalTier)
			cardColor.A = 80
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(cellW), float32(cellH), cardColor, false)
		
		// Border
		borderColor := color.RGBA{100, 100, 100, 255}
		if ach.completed {
			borderColor = GetMedalColor(ach.medalTier)
		}
		vector.StrokeRect(screen, float32(x), float32(y), float32(cellW), float32(cellH), 2, borderColor, false)
		
		// Icon/Medal
		icon := "🔒"
		if ach.completed {
			icon = ach.icon
		}
		ebitenutil.DebugPrintAt(screen, icon, x+10, y+10)
		
		// Title
		titleColor := color.RGBA{200, 200, 200, 255}
		if ach.completed {
			titleColor = GetMedalColor(ach.medalTier)
		}
		_ = titleColor
		ebitenutil.DebugPrintAt(screen, ach.title, x+40, y+12)
		
		// Description
		descY := y + 35
		ebitenutil.DebugPrintAt(screen, ach.description, x+10, descY)
		
		// Tier
		tierY := y + 60
		tierText := GetTierName(ach.medalTier)
		ebitenutil.DebugPrintAt(screen, tierText, x+10, tierY)
		
		// Progress bar
		if !ach.completed {
			barY := y + 80
			barW := cellW - 20
			barH := 10
			
			// Background
			vector.DrawFilledRect(screen, float32(x+10), float32(barY), float32(barW), float32(barH), color.RGBA{80, 80, 80, 255}, false)
			
			// Progress
			progressW := int(float32(barW) * float32(ach.progress) / float32(ach.maxProgress))
			if progressW > 0 {
				vector.DrawFilledRect(screen, float32(x+10), float32(barY), float32(progressW), float32(barH), color.RGBA{100, 200, 100, 255}, false)
			}
			
			// Progress text
			progressText := fmt.Sprintf("%d/%d", ach.progress, ach.maxProgress)
			ebitenutil.DebugPrintAt(screen, progressText, x+10, barY+15)
		} else {
			// Unlocked text
			ebitenutil.DebugPrintAt(screen, "✓ UNLOCKED", x+10, y+80)
		}
	}
	
	// Close hint
	closeHint := "Press B to close"
	ebitenutil.DebugPrintAt(screen, closeHint, screenWidth/2-80, screenHeight-40)
}

// drawAchievementNotification - уведомление о получении ачивки
func (g *Game) drawAchievementNotification(screen *ebiten.Image, ach Achievement) {
	if !ach.completed {
		return
	}
	
	// Draw notification at bottom center
	panelX := screenWidth/2 - 150
	panelY := screenHeight - 100
	panelW := 300
	panelH := 80
	
	// Background gradient (gold for achievement)
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), color.RGBA{255, 215, 0, 200}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 3, color.RGBA{255, 255, 255, 255}, false)
	
	// Icon
	ebitenutil.DebugPrintAt(screen, ach.icon, panelX+10, panelY+20)
	
	// Title
	ebitenutil.DebugPrintAt(screen, "ACHIEVEMENT UNLOCKED!", panelX+50, panelY+15)
	
	// Achievement name
	ebitenutil.DebugPrintAt(screen, ach.title, panelX+50, panelY+35)
	
	// Tier
	tierText := GetTierName(ach.medalTier)
	ebitenutil.DebugPrintAt(screen, tierText, panelX+50, panelY+55)
}

// drawCurrentHint - отрисовка текущей подсказки туториала
func (g *Game) drawCurrentHint(screen *ebiten.Image) {
	if g.tutorial == nil || !g.tutorial.visible {
		return
	}
	
	hint := g.tutorial.GetCurrentHint()
	if hint == "" || !g.tutorial.showHint {
		return
	}
	
	// Draw hint at top center
	hintX := screenWidth/2 - len(hint)*6
	hintY := 55
	
	// Background
	vector.DrawFilledRect(screen, float32(hintX-10), float32(hintY-5), float32(len(hint)*12+20), 25, color.RGBA{0, 0, 0, 180}, false)
	
	// Text
	ebitenutil.DebugPrintAt(screen, hint, hintX, hintY)
}

// getBlockColor возвращает цвет блока
func getBlockColor(block BlockType) color.RGBA {
	switch block {
	case Dirt:
		return color.RGBA{139, 69, 19, 255}
	case Grass:
		return color.RGBA{34, 139, 34, 255}
	case Stone:
		return color.RGBA{128, 128, 128, 255}
	case Wood:
		return color.RGBA{101, 67, 33, 255}
	case Leaves:
		return color.RGBA{34, 100, 34, 255}
	case Sand:
		return color.RGBA{238, 214, 130, 255}
	case Coal_Ore:
		return color.RGBA{50, 50, 50, 255}
	case Iron_Ore:
		return color.RGBA{205, 127, 50, 255}
	case Gold_Ore:
		return color.RGBA{255, 215, 0, 255}
	case Diamond_Ore:
		return color.RGBA{0, 255, 255, 255}
	case Bricks:
		return color.RGBA{178, 34, 34, 255}
	case Plank:
		return color.RGBA{222, 184, 135, 255}
	case Crafting_Table:
		return color.RGBA{139, 90, 43, 255}
	default:
		return color.RGBA{255, 0, 255, 255} // Magenta for unknown
	}
}

// drawWorld - отрисовка мира
func (g *Game) drawWorld(screen *ebiten.Image) {
	if g.world == nil {
		return
	}
	
	// Calculate visible area based on camera
	startX := int(g.camera.x) / blockSize
	startY := int(g.camera.y) / blockSize
	endX := startX + screenWidth/blockSize + 2
	endY := startY + screenHeight/blockSize + 2
	
	// Draw blocks
	for x := startX; x < endX && x < g.world.width; x++ {
		for y := startY; y < endY && y < g.world.height; y++ {
			block := g.world.blocks[x][y]
			if block.typ != Air {
				drawX := float32(x*blockSize) - float32(g.camera.x)
				drawY := float32(y*blockSize) - float32(g.camera.y)
				
				blockColor := getBlockColor(block.typ)
				vector.DrawFilledRect(screen, drawX, drawY, blockSize, blockSize, blockColor, false)
				
				// Add some texture/detail
				if block.typ == Stone || block.typ == Coal_Ore || block.typ == Iron_Ore {
					// Add some noise
					if (x+y)%3 == 0 {
						darkerColor := color.RGBA{
							R: blockColor.R - 20,
							G: blockColor.G - 20,
							B: blockColor.B - 20,
							A: 255,
						}
						vector.DrawFilledRect(screen, drawX+5, drawY+5, 10, 10, darkerColor, false)
					}
				}
			}
		}
	}
}

// mineBlock - добыча блока
func (g *Game) mineBlock(screenX, screenY int) {
	if g.world == nil {
		return
	}
	
	// Convert screen coordinates to world coordinates
	worldX := int((float64(screenX) + g.camera.x) / blockSize)
	worldY := int((float64(screenY) + g.camera.y) / blockSize)
	
	block := g.world.GetBlock(worldX, worldY)
	if block != nil && block.minable && block.typ != Air {
		// Add to inventory
		g.inventory.AddItem(block.typ, 1)
		g.audio.PlayCollect()
		g.spawnSparkParticles(float32(screenX), float32(screenY), 10, getBlockColor(block.typ))
		
		// Remove block
		block.typ = Air
		block.solid = false
	}
}

// placeBlock - размещение блока
func (g *Game) placeBlock(screenX, screenY int) {
	if g.world == nil {
		return
	}
	
	// Get selected block from inventory
	selectedSlot := g.inventory.slots[g.inventory.selected]
	if selectedSlot.count <= 0 {
		return
	}
	
	// Convert screen coordinates to world coordinates
	worldX := int((float64(screenX) + g.camera.x) / blockSize)
	worldY := int((float64(screenY) + g.camera.y) / blockSize)
	
	block := g.world.GetBlock(worldX, worldY)
	if block != nil && block.typ == Air {
		// Place block
		block.typ = selectedSlot.item
		block.solid = true
		block.minable = true
		
		// Remove from inventory
		g.inventory.RemoveItem(selectedSlot.item, 1)
		g.audio.PlayJumpBump()
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

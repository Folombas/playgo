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

	// Card dimensions
	CardWidth  = 140
	CardHeight = 200

	// Game states
	StateMenu      = 0
	StateMap       = 1
	StateBattle    = 2
	StateCardSelect = 3
	StateSettings  = 4
	StateGameOver  = 5
	StateVictory   = 6

	// Card types
	CardAttack  = 1
	CardDefense = 2
	CardBuff    = 3
	CardDebuff  = 4
	CardSpecial = 5

	// Rarity
	Common    = 1
	Uncommon  = 2
	Rare      = 3
	Legendary = 4

	// Max values
	MaxDeckSize    = 20
	MaxHandSize    = 7
	MaxEnergy      = 10
	MaxHealth      = 100
)

// ============================================================================
// CARD DEFINITIONS
// ============================================================================

type Card struct {
	id          int
	name        string
	description string
	cardType    int
	cost        int
	damage      int
	block       int
	effect      string
	rarity      int
	image       *ebiten.Image
	selected    bool
	canPlay     bool
	
	// Evolution/Upgrade system
	level       int      // +1 = +damage/+block
	upgraded    bool     // Enhanced version
	enchantments []string // Additional effects
}

// Achievement - система достижений
type Achievement struct {
	id          string
	name        string
	description string
	unlocked    bool
	progress    int
	requirement int
}

var achievementDatabase = map[string]Achievement{
	"first_blood": {id: "first_blood", name: "Первая кровь", description: "Победите первого врага", unlocked: false, requirement: 1},
	"deck_master": {id: "deck_master", name: "Мастер колоды", description: "Соберите 20 карт", unlocked: false, requirement: 20},
	"relic_hunter": {id: "relic_hunter", name: "Охотник за реликвиями", description: "Найдите 5 реликвий", unlocked: false, requirement: 5},
	"floor_10": {id: "floor_10", name: "Покоритель", description: "Достигните 10 этажа", unlocked: false, requirement: 10},
	"floor_20": {id: "floor_20", name: "Легенда", description: "Достигните 20 этажа", unlocked: false, requirement: 20},
	"no_damage": {id: "no_damage", name: "Неуязвимый", description: "Пройдите этаж без урона", unlocked: false, requirement: 1},
	"combo_master": {id: "combo_master", name: "Комбо-мастер", description: "Сыграйте 10 карт за ход", unlocked: false, requirement: 10},
	"rich": {id: "rich", name: "Богач", description: "Накопите 100 золота", unlocked: false, requirement: 100},
}

// Perk/Talent - перки игрока
type Perk struct {
	id          string
	name        string
	description string
	tier        int      // 1, 2, 3
	purchased   bool
}

var perkDatabase = map[string]Perk{
	"strength_1": {id: "strength_1", name: "Сила I", description: "+2 урона атаками", tier: 1},
	"strength_2": {id: "strength_2", name: "Сила II", description: "+4 урона атаками", tier: 2},
	"strength_3": {id: "strength_3", name: "Сила III", description: "+6 урона атаками", tier: 3},
	"vitality_1": {id: "vitality_1", name: "Жизнеспособность I", description: "+10 HP", tier: 1},
	"vitality_2": {id: "vitality_2", name: "Жизнеспособность II", description: "+25 HP", tier: 2},
	"vitality_3": {id: "vitality_3", name: "Жизнеспособность III", description: "+50 HP", tier: 3},
	"wisdom_1": {id: "wisdom_1", name: "Мудрость I", description: "+1 энергия", tier: 2},
	"wisdom_2": {id: "wisdom_2", name: "Мудрость II", description: "+2 энергия", tier: 3},
	"luck_1": {id: "luck_1", name: "Удача I", description: "Шанс найти редкую карту", tier: 1},
	"luck_2": {id: "luck_2", name: "Удача II", description: "Больше редкости", tier: 2},
	"defense_1": {id: "defense_1", name: "Оборона I", description: "+3 блок к картам", tier: 1},
	"defense_2": {id: "defense_2", name: "Оборона II", description: "+6 блок к картам", tier: 2},
}

var cardDatabase = map[int]Card{
	// === COMMON CARDS ===
	1:  {id: 1, name: "Удар", description: "Нанесите 6 урона", cardType: CardAttack, cost: 1, damage: 6, rarity: Common},
	2:  {id: 2, name: "Защита", description: "Получите 5 блока", cardType: CardDefense, cost: 1, block: 5, rarity: Common},
	3:  {id: 3, name: "Мощный удар", description: "Нанесите 12 урона", cardType: CardAttack, cost: 2, damage: 12, rarity: Uncommon},
	4:  {id: 4, name: "Щит", description: "Получите 12 блока", cardType: CardDefense, cost: 2, block: 12, rarity: Uncommon},
	5:  {id: 5, name: "Огненный шар", description: "Нанесите 20 урона всем врагам", cardType: CardAttack, cost: 3, damage: 20, rarity: Rare},
	6:  {id: 6, name: "Боевой клич", description: "+3 урона в этом ходу", cardType: CardBuff, cost: 1, effect: "buff_attack", rarity: Common},
	7:  {id: 7, name: "Яд", description: "Враг получает 3 урона каждый ход", cardType: CardDebuff, cost: 1, effect: "poison", rarity: Uncommon},
	8:  {id: 8, name: "Лечение", description: "Восстановите 10 HP", cardType: CardBuff, cost: 2, effect: "heal_10", rarity: Uncommon},
	9:  {id: 9, name: "Мега удар", description: "Нанесите 30 урона", cardType: CardAttack, cost: 4, damage: 30, rarity: Legendary},
	10: {id: 10, name: "Неуязвимость", description: "Получите 25 блока и неуязвимость", cardType: CardDefense, cost: 3, block: 25, effect: "invincible", rarity: Rare},
	11: {id: 11, name: "Кровожадность", description: "Нанесите 8 урона, получите 5 HP", cardType: CardAttack, cost: 2, damage: 8, effect: "lifesteal", rarity: Rare},
	12: {id: 12, name: "Контратака", description: "Получите 8 блока, отразите урон", cardType: CardDefense, cost: 1, block: 8, effect: "reflect", rarity: Uncommon},
	
	// === NEW CARDS ===
	13: {id: 13, name: "Смерч", description: "Нанесите 8 урона всем врагам 3 раза", cardType: CardAttack, cost: 3, damage: 8, effect: "aoe_3x", rarity: Legendary},
	14: {id: 14, name: "Каменная кожа", description: "Получите 15 блока, +5 блок каждый ход", cardType: CardDefense, cost: 2, block: 15, effect: "growing_block", rarity: Rare},
	15: {id: 15, name: "Ярость", description: "Получите 1 энергию, наносите +50% урона", cardType: CardBuff, cost: 0, effect: "rage", rarity: Rare},
	16: {id: 16, name: "Слабость", description: "Враг наносит на 40% меньше урона", cardType: CardDebuff, cost: 1, effect: "weak", rarity: Common},
	17: {id: 17, name: "Уязвимость", description: "Враг получает на 50% больше урона", cardType: CardDebuff, cost: 2, effect: "vulnerable", rarity: Uncommon},
	18: {id: 18, name: "Молния", description: "Нанесите 10 урона, игнорируя блок", cardType: CardAttack, cost: 2, damage: 10, effect: "pierce", rarity: Uncommon},
	19: {id: 19, name: "Вампиризм", description: "Нанесите 12 урона, украдите 6 HP", cardType: CardAttack, cost: 3, damage: 12, effect: "steal_hp", rarity: Rare},
	20: {id: 20, name: "Благословение", description: "Получите 20 блока и 5 HP", cardType: CardBuff, cost: 2, block: 20, effect: "heal_5", rarity: Uncommon},
	21: {id: 21, name: "Двойной удар", description: "Нанесите 5 урона дважды", cardType: CardAttack, cost: 1, damage: 5, effect: "double_strike", rarity: Common},
	22: {id: 22, name: "Ледяная броня", description: "Получите 10 блока, заморозьте врага", cardType: CardDefense, cost: 2, block: 10, effect: "freeze", rarity: Rare},
	23: {id: 23, name: "Тёмная магия", description: "Нанесите 25 урона, получите 5 урона", cardType: CardAttack, cost: 2, damage: 25, effect: "self_damage", rarity: Rare},
	24: {id: 24, name: "Медитация", description: "Получите 2 энергии, пропустите ход", cardType: CardBuff, cost: 0, effect: "meditate", rarity: Rare},
	25: {id: 25, name: "Божественный щит", description: "Получите 99 блока", cardType: CardDefense, cost: 5, block: 99, rarity: Legendary},
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Player struct {
	maxHealth    int
	health       int
	block        int
	energy       int
	maxEnergy    int
	damageBuff   int
	poison       int
	poisonTurns  int
	invincible   bool
	reflect      bool
	
	// Relics & Effects
	weak         int      // Уменьшение урона (ходы)
	vulnerable   int      // Увеличение получаемого урона
	frozen       int      // Пропуск хода
	growingBlock int      // Растущий блок за ход
	rage         bool     // +50% урона
	extraEnergy  int      // Дополнительная энергия
	
	// Relics owned
	relics       []Relic
}

type Relic struct {
	id          int
	name        string
	description string
	effect      string
	stacks      int
}

var relicDatabase = map[int]Relic{
	1: {id: 1, name: "Кровавый череп", description: "+1 урона за каждую сыгранную карту", effect: "damage_per_card"},
	2: {id: 2, name: "Перо орла", description: "Начинайте бой с 5 блока", effect: "start_block"},
	3: {id: 3, name: "Амулет жизни", description: "+10 максимального HP", effect: "max_hp"},
	4: {id: 4, name: "Кольцо силы", description: "+2 к начальной энергии", effect: "max_energy"},
	5: {id: 5, name: "Щит предков", description: "Шанс 25% получить 10 блока", effect: "block_chance"},
	6: {id: 6, name: "Книга магии", description: "Карты стоят на 1 меньше (минимум 0)", effect: "cost_reduce"},
	7: {id: 7, name: "Сердце дракона", description: "Наносите +2 урона всеми атаками", effect: "flat_damage"},
	8: {id: 8, name: "Сапоги скорости", description: "Берите +1 карту в начале хода", effect: "draw_extra"},
}

type Enemy struct {
	name        string
	maxHealth   int
	health      int
	block       int
	damage      int
	intent      string // "attack", "defend", "buff"
	intentValue int
	image       *ebiten.Image
	poison      int
	vulnerable  int    // Увеличенный получаемый урон
}

type GameState struct {
	state        int
	player       *Player
	enemies      []*Enemy
	hand         []*Card
	deck         []*Card
	discard      []*Card
	energy       int
	turn         int
	_floor       int
	gold         int
	score        int
	
	// Progression systems
	achievements map[string]*Achievement
	perks        []Perk
	perkPoints   int      // Очки перков за победы
	ascension    int      // Уровень сложности (0-20)
	dailySeed    int64    // Для ежедневных испытаний
	isDailyRun   bool     // Ежедневный забег
	
	// Save system
	saveSlot     int
	totalWins    int
	totalFloors  int
	highestFloor int
	
	// UI State
	selectedCard    int
	hoveredCard     int
	hoveredButton   int
	settingsOpen    bool
	showTooltips    bool
	soundEnabled    bool
	musicVolume     float64
	sfxVolume       float64
	
	// Visual effects
	particles      []*Particle
	screenShake    int
	damageNumbers  []*DamageNumber
	
	// Assets
	gameFont       font.Face
	smallFont      font.Face
	cardImages     map[int]*ebiten.Image
	buttonImages   map[int]*ebiten.Image
	
	// Audio
	audioCtx       *audio.Context
	sounds         map[int][]byte
}

type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	color    color.RGBA
	size     float32
}

type DamageNumber struct {
	x, y     float64
	value    int
	life     int
	isHeal   bool
	isCrit   bool
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
	sounds[0] = GenerateSound(600, 0.1)  // Card play
	sounds[1] = GenerateSound(200, 0.15) // Damage
	sounds[2] = GenerateSound(800, 0.2)  // Heal
	sounds[3] = GenerateSound(400, 0.3)  // Turn start
	sounds[4] = GenerateSound(150, 0.5)  // Death
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
		state: StateMenu,
		player: &Player{
			maxHealth: MaxHealth,
			health:    MaxHealth,
			energy:    3,
			maxEnergy: 3,
			relics:    make([]Relic, 0),
		},
		deck:         make([]*Card, 0),
		hand:         make([]*Card, 0),
		discard:      make([]*Card, 0),
		enemies:      make([]*Enemy, 0),
		particles:    make([]*Particle, 0),
		damageNumbers: make([]*DamageNumber, 0),
		showTooltips: true,
		soundEnabled: true,
		musicVolume:  0.5,
		sfxVolume:    0.5,
		gold:         50,
		_floor:       1,
		achievements: make(map[string]*Achievement),
		perks:        make([]Perk, 0),
		perkPoints:   0,
		ascension:    0,
		totalWins:    0,
		totalFloors:  0,
		highestFloor: 0,
	}

	// Initialize achievements
	for id, ach := range achievementDatabase {
		achCopy := ach
		g.achievements[id] = &achCopy
	}

	// Load fonts
	g.gameFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 24)
	g.smallFont = LoadFont("assets/fonts/SuperFeel-JpZqa.ttf", 16)

	// Load sounds
	g.audioCtx = InitAudio()
	g.sounds = LoadSounds()

	// Initialize starter deck
	g.InitializeStarterDeck()

	return g
}

func (g *GameState) InitializeStarterDeck() {
	// Add starter cards
	starterCards := []int{1, 1, 1, 2, 2, 2, 6, 6}
	for _, cardID := range starterCards {
		if card, ok := cardDatabase[cardID]; ok {
			cardCopy := card
			g.deck = append(g.deck, &cardCopy)
		}
	}
	g.ShuffleDeck()
}

// UpgradeCard - улучшение карты (+урон/+блок)
func (g *GameState) UpgradeCard(cardIndex int) {
	if cardIndex >= 0 && cardIndex < len(g.hand) {
		card := g.hand[cardIndex]
		if !card.upgraded {
			card.upgraded = true
			card.level++
			
			// Улучшение эффектов
			if card.cardType == CardAttack && card.damage > 0 {
				card.damage += 3
			}
			if card.cardType == CardDefense && card.block > 0 {
				card.block += 3
			}
			if card.cost > 0 {
				// Иногда уменьшаем стоимость
				if rand.Float32() < 0.3 {
					card.cost--
				}
			}
			
			// Визуальный эффект
			g.spawnParticles(float64(ScreenWidth/2), float64(ScreenHeight/2), 30, color.RGBA{255, 215, 0, 255})
			PlaySound(g, 5) // Victory sound
		}
	}
}

// CheckAchievements - проверка достижений
func (g *GameState) CheckAchievements() {
	for id, ach := range g.achievements {
		if ach.unlocked {
			continue
		}
		
		unlocked := false
		switch id {
		case "first_blood":
			unlocked = g.totalWins >= 1
		case "deck_master":
			unlocked = len(g.deck) >= 20
		case "relic_hunter":
			unlocked = len(g.player.relics) >= 5
		case "floor_10":
			unlocked = g.highestFloor >= 10
		case "floor_20":
			unlocked = g.highestFloor >= 20
		case "rich":
			unlocked = g.gold >= 100
		case "combo_master":
			unlocked = len(g.hand) <= MaxHandSize-10 // Played 10 cards
		}
		
		if unlocked {
			ach.unlocked = true
			g.achievements[id] = ach
			// Show achievement notification
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/2), y: 100,
				value: 0, life: 180, isHeal: true, // 3 seconds
			})
			PlaySound(g, 5)
		}
	}
}

func (g *GameState) ShuffleDeck() {
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
}

func (g *GameState) DrawCards(count int) {
	for i := 0; i < count && len(g.hand) < MaxHandSize; i++ {
		if len(g.deck) == 0 {
			if len(g.discard) == 0 {
				break
			}
			// Reshuffle discard into deck
			g.deck = append(g.deck, g.discard...)
			g.discard = make([]*Card, 0)
			g.ShuffleDeck()
		}
		if len(g.deck) > 0 {
			g.hand = append(g.hand, g.deck[len(g.deck)-1])
			g.deck = g.deck[:len(g.deck)-1]
		}
	}
}

func (g *GameState) StartBattle() {
	g.state = StateBattle
	g.player.block = 0
	g.player.damageBuff = 0
	g.player.invincible = false
	g.player.reflect = false
	g.hand = make([]*Card, 0)
	g.discard = make([]*Card, 0)
	g.energy = g.player.maxEnergy
	g.turn++

	// Generate enemy based on floor
	enemy := g.GenerateEnemy(g._floor)
	g.enemies = []*Enemy{enemy}

	g.DrawCards(MaxHandSize)
	PlaySound(g, 3)
}

func (g *GameState) GenerateEnemy(floor int) *Enemy {
	enemyTypes := []struct {
		name   string
		hp     int
		damage int
	}{
		{"Гоблин", 30 + floor*5, 6 + floor},
		{"Скелет", 40 + floor*5, 8 + floor},
		{"Орк", 60 + floor*8, 10 + floor},
		{"Тролль", 80 + floor*10, 12 + floor},
		{"Дракон", 150 + floor*15, 18 + floor*2},
	}

	idx := min(floor-1, len(enemyTypes)-1)
	e := enemyTypes[idx]

	return &Enemy{
		name:      e.name,
		maxHealth: e.hp,
		health:    e.hp,
		damage:    e.damage,
		intent:    "attack",
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

func (g *GameState) Update() error {
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

	// Update damage numbers
	for i := len(g.damageNumbers) - 1; i >= 0; i-- {
		d := g.damageNumbers[i]
		d.y -= 0.5
		d.life--
		if d.life <= 0 {
			g.damageNumbers = append(g.damageNumbers[:i], g.damageNumbers[i+1:]...)
		}
	}

	// Screen shake decay
	if g.screenShake > 0 {
		g.screenShake--
	}

	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StateMap:
		g.updateMap()
	case StateBattle:
		g.updateBattle()
	case StateSettings:
		g.updateSettings()
	case StateGameOver, StateVictory:
		g.updateEndScreen()
	}

	return nil
}

func (g *GameState) updateMenu() {
	mx, my := ebiten.CursorPosition()

	// Check button hover
	buttons := []struct{ x, y, w, h int }{
		{ScreenWidth/2 - 150, 250, 300, 60}, // Start
		{ScreenWidth/2 - 150, 330, 300, 60}, // Settings
		{ScreenWidth/2 - 150, 410, 300, 60}, // Quit
	}

	for i, btn := range buttons {
		if mx >= btn.x && mx <= btn.x+btn.w && my >= btn.y && my <= btn.y+btn.h {
			g.hoveredButton = i
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				switch i {
				case 0:
					g.state = StateMap
				case 1:
					g.state = StateSettings
				case 2:
					os.Exit(0)
				}
			}
		}
	}
}

func (g *GameState) updateMap() {
	mx, my := ebiten.CursorPosition()

	// Simple map with nodes
	nodeY := 150
	for floor := 1; floor <= 10; floor++ {
		nodeX := ScreenWidth/2 + (floor-5)*100
		if mx >= nodeX-40 && mx <= nodeX+40 && my >= nodeY-40 && my <= nodeY+40 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && floor == g._floor {
				g.StartBattle()
			}
		}
		nodeY += 50
	}

	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateMenu
		}
	}
}

func (g *GameState) updateBattle() {
	mx, my := ebiten.CursorPosition()

	// Hotkey for upgrade (U key)
	if inpututil.IsKeyJustPressed(ebiten.KeyU) && len(g.hand) > 0 {
		// Upgrade random card or first card
		g.UpgradeCard(0)
	}

	// Check card hover/click
	cardStartX := (ScreenWidth - len(g.hand)*CardWidth - (len(g.hand)-1)*20) / 2
	for i, card := range g.hand {
		x := cardStartX + i*(CardWidth+20)
		y := ScreenHeight - CardHeight - 20

		// Hover effect
		if mx >= x && mx <= x+CardWidth && my >= y && my <= y+CardHeight {
			g.hoveredCard = i
			card.selected = true

			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.playCard(i)
			}
		} else {
			card.selected = false
		}
	}

	// End turn button
	endTurnX := ScreenWidth - 180
	endTurnY := ScreenHeight/2 - 30
	if mx >= endTurnX && mx <= endTurnX+160 && my >= endTurnY && my <= endTurnY+60 {
		g.hoveredButton = 0
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.endTurn()
		}
	}

	// Check if battle is won
	allEnemiesDead := true
	for _, e := range g.enemies {
		if e.health > 0 {
			allEnemiesDead = false
			break
		}
	}
	if allEnemiesDead && len(g.enemies) > 0 {
		g.victory()
	}

	// Check if player is dead
	if g.player.health <= 0 {
		g.state = StateGameOver
		PlaySound(g, 4)
	}
}

func (g *GameState) updateSettings() {
	mx, my := ebiten.CursorPosition()

	// Back button
	if mx >= 20 && mx <= 120 && my >= 20 && my <= 60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateMenu
		}
	}

	// Sound toggle
	toggleX := ScreenWidth/2 + 100
	toggleY := 250
	if mx >= toggleX && mx <= toggleX+60 && my >= toggleY-20 && my <= toggleY+20 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.soundEnabled = !g.soundEnabled
		}
	}

	// Tooltip toggle
	toggleY = 320
	if mx >= toggleX && mx <= toggleX+60 && my >= toggleY-20 && my <= toggleY+20 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.showTooltips = !g.showTooltips
		}
	}

	// Volume sliders
	sliderY := 400
	if mx >= ScreenWidth/2-100 && mx <= ScreenWidth/2+100 && my >= sliderY-10 && my <= sliderY+10 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.sfxVolume = float64(mx-(ScreenWidth/2-100)) / 200
			if g.sfxVolume < 0 {
				g.sfxVolume = 0
			}
			if g.sfxVolume > 1 {
				g.sfxVolume = 1
			}
		}
	}
}

func (g *GameState) updateEndScreen() {
	mx, my := ebiten.CursorPosition()

	// Continue button
	btnX := ScreenWidth/2 - 150
	btnY := ScreenHeight/2 + 50
	if mx >= btnX && mx <= btnX+300 && my >= btnY && my <= btnY+60 {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = StateMenu
			// Reset game
			g.player.health = g.player.maxHealth
			g._floor = 1
			g.gold = 50
			g.deck = make([]*Card, 0)
			g.InitializeStarterDeck()
		}
	}
}

func (g *GameState) playCard(cardIndex int) {
	card := g.hand[cardIndex]
	
	// Calculate cost with relics
	cost := card.cost
	for _, relic := range g.player.relics {
		if relic.effect == "cost_reduce" {
			cost--
		}
	}
	if cost < 0 {
		cost = 0
	}
	
	if g.energy < cost {
		return // Not enough energy
	}

	g.energy -= cost
	PlaySound(g, 0)
	
	// Relic: Damage per card
	damageBonus := 0
	for _, relic := range g.player.relics {
		if relic.effect == "flat_damage" {
			damageBonus += 2
		}
	}

	// Apply card effects
	switch card.cardType {
	case CardAttack:
		damage := card.damage + g.player.damageBuff + damageBonus
		
		// Rage effect
		if g.player.rage {
			damage = damage * 150 / 100
		}
		
		if len(g.enemies) > 0 {
			target := g.enemies[0]
			
			// Pierce effect - ignore block
			if card.effect == "pierce" {
				target.health -= damage
			} else {
				// Normal damage calculation
				actualDamage := damage
				if target.block > 0 {
					if target.block >= actualDamage {
						target.block -= actualDamage
						actualDamage = 0
					} else {
						actualDamage -= target.block
						target.block = 0
					}
				}
				
				// Vulnerable effect
				if target.vulnerable > 0 {
					actualDamage = actualDamage * 150 / 100
				}
				
				target.health -= actualDamage
			}
			
			g.screenShake = 5
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/2), y: float64(ScreenHeight/3),
				value: damage, life: 60,
			})
			g.spawnParticles(float64(ScreenWidth/2), float64(ScreenHeight/3), 10, color.RGBA{255, 0, 0, 255})
			
			// Lifesteal
			if card.effect == "lifesteal" || card.effect == "steal_hp" {
				heal := 5
				if card.effect == "steal_hp" {
					heal = 6
				}
				g.player.health = min(g.player.health+heal, g.player.maxHealth)
				g.damageNumbers = append(g.damageNumbers, &DamageNumber{
					x: float64(ScreenWidth/4), y: float64(ScreenHeight/2),
					value: heal, life: 60, isHeal: true,
				})
			}
			
			// Double strike
			if card.effect == "double_strike" {
				target.health -= damage
				g.damageNumbers = append(g.damageNumbers, &DamageNumber{
					x: float64(ScreenWidth/2), y: float64(ScreenHeight/3),
					value: damage, life: 60,
				})
			}
			
			// AOE 3x
			if card.effect == "aoe_3x" {
				for i := 0; i < 2; i++ {
					target.health -= damage
					g.spawnParticles(float64(ScreenWidth/2), float64(ScreenHeight/3), 5, color.RGBA{255, 100, 0, 255})
				}
			}
			
			// Self damage
			if card.effect == "self_damage" {
				g.player.health -= 5
				g.damageNumbers = append(g.damageNumbers, &DamageNumber{
					x: float64(ScreenWidth/4), y: float64(ScreenHeight/2),
					value: 5, life: 60,
				})
			}
		}
		
	case CardDefense:
		g.player.block += card.block
		
		// Growing block
		if card.effect == "growing_block" {
			g.player.growingBlock = 5
		}
		
		g.spawnParticles(float64(ScreenWidth/4), float64(ScreenHeight/2), 8, color.RGBA{0, 100, 255, 255})
		
	case CardBuff:
		switch card.effect {
		case "buff_attack":
			g.player.damageBuff += 3
		case "heal_10":
			g.player.health = min(g.player.health+10, g.player.maxHealth)
			PlaySound(g, 2)
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/4), y: float64(ScreenHeight/2),
				value: 10, life: 60, isHeal: true,
			})
		case "heal_5":
			g.player.health = min(g.player.health+5, g.player.maxHealth)
			PlaySound(g, 2)
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/4), y: float64(ScreenHeight/2),
				value: 5, life: 60, isHeal: true,
			})
		case "rage":
			g.player.rage = true
			g.player.extraEnergy = 1
		case "meditate":
			g.player.extraEnergy = 2
		case "invincible":
			g.player.invincible = true
		case "reflect":
			g.player.reflect = true
		}
		
	case CardDebuff:
		if len(g.enemies) > 0 {
			switch card.effect {
			case "poison":
				g.enemies[0].poison = 3
			case "weak":
				g.enemies[0].damage = g.enemies[0].damage * 60 / 100
			case "vulnerable":
				g.enemies[0].vulnerable = 3
			case "freeze":
				g.enemies[0].damage = 0 // Frozen for 1 turn
			}
		}
	}

	// Move card to discard
	g.discard = append(g.discard, card)
	g.hand = append(g.hand[:cardIndex], g.hand[cardIndex+1:]...)
}

func (g *GameState) endTurn() {
	// Enemy turn
	for _, enemy := range g.enemies {
		if enemy.health <= 0 {
			continue
		}

		// Apply poison
		if enemy.poison > 0 {
			enemy.health -= enemy.poison
			g.damageNumbers = append(g.damageNumbers, &DamageNumber{
				x: float64(ScreenWidth/2), y: float64(ScreenHeight/3),
				value: enemy.poison, life: 60,
			})
			enemy.poison--
		}
		
		// Apply vulnerable decay
		if enemy.vulnerable > 0 {
			enemy.vulnerable--
		}

		// Enemy attack
		if enemy.intent == "attack" && enemy.damage > 0 {
			damage := enemy.damage
			
			// Weak effect on enemy
			if enemy.damage < g.enemies[0].damage {
				damage = damage * 60 / 100
			}
			
			if g.player.block > 0 {
				if g.player.block >= damage {
					g.player.block -= damage
					damage = 0
				} else {
					damage -= g.player.block
					g.player.block = 0
				}
			}
			if g.player.invincible {
				damage = 0
			}
			
			// Reflect damage
			if g.player.reflect && damage > 0 {
				enemy.health -= damage / 2
				damage = damage / 2
			}
			
			g.player.health -= damage
			if damage > 0 {
				g.screenShake = 10
				g.spawnParticles(float64(ScreenWidth/4), float64(ScreenHeight/2), 15, color.RGBA{255, 0, 0, 255})
			}
		}
	}

	// Start of player turn
	// Apply growing block
	if g.player.growingBlock > 0 {
		g.player.block += g.player.growingBlock
	}
	
	// Apply rage decay
	g.player.rage = false
	g.player.extraEnergy = 0
	
	// Apply weak/vulnerable decay
	if g.player.weak > 0 {
		g.player.weak--
	}
	if g.player.vulnerable > 0 {
		g.player.vulnerable--
	}
	if g.player.frozen > 0 {
		g.player.frozen--
	}

	// Reset player
	g.player.block = 0
	g.player.damageBuff = 0
	g.player.invincible = false
	g.player.reflect = false
	
	// Calculate energy
	g.energy = g.player.maxEnergy
	for _, relic := range g.player.relics {
		if relic.effect == "max_energy" {
			g.energy += 2
		}
	}

	// Draw new cards
	for _, card := range g.hand {
		g.discard = append(g.discard, card)
	}
	g.hand = make([]*Card, 0)
	
	drawCount := MaxHandSize
	for _, relic := range g.player.relics {
		if relic.effect == "draw_extra" {
			drawCount++
		}
	}
	g.DrawCards(drawCount)

	// Enemy intent for next turn
	for _, enemy := range g.enemies {
		if enemy.health > 0 {
			intents := []string{"attack", "defend", "buff"}
			enemy.intent = intents[rand.Intn(len(intents))]
			if enemy.intent == "defend" {
				enemy.block += 10
			} else if enemy.intent == "buff" {
				enemy.damage = enemy.damage * 120 / 100
			}
		}
	}

	PlaySound(g, 3)
}

func (g *GameState) victory() {
	g.state = StateVictory
	g.score += 100 * g._floor
	g.gold += 20
	g._floor++
	g.totalWins++
	g.totalFloors++
	if g._floor > g.highestFloor {
		g.highestFloor = g._floor
	}
	PlaySound(g, 5)

	// Award perk point
	g.perkPoints++

	// Add reward card OR relic
	if len(g.deck) < MaxDeckSize && rand.Float32() < 0.7 {
		// 70% chance: new card
		rewardCards := []int{3, 4, 5, 7, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
		cardID := rewardCards[rand.Intn(len(rewardCards))]
		if card, ok := cardDatabase[cardID]; ok {
			cardCopy := card
			g.deck = append(g.deck, &cardCopy)
		}
	} else if len(g.player.relics) < 5 {
		// 30% chance: new relic
		relicID := rand.Intn(8) + 1
		if relic, ok := relicDatabase[relicID]; ok {
			g.player.relics = append(g.player.relics, relic)
		}
	}
	
	// Heal player after victory
	g.player.health = min(g.player.health+20, g.player.maxHealth)
	
	// Check achievements
	g.CheckAchievements()
	
	// Ascension scaling (harder enemies)
	ascensionBonus := g.ascension * 10
	for _, enemy := range g.enemies {
		enemy.maxHealth += ascensionBonus
		enemy.health = enemy.maxHealth
		enemy.damage += g.ascension * 2
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
	case StateMap:
		g.drawMap(screen)
	case StateBattle:
		g.drawBattle(screen)
	case StateSettings:
		g.drawSettings(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateVictory:
		g.drawVictory(screen)
	}
}

func (g *GameState) drawMenu(screen *ebiten.Image) {
	// Background gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(30 + y/30)
		g := uint8(20 + y/40)
		b := uint8(60 + y/20)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	// Title
	title := "⚔️ GO LEGENDS ⚔️"
	titleX := ScreenWidth/2 - 180
	titleY := 120

	if g.gameFont != nil {
		// Shadow
		text.Draw(screen, title, g.gameFont, titleX+4, titleY+4, color.RGBA{0, 0, 0, 150})
		// Main title with gradient effect
		text.Draw(screen, title, g.gameFont, titleX, titleY, color.RGBA{255, 215, 0, 255})

		subtitle := "Roguelike Card Battler"
		text.Draw(screen, subtitle, g.smallFont, ScreenWidth/2-100, 180, color.RGBA{200, 200, 200, 255})
	}

	// Buttons
	buttons := []string{"🎮 Начать игру", "⚙️ Настройки", "🚪 Выход"}
	buttonY := 250
	for i, btnText := range buttons {
		x := ScreenWidth/2 - 150
		y := buttonY + i*80

		// Button background
		btnColor := color.RGBA{60, 60, 100, 255}
		if i == g.hoveredButton {
			btnColor = color.RGBA{80, 80, 140, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), 300, 60, btnColor, false)
		vector.StrokeRect(screen, float32(x), float32(y), 300, 60, 2, color.RGBA{255, 215, 0, 255}, false)

		// Button text
		if g.gameFont != nil {
			text.Draw(screen, btnText, g.gameFont, ScreenWidth/2-80, y+35, color.RGBA{255, 255, 255, 255})
		}
	}

	// Version info
	if g.smallFont != nil {
		versionText := "Go365 Day 84 | v1.0 | Pure Go + Ebitengine"
		text.Draw(screen, versionText, g.smallFont, 20, ScreenHeight-30, color.RGBA{150, 150, 150, 255})
	}
}

func (g *GameState) drawMap(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{20, 20, 40, 255})

	// Draw path nodes
	nodeY := 150
	for floor := 1; floor <= 10; floor++ {
		nodeX := ScreenWidth/2 + (floor-5)*100

		// Connection line
		if floor > 1 {
			prevX := ScreenWidth/2 + (floor-6)*100
			prevY := nodeY - 50
			vector.StrokeLine(screen, float32(prevX), float32(prevY), float32(nodeX), float32(nodeY), 3, color.RGBA{100, 100, 100, 255}, false)
		}

		// Node circle
		nodeColor := color.RGBA{60, 60, 80, 255}
		if floor == g._floor {
			nodeColor = color.RGBA{255, 215, 0, 255}
		} else if floor < g._floor {
			nodeColor = color.RGBA{100, 200, 100, 255}
		}
		vector.DrawFilledCircle(screen, float32(nodeX), float32(nodeY), 40, nodeColor, false)
		vector.StrokeCircle(screen, float32(nodeX), float32(nodeY), 40, 3, color.RGBA{255, 255, 255, 255}, false)

		// Floor number
		if g.smallFont != nil {
			floorText := fmt.Sprintf("%d", floor)
			text.Draw(screen, floorText, g.smallFont, nodeX-5, nodeY+5, color.RGBA{255, 255, 255, 255})
		}
	}

	// Player stats panel
	g.drawStatsPanel(screen)

	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{100, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}

	// Instructions
	if g.gameFont != nil {
		instruction := "Выберите этаж для боя (жёлтый)"
		text.Draw(screen, instruction, g.gameFont, ScreenWidth/2-150, ScreenHeight-50, color.RGBA{200, 200, 200, 255})
	}
}

func (g *GameState) drawBattle(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{30, 20, 40, 255})

	// Draw enemy
	if len(g.enemies) > 0 && g.enemies[0].health > 0 {
		enemy := g.enemies[0]
		ex := ScreenWidth/2
		ey := ScreenHeight/3

		// Enemy sprite (placeholder)
		vector.DrawFilledCircle(screen, float32(ex), float32(ey), 60, color.RGBA{200, 50, 50, 255}, false)

		// Enemy name
		if g.gameFont != nil {
			text.Draw(screen, enemy.name, g.gameFont, ex-50, ey-80, color.RGBA{255, 100, 100, 255})
		}

		// Health bar
		hpPercent := float32(enemy.health) / float32(enemy.maxHealth)
		vector.DrawFilledRect(screen, float32(ex-80), float32(ey-60), 160, 20, color.RGBA{100, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(ex-80), float32(ey-60), 160*hpPercent, 20, color.RGBA{255, 0, 0, 255}, false)

		// HP text
		if g.smallFont != nil {
			hpText := fmt.Sprintf("%d/%d", enemy.health, enemy.maxHealth)
			text.Draw(screen, hpText, g.smallFont, ex-30, ey-45, color.RGBA{255, 255, 255, 255})
		}

		// Intent indicator
		intentIcon := "⚔️"
		if enemy.intent == "defend" {
			intentIcon = "🛡️"
		} else if enemy.intent == "buff" {
			intentIcon = "✨"
		}
		if g.gameFont != nil {
			intentText := fmt.Sprintf("%s %d", intentIcon, enemy.damage)
			text.Draw(screen, intentText, g.gameFont, ex-40, ey+80, color.RGBA{255, 200, 0, 255})
		}
	}

	// Draw player
	px := ScreenWidth / 4
	py := ScreenHeight / 2

	// Player sprite (placeholder)
	vector.DrawFilledRect(screen, float32(px-30), float32(py-50), 60, 100, color.RGBA{50, 100, 200, 255}, false)

	// Player stats
	g.drawStatsPanel(screen)

	// Energy
	if g.gameFont != nil {
		energyText := fmt.Sprintf("⚡ %d/%d", g.energy, g.player.maxEnergy)
		text.Draw(screen, energyText, g.gameFont, px-40, py+70, color.RGBA{255, 255, 0, 255})
	}

	// Cards in hand
	g.drawHand(screen)

	// End turn button
	endTurnX := ScreenWidth - 180
	endTurnY := ScreenHeight/2 - 30
	btnColor := color.RGBA{150, 50, 50, 255}
	if g.hoveredButton == 0 {
		btnColor = color.RGBA{200, 70, 70, 255}
	}
	vector.DrawFilledRect(screen, float32(endTurnX), float32(endTurnY), 160, 60, btnColor, false)
	vector.StrokeRect(screen, float32(endTurnX), float32(endTurnY), 160, 60, 3, color.RGBA{255, 200, 0, 255}, false)
	if g.gameFont != nil {
		text.Draw(screen, "Конец хода", g.gameFont, endTurnX+15, endTurnY+30, color.RGBA{255, 255, 255, 255})
	}

	// Draw particles
	g.drawParticles(screen)

	// Draw damage numbers
	g.drawDamageNumbers(screen)
}

func (g *GameState) drawHand(screen *ebiten.Image) {
	if len(g.hand) == 0 {
		return
	}

	cardStartX := (ScreenWidth - len(g.hand)*CardWidth - (len(g.hand)-1)*20) / 2
	cardY := ScreenHeight - CardHeight - 20

	for i, card := range g.hand {
		x := cardStartX + i*(CardWidth+20)
		y := cardY

		// Hover effect - lift card
		if i == g.hoveredCard {
			y -= 20
		}

		// Card background
		cardColor := color.RGBA{40, 40, 60, 255}
		if card.selected {
			cardColor = color.RGBA{60, 60, 100, 255}
		}
		if g.energy < card.cost {
			cardColor = color.RGBA{60, 40, 40, 255} // Red tint if can't play
		}

		vector.DrawFilledRect(screen, float32(x), float32(y), CardWidth, CardHeight, cardColor, false)
		vector.StrokeRect(screen, float32(x), float32(y), CardWidth, CardHeight, 2, color.RGBA{255, 215, 0, 255}, false)

		// Card name
		if g.smallFont != nil {
			text.Draw(screen, card.name, g.smallFont, x+10, y+25, color.RGBA{255, 255, 255, 255})
		}

		// Cost circle
		vector.DrawFilledCircle(screen, float32(x+20), float32(y+20), 15, color.RGBA{0, 100, 255, 255}, false)
		if g.smallFont != nil {
			costText := fmt.Sprintf("%d", card.cost)
			text.Draw(screen, costText, g.smallFont, x+15, y+25, color.RGBA{255, 255, 255, 255})
		}

		// Card description
		if g.smallFont != nil {
			desc := card.description
			if len(desc) > 20 {
				desc = desc[:20] + "..."
			}
			text.Draw(screen, desc, g.smallFont, x+10, y+100, color.RGBA{200, 200, 200, 255})
		}

		// Card type icon
		typeIcon := "⚔️"
		if card.cardType == CardDefense {
			typeIcon = "🛡️"
		} else if card.cardType == CardBuff {
			typeIcon = "✨"
		} else if card.cardType == CardDebuff {
			typeIcon = "☠️"
		}
		if g.smallFont != nil {
			text.Draw(screen, typeIcon, g.smallFont, x+CardWidth-30, y+10, color.RGBA{255, 255, 255, 255})
		}

		// Rarity indicator
		rarityColor := color.RGBA{150, 150, 150, 255}
		if card.rarity == Uncommon {
			rarityColor = color.RGBA{50, 200, 50, 255}
		} else if card.rarity == Rare {
			rarityColor = color.RGBA{50, 50, 255, 255}
		} else if card.rarity == Legendary {
			rarityColor = color.RGBA{255, 215, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y+CardHeight-5), CardWidth, 5, rarityColor, false)
	}

	// Tooltip for hovered card
	if g.hoveredCard >= 0 && g.hoveredCard < len(g.hand) && g.showTooltips {
		card := g.hand[g.hoveredCard]
		tooltipX := cardStartX + g.hoveredCard*(CardWidth+20)
		tooltipY := cardY - 120

		vector.DrawFilledRect(screen, float32(tooltipX-20), float32(tooltipY), CardWidth+40, 100, color.RGBA{0, 0, 0, 200}, false)
		vector.StrokeRect(screen, float32(tooltipX-20), float32(tooltipY), CardWidth+40, 100, 2, color.RGBA{255, 215, 0, 255}, false)

		if g.smallFont != nil {
			text.Draw(screen, card.name, g.smallFont, tooltipX, tooltipY+20, color.RGBA{255, 215, 0, 255})
			text.Draw(screen, card.description, g.smallFont, tooltipX, tooltipY+45, color.RGBA{255, 255, 255, 255})

			typeName := "Атака"
			if card.cardType == CardDefense {
				typeName = "Защита"
			} else if card.cardType == CardBuff {
				typeName = "Бафф"
			} else if card.cardType == CardDebuff {
				typeName = "Дебафф"
			}
			rarityName := "Обычная"
			if card.rarity == Uncommon {
				rarityName = "Необычная"
			} else if card.rarity == Rare {
				rarityName = "Редкая"
			} else if card.rarity == Legendary {
				rarityName = "Легендарная"
			}
			text.Draw(screen, fmt.Sprintf("%s | %s", typeName, rarityName), g.smallFont, tooltipX, tooltipY+70, color.RGBA{150, 150, 150, 255})
		}
	}
}

func (g *GameState) drawStatsPanel(screen *ebiten.Image) {
	// Panel background
	vector.DrawFilledRect(screen, 20, 20, 320, 200, color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeRect(screen, 20, 20, 320, 200, 2, color.RGBA{100, 100, 100, 255}, false)

	if g.gameFont != nil {
		// Health
		hpText := fmt.Sprintf("❤️ %d/%d", g.player.health, g.player.maxHealth)
		text.Draw(screen, hpText, g.gameFont, 35, 50, color.RGBA{255, 100, 100, 255})

		// Block
		if g.player.block > 0 {
			blockText := fmt.Sprintf("🛡️ %d", g.player.block)
			text.Draw(screen, blockText, g.gameFont, 35, 80, color.RGBA{100, 150, 255, 255})
		}

		// Gold
		goldText := fmt.Sprintf("💰 %d", g.gold)
		text.Draw(screen, goldText, g.gameFont, 35, 110, color.RGBA{255, 215, 0, 255})

		// Floor
		floorText := fmt.Sprintf("🏰 Этаж %d", g._floor)
		text.Draw(screen, floorText, g.gameFont, 180, 50, color.RGBA{200, 150, 100, 255})

		// Score
		scoreText := fmt.Sprintf("⭐ %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, 180, 80, color.RGBA{150, 100, 255, 255})
		
		// Relics count
		if len(g.player.relics) > 0 {
			relicText := fmt.Sprintf("🎁 Реликвий: %d", len(g.player.relics))
			text.Draw(screen, relicText, g.smallFont, 35, 145, color.RGBA{255, 150, 50, 255})
		}
		
		// Active effects
		y := 170
		if g.player.rage {
			text.Draw(screen, "⚔️ Ярость", g.smallFont, 35, y, color.RGBA{255, 50, 50, 255})
			y += 20
		}
		if g.player.growingBlock > 0 {
			text.Draw(screen, "🛡️ Растущий блок", g.smallFont, 35, y, color.RGBA{50, 150, 255, 255})
		}
	}
}

func (g *GameState) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), p.size, p.color, false)
	}
}

func (g *GameState) drawDamageNumbers(screen *ebiten.Image) {
	for _, d := range g.damageNumbers {
		c := color.RGBA{255, 255, 255, 255}
		if d.isHeal {
			c = color.RGBA{0, 255, 100, 255}
		}
		if g.smallFont != nil {
			numText := fmt.Sprintf("%d", d.value)
			text.Draw(screen, numText, g.smallFont, int(d.x), int(d.y), c)
		}
	}
}

func (g *GameState) drawSettings(screen *ebiten.Image) {
	// Background
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
		vector.DrawFilledRect(screen, float32(ScreenWidth/2+100), float32(230), 60, 30, toggleColor, false)
		status := "ВКЛ"
		if !g.soundEnabled {
			status = "ВЫКЛ"
		}
		text.Draw(screen, status, g.smallFont, ScreenWidth/2+115, 250, color.RGBA{255, 255, 255, 255})

		// Tooltips toggle
		tooltipText := "💡 Подсказки:"
		text.Draw(screen, tooltipText, g.gameFont, ScreenWidth/2-150, 310, color.RGBA{255, 255, 255, 255})
		toggleColor = color.RGBA{0, 200, 0, 255}
		if !g.showTooltips {
			toggleColor = color.RGBA{200, 0, 0, 255}
		}
		vector.DrawFilledRect(screen, float32(ScreenWidth/2+100), float32(300), 60, 30, toggleColor, false)
		status = "ВКЛ"
		if !g.showTooltips {
			status = "ВЫКЛ"
		}
		text.Draw(screen, status, g.smallFont, ScreenWidth/2+115, 320, color.RGBA{255, 255, 255, 255})

		// Volume slider
		volText := "🔈 Громкость SFX:"
		text.Draw(screen, volText, g.gameFont, ScreenWidth/2-150, 380, color.RGBA{255, 255, 255, 255})
		vector.StrokeLine(screen, float32(ScreenWidth/2-100), float32(400), float32(ScreenWidth/2+100), float32(400), 4, color.RGBA{100, 100, 100, 255}, false)
		sliderX := float32(ScreenWidth/2-100) + float32(g.sfxVolume)*200
		vector.DrawFilledCircle(screen, sliderX, 400, 10, color.RGBA{0, 200, 255, 255}, false)
	}

	// Back button
	vector.DrawFilledRect(screen, 20, 20, 100, 40, color.RGBA{100, 50, 50, 255}, false)
	if g.smallFont != nil {
		text.Draw(screen, "← Назад", g.smallFont, 35, 45, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255})

	if g.gameFont != nil {
		title := "💀 ПОРАЖЕНИЕ"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-140, ScreenHeight/2-80, color.RGBA{255, 50, 50, 255})

		scoreText := fmt.Sprintf("Счёт: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, ScreenWidth/2-60, ScreenHeight/2-20, color.RGBA{255, 255, 255, 255})

		floorText := fmt.Sprintf("Этаж: %d", g._floor)
		text.Draw(screen, floorText, g.gameFont, ScreenWidth/2-50, ScreenHeight/2+20, color.RGBA{255, 255, 255, 255})
	}

	// Continue button
	btnX := ScreenWidth/2 - 150
	btnY := ScreenHeight/2 + 50
	vector.DrawFilledRect(screen, float32(btnX), float32(btnY), 300, 60, color.RGBA{100, 50, 50, 255}, false)
	vector.StrokeRect(screen, float32(btnX), float32(btnY), 300, 60, 3, color.RGBA{255, 100, 100, 255}, false)
	if g.gameFont != nil {
		text.Draw(screen, "Продолжить", g.gameFont, btnX+80, btnY+30, color.RGBA{255, 255, 255, 255})
	}
}

func (g *GameState) drawVictory(screen *ebiten.Image) {
	// Victory gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(50 + y/30)
		g := uint8(80 + y/40)
		b := uint8(50 + y/30)
		screen.Fill(color.RGBA{r, g, b, 255})
	}

	if g.gameFont != nil {
		title := "🏆 ПОБЕДА!"
		text.Draw(screen, title, g.gameFont, ScreenWidth/2-120, ScreenHeight/2-150, color.RGBA{255, 215, 0, 255})

		scoreText := fmt.Sprintf("Счёт: %d", g.score)
		text.Draw(screen, scoreText, g.gameFont, ScreenWidth/2-60, ScreenHeight/2-90, color.RGBA{255, 255, 255, 255})

		floorText := fmt.Sprintf("Этаж: %d", g._floor)
		text.Draw(screen, floorText, g.gameFont, ScreenWidth/2-50, ScreenHeight/2-50, color.RGBA{255, 255, 255, 255})
		
		// Show reward
		rewardText := "✨ Награда получена! ✨"
		text.Draw(screen, rewardText, g.smallFont, ScreenWidth/2-120, ScreenHeight/2-10, color.RGBA{100, 255, 100, 255})
		
		nextText := "Следующий этаж разблокирован!"
		text.Draw(screen, nextText, g.smallFont, ScreenWidth/2-140, ScreenHeight/2+20, color.RGBA{100, 255, 100, 255})
		
		// Show relics
		if len(g.player.relics) > 0 {
			relic := g.player.relics[len(g.player.relics)-1]
			relicName := fmt.Sprintf("🎁 %s", relic.name)
			text.Draw(screen, relicName, g.smallFont, ScreenWidth/2-100, ScreenHeight/2+55, color.RGBA{255, 150, 50, 255})
		}
	}

	// Continue button
	btnX := ScreenWidth/2 - 150
	btnY := ScreenHeight/2 + 100
	vector.DrawFilledRect(screen, float32(btnX), float32(btnY), 300, 60, color.RGBA{50, 100, 50, 255}, false)
	vector.StrokeRect(screen, float32(btnX), float32(btnY), 300, 60, 3, color.RGBA{100, 255, 100, 255}, false)
	if g.gameFont != nil {
		text.Draw(screen, "Продолжить", g.gameFont, btnX+80, btnY+30, color.RGBA{255, 255, 255, 255})
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
	ebiten.SetWindowTitle("⚔️ Go Legends - Roguelike Card Battler | Go365 Day 84")

	game := NewGameState()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

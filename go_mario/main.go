// Go365 Day 86 - GO MARIO: CARD BATTLES v3.0.0
// Карточный Roguelike (Slay the Spire-style)
// Колода, карты, ходы, враги, реликвии

package main

import (
	"fmt"
	"image/color"
	"log"
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

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	ScreenWidth  = 1024
	ScreenHeight = 768

	// Card dimensions
	CardWidth  = 140
	CardHeight = 200

	// Game
	MaxHandSize  = 5
	MaxDeckSize  = 30
	MaxEnergy    = 3
)

// ============================================================================
// COLORS
// ============================================================================

var (
	ColorBG            = color.RGBA{20, 25, 35, 255}
	ColorCardBG        = color.RGBA{40, 50, 70, 255}
	ColorCardBorder    = color.RGBA{100, 120, 150, 255}
	ColorEnergy        = color.RGBA{0, 200, 255, 255}
	ColorHealth        = color.RGBA{220, 50, 50, 255}
	ColorBlock         = color.RGBA{100, 150, 255, 255}
	ColorGold          = color.RGBA{255, 215, 0, 255}
	ColorCardAttack    = color.RGBA{200, 80, 80, 255}
	ColorCardSkill     = color.RGBA{80, 120, 200, 255}
	ColorCardPower     = color.RGBA{180, 80, 200, 255}
	ColorCardCommon    = color.RGBA{150, 150, 150, 255}
	ColorCardUncommon  = color.RGBA{100, 200, 100, 255}
	ColorCardRare      = color.RGBA{255, 200, 50, 255}
	ColorEnemyBG       = color.RGBA{80, 40, 40, 255}
	ColorButtonNormal  = color.RGBA{60, 80, 100, 255}
	ColorButtonHover   = color.RGBA{80, 120, 160, 255}
	ColorButtonActive  = color.RGBA{100, 180, 255, 255}
)

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	playerStand  *ebiten.Image
	slimeGreen   *ebiten.Image
	slimeBlue    *ebiten.Image
	frog         *ebiten.Image
	bee          *ebiten.Image
	gameFont     font.Face
	largeFont    font.Face
}

var gameAssets *Assets

func LoadAssets() *Assets {
	assets := &Assets{}

	assets.playerStand, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	assets.slimeGreen, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeGreen.png")
	assets.slimeBlue, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeBlue.png")
	assets.frog, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/frog.png")
	assets.bee, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/bee.png")

	assets.gameFont, _ = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 18)
	assets.largeFont, _ = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 48)

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
// CARD TYPES
// ============================================================================

type CardType int

const (
	CardAttack CardType = iota
	CardSkill
	CardPower
)

type CardRarity int

const (
	Common CardRarity = iota
	Uncommon
	Rare
)

type Card struct {
	id          int
	name        string
	description string
	cardType    CardType
	rarity      CardRarity
	cost        int
	damage      int
	block       int
	magic       int
	isSelected  bool
	isHovered   bool
	x, y        float64
}

// ============================================================================
// CARD DATABASE
// ============================================================================

var cardDatabase = []Card{
	// Attack cards
	{0, "Удар", "Нанесите 6 урона", CardAttack, Common, 1, 6, 0, 0, false, false, 0, 0},
	{1, "Сильный удар", "Нанесите 10 урона", CardAttack, Common, 2, 10, 0, 0, false, false, 0, 0},
	{2, "Критический удар", "Нанесите 8 урона. Если у врага меньше 10 HP, убивает", CardAttack, Uncommon, 1, 8, 0, 0, false, false, 0, 0},
	{3, "Огненный шар", "Нанесите 12 урона всем врагам", CardAttack, Rare, 2, 12, 0, 0, false, false, 0, 0},
	{4, "Смертельный бросок", "Нанесите 15 урона. Получите 2 урона", CardAttack, Common, 1, 15, 0, 0, false, false, 0, 0},
	
	// Skill cards
	{5, "Защита", "Получите 5 блока", CardSkill, Common, 1, 0, 5, 0, false, false, 0, 0},
	{6, "Сильная защита", "Получите 12 блока", CardSkill, Common, 2, 0, 12, 0, false, false, 0, 0},
	{7, "Уклонение", "Получите 8 блока. Возьмите 1 карту", CardSkill, Uncommon, 1, 0, 8, 0, false, false, 0, 0},
	{8, "Медитация", "Получите 2 энергии. В конце хода потеряйте 1 HP", CardSkill, Rare, 0, 0, 0, 0, false, false, 0, 0},
	{9, "Исцеление", "Восстановите 8 HP", CardSkill, Uncommon, 1, 0, 0, 0, false, false, 0, 0},
	
	// Power cards
	{10, "Ярость", "В начале хода получите 1 дополнительную энергию", CardPower, Rare, 1, 0, 0, 0, false, false, 0, 0},
	{11, "Броня", "В начале хода получите 3 блока", CardPower, Uncommon, 1, 0, 0, 0, false, false, 0, 0},
	{12, "Сила", "Все ваши атаки наносят на 2 урона больше", CardPower, Rare, 2, 0, 0, 0, false, false, 0, 0},
	{13, "Регенерация", "В конце хода восстановите 2 HP", CardPower, Uncommon, 1, 0, 0, 0, false, false, 0, 0},
	{14, "Концентрация", "В начале хода возьмите 1 дополнительную карту", CardPower, Rare, 1, 0, 0, 0, false, false, 0, 0},
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
	strength     int
	extraEnergy  int
	extraBlock   int
	extraDraw    int
	regen        int
	deck         []*Card
	hand         []*Card
	discard      []*Card
	exhaust      []*Card
	relics       []string
}

type Enemy struct {
	id           int
	name         string
	maxHealth    int
	health       int
	block        int
	damage       int
	intent       string
	intentValue  int
	sprite       *ebiten.Image
	x, y         float64
	width, height float32
	isDead       bool
}

type GameState int

const (
	StateMenu GameState = iota
	StateBattle
	StateMap
	StateReward
	StateGameOver
	StateVictory
)

type BattleState int

const (
	BattlePlayerTurn BattleState = iota
	BattleEnemyTurn
	BattleEnd
)

type Game struct {
	player      *Player
	enemies     []*Enemy
	deck        []*Card
	hand        []*Card
	discard     []*Card
	
	gameState   GameState
	battleState BattleState
	turn        int
	room        int
	floor       int
	
	battleReward *Card
	mapRooms     []Room
	
	mouseX, mouseY int
	frameCount     int
}

type Room struct {
	id        int
	roomType  int // 0: battle, 1: elite, 2: treasure, 3: rest
	isCleared bool
	x, y      int
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	gameAssets = LoadAssets()

	g := &Game{
		player: &Player{
			maxHealth: 72,
			health:    72,
			maxEnergy: 3,
			energy:    3,
		},
		gameState:   StateMenu,
		battleState: BattlePlayerTurn,
		turn:        1,
		room:        0,
		floor:       1,
		hand:        make([]*Card, 0),
		discard:     make([]*Card, 0),
		deck:        make([]*Card, 0),
	}

	// Starting deck
	g.addStartingDeck()
	g.generateMap()

	return g
}

func (g *Game) addStartingDeck() {
	// 4x Удар, 4x Защита
	for i := 0; i < 4; i++ {
		g.deck = append(g.deck, &Card{id: 0, name: "Удар", description: "Нанесите 6 урона", cardType: CardAttack, rarity: Common, cost: 1, damage: 6})
	}
	for i := 0; i < 4; i++ {
		g.deck = append(g.deck, &Card{id: 5, name: "Защита", description: "Получите 5 блока", cardType: CardSkill, rarity: Common, cost: 1, block: 5})
	}
}

func (g *Game) generateMap() {
	g.mapRooms = make([]Room, 15)
	for i := range g.mapRooms {
		roomType := rand.Intn(100)
		if roomType < 60 {
			g.mapRooms[i] = Room{id: i, roomType: 0, x: (i % 5) * 200 + 100, y: (i / 5) * 150 + 100}
		} else if roomType < 80 {
			g.mapRooms[i] = Room{id: i, roomType: 1, x: (i % 5) * 200 + 100, y: (i / 5) * 150 + 100}
		} else if roomType < 95 {
			g.mapRooms[i] = Room{id: i, roomType: 2, x: (i % 5) * 200 + 100, y: (i / 5) * 150 + 100}
		} else {
			g.mapRooms[i] = Room{id: i, roomType: 3, x: (i % 5) * 200 + 100, y: (i / 5) * 150 + 100}
		}
	}
}

func (g *Game) startBattle(room Room) {
	g.gameState = StateBattle
	g.battleState = BattlePlayerTurn
	g.player.energy = g.player.maxEnergy
	g.player.block = 0
	
	// Create enemy based on room type
	var enemy *Enemy
	if room.roomType == 1 { // Elite
		enemy = g.createEliteEnemy()
	} else {
		enemy = g.createEnemy(g.floor)
	}
	
	g.enemies = []*Enemy{enemy}
	
	// Draw hand
	g.drawCards(MaxHandSize)
	g.turn = 1
}

func (g *Game) createEnemy(floor int) *Enemy {
	enemyTypes := []struct {
		name   string
		hp     int
		damage int
		sprite *ebiten.Image
	}{
		{"Слизень", 40 + floor*5, 8 + floor*2, gameAssets.slimeGreen},
		{"Летун", 30 + floor*4, 10 + floor*2, gameAssets.bee},
		{"Лягушка", 50 + floor*6, 7 + floor*2, gameAssets.frog},
	}
	
	t := enemyTypes[rand.Intn(len(enemyTypes))]
	return &Enemy{
		name:   t.name,
		maxHealth: t.hp,
		health: t.hp,
		damage: t.damage,
		sprite: t.sprite,
		x: ScreenWidth/2 + 200,
		y: ScreenHeight/2 - 50,
		width: 80,
		height: 80,
	}
}

func (g *Game) createEliteEnemy() *Enemy {
	return &Enemy{
		name:   "ЭЛИТНЫЙ ВРАГ",
		maxHealth: 80 + g.floor*10,
		health: 80 + g.floor*10,
		damage: 15 + g.floor*3,
		sprite: gameAssets.frog,
		x: ScreenWidth/2 + 200,
		y: ScreenHeight/2 - 50,
		width: 100,
		height: 100,
	}
}

func (g *Game) drawCards(count int) {
	for i := 0; i < count && len(g.deck) > 0; i++ {
		if len(g.deck) == 0 {
			if len(g.discard) == 0 {
				break
			}
			// Shuffle discard into deck
			g.deck = g.discard
			g.discard = make([]*Card, 0)
			rand.Shuffle(len(g.deck), func(i, j int) {
				g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
			})
		}
		
		if len(g.deck) > 0 {
			card := g.deck[len(g.deck)-1]
			g.deck = g.deck[:len(g.deck)-1]
			g.hand = append(g.hand, card)
		}
	}
}

func (g *Game) playCard(card *Card, target *Enemy) {
	p := g.player
	
	if p.energy < card.cost {
		return
	}
	
	p.energy -= card.cost
	
	switch card.cardType {
	case CardAttack:
		damage := card.damage + p.strength
		if target != nil {
			target.health -= damage
			target.block -= damage
			if target.block < 0 {
				target.health += target.block
				target.block = 0
			}
			if target.health <= 0 {
				target.isDead = true
				g.endBattle(true)
			}
		}
	case CardSkill:
		p.block += card.block
		if card.id == 7 { // Уклонение
			g.drawCards(1)
		}
		if card.id == 8 { // Медитация
			p.energy += 2
		}
		if card.id == 9 { // Исцеление
			p.health = min(p.health+8, p.maxHealth)
		}
	case CardPower:
		switch card.id {
		case 10: p.extraEnergy = 999 // Ярость
		case 11: p.extraBlock = 999  // Броня
		case 12: p.strength = 999    // Сила
		case 13: p.regen = 999       // Регенерация
		case 14: p.extraDraw = 999   // Концентрация
		}
	}
	
	// Move card to discard
	for i, c := range g.hand {
		if c == card {
			g.hand = append(g.hand[:i], g.hand[i+1:]...)
			g.discard = append(g.discard, card)
			break
		}
	}
}

func (g *Game) endPlayerTurn() {
	g.battleState = BattleEnemyTurn
	
	// Enemy turn
	for _, enemy := range g.enemies {
		if enemy.isDead {
			continue
		}
		
		// Simple AI
		enemy.intent = "attack"
		enemy.intentValue = enemy.damage
		
		// Deal damage
		damage := enemy.damage
		block := g.player.block
		if block >= damage {
			g.player.block -= damage
		} else {
			g.player.health -= (damage - block)
			g.player.block = 0
		}
		
		if g.player.health <= 0 {
			g.gameState = StateGameOver
		}
	}
	
	// End of turn cleanup
	g.endTurn()
}

func (g *Game) endTurn() {
	p := g.player
	
	// Regen
	if p.regen > 0 {
		p.health = min(p.health+2, p.maxHealth)
	}
	
	// Clear block
	p.block = 0
	
	// Discard hand
	for _, card := range g.hand {
		g.discard = append(g.discard, card)
	}
	g.hand = make([]*Card, 0)
	
	// Start new turn
	p.energy = p.maxEnergy
	if p.extraEnergy > 0 {
		p.energy++
	}
	
	drawCount := MaxHandSize
	if p.extraDraw > 0 {
		drawCount++
	}
	g.drawCards(drawCount)
	
	g.turn++
	g.battleState = BattlePlayerTurn
	
	// Set enemy intent
	for _, enemy := range g.enemies {
		if !enemy.isDead {
			enemy.intent = "attack"
			enemy.intentValue = enemy.damage
		}
	}
}

func (g *Game) endBattle(victory bool) {
	if victory {
		g.gameState = StateReward
		g.battleReward = g.getRandomCard()
		g.room++
		if g.room >= len(g.mapRooms) {
			g.gameState = StateVictory
		}
	}
}

func (g *Game) getRandomCard() *Card {
	rarityRoll := rand.Intn(100)
	var rarity CardRarity
	if rarityRoll < 60 {
		rarity = Common
	} else if rarityRoll < 90 {
		rarity = Uncommon
	} else {
		rarity = Rare
	}
	
	availableCards := make([]Card, 0)
	for _, card := range cardDatabase {
		if card.rarity == rarity {
			availableCards = append(availableCards, card)
		}
	}
	
	if len(availableCards) == 0 {
		availableCards = cardDatabase
	}
	
	card := availableCards[rand.Intn(len(availableCards))]
	return &card
}

func (g *Game) addCardToDeck(card *Card) {
	if len(g.deck) < MaxDeckSize {
		g.deck = append(g.deck, card)
	}
	g.gameState = StateMap
}

func (g *Game) rest() {
	g.player.health = min(g.player.health+15, g.player.maxHealth)
	g.room++
	g.gameState = StateMap
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frameCount++
	
	// Track mouse
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	
	switch g.gameState {
	case StateMenu:
		g.updateMenu()
	case StateMap:
		g.updateMap()
	case StateBattle:
		g.updateBattle()
	case StateReward:
		g.updateReward()
	case StateGameOver, StateVictory:
		g.updateEnd()
	}
	
	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.gameState = StateMap
	}
}

func (g *Game) updateMap() {
	// Check room click
	for i := range g.mapRooms {
		room := &g.mapRooms[i]
		if i == g.room && !room.isCleared {
			if g.isMouseOver(room.x, room.y, 80, 80) {
				if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
					if room.roomType == 3 {
						g.rest()
					} else {
						g.startBattle(*room)
					}
				}
			}
		}
	}
}

func (g *Game) updateBattle() {
	// Update card hover
	for _, card := range g.hand {
		card.isHovered = g.isMouseOver(int(card.x), int(card.y), CardWidth, CardHeight)
		if card.isHovered && inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Play card on first enemy
			for _, enemy := range g.enemies {
				if !enemy.isDead {
					g.playCard(card, enemy)
					break
				}
			}
		}
	}
	
	// End turn button
	if g.isMouseOver(ScreenWidth-150, ScreenHeight-80, 130, 50) {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.endPlayerTurn()
		}
	}
}

func (g *Game) updateReward() {
	// Select card
	if g.battleReward != nil {
		g.battleReward.isHovered = g.isMouseOver(ScreenWidth/2-CardWidth/2, ScreenHeight/2-CardHeight/2, CardWidth, CardHeight)
		if g.battleReward.isHovered && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
			g.addCardToDeck(g.battleReward)
		}
	}
	
	// Skip button
	if g.isMouseOver(ScreenWidth/2-65, ScreenHeight/2+150, 130, 40) {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.gameState = StateMap
		}
	}
}

func (g *Game) updateEnd() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		*g = *NewGame()
	}
}

func (g *Game) isMouseOver(x, y, w, h int) bool {
	return g.mouseX >= x && g.mouseX <= x+w && g.mouseY >= y && g.mouseY <= y+h
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameState {
	case StateMenu:
		g.drawMenu(screen)
	case StateMap:
		g.drawMap(screen)
	case StateBattle:
		g.drawBattle(screen)
	case StateReward:
		g.drawReward(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateVictory:
		g.drawVictory(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(ColorBG)
	
	if gameAssets.largeFont != nil {
		title := "🍄 GO MARIO: CARD BATTLES 🍄"
		bounds := text.BoundString(gameAssets.largeFont, title)
		text.Draw(screen, title, gameAssets.largeFont, ScreenWidth/2-bounds.Dx()/2, 200, ColorGold)
		
		subtitle := "Карточный Roguelike"
		bounds = text.BoundString(gameAssets.gameFont, subtitle)
		text.Draw(screen, subtitle, gameAssets.gameFont, ScreenWidth/2-bounds.Dx()/2, 280, color.White)
		
		instructions := []string{
			"🎴 Собирайте карты",
			"⚔️ Сражайтесь с врагами",
			"📈 Улучшайте колоду",
			"🏆 Достигните вершины",
			"",
			"Нажмите ENTER для старта",
		}
		
		y := 380
		for _, line := range instructions {
			bounds = text.BoundString(gameAssets.gameFont, line)
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-bounds.Dx()/2, y, color.White)
			y += 35
		}
	}
}

func (g *Game) drawMap(screen *ebiten.Image) {
	screen.Fill(ColorBG)
	
	// Draw map rooms
	for i, room := range g.mapRooms {
		roomColor := ColorCardBorder
		if room.isCleared {
			roomColor = color.RGBA{50, 50, 50, 255}
		}
		if i == g.room {
			roomColor = ColorGold
		}
		
		vector.DrawFilledRect(screen, float32(room.x), float32(room.y), 80, 80, roomColor, true)
		
		// Room icon
		icon := "?"
		switch room.roomType {
		case 0: icon = "⚔️"
		case 1: icon = "💀"
		case 2: icon = "📦"
		case 3: icon = "🔥"
		}
		
		if gameAssets.gameFont != nil {
			text.Draw(screen, icon, gameAssets.gameFont, room.x+30, room.y+50, color.White)
		}
		
		// Connection line
		if i < len(g.mapRooms)-1 {
			nextRoom := g.mapRooms[i+1]
			vector.StrokeLine(screen, float32(room.x+40), float32(room.y+80),
				float32(nextRoom.x+40), float32(nextRoom.y), 2, ColorCardBorder, true)
		}
	}
	
	// Player info
	g.drawPlayerInfo(screen, 20, 20)
	
	// Instructions
	if gameAssets.gameFont != nil {
		text.Draw(screen, "Нажмите ENTER для входа в комнату", gameAssets.gameFont, ScreenWidth/2-150, ScreenHeight-50, color.White)
	}
}

func (g *Game) drawBattle(screen *ebiten.Image) {
	screen.Fill(ColorBG)
	
	// Draw enemies
	for _, enemy := range g.enemies {
		if enemy.isDead {
			continue
		}

		// Enemy sprite background
		vector.DrawFilledRect(screen, float32(enemy.x-40), float32(enemy.y-40), enemy.width, enemy.height, ColorEnemyBG, true)

		if enemy.sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(enemy.x-40), float64(enemy.y-40))
			screen.DrawImage(enemy.sprite, op)
		}

		// Health bar
		vector.DrawFilledRect(screen, float32(enemy.x-40), float32(enemy.y-60), enemy.width, 10, color.RGBA{80, 0, 0, 255}, true)
		healthPercent := float32(enemy.health) / float32(enemy.maxHealth)
		vector.DrawFilledRect(screen, float32(enemy.x-40), float32(enemy.y-60), enemy.width*healthPercent, 10, ColorHealth, true)

		// Name and HP
		if gameAssets.gameFont != nil {
			text.Draw(screen, fmt.Sprintf("%s", enemy.name), gameAssets.gameFont, int(enemy.x-30), int(enemy.y-70), color.White)
			text.Draw(screen, fmt.Sprintf("%d/%d", enemy.health, enemy.maxHealth), gameAssets.gameFont, int(enemy.x+20), int(enemy.y-70), ColorHealth)

			// Intent
			intentIcon := "⚔️"
			text.Draw(screen, fmt.Sprintf("%s %d", intentIcon, enemy.intentValue), gameAssets.gameFont, int(enemy.x-30), int(enemy.y+50), color.RGBA{255, 100, 100, 255})
		}
	}

	// Draw player
	if gameAssets.playerStand != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(200, ScreenHeight/2-50)
		op.GeoM.Scale(0.5, 0.5)
		screen.DrawImage(gameAssets.playerStand, op)
	}

	// Player stats
	g.drawPlayerInfo(screen, 20, 20)
	
	// Energy
	if gameAssets.gameFont != nil {
		for i := 0; i < g.player.energy; i++ {
			vector.DrawFilledCircle(screen, float32(250+i*30), 100, 12, ColorEnergy, true)
		}
	}
	
	// Cards in hand
	cardStartX := ScreenWidth/2 - (len(g.hand)*CardWidth)/2 - (len(g.hand)-1)*20
	for i, card := range g.hand {
		card.x = float64(cardStartX + i*(CardWidth+40))
		card.y = float64(ScreenHeight - CardHeight - 20)
		
		if card.isHovered {
			card.y -= 30
		}
		
		g.drawCard(screen, card)
	}
	
	// End turn button
	buttonColor := ColorButtonNormal
	if g.isMouseOver(ScreenWidth-150, ScreenHeight-80, 130, 50) {
		buttonColor = ColorButtonHover
	}
	vector.DrawFilledRect(screen, float32(ScreenWidth-150), float32(ScreenHeight-80), 130, 50, buttonColor, true)
	vector.StrokeRect(screen, float32(ScreenWidth-150), float32(ScreenHeight-80), 130, 50, 2, ColorCardBorder, true)
	if gameAssets.gameFont != nil {
		text.Draw(screen, "Конец хода", gameAssets.gameFont, ScreenWidth-130, ScreenHeight-50, color.White)
	}
	
	// Turn info
	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("Ход %d", g.turn), gameAssets.gameFont, ScreenWidth/2-30, 20, color.White)
	}
}

func (g *Game) drawCard(screen *ebiten.Image, card *Card) {
	x, y := card.x, card.y
	
	// Card background
	cardColor := ColorCardBG
	if card.isHovered {
		cardColor = ColorCardBorder
	}
	
	vector.DrawFilledRect(screen, float32(x), float32(y), CardWidth, CardHeight, cardColor, true)
	vector.StrokeRect(screen, float32(x), float32(y), CardWidth, CardHeight, 2, ColorCardBorder, true)

	// Card type color
	typeColor := ColorCardAttack
	if card.cardType == CardSkill {
		typeColor = ColorCardSkill
	} else if card.cardType == CardPower {
		typeColor = ColorCardPower
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), CardWidth, 5, typeColor, true)

	// Cost
	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("%d", card.cost), gameAssets.gameFont, int(x)+8, int(y)+25, ColorEnergy)
	}

	// Card name
	if gameAssets.gameFont != nil {
		bounds := text.BoundString(gameAssets.gameFont, card.name)
		text.Draw(screen, card.name, gameAssets.gameFont, int(x)+CardWidth/2-bounds.Dx()/2, int(y)+50, color.White)

		// Description
		descLines := wrapText(card.description, 18)
		for i, line := range descLines {
			text.Draw(screen, line, gameAssets.gameFont, int(x)+10, int(y)+80+i*20, color.RGBA{200, 200, 200, 255})
		}

		// Stats
		stats := ""
		if card.damage > 0 {
			stats += fmt.Sprintf("⚔️%d ", card.damage)
		}
		if card.block > 0 {
			stats += fmt.Sprintf("🛡️%d ", card.block)
		}
		text.Draw(screen, stats, gameAssets.gameFont, int(x)+10, int(y)+CardHeight-30, color.White)
	}
}

func wrapText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}
	
	words := []string{text}
	lines := []string{}
	currentLine := ""
	
	for _, word := range words {
		if len(currentLine)+len(word) > maxLen {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine += word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}

func (g *Game) drawPlayerInfo(screen *ebiten.Image, x, y int) {
	p := g.player
	
	// Health
	vector.DrawFilledRect(screen, float32(x), float32(y), 150, 20, color.RGBA{80, 0, 0, 255}, true)
	healthPercent := float32(p.health) / float32(p.maxHealth)
	vector.DrawFilledRect(screen, float32(x), float32(y), 150*healthPercent, 20, ColorHealth, true)
	
	// Block
	vector.DrawFilledRect(screen, float32(x), float32(y+25), 80, 16, color.RGBA{0, 0, 80, 255}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y+25), float32(min(p.block, 80)), 16, ColorBlock, true)
	
	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("HP %d/%d", p.health, p.maxHealth), gameAssets.gameFont, x+10, y+15, color.White)
		text.Draw(screen, fmt.Sprintf("🛡️%d", p.block), gameAssets.gameFont, x+10, y+38, ColorBlock)
	}
}

func (g *Game) drawReward(screen *ebiten.Image) {
	screen.Fill(ColorBG)
	
	if gameAssets.largeFont != nil {
		text.Draw(screen, "🎉 НАГРАДА!", gameAssets.largeFont, ScreenWidth/2-100, 150, ColorGold)
	}
	
	if gameAssets.gameFont != nil {
		text.Draw(screen, "Выберите карту для добавления в колоду", gameAssets.gameFont, ScreenWidth/2-180, 220, color.White)
	}
	
	// Draw reward card
	if g.battleReward != nil {
		g.battleReward.x = float64(ScreenWidth/2 - CardWidth/2)
		g.battleReward.y = float64(ScreenHeight/2 - CardHeight/2)
		g.drawCard(screen, g.battleReward)
	}
	
	// Skip button
	buttonColor := ColorButtonNormal
	if g.isMouseOver(ScreenWidth/2-65, ScreenHeight/2+150, 130, 40) {
		buttonColor = ColorButtonHover
	}
	vector.DrawFilledRect(screen, float32(ScreenWidth/2-65), float32(ScreenHeight/2+150), 130, 40, buttonColor, true)
	if gameAssets.gameFont != nil {
		text.Draw(screen, "Пропустить", gameAssets.gameFont, ScreenWidth/2-50, ScreenHeight/2+175, color.White)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 20, 20, 255})
	
	if gameAssets.largeFont != nil {
		text.Draw(screen, "💀 ВЫ ПРОИГРАЛИ", gameAssets.largeFont, ScreenWidth/2-180, ScreenHeight/2-50, ColorHealth)
		text.Draw(screen, fmt.Sprintf("Этаж: %d | Комната: %d", g.floor, g.room), gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2+20, color.White)
		text.Draw(screen, "Нажмите ENTER для рестарта", gameAssets.gameFont, ScreenWidth/2-150, ScreenHeight/2+80, color.White)
	}
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 50, 20, 255})
	
	if gameAssets.largeFont != nil {
		text.Draw(screen, "🏆 ПОБЕДА!", gameAssets.largeFont, ScreenWidth/2-120, ScreenHeight/2-50, ColorGold)
		text.Draw(screen, fmt.Sprintf("Ходов: %d", g.turn), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+20, color.White)
		text.Draw(screen, "Нажмите ENTER для рестарта", gameAssets.gameFont, ScreenWidth/2-150, ScreenHeight/2+80, color.White)
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
	ebiten.SetWindowTitle("GO MARIO: CARD BATTLES - Go365 Day 86 | Card Roguelike")
	ebiten.SetVsyncEnabled(true)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

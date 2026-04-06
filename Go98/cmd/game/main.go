package main

import (
	"log"
	"math/rand"
	"time"

	"dungeon_crawler/internal/audio"
	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/dungeon"
	"dungeon_crawler/internal/entity"
	"dungeon_crawler/internal/engine"
	"dungeon_crawler/internal/sprite"
	"dungeon_crawler/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// GameState represents current game state
type GameState string

const (
	StateMenu      GameState = "menu"
	StatePlaying   GameState = "playing"
	StatePaused    GameState = "paused"
	StateGameOver  GameState = "game_over"
	StateVictory   GameState = "victory"
	StateNextFloor GameState = "next_floor"
)

// Game is the main game struct
type Game struct {
	State          GameState
	Player         *entity.Player
	Enemies        []*entity.Enemy
	Items          []*entity.Item
	Dungeon        *dungeon.Dungeon
	Renderer       *engine.Renderer
	Sound          *audio.SoundManager
	SpriteMgr      *sprite.SpriteManager
	DamageNums     []*entity.DamageNumber
	Rng            *rand.Rand
	Floor          int
	Score          int
	MenuSelect     int
	NextFloorTimer int
	ShowControls   bool
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	sm := sprite.NewSpriteManager()
	renderer := engine.NewRenderer()

	sound, err := audio.NewSoundManager()
	if err != nil {
		log.Printf("Warning: Audio not available: %v", err)
	}

	g := &Game{
		State:     StateMenu,
		Rng:       rng,
		Renderer:  renderer,
		Sound:     sound,
		SpriteMgr: sm,
	}

	return g
}

func (g *Game) Update() error {
	switch g.State {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		g.updatePaused()
	case StateGameOver:
		g.updateGameOver()
	case StateVictory:
		g.updateVictory()
	case StateNextFloor:
		g.NextFloorTimer--
		if g.NextFloorTimer <= 0 {
			g.State = StatePlaying
		}
	}
	return nil
}

func (g *Game) updateMenu() {
	// Menu navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.MenuSelect--
		if g.MenuSelect < 0 {
			g.MenuSelect = 2
		}
		if g.Sound != nil {
			g.Sound.PlayMenu()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.MenuSelect++
		if g.MenuSelect > 2 {
			g.MenuSelect = 0
		}
		if g.Sound != nil {
			g.Sound.PlayMenu()
		}
	}

	// Select
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(config.GameKeys.Attack) {
		switch g.MenuSelect {
		case 0: // Start
			g.startGame()
		case 1: // How to play
			g.ShowControls = !g.ShowControls
		case 2: // Exit
			// Can't exit in browser, just ignore
		}
	}
}

func (g *Game) updatePlaying() {
	// Check pause
	if inpututil.IsKeyJustPressed(config.GameKeys.Pause) {
		g.State = StatePaused
		return
	}

	// Check use potion
	if inpututil.IsKeyJustPressed(config.GameKeys.UseItems) {
		if g.Player.Potions > 0 && g.Player.HP < g.Player.MaxHP {
			g.Player.Potions--
			g.Player.Heal(config.HealPotionValue)
			if g.Sound != nil {
				g.Sound.PlayHeal()
			}
		}
	}

	// Update player
	g.Player.Update(g.Dungeon)

	// Check tile effects
	tileX := int((g.Player.X + g.Player.Width/2) / config.TileSize)
	tileY := int((g.Player.Y + g.Player.Height/2) / config.TileSize)
	tile := g.Dungeon.GetTile(tileX, tileY)

	switch tile {
	case config.TileSpikes:
		g.Player.TakeDamage(5)
		g.addDamageNumber(g.Player.X, g.Player.Y, 5, entity.DamageColor{R: 255, G: 0, B: 0})
	case config.TileStairs:
		if g.Floor < config.MaxFloors {
			g.nextFloor()
		} else {
			g.victory()
		}
	}

	// Update enemies
	for _, enemy := range g.Enemies {
		enemy.Update(g.Player, g.Dungeon)

		// Check collision with player
		if enemy.IsAlive() && entity.CollidesWith(g.Player, enemy) {
			g.Player.TakeDamage(enemy.Damage)
			g.addDamageNumber(g.Player.X, g.Player.Y-20, enemy.Damage, entity.DamageColor{R: 255, G: 0, B: 0})
			if g.Sound != nil {
				g.Sound.PlayHit()
			}
		}
	}

	// Player attack
	if g.Player.Attacking && g.Player.AttackFrame == 10 { // Hit frame
		ax, ay, aw, ah := g.Player.GetAttackBox()
		attackBox := &entity.AttackHitBox{X: ax, Y: ay, Width: aw, Height: ah}

		for _, enemy := range g.Enemies {
			if enemy.IsAlive() && entity.CollidesWith(attackBox, enemy) {
				enemy.TakeDamage(g.Player.AttackDMG)
				g.addDamageNumber(enemy.X, enemy.Y-20, g.Player.AttackDMG, entity.DamageColor{R: 255, G: 255, B: 0})
				g.Score += 10
				if g.Sound != nil {
					g.Sound.PlayAttack()
				}

				if !enemy.IsAlive() {
					g.Score += 50 // Kill bonus
				}
			}
		}
	}

	// Update items
	for _, item := range g.Items {
		item.Update()

		// Check pickup
		if item.Active && !item.Collected && entity.CollidesWith(g.Player, item) {
			switch item.Type {
			case entity.ItemCoin:
				g.Player.Coins += item.Value
				g.Score += config.CoinValue
				if g.Sound != nil {
					g.Sound.PlayCoin()
				}
			case entity.ItemGem:
				g.Player.Gems += item.Value
				g.Score += config.GemValue
				if g.Sound != nil {
					g.Sound.PlayGem()
				}
			case entity.ItemKey:
				g.Player.Keys += item.Value
				if g.Sound != nil {
					g.Sound.PlayKey()
				}
			case entity.ItemPotion:
				g.Player.Potions += item.Value
				if g.Sound != nil {
					g.Sound.PlayHeal()
				}
			}
			item.Collect()
		}
	}

	// Update damage numbers
	for _, dmg := range g.DamageNums {
		dmg.Update()
	}

	// Clean up damage numbers
	activeDamages := make([]*entity.DamageNumber, 0)
	for _, dmg := range g.DamageNums {
		if dmg.IsActive() {
			activeDamages = append(activeDamages, dmg)
		}
	}
	g.DamageNums = activeDamages

	// Check death
	if g.Player.IsDead() {
		g.gameOver()
	}

	// Update camera
	g.Renderer.Camera.Follow(g.Player)
}

func (g *Game) updatePaused() {
	if inpututil.IsKeyJustPressed(config.GameKeys.Pause) {
		g.State = StatePlaying
	}
}

func (g *Game) updateGameOver() {
	if inpututil.IsKeyJustPressed(config.GameKeys.Restart) {
		g.startGame()
	}
}

func (g *Game) updateVictory() {
	if inpututil.IsKeyJustPressed(config.GameKeys.Restart) {
		g.startGame()
	}
}

func (g *Game) startGame() {
	g.Floor = 1
	g.Score = 0
	g.generateFloor()
	g.State = StatePlaying
	if g.Sound != nil {
		g.Sound.PlayMenu()
	}
}

func (g *Game) generateFloor() {
	g.Dungeon = dungeon.NewDungeon(g.Floor, g.Rng)

	// Spawn player
	spawnX, spawnY := g.Dungeon.GetSpawnPoint()
	g.Player = entity.NewPlayer(spawnX, spawnY)
	g.Player.Floor = g.Floor
	g.Player.Score = g.Score

	// Generate enemies
	numEnemies := 5 + g.Floor*3
	if numEnemies > 20 {
		numEnemies = 20
	}

	g.Enemies = make([]*entity.Enemy, 0)
	for i := 0; i < numEnemies; i++ {
		roomIdx := 1 + g.Rng.Intn(len(g.Dungeon.Rooms)-1)
		room := g.Dungeon.Rooms[roomIdx]
		ex := room.X + 1 + g.Rng.Intn(room.W-2)
		ey := room.Y + 1 + g.Rng.Intn(room.H-2)

		var enemyType entity.EnemyType
		r := g.Rng.Intn(10)
		if r < 4 {
			enemyType = entity.EnemySlime
		} else if r < 7 {
			enemyType = entity.EnemyBee
		} else {
			enemyType = entity.EnemyFly
		}

		g.Enemies = append(g.Enemies, entity.NewEnemy(ex, ey, enemyType, g.Floor))
	}

	// Generate items
	g.Items = make([]*entity.Item, 0)

	// Coins in each room
	for _, room := range g.Dungeon.Rooms {
		numCoins := g.Rng.Intn(3) + 1
		for i := 0; i < numCoins; i++ {
			ix := room.X + 1 + g.Rng.Intn(room.W-2)
			iy := room.Y + 1 + g.Rng.Intn(room.H-2)
			g.Items = append(g.Items, entity.NewItem(ix, iy, entity.ItemCoin, 1))
		}
	}

	// Gems (rarer)
	numGems := 2 + g.Rng.Intn(3)
	for i := 0; i < numGems; i++ {
		room := g.Dungeon.Rooms[g.Rng.Intn(len(g.Dungeon.Rooms))]
		ix := room.X + 1 + g.Rng.Intn(room.W-2)
		iy := room.Y + 1 + g.Rng.Intn(room.H-2)
		g.Items = append(g.Items, entity.NewItem(ix, iy, entity.ItemGem, 1))
	}

	// Potions
	numPotions := 1 + g.Rng.Intn(3)
	for i := 0; i < numPotions; i++ {
		room := g.Dungeon.Rooms[g.Rng.Intn(len(g.Dungeon.Rooms))]
		ix := room.X + 1 + g.Rng.Intn(room.W-2)
		iy := room.Y + 1 + g.Rng.Intn(room.H-2)
		g.Items = append(g.Items, entity.NewItem(ix, iy, entity.ItemPotion, 1))
	}

	// Keys (rare, needed for some doors in future updates)
	numKeys := g.Rng.Intn(2)
	for i := 0; i < numKeys; i++ {
		room := g.Dungeon.Rooms[g.Rng.Intn(len(g.Dungeon.Rooms))]
		ix := room.X + 1 + g.Rng.Intn(room.W-2)
		iy := room.Y + 1 + g.Rng.Intn(room.H-2)
		g.Items = append(g.Items, entity.NewItem(ix, iy, entity.ItemKey, 1))
	}

	g.DamageNums = make([]*entity.DamageNumber, 0)
}

func (g *Game) nextFloor() {
	g.Floor++
	g.Score += 100 // Floor completion bonus
	g.Player.Floor = g.Floor
	g.generateFloor()
	g.NextFloorTimer = 120
	g.State = StateNextFloor
	if g.Sound != nil {
		g.Sound.PlayStairs()
	}
}

func (g *Game) gameOver() {
	g.State = StateGameOver
	g.Score = g.Player.Score
	if g.Sound != nil {
		g.Sound.PlayDeath()
	}
}

func (g *Game) victory() {
	g.State = StateVictory
	g.Score = g.Player.Score + 1000 // Victory bonus
	if g.Sound != nil {
		g.Sound.PlayVictory()
	}
}

func (g *Game) addDamageNumber(x, y float64, value int, col entity.DamageColor) {
	g.DamageNums = append(g.DamageNums, entity.NewDamageNumber(x, y, value, col))
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateMenu:
		ui.DrawMenu(screen, g.MenuSelect)
	case StatePlaying:
		g.renderGame(screen)
	case StatePaused:
		g.renderGame(screen)
		ui.DrawPause(screen)
	case StateGameOver:
		g.renderGame(screen)
		ui.DrawGameOver(screen, g.Score, g.Floor)
	case StateVictory:
		ui.DrawVictory(screen, g.Score)
	case StateNextFloor:
		g.renderGame(screen)
		ui.DrawNextFloor(screen, g.Floor)
	}
}

func (g *Game) renderGame(screen *ebiten.Image) {
	// Clear with dark background
	screen.Fill(entity.ColorDarkBG)

	// Render dungeon
	if g.Dungeon != nil {
		g.Renderer.RenderDungeon(screen, g.Dungeon)
		g.Renderer.RenderItems(screen, g.Items)
		g.Renderer.RenderEnemies(screen, g.Enemies)
		g.Renderer.RenderPlayer(screen, g.Player)
		g.Renderer.RenderDamageNumbers(screen, g.DamageNums)
	}

	// HUD
	if g.Player != nil {
		ui.DrawHUD(screen, g.Player)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

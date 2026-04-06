package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"towerdefense/internal/audio"
	"towerdefense/internal/config"
	"towerdefense/internal/enemy"
	gamemap "towerdefense/internal/map"
	"towerdefense/internal/particle"
	"towerdefense/internal/projectile"
	"towerdefense/internal/tower"
	"towerdefense/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameState
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
)

// Game main struct
type Game struct {
	state        GameState
	gameMap      *gamemap.Map
	towers       []*tower.Tower
	enemies      []*enemy.Enemy
	projectiles  []*projectile.Projectile
	particles    *particle.System
	audio        *audio.Manager
	
	gold         int
	lives        int
	score        int
	wave         int
	waveTimer    int
	waveActive   bool
	enemiesSpawned int
	enemiesInWave  int
	spawnTimer   int
	
	selectedTowerType int
	selectedTower     *tower.Tower
	hoverGridX        int
	hoverGridY        int
	showTowerMenu     bool
	
	pathPositions    []enemy.PathPosition
	menuSelection    int
	gameOverTimer    int
	
	// Visual effects
	menuStars []Star
	frameCount int
}

type Star struct {
	X, Y   float64
	Size   float64
	Brightness float64
	TwinkleSpeed float64
}

func NewGame() *Game {
	am, _ := audio.NewManager()
	
	g := &Game{
		state:             StateMenu,
		particles:         particle.NewSystem(),
		audio:             am,
		gold:              config.InitialGold,
		lives:             config.InitialLives,
		selectedTowerType: -1,
		hoverGridX:        -1,
		hoverGridY:        -1,
		menuSelection:     0,
		menuStars:         make([]Star, 100),
	}
	
	// Initialize stars
	for i := range g.menuStars {
		g.menuStars[i] = Star{
			X:            rand.Float64() * config.ScreenWidth,
			Y:            rand.Float64() * config.ScreenHeight,
			Size:         1 + rand.Float64()*2,
			Brightness:   0.5 + rand.Float64()*0.5,
			TwinkleSpeed: 0.02 + rand.Float64()*0.03,
		}
	}
	
	g.resetGame()
	return g
}

func (g *Game) resetGame() {
	g.gameMap = gamemap.NewMap()
	g.towers = make([]*tower.Tower, 0)
	g.enemies = make([]*enemy.Enemy, 0)
	g.projectiles = make([]*projectile.Projectile, 0)
	g.gold = config.InitialGold
	g.lives = config.InitialLives
	g.score = 0
	g.wave = 0
	g.waveTimer = 180
	g.waveActive = false
	g.selectedTowerType = -1
	g.selectedTower = nil
	g.showTowerMenu = false
	
	// Precalculate path positions
	pathNodes := g.gameMap.Path
	g.pathPositions = enemy.PrecalculatePathPositions(pathNodes)
}

func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		if inpututil.IsKeyJustPressed(config.KeyPause) {
			g.state = StatePlaying
		}
	case StateGameOver, StateVictory:
		if inpututil.IsKeyJustPressed(config.KeyRestart) {
			g.resetGame()
			g.state = StatePlaying
		}
	}
	
	g.particles.Update()
	return nil
}

func (g *Game) updateMenu() {
	g.frameCount++
	
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.menuSelection--
		if g.menuSelection < 0 {
			g.menuSelection = 2
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.menuSelection++
		if g.menuSelection > 2 {
			g.menuSelection = 0
		}
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		switch g.menuSelection {
		case 0:
			g.resetGame()
			g.state = StatePlaying
		case 1:
			// How to play - skip for now
		case 2:
			// Exit - just close
			return
		}
	}
}

func (g *Game) updatePlaying() {
	g.frameCount++
	
	// Pause
	if inpututil.IsKeyJustPressed(config.KeyPause) {
		g.state = StatePaused
		return
	}
	
	// Mouse handling
	mx, my := ebiten.CursorPosition()
	g.hoverGridX = (mx - config.GridOffsetX) / config.TileSize
	g.hoverGridY = (my - config.GridOffsetY) / config.TileSize
	
	// Tower placement
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.hoverGridX >= 0 && g.hoverGridX < config.GridWidth &&
		   g.hoverGridY >= 0 && g.hoverGridY < config.GridHeight {
			
			// Check if clicking on existing tower
			clickedTower := g.getTowerAt(g.hoverGridX, g.hoverGridY)
			
			if clickedTower != nil {
				g.selectedTower = clickedTower
				g.showTowerMenu = true
			} else if g.selectedTowerType >= 0 && g.gameMap.CanPlaceTower(g.hoverGridX, g.hoverGridY) {
				stats := tower.GetType(g.selectedTowerType)
				if g.gold >= stats.Cost {
					g.gold -= stats.Cost
					newTower := tower.NewTower(g.selectedTowerType, g.hoverGridX, g.hoverGridY)
					g.towers = append(g.towers, newTower)
					g.gameMap.PlaceTower(g.hoverGridX, g.hoverGridY)
					g.audio.PlayPlace()
					g.particles.Emit(newTower.X, newTower.Y, particle.PTLevelUp, 10)
				}
			}
		}
	}
	
	// Right click to deselect
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		g.selectedTowerType = -1
		g.showTowerMenu = false
		g.selectedTower = nil
	}
	
	// Number keys for tower selection
	for i := 0; i < 5; i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(i)) {
			if g.selectedTowerType == i {
				g.selectedTowerType = -1
			} else {
				g.selectedTowerType = i
				g.showTowerMenu = false
				g.selectedTower = nil
			}
		}
	}
	
	// Wave management
	if !g.waveActive && g.wave < config.MaxWaves {
		g.waveTimer--
		if g.waveTimer <= 0 || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startWave()
		}
	}
	
	if g.waveActive {
		g.spawnEnemies()
	}
	
	// Update towers
	for _, t := range g.towers {
		t.Update(g.enemies)
		if t.CanFire() {
			target := t.GetTarget()
			if target != nil {
				proj := projectile.NewProjectile(t, target)
				if proj != nil {
					g.projectiles = append(g.projectiles, proj)
					g.audio.PlayShoot()
				}
				t.Cooldown = t.FireRate // Reset cooldown
				t.Targets = nil
			}
		}
	}
	
	// Update enemies
	for _, e := range g.enemies {
		e.Update(g.pathPositions)
		if !e.IsAlive() && e.Progress >= 1.0 {
			g.lives--
			g.audio.PlayGameOver()
			if g.lives <= 0 {
				g.state = StateGameOver
			}
		}
	}
	
	// Update projectiles
	for _, p := range g.projectiles {
		p.Update()
		
		// Check splash damage
		if p.Type == tower.TowerSplash && !p.Alive {
			for _, e := range g.enemies {
				if e.IsAlive() {
					dist := math.Hypot(e.X-p.X, e.Y-p.Y)
					if dist < 60 {
						e.TakeDamage(p.Damage)
					}
				}
			}
			g.particles.Emit(p.X, p.Y, particle.PTExplosion, 15)
			g.audio.PlayExplosion()
		}
	}
	
	// Clean up
	activeEnemies := make([]*enemy.Enemy, 0)
	for _, e := range g.enemies {
		if e.IsAlive() {
			activeEnemies = append(activeEnemies, e)
		} else if !e.IsAlive() && e.Progress < 1.0 {
			// Died from damage
			g.gold += e.Reward
			g.score += e.Reward * 10
			g.audio.PlayEnemyDeath()
			g.particles.Emit(e.X, e.Y, particle.PTDeath, 10)
			g.particles.Emit(e.X, e.Y, particle.PTCoin, 5)
		}
	}
	g.enemies = activeEnemies
	
	g.projectiles = filterProjectiles(g.projectiles)
	
	// Check wave complete
	if g.waveActive && len(g.enemies) == 0 && g.enemiesSpawned >= g.enemiesInWave {
		g.waveActive = false
		g.waveTimer = config.WaveInterval
		g.score += g.wave * 100 // Wave bonus
		
		if g.wave >= config.MaxWaves {
			g.state = StateVictory
			g.audio.PlayVictory()
		}
	}
}

func (g *Game) startWave() {
	g.wave++
	g.waveActive = true
	g.enemiesSpawned = 0
	g.enemiesInWave = 5 + g.wave*3
	g.spawnTimer = 0
	g.audio.PlayWaveStart()
}

func (g *Game) spawnEnemies() {
	if g.enemiesSpawned >= g.enemiesInWave {
		return
	}
	
	g.spawnTimer--
	if g.spawnTimer > 0 {
		return
	}
	
	g.spawnTimer = 30 - g.wave // Spawn faster in later waves
	if g.spawnTimer < 10 {
		g.spawnTimer = 10
	}
	
	// Determine enemy type based on wave
	var enemyType int
	r := rand.Intn(100)
	
	if g.wave < 5 {
		if r < 70 {
			enemyType = enemy.EnemyBasic
		} else {
			enemyType = enemy.EnemyFast
		}
	} else if g.wave < 10 {
		if r < 50 {
			enemyType = enemy.EnemyBasic
		} else if r < 80 {
			enemyType = enemy.EnemyFast
		} else {
			enemyType = enemy.EnemyTank
		}
	} else if g.wave < 20 {
		if r < 30 {
			enemyType = enemy.EnemyBasic
		} else if r < 50 {
			enemyType = enemy.EnemyFast
		} else if r < 80 {
			enemyType = enemy.EnemyTank
		} else {
			enemyType = enemy.EnemySwarm
		}
	} else {
		if r < 20 {
			enemyType = enemy.EnemyBasic
		} else if r < 40 {
			enemyType = enemy.EnemyFast
		} else if r < 65 {
			enemyType = enemy.EnemyTank
		} else if r < 90 {
			enemyType = enemy.EnemySwarm
		} else {
			enemyType = enemy.EnemyBoss
		}
	}
	
	// Scale enemy HP with wave
	scaleFactor := 1.0 + float64(g.wave)*0.1
	newEnemy := enemy.NewEnemy(enemyType)
	newEnemy.HP *= scaleFactor
	newEnemy.MaxHP *= scaleFactor
	
	g.enemies = append(g.enemies, newEnemy)
	g.enemiesSpawned++
}

func (g *Game) getTowerAt(gx, gy int) *tower.Tower {
	for _, t := range g.towers {
		if t.GridX == gx && t.GridY == gy {
			return t
		}
	}
	return nil
}

func filterProjectiles(projs []*projectile.Projectile) []*projectile.Projectile {
	active := make([]*projectile.Projectile, 0)
	for _, p := range projs {
		if p.Alive {
			active = append(active, p)
		}
	}
	return active
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear
	screen.Fill(color.RGBA{20, 30, 20, 255})
	
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StatePaused:
		g.drawGame(screen)
		if g.state == StatePaused {
			ui.DrawText(screen, "PAUSED", config.ScreenWidth/2-20, config.ScreenHeight/2-10, color.White)
		}
	case StateGameOver:
		g.drawGame(screen)
		ui.DrawText(screen, "GAME OVER", config.ScreenWidth/2-35, config.ScreenHeight/2-20, color.RGBA{255, 50, 50, 255})
		ui.DrawText(screen, fmt.Sprintf("Score: %d", g.score), config.ScreenWidth/2-30, config.ScreenHeight/2+20, color.White)
		ui.DrawText(screen, "Press ENTER to restart", config.ScreenWidth/2-55, config.ScreenHeight/2+50, color.RGBA{200, 200, 200, 255})
	case StateVictory:
		g.drawGame(screen)
		ui.DrawText(screen, "VICTORY!", config.ScreenWidth/2-35, config.ScreenHeight/2-20, color.RGBA{255, 215, 0, 255})
		ui.DrawText(screen, fmt.Sprintf("Final Score: %d", g.score), config.ScreenWidth/2-40, config.ScreenHeight/2+20, color.White)
		ui.DrawText(screen, "Press ENTER to play again", config.ScreenWidth/2-60, config.ScreenHeight/2+50, color.White)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Background gradient (dark blue to dark green)
	vector.DrawFilledRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{10, 20, 40, 255}, false)
	
	// Stars
	for _, star := range g.menuStars {
		twinkle := 0.5 + 0.5*math.Sin(float64(g.frameCount)*star.TwinkleSpeed)
		alpha := uint8(star.Brightness * twinkle * 255)
		vector.DrawFilledCircle(screen, float32(star.X), float32(star.Y), float32(star.Size), color.RGBA{200, 220, 255, alpha}, false)
	}
	
	// Title with glow effect
	for i := 3; i > 0; i-- {
		alpha := uint8(50 * i)
		vector.StrokeCircle(screen, config.ScreenWidth/2, 140, 60+float32(i)*5, 3, color.RGBA{255, 215, 0, alpha}, false)
	}
	ui.DrawText(screen, "TOWER DEFENSE", config.ScreenWidth/2-65, 150, color.RGBA{255, 215, 0, 255})
	ui.DrawText(screen, "Go365 Day 100 - Ebitengine", config.ScreenWidth/2-65, 190, color.RGBA{200, 200, 200, 255})
	
	// Menu items
	items := []string{"START GAME", "HOW TO PLAY", "EXIT"}
	for i, item := range items {
		y := 280 + i*50
		x := config.ScreenWidth/2 - 70
		
		if i == g.menuSelection {
			vector.DrawFilledRect(screen, float32(x-10), float32(y-20), 170, 35, color.RGBA{50, 80, 50, 200}, false)
			ui.DrawText(screen, "> "+item, x, y, color.RGBA{255, 255, 150, 255})
		} else {
			ui.DrawText(screen, "  "+item, x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	
	// Instructions
	ui.DrawText(screen, "W/S or Up/Down: Navigate", config.ScreenWidth/2-65, 480, color.RGBA{150, 150, 150, 255})
	ui.DrawText(screen, "ENTER or J: Select", config.ScreenWidth/2-45, 500, color.RGBA{150, 150, 150, 255})
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Draw map
	g.gameMap.Draw(screen)
	
	// Draw tower placement preview
	if g.selectedTowerType >= 0 && g.hoverGridX >= 0 && g.hoverGridX < config.GridWidth &&
	   g.hoverGridY >= 0 && g.hoverGridY < config.GridHeight {
		sx := float32(g.hoverGridX*config.TileSize + config.GridOffsetX)
		sy := float32(g.hoverGridY*config.TileSize + config.GridOffsetY)
		ts := float32(config.TileSize)
		
		canPlace := g.gameMap.CanPlaceTower(g.hoverGridX, g.hoverGridY)
		stats := tower.GetType(g.selectedTowerType)
		canAfford := g.gold >= stats.Cost
		
		if canPlace && canAfford {
			vector.DrawFilledRect(screen, sx, sy, ts, ts, color.RGBA{0, 255, 0, 100}, false)
			// Range preview
			cx := float32(g.hoverGridX*config.TileSize + config.GridOffsetX + config.TileSize/2)
			cy := float32(g.hoverGridY*config.TileSize + config.GridOffsetY + config.TileSize/2)
			vector.StrokeCircle(screen, cx, cy, float32(stats.Range), 1, color.RGBA{0, 255, 0, 80}, false)
		} else {
			vector.DrawFilledRect(screen, sx, sy, ts, ts, color.RGBA{255, 0, 0, 100}, false)
		}
	}
	
	// Draw towers
	for _, t := range g.towers {
		t.Draw(screen)
	}
	
	// Draw selected tower range
	if g.selectedTower != nil {
		g.selectedTower.DrawRange(screen)
	}
	
	// Draw enemies
	for _, e := range g.enemies {
		e.Draw(screen)
	}
	
	// Draw projectiles
	for _, p := range g.projectiles {
		p.Draw(screen)
	}
	
	// Draw particles
	g.particles.Draw(screen)
	
	// Draw HUD
	g.drawHUD(screen)
	
	// Draw tower selection panel
	if g.showTowerMenu && g.selectedTower != nil {
		g.drawTowerMenu(screen)
	}
	
	// Draw tower build panel
	if g.selectedTowerType >= 0 {
		g.drawBuildPanel(screen)
	}
	
	// Wave incoming text
	if !g.waveActive && g.wave < config.MaxWaves {
		msg := fmt.Sprintf("Wave %d in %d...", g.wave+1, (g.waveTimer+59)/60)
		ui.DrawText(screen, msg, config.ScreenWidth/2-50, config.ScreenHeight-40, color.RGBA{255, 255, 100, 255})
		ui.DrawText(screen, "Press SPACE to start now", config.ScreenWidth/2-65, config.ScreenHeight-20, color.RGBA{200, 200, 200, 255})
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, config.ScreenWidth, 50, color.RGBA{20, 20, 30, 230}, false)
	
	// Gold
	ui.DrawText(screen, fmt.Sprintf("Gold: %d", g.gold), 20, 15, color.RGBA{255, 215, 0, 255})
	
	// Lives
	ui.DrawText(screen, fmt.Sprintf("Lives: %d", g.lives), 200, 15, color.RGBA{255, 100, 100, 255})
	
	// Score
	ui.DrawText(screen, fmt.Sprintf("Score: %d", g.score), 400, 15, color.White)
	
	// Wave
	ui.DrawText(screen, fmt.Sprintf("Wave: %d/%d", g.wave, config.MaxWaves), 600, 15, color.RGBA{150, 200, 255, 255})
	
	// Tower hotkeys
	for i := 0; i < 5; i++ {
		stats := tower.GetType(i)
		x := 20 + i*120
		y := 55
		
		if g.selectedTowerType == i {
			vector.DrawFilledRect(screen, float32(x), float32(y), 100, 45, color.RGBA{50, 80, 50, 200}, false)
		}
		
		// Tower icon
		vector.DrawFilledCircle(screen, float32(x+15), float32(y+20), 10, stats.Color, false)
		
		// Tower name and cost
		ui.DrawText(screen, fmt.Sprintf("[%d]%s", i+1, stats.Name), x+30, y+8, color.White)
		ui.DrawText(screen, fmt.Sprintf("%dg", stats.Cost), x+30, y+24, color.RGBA{255, 215, 0, 255})
	}
}

func (g *Game) drawTowerMenu(screen *ebiten.Image) {
	t := g.selectedTower
	if t == nil {
		return
	}
	
	// Panel background
	px := float32(config.ScreenWidth - 220)
	py := float32(60)
	vector.DrawFilledRect(screen, px, py, 210, 150, color.RGBA{20, 20, 40, 230}, false)
	
	// Tower info
	ui.DrawText(screen, t.GetInfo(), int(px)+10, int(py)+15, color.White)
	
	// Upgrade button
	upgradeCost := t.GetUpgradeCost()
	if upgradeCost > 0 && g.gold >= upgradeCost {
		ui.DrawText(screen, fmt.Sprintf("[U] Upgrade (%dg)", upgradeCost), int(px)+10, int(py)+60, color.RGBA{100, 255, 100, 255})
	} else if upgradeCost > 0 {
		ui.DrawText(screen, fmt.Sprintf("[U] Upgrade (%dg) - Need %d more", upgradeCost, upgradeCost-g.gold), int(px)+10, int(py)+60, color.RGBA{255, 100, 100, 255})
	} else {
		ui.DrawText(screen, "[U] MAX LEVEL", int(px)+10, int(py)+60, color.RGBA{255, 215, 0, 255})
	}
	
	// Sell button
	sellValue := t.GetSellValue()
	ui.DrawText(screen, fmt.Sprintf("[S] Sell (%dg)", sellValue), int(px)+10, int(py)+85, color.RGBA{255, 150, 50, 255})
}

func (g *Game) drawBuildPanel(screen *ebiten.Image) {
	// Already shown in HUD as hotkeys
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	rand.Seed(time.Now().UnixNano())
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

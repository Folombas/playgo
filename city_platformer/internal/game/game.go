// Package game - пиксельная игра Platformer
// Go365 Day 93 - Полностью пиксельная игра!
package game

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"city_platformer/internal/entity"
	"city_platformer/internal/sprite"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
)

type Game struct {
	state        GameState
	player       *entity.Player
	platforms    []*entity.Platform
	houses       []*entity.House
	trees        []*entity.Tree
	enemies      []*entity.Enemy
	items        []*entity.Item
	collectibles []*entity.Collectible
	particles    []Particle

	cameraX float64
	cameraY float64
	score   int
	level   int
	coins   int

	rng         *rand.Rand
	spriteSheet *sprite.SpriteSheet

	jumpPressed bool
}

type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.Color
	Size         float64
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		state: StateMenu,
		rng:   rng,
	}

	var err error
	g.spriteSheet, err = sprite.LoadSpriteSheet()
	if err != nil {
		println("Warning: sprite loading error:", err.Error())
	}

	return g
}

func (g *Game) Reset() {
	g.level = 1
	g.score = 0
	g.coins = 0
	g.startLevel()
}

func (g *Game) startLevel() {
	g.generateLevel(g.level)
}

func (g *Game) generateLevel(levelNum int) {
	levelWidth := 2500 + levelNum*500
	groundY := float64(levelHeight) - 32

	// Игрок
	g.player = entity.NewPlayer(100, groundY-64, g.spriteSheet)
	g.player.Physics.OnGround = true

	// Платформы (земля)
	g.platforms = make([]*entity.Platform, 0)
	g.platforms = append(g.platforms, entity.NewPlatform(0, groundY, float64(levelWidth), 32, "grass", g.spriteSheet))

	// Холмы
	numHills := 5 + levelNum
	for i := 0; i < numHills; i++ {
		hillX := float64(300 + i*400)
		hillY := groundY - 40
		hillWidth := 120.0 + g.rng.Float64()*80
		g.platforms = append(g.platforms, entity.NewPlatform(hillX, hillY, hillWidth, 32, "grassHalf", g.spriteSheet))
	}

	// Домики
	g.houses = make([]*entity.House, 0)
	numHouses := 2 + levelNum
	for i := 0; i < numHouses; i++ {
		houseX := 500.0 + float64(i)*600 + g.rng.Float64()*100
		houseType := "houseBeige"
		if i%3 == 1 {
			houseType = "houseDark"
		} else if i%3 == 2 {
			houseType = "houseGray"
		}
		house := entity.NewHouse(houseX, groundY, houseType, g.spriteSheet)
		g.houses = append(g.houses, house)
	}

	// Деревья
	g.trees = make([]*entity.Tree, 0)
	numTrees := 8 + levelNum*2
	for i := 0; i < numTrees; i++ {
		treeX := 200.0 + float64(i)*250 + g.rng.Float64()*50
		treeType := "pine"
		if i%2 == 0 {
			treeType = "tree"
		}
		tree := entity.NewTree(treeX, groundY, treeType, g.spriteSheet)
		g.trees = append(g.trees, tree)
	}

	// Враги
	g.enemies = make([]*entity.Enemy, 0)
	enemyTypes := []string{"slime", "snake", "spider"}
	numEnemies := 3 + levelNum
	for i := 0; i < numEnemies; i++ {
		x := 400.0 + float64(i)*350 + g.rng.Float64()*100
		y := groundY - 40
		enemy := entity.NewEnemy(x, y, enemyTypes[g.rng.Intn(len(enemyTypes))], g.spriteSheet)
		enemy.PatrolStart = x - 60
		enemy.PatrolEnd = x + 60
		g.enemies = append(g.enemies, enemy)
	}

	// Монетки
	g.items = make([]*entity.Item, 0)
	numCoins := 15 + levelNum*3
	for i := 0; i < numCoins; i++ {
		x := 150.0 + float64(i)*120 + g.rng.Float64()*60
		y := groundY - 50 - g.rng.Float64()*80
		item := entity.NewItem(x, y, entity.ItemCoinGold, 10, g.spriteSheet)
		g.items = append(g.items, item)
	}

	// Звёзды
	g.collectibles = make([]*entity.Collectible, 0)
	numStars := 5 + levelNum
	for i := 0; i < numStars; i++ {
		x := 600.0 + float64(i)*400 + g.rng.Float64()*100
		y := groundY - 100 - g.rng.Float64()*60
		c := entity.NewCollectible(x, y, entity.ItemStar, 50, g.spriteSheet)
		g.collectibles = append(g.collectibles, c)
	}

	g.particles = make([]Particle, 0)
	g.cameraX = 0
	g.cameraY = 0
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused:
			g.state = StatePlaying
		case StateMenu, StateGameOver, StateVictory:
			return ebiten.Termination
		}
	}

	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	if g.state == StateGameOver && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	if g.state == StateVictory && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.level++
		g.startLevel()
		g.state = StatePlaying
	}

	if g.state == StatePlaying {
		g.updateGame()
	}

	return nil
}

func (g *Game) updateGame() {
	dt := 1.0 / 60.0

	g.handleInput()
	g.player.Update(dt)
	g.applyPhysics(dt)
	g.updateCamera()
	g.collectItems()
	g.collectCollectibles()
	g.updateEnemies(dt)
	g.updateParticles(dt)
	g.updateHouses(dt)
	g.checkLevelExit()

	if g.player.Health.Dead {
		g.player.Health.Dead = false
		g.player.Health.Current = g.player.Health.Max
		g.startLevel()
		g.score -= 50
		if g.score < 0 {
			g.score = 0
		}
	}
}

func (g *Game) handleInput() {
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.MoveLeft()
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.MoveRight()
	} else {
		g.player.Stop()
	}

	jumpKey := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)
	if jumpKey && !g.jumpPressed {
		g.jumpPressed = true
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{100, 255, 100, 255})
	} else if !jumpKey {
		g.jumpPressed = false
	}
}

func (g *Game) applyPhysics(dt float64) {
	g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	if g.player.Physics.VelocityY > 800 {
		g.player.Physics.VelocityY = 800
	}

	oldY := g.player.Transform.Y
	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.player.Transform.Y += g.player.Physics.VelocityY * dt
	g.player.Physics.OnGround = false

	for _, p := range g.platforms {
		if entity.CheckCollision(g.player.Transform, p.Transform) {
			if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Transform.Y+10 {
				g.player.Transform.Y = p.Transform.Y - g.player.Transform.Height
				g.player.Physics.VelocityY = 0
				g.player.Physics.OnGround = true
				g.player.ResetJump()
			}
		}
	}

	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
	}
	if g.player.Transform.Y > 800 {
		g.player.Health.TakeDamage(100)
	}
}

func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}
}

func (g *Game) collectItems() {
	for _, item := range g.items {
		if item.Collected {
			continue
		}
		if entity.CheckCollision(g.player.Transform, item.Transform) {
			item.Collected = true
			g.coins++
			g.score += item.Value
			g.spawnParticles(item.Transform.X+16, item.Transform.Y+16, 0, -50, 10, color.RGBA{255, 215, 0, 255})
		}
	}
}

func (g *Game) collectCollectibles() {
	for _, c := range g.collectibles {
		if c.Collected {
			continue
		}
		if entity.CheckCollision(g.player.Transform, c.Transform) {
			c.Collected = true
			g.score += c.Value
			g.spawnParticles(c.Transform.X+16, c.Transform.Y+16, 0, -50, 15, color.RGBA{255, 255, 255, 255})
		}
	}
}

func (g *Game) updateEnemies(dt float64) {
	for _, enemy := range g.enemies {
		enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y)
		if !enemy.Health.Dead && entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 {
				g.player.Health.TakeDamage(enemy.Damage)
				g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
			}
		}
	}
}

func (g *Game) updateHouses(dt float64) {
	for _, house := range g.houses {
		house.Update(dt)
	}
}

func (g *Game) updateParticles(dt float64) {
	active := make([]Particle, 0)
	for i := range g.particles {
		p := &g.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt
		p.Life -= dt * 0.5
		if p.Life > 0 {
			active = append(active, *p)
		}
	}
	g.particles = active
}

func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life: 1.0,
			Color: c,
			Size: 3 + g.rng.Float64()*4,
		})
	}
}

func (g *Game) checkLevelExit() {
	allCollected := true
	for _, c := range g.collectibles {
		if !c.Collected {
			allCollected = false
			break
		}
	}
	if allCollected && g.player.Transform.X > 2000 {
		g.state = StateVictory
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)

	if g.state == StateMenu {
		g.drawMenu(screen)
		return
	}

	for _, p := range g.platforms {
		p.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, house := range g.houses {
		house.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, tree := range g.trees {
		tree.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, item := range g.items {
		item.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, c := range g.collectibles {
		c.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.cameraX, g.cameraY)
	}

	if g.player != nil {
		g.player.Draw(screen, g.cameraX, g.cameraY)
	}

	for _, p := range g.particles {
		vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y-g.cameraY), float32(p.Size), float32(p.Size), p.Color, false)
	}

	if g.state == StatePlaying || g.state == StatePaused {
		g.drawHUD(screen)
	}

	if g.state == StatePaused {
		g.drawPause(screen)
	}
	if g.state == StateGameOver {
		g.drawGameOver(screen)
	}
	if g.state == StateVictory {
		g.drawVictory(screen)
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	// Пиксельное небо - градиент
	for y := 0; y < screenHeight; y++ {
		percent := float64(y) / float64(screenHeight)
		r := uint8(135 + percent*30)
		g_ := uint8(206 - percent*50)
		b := uint8(235 - percent*50)
		vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g_, b, 255}, false)
	}

	// Пиксельные облака
	g.drawClouds(screen)

	// Дальний план с параллаксом
	g.drawDistantScenery(screen)
}

func (g *Game) drawClouds(screen *ebiten.Image) {
	cloudColor := color.RGBA{255, 255, 255, 200}
	for i := 0; i < 8; i++ {
		cloudX := float32((i*180 - int(g.cameraX*0.2)) % (screenWidth + 200))
		cloudY := float32(40 + (i%4)*40)
		if cloudX < 0 {
			cloudX += screenWidth + 200
		}
		vector.DrawFilledRect(screen, cloudX, cloudY, 70, 25, cloudColor, false)
		vector.DrawFilledRect(screen, cloudX+15, cloudY-12, 45, 35, cloudColor, false)
		vector.DrawFilledRect(screen, cloudX+35, cloudY-10, 35, 30, cloudColor, false)
	}
}

func (g *Game) drawDistantScenery(screen *ebiten.Image) {
	// Дальние холмы
	hillColor := color.RGBA{100, 160, 100, 180}
	for i := 0; i < 12; i++ {
		hillX := float32(i*120 - int(g.cameraX*0.3)%120)
		hillHeight := float32(80 + (i*13)%60)
		for x := 0.0; x < 100.0; x++ {
			y := float32(math.Sqrt(100*100 - (x-50)*(x-50))) * (hillHeight / 50)
			vector.DrawFilledRect(screen, hillX+float32(x), float32(screenHeight)-float32(y)-20, 1, float32(y), hillColor, false)
		}
	}

	// Дальние деревья (пиксельные ёлки)
	treeColor := color.RGBA{34, 100, 34, 200}
	for i := 0; i < 15; i++ {
		treeX := float32(i*100 - int(g.cameraX*0.4)%100)
		treeY := float32(screenHeight) - 35
		vector.DrawFilledRect(screen, treeX+35, treeY-60, 30, 60, treeColor, false)
		vector.DrawFilledRect(screen, treeX+42, treeY-90, 16, 35, treeColor, false)
		vector.DrawFilledRect(screen, treeX+47, treeY, 6, 15, color.RGBA{101, 67, 33, 200}, false)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "🎮 PIXEL PLATFORMER"
	ebitenutil.DebugPrintAt(screen, title, screenWidth/2-120, 150)

	subtitle := "Пиксельная прогулка по деревне"
	ebitenutil.DebugPrintAt(screen, subtitle, screenWidth/2-130, 200)

	instructions := []string{
		"",
		"[SPACE] Начать игру",
		"",
		"Управление:",
		"A/D - Ходить",
		"W/Пробел - Прыжок (двойной!)",
		"",
		"Цель: Собрать все звёзды!",
		"Наслаждайтесь пиксель-артом! 🎨",
	}

	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-150, 260+i*22)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	y := 10

	// Здоровье
	vector.DrawFilledRect(screen, 10, float32(y), 200, 20, color.RGBA{50, 50, 50, 255}, false)
	hpPercent := float32(g.player.Health.Current) / float32(g.player.Health.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*hpPercent, 20, color.RGBA{100, 200, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "HP", 220, y)

	y += 25
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), 10, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "COINS: "+string(rune(g.coins)), 10, y)
	y += 20

	collected := 0
	for _, c := range g.collectibles {
		if c.Collected {
			collected++
		}
	}
	ebitenutil.DebugPrintAt(screen, "STARS: "+string(rune(collected))+"/"+string(rune(len(g.collectibles))), 10, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "LEVEL: "+string(rune(g.level)), screenWidth-100, 10)

	ebitenutil.DebugPrintAt(screen, "[ESC] Пауза  [W] Прыжок", 10, screenHeight-30)
}

func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", screenWidth/2-50, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "[ESC] Продолжить", screenWidth/2-80, screenHeight/2)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{150, 50, 50, 200}, false)
	ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-80, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Заново", screenWidth/2-80, screenHeight/2+50)
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 150, 50, 200}, false)
	ebitenutil.DebugPrintAt(screen, "УРОВЕНЬ ПРОЙДЕН!", screenWidth/2-100, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Следующий уровень", screenWidth/2-120, screenHeight/2+50)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

const levelHeight = 600

// Package game - бесконечно генерируемый мир
// Go365 Day 93 - Infinite Pixel Platformer
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
	chunkWidth   = 400  // Ширина чанка в пикселях
	groundY      = 580  // Уровень земли
	maxChunks    = 5    // Максимум активных чанков
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// Chunk - чанк мира
type Chunk struct {
	X            int
	Platforms    []*entity.Platform
	Houses       []*entity.House
	Trees        []*entity.Tree
	Enemies      []*entity.Enemy
	Items        []*entity.Item
	Collectibles []*entity.Collectible
	Generated    bool
}

// World - бесконечный мир
type World struct {
	Chunks       map[int]*Chunk
	ActiveChunks []int
	Seed         int64
	Rng          *rand.Rand
}

// NewWorld создаёт новый мир
func NewWorld(seed int64) *World {
	return &World{
		Chunks:       make(map[int]*Chunk),
		ActiveChunks: make([]int, 0),
		Seed:         seed,
		Rng:          rand.New(rand.NewSource(seed)),
	}
}

// GetChunk получает или создаёт чанк
func (w *World) GetChunk(chunkX int) *Chunk {
	if chunk, ok := w.Chunks[chunkX]; ok {
		return chunk
	}

	// Создаём новый чанк
	chunk := &Chunk{
		X:         chunkX * chunkWidth,
		Generated: false,
	}
	w.Chunks[chunkX] = chunk
	w.generateChunk(chunk)
	w.ActiveChunks = append(w.ActiveChunks, chunkX)

	// Удаляем старые чанки
	w.cleanupOldChunks()

	return chunk
}

// generateChunk генерирует чанк
func (w *World) generateChunk(chunk *Chunk) {
	w.Rng.Seed(w.Seed + int64(chunk.X))

	// Земля (с возможными ямами)
	hasPit := w.Rng.Float64() < 0.2 // 20% шанс ямы
	if !hasPit {
		chunk.Platforms = append(chunk.Platforms,
			entity.NewPlatform(float64(chunk.X), groundY, chunkWidth, 32, "grass", nil),
		)
	} else {
		// Земля с ямой посередине
		chunk.Platforms = append(chunk.Platforms,
			entity.NewPlatform(float64(chunk.X), groundY, chunkWidth*0.4, 32, "grass", nil),
			entity.NewPlatform(float64(chunk.X)+chunkWidth*0.6, groundY, chunkWidth*0.4, 32, "grass", nil),
		)
	}

	// Холмы и платформы
	numPlatforms := w.Rng.Intn(4) + 2
	for i := 0; i < numPlatforms; i++ {
		px := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-100)
		py := groundY - 80 - w.Rng.Float64()*120
		width := 80 + w.Rng.Float64()*60
		tileType := "stone"
		if w.Rng.Float64() < 0.3 {
			tileType = "grassHalf"
		}
		chunk.Platforms = append(chunk.Platforms,
			entity.NewPlatform(px, py, width, 32, tileType, nil),
		)
	}

	// Домики (если есть место)
	if w.Rng.Float64() < 0.4 && !hasPit {
		houseX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-100)
		houseType := "houseBeige"
		r := w.Rng.Float64()
		if r < 0.33 {
			houseType = "houseDark"
		} else if r < 0.66 {
			houseType = "houseGray"
		}
		house := entity.NewHouse(houseX, groundY, houseType, nil)
		chunk.Houses = append(chunk.Houses, house)
	}

	// Деревья
	numTrees := w.Rng.Intn(3) + 1
	for i := 0; i < numTrees; i++ {
		treeX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		treeType := "pine"
		if w.Rng.Float64() < 0.5 {
			treeType = "tree"
		}
		tree := entity.NewTree(treeX, groundY, treeType, nil)
		chunk.Trees = append(chunk.Trees, tree)
	}

	// Враги
	if w.Rng.Float64() < 0.6 && !hasPit {
		enemyX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-100)
		enemyType := "slime"
		r := w.Rng.Float64()
		if r < 0.33 {
			enemyType = "snake"
		} else if r < 0.66 {
			enemyType = "spider"
		}
		enemy := entity.NewEnemy(enemyX, groundY-40, enemyType, nil)
		enemy.PatrolStart = enemyX - 60
		enemy.PatrolEnd = enemyX + 60
		chunk.Enemies = append(chunk.Enemies, enemy)
	}

	// Монетки
	numCoins := w.Rng.Intn(5) + 3
	for i := 0; i < numCoins; i++ {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 50 - w.Rng.Float64()*150
		item := entity.NewItem(x, y, entity.ItemCoinGold, 10, nil)
		chunk.Items = append(chunk.Items, item)
	}

	// Звёзды (редко)
	if w.Rng.Float64() < 0.3 {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 120 - w.Rng.Float64()*80
		c := entity.NewCollectible(x, y, entity.ItemStar, 50, nil)
		chunk.Collectibles = append(chunk.Collectibles, c)
	}

	chunk.Generated = true
}

// cleanupOldChunks удаляет старые чанки
func (w *World) cleanupOldChunks() {
	if len(w.ActiveChunks) > maxChunks {
		// Сортируем чанки
		sorted := make([]int, len(w.ActiveChunks))
		copy(sorted, w.ActiveChunks)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i+1; j < len(sorted); j++ {
				if sorted[i] > sorted[j] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		// Удаляем самый старый (левый) чанк
		oldChunkX := sorted[0]
		delete(w.Chunks, oldChunkX)

		// Обновляем активные чанки
		newActive := make([]int, 0)
		for _, x := range w.ActiveChunks {
			if x != oldChunkX {
				newActive = append(newActive, x)
			}
		}
		w.ActiveChunks = newActive
	}
}

// Update обновляет все чанки
func (w *World) Update(dt float64) {
	for _, chunk := range w.Chunks {
		for _, house := range chunk.Houses {
			house.Update(dt)
		}
		for _, enemy := range chunk.Enemies {
			enemy.Update(dt, 0, 0) // Упрощённо
		}
		for _, item := range chunk.Items {
			item.Update(dt)
		}
		for _, c := range chunk.Collectibles {
			c.Update(dt)
		}
	}
}

// Draw отрисовывает чанки
func (w *World) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	for _, chunk := range w.Chunks {
		for _, p := range chunk.Platforms {
			p.Draw(screen, cameraX, cameraY)
		}
		for _, house := range chunk.Houses {
			house.Draw(screen, cameraX, cameraY)
		}
		for _, tree := range chunk.Trees {
			tree.Draw(screen, cameraX, cameraY)
		}
		for _, item := range chunk.Items {
			item.Draw(screen, cameraX, cameraY)
		}
		for _, c := range chunk.Collectibles {
			c.Draw(screen, cameraX, cameraY)
		}
		for _, enemy := range chunk.Enemies {
			enemy.Draw(screen, cameraX, cameraY)
		}
	}
}

// Game - основная игра
type Game struct {
	state   GameState
	player  *entity.Player
	world   *World
	cameraX float64

	score       int
	coins       int
	distance    int
	maxDistance int

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

var particles []Particle

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
	g.score = 0
	g.coins = 0
	g.distance = 0
	g.maxDistance = 0
	g.cameraX = 0

	// Создаём мир с случайным seed
	g.world = NewWorld(time.Now().UnixNano())

	// Игрок
	g.player = entity.NewPlayer(100, groundY-64, g.spriteSheet)
	g.player.Physics.OnGround = true

	// Генерируем первые чанки
	for i := 0; i < 3; i++ {
		g.world.GetChunk(i)
	}

	particles = make([]Particle, 0)
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused:
			g.state = StatePlaying
		case StateMenu, StateGameOver:
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
	g.world.Update(dt)
	g.checkCollisions()

	// Обновляем дистанцию
	playerChunk := int(g.player.Transform.X / chunkWidth)
	if playerChunk*chunkWidth > g.maxDistance {
		g.maxDistance = playerChunk * chunkWidth
		g.distance = g.maxDistance
		g.score = g.maxDistance / 10
	}

	// Смерть игрока
	if g.player.Transform.Y > 800 {
		g.player.Health.TakeDamage(100)
	}

	if g.player.Health.Dead {
		g.state = StateGameOver
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

	// Коллизия с платформами всех чанков
	for _, chunk := range g.world.Chunks {
		for _, p := range chunk.Platforms {
			if entity.CheckCollision(g.player.Transform, p.Transform) {
				if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Transform.Y+10 {
					g.player.Transform.Y = p.Transform.Y - g.player.Transform.Height
					g.player.Physics.VelocityY = 0
					g.player.Physics.OnGround = true
					g.player.ResetJump()
				}
			}
		}
	}

	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
		g.player.Physics.VelocityX = 0
	}
}

func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1

	// Генерируем новые чанки впереди
	chunkAhead := int((g.cameraX + screenWidth) / chunkWidth)
	g.world.GetChunk(chunkAhead + 1)
}

func (g *Game) collectItems() {
	for _, chunk := range g.world.Chunks {
		for _, item := range chunk.Items {
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
}

func (g *Game) collectCollectibles() {
	for _, chunk := range g.world.Chunks {
		for _, c := range chunk.Collectibles {
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
}

func (g *Game) updateEnemies(dt float64) {
	for _, chunk := range g.world.Chunks {
		for _, enemy := range chunk.Enemies {
			enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y)
			if !enemy.Health.Dead && entity.CheckCollision(g.player.Transform, enemy.Transform) {
				if g.player.Health.Invincible <= 0 {
					g.player.Health.TakeDamage(enemy.Damage)
					g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
				}
			}
		}
	}
}

func (g *Game) checkCollisions() {
	// Проверка коллизий с домиками
	for _, chunk := range g.world.Chunks {
		for _, house := range chunk.Houses {
			if entity.CheckCollision(g.player.Transform, house.Transform) {
				// Отталкиваем игрока
				if g.player.Transform.X < house.Transform.X {
					g.player.Transform.X = house.Transform.X - g.player.Transform.Width
				} else {
					g.player.Transform.X = house.Transform.X + house.Transform.Width
				}
				g.player.Physics.VelocityX = 0
			}
		}
	}
}

func (g *Game) updateParticles(dt float64) {
	active := make([]Particle, 0)
	for i := range particles {
		p := &particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt
		p.Life -= dt * 0.5
		if p.Life > 0 {
			active = append(active, *p)
		}
	}
	particles = active
}

func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		particles = append(particles, Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life: 1.0,
			Color: c,
			Size: 3 + g.rng.Float64()*4,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)

	if g.state == StateMenu {
		g.drawMenu(screen)
		return
	}

	// Отрисовка мира
	g.world.Draw(screen, g.cameraX, 0)

	// Игрок
	if g.player != nil {
		g.player.Draw(screen, g.cameraX, 0)
	}

	// Частицы
	for _, p := range particles {
		vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y), float32(p.Size), float32(p.Size), p.Color, false)
	}

	// HUD
	if g.state == StatePlaying || g.state == StatePaused {
		g.drawHUD(screen)
	}

	if g.state == StatePaused {
		g.drawPause(screen)
	}
	if g.state == StateGameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	// Небо
	for y := 0; y < screenHeight; y++ {
		percent := float64(y) / float64(screenHeight)
		r := uint8(135 + percent*30)
		g_ := uint8(206 - percent*50)
		b := uint8(235 - percent*50)
		vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g_, b, 255}, false)
	}

	// Облака (параллакс)
	cloudColor := color.RGBA{255, 255, 255, 200}
	for i := 0; i < 10; i++ {
		cloudX := float32((i*200 - int(g.cameraX*0.1)) % (screenWidth + 200))
		if cloudX < 0 {
			cloudX += screenWidth + 200
		}
		cloudY := float32(40 + (i%5)*30)
		vector.DrawFilledRect(screen, cloudX, cloudY, 80, 25, cloudColor, false)
		vector.DrawFilledRect(screen, cloudX+15, cloudY-12, 50, 35, cloudColor, false)
	}

	// Дальние холмы (параллакс)
	hillColor := color.RGBA{100, 160, 100, 180}
	for i := 0; i < 15; i++ {
		hillX := float32(i*100 - int(g.cameraX*0.2)%100)
		hillHeight := float32(80 + (i*13)%60)
		for x := 0.0; x < 100.0; x++ {
			y := float32(math.Sqrt(100*100 - (x-50)*(x-50))) * (hillHeight / 50)
			vector.DrawFilledRect(screen, hillX+float32(x), float32(screenHeight)-float32(y)-20, 1, float32(y), hillColor, false)
		}
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "🎮 INFINITE PIXEL PLATFORMER"
	ebitenutil.DebugPrintAt(screen, title, screenWidth/2-160, 150)

	subtitle := "Бесконечная процедурная генерация!"
	ebitenutil.DebugPrintAt(screen, subtitle, screenWidth/2-140, 200)

	instructions := []string{
		"",
		"[SPACE] Начать игру",
		"",
		"Управление:",
		"A/D - Ходить влево/вправо",
		"W/Пробел - Прыжок (двойной!)",
		"",
		"Цель: Пройти как можно дальше!",
		"Собирай монетки и звёзды! 💰⭐",
		"",
		"Мир генерируется бесконечно! 🌍",
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
	ebitenutil.DebugPrintAt(screen, "DISTANCE: "+string(rune(g.distance))+"m", 10, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, "MAX: "+string(rune(g.maxDistance))+"m", 10, y)

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
	ebitenutil.DebugPrintAt(screen, "DISTANCE: "+string(rune(g.maxDistance))+"m", screenWidth/2-90, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2+30)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Заново", screenWidth/2-80, screenHeight/2+80)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

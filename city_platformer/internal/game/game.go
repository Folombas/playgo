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
	Mushrooms    []*entity.Mushroom
	Frogs        []*entity.Frog
	Butterflies  []*entity.Butterfly
	Cacti        []*entity.Cactus
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
	SpriteSheet  *sprite.SpriteSheet
}

// NewWorld создаёт новый мир
func NewWorld(seed int64, ss *sprite.SpriteSheet) *World {
	return &World{
		Chunks:       make(map[int]*Chunk),
		ActiveChunks: make([]int, 0),
		Seed:         seed,
		Rng:          rand.New(rand.NewSource(seed)),
		SpriteSheet:  ss,
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
	w.generateChunk(chunk, w.SpriteSheet)
	w.ActiveChunks = append(w.ActiveChunks, chunkX)

	// Удаляем старые чанки
	w.cleanupOldChunks()

	return chunk
}

// generateChunk генерирует чанк
func (w *World) generateChunk(chunk *Chunk, ss *sprite.SpriteSheet) {
	w.Rng.Seed(w.Seed + int64(chunk.X))

	// Твёрдая земля с газоном - несколько слоёв чтобы не было просвета!
	chunk.Platforms = append(chunk.Platforms,
		// Слой 1: Трава
		entity.NewPlatform(float64(chunk.X), groundY, chunkWidth, 32, "grass", ss),
		// Слой 2: Земля
		entity.NewPlatform(float64(chunk.X), groundY+32, chunkWidth, 32, "dirt", ss),
		// Слой 3: Ещё земля
		entity.NewPlatform(float64(chunk.X), groundY+64, chunkWidth, 32, "dirt", ss),
		// Слой 4: Камень внизу
		entity.NewPlatform(float64(chunk.X), groundY+96, chunkWidth, 32, "stone", ss),
	)

	// Ямы (20% шанс) - убираем среднюю платформу
	hasPit := w.Rng.Float64() < 0.2
	if hasPit {
		// Разрываем землю на две части (все слои)
		pitStart := float64(chunk.X) + chunkWidth*0.4
		
		chunk.Platforms = append(chunk.Platforms,
			// Левая сторона
			entity.NewPlatform(float64(chunk.X), groundY, chunkWidth*0.4, 32, "grass", ss),
			entity.NewPlatform(float64(chunk.X), groundY+32, chunkWidth*0.4, 32, "dirt", ss),
			entity.NewPlatform(float64(chunk.X), groundY+64, chunkWidth*0.4, 32, "dirt", ss),
			entity.NewPlatform(float64(chunk.X), groundY+96, chunkWidth*0.4, 32, "stone", ss),
			// Правая сторона
			entity.NewPlatform(pitStart, groundY, chunkWidth*0.4, 32, "grass", ss),
			entity.NewPlatform(pitStart, groundY+32, chunkWidth*0.4, 32, "dirt", ss),
			entity.NewPlatform(pitStart, groundY+64, chunkWidth*0.4, 32, "dirt", ss),
			entity.NewPlatform(pitStart, groundY+96, chunkWidth*0.4, 32, "stone", ss),
		)
	}

	// Холмы и платформы - аккуратно, без наложений!
	numPlatforms := w.Rng.Intn(3) + 2
	usedY := make(map[int]bool) // Чтобы платформы не накладывались
	
	for i := 0; i < numPlatforms; i++ {
		px := float64(chunk.X) + 50 + w.Rng.Float64()*float64(chunkWidth-150)
		
		// Выбираем высоту из фиксированных уровней (чтобы не накладывались)
		level := w.Rng.Intn(3) // 0, 1, или 2
		py := groundY - 80 - float64(level)*70
		
		// Проверяем чтобы не было наложения по Y
		key := int(py)
		if usedY[key] {
			py += 35 // Сдвигаем если занято
		}
		usedY[int(py)] = true
		
		width := 80 + w.Rng.Float64()*40
		tileType := "stone"
		if w.Rng.Float64() < 0.4 {
			tileType = "grassHalf"
		}
		chunk.Platforms = append(chunk.Platforms,
			entity.NewPlatform(px, py, width, 32, tileType, ss),
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
		tree := entity.NewTree(treeX, groundY, treeType, ss)
		chunk.Trees = append(chunk.Trees, tree)
	}

	// Грибы
	numMushrooms := w.Rng.Intn(4) + 2
	for i := 0; i < numMushrooms; i++ {
		mushX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		mushY := float64(groundY - 16)
		mushType := "red"
		r := w.Rng.Float64()
		if r < 0.33 {
			mushType = "brown"
		} else if r < 0.66 {
			mushType = "tan"
		}
		mush := entity.NewMushroom(mushX, mushY, mushType, ss)
		chunk.Mushrooms = append(chunk.Mushrooms, mush)
	}

	// Лягушки
	if w.Rng.Float64() < 0.4 {
		frogX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		frogY := float64(groundY - 16)
		frog := entity.NewFrog(frogX, frogY, ss)
		chunk.Frogs = append(chunk.Frogs, frog)
	}

	// Бабочки
	numButterflies := w.Rng.Intn(3) + 1
	for i := 0; i < numButterflies; i++ {
		bflyX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		bflyY := groundY - 60 - w.Rng.Float64()*40
		bfly := entity.NewButterfly(bflyX, bflyY, ss)
		chunk.Butterflies = append(chunk.Butterflies, bfly)
	}

	// Кактусы (редко, 20% шанс)
	if w.Rng.Float64() < 0.2 {
		cactusX := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		cactus := entity.NewCactus(cactusX, groundY, ss)
		chunk.Cacti = append(chunk.Cacti, cactus)
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
		enemy := entity.NewEnemy(enemyX, groundY-40, enemyType, ss)
		enemy.PatrolStart = enemyX - 60
		enemy.PatrolEnd = enemyX + 60
		chunk.Enemies = append(chunk.Enemies, enemy)
	}

	// Монетки
	numCoins := w.Rng.Intn(5) + 3
	for i := 0; i < numCoins; i++ {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 50 - w.Rng.Float64()*150
		coinType := entity.ItemCoinGold
		r := w.Rng.Float64()
		if r < 0.3 {
			coinType = entity.ItemCoinSilver
		} else if r < 0.5 {
			coinType = entity.ItemCoinBronze
		}
		item := entity.NewItem(x, y, coinType, 10, ss)
		chunk.Items = append(chunk.Items, item)
	}

	// Кристаллы
	if w.Rng.Float64() < 0.5 {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 80 - w.Rng.Float64()*100
		gemType := entity.ItemGemRed
		r := w.Rng.Float64()
		if r < 0.25 {
			gemType = entity.ItemGemBlue
		} else if r < 0.5 {
			gemType = entity.ItemGemGreen
		} else if r < 0.75 {
			gemType = entity.ItemGemYellow
		}
		item := entity.NewItem(x, y, gemType, 25, ss)
		chunk.Items = append(chunk.Items, item)
	}

	// Грибы-предметы
	if w.Rng.Float64() < 0.3 {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := float64(groundY - 40)
		mushType := entity.ItemMushroom
		item := entity.NewItem(x, y, mushType, 15, ss)
		chunk.Items = append(chunk.Items, item)
	}

	// Звёзды (редко)
	if w.Rng.Float64() < 0.3 {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 120 - w.Rng.Float64()*80
		c := entity.NewCollectible(x, y, entity.ItemStar, 50, ss)
		chunk.Collectibles = append(chunk.Collectibles, c)
	}

	// Ключи (очень редко)
	if w.Rng.Float64() < 0.15 {
		x := float64(chunk.X) + w.Rng.Float64()*float64(chunkWidth-50)
		y := groundY - 60 - w.Rng.Float64()*80
		keyType := entity.ItemKeyBlue
		r := w.Rng.Float64()
		if r < 0.33 {
			keyType = entity.ItemKeyGreen
		} else if r < 0.66 {
			keyType = entity.ItemKeyRed
		}
		item := entity.NewItem(x, y, keyType, 30, ss)
		chunk.Items = append(chunk.Items, item)
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
		for _, mush := range chunk.Mushrooms {
			mush.Update(dt)
		}
		for _, frog := range chunk.Frogs {
			frog.Update(dt)
		}
		for _, bfly := range chunk.Butterflies {
			bfly.Update(dt)
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
		for _, mush := range chunk.Mushrooms {
			mush.Draw(screen, cameraX, cameraY)
		}
		for _, frog := range chunk.Frogs {
			frog.Draw(screen, cameraX, cameraY)
		}
		for _, bfly := range chunk.Butterflies {
			bfly.Draw(screen, cameraX, cameraY)
		}
		for _, cactus := range chunk.Cacti {
			cactus.Draw(screen, cameraX, cameraY)
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
	g.world = NewWorld(time.Now().UnixNano(), g.spriteSheet)

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
	// Красивое небо (sky.png) - заполняем весь экран
	skyImg := g.spriteSheet.GetBackground("sky")
	if skyImg != nil {
		skyWidth := float64(skyImg.Bounds().Dx())
		
		// Рисуем небо с повторением по горизонтали
		for i := 0; i < 3; i++ {
			x := float64(i)*skyWidth - math.Mod(g.cameraX*0.02, skyWidth)
			if x < -skyWidth {
				x += skyWidth * 3
			}
			
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(x, 0)
			screen.DrawImage(skyImg, opts)
		}
	} else {
		// Запасной вариант - градиент
		for y := 0; y < screenHeight; y++ {
			percent := float64(y) / float64(screenHeight)
			r := uint8(135 + percent*30)
			g_ := uint8(206 - percent*50)
			b := uint8(235 - percent*50)
			vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g_, b, 255}, false)
		}
	}

	// Дальние скалы/горы (параллакс слой 1)
	g.drawParallaxLayer(screen, "rocks_1", 0.03)
	
	// Ближние скалы (параллакс слой 2)
	g.drawParallaxLayer(screen, "rocks_2", 0.05)
	
	// Деревья на заднем плане (параллакс слой 3)
	g.drawParallaxLayer(screen, "parallax-mountain-trees", 0.1)

	// Облака - спрайты!
	cloudNames := []string{"clouds_1", "clouds_2", "clouds_3", "clouds_4"}
	
	for i := 0; i < 5; i++ {
		cloudName := cloudNames[i%len(cloudNames)]
		cloudImg := g.spriteSheet.GetBackground(cloudName)
		
		if cloudImg == nil {
			// Запасной вариант - векторы
			cloudColor := color.RGBA{255, 255, 255, 200}
			cloudX := float32((i*200 - int(g.cameraX*0.05)) % (screenWidth + 200))
			if cloudX < 0 {
				cloudX += screenWidth + 200
			}
			cloudY := float32(60 + (i%3)*35)
			vector.DrawFilledRect(screen, cloudX, cloudY, 90, 30, cloudColor, false)
			vector.DrawFilledRect(screen, cloudX+20, cloudY-15, 60, 40, cloudColor, false)
			continue
		}
		
		cloudX := float64((i*280 - int(g.cameraX*0.02)) % (screenWidth + 300))
		if cloudX < 0 {
			cloudX += screenWidth + 300
		}
		cloudY := float64(80 + (i%3)*40)
		
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(cloudX, cloudY)
		screen.DrawImage(cloudImg, opts)
	}
}

// drawParallaxLayer рисует слой параллакса
func (g *Game) drawParallaxLayer(screen *ebiten.Image, name string, parallaxSpeed float64) {
	img := g.spriteSheet.GetBackground(name)
	if img == nil {
		return
	}
	
	imgWidth := float64(img.Bounds().Dx())
	offset := g.cameraX * parallaxSpeed
	
	// Рисуем два раза для бесшовности
	for i := 0; i < 2; i++ {
		x := float64(i)*imgWidth - math.Mod(offset, imgWidth)
		if x < -imgWidth {
			x += imgWidth * 2
		}
		
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(x, 0)
		screen.DrawImage(img, opts)
	}
}

func (g *Game) drawDistantScenery(screen *ebiten.Image) {
	// Дальние холмы - зелёные (параллакс)
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

// Package game содержит основную игровую логику Kingdom Garden
package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/playgo/go90/internal/entity"
)

// State определяет текущее состояние игры
type State int

const (
	StateMenu State = iota
	StatePlaying
	StatePaused
	StateBuild
	StateInventory
)

// Config содержит конфигурацию игры
type Config struct {
	ScreenWidth  int
	ScreenHeight int
	TargetFPS    int
	TileSize     int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		ScreenWidth:  1280,
		ScreenHeight: 720,
		TargetFPS:    60,
		TileSize:     48,
	}
}

// Game представляет основную игру
type Game struct {
	config      *Config
	state       State
	cameraX     float64
	cameraY     float64
	world       *entity.World
	player      *entity.Player
	clouds      []*entity.Cloud
	particles   []*entity.Particle
	animals     []*entity.Animal
	dayTime     float64
	season      Season
	coins       int
	unlocked    map[string]bool
	rng         *rand.Rand
	selectedItem string
	buildMode   bool
}

// Season представляет время года
type Season int

const (
	SeasonSpring Season = iota
	SeasonSummer
	SeasonFall
	SeasonWinter
)

// NewGame создаёт новую игру
func NewGame() *Game {
	cfg := DefaultConfig()
	entCfg := &entity.Config{
		ScreenWidth:  cfg.ScreenWidth,
		ScreenHeight: cfg.ScreenHeight,
		TargetFPS:    cfg.TargetFPS,
		TileSize:     cfg.TileSize,
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		config:    cfg,
		state:     StateMenu,
		world:     entity.NewWorld(entCfg),
		player:    entity.NewPlayer(640, 360),
		clouds:    make([]*entity.Cloud, 0),
		particles: make([]*entity.Particle, 0),
		animals:   make([]*entity.Animal, 0),
		dayTime:   0.5, // 0-1, 0.5 = полдень
		season:    SeasonSpring,
		coins:     100,
		unlocked:  make(map[string]bool),
		rng:       rng,
	}

	// Инициализация облаков
	for i := 0; i < 15; i++ {
		g.clouds = append(g.clouds, entity.NewCloud(rng.Float64()*float64(cfg.ScreenWidth), rng.Float64()*200, rng))
	}

	// Инициализация животных
	g.spawnAnimal("bunny", 400, 500)
	g.spawnAnimal("bird", 800, 300)
	g.spawnAnimal("butterfly", 600, 400)

	// Размещение деревьев и растений
	g.world.PlaceTree(300, 400)
	g.world.PlaceTree(900, 350)
	g.world.PlaceTree(500, 200)
	g.world.PlaceFlower(400, 450, "pink")
	g.world.PlaceFlower(420, 460, "yellow")
	g.world.PlaceFlower(380, 470, "blue")
	g.world.PlaceCastle(700, 250)

	return g
}

// spawnAnimal создаёт животное
func (g *Game) spawnAnimal(animalType string, x, y float64) {
	g.animals = append(g.animals, entity.NewAnimal(animalType, x, y, g.rng))
}

// Update обновляет логику игры
func (g *Game) Update() error {
	// Обработка ввода
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused, StateBuild, StateInventory:
			g.state = StatePlaying
			g.buildMode = false
		case StateMenu:
			return ebiten.Termination
		}
	}

	// Переход из меню в игру
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.state = StatePlaying
	}

	// Режим строительства
	if g.state == StatePlaying && ebiten.IsKeyPressed(ebiten.KeyB) {
		g.buildMode = !g.buildMode
		if g.buildMode {
			g.state = StateBuild
		} else {
			g.state = StatePlaying
		}
	}

	// Инвентарь
	if g.state == StatePlaying && ebiten.IsKeyPressed(ebiten.KeyI) {
		if g.state == StateInventory {
			g.state = StatePlaying
		} else {
			g.state = StateInventory
		}
	}

	// Обновление игры
	if g.state == StatePlaying || g.state == StateBuild {
		g.updateGame()
	}

	return nil
}

// updateGame обновляет игровую логику
func (g *Game) updateGame() {
	// Обновление времени суток
	g.dayTime += 0.0001
	if g.dayTime > 1 {
		g.dayTime = 0
	}

	// Обновление игрока
	g.player.Update()

	// Управление камерой (слежение за игроком)
	g.cameraX = g.player.X - float64(g.config.ScreenWidth)/2
	g.cameraY = g.player.Y - float64(g.config.ScreenHeight)/2

	// Ограничение камеры границами мира
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraY < 0 {
		g.cameraY = 0
	}
	if g.cameraX > float64(g.world.Width*g.config.TileSize)-float64(g.config.ScreenWidth) {
		g.cameraX = float64(g.world.Width*g.config.TileSize) - float64(g.config.ScreenWidth)
	}
	if g.cameraY > float64(g.world.Height*g.config.TileSize)-float64(g.config.ScreenHeight) {
		g.cameraY = float64(g.world.Height*g.config.TileSize) - float64(g.config.ScreenHeight)
	}

	// Управление игроком
	g.handleInput()

	// Обновление облаков
	for _, cloud := range g.clouds {
		cloud.Update()
	}

	// Обновление животных
	for _, animal := range g.animals {
		animal.Update()
	}

	// Обновление частиц
	g.updateParticles()

	// Спавн новых животных
	if g.rng.Float64() < 0.001 {
		x := g.player.X + (g.rng.Float64()-0.5)*600
		y := g.player.Y + (g.rng.Float64()-0.5)*400
		types := []string{"bunny", "bird", "butterfly"}
		g.spawnAnimal(types[g.rng.Intn(len(types))], x, y)
	}

	// Режим строительства
	if g.buildMode {
		g.handleBuildInput()
	}
}

// handleInput обрабатывает ввод игрока
func (g *Game) handleInput() {
	speed := g.player.MoveSpeed

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Y += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += speed
	}

	// Взаимодействие (посадка растения)
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.plantSomething()
	}
}

// handleBuildInput обрабатывает ввод в режиме строительства
func (g *Game) handleBuildInput() {
	// Выбор предмета цифрами
	if ebiten.IsKeyPressed(ebiten.Key1) {
		g.selectedItem = "tree"
	}
	if ebiten.IsKeyPressed(ebiten.Key2) {
		g.selectedItem = "flower"
	}
	if ebiten.IsKeyPressed(ebiten.Key3) {
		g.selectedItem = "house"
	}
	if ebiten.IsKeyPressed(ebiten.Key4) {
		g.selectedItem = "fence"
	}
	if ebiten.IsKeyPressed(ebiten.Key5) {
		g.selectedItem = "path"
	}

	// Размещение левой кнопкой мыши
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		worldX := float64(mx) + g.cameraX
		worldY := float64(my) + g.cameraY
		g.placeBuilding(worldX, worldY)
	}
}

// plantSomething сажает растение рядом с игроком
func (g *Game) plantSomething() {
	types := []string{"flower", "tree", "bush"}
	chosen := types[g.rng.Intn(len(types))]

	// Размещение перед игроком
	offsetX := 50.0
	offsetY := 0.0

	targetX := g.player.X + offsetX
	targetY := g.player.Y + offsetY

	switch chosen {
	case "flower":
		colors := []string{"pink", "yellow", "blue", "red"}
		g.world.PlaceFlower(targetX, targetY, colors[g.rng.Intn(len(colors))])
		g.spawnParticles(targetX, targetY, 10, color.RGBA{255, 100, 200, 255})
	case "tree":
		g.world.PlaceTree(targetX, targetY)
		g.spawnParticles(targetX, targetY, 15, color.RGBA{50, 200, 50, 255})
	case "bush":
		g.world.PlaceBush(targetX, targetY)
		g.spawnParticles(targetX, targetY, 8, color.RGBA{100, 150, 50, 255})
	}

	g.coins += 5 // Награда за посадку
}

// placeBuilding размещает постройку
func (g *Game) placeBuilding(x, y float64) {
	switch g.selectedItem {
	case "tree":
		if g.coins >= 10 {
			g.world.PlaceTree(x, y)
			g.coins -= 10
			g.spawnParticles(x, y, 10, color.RGBA{50, 200, 50, 255})
		}
	case "flower":
		if g.coins >= 5 {
			colors := []string{"pink", "yellow", "blue", "red", "purple"}
			g.world.PlaceFlower(x, y, colors[g.rng.Intn(len(colors))])
			g.coins -= 5
			g.spawnParticles(x, y, 8, color.RGBA{255, 100, 200, 255})
		}
	case "house":
		if g.coins >= 50 {
			g.world.PlaceHouse(x, y)
			g.coins -= 50
			g.spawnParticles(x, y, 20, color.RGBA{200, 150, 100, 255})
		}
	case "fence":
		if g.coins >= 3 {
			g.world.PlaceFence(x, y)
			g.coins -= 3
		}
	case "path":
		if g.coins >= 2 {
			g.world.PlacePath(x, y)
			g.coins -= 2
		}
	}
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, entity.NewParticle(
			x,
			y,
			(g.rng.Float64()-0.5)*4,
			(g.rng.Float64()-0.5)*4,
			c,
		))
	}
}

// updateParticles обновляет частицы
func (g *Game) updateParticles() {
	activeParticles := make([]*entity.Particle, 0)
	for _, p := range g.particles {
		p.Update()
		if p.Life > 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	g.particles = activeParticles
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана (небо)
	screen.Fill(g.getSkyColor())

	// Отрисовка мира
	g.world.Draw(screen, g.cameraX, g.cameraY)

	// Отрисовка животных
	for _, animal := range g.animals {
		animal.Draw(screen, g.cameraX, g.cameraY)
	}

	// Отрисовка частиц
	for _, p := range g.particles {
		p.Draw(screen, g.cameraX, g.cameraY)
	}

	// Отрисовка игрока
	g.player.Draw(screen, g.cameraX, g.cameraY)

	// Отрисовка облаков
	for _, cloud := range g.clouds {
		cloud.Draw(screen)
	}

	// UI
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StateBuild:
		g.drawHUD(screen)
		if g.buildMode {
			g.drawBuildMenu(screen)
		}
	case StateInventory:
		g.drawHUD(screen)
		g.drawInventory(screen)
	case StatePaused:
		g.drawHUD(screen)
		g.drawPause(screen)
	}
}

// getSkyColor возвращает цвет неба в зависимости от времени суток
func (g *Game) getSkyColor() color.Color {
	if g.dayTime < 0.25 || g.dayTime > 0.75 {
		// Ночь
		return color.RGBA{20, 20, 60, 255}
	} else if g.dayTime < 0.3 || g.dayTime > 0.7 {
		// Рассвет/закат
		return color.RGBA{255, 150, 100, 255}
	}
	// День
	return color.RGBA{100, 180, 255, 255}
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.config.ScreenWidth, g.config.ScreenHeight
}

// drawMenu отрисовывает главное меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := `
╔═══════════════════════════════════════════╗
║      🌸 KINGDOM GARDEN 🏰                 ║
║         Мир Сад Королевства               ║
╠═══════════════════════════════════════════╣
║                                           ║
║      [SPACE] - Начать игру                ║
║      [ESC] - Выход                        ║
║                                           ║
║  🎮 Управление:                           ║
║     WASD / Стрелки - Перемещение          ║
║     E - Посадить растение                 ║
║     B - Режим строительства               ║
║     I - Инвентарь                         ║
║                                           ║
║  🌸 Строй, сажай, украшай!                ║
║  🦋 Никаких врагов - только красота!      ║
║                                           ║
╚═══════════════════════════════════════════╝
`
	ebitenutil.DebugPrint(screen, title)
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	// Фон HUD
	vector.DrawFilledRect(screen, 0, 0, 200, 120, color.RGBA{0, 0, 0, 150}, true)

	hudText := fmt.Sprintf(`┌──────────────────┐
│  🌸 KINGDOM      │
├──────────────────┤
│  💰 Монеты: %4d   │
│  🏰 Построек: %3d │
│  🌳 Деревьев: %3d │
│  🌷 Цветов: %4d   │
└──────────────────┘

💰 +5 за посадку

[B] Строительство
[I] Инвентарь
[ESC] Пауза
`, g.coins, g.world.BuildingCount(), g.world.TreeCount(), g.world.FlowerCount())

	ebitenutil.DebugPrint(screen, hudText)

	// Выбранный предмет в режиме строительства
	if g.buildMode {
		itemText := fmt.Sprintf("🔨 Строим: %s", g.selectedItem)
		ebitenutil.DebugPrintAt(screen, itemText, g.config.ScreenWidth-200, 20)
	}
}

// drawBuildMenu отрисовывает меню строительства
func (g *Game) drawBuildMenu(screen *ebiten.Image) {
	menu := `
╔═══════════════════════════════════════╗
║         🛠️ СТРОИТЕЛЬСТВО 🛠️            ║
╠═══════════════════════════════════════╣
║  [1] 🌳 Дерево (10💰)                 ║
║  [2] 🌷 Цветок (5💰)                  ║
║  [3] 🏠 Домик (50💰)                  ║
║  [4] 🚪 Забор (3💰)                   ║
║  [5] 🛤️ Дорожка (2💰)                 ║
╠═══════════════════════════════════════╣
║  ЛКМ - Разместить                     ║
║  [B] - Выход                          ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, menu, g.config.ScreenWidth/2-180, g.config.ScreenHeight/2-150)
}

// drawInventory отрисовывает инвентарь
func (g *Game) drawInventory(screen *ebiten.Image) {
	// Полупрозрачный фон
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	invText := `
╔═══════════════════════════════════════╗
║            🎒 ИНВЕНТАРЬ 🎒            ║
╠═══════════════════════════════════════╣
║                                       ║
║  🌳 Семена деревьев: ∞                ║
║  🌷 Семена цветов: ∞                  ║
║  🏠 Чертежи домов: 3                  ║
║  🚪 Секции забора: 20                 ║
║  🛤️ Плитки дорожки: 50                ║
║                                       ║
║  💰 Монеты: %d                         ║
║                                       ║
╠═══════════════════════════════════════╣
║  [I] - Закрыть                        ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(invText, g.coins), g.config.ScreenWidth/2-180, g.config.ScreenHeight/2-150)
}

// drawPause отрисовывает меню паузы
func (g *Game) drawPause(screen *ebiten.Image) {
	// Полупрозрачный фон
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 128})

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	pauseText := `
╔═══════════════════════════════════════╗
║              ⏸️ ПАУЗА                  ║
╠═══════════════════════════════════════╣
║                                       ║
║     [ESC] - Продолжить                ║
║                                       ║
║     Твой сад прекрасен! 🌸            ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, g.config.ScreenWidth/2-180, g.config.ScreenHeight/2-100)
}

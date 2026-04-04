package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ======================== КОНСТАНТЫ ========================
const (
	Tile   = 48
	GW     = 15
	GH     = 13
	HUD    = 40
	WinW   = GW * Tile
	WinH   = GH*Tile + HUD
	CD     = 8
	ECMD   = 15
	BOMB_T = 180
	EXPL_T = 30
)

// ======================== ТАЙЛЫ ========================
const (
	T_EMPTY = iota
	T_STONE
	T_BRICK
)

// ======================== НАПРАВЛЕНИЯ ========================
const (
	D_DOWN = iota
	D_UP
	D_LEFT
	D_RIGHT
)

// ======================== ЦВЕТА ========================
var (
	C_BG        = color.RGBA{50, 50, 80, 255}
	C_HUD_BG    = color.RGBA{20, 20, 40, 255}
	C_STONE     = color.RGBA{120, 120, 140, 255}
	C_BRICK     = color.RGBA{180, 100, 60, 255}
	C_GRASS     = color.RGBA{60, 140, 60, 255}
	C_PLAYER    = color.RGBA{255, 150, 200, 255}
	C_PLAYER_H  = color.RGBA{255, 200, 230, 255}
	C_ENEMY_BAL = color.RGBA{255, 120, 120, 255}
	C_ENEMY_CHA = color.RGBA{100, 100, 255, 255}
	C_ENEMY_SPD = color.RGBA{100, 255, 100, 255}
	C_ENEMY_E   = color.RGBA{255, 255, 100, 255}
	C_ENEMY_TEL = color.RGBA{200, 100, 255, 255}
	C_BOMB      = color.RGBA{40, 40, 40, 255}
	C_BOMB_H    = color.RGBA{80, 80, 80, 255}
	C_BOMB_KICK = color.RGBA{100, 100, 100, 255}
	C_EXPL      = color.RGBA{255, 200, 50, 255}
	C_EXPL_C    = color.RGBA{255, 255, 200, 255}
	C_WHITE     = color.RGBA{255, 255, 255, 255}
	C_GREEN     = color.RGBA{100, 255, 100, 255}
	C_RED       = color.RGBA{255, 80, 80, 255}
	C_YELLOW    = color.RGBA{255, 255, 80, 255}
	C_KEY_DEBUG = color.RGBA{0, 255, 150, 255}
	C_DOOR      = color.RGBA{200, 180, 100, 255}
	C_SHIELD    = color.RGBA{100, 200, 255, 200}

	// Power-up цвета
	C_PU_FIRE  = color.RGBA{255, 100, 0, 255}
	C_PU_BOMB  = color.RGBA{80, 80, 80, 255}
	C_PU_SPEED = color.RGBA{255, 255, 0, 255}
	C_PU_HEART = color.RGBA{255, 50, 100, 255}
	C_PU_SHIELD = color.RGBA{100, 200, 255, 255}
	C_PU_KICK  = color.RGBA{200, 150, 100, 255}
)

// ======================== СОСТОЯНИЕ ========================
type State int

const (
	S_MENU State = iota
	S_PLAY
	S_DEAD
	S_WIN
)

// ======================== POWER-UPS ========================
type PowerUp struct {
	gx, gy int
	kind   int // 0=fire, 1=bomb, 2=speed, 3=heart, 4=shield, 5=kick
	anim   int
	frame  int
}

const (
	PU_FIRE = iota
	PU_BOMB
	PU_SPEED
	PU_HEART
	PU_SHIELD
	PU_KICK
)

func puColor(kind int) color.RGBA {
	switch kind {
	case PU_FIRE:
		return C_PU_FIRE
	case PU_BOMB:
		return C_PU_BOMB
	case PU_SPEED:
		return C_PU_SPEED
	case PU_HEART:
		return C_PU_HEART
	case PU_SHIELD:
		return C_PU_SHIELD
	case PU_KICK:
		return C_PU_KICK
	default:
		return C_WHITE
	}
}

func puSymbol(kind int) string {
	switch kind {
	case PU_FIRE:
		return "🔥"
	case PU_BOMB:
		return "💣"
	case PU_SPEED:
		return "⚡"
	case PU_HEART:
		return "❤"
	case PU_SHIELD:
		return "🛡"
	case PU_KICK:
		return "👟"
	default:
		return "?"
	}
}

func puName(kind int) string {
	switch kind {
	case PU_FIRE:
		return "Fire+1"
	case PU_BOMB:
		return "Bomb+1"
	case PU_SPEED:
		return "Speed+"
	case PU_HEART:
		return "Life+1"
	case PU_SHIELD:
		return "Shield"
	case PU_KICK:
		return "Kick"
	default:
		return "?"
	}
}

// ======================== ЧАСТИЦЫ ========================
type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	maxLife  int
	clr      color.RGBA
	size     int
}

func spawnParticles(gx, gy int, clr color.RGBA, count int) []Particle {
	ps := make([]Particle, 0, count)
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1.0 + rand.Float64()*3.0
		ps = append(ps, Particle{
			x: float64(gx*Tile + Tile/2),
			y: float64(gy*Tile + HUD + Tile/2),
			vx: math.Cos(angle) * speed,
			vy: math.Sin(angle) * speed - 1,
			life: 20 + rand.Intn(20),
			maxLife: 40,
			clr: clr,
			size: 2 + rand.Intn(4),
		})
	}
	return ps
}

// ======================== ТРЯСКА ЭКРАНА ========================
type ScreenShake struct {
	intensity float64
	timer     int
}

func (s *ScreenShake) shake(intensity float64, duration int) {
	s.intensity = intensity
	s.timer = duration
}

func (s *ScreenShake) update() {
	if s.timer > 0 {
		s.timer--
	} else {
		s.intensity = 0
	}
}

func (s *ScreenShake) offset() (float64, float64) {
	if s.intensity == 0 {
		return 0, 0
	}
	return (rand.Float64() - 0.5) * s.intensity * 2,
		(rand.Float64() - 0.5) * s.intensity * 2
}

// ======================== ИГРОК ========================
type Player struct {
	gx, gy    int
	dir       int
	lives     int
	bombs     int
	active    int
	radius    int
	cd        int
	anim      int
	frame     int
	inv       int
	shield    bool
	speedBoost int // бонус скорости
	kick      bool  // пинать бомбы
	score     int
	combos    int
	comboTimer int
}

func NewPlayer() *Player {
	return &Player{gx: 1, gy: 1, dir: D_DOWN, lives: 3, bombs: 1, radius: 2, inv: 90, cd: CD}
}

func (p *Player) Update(g *Game) {
	if p.inv > 0 {
		p.inv--
	}
	if p.cd > 0 {
		p.cd--
	}
	if p.comboTimer > 0 {
		p.comboTimer--
		if p.comboTimer == 0 {
			p.combos = 0
		}
	}

	if p.cd > 0 {
		return
	}

	dx, dy := 0, 0
	pressed := false

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dy, dx, pressed = -1, 0, true
		p.dir = D_UP
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		dy, dx, pressed = 1, 0, true
		p.dir = D_DOWN
	} else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dy, dx, pressed = 0, -1, true
		p.dir = D_LEFT
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dy, dx, pressed = 0, 1, true
		p.dir = D_RIGHT
	}

	if pressed {
		g.keys = ""
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.keys += "W"
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.keys += "A"
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.keys += "S"
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.keys += "D"
		}

		p.frame++
		p.anim = 1 + (p.frame/6)%2
	} else {
		p.anim = 0
		p.frame = 0
		g.keys = ""
	}

	if dx == 0 && dy == 0 {
		return
	}

	// Пытаемся пнуть бомбу
	if p.kick {
		for i, b := range g.bombs {
			if (b.gx == p.gx+dx && b.gy == p.gy) || (b.gx == p.gx && b.gy == p.gy+dy) {
				if b.timer > BOMB_T - 60 { // только что поставленную
					// Пинуть бомбу
					newX, newY := b.gx+dx, b.gy+dy
					if g.walkable(newX, newY, p.gx, p.gy) {
						g.bombs[i].gx = newX
						g.bombs[i].gy = newY
						p.cd = CD
						return
					}
				}
			}
		}
	}

	nx, ny := p.gx+dx, p.gy+dy
	if g.walkable(nx, ny, p.gx, p.gy) {
		p.gx, p.gy = nx, ny
		p.cd = CD
		if p.speedBoost > 0 {
			p.cd = CD / 2
		}
	} else if g.walkable(nx, p.gy, p.gx, p.gy) {
		p.gx = nx
		p.cd = CD
	} else if g.walkable(p.gx, ny, p.gx, p.gy) {
		p.gy = ny
		p.cd = CD
	}
}

func (p *Player) Draw(s *ebiten.Image) {
	if p.inv > 0 && p.inv%8 < 4 {
		return
	}

	px, py := p.gx*Tile, p.gy*Tile+HUD

	// Shield эффект
	if p.shield {
		rect(s, px+2, py+2, Tile-4, Tile-4, C_SHIELD)
	}

	// Тело
	rect(s, px+4, py+4, Tile-8, Tile-8, C_PLAYER)
	rect(s, px+12, py+2, Tile-24, 16, C_PLAYER_H)

	// Глаза
	ex, ey := px+Tile/2, py+Tile/3
	switch p.dir {
	case D_UP:
		ey -= 4
	case D_DOWN:
		ey += 4
	case D_LEFT:
		ex -= 6
	case D_RIGHT:
		ex += 6
	}
	rect(s, ex-3, ey-3, 3, 3, color.Black)
	rect(s, ex+3, ey-3, 3, 3, color.Black)

	if p.anim > 0 {
		rect(s, px+8, py+Tile-6, 8, 6, C_PLAYER)
		rect(s, px+Tile-16, py+Tile-6, 8, 6, C_PLAYER)
	}
}

// ======================== ВРАГИ ========================
const (
	E_BALLOON = iota // Случайное движение, медленный
	E_CHASER         // Преследует игрока
	E_SPLITTER       // Делится при смерти
	E_TELEPORTER     // Телепортируется
)

type Enemy struct {
	gx, gy int
	kind   int
	dir    int
	alive  bool
	cd     int
	anim   int
	frame  int
	hp     int // HP для сплиттера
}

func NewEnemy(gx, gy int, kind int) *Enemy {
	hp := 1
	if kind == E_SPLITTER {
		hp = 2
	}
	return &Enemy{gx: gx, gy: gy, kind: kind, dir: rand.Intn(4), alive: true, cd: ECMD/2, hp: hp}
}

func (e *Enemy) Update(g *Game) {
	if !e.alive {
		return
	}
	if e.cd > 0 {
		e.cd--
		return
	}

	e.frame++
	e.anim = (e.frame / 8) % 2

	switch e.kind {
	case E_BALLOON:
		e.updateBalloon(g)
	case E_CHASER:
		e.updateChaser(g)
	case E_SPLITTER:
		e.updateSplitter(g)
	case E_TELEPORTER:
		e.updateTeleporter(g)
	}
}

func (e *Enemy) updateBalloon(g *Game) {
	e.cd = ECMD + rand.Intn(20)
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateChaser(g *Game) {
	e.cd = ECMD - 3 + rand.Intn(10) // Быстрее чем Balloon

	// Движение к игроку с небольшим рандомом
	dx := g.player.gx - e.gx
	dy := g.player.gy - e.gy

	moves := [][2]int{}
	if rand.Float64() < 0.7 { // 70% идти к игроку
		if dx > 0 {
			moves = append(moves, [2]int{1, 0})
		} else if dx < 0 {
			moves = append(moves, [2]int{-1, 0})
		}
		if dy > 0 {
			moves = append(moves, [2]int{0, 1})
		} else if dy < 0 {
			moves = append(moves, [2]int{0, -1})
		}
	}
	// 30% случайное
	moves = append(moves, [2]int{0, -1}, [2]int{0, 1}, [2]int{-1, 0}, [2]int{1, 0})

	rand.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })

	for _, d := range moves {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateSplitter(g *Game) {
	e.cd = ECMD + rand.Intn(15)
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		nx, ny := e.gx+d[0], e.gy+d[1]
		if g.walkableEnemy(nx, ny) {
			e.gx, e.gy = nx, ny
			return
		}
	}
}

func (e *Enemy) updateTeleporter(g *Game) {
	e.cd = 60 + rand.Intn(40) // Телепорт каждые 1-1.5 сек

	// Найти случайное свободное место
	for attempts := 0; attempts < 20; attempts++ {
		nx := 1 + rand.Intn(GW-2)
		ny := 1 + rand.Intn(GH-2)
		if g.grid[ny][nx] == T_EMPTY {
			// Не рядом с игроком
			dist := abs(nx-g.player.gx) + abs(ny-g.player.gy)
			if dist > 3 {
				e.gx, e.gy = nx, ny
				// Частицы при телепортации
				g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_ENEMY_TEL, 8)...)
				return
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (e *Enemy) Draw(s *ebiten.Image) {
	if !e.alive {
		return
	}
	px, py := e.gx*Tile, e.gy*Tile+HUD

	// Цвет по типу
	clr := C_ENEMY_BAL
	switch e.kind {
	case E_CHASER:
		clr = C_ENEMY_CHA
	case E_SPLITTER:
		clr = C_ENEMY_SPD
	case E_TELEPORTER:
		clr = C_ENEMY_TEL
	}

	// Тело (разные формы для разных типов)
	switch e.kind {
	case E_BALLOON:
		// Круг
		rect(s, px+4, py+4, Tile-8, Tile-8, clr)
		rect(s, px+8, py+2, Tile-16, 4, clr)
	case E_CHASER:
		// Стрелка к игроку
		rect(s, px+4, py+4, Tile-8, Tile-8, clr)
		if g_player != nil {
			if e.gx < g_player.gx {
				rect(s, px+Tile-6, py+Tile/2-4, 6, 8, C_ENEMY_E)
			} else if e.gx > g_player.gx {
				rect(s, px, py+Tile/2-4, 6, 8, C_ENEMY_E)
			}
		}
	case E_SPLITTER:
		// Квадрат + HP индикатор
		rect(s, px+2, py+2, Tile-4, Tile-4, clr)
		// Полоска HP
		if e.hp > 1 {
			rect(s, px+4, py+Tile-8, (Tile-8)*e.hp/2, 4, C_ENEMY_E)
		}
	case E_TELEPORTER:
		// Ромб
		rect(s, px+Tile/2-4, py+4, 8, Tile-8, clr)
		rect(s, px+4, py+Tile/2-4, Tile-8, 8, clr)
	}

	// Глаза
	rect(s, px+10, py+14, 8, 8, C_ENEMY_E)
	rect(s, px+Tile-18, py+14, 8, 8, C_ENEMY_E)
}

// Глобальная ссылка для Chaser AI
var g_player *Player

// ======================== БОМБА ========================
type Bomb struct {
	gx, gy   int
	timer    int
	radius   int
	kicked   bool
	owner    int // ID владельца (для подсчёта)
}

// ======================== ВЗРЫВ ========================
type Cell struct {
	gx, gy int
	t      int
}

// ======================== ИГРА ========================
type Game struct {
	grid   [][]int
	player *Player
	enemies []*Enemy
	bombs  []Bomb
	explos []Cell
	pus    []PowerUp     // power-ups на поле
	particles []Particle
	state  State
	keys   string
	sprites map[string]*ebiten.Image
	level  int
	spacePrev bool
	enterPrev bool
	shake  ScreenShake
	score  int
}

func NewGame() *Game {
	g := &Game{state: S_MENU, sprites: make(map[string]*ebiten.Image), level: 1}
	g.loadSprites()
	initAudio()
	return g
}

func (g *Game) loadSprites() {
	tryLoad := func(name, file string) {
		img, _, err := ebitenutil.NewImageFromFile("assets/sprites/" + file)
		if err == nil {
			g.sprites[name] = img
		}
	}
	tryLoad("player", "player_stand.png")
	tryLoad("enemy", "enemy1.png")
	tryLoad("bomb", "bomb.png")
	tryLoad("brick", "brick.png")
	tryLoad("stone", "stone.png")
	tryLoad("grass", "grass.png")
	tryLoad("explosion", "explosion.png")

	if len(g.sprites) > 0 {
		fmt.Printf("✓ Loaded %d sprites\n", len(g.sprites))
	} else {
		fmt.Println("✗ No sprites loaded, using primitives")
	}
}

func (g *Game) initLevel() {
	g.grid = make([][]int, GH)
	for y := range g.grid {
		g.grid[y] = make([]int, GW)
	}

	// Каменная сетка
	for y := 0; y < GH; y += 2 {
		for x := 0; x < GW; x += 2 {
			g.grid[y][x] = T_STONE
		}
	}

	// Кирпичи + power-ups под ними
	r := rand.New(rand.NewSource(42 + int64(g.level)*7))
	g.pus = nil

	for y := 1; y < GH-1; y++ {
		for x := 1; x < GW-1; x++ {
			if x <= 2 && y <= 2 {
				continue
			}
			if r.Float32() < 0.30 {
				g.grid[y][x] = T_BRICK
				// 25% шанс power-up под кирпичом
				if r.Float32() < 0.25 {
					kind := r.Intn(6)
					g.pus = append(g.pus, PowerUp{gx: x, gy: y, kind: kind})
				}
			}
		}
	}

	g.player = NewPlayer()
	g.bombs = nil
	g.explos = nil
	g.particles = nil

	// Враги: сложнее с каждым уровнем
	g.enemies = nil
	count := 3 + g.level
	positions := [][2]int{}
	for y := 1; y < GH-1; y++ {
		for x := 4; x < GW-1; x++ {
			if g.grid[y][x] == T_EMPTY {
				positions = append(positions, [2]int{x, y})
			}
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	for i := 0; i < count && i < len(positions); i++ {
		kind := E_BALLOON
		if g.level >= 2 && i < count/3 {
			kind = E_CHASER
		}
		if g.level >= 3 && i < count/4 {
			kind = E_SPLITTER
		}
		if g.level >= 4 && i < count/5 {
			kind = E_TELEPORTER
		}
		g.enemies = append(g.enemies, NewEnemy(positions[i][0], positions[i][1], kind))
	}

	g_player = g.player // для AI
}

func (g *Game) walkable(nx, ny, fromX, fromY int) bool {
	if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
		return false
	}
	if g.grid[ny][nx] != T_EMPTY {
		return false
	}
	for _, b := range g.bombs {
		if b.gx == nx && b.gy == ny && !(b.gx == fromX && b.gy == fromY) {
			return false
		}
	}
	return true
}

func (g *Game) walkableEnemy(nx, ny int) bool {
	if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
		return false
	}
	if g.grid[ny][nx] != T_EMPTY {
		return false
	}
	for _, b := range g.bombs {
		if b.gx == nx && b.gy == ny {
			return false
		}
	}
	return true
}

func (g *Game) Update() error {
	frameCount++

	// ===== МЕНЮ =====
	if g.state == S_MENU {
		enter := ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)
		if enter && !g.enterPrev {
			g.state = S_PLAY
			g.initLevel()
			g.score = 0
			playSound(soundMenu)
			fmt.Println("▶ Game started! Level", g.level)
		}
		g.enterPrev = enter
		return nil
	}

	// ===== GAME OVER / WIN =====
	if g.state == S_DEAD || g.state == S_WIN {
		enter := ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)
		if enter && !g.enterPrev {
			if g.state == S_WIN {
				g.level++
				g.initLevel()
				g.state = S_PLAY
				playSound(soundMenu)
				fmt.Println("▶ Level", g.level)
			} else {
				g.level = 1
				g.initLevel()
				g.state = S_PLAY
				g.score = 0
				playSound(soundMenu)
				fmt.Println("▶ Restart")
			}
		}
		g.enterPrev = enter
		return nil
	}

	// ===== ИГРА =====
	space := ebiten.IsKeyPressed(ebiten.KeySpace)
	g_player = g.player

	g.player.Update(g)
	g.shake.update()

	// Бомба
	if space && !g.spacePrev && g.player.active < g.player.bombs {
		hasBomb := false
		for _, b := range g.bombs {
			if b.gx == g.player.gx && b.gy == g.player.gy {
				hasBomb = true
				break
			}
		}
		if !hasBomb {
			g.bombs = append(g.bombs, Bomb{
				gx: g.player.gx, gy: g.player.gy,
				timer: BOMB_T, radius: g.player.radius,
			})
			g.player.active++
			playSound(soundPlace)
		}
	}
	g.spacePrev = space

	// Рестарт
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.level = 1
		g.initLevel()
		g.state = S_PLAY
		g.score = 0
	}

	// Бомбы
	for i := len(g.bombs) - 1; i >= 0; i-- {
		g.bombs[i].timer--
		if g.bombs[i].timer <= 0 {
			g.doExplosion(i)
		}
	}

	// Взрывы
	for i := len(g.explos) - 1; i >= 0; i-- {
		g.explos[i].t--
		if g.explos[i].t <= 0 {
			g.explos = append(g.explos[:i], g.explos[i+1:]...)
		}
	}

	// Частицы
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.15 // гравитация
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Враги
	for _, e := range g.enemies {
		e.Update(g)
	}

	// Коллизия: игрок ↔ враг
	if g.player.inv <= 0 {
		for _, e := range g.enemies {
			if e.alive && e.gx == g.player.gx && e.gy == g.player.gy {
				g.playerHit()
			}
		}
	}

	// Коллизия: игрок ↔ взрыв
	for _, c := range g.explos {
		if c.gx == g.player.gx && c.gy == g.player.gy && g.player.inv <= 0 {
			g.playerHit()
		}
	}

	// Коллизия: враг ↔ взрыв
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		for _, c := range g.explos {
			if c.gx == e.gx && c.gy == e.gy {
				e.hp--
				if e.hp <= 0 {
					e.alive = false
					g.player.combos++
					g.player.comboTimer = 60
					bonus := 100 * g.player.combos
					g.player.score += bonus
					g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_EXPL, 12)...)

					// Splitter: делится на 2
					if e.kind == E_SPLITTER {
						for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
							nx, ny := e.gx+d[0], e.gy+d[1]
							if g.walkableEnemy(nx, ny) {
								child := NewEnemy(nx, ny, E_BALLOON)
								child.cd = 30
								g.enemies = append(g.enemies, child)
								break
							}
						}
					}

					playSound(soundKill)
				} else {
					// Hit но не dead
					g.particles = append(g.particles, spawnParticles(e.gx, e.gy, C_ENEMY_E, 6)...)
					playSound(soundPlace)
				}
			}
		}
	}

	// Проверка победы
	alive := 0
	for _, e := range g.enemies {
		if e.alive {
			alive++
		}
	}
	if alive == 0 && len(g.enemies) > 0 {
		g.state = S_WIN
		g.player.score += 500 * g.level
		playSound(soundWin)
		g.shake.shake(5, 20)
		fmt.Println("🏆 Level complete! Score:", g.player.score)
	}

	return nil
}

func (g *Game) playerHit() {
	if g.player.shield {
		g.player.shield = false
		g.player.inv = 60
		g.particles = append(g.particles, spawnParticles(g.player.gx, g.player.gy, C_SHIELD, 16)...)
		playSound(soundPlace)
		return
	}

	g.player.lives--
	g.player.inv = 90
	g.shake.shake(4, 15)
	g.particles = append(g.particles, spawnParticles(g.player.gx, g.player.gy, C_RED, 20)...)
	playSound(soundDie)

	if g.player.lives <= 0 {
		g.state = S_DEAD
		fmt.Println("💀 Game Over! Score:", g.player.score)
	}
}

func (g *Game) doExplosion(idx int) {
	b := g.bombs[idx]
	g.bombs = append(g.bombs[:idx], g.bombs[idx+1:]...)
	g.player.active--
	playSound(soundExpl)
	g.shake.shake(3, 10)

	// Центр
	g.explos = append(g.explos, Cell{b.gx, b.gy, EXPL_T})
	g.particles = append(g.particles, spawnParticles(b.gx, b.gy, C_EXPL, 8)...)

	// 4 направления
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range dirs {
		for i := 1; i <= b.radius; i++ {
			nx, ny := b.gx+d[0]*i, b.gy+d[1]*i
			if nx < 0 || nx >= GW || ny < 0 || ny >= GH {
				break
			}
			if g.grid[ny][nx] == T_STONE {
				g.particles = append(g.particles, spawnParticles(nx, ny, C_STONE, 4)...)
				break
			}
			g.explos = append(g.explos, Cell{nx, ny, EXPL_T})

			// Разрушение кирпича
			if g.grid[ny][nx] == T_BRICK {
				g.grid[ny][nx] = T_EMPTY
				g.particles = append(g.particles, spawnParticles(nx, ny, C_BRICK, 10)...)

				// Проверяем power-up под кирпичом
				for j := len(g.pus) - 1; j >= 0; j-- {
					if g.pus[j].gx == nx && g.pus[j].gy == ny {
						// Оставляем на поле
						g.pus = append(g.pus[:j], g.pus[j+1:]...)
						break
					}
				}
				break
			}

			// Цепная реакция
			for j := range g.bombs {
				if g.bombs[j].gx == nx && g.bombs[j].gy == ny {
					g.bombs[j].timer = 1
				}
			}
		}
	}
}

func (g *Game) applyPowerUp(pu *PowerUp) {
	playSound(soundMenu)
	g.particles = append(g.particles, spawnParticles(pu.gx, pu.gy, puColor(pu.kind), 12)...)

	switch pu.kind {
	case PU_FIRE:
		g.player.radius++
		fmt.Println("🔥 Fire +1 → radius", g.player.radius)
	case PU_BOMB:
		g.player.bombs++
		fmt.Println("💣 Bomb +1 → max", g.player.bombs)
	case PU_SPEED:
		g.player.speedBoost++
		g.player.cd = CD / 2
		fmt.Println("⚡ Speed Up!")
	case PU_HEART:
		g.player.lives++
		fmt.Println("❤ Life +1 →", g.player.lives)
	case PU_SHIELD:
		g.player.shield = true
		fmt.Println("🛡 Shield ON!")
	case PU_KICK:
		g.player.kick = true
		fmt.Println("👟 Kick ON!")
	}
}

func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(C_BG)

	if g.state == S_MENU {
		g.drawMenu(s)
		return
	}

	// Тряска экрана
	sx, sy := g.shake.offset()

	// Поле
	for y := 0; y < GH; y++ {
		for x := 0; x < GW; x++ {
			px, py := x*Tile+int(sx), y*Tile+HUD+int(sy)

			// Трава
			if spr := g.sprites["grass"]; spr != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
				op.GeoM.Translate(float64(px), float64(py))
				s.DrawImage(spr, op)
			} else {
				rect(s, px, py, Tile, Tile, C_GRASS)
			}

			switch g.grid[y][x] {
			case T_STONE:
				if spr := g.sprites["stone"]; spr != nil {
					drawTile(spr, s, px, py)
				} else {
					rect(s, px+1, py+1, Tile-2, Tile-2, C_STONE)
					rect(s, px+4, py+4, Tile-8, Tile-8, color.RGBA{140, 140, 160, 255})
				}
			case T_BRICK:
				if spr := g.sprites["brick"]; spr != nil {
					drawTile(spr, s, px, py)
				} else {
					rect(s, px+1, py+1, Tile-2, Tile-2, C_BRICK)
					rect(s, px, py+Tile/2, Tile, 2, color.RGBA{140, 70, 40, 255})
					rect(s, px+Tile/2, py, 2, Tile/2, color.RGBA{140, 70, 40, 255})
					rect(s, px+Tile/4, py+Tile/2, 2, Tile/2, color.RGBA{140, 70, 40, 255})
					rect(s, px+Tile*3/4, py+Tile/2, 2, Tile/2, color.RGBA{140, 70, 40, 255})
				}

				// Индикатор power-up под кирпичом
				for _, pu := range g.pus {
					if pu.gx == x && pu.gy == y {
						pu.frame++
						if pu.frame%20 < 10 {
							rect(s, px+Tile-8, py+2, 6, 6, puColor(pu.kind))
						}
					}
				}
			}
		}
	}

	// Power-ups на поле
	for i := len(g.pus) - 1; i >= 0; i-- {
		pu := &g.pus[i]
		pu.anim = (pu.frame / 10) % 2

		px, py := pu.gx*Tile, pu.gy*Tile+HUD

		// Проверяем, есть ли здесь игрок
		if g.player.gx == pu.gx && g.player.gy == pu.gy {
			g.applyPowerUp(pu)
			g.pus = append(g.pus[:i], g.pus[i+1:]...)
			continue
		}

		// Рисуем
		clr := puColor(pu.kind)
		bobY := py
		if pu.anim == 1 {
			bobY -= 4
		}
		rect(s, px+8, bobY+8, Tile-16, Tile-16, clr)
		rect(s, px+12, bobY+4, Tile-24, 6, puColor(pu.kind))

		// Символ
		sym := puSymbol(pu.kind)
		text.Draw(s, sym, basicfont.Face7x13, px+16, bobY+Tile-12, C_WHITE)
	}

	// Взрывы
	for _, c := range g.explos {
		px, py := c.gx*Tile, c.gy*Tile+HUD
		a := float64(c.t) / EXPL_T
		if spr := g.sprites["explosion"]; spr != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
			op.GeoM.Translate(float64(px), float64(py))
			op.ColorM.Scale(1, 1, 1, a)
			s.DrawImage(spr, op)
		} else {
			rectAlpha(s, px, py, Tile, Tile, C_EXPL, a)
		}
	}

	// Частицы
	for _, p := range g.particles {
		a := float64(p.life) / float64(p.maxLife)
		rectAlpha(s, int(p.x)-p.size/2, int(p.y)-p.size/2, p.size, p.size, p.clr, a)
	}

	// Бомбы
	for _, b := range g.bombs {
		px, py := b.gx*Tile, b.gy*Tile+HUD
		pulse := 1.0
		if b.timer < 60 {
			pulse = 1.0 + 0.15*(float64(b.timer%12)/12)
		}
		if spr := g.sprites["bomb"]; spr != nil {
			op := &ebiten.DrawImageOptions{}
			sc := float64(Tile) / float64(spr.Bounds().Dx()) * pulse
			op.GeoM.Scale(sc, sc)
			off := (float64(Tile) - float64(spr.Bounds().Dx())*sc) / 2
			op.GeoM.Translate(float64(px)+off, float64(py)+off)
			s.DrawImage(spr, op)
		} else {
			sz := int(float64(Tile-8) * pulse)
			off := (Tile - sz) / 2
			rect(s, px+off, py+off, sz, sz, C_BOMB)
			rect(s, px+off+4, py+off+4, sz-8, sz-8, C_BOMB_H)
			rect(s, px+Tile/2-1, py+2, 3, 8, C_BRICK)

			// Мигание когда скоро взорвётся
			if b.timer < 30 && b.timer%4 < 2 {
				rect(s, px+off, py+off, sz, sz, C_EXPL)
			}
		}
	}

	// Враги
	for _, e := range g.enemies {
		e.Draw(s)
	}

	// Игрок
	g.player.Draw(s)

	// HUD (без тряски)
	g.drawHUD(s)

	// Debug: клавиши
	if g.keys != "" {
		text.Draw(s, "KEYS:"+g.keys, basicfont.Face7x13, WinW-70, WinH-8, C_KEY_DEBUG)
	}

	// Game Over / Win overlay
	if g.state == S_DEAD {
		g.drawOverlay(s, "GAME OVER", C_RED)
	} else if g.state == S_WIN {
		g.drawOverlay(s, fmt.Sprintf("LEVEL %d COMPLETE! Score: %d", g.level, g.player.score), C_GREEN)
	}
}

func (g *Game) drawHUD(s *ebiten.Image) {
	rect(s, 0, 0, WinW, HUD, C_HUD_BG)

	alive := 0
	for _, e := range g.enemies {
		if e.alive {
			alive++
		}
	}

	hud := fmt.Sprintf("♥%d  💣%d/%d  🔥%d  👾%d  Score:%d  [R]Restart",
		g.player.lives, g.player.active, g.player.bombs, g.player.radius, alive, g.player.score)
	text.Draw(s, hud, basicfont.Face7x13, 8, 25, C_WHITE)
	text.Draw(s, fmt.Sprintf("Lv.%d", g.level), basicfont.Face7x13, WinW-40, 25, C_GREEN)

	// Индикаторы бонусов
	x := 200
	if g.player.speedBoost > 0 {
		text.Draw(s, "⚡", basicfont.Face7x13, x, 25, C_PU_SPEED)
		x += 15
	}
	if g.player.shield {
		text.Draw(s, "🛡", basicfont.Face7x13, x, 25, C_PU_SHIELD)
		x += 15
	}
	if g.player.kick {
		text.Draw(s, "👟", basicfont.Face7x13, x, 25, C_PU_KICK)
		x += 15
	}
	if g.player.combos > 1 {
		text.Draw(s, fmt.Sprintf("x%d COMBO!", g.player.combos), basicfont.Face7x13, x, 25, C_YELLOW)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	s.Fill(color.RGBA{10, 10, 30, 255})

	title := "BOMBERMAN GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, WinW/2-bw.Dx()/2, WinH/2-100, C_WHITE)

	frame := frameCount / 30
	c := C_GREEN
	if frame%2 == 0 {
		c = C_YELLOW
	}
	text.Draw(s, "Press ENTER or SPACE", basicfont.Face7x13, WinW/2-90, WinH/2-50, c)

	text.Draw(s, "WASD / Arrows - Move", basicfont.Face7x13, WinW/2-80, WinH/2, C_WHITE)
	text.Draw(s, "SPACE - Place Bomb", basicfont.Face7x13, WinW/2-70, WinH/2+20, C_WHITE)
	text.Draw(s, "R - Restart Level", basicfont.Face7x13, WinW/2-65, WinH/2+40, C_WHITE)

	// Типы врагов
	text.Draw(s, "Enemies:", basicfont.Face7x13, WinW/2-50, WinH/2+80, C_GREEN)
	text.Draw(s, "● Balloon  ■ Chaser  ◆ Splitter  ✦ Teleporter", basicfont.Face7x13, WinW/2-160, WinH/2+100, C_WHITE)

	// Power-ups
	text.Draw(s, "Power-ups under bricks:", basicfont.Face7x13, WinW/2-110, WinH/2+130, C_YELLOW)
	text.Draw(s, "🔥Fire  💣Bomb  ⚡Speed  ❤Life  🛡Shield  👟Kick", basicfont.Face7x13, WinW/2-180, WinH/2+150, C_WHITE)

	// Анимированные бомбы
	t := frameCount
	for i := 0; i < 5; i++ {
		bx := WinW/2 - 80 + i*40
		by := WinH/2 - 150 + int(t/20+int64(i)*10)%10
		rect(s, bx, by, 20, 20, C_BOMB)
		rect(s, bx+4, by+4, 12, 12, C_BOMB_H)
	}

	text.Draw(s, "Go365 Challenge — Day 95", basicfont.Face7x13, WinW/2-95, WinH-30, color.RGBA{150, 150, 150, 255})
}

func (g *Game) drawOverlay(s *ebiten.Image, msg string, clr color.Color) {
	rect(s, 0, 0, WinW, WinH, color.RGBA{0, 0, 0, 180})

	bw := text.BoundString(basicfont.Face7x13, msg)
	text.Draw(s, msg, basicfont.Face7x13, WinW/2-bw.Dx()/2, WinH/2-20, clr)

	sub := "Press ENTER to continue"
	bw2 := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, WinW/2-bw2.Dx()/2, WinH/2+30, C_WHITE)

	text.Draw(s, fmt.Sprintf("Final Score: %d", g.player.score), basicfont.Face7x13, WinW/2-60, WinH/2+60, C_YELLOW)
}

func (g *Game) Layout(w, h int) (int, int) {
	return WinW, WinH
}

// ======================== УТИЛИТЫ ========================
func rect(s *ebiten.Image, x, y, w, h int, c color.Color) {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	s.DrawImage(img, op)
}

func rectAlpha(s *ebiten.Image, x, y, w, h int, c color.Color, a float64) {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorM.Scale(1, 1, 1, a)
	s.DrawImage(img, op)
}

func drawTile(spr *ebiten.Image, dst *ebiten.Image, px, py int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(Tile)/float64(spr.Bounds().Dx()), float64(Tile)/float64(spr.Bounds().Dy()))
	op.GeoM.Translate(float64(px), float64(py))
	dst.DrawImage(spr, op)
}

// ======================== ЗВУК ========================
var (
	audioCtx   *audio.Context
	soundPlace *audio.Player
	soundExpl  *audio.Player
	soundKill  *audio.Player
	soundDie   *audio.Player
	soundWin   *audio.Player
	soundMenu  *audio.Player
)

func generateWAV(samples []float64, sampleRate int) []byte {
	buf := new(bytes.Buffer)
	numCh := uint16(1)
	bitsPS := uint16(16)
	byteRate := sampleRate * int(numCh) * int(bitsPS) / 8
	blockAlign := numCh * bitsPS / 8
	dataSize := len(samples) * 2

	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, numCh)
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(buf, binary.LittleEndian, blockAlign)
	binary.Write(buf, binary.LittleEndian, bitsPS)
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))

	for _, s := range samples {
		if s > 1 { s = 1 }
		if s < -1 { s = -1 }
		binary.Write(buf, binary.LittleEndian, int16(s*32767))
	}
	return buf.Bytes()
}

func makeBeep(duration float64, freq float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = math.Sin(2*math.Pi*freq*t) * math.Exp(-t*8) * 0.5
	}
	return out
}

func makeNoise(duration float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		out[i] = (rand.Float64()*2 - 1) * math.Exp(-t*6) * 0.6
	}
	return out
}

func makeSweep(duration float64, f1, f2 float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		p := float64(i) / float64(n)
		freq := f1 + (f2-f1)*p
		out[i] = math.Sin(2*math.Pi*freq*t) * math.Exp(-t*4) * 0.5
	}
	return out
}

func makeArp(duration float64, freqs []float64, sr int) []float64 {
	n := int(float64(sr) * duration)
	out := make([]float64, n)
	stepDur := duration / float64(len(freqs))
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sr)
		idx := int(t / stepDur)
		if idx >= len(freqs) {
			idx = len(freqs) - 1
		}
		out[i] = math.Sin(2*math.Pi*freqs[idx]*t) * math.Exp(-t*5) * 0.4
	}
	return out
}

func initAudio() {
	audioCtx = audio.NewContext(44100)

	wavPlace := generateWAV(makeBeep(0.08, 880, 44100), 44100)
	soundPlace, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavPlace))

	wavExpl := generateWAV(makeNoise(0.4, 44100), 44100)
	soundExpl, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavExpl))

	wavKill := generateWAV(makeSweep(0.2, 600, 200, 44100), 44100)
	soundKill, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavKill))

	wavDie := generateWAV(makeSweep(0.5, 400, 80, 44100), 44100)
	soundDie, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavDie))

	wavWin := generateWAV(makeArp(0.5, []float64{523, 659, 784, 1047}, 44100), 44100)
	soundWin, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavWin))

	wavMenu := generateWAV(makeBeep(0.05, 660, 44100), 44100)
	soundMenu, _ = audio.NewPlayer(audioCtx, bytes.NewReader(wavMenu))
}

func playSound(player *audio.Player) {
	if player == nil {
		return
	}
	player.Rewind()
	player.Play()
}

// ======================== ГЛОБАЛЬНЫЕ ========================
var frameCount int64

// ======================== MAIN ========================
func main() {
	fmt.Println("═══════════════════════════════════")
	fmt.Println("  BOMBERMAN GO — Go365 Day 95")
	fmt.Println("  4 Enemy Types | 6 Power-Ups")
	fmt.Println("═══════════════════════════════════")
	fmt.Println()

	ebiten.SetWindowSize(WinW, WinH)
	ebiten.SetWindowTitle("Bomberman Go — Go365 Day 95")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

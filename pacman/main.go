package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

type Vec struct{ X, Y int }

type Player struct {
	Pos Vec
	Dir Vec
}

type Ghost struct {
	Pos   Vec
	Dir   Vec
	Color termbox.Attribute
}

type Game struct {
	W, H         int
	Layout       [][]rune
	Player       Player
	Ghosts       []Ghost
	Coins        map[Vec]bool
	Score        int
	Running      bool
	Paused       bool
	InMenu       bool
	GameOver     bool
	Win          bool
	TickInterval time.Duration
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	g := NewGame()
	g.Run()
}

/* ------------------ Game setup ------------------ */

func NewGame() *Game {
	w, h := termbox.Size()
	g := &Game{
		W:            w,
		H:            h,
		TickInterval: 100 * time.Millisecond,
		InMenu:       true,
	}
	g.makeLayout()
	g.placeCoins()
	g.placePlayer()
	g.placeGhosts()
	return g
}

// Simple geometric maze generator (hand-crafted pattern scaled to terminal)
func (g *Game) makeLayout() {
	W, H := g.W, g.H
	grid := make([][]rune, H)
	for y := 0; y < H; y++ {
		grid[y] = make([]rune, W)
		for x := 0; x < W; x++ {
			grid[y][x] = ' '
		}
	}
	// border
	for x := 0; x < W; x++ {
		grid[0][x] = '#'
		grid[H-1][x] = '#'
	}
	for y := 0; y < H; y++ {
		grid[y][0] = '#'
		grid[y][W-1] = '#'
	}
	// geometric patterns: concentric rectangles and cross lines
	padding := 2
	for layer := 0; layer < 5; layer++ {
		x0 := padding + layer*4
		y0 := padding + layer*2
		x1 := W - 1 - x0
		y1 := H - 1 - y0
		if x1-x0 < 6 || y1-y0 < 4 {
			break
		}
		for x := x0; x <= x1; x++ {
			grid[y0][x] = '#'
			grid[y1][x] = '#'
		}
		for y := y0; y <= y1; y++ {
			grid[y][x0] = '#'
			grid[y][x1] = '#'
		}
		// add some internal walls / corridors
		for i := 0; i < 3; i++ {
			cx := (x0 + x1) / 2
			cy := y0 + 1 + i*(y1-y0-2)/2
			for x := cx - (layer+1)*2; x <= cx+(layer+1)*2; x++ {
				if x > x0 && x < x1 {
					grid[cy][x] = '#'
				}
			}
		}
	}
	// remove some openings to make paths
	for y := 1; y < H-1; y += 4 {
		for x := 2; x < W-2; x += 7 {
			grid[y][x] = ' '
		}
	}
	g.Layout = grid
}

func (g *Game) placeCoins() {
	g.Coins = make(map[Vec]bool)
	for y := 1; y < g.H-1; y++ {
		for x := 1; x < g.W-1; x++ {
			if g.Layout[y][x] == ' ' {
				// skip some to create corridors and space for player/ghosts
				if rand.Float64() < 0.7 {
					g.Coins[Vec{x, y}] = true
				}
			}
		}
	}
}

func (g *Game) placePlayer() {
	// place at center-ish
	cx, cy := g.W/4, g.H/2
	// find nearest empty
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			x := cx + dx
			y := cy + dy
			if inside(g, x, y) && g.Layout[y][x] == ' ' {
				g.Player = Player{Pos: Vec{x, y}, Dir: Vec{0, 0}}
				delete(g.Coins, Vec{x, y})
				return
			}
		}
	}
	g.Player = Player{Pos: Vec{cx, cy}}
	delete(g.Coins, Vec{cx, cy})
}

func (g *Game) placeGhosts() {
	colors := []termbox.Attribute{termbox.ColorRed, termbox.ColorMagenta, termbox.ColorCyan, termbox.ColorYellow}
	g.Ghosts = nil
	positions := []Vec{
		{g.W - g.W/4, g.H / 3},
		{g.W - g.W/4, g.H / 2},
		{g.W - g.W/4, g.H - g.H/3},
		{g.W/2 + 2, g.H / 2},
	}
	for i, p := range positions {
		// find free
		found := false
		for dy := -2; dy <= 2 && !found; dy++ {
			for dx := -2; dx <= 2 && !found; dx++ {
				x := p.X + dx
				y := p.Y + dy
				if inside(g, x, y) && g.Layout[y][x] == ' ' {
					g.Ghosts = append(g.Ghosts, Ghost{Pos: Vec{x, y}, Dir: Vec{0, 0}, Color: colors[i%len(colors)]})
					delete(g.Coins, Vec{x, y})
					found = true
				}
			}
		}
	}
}

/* ------------------ Game loop ------------------ */

func (g *Game) Run() {
	eventCh := make(chan termbox.Event)
	go func() {
		for {
			eventCh <- termbox.PollEvent()
		}
	}()

	ticker := time.NewTicker(g.TickInterval)
	defer ticker.Stop()

	for {
		g.draw()
		select {
		case ev := <-eventCh:
			if ev.Type == termbox.EventKey {
				if g.InMenu {
					g.handleMenuKey(ev)
				} else {
					g.handleKey(ev)
				}
			}
			if ev.Type == termbox.EventResize {
				g.W, g.H = ev.Width, ev.Height
				g.makeLayout()
				g.placeCoins()
				g.placePlayer()
				g.placeGhosts()
			}
		case <-ticker.C:
			if g.InMenu {
				// nothing
			} else if !g.Paused && !g.GameOver && !g.Win {
				g.update()
			}
		}
		if g.GameOver || g.Win {
			// still allow quitting or restart
		}
	}
}

/* ------------------ Input handling ------------------ */

func (g *Game) handleMenuKey(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyArrowUp, termbox.KeyArrowLeft:
		// no options to navigate for now
	case termbox.KeyEnter:
		g.InMenu = false
		g.Running = true
		g.Paused = false
	default:
		// also allow 'q' to quit
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			termbox.Close()
			fmt.Println("Bye")
			return
		}
		if ev.Ch == 's' || ev.Ch == 'S' {
			g.InMenu = false
			g.Running = true
		}
	}
}

func (g *Game) handleKey(ev termbox.Event) {
	if ev.Key == termbox.KeyCtrlC {
		termbox.Close()
		fmt.Println("Bye")
		return
	}
	switch ev.Ch {
	case 'p', 'P':
		g.Paused = !g.Paused
	case 'r', 'R':
		// restart
		*g = *NewGame()
		g.InMenu = false
		return
	case 'q', 'Q':
		termbox.Close()
		fmt.Println("Bye")
		return
	}
	switch ev.Key {
	case termbox.KeyArrowUp:
		g.Player.Dir = Vec{0, -1}
	case termbox.KeyArrowDown:
		g.Player.Dir = Vec{0, 1}
	case termbox.KeyArrowLeft:
		g.Player.Dir = Vec{-1, 0}
	case termbox.KeyArrowRight:
		g.Player.Dir = Vec{1, 0}
	}
	// WASD
	if ev.Ch == 'w' || ev.Ch == 'W' {
		g.Player.Dir = Vec{0, -1}
	}
	if ev.Ch == 's' || ev.Ch == 'S' {
		g.Player.Dir = Vec{0, 1}
	}
	if ev.Ch == 'a' || ev.Ch == 'A' {
		g.Player.Dir = Vec{-1, 0}
	}
	if ev.Ch == 'd' || ev.Ch == 'D' {
		g.Player.Dir = Vec{1, 0}
	}
}

/* ------------------ Update ------------------ */

func (g *Game) update() {
	// move player if possible
	nx := g.Player.Pos.X + g.Player.Dir.X
	ny := g.Player.Pos.Y + g.Player.Dir.Y
	if inside(g, nx, ny) && g.Layout[ny][nx] == ' ' {
		g.Player.Pos = Vec{nx, ny}
		// collect coin
		if g.Coins[g.Player.Pos] {
			delete(g.Coins, g.Player.Pos)
			g.Score += 10
			if len(g.Coins) == 0 {
				g.Win = true
			}
		}
	} else {
		// blocked -> stop moving
		g.Player.Dir = Vec{0, 0}
	}
	// move ghosts (simple random + tendency to move towards player)
	for i := range g.Ghosts {
		ghost := &g.Ghosts[i]
		// occasionally change direction
		if rand.Float64() < 0.3 || !canMove(g, ghost.Pos, ghost.Dir) {
			ghost.Dir = chooseGhostDir(g, ghost.Pos, g.Player.Pos)
		}
		nx := ghost.Pos.X + ghost.Dir.X
		ny := ghost.Pos.Y + ghost.Dir.Y
		if inside(g, nx, ny) && g.Layout[ny][nx] == ' ' {
			ghost.Pos = Vec{nx, ny}
		} else {
			ghost.Dir = chooseGhostDir(g, ghost.Pos, g.Player.Pos)
		}
		// collision
		if ghost.Pos == g.Player.Pos {
			g.GameOver = true
		}
	}
}

/* ------------------ Helpers ------------------ */

func inside(g *Game, x, y int) bool {
	return x >= 0 && x < g.W && y >= 0 && y < g.H
}

func canMove(g *Game, pos Vec, dir Vec) bool {
	nx := pos.X + dir.X
	ny := pos.Y + dir.Y
	return inside(g, nx, ny) && g.Layout[ny][nx] == ' '
}

func chooseGhostDir(g *Game, pos Vec, target Vec) Vec {
	// preferences: move towards player with some randomness
	best := Vec{0, 0}
	bestScore := -9999.0
	cands := []Vec{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {0, 0}}
	for _, d := range cands {
		nx := pos.X + d.X
		ny := pos.Y + d.Y
		if !inside(g, nx, ny) || g.Layout[ny][nx] != ' ' {
			continue
		}
		// score: negative distance to target plus random small factor
		dist := float64((nx-target.X)*(nx-target.X) + (ny-target.Y)*(ny-target.Y))
		score := -dist + rand.Float64()*20.0
		if score > bestScore {
			bestScore = score
			best = d
		}
	}
	return best
}

/* ------------------ Rendering ------------------ */

func (g *Game) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if g.InMenu {
		g.drawMenu()
		termbox.Flush()
		return
	}
	// draw layout
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			ch := g.Layout[y][x]
			fg, bg := termbox.ColorWhite, termbox.ColorBlack
			if ch == '#' {
				drawAt(x, y, '#', termbox.ColorBlue|termbox.AttrBold, bg)
			} else {
				// empty space
				drawAt(x, y, ' ', fg, bg)
			}
		}
	}
	// draw coins
	for pos := range g.Coins {
		drawAt(pos.X, pos.Y, '.', termbox.ColorYellow, termbox.ColorBlack)
	}
	// draw player (geometric Pacman)
	drawPacman(g.Player.Pos.X, g.Player.Pos.Y, termbox.ColorGreen)

	// ghosts
	for _, ghost := range g.Ghosts {
		drawGhost(ghost.Pos.X, ghost.Pos.Y, ghost.Color)
	}

	// HUD
	printStr(1, 1, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack, fmt.Sprintf("Score: %d  Coins left: %d", g.Score, len(g.Coins)))
	printStr(1, 2, termbox.ColorWhite, termbox.ColorBlack, "P - пауза, R - рестарт, Q - выход")

	if g.Paused {
		printCentered(g.H/2, "PAUSED — нажмите P для продолжения")
	}
	if g.GameOver {
		printCentered(g.H/2, "GAME OVER — нажмите R чтобы перезапустить или Q чтобы выйти")
	}
	if g.Win {
		printCentered(g.H/2, "YOU WIN! — нажмите R чтобы сыграть снова")
	}

	termbox.Flush()
}

func (g *Game) drawMenu() {
	title := "PACMAN (терминал)"
	sub := "Нажмите Enter или S чтобы начать. P — пауза во время игры. Q — выход."
	h := g.H
	printCentered(h/2-2, title)
	printCentered(h/2, sub)
	printCentered(h/2+2, "Управление: стрелки или WASD")
	printCentered(h/2+4, "Красивая геометрия лабиринта и персонажи")
	printCentered(h-2, "Версия: терминальная demo")
}

func printCentered(y int, s string) {
	w, _ := termbox.Size()
	x := (w - len(s)) / 2
	printStr(x, y, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack, s)
}

func printStr(x, y int, fg, bg termbox.Attribute, s string) {
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

func drawAt(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x, y, ch, fg, bg)
}

func drawPacman(x, y int, color termbox.Attribute) {
	// geometric Pacman using 3x3 pattern
	// mouth direction determined by any non-zero dir? keep facing right for simplicity
	pattern := []string{
		" ◉ ",
		"◐◉◑",
		" ◉ ",
	}
	for dy, row := range pattern {
		for dx, c := range row {
			if c == ' ' {
				continue
			}
			termbox.SetCell(x+dx-1, y+dy-1, c, color|termbox.AttrBold, termbox.ColorBlack)
		}
	}
}

func drawGhost(x, y int, color termbox.Attribute) {
	// simple geometric ghost 3x3
	pattern := []string{
		"╭╮ ",
		"███",
		"⎺⎺⎺",
	}
	for dy, row := range pattern {
		for dx, c := range row {
			if c == ' ' {
				continue
			}
			termbox.SetCell(x+dx-1, y+dy-1, c, color, termbox.ColorBlack)
		}
	}
}

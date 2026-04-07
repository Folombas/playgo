// Food Match-3 - Go365 Challenge Day 99
package main

import (
	"bytes"
	"embed"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed assets/food/*.png
var foodFS embed.FS

const (
	screenWidth  = 1024
	screenHeight = 768
	boardCols    = 8
	boardRows    = 8
	cellSize     = 60
	cellPad      = 6
	boardX       = 300
	boardY       = 80
	foodCount    = 6
)

var foodNames = []string{"apple", "orange", "banana", "strawberry", "grapes", "cupcake"}

var foodColors = []color.RGBA{
	{255, 60, 60, 255},
	{255, 170, 0, 255},
	{255, 255, 80, 255},
	{255, 100, 150, 255},
	{160, 100, 255, 255},
	{255, 180, 220, 255},
}

type State int

const (
	StateIdle State = iota
	StateSwapping
	StateRemoving
	StateDropping
)

type Piece struct {
	Type    int
	Col     int
	Row     int
	X       float64
	Y       float64
	TargetX float64
	TargetY float64
	Alpha   float64
	Scale   float64
	Matched bool
}

func newPiece(t, c, r int, fromAbove bool) *Piece {
	x := float64(boardX + c*(cellSize+cellPad))
	y := float64(boardY + r*(cellSize+cellPad))
	if fromAbove {
		y = float64(boardY - (cellSize+cellPad)*4)
	}
	return &Piece{Type: t, Col: c, Row: r, X: x, Y: y, TargetX: x, TargetY: y, Alpha: 1, Scale: 1}
}

func (p *Piece) update() bool {
	dx := p.TargetX - p.X
	dy := p.TargetY - p.Y
	if dx*dx+dy*dy > 0.25 {
		p.X += dx * 0.25
		p.Y += dy * 0.25
		return true
	}
	p.X = p.TargetX
	p.Y = p.TargetY
	if p.Matched {
		p.Scale *= 0.8
		p.Alpha *= 0.75
		return p.Alpha > 0.02
	}
	return false
}

func (p *Piece) contains(mx, my float64) bool {
	s := float64(cellSize) * p.Scale / 2
	cx := p.X + float64(cellSize)/2
	cy := p.Y + float64(cellSize)/2
	return mx >= cx-s && mx <= cx+s && my >= cy-s && my <= cy+s
}

type Board struct {
	g [][]*Piece
}

func newBoard() *Board {
	b := &Board{g: make([][]*Piece, boardCols)}
	for c := 0; c < boardCols; c++ {
		b.g[c] = make([]*Piece, boardRows)
	}
	b.fill()
	return b
}

func (b *Board) fill() {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			var t int
			for {
				t = rand.Intn(foodCount)
				if !b.wouldMatch(c, r, t) {
					break
				}
			}
			b.g[c][r] = newPiece(t, c, r, false)
		}
	}
}

func (b *Board) wouldMatch(c, r, t int) bool {
	n := 1
	for i := c - 1; i >= 0 && b.g[i][r] != nil && b.g[i][r].Type == t; i-- {
		n++
	}
	for i := c + 1; i < boardCols && b.g[i][r] != nil && b.g[i][r].Type == t; i++ {
		n++
	}
	if n >= 3 {
		return true
	}
	n = 1
	for i := r - 1; i >= 0 && b.g[c][i] != nil && b.g[c][i].Type == t; i-- {
		n++
	}
	for i := r + 1; i < boardRows && b.g[c][i] != nil && b.g[c][i].Type == t; i++ {
		n++
	}
	return n >= 3
}

func (b *Board) get(c, r int) *Piece {
	if c < 0 || c >= boardCols || r < 0 || r >= boardRows {
		return nil
	}
	return b.g[c][r]
}

func (b *Board) swap(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if dc < 0 {
		dc = -dc
	}
	if dr < 0 {
		dr = -dr
	}
	if !((dc == 1 && dr == 0) || (dc == 0 && dr == 1)) {
		return false
	}
	p1 := b.g[c1][r1]
	p2 := b.g[c2][r2]
	if p1 == nil || p2 == nil {
		return false
	}
	b.g[c1][r1] = p2
	b.g[c2][r2] = p1
	p1.Col, p1.Row = c2, r2
	p2.Col, p2.Row = c1, r1
	p1.TargetX = float64(boardX + c2*(cellSize+cellPad))
	p1.TargetY = float64(boardY + r2*(cellSize+cellPad))
	p2.TargetX = float64(boardX + c1*(cellSize+cellPad))
	p2.TargetY = float64(boardY + r1*(cellSize+cellPad))
	return true
}

func (b *Board) findMatches() int {
	matched := make(map[[2]int]bool)
	count := 0

	for r := 0; r < boardRows; r++ {
		for c := 0; c <= boardCols-3; c++ {
			p := b.g[c][r]
			if p == nil || p.Matched {
				continue
			}
			match := []*Piece{p}
			for cc := c + 1; cc < boardCols && b.g[cc][r] != nil && b.g[cc][r].Type == p.Type; cc++ {
				match = append(match, b.g[cc][r])
			}
			if len(match) >= 3 {
				for _, m := range match {
					k := [2]int{m.Col, m.Row}
					if !matched[k] {
						matched[k] = true
						m.Matched = true
						count++
					}
				}
			}
		}
	}

	for c := 0; c < boardCols; c++ {
		for r := 0; r <= boardRows-3; r++ {
			p := b.g[c][r]
			if p == nil || p.Matched {
				continue
			}
			match := []*Piece{p}
			for rr := r + 1; rr < boardRows && b.g[c][rr] != nil && b.g[c][rr].Type == p.Type; rr++ {
				match = append(match, b.g[c][rr])
			}
			if len(match) >= 3 {
				for _, m := range match {
					k := [2]int{m.Col, m.Row}
					if !matched[k] {
						matched[k] = true
						m.Matched = true
						count++
					}
				}
			}
		}
	}
	return count
}

func (b *Board) remove() {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			if b.g[c][r] != nil && b.g[c][r].Matched {
				b.g[c][r] = nil
			}
		}
	}
}

func (b *Board) drop() bool {
	dropped := false
	for c := 0; c < boardCols; c++ {
		empty := -1
		for r := boardRows - 1; r >= 0; r-- {
			if b.g[c][r] == nil {
				if empty == -1 {
					empty = r
				}
				continue
			}
			if empty != -1 {
				p := b.g[c][r]
				b.g[c][empty] = p
				b.g[c][r] = nil
				p.Row = empty
				p.TargetY = float64(boardY + empty*(cellSize+cellPad))
				empty--
				dropped = true
			}
		}
		for r := empty; r >= 0; r-- {
			t := rand.Intn(foodCount)
			p := newPiece(t, c, r, true)
			p.TargetY = float64(boardY + r*(cellSize+cellPad))
			b.g[c][r] = p
			dropped = true
		}
	}
	return dropped
}

func (b *Board) isAnimating() bool {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := b.g[c][r]
			if p == nil {
				continue
			}
			dx := p.TargetX - p.X
			dy := p.TargetY - p.Y
			if dx*dx+dy*dy > 0.5 {
				return true
			}
			if p.Matched && p.Alpha > 0.02 {
				return true
			}
		}
	}
	return false
}

func (b *Board) at(mx, my float64) *Piece {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := b.g[c][r]
			if p != nil && p.contains(mx, my) {
				return p
			}
		}
	}
	return nil
}

type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.RGBA
	Size         float64
}

type Game struct {
	board      *Board
	foods      []*ebiten.Image
	particles  []*Particle
	state      State
	score      int
	combo      int
	drag       *Piece
	dragOX     float64
	dragOY     float64
	dragSX     float64
	dragSY     float64
	isDragging bool
	lastC1     int
	lastR1     int
	lastC2     int
	lastR2     int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		board:     newBoard(),
		foods:     make([]*ebiten.Image, foodCount),
		particles: []*Particle{},
	}
	g.loadSprites()
	return g
}

func (g *Game) loadSprites() {
	for i, name := range foodNames {
		data, err := foodFS.ReadFile("assets/food/" + name + ".png")
		if err != nil {
			continue
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			continue
		}
		g.foods[i] = ebiten.NewImageFromImage(img)
	}
}

func (g *Game) Update() error {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.2
		p.Life -= 0.02
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := g.board.get(c, r)
			if p != nil {
				p.update()
			}
		}
	}

	switch g.state {
	case StateIdle:
		g.input()
	case StateSwapping:
		if !g.board.isAnimating() {
			m := g.board.findMatches()
			if m > 0 {
				g.combo++
				g.state = StateRemoving
			} else {
				g.board.swap(g.lastC1, g.lastR1, g.lastC2, g.lastR2)
				g.combo = 0
				g.state = StateIdle
			}
			g.drag = nil
			g.isDragging = false
		}
	case StateRemoving:
		if !g.board.isAnimating() {
			g.board.remove()
			if g.board.drop() {
				g.state = StateDropping
			} else {
				m := g.board.findMatches()
				if m > 0 {
					g.combo++
					g.state = StateRemoving
				} else {
					g.combo = 0
					g.state = StateIdle
				}
			}
		}
	case StateDropping:
		if !g.board.isAnimating() {
			m := g.board.findMatches()
			if m > 0 {
				g.combo++
				g.state = StateRemoving
			} else {
				g.combo = 0
				g.state = StateIdle
			}
		}
	}
	return nil
}

func (g *Game) input() {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		p := g.board.at(fx, fy)
		if p != nil {
			g.drag = p
			g.dragOX = fx
			g.dragOY = fy
			g.dragSX = p.X
			g.dragSY = p.Y
			g.isDragging = false
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.drag != nil {
		dx := fx - g.dragOX
		dy := fy - g.dragOY
		if !g.isDragging && (dx*dx+dy*dy > 200) {
			g.isDragging = true
		}
		if g.isDragging {
			g.drag.X = g.dragSX + dx
			g.drag.Y = g.dragSY + dy
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && g.drag != nil {
		if g.isDragging {
			dx := g.drag.X - g.dragSX
			dy := g.drag.Y - g.dragSY
			tc, tr := g.drag.Col, g.drag.Row
			if dx*dx > dy*dy {
				if dx > 0 {
					tc++
				} else {
					tc--
				}
			} else {
				if dy > 0 {
					tr++
				} else {
					tr--
				}
			}
			if tc >= 0 && tc < boardCols && tr >= 0 && tr < boardRows {
				target := g.board.get(tc, tr)
				if target != nil {
					g.lastC1 = g.drag.Col
					g.lastR1 = g.drag.Row
					g.lastC2 = tc
					g.lastR2 = tr
					g.board.swap(g.drag.Col, g.drag.Row, tc, tr)
					g.state = StateSwapping
				}
			}
			g.drag.X = g.dragSX
			g.drag.Y = g.dragSY
		}
		g.drag = nil
		g.isDragging = false
	}
}

func (g *Game) spawnParticles(count int) {
	for i := 0; i < count*6; i++ {
		c := rand.Intn(boardCols)
		r := rand.Intn(boardRows)
		p := g.board.get(c, r)
		if p != nil {
			col := foodColors[p.Type]
			g.particles = append(g.particles, &Particle{
				X:     float64(boardX+c*(cellSize+cellPad)+cellSize/2),
				Y:     float64(boardY+r*(cellSize+cellPad)+cellSize/2),
				VX:    float64(rand.Intn(7)-3) * 1.5,
				VY:    float64(rand.Intn(7)-3) * 1.5 - 2,
				Life:  1,
				Color: col,
				Size:  float64(3 + rand.Intn(4)),
			})
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})

	// Фон доски
	bw := boardCols*(cellSize+cellPad) + 16
	bh := boardRows*(cellSize+cellPad) + 16
	bg := ebiten.NewImage(bw, bh)
	bg.Fill(color.RGBA{30, 30, 50, 250})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(boardX-8), float64(boardY-8))
	screen.DrawImage(bg, op)

	// Рамка
	border := ebiten.NewImage(bw, 3)
	border.Fill(color.RGBA{80, 160, 255, 255})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(boardX-8), float64(boardY-10))
	screen.DrawImage(border, op2)

	// Сетка
	for c := 0; c <= boardCols; c++ {
		x := float64(boardX + c*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, x, float64(boardY), x, float64(boardY+boardRows*(cellSize+cellPad)),
			color.RGBA{50, 50, 80, 120})
	}
	for r := 0; r <= boardRows; r++ {
		y := float64(boardY + r*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, float64(boardX), y, float64(boardX+boardCols*(cellSize+cellPad)), y,
			color.RGBA{50, 50, 80, 120})
	}

	// Еда
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			p := g.board.get(c, r)
			if p == nil || p.Alpha < 0.02 || p == g.drag {
				continue
			}
			g.drawPiece(screen, p)
		}
	}
	if g.drag != nil && g.drag.Alpha > 0.02 {
		g.drawPiece(screen, g.drag)
	}

	// Частицы
	for _, p := range g.particles {
		s := int(p.Size * p.Life)
		if s < 1 {
			continue
		}
		img := ebiten.NewImage(s, s)
		img.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, uint8(p.Life * 255)})
		opp := &ebiten.DrawImageOptions{}
		opp.GeoM.Translate(p.X-float64(s)/2, p.Y-float64(s)/2)
		screen.DrawImage(img, opp)
	}

	// UI
	g.drawUI(screen)
}

func (g *Game) drawPiece(screen *ebiten.Image, p *Piece) {
	img := g.foods[p.Type]
	if img == nil {
		return
	}
	s := int(float64(cellSize) * p.Scale)
	if s < 4 {
		return
	}
	sc := float64(s) / float64(img.Bounds().Dx())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sc, sc)
	op.GeoM.Translate(p.X, p.Y)
	op.ColorM.Scale(1, 1, 1, p.Alpha)
	screen.DrawImage(img, op)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Панель
	panel := ebiten.NewImage(240, 260)
	panel.Fill(color.RGBA{25, 25, 45, 230})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(20, 50)
	screen.DrawImage(panel, op)

	// Рамка
	top := ebiten.NewImage(240, 3)
	top.Fill(color.RGBA{80, 160, 255, 255})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(20, 50)
	screen.DrawImage(top, op2)

	// Очки
	sw := 180
	if g.score > 10000 {
		sw = 220
	}
	sp := ebiten.NewImage(sw, 30)
	sp.Fill(color.RGBA{40, 100, 140, 200})
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(40, 100)
	screen.DrawImage(sp, op3)

	// Комбо
	if g.combo > 1 {
		cw := 100 + g.combo*20
		if cw > 220 {
			cw = 220
		}
		cp := ebiten.NewImage(cw, 30)
		cp.Fill(color.RGBA{180, 150, 50, 200})
		op4 := &ebiten.DrawImageOptions{}
		op4.GeoM.Translate(40, 150)
		screen.DrawImage(cp, op4)
	}

	// Уровень
	lv := g.score/1000 + 1
	lw := 60 + lv*30
	if lw > 200 {
		lw = 200
	}
	lp := ebiten.NewImage(lw, 25)
	lp.Fill(color.RGBA{50, 100, 80, 200})
	op5 := &ebiten.DrawImageOptions{}
	op5.GeoM.Translate(40, 200)
	screen.DrawImage(lp, op5)

	// Подсказка
	hp := ebiten.NewImage(450, 30)
	hp.Fill(color.RGBA{25, 25, 45, 180})
	op6 := &ebiten.DrawImageOptions{}
	op6.GeoM.Translate(280, 715)
	screen.DrawImage(hp, op6)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return screenWidth, screenHeight
}

func itos(n int) string {
	if n == 0 {
		return "0"
	}
	var s string
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Food Match-3 - Go365")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ==================== КОНСТАНТЫ ====================
const (
	screenWidth  = 1024
	screenHeight = 768
	
	boardCols = 8
	boardRows = 8
	cellSize  = 64
	cellPad   = 4
	
	boardX = 280
	boardY = 60
	
	gemCount = 6
)

// ==================== ЦВЕТА ====================
var gemColors = []color.RGBA{
	{255, 50, 50, 255},   // красный
	{50, 100, 255, 255},  // синий
	{50, 200, 50, 255},   // зелёный
	{255, 255, 50, 255},  // жёлтый
	{200, 50, 255, 255},  // фиолетовый
	{255, 165, 0, 255},   // оранжевый
}

var gemBorderColors = []color.RGBA{
	{180, 30, 30, 255},
	{30, 60, 180, 255},
	{30, 120, 30, 255},
	{180, 180, 30, 255},
	{120, 30, 180, 255},
	{180, 100, 0, 255},
}

// ==================== СОСТОЯНИЯ ====================
type State int
const (
	StateIdle State = iota
	StateSwapping
	StateChecking
	StateRemoving
	StateDropping
	StateGameOver
)

// ==================== КРИСТАЛЛ ====================
type Gem struct {
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

func NewGem(gType, col, row int, aboveBoard bool) *Gem {
	x := float64(boardX + col*(cellSize+cellPad))
	y := float64(boardY + row*(cellSize+cellPad))
	
	if aboveBoard {
		y = float64(boardY - (cellSize+cellPad)*(3))
	}
	
	return &Gem{
		Type:    gType,
		Col:     col,
		Row:     row,
		X:       x,
		Y:       y,
		TargetX: x,
		TargetY: y,
		Alpha:   1.0,
		Scale:   1.0,
	}
}

func (g *Gem) Update() bool {
	// Плавное движение к цели
	dx := g.TargetX - g.X
	dy := g.TargetY - g.Y
	
	if dx*dx+dy*dy > 0.5 {
		g.X += dx * 0.2
		g.Y += dy * 0.2
		return true
	}
	
	g.X = g.TargetX
	g.Y = g.TargetY
	
	// Анимация исчезновения
	if g.Matched {
		g.Scale *= 0.85
		g.Alpha *= 0.8
		return g.Alpha > 0.01
	}
	
	return false
}

func (g *Gem) Contains(mx, my float64) bool {
	size := float64(cellSize) * g.Scale / 2
	cx := g.X + float64(cellSize)/2
	cy := g.Y + float64(cellSize)/2
	return mx >= cx-size && mx <= cx+size && my >= cy-size && my <= cy+size
}

// ==================== ИГРОВОЕ ПОЛЕ ====================
type Board struct {
	gems [][]*Gem
}

func NewBoard() *Board {
	b := &Board{
		gems: make([][]*Gem, boardCols),
	}
	
	for c := 0; c < boardCols; c++ {
		b.gems[c] = make([]*Gem, boardRows)
	}
	
	b.Fill()
	return b
}

func (b *Board) Fill() {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			var gemType int
			for {
				gemType = rand.Intn(gemCount)
				if !b.wouldMatch(c, r, gemType) {
					break
				}
			}
			b.gems[c][r] = NewGem(gemType, c, r, false)
		}
	}
}

func (b *Board) wouldMatch(col, row, gemType int) bool {
	// Горизонталь
	count := 1
	for c := col - 1; c >= 0 && b.gems[c][row] != nil && b.gems[c][row].Type == gemType; c-- {
		count++
	}
	for c := col + 1; c < boardCols && b.gems[c][row] != nil && b.gems[c][row].Type == gemType; c++ {
		count++
	}
	if count >= 3 {
		return true
	}
	
	// Вертикаль
	count = 1
	for r := row - 1; r >= 0 && b.gems[col][r] != nil && b.gems[col][r].Type == gemType; r-- {
		count++
	}
	for r := row + 1; r < boardRows && b.gems[col][r] != nil && b.gems[col][r].Type == gemType; r++ {
		count++
	}
	return count >= 3
}

func (b *Board) GetGem(col, row int) *Gem {
	if col < 0 || col >= boardCols || row < 0 || row >= boardRows {
		return nil
	}
	return b.gems[col][row]
}

func (b *Board) SetGem(col, row int, gem *Gem) {
	if col >= 0 && col < boardCols && row >= 0 && row < boardRows {
		b.gems[col][row] = gem
	}
}

func (b *Board) Swap(col1, row1, col2, row2 int) {
	if !b.isAdjacent(col1, row1, col2, row2) {
		return
	}
	
	gem1 := b.gems[col1][row1]
	gem2 := b.gems[col2][row2]
	
	b.gems[col1][row1] = gem2
	b.gems[col2][row2] = gem1
	
	gem1.Col, gem1.Row = col2, row2
	gem2.Col, gem2.Row = col1, row1
	
	gem1.TargetX = float64(boardX + col2*(cellSize+cellPad))
	gem1.TargetY = float64(boardY + row2*(cellSize+cellPad))
	
	gem2.TargetX = float64(boardX + col1*(cellSize+cellPad))
	gem2.TargetY = float64(boardY + row1*(cellSize+cellPad))
}

func (b *Board) isAdjacent(c1, r1, c2, r2 int) bool {
	dc := abs(c1 - c2)
	dr := abs(r1 - r2)
	return (dc == 1 && dr == 0) || (dc == 0 && dr == 1)
}

func (b *Board) FindMatches() [][]*Gem {
	var matches [][]*Gem
	matched := make(map[[2]int]bool)
	
	// Горизонтальные
	for r := 0; r < boardRows; r++ {
		for c := 0; c <= boardCols-3; c++ {
			gem := b.gems[c][r]
			if gem == nil || gem.Matched {
				continue
			}
			
			match := []*Gem{gem}
			for cc := c + 1; cc < boardCols && b.gems[cc][r] != nil && b.gems[cc][r].Type == gem.Type; cc++ {
				match = append(match, b.gems[cc][r])
			}
			
			if len(match) >= 3 {
				for _, g := range match {
					key := [2]int{g.Col, g.Row}
					if !matched[key] {
						matched[key] = true
						g.Matched = true
					}
				}
				matches = append(matches, match)
			}
		}
	}
	
	// Вертикальные
	for c := 0; c < boardCols; c++ {
		for r := 0; r <= boardRows-3; r++ {
			gem := b.gems[c][r]
			if gem == nil || gem.Matched {
				continue
			}
			
			match := []*Gem{gem}
			for rr := r + 1; rr < boardRows && b.gems[c][rr] != nil && b.gems[c][rr].Type == gem.Type; rr++ {
				match = append(match, b.gems[c][rr])
			}
			
			if len(match) >= 3 {
				for _, g := range match {
					key := [2]int{g.Col, g.Row}
					if !matched[key] {
						matched[key] = true
						g.Matched = true
					}
				}
				matches = append(matches, match)
			}
		}
	}
	
	return matches
}

func (b *Board) RemoveMatched() int {
	count := 0
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			if b.gems[c][r] != nil && b.gems[c][r].Matched {
				b.gems[c][r] = nil
				count++
			}
		}
	}
	return count
}

func (b *Board) Drop() bool {
	dropped := false
	
	for c := 0; c < boardCols; c++ {
		emptyRow := -1
		
		for r := boardRows - 1; r >= 0; r-- {
			if b.gems[c][r] == nil {
				if emptyRow == -1 {
					emptyRow = r
				}
				continue
			}
			
			if emptyRow != -1 {
				gem := b.gems[c][r]
				b.gems[c][emptyRow] = gem
				b.gems[c][r] = nil
				
				gem.Row = emptyRow
				gem.TargetY = float64(boardY + emptyRow*(cellSize+cellPad))
				
				emptyRow--
				dropped = true
			}
		}
		
		// Новые кристаллы сверху
		for r := emptyRow; r >= 0; r-- {
			gemType := rand.Intn(gemCount)
			gem := NewGem(gemType, c, r, true)
			gem.TargetY = float64(boardY + r*(cellSize+cellPad))
			b.gems[c][r] = gem
			dropped = true
		}
	}
	
	return dropped
}

func (b *Board) IsAnimating() bool {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			gem := b.gems[c][r]
			if gem == nil {
				continue
			}
			
			dx := gem.TargetX - gem.X
			dy := gem.TargetY - gem.Y
			if dx*dx+dy*dy > 1 {
				return true
			}
			
			if gem.Matched && gem.Alpha > 0.01 {
				return true
			}
		}
	}
	return false
}

func (b *Board) GetGemAt(mx, my float64) *Gem {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			gem := b.gems[c][r]
			if gem != nil && gem.Contains(mx, my) {
				return gem
			}
		}
	}
	return nil
}

// ==================== ЧАСТИЦЫ ====================
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
}

func NewParticle(x, y float64, col color.RGBA) *Particle {
	return &Particle{
		X:     x,
		Y:     y,
		VX:    (rand.Float64() - 0.5) * 8,
		VY:    (rand.Float64() - 0.5) * 8,
		Life:  1.0,
		Color: col,
		Size:  float64(3 + rand.Intn(5)),
	}
}

func (p *Particle) Update() bool {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.3
	p.Life -= 0.025
	return p.Life > 0
}

// ==================== ИГРА ====================
type Game struct {
	board       *Board
	particles   []*Particle
	state       State
	score       int
	combo       int
	selected    *Gem
	hintTimer   int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		board:     NewBoard(),
		state:     StateIdle,
		score:     0,
		combo:     0,
		particles: make([]*Particle, 0),
	}
	
	return g
}

func (g *Game) Update() error {
	// Обновление частиц
	for i := len(g.particles) - 1; i >= 0; i-- {
		if !g.particles[i].Update() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
	
	// Обновление кристаллов
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			gem := g.board.GetGem(c, r)
			if gem != nil {
				gem.Update()
			}
		}
	}
	
	// Машина состояий
	switch g.state {
	case StateIdle:
		g.handleInput()
		
	case StateSwapping:
		if !g.board.IsAnimating() {
			matches := g.board.FindMatches()
			if len(matches) > 0 {
				g.combo++
				g.state = StateRemoving
			} else {
				// Отмена обмена
				g.board.Swap(g.selected.Col, g.selected.Row, 
					g.selected.Col, g.selected.Row)
				g.selected = nil
				g.state = StateIdle
			}
		}
		
	case StateRemoving:
		if !g.board.IsAnimating() {
			count := g.board.RemoveMatched()
			if count > 0 {
				baseScore := count * 100
				comboBonus := int(float64(baseScore) * (1 + float64(g.combo-1)*0.5))
				g.score += comboBonus
				
				// Частицы
				g.spawnParticles(count)
			}
			g.board.Drop()
			g.state = StateDropping
		}
		
	case StateDropping:
		if !g.board.IsAnimating() {
			g.state = StateRemoving
		}
	}
	
	g.hintTimer++
	return nil
}

func (g *Game) handleInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		gem := g.board.GetGemAt(float64(mx), float64(my))
		
		if gem != nil {
			if g.selected == nil {
				g.selected = gem
			} else if g.board.isAdjacent(g.selected.Col, g.selected.Row, gem.Col, gem.Row) {
				g.board.Swap(g.selected.Col, g.selected.Row, gem.Col, gem.Row)
				g.selected = nil
				g.state = StateSwapping
			} else {
				g.selected = gem
			}
		}
	}
	
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.selected = nil
	}
}

func (g *Game) spawnParticles(count int) {
	for i := 0; i < count*8; i++ {
		c := rand.Intn(boardCols)
		r := rand.Intn(boardRows)
		gem := g.board.GetGem(c, r)
		if gem != nil {
			col := gemColors[gem.Type]
			x := float64(boardX + c*(cellSize+cellPad) + cellSize/2)
			y := float64(boardY + r*(cellSize+cellPad) + cellSize/2)
			g.particles = append(g.particles, NewParticle(x, y, col))
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Фон - градиент
	g.drawBackground(screen)
	
	// Игровое поле
	g.drawBoard(screen)
	
	// Кристаллы
	g.drawGems(screen)
	
	// Частицы
	g.drawParticles(screen)
	
	// UI
	g.drawUI(screen)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 35, 255})
	
	// Звёзды на фоне
	for i := 0; i < 50; i++ {
		x := float64((i * 137 + 50) % screenWidth)
		y := float64((i * 251 + 30) % screenHeight)
		brightness := uint8(100 + (i*31)%155)
		ebitenutil.DrawLine(screen, x, y, x, y, 
			color.RGBA{brightness, brightness, brightness, 255})
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	// Фон поля
	w := boardCols*(cellSize+cellPad) + 20
	h := boardRows*(cellSize+cellPad) + 20
	
	board := ebiten.NewImage(w, h)
	board.Fill(color.RGBA{25, 25, 50, 240})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(boardX-10), float64(boardY-10))
	screen.DrawImage(board, op)
	
	// Рамка
	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 180, 255, 255})
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(boardX-10), float64(boardY-12))
	screen.DrawImage(border, op2)
	
	// Сетка
	for c := 0; c <= boardCols; c++ {
		x := float64(boardX + c*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, x, float64(boardY), x, float64(boardY+boardRows*(cellSize+cellPad)),
			color.RGBA{60, 60, 100, 100})
	}
	for r := 0; r <= boardRows; r++ {
		y := float64(boardY + r*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, float64(boardX), y, float64(boardX+boardCols*(cellSize+cellPad)), y,
			color.RGBA{60, 60, 100, 100})
	}
}

func (g *Game) drawGems(screen *ebiten.Image) {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			gem := g.board.GetGem(c, r)
			if gem == nil || gem.Alpha < 0.01 {
				continue
			}
			
			g.drawSingleGem(screen, gem)
		}
	}
}

func (g *Game) drawSingleGem(screen *ebiten.Image, gem *Gem) {
	col := gemColors[gem.Type]
	borderCol := gemBorderColors[gem.Type]
	
	// Выделение
	if g.selected == gem {
		highlight := ebiten.NewImage(cellSize+8, cellSize+8)
		highlight.Fill(color.RGBA{255, 255, 255, 100})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(gem.X-4, gem.Y-4)
		screen.DrawImage(highlight, op)
	}
	
	// Основной кристалл
	size := int(float64(cellSize) * gem.Scale)
	if size < 4 {
		return
	}
	
	// Рамка
	gemImg := ebiten.NewImage(size, size)
	gemImg.Fill(borderCol)
	
	// Внутренняя часть
	inner := ebiten.NewImage(size-4, size-4)
	inner.Fill(col)
	
	// Блик
	shine := ebiten.NewImage(size/3, size/3)
	shine.Fill(color.RGBA{255, 255, 255, 120})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(gem.X, gem.Y)
	op.ColorM.Scale(1, 1, 1, gem.Alpha)
	screen.DrawImage(gemImg, op)
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(gem.X+2, gem.Y+2)
	op2.ColorM.Scale(1, 1, 1, gem.Alpha)
	screen.DrawImage(inner, op2)
	
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(gem.X+float64(size)/4, gem.Y+float64(size)/4)
	op3.ColorM.Scale(1, 1, 1, gem.Alpha)
	screen.DrawImage(shine, op3)
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		size := int(p.Size * p.Life)
		if size < 1 {
			continue
		}
		
		particle := ebiten.NewImage(size, size)
		alpha := uint8(p.Life * 255)
		particle.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(particle, op)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Панель счёта слева
	panel := ebiten.NewImage(240, 300)
	panel.Fill(color.RGBA{20, 20, 50, 220})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(20, 40)
	screen.DrawImage(panel, op)
	
	// Рамка панели
	topBorder := ebiten.NewImage(240, 4)
	topBorder.Fill(color.RGBA{100, 180, 255, 255})
	
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(20, 40)
	screen.DrawImage(topBorder, op2)
	
	// Заголовок - красивый баннер
	titlePanel := ebiten.NewImage(200, 40)
	titlePanel.Fill(color.RGBA{40, 40, 80, 200})
	
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(40, 70)
	screen.DrawImage(titlePanel, op3)
	
	// Рамка заголовка
	titleBorder := ebiten.NewImage(200, 3)
	titleBorder.Fill(color.RGBA{100, 200, 255, 255})
	
	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(40, 70)
	screen.DrawImage(titleBorder, op4)
	
	// Очки - зелёная панель
	scoreWidth := 180
	if g.score > 10000 {
		scoreWidth = 220
	}
	scorePanel := ebiten.NewImage(scoreWidth, 35)
	scorePanel.Fill(color.RGBA{40, 120, 40, 200})
	
	op5 := &ebiten.DrawImageOptions{}
	op5.GeoM.Translate(40, 140)
	screen.DrawImage(scorePanel, op5)
	
	// Комбо - золотая панель
	if g.combo > 1 {
		comboWidth := 120 + g.combo*20
		if comboWidth > 220 {
			comboWidth = 220
		}
		comboPanel := ebiten.NewImage(comboWidth, 30)
		comboPanel.Fill(color.RGBA{200, 180, 50, 200})
		
		op6 := &ebiten.DrawImageOptions{}
		op6.GeoM.Translate(40, 190)
		screen.DrawImage(comboPanel, op6)
	}
	
	// Уровень - синяя панель
	level := g.score/1000 + 1
	levelWidth := 60 + level*30
	if levelWidth > 200 {
		levelWidth = 200
	}
	levelPanel := ebiten.NewImage(levelWidth, 30)
	levelPanel.Fill(color.RGBA{50, 50, 150, 200})
	
	op7 := &ebiten.DrawImageOptions{}
	op7.GeoM.Translate(40, 240)
	screen.DrawImage(levelPanel, op7)
	
	// Подсказка внизу
	hintPanel := ebiten.NewImage(500, 35)
	hintPanel.Fill(color.RGBA{20, 20, 50, 180})
	
	op8 := &ebiten.DrawImageOptions{}
	op8.GeoM.Translate(260, 710)
	screen.DrawImage(hintPanel, op8)
	
	// Рамка подсказки
	hintBorder := ebiten.NewImage(500, 2)
	hintBorder.Fill(color.RGBA{150, 150, 200, 150})
	
	op9 := &ebiten.DrawImageOptions{}
	op9.GeoM.Translate(260, 710)
	screen.DrawImage(hintBorder, op9)
	
	// Логотип Go365
	logoPanel := ebiten.NewImage(180, 30)
	logoPanel.Fill(color.RGBA{30, 60, 100, 180})
	
	op10 := &ebiten.DrawImageOptions{}
	op10.GeoM.Translate(820, 20)
	screen.DrawImage(logoPanel, op10)
	
	logoBorder := ebiten.NewImage(180, 2)
	logoBorder.Fill(color.RGBA{100, 200, 255, 200})
	
	op11 := &ebiten.DrawImageOptions{}
	op11.GeoM.Translate(820, 20)
	screen.DrawImage(logoBorder, op11)
}

func itos(n int) string {
	if n == 0 {
		return "0"
	}
	
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// ==================== ЗАПУСК ====================
func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Crystal Cascade - Puzzle GO | Go365 Challenge")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

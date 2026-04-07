// Crystal Cascade - Match-3 Game with Food Sprites
// Go365 Challenge - Day 99
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

	foodCount = 6
)

// ==================== ЕДА ====================
var foodNames = []string{
	"apple",
	"orange",
	"banana",
	"strawberry",
	"grapes",
	"cupcake",
}

var foodColors = []color.RGBA{
	{255, 50, 50, 255},
	{255, 165, 0, 255},
	{255, 255, 100, 255},
	{255, 100, 150, 255},
	{150, 100, 255, 255},
	{255, 180, 220, 255},
}

// ==================== СОСТОЯНИЯ ====================
type GameState int

const (
	StateIdle GameState = iota
	StateSwapping
	StateChecking
	StateRemoving
	StateDropping
)

// ==================== ЕДА-КРИСТАЛЛ ====================
type FoodPiece struct {
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

func NewFood(fType, col, row int, fromAbove bool) *FoodPiece {
	x := float64(boardX + col*(cellSize+cellPad))
	y := float64(boardY + row*(cellSize+cellPad))

	if fromAbove {
		y = float64(boardY - (cellSize+cellPad)*3)
	}

	return &FoodPiece{
		Type:    fType,
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

func (f *FoodPiece) Update() bool {
	dx := f.TargetX - f.X
	dy := f.TargetY - f.Y

	if dx*dx+dy*dy > 0.5 {
		f.X += dx * 0.2
		f.Y += dy * 0.2
		return true
	}

	f.X = f.TargetX
	f.Y = f.TargetY

	if f.Matched {
		f.Scale *= 0.85
		f.Alpha *= 0.8
		return f.Alpha > 0.01
	}

	return false
}

func (f *FoodPiece) Contains(mx, my float64) bool {
	size := float64(cellSize) * f.Scale / 2
	cx := f.X + float64(cellSize)/2
	cy := f.Y + float64(cellSize)/2
	return mx >= cx-size && mx <= cx+size && my >= cy-size && my <= cy+size
}

// ==================== ИГРОВОЕ ПОЛЕ ====================
type Board struct {
	pieces [][]*FoodPiece
}

func NewBoard() *Board {
	b := &Board{
		pieces: make([][]*FoodPiece, boardCols),
	}

	for c := 0; c < boardCols; c++ {
		b.pieces[c] = make([]*FoodPiece, boardRows)
	}

	b.Fill()
	return b
}

func (b *Board) Fill() {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			var fType int
			for {
				fType = rand.Intn(foodCount)
				if !b.wouldMatch(c, r, fType) {
					break
				}
			}
			b.pieces[c][r] = NewFood(fType, c, r, false)
		}
	}
}

func (b *Board) wouldMatch(col, row, fType int) bool {
	count := 1
	for c := col - 1; c >= 0 && b.pieces[c][row] != nil && b.pieces[c][row].Type == fType; c-- {
		count++
	}
	for c := col + 1; c < boardCols && b.pieces[c][row] != nil && b.pieces[c][row].Type == fType; c++ {
		count++
	}
	if count >= 3 {
		return true
	}

	count = 1
	for r := row - 1; r >= 0 && b.pieces[col][r] != nil && b.pieces[col][r].Type == fType; r-- {
		count++
	}
	for r := row + 1; r < boardRows && b.pieces[col][r] != nil && b.pieces[col][r].Type == fType; r++ {
		count++
	}
	return count >= 3
}

func (b *Board) GetPiece(col, row int) *FoodPiece {
	if col < 0 || col >= boardCols || row < 0 || row >= boardRows {
		return nil
	}
	return b.pieces[col][row]
}

func (b *Board) Swap(col1, row1, col2, row2 int) {
	if !b.isAdjacent(col1, row1, col2, row2) {
		return
	}

	p1 := b.pieces[col1][row1]
	p2 := b.pieces[col2][row2]

	if p1 == nil || p2 == nil {
		return
	}

	b.pieces[col1][row1] = p2
	b.pieces[col2][row2] = p1

	p1.Col, p1.Row = col2, row2
	p2.Col, p2.Row = col1, row1

	p1.TargetX = float64(boardX + col2*(cellSize+cellPad))
	p1.TargetY = float64(boardY + row2*(cellSize+cellPad))

	p2.TargetX = float64(boardX + col1*(cellSize+cellPad))
	p2.TargetY = float64(boardY + row1*(cellSize+cellPad))
}

func (b *Board) isAdjacent(c1, r1, c2, r2 int) bool {
	dc := c1 - c2
	dr := r1 - r2
	if dc < 0 {
		dc = -dc
	}
	if dr < 0 {
		dr = -dr
	}
	return (dc == 1 && dr == 0) || (dc == 0 && dr == 1)
}

func (b *Board) FindMatches() [][]*FoodPiece {
	var matches [][]*FoodPiece
	matched := make(map[[2]int]bool)

	for r := 0; r < boardRows; r++ {
		for c := 0; c <= boardCols-3; c++ {
			piece := b.pieces[c][r]
			if piece == nil || piece.Matched {
				continue
			}

			match := []*FoodPiece{piece}
			for cc := c + 1; cc < boardCols && b.pieces[cc][r] != nil && b.pieces[cc][r].Type == piece.Type; cc++ {
				match = append(match, b.pieces[cc][r])
			}

			if len(match) >= 3 {
				for _, p := range match {
					key := [2]int{p.Col, p.Row}
					if !matched[key] {
						matched[key] = true
						p.Matched = true
					}
				}
				matches = append(matches, match)
			}
		}
	}

	for c := 0; c < boardCols; c++ {
		for r := 0; r <= boardRows-3; r++ {
			piece := b.pieces[c][r]
			if piece == nil || piece.Matched {
				continue
			}

			match := []*FoodPiece{piece}
			for rr := r + 1; rr < boardRows && b.pieces[c][rr] != nil && b.pieces[c][rr].Type == piece.Type; rr++ {
				match = append(match, b.pieces[c][rr])
			}

			if len(match) >= 3 {
				for _, p := range match {
					key := [2]int{p.Col, p.Row}
					if !matched[key] {
						matched[key] = true
						p.Matched = true
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
			if b.pieces[c][r] != nil && b.pieces[c][r].Matched {
				b.pieces[c][r] = nil
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
			if b.pieces[c][r] == nil {
				if emptyRow == -1 {
					emptyRow = r
				}
				continue
			}

			if emptyRow != -1 {
				piece := b.pieces[c][r]
				b.pieces[c][emptyRow] = piece
				b.pieces[c][r] = nil

				piece.Row = emptyRow
				piece.TargetY = float64(boardY + emptyRow*(cellSize+cellPad))

				emptyRow--
				dropped = true
			}
		}

		for r := emptyRow; r >= 0; r-- {
			fType := rand.Intn(foodCount)
			piece := NewFood(fType, c, r, true)
			piece.TargetY = float64(boardY + r*(cellSize+cellPad))
			b.pieces[c][r] = piece
			dropped = true
		}
	}

	return dropped
}

func (b *Board) IsAnimating() bool {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			piece := b.pieces[c][r]
			if piece == nil {
				continue
			}

			dx := piece.TargetX - piece.X
			dy := piece.TargetY - piece.Y
			if dx*dx+dy*dy > 1 {
				return true
			}

			if piece.Matched && piece.Alpha > 0.01 {
				return true
			}
		}
	}
	return false
}

func (b *Board) GetPieceAt(mx, my float64) *FoodPiece {
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			piece := b.pieces[c][r]
			if piece != nil && piece.Contains(mx, my) {
				return piece
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
	board         *Board
	foods         []*ebiten.Image
	particles     []*Particle
	state         GameState
	score         int
	combo         int
	dragging      *FoodPiece
	dragStartX    float64
	dragStartY    float64
	dragOffsetX   float64
	dragOffsetY   float64
	isDragging    bool
	lastSwapC1    int
	lastSwapR1    int
	lastSwapC2    int
	lastSwapR2    int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		board:     NewBoard(),
		foods:     make([]*ebiten.Image, foodCount),
		particles: make([]*Particle, 0),
		state:     StateIdle,
	}

	g.loadFoodSprites()
	return g
}

func (g *Game) loadFoodSprites() {
	for i, name := range foodNames {
		path := "assets/food/" + name + ".png"
		data, err := foodFS.ReadFile(path)
		if err != nil {
			log.Printf("Warning: could not load %s: %v", name, err)
			continue
		}

		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("Warning: could not decode %s: %v", name, err)
			continue
		}

		g.foods[i] = ebiten.NewImageFromImage(img)
		log.Printf("Loaded food sprite: %s", name)
	}
}

func (g *Game) Update() error {
	for i := len(g.particles) - 1; i >= 0; i-- {
		if !g.particles[i].Update() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			piece := g.board.GetPiece(c, r)
			if piece != nil {
				piece.Update()
			}
		}
	}

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
				// Отмена - меняем обратно
				g.board.Swap(g.lastSwapC1, g.lastSwapR1, g.lastSwapC2, g.lastSwapR2)
				g.dragging = nil
				g.isDragging = false
				g.combo = 0
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
				g.spawnParticles(count)
			}
			g.board.Drop()
			g.state = StateDropping
		}

	case StateDropping:
		if !g.board.IsAnimating() {
			matches := g.board.FindMatches()
			if len(matches) > 0 {
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

func (g *Game) handleInput() {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	// Начало перетаскивания
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.state == StateIdle {
		piece := g.board.GetPieceAt(fx, fy)
		if piece != nil {
			g.dragging = piece
			g.dragStartX = fx
			g.dragStartY = fy
			g.dragOffsetX = piece.X
			g.dragOffsetY = piece.Y
			g.isDragging = false
		}
	}

	// Процесс перетаскивания
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.dragging != nil {
		dx := fx - g.dragStartX
		dy := fy - g.dragStartY

		// Начинаем тащить если сдвинулись достаточно
		if !g.isDragging && (dx*dx+dy*dy > 100) {
			g.isDragging = true
		}

		if g.isDragging {
			g.dragging.X = g.dragOffsetX + dx
			g.dragging.Y = g.dragOffsetY + dy
		}
	}

	// Конец перетаскивания
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && g.dragging != nil {
		if g.isDragging {
			// Определяем направление перетаскивания
			dx := g.dragging.X - g.dragOffsetX
			dy := g.dragging.Y - g.dragOffsetY

			targetCol := g.dragging.Col
			targetRow := g.dragging.Row

			if dx*dx > dy*dy {
				// Горизонтально
				if dx > 0 {
					targetCol++
				} else {
					targetCol--
				}
			} else {
				// Вертикально
				if dy > 0 {
					targetRow++
				} else {
					targetRow--
				}
			}

			// Проверяем что целевая клетка в пределах поля
			if targetCol >= 0 && targetCol < boardCols && targetRow >= 0 && targetRow < boardRows {
				target := g.board.GetPiece(targetCol, targetRow)
				if target != nil {
					// Сохраняем координаты для отмены
					g.lastSwapC1 = g.dragging.Col
					g.lastSwapR1 = g.dragging.Row
					g.lastSwapC2 = targetCol
					g.lastSwapR2 = targetRow

					g.board.Swap(g.dragging.Col, g.dragging.Row, targetCol, targetRow)
					g.state = StateSwapping
				}
			}

			// Возвращаем на место
			g.dragging.X = g.dragOffsetX
			g.dragging.Y = g.dragOffsetY
		}

		g.dragging = nil
		g.isDragging = false
	}
}

func (g *Game) spawnParticles(count int) {
	for i := 0; i < count*8; i++ {
		c := rand.Intn(boardCols)
		r := rand.Intn(boardRows)
		piece := g.board.GetPiece(c, r)
		if piece != nil {
			col := foodColors[piece.Type]
			x := float64(boardX + c*(cellSize+cellPad) + cellSize/2)
			y := float64(boardY + r*(cellSize+cellPad) + cellSize/2)
			g.particles = append(g.particles, NewParticle(x, y, col))
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.drawBoard(screen)
	g.drawFoods(screen)
	g.drawParticles(screen)
	g.drawUI(screen)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 35, 255})

	for i := 0; i < 50; i++ {
		x := float64((i * 137 + 50) % screenWidth)
		y := float64((i * 251 + 30) % screenHeight)
		brightness := uint8(100 + (i*31)%155)
		ebitenutil.DrawLine(screen, x, y, x, y,
			color.RGBA{brightness, brightness, brightness, 255})
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	w := boardCols*(cellSize+cellPad) + 20
	h := boardRows*(cellSize+cellPad) + 20

	board := ebiten.NewImage(w, h)
	board.Fill(color.RGBA{25, 25, 50, 240})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(boardX-10), float64(boardY-10))
	screen.DrawImage(board, op)

	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 200, 150, 255})

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(boardX-10), float64(boardY-12))
	screen.DrawImage(border, op2)

	for c := 0; c <= boardCols; c++ {
		x := float64(boardX + c*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, x, float64(boardY), x, float64(boardY+boardRows*(cellSize+cellPad)),
			color.RGBA{60, 80, 60, 100})
	}
	for r := 0; r <= boardRows; r++ {
		y := float64(boardY + r*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, float64(boardX), y, float64(boardX+boardCols*(cellSize+cellPad)), y,
			color.RGBA{60, 80, 60, 100})
	}
}

func (g *Game) drawFoods(screen *ebiten.Image) {
	// Сначала все кроме перетаскиваемого
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			piece := g.board.GetPiece(c, r)
			if piece == nil || piece.Alpha < 0.01 || piece == g.dragging {
				continue
			}

			g.drawSingleFood(screen, piece)
		}
	}

	// Перетаскиваемый поверх всех
	if g.dragging != nil && g.dragging.Alpha > 0.01 {
		g.drawSingleFood(screen, g.dragging)
	}
}

func (g *Game) drawSingleFood(screen *ebiten.Image, piece *FoodPiece) {
	// Подсветка для перетаскиваемого
	if piece == g.dragging && g.isDragging {
		highlight := ebiten.NewImage(cellSize+12, cellSize+12)
		highlight.Fill(color.RGBA{255, 255, 255, 60})

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(piece.X-6, piece.Y-6)
		screen.DrawImage(highlight, op)
	}

	img := g.foods[piece.Type]
	if img == nil {
		return
	}

	scale := piece.Scale
	if piece == g.dragging && g.isDragging {
		scale = 1.2
	}

	size := int(float64(cellSize) * scale)
	if size < 4 {
		return
	}

	imgScale := float64(size) / float64(img.Bounds().Dx())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(imgScale, imgScale)
	op.GeoM.Translate(piece.X, piece.Y)

	alpha := piece.Alpha
	if piece == g.dragging && g.isDragging {
		alpha = 0.85
	}
	op.ColorM.Scale(1, 1, 1, alpha)
	screen.DrawImage(img, op)
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
	panel := ebiten.NewImage(240, 300)
	panel.Fill(color.RGBA{20, 40, 20, 220})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(20, 40)
	screen.DrawImage(panel, op)

	topBorder := ebiten.NewImage(240, 4)
	topBorder.Fill(color.RGBA{100, 200, 150, 255})

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(20, 40)
	screen.DrawImage(topBorder, op2)

	titlePanel := ebiten.NewImage(200, 40)
	titlePanel.Fill(color.RGBA{40, 80, 40, 200})

	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(40, 70)
	screen.DrawImage(titlePanel, op3)

	titleBorder := ebiten.NewImage(200, 3)
	titleBorder.Fill(color.RGBA{100, 255, 150, 255})

	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(40, 70)
	screen.DrawImage(titleBorder, op4)

	scoreWidth := 180
	if g.score > 10000 {
		scoreWidth = 220
	}
	scorePanel := ebiten.NewImage(scoreWidth, 35)
	scorePanel.Fill(color.RGBA{40, 120, 40, 200})

	op5 := &ebiten.DrawImageOptions{}
	op5.GeoM.Translate(40, 140)
	screen.DrawImage(scorePanel, op5)

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

	level := g.score/1000 + 1
	levelWidth := 60 + level*30
	if levelWidth > 200 {
		levelWidth = 200
	}
	levelPanel := ebiten.NewImage(levelWidth, 30)
	levelPanel.Fill(color.RGBA{50, 100, 50, 200})

	op7 := &ebiten.DrawImageOptions{}
	op7.GeoM.Translate(40, 240)
	screen.DrawImage(levelPanel, op7)

	hintPanel := ebiten.NewImage(500, 35)
	hintPanel.Fill(color.RGBA{20, 40, 20, 180})

	op8 := &ebiten.DrawImageOptions{}
	op8.GeoM.Translate(260, 710)
	screen.DrawImage(hintPanel, op8)

	hintBorder := ebiten.NewImage(500, 2)
	hintBorder.Fill(color.RGBA{150, 200, 150, 150})

	op9 := &ebiten.DrawImageOptions{}
	op9.GeoM.Translate(260, 710)
	screen.DrawImage(hintBorder, op9)

	logoPanel := ebiten.NewImage(180, 30)
	logoPanel.Fill(color.RGBA{30, 60, 30, 180})

	op10 := &ebiten.DrawImageOptions{}
	op10.GeoM.Translate(820, 20)
	screen.DrawImage(logoPanel, op10)

	logoBorder := ebiten.NewImage(180, 2)
	logoBorder.Fill(color.RGBA{100, 255, 150, 200})

	op11 := &ebiten.DrawImageOptions{}
	op11.GeoM.Translate(820, 20)
	screen.DrawImage(logoBorder, op11)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Food Match-3 - Go365 Challenge")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Crystal Cascade - Match-3 Game with Food Sprites
// Go365 Challenge - Day 99
package main

import (
	"bytes"
	"embed"
	"image/color"
	"image/png"
	"log"
	"math"
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

var foodGlowColors = []color.RGBA{
	{255, 100, 100, 150},
	{255, 200, 100, 150},
	{255, 255, 150, 150},
	{255, 150, 200, 150},
	{200, 150, 255, 150},
	{255, 220, 240, 150},
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
	Type      int
	Col       int
	Row       int
	X         float64
	Y         float64
	TargetX   float64
	TargetY   float64
	Alpha     float64
	Scale     float64
	Matched   bool
	BobPhase  float64
	GlowPhase float64
}

func NewFood(fType, col, row int, fromAbove bool) *FoodPiece {
	x := float64(boardX + col*(cellSize+cellPad))
	y := float64(boardY + row*(cellSize+cellPad))

	if fromAbove {
		y = float64(boardY - (cellSize+cellPad)*3)
	}

	return &FoodPiece{
		Type:     fType,
		Col:      col,
		Row:      row,
		X:        x,
		Y:        y,
		TargetX:  x,
		TargetY:  y,
		Alpha:    1.0,
		Scale:    1.0,
		BobPhase: float64(rand.Intn(360)),
		GlowPhase: float64(rand.Intn(360)),
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

	// Анимация покачивания
	f.BobPhase += 2
	if f.BobPhase > 360 {
		f.BobPhase -= 360
	}
	f.GlowPhase += 3
	if f.GlowPhase > 360 {
		f.GlowPhase -= 360
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
	X, Y    float64
	VX, VY  float64
	Life    float64
	MaxLife float64
	Color   color.RGBA
	Size    float64
	Type    string
	Rotation float64
	RotSpeed float64
}

func NewParticle(x, y float64, col color.RGBA, ptype string) *Particle {
	angle := rand.Float64() * math.Pi * 2
	speed := 2 + rand.Float64()*6

	return &Particle{
		X:        x,
		Y:        y,
		VX:       math.Cos(angle) * speed,
		VY:       math.Sin(angle) * speed - 2,
		Life:     1.0,
		MaxLife:  1.0,
		Color:    col,
		Size:     float64(3 + rand.Intn(6)),
		Type:     ptype,
		Rotation: rand.Float64() * math.Pi * 2,
		RotSpeed: (rand.Float64() - 0.5) * 0.2,
	}
}

func (p *Particle) Update() bool {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.15
	p.VX *= 0.98
	p.Rotation += p.RotSpeed
	p.Life -= 0.015
	return p.Life > 0
}

// ==================== ЭФФЕКТЫ ТЕКСТА ====================
type FloatingText struct {
	X, Y    float64
	Text    string
	Life    float64
	Color   color.RGBA
	Size    float64
}

func NewFloatingText(x, y float64, text string, col color.RGBA) *FloatingText {
	return &FloatingText{
		X:     x,
		Y:     y,
		Text:  text,
		Life:  1.0,
		Color: col,
		Size:  20,
	}
}

func (ft *FloatingText) Update() bool {
	ft.Y -= 1.5
	ft.Life -= 0.012
	return ft.Life > 0
}

// ==================== ЭФФЕКТ ВСПЫШКИ ====================
type FlashEffect struct {
	X, Y  float64
	Life  float64
	MaxR  float64
	Color color.RGBA
}

func NewFlashEffect(x, y float64, col color.RGBA) *FlashEffect {
	return &FlashEffect{
		X:     x,
		Y:     y,
		Life:  1.0,
		MaxR:  80,
		Color: col,
	}
}

func (fe *FlashEffect) Update() bool {
	fe.Life -= 0.05
	return fe.Life > 0
}

// ==================== ИГРА ====================
type Game struct {
	board         *Board
	foods         []*ebiten.Image
	particles     []*Particle
	floatingTexts []*FloatingText
	flashEffects  []*FlashEffect
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
	time          float64
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		board:         NewBoard(),
		foods:         make([]*ebiten.Image, foodCount),
		particles:     make([]*Particle, 0),
		floatingTexts: make([]*FloatingText, 0),
		flashEffects:  make([]*FlashEffect, 0),
		state:         StateIdle,
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
	}
}

func (g *Game) Update() error {
	g.time += 0.016

	// Обновление частиц
	for i := len(g.particles) - 1; i >= 0; i-- {
		if !g.particles[i].Update() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Обновление плавающего текста
	for i := len(g.floatingTexts) - 1; i >= 0; i-- {
		if !g.floatingTexts[i].Update() {
			g.floatingTexts = append(g.floatingTexts[:i], g.floatingTexts[i+1:]...)
		}
	}

	// Обновление вспышек
	for i := len(g.flashEffects) - 1; i >= 0; i-- {
		if !g.flashEffects[i].Update() {
			g.flashEffects = append(g.flashEffects[:i], g.flashEffects[i+1:]...)
		}
	}

	// Обновление кристаллов
	for c := 0; c < boardCols; c++ {
		for r := 0; r < boardRows; r++ {
			piece := g.board.GetPiece(c, r)
			if piece != nil {
				piece.Update()
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

				// Эффекты при удалении
				g.spawnMatchParticles(count)
				g.spawnFloatingText(count, comboBonus)

				// Вспышка для больших комбо
				if g.combo >= 2 {
					g.flashEffects = append(g.flashEffects, 
						NewFlashEffect(float64(boardX+boardCols*(cellSize+cellPad)/2), 
							float64(boardY+boardRows*(cellSize+cellPad)/2), 
							color.RGBA{255, 255, 200, 100}))
				}
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

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.dragging != nil {
		dx := fx - g.dragStartX
		dy := fy - g.dragStartY

		if !g.isDragging && (dx*dx+dy*dy > 100) {
			g.isDragging = true
		}

		if g.isDragging {
			g.dragging.X = g.dragOffsetX + dx
			g.dragging.Y = g.dragOffsetY + dy
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && g.dragging != nil {
		if g.isDragging {
			dx := g.dragging.X - g.dragOffsetX
			dy := g.dragging.Y - g.dragOffsetY

			targetCol := g.dragging.Col
			targetRow := g.dragging.Row

			if dx*dx > dy*dy {
				if dx > 0 {
					targetCol++
				} else {
					targetCol--
				}
			} else {
				if dy > 0 {
					targetRow++
				} else {
					targetRow--
				}
			}

			if targetCol >= 0 && targetCol < boardCols && targetRow >= 0 && targetRow < boardRows {
				target := g.board.GetPiece(targetCol, targetRow)
				if target != nil {
					g.lastSwapC1 = g.dragging.Col
					g.lastSwapR1 = g.dragging.Row
					g.lastSwapC2 = targetCol
					g.lastSwapR2 = targetRow

					g.board.Swap(g.dragging.Col, g.dragging.Row, targetCol, targetRow)
					g.state = StateSwapping
				}
			}

			g.dragging.X = g.dragOffsetX
			g.dragging.Y = g.dragOffsetY
		}

		g.dragging = nil
		g.isDragging = false
	}
}

func (g *Game) spawnMatchParticles(count int) {
	for i := 0; i < count*12; i++ {
		c := rand.Intn(boardCols)
		r := rand.Intn(boardRows)
		piece := g.board.GetPiece(c, r)
		if piece != nil {
			col := foodColors[piece.Type]
			x := float64(boardX + c*(cellSize+cellPad) + cellSize/2)
			y := float64(boardY + r*(cellSize+cellPad) + cellSize/2)
			
			// Искры
			g.particles = append(g.particles, NewParticle(x, y, col, "spark"))
			
			// Звёздочки
			if rand.Float64() > 0.5 {
				g.particles = append(g.particles, NewParticle(x, y, 
					color.RGBA{255, 255, 200, 255}, "star"))
			}
		}
	}
}

func (g *Game) spawnFloatingText(count int, score int) {
	cx := boardX + boardCols*(cellSize+cellPad)/2
	cy := boardY + boardRows*(cellSize+cellPad)/2
	
	text := ""
	if g.combo > 1 {
		text = "COMBO x" + itos(g.combo) + "!"
	} else {
		text = "+" + itos(score)
	}
	
	col := color.RGBA{255, 255, 100, 255}
	if g.combo >= 3 {
		col = color.RGBA{255, 200, 50, 255}
	}
	if g.combo >= 5 {
		col = color.RGBA{255, 100, 100, 255}
	}
	
	g.floatingTexts = append(g.floatingTexts, 
		NewFloatingText(float64(cx), float64(cy), text, col))
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.drawBoard(screen)
	g.drawFlashEffects(screen)
	g.drawFoods(screen)
	g.drawParticles(screen)
	g.drawFloatingTexts(screen)
	g.drawUI(screen)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	// Градиентный фон
	for y := 0; y < screenHeight; y++ {
		ratio := float64(y) / float64(screenHeight)
		r := uint8(10 + float64(20)*ratio)
		gr := uint8(15 + float64(25)*ratio)
		b := uint8(30 + float64(40)*ratio)
		
		line := ebiten.NewImage(screenWidth, 1)
		line.Fill(color.RGBA{r, gr, b, 255})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(y))
		screen.DrawImage(line, op)
	}

	// Анимированные звёзды
	for i := 0; i < 60; i++ {
		x := float64((i * 137 + 50) % screenWidth)
		y := float64((i * 251 + 30) % screenHeight)
		twinkle := 0.5 + 0.5*math.Sin(g.time*2+float64(i))
		brightness := uint8(80 + 175*twinkle)
		size := 1 + int(twinkle)
		
		for dx := 0; dx < size; dx++ {
			for dy := 0; dy < size; dy++ {
				ebitenutil.DrawLine(screen, x+float64(dx), y+float64(dy), 
					x+float64(dx), y+float64(dy),
					color.RGBA{brightness, brightness, brightness, 255})
			}
		}
	}
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	w := boardCols*(cellSize+cellPad) + 20
	h := boardRows*(cellSize+cellPad) + 20

	// Фон с градиентом
	board := ebiten.NewImage(w, h)
	board.Fill(color.RGBA{20, 30, 40, 240})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(boardX-10), float64(boardY-10))
	screen.DrawImage(board, op)

	// Светящаяся рамка
	border := ebiten.NewImage(w, 4)
	border.Fill(color.RGBA{100, 200, 255, 255})

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(boardX-10), float64(boardY-12))
	screen.DrawImage(border, op2)

	// Сетка
	for c := 0; c <= boardCols; c++ {
		x := float64(boardX + c*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, x, float64(boardY), x, float64(boardY+boardRows*(cellSize+cellPad)),
			color.RGBA{40, 60, 80, 150})
	}
	for r := 0; r <= boardRows; r++ {
		y := float64(boardY + r*(cellSize+cellPad) - cellPad/2)
		ebitenutil.DrawLine(screen, float64(boardX), y, float64(boardX+boardCols*(cellSize+cellPad)), y,
			color.RGBA{40, 60, 80, 150})
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
	img := g.foods[piece.Type]
	if img == nil {
		return
	}

	// Свечение вокруг еды (пульсирующее)
	glowIntensity := 0.5 + 0.5*math.Sin(piece.GlowPhase*math.Pi/180)
	glowSize := int(5 + 3*glowIntensity)
	
	if piece != g.dragging || !g.isDragging {
		glow := ebiten.NewImage(cellSize+glowSize*2, cellSize+glowSize*2)
		glowColor := foodGlowColors[piece.Type]
		glowColor.A = uint8(float64(glowColor.A) * glowIntensity * piece.Alpha)
		glow.Fill(glowColor)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(piece.X-float64(glowSize), piece.Y-float64(glowSize))
		op.ColorM.Scale(1, 1, 1, piece.Alpha)
		screen.DrawImage(glow, op)
	}

	// Подсветка для перетаскиваемого
	if piece == g.dragging && g.isDragging {
		highlight := ebiten.NewImage(cellSize+16, cellSize+16)
		highlight.Fill(color.RGBA{255, 255, 255, 100})

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(piece.X-8, piece.Y-8)
		screen.DrawImage(highlight, op)
	}

	scale := piece.Scale
	if piece == g.dragging && g.isDragging {
		scale = 1.2
	}

	// Покачивание для не-перетаскиваемых
	bobOffset := 0.0
	if piece != g.dragging {
		bobOffset = math.Sin(piece.BobPhase*math.Pi/180) * 2
	}

	size := int(float64(cellSize) * scale)
	if size < 4 {
		return
	}

	imgScale := float64(size) / float64(img.Bounds().Dx())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(imgScale, imgScale)
	op.GeoM.Translate(piece.X, piece.Y+bobOffset)

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

		if p.Type == "star" {
			// Рисуем звёздочку
			g.drawStarParticle(screen, p)
		} else {
			// Обычная частица
			particle := ebiten.NewImage(size, size)
			alpha := uint8(p.Life * 255)
			particle.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha})

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
			op.GeoM.Rotate(p.Rotation)
			screen.DrawImage(particle, op)
		}
	}
}

func (g *Game) drawStarParticle(screen *ebiten.Image, p *Particle) {
	size := int(p.Size * p.Life * 1.5)
	if size < 2 {
		return
	}

	// Создаём звёздочку
	star := ebiten.NewImage(size, size)
	
	// Рисуем звезду
	center := size / 2
	alpha := uint8(p.Life * 255)
	col := color.RGBA{255, 255, 200, alpha}
	
	for i := 0; i < 5; i++ {
		angle := float64(i) * math.Pi * 2 / 5 - math.Pi/2
		x := center + int(float64(center)*0.8*math.Cos(angle))
		y := center + int(float64(center)*0.8*math.Sin(angle))
		ebitenutil.DrawLine(star, float64(center), float64(center), 
			float64(x), float64(y), col)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
	op.GeoM.Rotate(p.Rotation)
	screen.DrawImage(star, op)
}

func (g *Game) drawFlashEffects(screen *ebiten.Image) {
	for _, fe := range g.flashEffects {
		radius := fe.MaxR * (1 - fe.Life)
		alpha := uint8(fe.Life * 150)
		
		// Кольцо вспышки
		segments := 32
		for i := 0; i < segments; i++ {
			angle1 := float64(i) * math.Pi * 2 / float64(segments)
			angle2 := float64(i+1) * math.Pi * 2 / float64(segments)
			
			x1 := fe.X + radius*math.Cos(angle1)
			y1 := fe.Y + radius*math.Sin(angle1)
			x2 := fe.X + radius*math.Cos(angle2)
			y2 := fe.Y + radius*math.Sin(angle2)
			
			ebitenutil.DrawLine(screen, x1, y1, x2, y2, 
				color.RGBA{fe.Color.R, fe.Color.G, fe.Color.B, alpha})
		}
	}
}

func (g *Game) drawFloatingTexts(screen *ebiten.Image) {
	for _, ft := range g.floatingTexts {
		if ft.Life <= 0 {
			continue
		}
		
		// Тень текста
		shadow := ebiten.NewImage(200, 30)
		shadow.Fill(color.RGBA{0, 0, 0, uint8(ft.Life * 200)})
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ft.X-98, ft.Y+2)
		screen.DrawImage(shadow, op)
		
		// Сам текст (пока прямоугольник с текстом внутри)
		textBg := ebiten.NewImage(200, 30)
		textBg.Fill(color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, uint8(ft.Life * 255)})
		
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(ft.X-100, ft.Y)
		screen.DrawImage(textBg, op2)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Панель счёта с тенью
	shadow := ebiten.NewImage(244, 304)
	shadow.Fill(color.RGBA{0, 0, 0, 100})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(22, 42)
	screen.DrawImage(shadow, op)

	panel := ebiten.NewImage(240, 300)
	panel.Fill(color.RGBA{15, 25, 35, 230})

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(20, 40)
	screen.DrawImage(panel, op2)

	// Градиентная рамка панели
	topBorder := ebiten.NewImage(240, 4)
	topBorder.Fill(color.RGBA{100, 200, 255, 255})

	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(20, 40)
	screen.DrawImage(topBorder, op3)

	// Заголовок с эффектом
	titlePanel := ebiten.NewImage(200, 40)
	titlePanel.Fill(color.RGBA{30, 50, 70, 200})

	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Translate(40, 70)
	screen.DrawImage(titlePanel, op4)

	titleBorder := ebiten.NewImage(200, 3)
	titleBorder.Fill(color.RGBA{100, 200, 255, 255})

	op5 := &ebiten.DrawImageOptions{}
	op5.GeoM.Translate(40, 70)
	screen.DrawImage(titleBorder, op5)

	// Очки
	scoreWidth := 180
	if g.score > 10000 {
		scoreWidth = 220
	}
	scorePanel := ebiten.NewImage(scoreWidth, 35)
	scorePanel.Fill(color.RGBA{40, 100, 140, 200})

	op6 := &ebiten.DrawImageOptions{}
	op6.GeoM.Translate(40, 140)
	screen.DrawImage(scorePanel, op6)

	// Комбо с пульсацией
	if g.combo > 1 {
		comboWidth := 120 + g.combo*20
		if comboWidth > 220 {
			comboWidth = 220
		}
		comboPanel := ebiten.NewImage(comboWidth, 30)
		comboPanel.Fill(color.RGBA{200, 150, 50, 200})

		op7 := &ebiten.DrawImageOptions{}
		op7.GeoM.Translate(40, 190)
		screen.DrawImage(comboPanel, op7)
		
		// Эффект свечения для комбо
		glowIntensity := 0.5 + 0.5*math.Sin(g.time*5)
		glowAlpha := uint8(glowIntensity * 100)
		comboGlow := ebiten.NewImage(comboWidth+10, 40)
		comboGlow.Fill(color.RGBA{255, 200, 100, glowAlpha})
		
		op7g := &ebiten.DrawImageOptions{}
		op7g.GeoM.Translate(35, 185)
		screen.DrawImage(comboGlow, op7g)
	}

	// Уровень
	level := g.score/1000 + 1
	levelWidth := 60 + level*30
	if levelWidth > 200 {
		levelWidth = 200
	}
	levelPanel := ebiten.NewImage(levelWidth, 30)
	levelPanel.Fill(color.RGBA{50, 120, 80, 200})

	op8 := &ebiten.DrawImageOptions{}
	op8.GeoM.Translate(40, 240)
	screen.DrawImage(levelPanel, op8)

	// Подсказка
	hintPanel := ebiten.NewImage(500, 35)
	hintPanel.Fill(color.RGBA{15, 25, 35, 180})

	op9 := &ebiten.DrawImageOptions{}
	op9.GeoM.Translate(260, 710)
	screen.DrawImage(hintPanel, op9)

	hintBorder := ebiten.NewImage(500, 2)
	hintBorder.Fill(color.RGBA{150, 200, 250, 150})

	op10 := &ebiten.DrawImageOptions{}
	op10.GeoM.Translate(260, 710)
	screen.DrawImage(hintBorder, op10)

	// Логотип
	logoPanel := ebiten.NewImage(180, 30)
	logoPanel.Fill(color.RGBA{20, 40, 60, 180})

	op11 := &ebiten.DrawImageOptions{}
	op11.GeoM.Translate(820, 20)
	screen.DrawImage(logoPanel, op11)

	logoBorder := ebiten.NewImage(180, 2)
	logoBorder.Fill(color.RGBA{100, 200, 255, 200})

	op12 := &ebiten.DrawImageOptions{}
	op12.GeoM.Translate(820, 20)
	screen.DrawImage(logoBorder, op12)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
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

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Food Match-3 - Go365 Challenge")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Package game содержит главный Game struct — оркестратор игры.
package game

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"

	"github.com/playgo/puzzle_go/internal/audio"
	"github.com/playgo/puzzle_go/internal/board"
	"github.com/playgo/puzzle_go/internal/config"
	"github.com/playgo/puzzle_go/internal/entity"
	"github.com/playgo/puzzle_go/internal/render"
	"github.com/playgo/puzzle_go/internal/ui"
)

// Game — главный orchestrator.
type Game struct {
	bd        [config.Rows][config.Cols]int
	score     int
	combo     int
	moves     int
	state     config.State
	selR      int
	selC      int
	hovR      int
	hovC      int
	particles []entity.Particle
	selectAnim *entity.SelectAnim
	rmAnims   []entity.RemoveAnim
	slAnims   []entity.SlideAnim
	busy      bool
	busyTimer int
	flash     int
	msg       string
	msgTimer  int
	buttons   []*entity.MenuButton
	menuAnim  int
	highScore int
	cascade   bool
	spr       *render.SpriteCache
	snd       *audio.Manager
}

// NewGame создаёт новую игру.
func NewGame(spr *render.SpriteCache, snd *audio.Manager) *Game {
	g := &Game{selR: -1, selC: -1, hovR: -1, hovC: -1, spr: spr, snd: snd}
	g.loadHighScore()
	g.buttons = ui.LoadMenuButtons()
	return g
}

func (g *Game) loadHighScore() {
	if d, err := os.ReadFile("highscore.txt"); err == nil {
		fmt.Sscanf(string(d), "%d", &g.highScore)
	}
}

func (g *Game) saveHighScore() {
	if g.score > g.highScore {
		g.highScore = g.score
		os.WriteFile("highscore.txt", []byte(fmt.Sprintf("%d", g.highScore)), 0644)
	}
}

// Start начинает новую игру.
func (g *Game) Start() {
	g.score, g.combo, g.moves = 0, 0, 0
	g.state = config.StatePlay
	g.selR, g.selC = -1, -1
	g.busy, g.busyTimer, g.cascade = false, 0, false
	g.slAnims, g.rmAnims, g.particles = nil, nil, nil
	g.selectAnim, g.msg = nil, ""
	board.FillBoard(&g.bd)
}

// Update — главный update loop.
func (g *Game) Update() error {
	g.menuAnim++

	// Update particles
	active := make([]entity.Particle, 0)
	for i := range g.particles {
		if g.particles[i].Update() {
			active = append(active, g.particles[i])
		}
	}
	g.particles = active

	// Update remove anims
	ra := make([]entity.RemoveAnim, 0)
	for i := range g.rmAnims {
		if !g.rmAnims[i].Update() {
			ra = append(ra, g.rmAnims[i])
		}
	}
	g.rmAnims = ra

	// Update slide anims
	sa := make([]entity.SlideAnim, 0)
	for i := range g.slAnims {
		if !g.slAnims[i].Update() {
			sa = append(sa, g.slAnims[i])
		}
	}
	g.slAnims = sa

	if g.selectAnim != nil { g.selectAnim.Update() }
	if g.msgTimer > 0 { g.msgTimer-- }
	if g.flash > 0 { g.flash-- }

	// Cascade check (state machine instead of goroutine!)
	if g.cascade && !g.busy {
		g.busyTimer--
		if g.busyTimer <= 0 {
			g.cascade = false
			moved := board.ApplyGravity(&g.bd)
			filled := board.FillEmpty(&g.bd)

			// Create slide anims for moved
			for _, pos := range moved {
				r, c := pos[0], pos[1]
				sy := float64(r*config.Tile) + config.BoardOffY
				g.slAnims = append(g.slAnims, entity.SlideAnim{
					R: r, C: c, SX: float64(c*config.Tile)+config.BoardOffX, SY: sy - float64(config.Tile),
					TX: float64(c*config.Tile)+config.BoardOffX, TY: sy,
					T: 0, Total: 12, TypeID: g.bd[r][c], Spr: g.spr.Gems[g.bd[r][c]],
				})
			}
			// Slide anims for filled (fall from top)
			for _, pos := range filled {
				r, c := pos[0], pos[1]
				sy := -float64((config.Rows - r) * config.Tile)
				ty := float64(r*config.Tile) + config.BoardOffY
				g.slAnims = append(g.slAnims, entity.SlideAnim{
					R: r, C: c, SX: float64(c*config.Tile)+config.BoardOffX, SY: sy,
					TX: float64(c*config.Tile)+config.BoardOffX, TY: ty,
					T: 0, Total: 15, TypeID: g.bd[r][c], Spr: g.spr.Gems[g.bd[r][c]],
				})
			}

			g.busy = len(g.slAnims) > 0
			if !g.busy {
				matches := board.FindMatches(g.bd)
				if len(matches) > 0 {
					g.combo++
					g.busy = true
					g.removeMatches(matches)
				} else {
					g.combo = 0
					if !board.HasValidMoves(g.bd) {
						board.ShuffleBoard(&g.bd)
						g.msg = "SHUFFLED!"; g.msgTimer = 90
					}
					if g.score >= config.TargetScore {
						g.state = config.StateWin
						g.saveHighScore()
						audio.Play(g.snd.Win)
					}
				}
			}
		}
	}

	switch g.state {
	case config.StateMenu:
		mx, my := ebiten.CursorPosition()
		ui.UpdateHover(g.buttons, mx, my)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.buttons {
				if b.Hover {
					if b.Label == "PLAY" { g.Start() }
					if b.Label == "OPTIONS" { g.state = config.StateOptions; g.buttons = ui.LoadOptionsButtons() }
					if b.Label == "EXIT" { os.Exit(0) }
				}
			}
		}
	case config.StatePlay:
		mx, my := ebiten.CursorPosition()
		g.hovR, g.hovC = px2rc(mx, my)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.hovR >= 0 {
			g.click(g.hovR, g.hovC)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.state = config.StatePause
			g.buttons = ui.LoadPauseButtons()
		}
	case config.StatePause:
		mx, my := ebiten.CursorPosition()
		ui.UpdateHover(g.buttons, mx, my)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.buttons {
				if b.Hover {
					if b.Label == "RESUME" { g.state = config.StatePlay }
					if b.Label == "RESTART" { g.Start() }
				}
			}
		}
	case config.StateOptions:
		mx, my := ebiten.CursorPosition()
		ui.UpdateHover(g.buttons, mx, my)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, b := range g.buttons {
				if b.Hover && b.Label == "BACK" { g.state = config.StateMenu; g.buttons = ui.LoadMenuButtons() }
			}
		}
	case config.StateWin:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) { g.state = config.StateMenu; g.buttons = ui.LoadMenuButtons() }
	}
	return nil
}

func px2rc(px, py int) (int, int) {
	r, c := (py-config.BoardOffY)/config.Tile, (px-config.BoardOffX)/config.Tile
	if r < 0 || r >= config.Rows || c < 0 || c >= config.Cols { return -1, -1 }
	return r, c
}

func abs(x int) int { if x < 0 { return -x }; return x }

func (g *Game) click(r, c int) {
	if g.busy || g.state != config.StatePlay { return }
	if g.selR < 0 {
		g.selR, g.selC = r, c
		g.selectAnim = &entity.SelectAnim{R: r, C: c, T: 0}
		return
	}
	if g.selR == r && g.selC == c {
		g.selR, g.selC = -1, -1; g.selectAnim = nil; return
	}
	if abs(g.selR-r)+abs(g.selC-c) == 1 {
		g.moves++; g.trySwap(g.selR, g.selC, r, c)
		g.selR, g.selC = -1, -1; g.selectAnim = nil
	} else {
		g.selR, g.selC = r, c
		g.selectAnim = &entity.SelectAnim{R: r, C: c, T: 0}
	}
}

func (g *Game) trySwap(r1, c1, r2, c2 int) {
	board.Swap(&g.bd, r1, c1, r2, c2)
	sx1, sy1 := float64(c1*config.Tile)+config.BoardOffX, float64(r1*config.Tile)+config.BoardOffY
	sx2, sy2 := float64(c2*config.Tile)+config.BoardOffX, float64(r2*config.Tile)+config.BoardOffY
	g.slAnims = append(g.slAnims, entity.SlideAnim{
		R: r1, C: c1, SX: sx1, SY: sy1, TX: sx2, TY: sy2, T: 0, Total: 10, TypeID: g.bd[r1][c1], Spr: g.spr.Gems[g.bd[r1][c1]],
	})
	g.slAnims = append(g.slAnims, entity.SlideAnim{
		R: r2, C: c2, SX: sx2, SY: sy2, TX: sx1, TY: sy1, T: 0, Total: 10, TypeID: g.bd[r2][c2], Spr: g.spr.Gems[g.bd[r2][c2]],
	})
	matches := board.FindMatches(g.bd)
	if len(matches) > 0 {
		audio.Play(g.snd.Swap); g.combo = 1; g.busy = true; g.removeMatches(matches)
	} else {
		board.Swap(&g.bd, r1, c1, r2, c2); audio.Play(g.snd.Bad); g.flash = 10
	}
}

func (g *Game) removeMatches(matches map[[2]int]bool) {
	pts := 0
	for pos := range matches {
		r, c := pos[0], pos[1]
		px := float64(c*config.Tile) + config.BoardOffX + config.Tile/2
		py := float64(r*config.Tile) + config.BoardOffY + config.Tile/2
		render.SpawnParts(px, py, render.GemColor(g.bd[r][c]), 8, &g.particles)
		g.rmAnims = append(g.rmAnims, entity.RemoveAnim{
			R: r, C: c, X: float64(c*config.Tile) + config.BoardOffX,
			Y: float64(r*config.Tile) + config.BoardOffY, T: 0, Total: 20, TypeID: g.bd[r][c],
		})
		pts += 10
	}
	board.ClearMatches(&g.bd, matches)
	g.score += pts * g.combo
	if g.combo > 1 {
		audio.Play(g.snd.Combo)
		g.msg = fmt.Sprintf("COMBO x%d!", g.combo); g.msgTimer = 60
	} else {
		audio.Play(g.snd.Match)
	}
	g.busyTimer = config.CascadeDelay; g.cascade = true
}

// Draw — главный draw.
func (g *Game) Draw(s *ebiten.Image) {
	s.Fill(color.RGBA{10, 10, 30, 255})
	switch g.state {
	case config.StateMenu:    g.drawMenu(s)
	case config.StatePlay:    g.drawPlay(s)
	case config.StatePause:   g.drawPlay(s); g.drawOverlay(s)
	case config.StateOptions: g.drawOptions(s)
	case config.StateWin:     g.drawWin(s)
	}
}

func (g *Game) drawMenu(s *ebiten.Image) {
	for y := 0; y < config.WinH/config.Tile+1; y++ {
		for x := 0; x < config.WinW/config.Tile+1; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*config.Tile), float64(y*config.Tile))
			s.DrawImage(g.spr.BGTile, op)
		}
	}
	render.DrawRect(s, g.spr.WhitePixel, 0, 0, config.WinW, config.WinH, color.RGBA{0, 0, 0, 0}, 0.6)
	title := "PUZZLE GO"
	bw := text.BoundString(basicfont.Face7x13, title)
	text.Draw(s, title, basicfont.Face7x13, config.WinW/2-bw.Dx()/2, 120, color.RGBA{255, 220, 100, 255})
	sub := "Match-3 Gem Crusher"
	bw2 := text.BoundString(basicfont.Face7x13, sub)
	text.Draw(s, sub, basicfont.Face7x13, config.WinW/2-bw2.Dx()/2, 145, color.RGBA{200, 200, 200, 255})
	ui.DrawButtons(s, g.buttons)
	hs := fmt.Sprintf("Best: %d", g.highScore)
	bw3 := text.BoundString(basicfont.Face7x13, hs)
	text.Draw(s, hs, basicfont.Face7x13, config.WinW/2-bw3.Dx()/2, 520, color.RGBA{255, 215, 0, 255})
}

func (g *Game) drawPlay(s *ebiten.Image) {
	// BG tiles
	for y := 0; y < config.Rows; y++ {
		for x := 0; x < config.Cols; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(config.BoardOffX+x*config.Tile), float64(config.BoardOffY+y*config.Tile))
			s.DrawImage(g.spr.BGTile, op)
		}
	}

	// Static gems
	animCells := make(map[[2]int]bool)
	for i := range g.slAnims { animCells[[2]int{g.slAnims[i].R, g.slAnims[i].C}] = true }

	for r := 0; r < config.Rows; r++ {
		for c := 0; c < config.Cols; c++ {
			if g.bd[r][c] < 0 || animCells[[2]int{r, c}] { continue }
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(config.BoardOffX+c*config.Tile), float64(config.BoardOffY+r*config.Tile))
			if g.spr.Gems[g.bd[r][c]] != nil { s.DrawImage(g.spr.Gems[g.bd[r][c]], op) }
		}
	}

	// Slide anims
	for i := range g.slAnims {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.slAnims[i].X(), g.slAnims[i].Y())
		if g.slAnims[i].Spr != nil { s.DrawImage(g.slAnims[i].Spr, op) }
	}

	// Remove anims
	for i := range g.rmAnims {
		a := &g.rmAnims[i]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(a.X, a.Y)
		op.GeoM.Translate(config.Tile/2, config.Tile/2)
		op.GeoM.Scale(a.Scale(), a.Scale())
		op.GeoM.Translate(-config.Tile/2, -config.Tile/2)
		op.ColorScale.ScaleAlpha(float32(a.Alpha()))
		if g.spr.Gems[a.TypeID] != nil {
			s.DrawImage(g.spr.Gems[a.TypeID], op)
		}
		if a.Progress() < 0.3 {
			op2 := &ebiten.DrawImageOptions{}
			op2.GeoM.Translate(a.X, a.Y)
			op2.ColorScale.ScaleAlpha(float32(0.5 * (1.0 - a.Progress()/0.3)))
			s.DrawImage(g.spr.WhitePixel, op2)
		}
	}

	// Hover
	if g.hovR >= 0 && !g.busy {
		render.DrawRect(s, g.spr.WhitePixel, config.BoardOffX+g.hovC*config.Tile, config.BoardOffY+g.hovR*config.Tile, config.Tile, config.Tile, color.White, 0.2)
	}

	// Selection
	if g.selR >= 0 && g.selectAnim != nil {
		px := float64(g.selectAnim.C*config.Tile) + config.BoardOffX
		py := float64(g.selectAnim.R*config.Tile) + config.BoardOffY
		pulse := g.selectAnim.Pulse()
		c := color.RGBA{255, 255, 255, uint8(150 + 105*pulse)}
		render.DrawRect(s, g.spr.WhitePixel, int(px)+2, int(py)+2, config.Tile-4, config.Tile-4, c, 1)
		if g.selectAnim.T%10 < 5 {
			cc := color.RGBA{255, 255, 100, 200}
			for _, off := range [][2]int{{4, 4}, {config.Tile - 6, 4}, {4, config.Tile - 6}, {config.Tile - 6, config.Tile - 6}} {
				render.DrawRect(s, g.spr.WhitePixel, int(px)+off[0], int(py)+off[1], 4, 4, cc, 1)
			}
		}
	}

	// Flash
	if g.flash > 0 { render.DrawRect(s, g.spr.WhitePixel, 0, 0, config.WinW, config.WinH, color.White, float32(g.flash)/20.0) }

	// Particles
	for i := range g.particles {
		p := &g.particles[i]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Rotate(p.Rotation)
		op.GeoM.Translate(-float64(p.Sz)/2, -float64(p.Sz)/2)
		op.GeoM.Scale(float64(p.Sz), float64(p.Sz))
		if col, ok := p.Clr.(color.RGBA); ok {
			op.ColorScale.SetR(float32(col.R)/255); op.ColorScale.SetG(float32(col.G)/255)
			op.ColorScale.SetB(float32(col.B)/255); op.ColorScale.SetA(p.Alpha() * float32(col.A)/255)
		}
		s.DrawImage(g.spr.WhitePixel, op)
	}

	// HUD
	if g.spr.HUDBG != nil {
		op := &ebiten.DrawImageOptions{}; s.DrawImage(g.spr.HUDBG, op)
	}
	text.Draw(s, fmt.Sprintf("Score: %d", g.score), basicfont.Face7x13, 15, 18, color.RGBA{255, 255, 255, 255})
	text.Draw(s, fmt.Sprintf("Moves: %d", g.moves), basicfont.Face7x13, 200, 18, color.RGBA{200, 200, 200, 255})
	text.Draw(s, fmt.Sprintf("Target: %d", config.TargetScore), basicfont.Face7x13, 350, 18, color.RGBA{150, 200, 255, 255})

	progress := float64(g.score) / float64(config.TargetScore)
	if progress > 1 { progress = 1 }
	render.DrawRect(s, g.spr.WhitePixel, 15, 30, 400, 8, color.RGBA{50, 50, 50, 255}, 1)
	render.DrawRect(s, g.spr.WhitePixel, 15, 30, int(400*progress), 8, color.RGBA{0, 200, 100, 255}, 1)

	if g.msgTimer > 0 {
		alpha := float32(g.msgTimer) / 60.0
		bw := text.BoundString(basicfont.Face7x13, g.msg)
		render.DrawRect(s, g.spr.WhitePixel, config.WinW/2-bw.Dx()/2-10, config.WinH/2-60, bw.Dx()+20, 30, color.RGBA{0, 0, 0, 0}, alpha*0.7)
		text.Draw(s, g.msg, basicfont.Face7x13, config.WinW/2-bw.Dx()/2, config.WinH/2-42, color.RGBA{255, 255, 100, 255})
	}
}

func (g *Game) drawOverlay(s *ebiten.Image) {
	render.DrawRect(s, g.spr.WhitePixel, 0, 0, config.WinW, config.WinH, color.RGBA{0, 0, 0, 0}, 0.7)
	ui.DrawButtons(s, g.buttons)
}

func (g *Game) drawOptions(s *ebiten.Image) {
	s.Fill(color.RGBA{20, 20, 40, 255})
	t := "CONTROLS"
	bw := text.BoundString(basicfont.Face7x13, t)
	text.Draw(s, t, basicfont.Face7x13, config.WinW/2-bw.Dx()/2, 100, color.White)
	text.Draw(s, "Click: Select/Swap gems", basicfont.Face7x13, config.WinW/2-90, 160, color.RGBA{200, 200, 200, 255})
	text.Draw(s, "ESC/P: Pause", basicfont.Face7x13, config.WinW/2-55, 190, color.RGBA{200, 200, 200, 255})
	text.Draw(s, "Match 3+ gems to score!", basicfont.Face7x13, config.WinW/2-85, 250, color.RGBA{255, 255, 100, 255})
	text.Draw(s, fmt.Sprintf("Target: %d points to win", config.TargetScore), basicfont.Face7x13, config.WinW/2-100, 280, color.RGBA{100, 200, 255, 255})
	ui.DrawButtons(s, g.buttons)
}

func (g *Game) drawWin(s *ebiten.Image) {
	g.drawPlay(s)
	render.DrawRect(s, g.spr.WhitePixel, 0, 0, config.WinW, config.WinH, color.RGBA{0, 0, 0, 0}, 0.8)
	t := "YOU WIN!"
	bw := text.BoundString(basicfont.Face7x13, t)
	text.Draw(s, t, basicfont.Face7x13, config.WinW/2-bw.Dx()/2, config.WinH/2-40, color.RGBA{255, 215, 0, 255})
	sc := fmt.Sprintf("Score: %d  Moves: %d", g.score, g.moves)
	bw2 := text.BoundString(basicfont.Face7x13, sc)
	text.Draw(s, sc, basicfont.Face7x13, config.WinW/2-bw2.Dx()/2, config.WinH/2+10, color.White)
	text.Draw(s, "Press ENTER for menu", basicfont.Face7x13, config.WinW/2-70, config.WinH/2+50, color.RGBA{200, 200, 200, 255})
}

// Layout возвращает размер экрана.
func (g *Game) Layout(w, h int) (int, int) { return config.WinW, config.WinH }

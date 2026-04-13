package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawUI renders all UI elements: score, timer, buttons, game over screen.
func drawUI(screen *ebiten.Image, g *Game) {
	// --- Score (top-left) ---
	scoreText := fmt.Sprintf("Score: %d", g.Score)
	ebitenutil.DebugPrintAt(screen, scoreText, 20, 20)

	// High score
	if g.HighScore > 0 {
		hsText := fmt.Sprintf("Best: %d", g.HighScore)
		ebitenutil.DebugPrintAt(screen, hsText, 20, 45)
	}

	// --- Timer (top-right) ---
	remaining := g.TimeRemaining
	if remaining < 0 {
		remaining = 0
	}
	minutes := int(remaining) / 60
	seconds := int(remaining) % 60
	timerText := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	ebitenutil.DebugPrintAt(screen, timerText, ScreenWidth-150, 20)

	// --- New Game Button ---
	drawNewGameButton(screen, g)

	// --- Game Over Screen ---
	if g.GameOver {
		drawGameOver(screen, g)
	}
}

// drawNewGameButton renders the "New Game" button.
func drawNewGameButton(screen *ebiten.Image, g *Game) {
	btnX := float32(ScreenWidth/2 - 80)
	btnY := float32(15)
	btnW := float32(160)
	btnH := float32(35)

	btnColor := color.RGBA{0x2E, 0xCC, 0x71, 0xCC}
	vector.DrawFilledRect(screen, btnX, btnY, btnW, btnH, btnColor, false)
	vector.StrokeRect(screen, btnX, btnY, btnW, btnH, 2, color.White, false)
	ebitenutil.DebugPrintAt(screen, "New Game (R)", int(btnX)+30, int(btnY)+10)

	g.BtnX = int(btnX)
	g.BtnY = int(btnY)
	g.BtnW = int(btnW)
	g.BtnH = int(btnH)
}

// drawGameOver renders the game over overlay.
func drawGameOver(screen *ebiten.Image, g *Game) {
	overlayColor := color.RGBA{0x00, 0x00, 0x00, 0xAA}
	vector.DrawFilledRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), overlayColor, false)

	ebitenutil.DebugPrintAt(screen, "GAME OVER!", ScreenWidth/2-70, ScreenHeight/2-60)

	scoreText := fmt.Sprintf("Final Score: %d", g.Score)
	ebitenutil.DebugPrintAt(screen, scoreText, ScreenWidth/2-80, ScreenHeight/2+10)

	ebitenutil.DebugPrintAt(screen, "Press R or click New Game", ScreenWidth/2-110, ScreenHeight/2+60)
}

// drawBoard renders the game board with all tiles.
func drawBoard(screen *ebiten.Image, g *Game) {
	for r := 0; r < BoardRows; r++ {
		for c := 0; c < BoardCols; c++ {
			tile := g.Board.Grid[r][c]
			if tile == nil {
				continue
			}

			targetX := g.BoardOffsetX + float64(c)*g.CellSize
			targetY := g.BoardOffsetY + float64(r)*g.CellSize

			drawX := targetX
			drawY := targetY

			// Fall animation
			if tile.Falling {
				t := easeInOutCubic(g.Anim.Current.Progress)
				startY := g.BoardOffsetY + tile.FallStart*g.CellSize
				drawY = startY + (targetY-startY)*t
				if g.Anim.Current.Done {
					tile.Falling = false
				}
			}

			// Shake animation offset
			if g.Anim.IsPlaying && g.Anim.Current.Type == AnimShake {
				drawX += g.Anim.Current.ShakeOffset
			}

			// Selected tile highlight
			if g.Selected == tile {
				highlightColor := color.RGBA{0xFF, 0xFF, 0x00, 0x40}
				vector.StrokeRect(screen, float32(drawX)+2, float32(drawY)+2, float32(g.CellSize)-4, float32(g.CellSize)-4, 3, highlightColor, false)
			}

			// Hint pulse tiles
			if g.HintActive {
				for _, hp := range g.HintPositions {
					if r == hp[0] && c == hp[1] {
						pulseColor := color.RGBA{0x00, 0xFF, 0x00, 0x60}
						vector.StrokeRect(screen, float32(drawX)+2, float32(drawY)+2, float32(g.CellSize)-4, float32(g.CellSize)-4, 4, pulseColor, false)
					}
				}
			}

			// Draw tile sprite
			op := &ebiten.DrawImageOptions{}
			img := tileImages[tile.Color]
			if img != nil {
				scale := float64(tile.Scale)
				cellPad := float64(g.CellSize - 8)
				op.GeoM.Translate(drawX+4, drawY+4)
				op.GeoM.Translate(cellPad/2, cellPad/2)
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(-cellPad/2, -cellPad/2)
				op.ColorScale.ScaleAlpha(float32(tile.Alpha))
				screen.DrawImage(img, op)
			} else {
				fallbackColor := tileColors[tile.Color%len(tileColors)]
				vector.DrawFilledCircle(screen, float32(drawX)+float32(g.CellSize)/2, float32(drawY)+float32(g.CellSize)/2, float32(g.CellSize)/2-4, fallbackColor, true)
			}
		}
	}

	// Draw particles on top
	g.Particles.Draw(screen)

	// Board border
	boardW := float32(g.CellSize * BoardCols)
	boardH := float32(g.CellSize * BoardRows)
	vector.StrokeRect(screen, float32(g.BoardOffsetX)-2, float32(g.BoardOffsetY)-2, boardW+4, boardH+4, 3, color.RGBA{0xFF, 0xFF, 0xFF, 0x30}, false)
}

// isNewGameButtonClicked checks if click coordinates are within the button bounds.
func (g *Game) isNewGameButtonClicked(x, y float64) bool {
	return x >= float64(g.BtnX) && x <= float64(g.BtnX+g.BtnW) &&
		y >= float64(g.BtnY) && y <= float64(g.BtnY+g.BtnH)
}

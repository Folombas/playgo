package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type TestGame struct{ frames int }

func (g *TestGame) Update() error {
	g.frames++
	if g.frames%60 == 0 {
		fmt.Printf("Frame %d (1 sec)\n", g.frames/60)
	}
	if g.frames > 600 { // 10 seconds
		return fmt.Errorf("Test complete - 10 seconds passed")
	}
	return nil
}

func (g *TestGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 50, 150, 255})
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Puzzle GO Test\nFrames: %d (~%d sec)\nPress ESC to exit early", g.frames, g.frames/60))
}

func (g *TestGame) Layout(w, h int) (int, int) {
	return 640, 480
}

func main() {
	fmt.Println("=== PUZZLE GO WINDOW TEST ===")
	fmt.Println("Blue window should stay open for 10 seconds")
	fmt.Println("Watch the counter - if it increments, game loop works!")
	fmt.Println()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowPosition(200, 200)
	ebiten.SetWindowTitle("PUZZLE GO - TEST WINDOW")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	err := ebiten.RunGame(&TestGame{})

	fmt.Printf("\nGame loop ended after test frames. Error: %v\n", err)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}

package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SimpleGame struct {
	frame int
}

func (g *SimpleGame) Update() error {
	g.frame++
	
	// Print every second
	if g.frame%60 == 0 {
		fmt.Printf("Frame %d (~%d sec)\n", g.frame, g.frame/60)
	}
	
	// ESC to quit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("ESC pressed")
	}
	
	// Close after 600 frames (10 sec)
	if g.frame > 600 {
		return fmt.Errorf("10 second test complete!")
	}
	
	return nil
}

func (g *SimpleGame) Draw(screen *ebiten.Image) {
	// Bright green background so you can SEE the window
	screen.Fill(color.RGBA{0, 200, 0, 255})
	
	// Big text
	ebitenutil.DebugPrint(screen, "=== PUZZLE GO TEST ===\n\nIf you see GREEN WINDOW,\nEbiten is working!\n\nPress ESC to exit early\nor wait 10 seconds\n\nWindow should be VISIBLE!")
}

func (g *SimpleGame) Layout(w, h int) (int, int) {
	return 640, 480
}

func main() {
	fmt.Println("=== PUZZLE GO - EBITEM TEST ===")
	fmt.Println("A GREEN window should appear for 10 seconds")
	fmt.Println("If you don't see it, check taskbar / ALT+TAB")
	fmt.Println()
	
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowPosition(300, 200) // Center-ish position
	ebiten.SetWindowTitle("PUZZLE GO - LOOK HERE! GREEN WINDOW")
	
	err := ebiten.RunGame(&SimpleGame{})
	
	fmt.Printf("\nGame ended: %v\n", err)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	os.Exit(0)
}

package web

import (
    \ github.com/hajimehoshi/ebiten/v2\
)

type Game struct{}
func (g *Game) Update() error { return nil }
func (g *Game) Draw(screen *ebiten.Image)   {}
func (g *Game) Layout(w, h int) (int, int) { return 800, 600 }
func main() {
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle(\Match-3 Web\)
    ebiten.RunGame(&Game{})
}

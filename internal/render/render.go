package render

import (
    \ image/color\
    \github.com/hajimehoshi/ebiten/v2\
)

type Renderer struct{}
func NewRenderer() *Renderer { return &Renderer{} }
func (r *Renderer) DrawBoard(screen *ebiten.Image, b interface{}) {}

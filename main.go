package main

import (
    \ fmt\
    \github.com/hajimehoshi/ebiten/v2\
    \github.com/Folombas/playgo/match3-game/internal/game\
    \github.com/Folombas/playgo/match3-game/internal/input\
    \github.com/Folombas/playgo/match3-game/internal/render\
)

type Game struct {
    gameState *game.Game
    input     *input.InputHandler
    renderer  *render.Renderer
}

func main() {
    g := &Game{
        gameState: game.NewGame(),
        input:     &input.InputHandler{},
        renderer:  render.NewRenderer(),
    }
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle(\Match-3\)
    ebiten.RunGame(g)
}

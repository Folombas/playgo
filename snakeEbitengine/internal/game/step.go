package game

import (
	"image/color"

	"snake/internal/types"
)

func (g *Game) step() {
	head := g.snake[0]
	newHead := types.Vec{head.X + g.dir.X, head.Y + g.dir.Y}

	if newHead.X < 0 || newHead.X >= types.GridW || newHead.Y < 0 || newHead.Y >= types.GridH {
		g.triggerExplosion(head, true)
		return
	}
	if !g.ghostModeActive() {
		for _, s := range g.snake {
			if s == newHead {
				g.triggerExplosion(newHead, true)
				return
			}
		}
	}

	g.snake = append([]types.Vec{newHead}, g.snake...)

	ateFruit := false
	if newHead.X == g.fruitX && newHead.Y == g.fruitY {
		ateFruit = true
		switch g.fruitType {
		case types.FruitApple:
			g.score += 1
			g.health = minInt(types.MaxHealth, g.health+25)
		case types.FruitStrawberry:
			g.score += 2
			g.health = minInt(types.MaxHealth, g.health+40)
		case types.FruitOrange:
			g.score += 3
			g.health = minInt(types.MaxHealth, g.health+35)
		case types.FruitBanana:
			g.score += 1
			g.health = minInt(types.MaxHealth, g.health+20)
		case types.FruitPineapple:
			g.score += 4
			g.health = minInt(types.MaxHealth, g.health+45)
		}
		g.placeFruit()
		g.spawnBombRandom()
		g.spawnIce()
		g.spawnGhost()
		g.sndHeal.Rewind()
		g.sndHeal.Play()
		g.addParticles(float64(newHead.X*types.TileSize+types.TileSize/2), float64(newHead.Y*types.TileSize+types.TileSize/2), 25, color.RGBA{50, 255, 80, 255}, true)
	}

	if g.frozenTimer > 0 && ateFruit {
		g.snake = g.snake[:len(g.snake)-1-1]
	} else if !ateFruit {
		g.snake = g.snake[:len(g.snake)-1]
	}

	for i := 0; i < len(g.bombs); i++ {
		if g.bombs[i].X == newHead.X && g.bombs[i].Y == newHead.Y {
			g.health -= 35
			g.triggerExplosion(newHead, g.health <= 0)
			g.bombs = append(g.bombs[:i], g.bombs[i+1:]...)
			return
		}
	}

	if g.iceActive && newHead.X == g.ice.X && newHead.Y == g.ice.Y {
		g.frozenTimer = 5.0
		g.iceActive = false
		g.addParticles(float64(newHead.X*types.TileSize+types.TileSize/2), float64(newHead.Y*types.TileSize+types.TileSize/2), 50, color.RGBA{100, 200, 255, 255}, true)
		g.sndHeal.Rewind()
		g.sndHeal.Play()
	}

	if g.ghostActive && newHead.X == g.ghostX && newHead.Y == g.ghostY {
		g.ghostModeTimer = 5.0
		g.ghostActive = false
		g.sndGhost.Rewind()
		g.sndGhost.Play()
		g.addParticles(float64(newHead.X*types.TileSize+types.TileSize/2), float64(newHead.Y*types.TileSize+types.TileSize/2), 60, color.RGBA{200, 200, 255, 200}, true)
	}

	g.addParticles(float64(newHead.X*types.TileSize+types.TileSize/2), float64(newHead.Y*types.TileSize+types.TileSize/2), 2, color.RGBA{0, 180, 220, 140}, false)
	if g.frozenTimer > 0 {
		g.addParticles(float64(newHead.X*types.TileSize+types.TileSize/2), float64(newHead.Y*types.TileSize+types.TileSize/2), 1, color.RGBA{150, 220, 255, 180}, true)
	}
}
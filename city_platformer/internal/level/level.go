package level

import (
	"math/rand"

	"city_platformer/internal/entity"
)

// Generator creates endless runner segments.
type Generator struct {
	rng    *rand.Rand
	ground float64
	speed  float64
	timer  float64
}

// NewGenerator creates a generator.
func NewGenerator(seed int64, groundY float64) *Generator {
	return &Generator{
		rng:    rand.New(rand.NewSource(seed)),
		ground: groundY,
		speed:  200,
	}
}

// SpawnObstacle creates a new obstacle.
func (g *Generator) SpawnObstacle(screenW float64) *entity.Obstacle {
	typ := g.rng.Intn(3)
	x := screenW + 50

	if typ == 0 {
		// Ground spike
		return &entity.Obstacle{
			X: x, Y: g.ground - 32, W: 32, H: 32, Type: 0,
		}
	} else if typ == 1 {
		// Flying obstacle
		return &entity.Obstacle{
			X: x, Y: g.ground - 100 - g.rng.Float64()*60, W: 40, H: 30, Type: 1,
		}
	}
	// Tall obstacle
	return &entity.Obstacle{
		X: x, Y: g.ground - 64, W: 32, H: 64, Type: 0,
	}
}

// SpawnCoin creates a coin at a random position.
func (g *Generator) SpawnCoin(screenW float64) *entity.Coin {
	return &entity.Coin{
		X: screenW + 50 + g.rng.Float64()*200,
		Y: g.ground - 40 - g.rng.Float64()*150,
		W: 16, H: 16,
	}
}

// Speed returns current game speed.
func (g *Generator) Speed() float64 {
	return g.speed
}

// IncreaseSpeed ramps up difficulty.
func (g *Generator) IncreaseSpeed() {
	g.speed += 10
}

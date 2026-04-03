package level

import (
	"math/rand"

	"city_platformer/internal/entity"
)

// Chunk represents a segment of the level.
type Chunk struct {
	X         float64
	Width     float64
	Platforms []*entity.Platform
	Enemies   []*entity.Enemy
	Items     []*entity.Item
	GroundY   float64
	HasPit    bool
}

// Generator generates level chunks procedurally.
type Generator struct {
	rng         *rand.Rand
	chunkWidth  float64
	groundY     float64
	difficulty  float64
	seed        int64
	lastChunkX  float64
}

// NewGenerator creates a new level generator.
func NewGenerator(seed int64) *Generator {
	return &Generator{
		rng:        rand.New(rand.NewSource(seed)),
		chunkWidth: 800,
		groundY:    500,
		difficulty: 1.0,
		seed:       seed,
	}
}

// Generate creates a new chunk at the given X position.
func (g *Generator) Generate(chunkX float64) *Chunk {
	chunk := &Chunk{
		X:       chunkX,
		Width:   g.chunkWidth,
		GroundY: g.groundY,
	}

	// 20% chance of a pit
	chunk.HasPit = g.rng.Float64() < 0.2

	// Generate ground platforms (skip if pit)
	if !chunk.HasPit {
		// Solid ground
		ground := entity.NewPlatform(chunkX, g.groundY, g.chunkWidth, 200)
		chunk.Platforms = append(chunk.Platforms, ground)
	} else {
		// Ground with gap
		pitStart := chunkX + g.chunkWidth*0.3
		pitEnd := chunkX + g.chunkWidth*0.7

		left := entity.NewPlatform(chunkX, g.groundY, pitStart-chunkX, 200)
		right := entity.NewPlatform(pitEnd, g.groundY, chunkX+g.chunkWidth-pitEnd, 200)
		chunk.Platforms = append(chunk.Platforms, left, right)
	}

	// Generate floating platforms
	numPlatforms := g.rng.Intn(3) + 1
	for i := 0; i < numPlatforms; i++ {
		px := chunkX + g.rng.Float64()*(g.chunkWidth-100)
		py := g.groundY - 80 - g.rng.Float64()*200
		pw := 64 + g.rng.Float64()*128
		ph := 16.0

		plat := entity.NewPlatform(px, py, pw, ph)
		chunk.Platforms = append(chunk.Platforms, plat)
	}

	// Generate enemies (more likely with higher difficulty)
	enemyChance := 0.3 + g.difficulty*0.1
	if g.rng.Float64() < enemyChance && !chunk.HasPit {
		numEnemies := g.rng.Intn(2) + 1
		for i := 0; i < numEnemies; i++ {
			ex := chunkX + 100 + g.rng.Float64()*(g.chunkWidth-200)
			et := entity.EnemyType(g.rng.Intn(3)) // slime, bat, snail
			patrolRange := 100 + g.rng.Float64()*100
			enemy := entity.NewEnemy(ex, g.groundY-32, et, ex-patrolRange, ex+patrolRange)
			chunk.Enemies = append(chunk.Enemies, enemy)
		}
	}

	// Generate items
	numCoins := g.rng.Intn(5) + 2
	for i := 0; i < numCoins; i++ {
		ix := chunkX + g.rng.Float64()*g.chunkWidth
		iy := g.groundY - 40 - g.rng.Float64()*250
		coin := entity.NewItem(ix, iy, entity.ItemCoin, 10)
		chunk.Items = append(chunk.Items, coin)
	}

	// Gems (rarer)
	if g.rng.Float64() < 0.4 {
		gx := chunkX + g.rng.Float64()*g.chunkWidth
		gy := g.groundY - 100 - g.rng.Float64()*200
		gem := entity.NewItem(gx, gy, entity.ItemGem, 25)
		chunk.Items = append(chunk.Items, gem)
	}

	// Stars (rare, high value)
	if g.rng.Float64() < 0.2 {
		sx := chunkX + g.rng.Float64()*g.chunkWidth
		sy := g.groundY - 150 - g.rng.Float64()*150
		star := entity.NewItem(sx, sy, entity.ItemStar, 100)
		chunk.Items = append(chunk.Items, star)
	}

	// Hearts (healing)
	if g.rng.Float64() < 0.15 {
		hx := chunkX + g.rng.Float64()*g.chunkWidth
		hy := g.groundY - 60 - g.rng.Float64()*180
		heart := entity.NewItem(hx, hy, entity.ItemHeart, 1)
		chunk.Items = append(chunk.Items, heart)
	}

	g.lastChunkX = chunkX
	return chunk
}

// IncreaseDifficulty ramps up the challenge.
func (g *Generator) IncreaseDifficulty() {
	g.difficulty += 0.2
}

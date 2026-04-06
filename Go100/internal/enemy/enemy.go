package enemy

import (
	"fmt"
	"image/color"
	"math"

	"towerdefense/internal/config"
	"towerdefense/internal/map"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PathNode alias
type PathNode = gamemap.PathNode

// Enemy represents an enemy unit
type Enemy struct {
	Type       int
	X, Y       float64
	HP         float64
	MaxHP      float64
	Speed      float64
	BaseSpeed  float64
	Reward     int
	Progress   float64 // 0.0 to 1.0 along path
	Alive      bool
	SlowTimer  int
	SlowFactor float64
	HitFlash   int
}

// EnemyStats holds stats for each enemy type
type EnemyStats struct {
	Name   string
	HP     float64
	Speed  float64
	Reward int
	Color  color.Color
	Size   float64
}

var enemyStats = []EnemyStats{
	{
		Name: "Basic", HP: 30, Speed: 1.5, Reward: 10,
		Color: color.RGBA{255, 100, 100, 255}, Size: 10,
	},
	{
		Name: "Fast", HP: 20, Speed: 3.0, Reward: 15,
		Color: color.RGBA{255, 255, 100, 255}, Size: 8,
	},
	{
		Name: "Tank", HP: 100, Speed: 0.8, Reward: 25,
		Color: color.RGBA{150, 150, 150, 255}, Size: 14,
	},
	{
		Name: "Boss", HP: 500, Speed: 0.5, Reward: 100,
		Color: color.RGBA{200, 50, 200, 255}, Size: 18,
	},
	{
		Name: "Swarm", HP: 10, Speed: 2.0, Reward: 5,
		Color: color.RGBA{100, 255, 100, 255}, Size: 6,
	},
}

// NewEnemy creates a new enemy
func NewEnemy(enemyType int) *Enemy {
	stats := enemyStats[enemyType]
	return &Enemy{
		Type:      enemyType,
		HP:        stats.HP,
		MaxHP:     stats.HP,
		Speed:     stats.Speed,
		BaseSpeed: stats.Speed,
		Reward:    stats.Reward,
		Alive:     true,
		Progress:  0,
	}
}

// Update updates enemy position along path
func (e *Enemy) Update(pathPositions []PathPosition) {
	if !e.Alive {
		return
	}

	// Apply slow effect
	if e.SlowTimer > 0 {
		e.Speed = e.BaseSpeed * e.SlowFactor
		e.SlowTimer--
	} else {
		e.Speed = e.BaseSpeed
	}

	// Move along path
	e.Progress += e.Speed * 0.002 // Normalize progress

	if e.Progress >= 1.0 {
		e.Alive = false // Reached the end
	}

	// Update position from path
	if len(pathPositions) > 0 {
		idx := int(e.Progress * float64(len(pathPositions)-1))
		if idx >= len(pathPositions) {
			idx = len(pathPositions) - 1
		}
		e.X = pathPositions[idx].X
		e.Y = pathPositions[idx].Y
	}

	if e.HitFlash > 0 {
		e.HitFlash--
	}
}

// TakeDamage applies damage
func (e *Enemy) TakeDamage(dmg float64) {
	e.HP -= dmg
	e.HitFlash = 5
	if e.HP <= 0 {
		e.Alive = false
	}
}

// ApplySlow applies slow debuff
func (e *Enemy) ApplySlow(factor float64, duration int) {
	e.SlowFactor = factor
	e.SlowTimer = duration
}

// IsAlive returns true if enemy is alive
func (e *Enemy) IsAlive() bool {
	return e.Alive
}

// Draw renders the enemy
func (e *Enemy) Draw(screen *ebiten.Image) {
	if !e.Alive {
		return
	}

	stats := enemyStats[e.Type]

	// Shadow
	vector.DrawFilledCircle(screen, e.X+2, e.Y+2, stats.Size, color.RGBA{0, 0, 0, 100}, true)

	// Body
	drawColor := stats.Color
	if e.HitFlash > 0 {
		drawColor = color.White
	} else if e.SlowTimer > 0 {
		// Blue tint when slowed
		drawColor = color.RGBA{150, 200, 255, 255}
	}

	vector.DrawFilledCircle(screen, e.X, e.Y, stats.Size, drawColor, true)
	
	// Border
	vector.StrokeCircle(screen, e.X, e.Y, stats.Size, 2, color.RGBA{0, 0, 0, 200}, true)

	// HP bar
	if e.HP < e.MaxHP {
		barWidth := stats.Size * 2
		barHeight := float64(4)
		hpRatio := e.HP / e.MaxHP

		// Background
		vector.DrawFilledRect(screen, e.X-barWidth/2, e.Y-stats.Size-8, barWidth, barHeight, color.RGBA{50, 50, 50, 255}, true)
		
		// HP fill
		hpColor := color.RGBA{0, 255, 0, 255}
		if hpRatio < 0.3 {
			hpColor = color.RGBA{255, 0, 0, 255}
		}
		vector.DrawFilledRect(screen, e.X-barWidth/2, e.Y-stats.Size-8, barWidth*hpRatio, barHeight, hpColor, true)
	}
}

// GetInfo returns enemy info
func (e *Enemy) GetInfo() string {
	stats := enemyStats[e.Type]
	return fmt.Sprintf("%s HP:%.0f/%.0f", stats.Name, e.HP, e.MaxHP)
}

// PathPosition stores precalculated path positions
type PathPosition struct {
	X, Y float64
}

// PrecalculatePathPositions precalculates all positions along the path
func PrecalculatePathPositions(path []PathNode) []PathPosition {
	positions := make([]PathPosition, 0)
	
	for i := 0; i < len(path)-1; i++ {
		p1 := path[i]
		p2 := path[i+1]
		
		// Interpolate 10 points per segment
		for t := 0; t < 10; t++ {
			frac := float64(t) / 10.0
			x := (p1.X + (p2.X-p1.X)*frac) * float64(config.TileSize) + float64(config.GridOffsetX) + float64(config.TileSize)/2
			y := (p1.Y + (p2.Y-p1.Y)*frac) * float64(config.TileSize) + float64(config.GridOffsetY) + float64(config.TileSize)/2
			positions = append(positions, PathPosition{X: x, Y: y})
		}
	}
	
	return positions
}

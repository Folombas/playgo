package tower

import (
	"fmt"
	"image/color"
	"math"

	"towerdefense/internal/config"
	"towerdefense/internal/enemy"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Tower represents a defensive tower
type Tower struct {
	Type        int
	Level       int
	X, Y        float64 // Screen coordinates
	GridX, GridY int    // Grid coordinates
	Damage      float64
	Range       float64
	FireRate    int // Frames between shots
	Cooldown    int
	TotalSpent  int // Gold spent on this tower + upgrades
	Targets     []*enemy.Enemy
	Angle       float64 // Rotation angle for visual
}

// TowerStats holds stats for each tower type
type TowerStats struct {
	Name     string
	Cost     int
	Damage   float64
	Range    float64
	FireRate int
	Color    color.Color
	Desc     string
}

var towerStats = []TowerStats{
	{
		Name: "Basic", Cost: 50, Damage: 10, Range: 120, FireRate: 30,
		Color: color.RGBA{100, 200, 100, 255},
		Desc: "Balanced tower",
	},
	{
		Name: "Sniper", Cost: 100, Damage: 50, Range: 250, FireRate: 90,
		Color: color.RGBA{100, 100, 255, 255},
		Desc: "Long range, high damage",
	},
	{
		Name: "Slow", Cost: 75, Damage: 5, Range: 100, FireRate: 60,
		Color: color.RGBA{100, 200, 255, 255},
		Desc: "Slows enemies",
	},
	{
		Name: "Splash", Cost: 125, Damage: 20, Range: 100, FireRate: 60,
		Color: color.RGBA{255, 150, 50, 255},
		Desc: "Area damage",
	},
	{
		Name: "Laser", Cost: 150, Damage: 2, Range: 150, FireRate: 1,
		Color: color.RGBA{255, 50, 50, 255},
		Desc: "Continuous beam",
	},
}

// NewTower creates a new tower
func NewTower(towerType, gridX, gridY int) *Tower {
	stats := towerStats[towerType]
	return &Tower{
		Type:     towerType,
		Level:    1,
		GridX:    gridX,
		GridY:    gridY,
		X:        float64(gridX*config.TileSize + config.GridOffsetX + config.TileSize/2),
		Y:        float64(gridY*config.TileSize + config.GridOffsetY + config.TileSize/2),
		Damage:   stats.Damage,
		Range:    stats.Range,
		FireRate: stats.FireRate,
		Cooldown: 0,
		TotalSpent: stats.Cost,
	}
}

// GetUpgradeCost returns the cost to upgrade
func (t *Tower) GetUpgradeCost() int {
	if t.Level >= 3 {
		return -1 // Max level
	}
	return towerStats[t.Type].Cost * t.Level
}

// Upgrade upgrades the tower
func (t *Tower) Upgrade() bool {
	if t.Level >= 3 {
		return false
	}
	t.Level++
	t.Damage *= 1.5
	t.Range *= 1.1
	if t.FireRate > 10 {
		t.FireRate = int(float64(t.FireRate) * 0.85)
	}
	t.TotalSpent += t.GetUpgradeCost()
	return true
}

// GetSellValue returns gold received when selling
func (t *Tower) GetSellValue() int {
	return int(float64(t.TotalSpent) * 0.6)
}

// Update updates tower logic
func (t *Tower) Update(enemies []*enemy.Enemy) {
	t.Cooldown--

	// Find target (prioritize furthest along path)
	t.Targets = nil
	bestProgress := -1.0
	var bestTarget *enemy.Enemy

	for _, e := range enemies {
		dist := math.Hypot(e.X-t.X, e.Y-t.Y)
		if dist <= t.Range && e.IsAlive() {
			if e.Progress > bestProgress {
				bestProgress = e.Progress
				bestTarget = e
			}
		}
	}

	if bestTarget != nil {
		t.Targets = append(t.Targets, bestTarget)
		t.Angle = math.Atan2(bestTarget.Y-t.Y, bestTarget.X-t.X)

		if t.Cooldown <= 0 {
			t.Cooldown = t.FireRate
			// Fire!
			return // Signal to game to create projectile
		}
	}
}

// CanFire returns true if tower can fire
func (t *Tower) CanFire() bool {
	return t.Cooldown <= 0 && len(t.Targets) > 0
}

// GetTarget returns current target
func (t *Tower) GetTarget() *enemy.Enemy {
	if len(t.Targets) == 0 {
		return nil
	}
	return t.Targets[0]
}

// Draw renders the tower
func (t *Tower) Draw(screen *ebiten.Image) {
	stats := towerStats[t.Type]

	// Base
	baseSize := float32(18 + t.Level*2)
	vector.DrawFilledCircle(screen, t.X, t.Y, float64(baseSize), stats.Color, true)
	
	// Border
	vector.StrokeCircle(screen, t.X, t.Y, float64(baseSize), 2, color.RGBA{255, 255, 255, 200}, true)

	// Turret barrel
	barrelLen := float64(baseSize) + 8
	barrelWidth := float32(4 + t.Level)
	endX := t.X + math.Cos(t.Angle)*barrelLen
	endY := t.Y + math.Sin(t.Angle)*barrelLen
	
	// Calculate perpendicular for width
	perpAngle := t.Angle + math.Pi/2
	dx := math.Cos(perpAngle) * float64(barrelWidth)
	dy := math.Sin(perpAngle) * float64(barrelWidth)
	
	points := []float32{
		float32(endX), float32(endY),
		float32(endX - dx), float32(endY - dy),
		float32(t.X - dx), float32(t.Y - dy),
		float32(t.X + dx), float32(t.Y + dy),
	}
	vector.DrawFilledPolygon(screen, points, stats.Color, true)

	// Level indicator
	levelText := fmt.Sprintf("%d", t.Level)
	// Draw level as small dots
	for i := 0; i < t.Level; i++ {
		lx := t.X - float64(t.Level-1)*4 + float64(i)*8
		ly := t.Y + float64(baseSize) + 6
		vector.DrawFilledCircle(screen, lx, ly, 2, color.White, true)
	}

	// Range indicator when hovered (handled by game)
}

// DrawRange draws the tower's range circle
func (t *Tower) DrawRange(screen *ebiten.Image) {
	op := &ebiten.DrawLinesOptions{}
	op.Width = 2
	vector.StrokeCircle(screen, t.X, t.Y, t.Range, op, color.RGBA{255, 255, 255, 100}, true)
}

// GetInfo returns tower info string
func (t *Tower) GetInfo() string {
	stats := towerStats[t.Type]
	return fmt.Sprintf("%s Lv.%d\nDMG: %.0f RNG: %.0f", stats.Name, t.Level, t.Damage, t.Range)
}

// GetType returns tower type
func GetType(towerType int) TowerStats {
	return towerStats[towerType]
}

package gamemap

import (
	"image/color"
	"math"

	"towerdefense/internal/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Tile types
const (
	TileGrass = iota
	TilePath
	TileBase
	TileTower
)

// Map represents the game map
type Map struct {
	Tiles    [][]int
	Path     []PathNode
	TowerSpots []TowerSpot
}

type PathNode struct {
	X, Y float64
}

type TowerSpot struct {
	X, Y   int
	Active bool
}

// NewMap creates a new game map with a winding path
func NewMap() *Map {
	m := &Map{
		Tiles:    make([][]int, config.GridHeight),
		Path:     make([]PathNode, 0),
		TowerSpots: make([]TowerSpot, 0),
	}

	// Initialize with grass
	for y := 0; y < config.GridHeight; y++ {
		m.Tiles[y] = make([]int, config.GridWidth)
		for x := 0; x < config.GridWidth; x++ {
			m.Tiles[y][x] = TileGrass
		}
	}

	// Create winding path
	m.generatePath()

	// Mark path tiles
	for _, p := range m.Path {
		gx := int(p.X)
		gy := int(p.Y)
		if gx >= 0 && gx < config.GridWidth && gy >= 0 && gy < config.GridHeight {
			m.Tiles[gy][gx] = TilePath
		}
	}

	// Generate tower placement spots
	m.generateTowerSpots()

	return m
}

func (m *Map) generatePath() {
	// Winding path from left to right
	pathPoints := [][2]int{
		{-1, 2},   // Entry
		{3, 2},
		{3, 6},
		{8, 6},
		{8, 3},
		{13, 3},
		{13, 9},
		{6, 9},
		{6, 11},
		{17, 11},
		{17, 5},
		{config.GridWidth, 5}, // Exit
	}

	// Convert to smooth path with all intermediate points
	for i := 0; i < len(pathPoints)-1; i++ {
		x1, y1 := pathPoints[i][0], pathPoints[i][1]
		x2, y2 := pathPoints[i+1][0], pathPoints[i+1][1]

		// Interpolate
		dx := float64(x2 - x1)
		dy := float64(y2 - y1)
		dist := math.Sqrt(dx*dx + dy*dy)
		steps := int(dist)

		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			x := x1 + int(float64(x2-x1)*t)
			y := y1 + int(float64(y2-y1)*t)
			m.Path = append(m.Path, PathNode{
				X: float64(x),
				Y: float64(y),
			})
		}
	}
}

func (m *Map) generateTowerSpots() {
	// Find valid tower placement spots (grass tiles near path)
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			if m.Tiles[y][x] == TileGrass {
				// Check if near path
				nearPath := false
				for _, p := range m.Path {
					px, py := int(p.X), int(p.Y)
					if math.Abs(float64(x-px)) <= 1 && math.Abs(float64(y-py)) <= 1 {
						nearPath = true
						break
					}
				}
				if nearPath {
					m.TowerSpots = append(m.TowerSpots, TowerSpot{
						X: x, Y: y, Active: false,
					})
				}
			}
		}
	}
}

// Draw renders the map
func (m *Map) Draw(screen *ebiten.Image) {
	// Draw tiles
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			sx := float32(x*config.TileSize + config.GridOffsetX)
			sy := float32(y*config.TileSize + config.GridOffsetY)
			ts := float32(config.TileSize)

			switch m.Tiles[y][x] {
			case TileGrass:
				vector.DrawFilledRect(screen, sx, sy, ts, ts, color.RGBA{40, 80, 40, 255}, false)
				vector.DrawFilledRect(screen, sx+2, sy+2, ts-4, ts-4, color.RGBA{50, 90, 50, 255}, false)
			case TilePath:
				vector.DrawFilledRect(screen, sx, sy, ts, ts, color.RGBA{180, 160, 120, 255}, false)
				vector.DrawFilledRect(screen, sx+4, sy+4, ts-8, ts-8, color.RGBA{160, 140, 100, 255}, false)
			case TileTower:
				vector.DrawFilledRect(screen, sx, sy, ts, ts, color.RGBA{80, 80, 100, 255}, false)
			}
		}
	}

	// Draw base (start point)
	if len(m.Path) > 0 {
		sx := float32(m.Path[0].X*float64(config.TileSize) + float64(config.GridOffsetX) + float64(config.TileSize)/2)
		sy := float32(m.Path[0].Y*float64(config.TileSize) + float64(config.GridOffsetY) + float64(config.TileSize)/2)
		vector.DrawFilledCircle(screen, sx, sy, 14, color.RGBA{0, 255, 0, 255}, false)
		vector.StrokeCircle(screen, sx, sy, 14, 2, color.White, false)
	}

	// Draw exit (end point)
	if len(m.Path) > 0 {
		last := m.Path[len(m.Path)-1]
		sx := float32(last.X*float64(config.TileSize) + float64(config.GridOffsetX) + float64(config.TileSize)/2)
		sy := float32(last.Y*float64(config.TileSize) + float64(config.GridOffsetY) + float64(config.TileSize)/2)
		vector.DrawFilledCircle(screen, sx, sy, 14, color.RGBA{255, 0, 0, 255}, false)
		vector.StrokeCircle(screen, sx, sy, 14, 2, color.White, false)
	}
}

// CanPlaceTower checks if a tower can be placed at grid position
func (m *Map) CanPlaceTower(gx, gy int) bool {
	if gx < 0 || gx >= config.GridWidth || gy < 0 || gy >= config.GridHeight {
		return false
	}
	return m.Tiles[gy][gx] == TileGrass
}

// PlaceTower marks a spot as occupied
func (m *Map) PlaceTower(gx, gy int) {
	for i := range m.TowerSpots {
		if m.TowerSpots[i].X == gx && m.TowerSpots[i].Y == gy {
			m.TowerSpots[i].Active = true
			break
		}
	}
	m.Tiles[gy][gx] = TileTower
}

// GetPathProgress returns interpolated position along path
func (m *Map) GetPositionAtProgress(progress float64) (float64, float64) {
	if len(m.Path) == 0 || progress <= 0 {
		return float64(m.Path[0].X) * float64(config.TileSize), float64(m.Path[0].Y) * float64(config.TileSize)
	}

	totalSegments := float64(len(m.Path) - 1)
	segment := progress * totalSegments
	idx := int(segment)
	if idx >= len(m.Path)-1 {
		last := m.Path[len(m.Path)-1]
		return last.X * float64(config.TileSize), last.Y * float64(config.TileSize)
	}

	t := segment - float64(idx)
	p1 := m.Path[idx]
	p2 := m.Path[idx+1]

	x := (p1.X + (p2.X-p1.X)*t) * float64(config.TileSize)
	y := (p1.Y + (p2.Y-p1.Y)*t) * float64(config.TileSize)

	return x, y
}

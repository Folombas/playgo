package dungeon

import (
	"math/rand"
	"dungeon_crawler/internal/config"
)

// Room represents a room in the dungeon
type Room struct {
	X, Y, W, H int
	CenterX    int
	CenterY    int
}

// Dungeon represents a generated dungeon floor
type Dungeon struct {
	Tiles    [][]int // 2D grid: 0=floor, 1=wall, 2=door, 3=stairs, 4=spikes, 5=water
	Rooms    []*Room
	Entities map[string][]EntityPosition // entityType -> positions
}

// EntityPosition stores position of an entity
type EntityPosition struct {
	X, Y int
}

// NewDungeon creates a new dungeon floor
func NewDungeon(floorNum int, rng *rand.Rand) *Dungeon {
	d := &Dungeon{
		Tiles:    make([][]int, config.MapHeight),
		Rooms:    make([]*Room, 0),
		Entities: make(map[string][]EntityPosition),
	}

	// Initialize with walls
	for y := 0; y < config.MapHeight; y++ {
		d.Tiles[y] = make([]int, config.MapWidth)
		for x := 0; x < config.MapWidth; x++ {
			d.Tiles[y][x] = config.TileWall
		}
	}

	// Generate rooms
	d.generateRooms(rng)

	// Connect rooms with corridors
	d.connectRooms(rng)

	// Place special items
	d.placeStairs(rng)
	d.placeSpikes(rng, floorNum)
	d.placeWater(rng, floorNum)

	return d
}

// generateRooms creates random rooms
func (d *Dungeon) generateRooms(rng *rand.Rand) {
	attempts := 0
	maxAttempts := 100

	for len(d.Rooms) < config.MaxRooms && attempts < maxAttempts {
		attempts++

		w := config.MinRoomSize + rng.Intn(config.MaxRoomSize-config.MinRoomSize+1)
		h := config.MinRoomSize + rng.Intn(config.MaxRoomSize-config.MinRoomSize+1)
		x := 1 + rng.Intn(config.MapWidth-w-2)
		y := 1 + rng.Intn(config.MapHeight-h-2)

		newRoom := &Room{
			X: x, Y: y, W: w, H: h,
			CenterX: x + w/2,
			CenterY: y + h/2,
		}

		// Check overlap
		overlap := false
		for _, room := range d.Rooms {
			if newRoom.X-1 < room.X+room.W && newRoom.X+newRoom.W+1 > room.X &&
				newRoom.Y-1 < room.Y+room.H && newRoom.Y+newRoom.H+1 > room.Y {
				overlap = true
				break
			}
		}

		if !overlap {
			d.carveRoom(newRoom)
			d.Rooms = append(d.Rooms, newRoom)
		}
	}
}

// carveRoom carves out a room in the grid
func (d *Dungeon) carveRoom(room *Room) {
	for y := room.Y; y < room.Y+room.H; y++ {
		for x := room.X; x < room.X+room.W; x++ {
			d.Tiles[y][x] = config.TileFloor
		}
	}
}

// connectRooms connects all rooms with L-shaped corridors
func (d *Dungeon) connectRooms(rng *rand.Rand) {
	for i := 1; i < len(d.Rooms); i++ {
		d.createCorridor(d.Rooms[i-1], d.Rooms[i], rng)
	}

	// Add some random extra connections for loops
	if len(d.Rooms) > 3 {
		for i := 0; i < 2; i++ {
			a := rng.Intn(len(d.Rooms))
			b := rng.Intn(len(d.Rooms))
			if a != b {
				d.createCorridor(d.Rooms[a], d.Rooms[b], rng)
			}
		}
	}
}

// createCorridor creates an L-shaped corridor between two rooms
func (d *Dungeon) createCorridor(a, b *Room, rng *rand.Rand) {
	x := a.CenterX
	y := a.CenterY
	targetX := b.CenterX
	targetY := b.CenterY

	// Randomly choose to go horizontal first or vertical first
	if rng.Intn(2) == 0 {
		// Horizontal first
		d.carveHorizontal(x, targetX, y)
		d.carveVertical(y, targetY, targetX)
	} else {
		// Vertical first
		d.carveVertical(y, targetY, x)
		d.carveHorizontal(x, targetX, targetY)
	}
}

// carveHorizontal carves a horizontal tunnel
func (d *Dungeon) carveHorizontal(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if y >= 0 && y < config.MapHeight && x >= 0 && x < config.MapWidth {
			d.Tiles[y][x] = config.TileFloor
			// Make corridor wider for better gameplay
			if y-1 >= 0 {
				d.Tiles[y-1][x] = config.TileFloor
			}
			if y+1 < config.MapHeight {
				d.Tiles[y+1][x] = config.TileFloor
			}
		}
	}
}

// carveVertical carves a vertical tunnel
func (d *Dungeon) carveVertical(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if y >= 0 && y < config.MapHeight && x >= 0 && x < config.MapWidth {
			d.Tiles[y][x] = config.TileFloor
			// Make corridor wider
			if x-1 >= 0 {
				d.Tiles[y][x-1] = config.TileFloor
			}
			if x+1 < config.MapWidth {
				d.Tiles[y][x+1] = config.TileFloor
			}
		}
	}
}

// placeStairs places stairs to next floor in the last room
func (d *Dungeon) placeStairs(rng *rand.Rand) {
	if len(d.Rooms) == 0 {
		return
	}

	lastRoom := d.Rooms[len(d.Rooms)-1]
	sx := lastRoom.X + 1 + rng.Intn(lastRoom.W-2)
	sy := lastRoom.Y + 1 + rng.Intn(lastRoom.H-2)
	d.Tiles[sy][sx] = config.TileStairs
}

// placeSpikes places spike traps randomly
func (d *Dungeon) placeSpikes(rng *rand.Rand, floorNum int) {
	numSpikes := 3 + floorNum*2 // More spikes on deeper floors
	if numSpikes > 15 {
		numSpikes = 15
	}

	for i := 0; i < numSpikes; i++ {
		roomIdx := 1 + rng.Intn(len(d.Rooms)-1) // Skip first room (spawn)
		if roomIdx >= len(d.Rooms) {
			continue
		}
		room := d.Rooms[roomIdx]
		sx := room.X + 1 + rng.Intn(room.W-2)
		sy := room.Y + 1 + rng.Intn(room.H-2)
		if d.Tiles[sy][sx] == config.TileFloor {
			d.Tiles[sy][sx] = config.TileSpikes
		}
	}
}

// placeWater places water hazards randomly
func (d *Dungeon) placeWater(rng *rand.Rand, floorNum int) {
	numWater := 2 + floorNum
	if numWater > 10 {
		numWater = 10
	}

	for i := 0; i < numWater; i++ {
		roomIdx := 2 + rng.Intn(len(d.Rooms)-2)
		if roomIdx >= len(d.Rooms) {
			continue
		}
		room := d.Rooms[roomIdx]
		
		// Create small water pool
		px := room.X + 1 + rng.Intn(room.W-3)
		py := room.Y + 1 + rng.Intn(room.H-3)
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				tx, ty := px+dx, py+dy
				if ty >= 0 && ty < config.MapHeight && tx >= 0 && tx < config.MapWidth {
					if d.Tiles[ty][tx] == config.TileFloor {
						d.Tiles[ty][tx] = config.TileWater
					}
				}
			}
		}
	}
}

// IsWalkable checks if a tile is walkable
func (d *Dungeon) IsWalkable(x, y int) bool {
	if x < 0 || x >= config.MapWidth || y < 0 || y >= config.MapHeight {
		return false
	}
	tile := d.Tiles[y][x]
	return tile == config.TileFloor || tile == config.TileDoor || tile == config.TileStairs || tile == config.TileSpikes
}

// GetTile returns tile type at position
func (d *Dungeon) GetTile(x, y int) int {
	if x < 0 || x >= config.MapWidth || y < 0 || y >= config.MapHeight {
		return config.TileWall
	}
	return d.Tiles[y][x]
}

// SetTile sets tile type at position
func (d *Dungeon) SetTile(x, y, tileType int) {
	if x >= 0 && x < config.MapWidth && y >= 0 && y < config.MapHeight {
		d.Tiles[y][x] = tileType
	}
}

// AddEntity adds an entity position
func (d *Dungeon) AddEntity(entityType string, x, y int) {
	d.Entities[entityType] = append(d.Entities[entityType], EntityPosition{X: x, Y: y})
}

// GetSpawnPoint returns a spawn point in a room (center of first room)
func (d *Dungeon) GetSpawnPoint() (int, int) {
	if len(d.Rooms) == 0 {
		return 5, 5
	}
	return d.Rooms[0].CenterX, d.Rooms[0].CenterY
}

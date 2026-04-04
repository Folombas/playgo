package entity

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
)

// Enemy - враг
type Enemy struct {
	X, Y     float64
	Speed    float64
	Alive    bool
	
	// AI
	dirX, dirY       float64
	dirChangeTimer   float64
	moveTimer        float64

	// Анимация
	animFrame int
	animTimer float64
}

// NewEnemy создает нового врага
func NewEnemy(gridX, gridY int, speed float64) *Enemy {
	return &Enemy{
		X:              float64(gridX * config.TileSize),
		Y:              float64(gridY * config.TileSize),
		Speed:          speed,
		Alive:          true,
		dirX:           0,
		dirY:           0,
		dirChangeTimer: 0,
		moveTimer:      0,
		animFrame:      0,
		animTimer:      0,
	}
}

// Update обновляет врага
func (e *Enemy) Update(grid *Grid) {
	if !e.Alive {
		return
	}

	// Меняем направление периодически
	e.dirChangeTimer += 1.0 / config.TPS
	if e.dirChangeTimer >= 1.0 { // Каждую секунду
		e.dirChangeTimer = 0
		e.changeDirection(grid)
	}

	// Двигаемся
	dx := e.dirX * e.Speed / config.TPS
	dy := e.dirY * e.Speed / config.TPS

	newX := e.X + dx
	newY := e.Y + dy

	// Проверяем коллизии
	if e.canMoveTo(newX, newY, grid) {
		e.X = newX
		e.Y = newY
	} else {
		// Меняем направление при столкновении
		e.changeDirection(grid)
	}

	// Обновляем анимацию
	e.moveTimer += 1.0 / config.TPS
	if e.moveTimer >= 0.2 {
		e.moveTimer = 0
		e.animFrame = (e.animFrame + 1) % 2
	}
}

// changeDirection меняет направление движения
func (e *Enemy) changeDirection(grid *Grid) {
	// Случайное направление
	directions := [][2]float64{
		{0, -1}, // Up
		{0, 1},  // Down
		{-1, 0}, // Left
		{1, 0},  // Right
	}

	// Перемешиваем направления
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	// Выбираем первое валидное направление
	for _, dir := range directions {
		newX := e.X + dir[0]*float64(config.TileSize)
		newY := e.Y + dir[1]*float64(config.TileSize)
		
		if e.canMoveTo(newX, newY, grid) {
			e.dirX = dir[0]
			e.dirY = dir[1]
			return
		}
	}

	// Если нет валидного направления, остаемся на месте
	e.dirX = 0
	e.dirY = 0
}

// canMoveTo проверяет, может ли враг переместиться
func (e *Enemy) canMoveTo(x, y float64, grid *Grid) bool {
	halfSize := float64(config.TileSize) * 0.4

	corners := [][2]float64{
		{x - halfSize, y - halfSize},
		{x + halfSize, y - halfSize},
		{x - halfSize, y + halfSize},
		{x + halfSize, y + halfSize},
	}

	for _, corner := range corners {
		gridX := int(corner[0] / float64(config.TileSize))
		gridY := int(corner[1] / float64(config.TileSize))

		if gridX < 0 || gridX >= config.GridWidth || gridY < 0 || gridY >= config.GridHeight {
			return false
		}

		if !grid.IsWalkable(gridX, gridY) {
			return false
		}
	}

	return true
}

// Draw отрисовывает врага
func (e *Enemy) Draw(screen *ebiten.Image, sprite1, sprite2 *ebiten.Image) {
	if !e.Alive {
		return
	}

	var sprite *ebiten.Image
	if e.animFrame == 0 {
		sprite = sprite1
	} else {
		sprite = sprite2
	}

	if sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	scale := float64(config.TileSize) / float64(sprite.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(e.X, e.Y)

	screen.DrawImage(sprite, op)
}

// Kill убивает врага
func (e *Enemy) Kill() {
	e.Alive = false
}

// IsDead проверяет, мертв ли враг
func (e *Enemy) IsDead() bool {
	return !e.Alive
}

// GetPosition возвращает позицию
func (e *Enemy) GetPosition() (float64, float64) {
	return e.X, e.Y
}

// SpawnEnemies спавнит врагов на случайных позициях
func SpawnEnemies(count int, grid *Grid) []*Enemy {
	enemies := make([]*Enemy, 0, count)
	positions := make([][2]int, 0)

	// Находим все свободные позиции
	for y := 0; y < config.GridHeight; y++ {
		for x := 0; x < config.GridWidth; x++ {
			// Не спавним близко к игроку
			if x <= 3 && y <= 3 {
				continue
			}

			if grid.IsWalkable(x, y) {
				positions = append(positions, [2]int{x, y})
			}
		}
	}

	// Перемешиваем
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Создаем врагов
	for i := 0; i < count && i < len(positions); i++ {
		pos := positions[i]
		enemy := NewEnemy(pos[0], pos[1], config.EnemySpeed)
		enemies = append(enemies, enemy)
	}

	return enemies
}

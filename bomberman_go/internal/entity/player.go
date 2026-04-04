package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/playgo/bomberman_go/internal/config"
)

// Direction - направление движения
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// Player - игрок
type Player struct {
	X, Y             float64 // Позиция в пикселях
	Lives            int
	MaxBombs         int
	ActiveBombs      int
	ExplosionRadius  int
	Speed            float64

	// Анимация
	animFrame   int
	animTimer   float64
	direction   Direction
	isMoving    bool
}

// NewPlayer создает нового игрока
func NewPlayer(gridX, gridY int) *Player {
	return &Player{
		X:               float64(gridX * config.TileSize),
		Y:               float64(gridY * config.TileSize),
		Lives:           config.StartingLives,
		MaxBombs:        config.MaxBombs,
		ActiveBombs:     0,
		ExplosionRadius: config.BaseExplosionRadius,
		Speed:           config.PlayerSpeed,
		animFrame:       0,
		animTimer:       0,
		direction:       DirDown,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update(grid *Grid, bombs []*Bomb) {
	dx, dy := 0.0, 0.0
	p.isMoving = false

	// Обработка ввода
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		p.direction = DirUp
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		p.direction = DirDown
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		p.direction = DirLeft
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		p.direction = DirRight
	}

	// Движение
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp):
		dy = -p.Speed / config.TPS
		p.isMoving = true
		p.direction = DirUp
	case ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown):
		dy = p.Speed / config.TPS
		p.isMoving = true
		p.direction = DirDown
	case ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft):
		dx = -p.Speed / config.TPS
		p.isMoving = true
		p.direction = DirLeft
	case ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight):
		dx = p.Speed / config.TPS
		p.isMoving = true
		p.direction = DirRight
	}

	// Проверяем коллизии перед движением
	newX := p.X + dx
	newY := p.Y + dy

	if p.canMoveTo(newX, newY, grid, bombs) {
		p.X = newX
		p.Y = newY
	} else if p.canMoveTo(newX, p.Y, grid, bombs) {
		p.X = newX // Движение только по X
	} else if p.canMoveTo(p.X, newY, grid, bombs) {
		p.Y = newY // Движение только по Y
	}

	// Обновляем анимацию
	if p.isMoving {
		p.animTimer++
		if p.animTimer >= 8 { // Смена кадра каждые 8 тиков
			p.animTimer = 0
			p.animFrame = (p.animFrame + 1) % 2
		}
	} else {
		p.animFrame = 0
		p.animTimer = 0
	}
}

// canMoveTo проверяет, может ли игрок переместиться в новую позицию
func (p *Player) canMoveTo(x, y float64, grid *Grid, bombs []*Bomb) bool {
	// Размер хитбокса игрока (немного меньше тайла)
	halfSize := float64(config.TileSize) * 0.4

	// Четыре угла хитбокса
	corners := [][2]float64{
		{x - halfSize, y - halfSize},
		{x + halfSize, y - halfSize},
		{x - halfSize, y + halfSize},
		{x + halfSize, y + halfSize},
	}

	for _, corner := range corners {
		gridX := int(corner[0] / float64(config.TileSize))
		gridY := int(corner[1] / float64(config.TileSize))

		// Проверяем границы
		if gridX < 0 || gridX >= config.GridWidth || gridY < 0 || gridY >= config.GridHeight {
			return false
		}

		// Проверяем стены
		if !grid.IsWalkable(gridX, gridY) {
			return false
		}

		// Проверяем бомбы (нельзя проходить сквозь бомбы)
		for _, bomb := range bombs {
			bombGridX := int(bomb.X / float64(config.TileSize))
			bombGridY := int(bomb.Y / float64(config.TileSize))
			if gridX == bombGridX && gridY == bombGridY {
				// Можно выходить с бомбы, но нельзя входить на неё
				currentGridX := int(p.X / float64(config.TileSize))
				currentGridY := int(p.Y / float64(config.TileSize))
				if currentGridX != bombGridX || currentGridY != bombGridY {
					return false
				}
			}
		}
	}

	return true
}

// Draw отрисовывает игрока
func (p *Player) Draw(screen *ebiten.Image, standSprite, walk1, walk2 *ebiten.Image) {
	var sprite *ebiten.Image

	if p.isMoving {
		if p.animFrame == 0 {
			sprite = walk1
		} else {
			sprite = walk2
		}
	} else {
		sprite = standSprite
	}

	if sprite == nil {
		return
	}

	// Масштабируем спрайт до размера тайла
	op := &ebiten.DrawImageOptions{}
	scale := float64(config.TileSize) / float64(sprite.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(p.X, p.Y)

	// Отражаем спрайт в зависимости от направления
	if p.direction == DirLeft {
		op.GeoM.Translate(float64(sprite.Bounds().Dx())*scale, 0)
		op.GeoM.Scale(-1, 1)
	}

	screen.DrawImage(sprite, op)
}

// CollidesWith проверяет коллизию с другой сущностью
func (p *Player) CollidesWith(other interface{ GetPosition() (float64, float64) }) bool {
	ox, oy := other.GetPosition()
	dist := math.Hypot(p.X-ox, p.Y-oy)
	return dist < float64(config.TileSize)*0.7
}

// IsInExplosion проверяет, находится ли игрок в взрыве
func (p *Player) IsInExplosion(explos *Explosion) bool {
	playerGridX := int(p.X / float64(config.TileSize))
	playerGridY := int(p.Y / float64(config.TileSize))
	return explos.Contains(playerGridX, playerGridY)
}

// TakeDamage наносит урон игроку
func (p *Player) TakeDamage() {
	p.Lives--
	if p.Lives < 0 {
		p.Lives = 0
	}
	//TODO: Добавить неуязвимость на короткое время
}

// TryPlaceBomb пытается установить бомбу
func (p *Player) TryPlaceBomb(bombs *[]*Bomb) {
	if p.ActiveBombs >= p.MaxBombs {
		return
	}

	// Определяем позицию бомы (в центре ближайшего тайла)
	bombGridX := int((p.X + float64(config.TileSize)/2) / float64(config.TileSize))
	bombGridY := int((p.Y + float64(config.TileSize)/2) / float64(config.TileSize))

	bombX := float64(bombGridX * config.TileSize)
	bombY := float64(bombGridY * config.TileSize)

	// Проверяем, нет ли уже бомбы здесь
	for _, bomb := range *bombs {
		if int(bomb.X) == int(bombX) && int(bomb.Y) == int(bombY) {
			return
		}
	}

	bomb := NewBomb(bombX, bombY, p.ExplosionRadius)
	*bombs = append(*bombs, bomb)
	p.ActiveBombs++
	bomb.OnExplode = func() {
		p.ActiveBombs--
	}
}

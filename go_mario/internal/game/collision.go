package game

// Rect - прямоугольник для коллизий
type Rect struct {
	X, Y     float32
	Width    float32
	Height   float32
}

// NewRect создаёт новый прямоугольник
func NewRect(x, y, w, h float32) Rect {
	return Rect{X: x, Y: y, Width: w, Height: h}
}

// CheckRectCollision проверяет столкновение двух прямоугольников
func CheckRectCollision(r1, r2 Rect) bool {
	return r1.X < r2.X+r2.Width &&
		r1.X+r1.Width > r2.X &&
		r1.Y < r2.Y+r2.Height &&
		r1.Y+r1.Height > r2.Y
}

// CheckRectCollisionSimple упрощённая проверка коллизий
func CheckRectCollisionSimple(x1, y1, w1, h1, x2, y2, w2, h2 float32) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// CheckCircleCollision проверяет столкновение двух кругов
func CheckCircleCollision(x1, y1, r1, x2, y2, r2 float32) bool {
	dx := x2 - x1
	dy := y2 - y1
	distance := dx*dx + dy*dy
	radiusSum := r1 + r2
	return distance < radiusSum*radiusSum
}

// GetPlayerRect возвращает прямоугольник игрока
func GetPlayerRect(x, y float64, width, height float32) Rect {
	return Rect{
		X:      float32(x),
		Y:      float32(y),
		Width:  width,
		Height: height,
	}
}

// Contains проверяет, содержит ли прямоугольник точку
func (r Rect) Contains(x, y float32) bool {
	return x >= r.X && x <= r.X+r.Width &&
		y >= r.Y && y <= r.Y+r.Height
}

// Overlaps проверяет перекрытие с другим прямоугольником
func (r Rect) Overlaps(other Rect) bool {
	return CheckRectCollision(r, other)
}

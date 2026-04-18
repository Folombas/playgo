package input

type Position struct { X, Y int }
func ScreenPosToBoardPos(sx, sy float64) *Position {
    tx := int((sx - 100) / 64)
    ty := int((sy - 100) / 64)
    if tx >= 0 && tx < 8 && ty >= 0 && ty < 8 {
        return &Position{tx, ty}
    }
    return nil
}

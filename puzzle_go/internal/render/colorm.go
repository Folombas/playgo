package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ColorMHelper помощник для работы с цветовыми матрицами
type ColorMHelper struct{}

// ScaleAlpha добавляет прозрачность к ColorM
func ScaleAlpha(cm *ebiten.ColorM, alpha float64) {
	cm.Scale(1, 1, 1, alpha)
}

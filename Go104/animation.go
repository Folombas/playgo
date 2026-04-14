package main

import (
	"math"
	"time"
)

// AnimationType определяет тип анимации
type AnimationType int

const (
	AnimSwap AnimationType = iota
	AnimShake
	AnimRemove
	AnimDrop
	AnimHintPulse
)

// Animation представляет одну анимацию
type Animation struct {
	Type       AnimationType
	StartTime  time.Time
	Duration   time.Duration
	Positions  []*Position //Affected позиции
	Complete   bool
}

// AnimationManager управляет всеми анимациями
type AnimationManager struct {
	Animations []*Animation
}

// NewAnimationManager создаёт менеджер анимаций
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{}
}

// AddSwap добавляет анимацию обмена
func (am *AnimationManager) AddSwap(p1, p2 *Position) {
	am.Animations = append(am.Animations, &Animation{
		Type:      AnimSwap,
		StartTime: time.Now(),
		Duration:  200 * time.Millisecond,
		Positions: []*Position{p1, p2},
	})
}

// AddShake добавляет анимацию дрожания (ошибка)
func (am *AnimationManager) AddShake(p1, p2 *Position) {
	am.Animations = append(am.Animations, &Animation{
		Type:      AnimShake,
		StartTime: time.Now(),
		Duration:  150 * time.Millisecond, // 3 цикла по 50мс
		Positions: []*Position{p1, p2},
	})
}

// AddRemove добавляет анимацию удаления
func (am *AnimationManager) AddRemove(positions []*Position) {
	am.Animations = append(am.Animations, &Animation{
		Type:      AnimRemove,
		StartTime: time.Now(),
		Duration:  150 * time.Millisecond,
		Positions: positions,
	})
}

// AddDrop добавляет анимацию падения
func (am *AnimationManager) AddDrop(positions []*Position) {
	am.Animations = append(am.Animations, &Animation{
		Type:      AnimDrop,
		StartTime: time.Now(),
		Duration:  200 * time.Millisecond,
		Positions: positions,
	})
}

// AddHintPulse добавляет анимацию подсказки
func (am *AnimationManager) AddHintPulse(p1, p2 *Position) {
	am.Animations = append(am.Animations, &Animation{
		Type:      AnimHintPulse,
		StartTime: time.Now(),
		Duration:  2000 * time.Millisecond, // 2 секунды пульсации
		Positions: []*Position{p1, p2},
	})
}

// Update обновляет все анимации
func (am *AnimationManager) Update(board *Board) bool {
	hasActive := false
	now := time.Now()

	for _, anim := range am.Animations {
		if anim.Complete {
			continue
		}

		elapsed := now.Sub(anim.StartTime)
		progress := float64(elapsed) / float64(anim.Duration)

		if progress >= 1.0 {
			anim.Complete = true
			am.applyAnimation(anim, board, 1.0)
		} else {
			hasActive = true
			am.applyAnimation(anim, board, progress)
		}
	}

	// Удаляем завершённые анимации
	cleaned := make([]*Animation, 0, len(am.Animations))
	for _, anim := range am.Animations {
		if !anim.Complete {
			cleaned = append(cleaned, anim)
		}
	}
	am.Animations = cleaned

	return hasActive
}

// applyAnimation применяет анимацию к фишкам
func (am *AnimationManager) applyAnimation(anim *Animation, board *Board, progress float64) {
	switch anim.Type {
	case AnimSwap:
		am.applySwap(anim, progress)
	case AnimShake:
		am.applyShake(anim, board, progress)
	case AnimRemove:
		am.applyRemove(anim, board, progress)
	case AnimDrop:
		am.applyDrop(anim, board, progress)
	case AnimHintPulse:
		am.applyHintPulse(anim, progress)
	}
}

// applySwap анимирует обмен фишек
func (am *AnimationManager) applySwap(anim *Animation, board *Board, progress float64) {
	if len(anim.Positions) < 2 {
		return
	}

	p1 := anim.Positions[0]
	p2 := anim.Positions[1]

	t1 := board.Grid[p1.Row][p1.Col]
	t2 := board.Grid[p2.Row][p2.Col]

	if t1 == nil || t2 == nil {
		return
	}

	// Начальные позиции - фактические координаты фишек
	startX1 := t1.X
	startY1 := t1.Y
	startX2 := t2.X
	startY2 := t2.Y

	// Целевые позиции
	targetX1 := float64(p2.Col * CellSize)
	targetY1 := float64(p2.Row * CellSize)
	targetX2 := float64(p1.Col * CellSize)
	targetY2 := float64(p1.Row * CellSize)

	// Линейная интерполяция
	t1.X = startX1 + (targetX1-startX1)*progress
	t1.Y = startY1 + (targetY1-startY1)*progress
	t2.X = startX2 + (targetX2-startX2)*progress
	t2.Y = startY2 + (targetY2-startY2)*progress
}

// applyShake анимирует дрожание при ошибке
func (am *AnimationManager) applyShake(anim *Animation, board *Board, progress float64) {
	// 3 цикла дрожания
	cycle := progress * 3.0
	shakeAmount := 0.0

	if cycle < 1.0 {
		shakeAmount = 4.0 * sinFast(cycle*3.14159*2)
	} else if cycle < 2.0 {
		shakeAmount = 4.0 * sinFast((cycle-1.0)*3.14159*2) * 0.66
	} else if cycle < 3.0 {
		shakeAmount = 4.0 * sinFast((cycle-2.0)*3.14159*2) * 0.33
	}

	for _, pos := range anim.Positions {
		if t := board.Grid[pos.Row][pos.Col]; t != nil {
			t.X = float64(pos.Col*CellSize) + shakeAmount
		}
	}
}

// applyRemove анимирует удаление фишек
func (am *AnimationManager) applyRemove(anim *Animation, board *Board, progress float64) {
	for _, pos := range anim.Positions {
		if t := board.Grid[pos.Row][pos.Col]; t != nil {
			t.Scale = 1.0 - progress
			t.Alpha = 1.0 - progress
		}
	}
}

// applyDrop анимирует падение фишек
func (am *AnimationManager) applyDrop(anim *Animation, board *Board, progress float64) {
	for _, pos := range anim.Positions {
		if t := board.Grid[pos.Row][pos.Col]; t != nil {
			startY := t.Y
			targetY := float64(pos.Row * CellSize)

			// Если это новая фишка, начинаем сверху
			if t.Y < 0 {
				startY = t.Y
			} else {
				startY = t.TargetY - float64(CellSize*2)
			}

			t.Y = startY + (targetY-startY)*easeOutQuad(progress)
		}
	}
}

// applyHintPulse анимирует подсказку
func (am *AnimationManager) applyHintPulse(anim *Animation, board *Board, progress float64) {
	// Пульсация масштаба
	pulse := 1.0 + 0.1*math.Sin(progress*math.Pi*4)

	for _, pos := range anim.Positions {
		if t := board.Grid[pos.Row][pos.Col]; t != nil {
			t.Scale = pulse
		}
	}
}

// easOutQuad функция плавности для анимации падения
func easeOutQuad(t float64) float64 {
	return t * (2 - t)
}

// sinFast быстрое приближённое вычисление sin(x)
func sinFast(x float64) float64 {
	return math.Sin(x)
}

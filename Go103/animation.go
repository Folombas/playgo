package main

import (
	"math"
	"time"
)

// AnimationType определяет тип анимации
type AnimationType int

const (
	AnimSwap AnimationType = iota // Обмен фишек
	AnimShake                       // Дрожание (невалидный обмен)
	AnimMatch                       // Удаление комбинации
	AnimDrop                        // Падение фишек
	AnimHint                        // Подсказка (пульсация)
	AnimScore                       // Всплывающие очки
)

// Animation представляет одну анимацию
type Animation struct {
	Type       AnimationType
	StartTime  time.Time
	Duration   time.Duration
	Row1, Col1 int // Первая позиция
	Row2, Col2 int // Вторая позиция (для swap/shake)
	Positions  [][2]int // Позиции для match/drop
}

// AnimationSystem управляет всеми анимациями
type AnimationSystem struct {
	animations []Animation
}

// NewAnimationSystem создаёт новую систему анимаций
func NewAnimationSystem() *AnimationSystem {
	return &AnimationSystem{
		animations: make([]Animation, 0),
	}
}

// AddSwap добавляет анимацию обмена
func (as *AnimationSystem) AddSwap(r1, c1, r2, c2 int) {
	as.animations = append(as.animations, Animation{
		Type:      AnimSwap,
		StartTime: time.Now(),
		Duration:  150 * time.Millisecond,
		Row1:      r1, Col1: c1,
		Row2:      r2, Col2: c2,
	})
}

// AddShake добавляет анимацию дрожания
func (as *AnimationSystem) AddShake(r1, c1, r2, c2 int) {
	as.animations = append(as.animations, Animation{
		Type:      AnimShake,
		StartTime: time.Now(),
		Duration:  150 * time.Millisecond, // 3 цикла по 50мс
		Row1:      r1, Col1: c1,
		Row2:      r2, Col2: c2,
	})
}

// AddMatch добавляет анимацию удаления
func (as *AnimationSystem) AddMatch(positions [][2]int) {
	as.animations = append(as.animations, Animation{
		Type:      AnimMatch,
		StartTime: time.Now(),
		Duration:  200 * time.Millisecond,
		Positions: positions,
	})
}

// AddDrop добавляет анимацию падения
func (as *AnimationSystem) AddDrop(newTiles [][2]int) {
	as.animations = append(as.animations, Animation{
		Type:      AnimDrop,
		StartTime: time.Now(),
		Duration:  250 * time.Millisecond,
		Positions: newTiles,
	})
}

// AddHint добавляет анимацию подсказки
func (as *AnimationSystem) AddHint(r1, c1, r2, c2 int) {
	as.animations = append(as.animations, Animation{
		Type:      AnimHint,
		StartTime: time.Now(),
		Duration:  2000 * time.Millisecond, // 2 секунды пульсации
		Row1:      r1, Col1: c1,
		Row2:      r2, Col2: c2,
	})
}

// IsAnimating возвращает true если есть активные анимации
func (as *AnimationSystem) IsAnimating() bool {
	as.cleanupFinished()
	return len(as.animations) > 0
}

// GetSwapProgress возвращает прогресс анимации обмена (0.0 - 1.0)
func (as *AnimationSystem) GetSwapProgress() float64 {
	for _, anim := range as.animations {
		if anim.Type == AnimSwap {
			elapsed := time.Since(anim.StartTime)
			progress := float64(elapsed) / float64(anim.Duration)
			if progress > 1.0 {
				progress = 1.0
			}
			return progress
		}
	}
	return 0.0
}

// GetShakeOffset возвращает смещение дрожания для позиции
func (as *AnimationSystem) GetShakeOffset(row, col int) (float64, float64) {
	for _, anim := range as.animations {
		if anim.Type == AnimShake {
			if (row == anim.Row1 && col == anim.Col1) || (row == anim.Row2 && col == anim.Col2) {
				elapsed := time.Since(anim.StartTime)
				// 3 цикла по 50мс
				cycle := int(elapsed.Milliseconds()) / 50
				if cycle >= 6 { // 6 полуциклов (3 полных)
					return 0, 0
				}
				// Чередуем +4 и -4 пикселя
				if cycle%2 == 0 {
					return 4, 0
				}
				return -4, 0
			}
		}
	}
	return 0, 0
}

// GetMatchAlpha возвращает альфа-канал для анимации удаления
func (as *AnimationSystem) GetMatchAlpha(row, col int) float64 {
	for _, anim := range as.animations {
		if anim.Type == AnimMatch {
			for _, pos := range anim.Positions {
				if pos[0] == row && pos[1] == col {
					elapsed := time.Since(anim.StartTime)
					progress := float64(elapsed) / float64(anim.Duration)
					if progress > 1.0 {
						return 0
					}
					// Fade out: от 1.0 до 0.0
					return 1.0 - progress
				}
			}
		}
	}
	return 1.0
}

// GetMatchScale возвращает масштаб для анимации удаления
func (as *AnimationSystem) GetMatchScale(row, col int) float64 {
	for _, anim := range as.animations {
		if anim.Type == AnimMatch {
			for _, pos := range anim.Positions {
				if pos[0] == row && pos[1] == col {
					elapsed := time.Since(anim.StartTime)
					progress := float64(elapsed) / float64(anim.Duration)
					if progress > 1.0 {
						return 0
					}
					// Scale down: от 1.0 до 0.0
					return 1.0 - progress
				}
			}
		}
	}
	return 1.0
}

// GetDropOffset возвращает смещение падения для новой фишки
func (as *AnimationSystem) GetDropOffset(row, col int) float64 {
	for _, anim := range as.animations {
		if anim.Type == AnimDrop {
			for _, pos := range anim.Positions {
				if pos[0] == row && pos[1] == col {
					elapsed := time.Since(anim.StartTime)
					progress := float64(elapsed) / float64(anim.Duration)
					if progress > 1.0 {
						return 0
					}
					// Падает сверху: от -cellSize*(row+1) до 0
					// Easing: квадратичное замедление
					eased := progress * progress
					return float64(row+1) * 64 * (1 - eased)
				}
			}
		}
	}
	return 0
}

// GetHintPulse возвращает пульсацию для подсказки (0.0 - 1.0)
func (as *AnimationSystem) GetHintPulse() float64 {
	for _, anim := range as.animations {
		if anim.Type == AnimHint {
			elapsed := time.Since(anim.StartTime)
			progress := float64(elapsed) / float64(anim.Duration)
			if progress > 1.0 {
				return 0
			}
			// Пульсация: sin-подобная кривая
			return (math.Sin(progress*math.Pi*4) + 1) / 2
		}
	}
	return 0
}

// IsHintActive возвращает true если подсказка активна
func (as *AnimationSystem) IsHintActive() bool {
	for _, anim := range as.animations {
		if anim.Type == AnimHint {
			elapsed := time.Since(anim.StartTime)
			return elapsed < anim.Duration
		}
	}
	return false
}

// GetHintPositions возвращает позиции подсказки
func (as *AnimationSystem) GetHintPositions() ([2]int, [2]int, bool) {
	for _, anim := range as.animations {
		if anim.Type == AnimHint && as.IsHintActive() {
			return [2]int{anim.Row1, anim.Col1}, [2]int{anim.Row2, anim.Col2}, true
		}
	}
	return [2]int{}, [2]int{}, false
}

// cleanupFinished удаляет завершённые анимации
func (as *AnimationSystem) cleanupFinished() {
	now := time.Now()
	active := make([]Animation, 0, len(as.animations))
	for _, anim := range as.animations {
		if now.Sub(anim.StartTime) < anim.Duration {
			active = append(active, anim)
		}
	}
	as.animations = active
}

// Clear очищает все анимации
func (as *AnimationSystem) Clear() {
	as.animations = make([]Animation, 0)
}

package logic

// Level представляет уровень игры
type Level struct {
	Number         int
	TargetScore    int
	MaxMoves       int
	GemTypes       int // Количество типов камней (сложность)
	BoardRows      int
	BoardCols      int
	TimeLimit      int // В секундах (0 = без лимита)
}

// LevelManager управляет уровнями и прогрессией
type LevelManager struct {
	currentLevel int
	totalScore   int
	levels       []Level
}

// NewLevelManager создаёт менеджер уровней
func NewLevelManager() *LevelManager {
	lm := &LevelManager{
		currentLevel: 1,
		totalScore:   0,
		levels:       generateLevels(),
	}
	return lm
}

// generateLevels генерирует список уровней
func generateLevels() []Level {
	return []Level{
		// Уровень 1 - Обучающий
		{
			Number:      1,
			TargetScore: 500,
			MaxMoves:    30,
			GemTypes:    4, // Только 4 типа камней
			BoardRows:   6,
			BoardCols:   6,
			TimeLimit:   0,
		},
		// Уровень 2 - Легкий
		{
			Number:      2,
			TargetScore: 1000,
			MaxMoves:    25,
			GemTypes:    5,
			BoardRows:   7,
			BoardCols:   7,
			TimeLimit:   0,
		},
		// Уровень 3 - Стандартный
		{
			Number:      3,
			TargetScore: 2000,
			MaxMoves:    25,
			GemTypes:    6,
			BoardRows:   8,
			BoardCols:   8,
			TimeLimit:   0,
		},
		// Уровень 4 - Сложнее
		{
			Number:      4,
			TargetScore: 3000,
			MaxMoves:    20,
			GemTypes:    6,
			BoardRows:   8,
			BoardCols:   8,
			TimeLimit:   180, // 3 минуты
		},
		// Уровень 5 - Продвинутый
		{
			Number:      5,
			TargetScore: 5000,
			MaxMoves:    20,
			GemTypes:    6,
			BoardRows:   9,
			BoardCols:   9,
			TimeLimit:   150,
		},
		// Уровень 6 - Эксперт
		{
			Number:      6,
			TargetScore: 7500,
			MaxMoves:    18,
			GemTypes:    6,
			BoardRows:   10,
			BoardCols:   10,
			TimeLimit:   120,
		},
		// Уровень 7 - Мастер
		{
			Number:      7,
			TargetScore: 10000,
			MaxMoves:    15,
			GemTypes:    6,
			BoardRows:   10,
			BoardCols:   10,
			TimeLimit:   90,
		},
		// Уровень 8 - Грандмастер
		{
			Number:      8,
			TargetScore: 15000,
			MaxMoves:    15,
			GemTypes:    6,
			BoardRows:   10,
			BoardCols:   10,
			TimeLimit:   60,
		},
		// Уровень 9+ - Бесконечная сложность
		{
			Number:      9,
			TargetScore: 20000,
			MaxMoves:    12,
			GemTypes:    6,
			BoardRows:   10,
			BoardCols:   10,
			TimeLimit:   45,
		},
		{
			Number:      10,
			TargetScore: 30000,
			MaxMoves:    10,
			GemTypes:    6,
			BoardRows:   10,
			BoardCols:   10,
			TimeLimit:   30,
		},
	}
}

// GetCurrentLevel возвращает текущий уровень
func (lm *LevelManager) GetCurrentLevel() Level {
	if lm.currentLevel <= len(lm.levels) {
		return lm.levels[lm.currentLevel-1]
	}
	// Генерация бесконечных уровней после 10
	return generateInfiniteLevel(lm.currentLevel)
}

// NextLevel переходит к следующему уровню
func (lm *LevelManager) NextLevel() {
	lm.currentLevel++
}

// GetLevelScore возвращает очки текущего уровня
func (lm *LevelManager) GetLevelScore() int {
	return lm.totalScore
}

// AddScore добавляет очки
func (lm *LevelManager) AddScore(score int) {
	lm.totalScore += score
}

// IsLevelComplete проверяет, пройден ли уровень
func (lm *LevelManager) IsLevelComplete(score int) bool {
	level := lm.GetCurrentLevel()
	return score >= level.TargetScore
}

// GetProgress возвращает прогресс к цели (0.0 - 1.0)
func (lm *LevelManager) GetProgress(score int) float64 {
	level := lm.GetCurrentLevel()
	if level.TargetScore == 0 {
		return 0.0
	}
	progress := float64(score) / float64(level.TargetScore)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// GetCurrentLevelNumber возвращает номер текущего уровня
func (lm *LevelManager) GetCurrentLevelNumber() int {
	return lm.currentLevel
}

// Reset сбрасывает прогресс уровня
func (lm *LevelManager) Reset() {
	lm.totalScore = 0
}

// generateInfiniteLevel генерирует уровни после 10
func generateInfiniteLevel(levelNum int) Level {
	baseTarget := 30000
	extraTarget := (levelNum - 10) * 15000
	
	return Level{
		Number:      levelNum,
		TargetScore: baseTarget + extraTarget,
		MaxMoves:    10,
		GemTypes:    6,
		BoardRows:   10,
		BoardCols:   10,
		TimeLimit:   30,
	}
}

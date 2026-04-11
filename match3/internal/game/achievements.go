package game

import "fmt"

// AchievementID определяет уникальный идентификатор достижения
type AchievementID string

const (
	AchievementFirstMatch     AchievementID = "first_match"
	AchievementCombo5         AchievementID = "combo_5"
	AchievementCombo10        AchievementID = "combo_10"
	AchievementLevel10        AchievementID = "level_10"
	AchievementScore5000      AchievementID = "score_5000"
	AchievementBombs10        AchievementID = "bombs_10"
	AchievementGamesPlayed10  AchievementID = "games_10"
	AchievementMoves100       AchievementID = "moves_100"
)

// Achievement представляет достижение игрока
type Achievement struct {
	ID          AchievementID
	Name        string
	Description string
	Unlocked    bool
	UnlockDate  string // Date in format "2006-01-02"
}

// AchievementSystem управляет системой достижений
type AchievementSystem struct {
	achievements map[AchievementID]*Achievement
	onUnlock     func(achievement Achievement)
}

// NewAchievementSystem создаёт новую систему достижений
func NewAchievementSystem() *AchievementSystem {
	as := &AchievementSystem{
		achievements: make(map[AchievementID]*Achievement),
		onUnlock:     nil,
	}
	
	// Инициализация достижений
	as.initAchievements()
	
	return as
}

// initAchievements инициализирует все достижения
func (as *AchievementSystem) initAchievements() {
	as.achievements[AchievementFirstMatch] = &Achievement{
		ID:          AchievementFirstMatch,
		Name:        "Первый матч",
		Description: "Соберите первый матч 3 в ряд",
		Unlocked:    false,
	}
	
	as.achievements[AchievementCombo5] = &Achievement{
		ID:          AchievementCombo5,
		Name:        "Комбо мастер",
		Description: "Достигните комбо x5",
		Unlocked:    false,
	}
	
	as.achievements[AchievementCombo10] = &Achievement{
		ID:          AchievementCombo10,
		Name:        "Комбо легенда",
		Description: "Достигните комбо x10",
		Unlocked:    false,
	}
	
	as.achievements[AchievementLevel10] = &Achievement{
		ID:          AchievementLevel10,
		Name:        "Десятый уровень",
		Description: "Пройдите 10 уровней",
		Unlocked:    false,
	}
	
	as.achievements[AchievementScore5000] = &Achievement{
		ID:          AchievementScore5000,
		Name:        "5000 очков",
		Description: "Наберите 5000 очков за одну игру",
		Unlocked:    false,
	}
	
	as.achievements[AchievementBombs10] = &Achievement{
		ID:          AchievementBombs10,
		Name:        "Подрывник",
		Description: "Создайте 10 бомб",
		Unlocked:    false,
	}
	
	as.achievements[AchievementGamesPlayed10] = &Achievement{
		ID:          AchievementGamesPlayed10,
		Name:        "Опытный игрок",
		Description: "Сыграйте 10 игр",
		Unlocked:    false,
	}
	
	as.achievements[AchievementMoves100] = &Achievement{
		ID:          AchievementMoves100,
		Name:        "Стратег",
		Description: "Сделайте 100 ходов за одну игру",
		Unlocked:    false,
	}
}

// SetOnUnlockCallback устанавливает callback при разблокировке достижения
func (as *AchievementSystem) SetOnUnlockCallback(callback func(achievement Achievement)) {
	as.onUnlock = callback
}

// CheckAndUnlock проверяет условия и разблокирует достижения
func (as *AchievementSystem) CheckAndUnlock(score int, moves int, combo int, level int, bombsCreated int, gamesPlayed int) {
	as.checkFirstMatch(score)
	as.checkCombo(combo)
	as.checkLevel(level)
	as.checkScore(score)
	as.checkBombs(bombsCreated)
	as.checkGamesPlayed(gamesPlayed)
	as.checkMoves(moves)
}

// checkFirstMatch проверяет достижение "Первый матч"
func (as *AchievementSystem) checkFirstMatch(score int) {
	if score > 0 {
		as.unlockAchievement(AchievementFirstMatch)
	}
}

// checkCombo проверяет достижения комбо
func (as *AchievementSystem) checkCombo(combo int) {
	if combo >= 5 {
		as.unlockAchievement(AchievementCombo5)
	}
	if combo >= 10 {
		as.unlockAchievement(AchievementCombo10)
	}
}

// checkLevel проверяет достижение уровня
func (as *AchievementSystem) checkLevel(level int) {
	if level >= 10 {
		as.unlockAchievement(AchievementLevel10)
	}
}

// checkScore проверяет достижение очков
func (as *AchievementSystem) checkScore(score int) {
	if score >= 5000 {
		as.unlockAchievement(AchievementScore5000)
	}
}

// checkBombs проверяет достижение бомб
func (as *AchievementSystem) checkBombs(bombsCreated int) {
	if bombsCreated >= 10 {
		as.unlockAchievement(AchievementBombs10)
	}
}

// checkGamesPlayed проверяет достижение количества игр
func (as *AchievementSystem) checkGamesPlayed(gamesPlayed int) {
	if gamesPlayed >= 10 {
		as.unlockAchievement(AchievementGamesPlayed10)
	}
}

// checkMoves проверяет достижение ходов
func (as *AchievementSystem) checkMoves(moves int) {
	if moves >= 100 {
		as.unlockAchievement(AchievementMoves100)
	}
}

// unlockAchievement разблокирует достижение
func (as *AchievementSystem) unlockAchievement(id AchievementID) {
	achievement, exists := as.achievements[id]
	if !exists || achievement.Unlocked {
		return
	}
	
	achievement.Unlocked = true
	achievement.UnlockDate = "2026-04-11" // Today's date
	
	fmt.Printf("🏆 Достижение разблокировано: %s - %s\n", achievement.Name, achievement.Description)
	
	// Вызвать callback
	if as.onUnlock != nil {
		as.onUnlock(*achievement)
	}
}

// GetAchievement возвращает достижение по ID
func (as *AchievementSystem) GetAchievement(id AchievementID) (Achievement, bool) {
	achievement, exists := as.achievements[id]
	if !exists {
		return Achievement{}, false
	}
	return *achievement, true
}

// GetUnlockedCount возвращает количество разблокированных достижений
func (as *AchievementSystem) GetUnlockedCount() int {
	count := 0
	for _, achievement := range as.achievements {
		if achievement.Unlocked {
			count++
		}
	}
	return count
}

// GetTotalCount возвращает общее количество достижений
func (as *AchievementSystem) GetTotalCount() int {
	return len(as.achievements)
}

// GetProgressString возвращает строку прогресса достижений
func (as *AchievementSystem) GetProgressString() string {
	return fmt.Sprintf("Достижения: %d/%d", as.GetUnlockedCount(), as.GetTotalCount())
}

package game

import (
	"fmt"
	"math/rand"
	"time"
)

// DailyChallenge представляет ежедневное задание
type DailyChallenge struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Target      int       `json:"target"`
	Reward      int       `json:"reward"` // Бонусные очки
	Completed   bool      `json:"completed"`
	Date        time.Time `json:"date"`
}

// DailyChallengeSystem управляет ежедневными заданиями
type DailyChallengeSystem struct {
	challenges   []DailyChallenge
	currentDate  time.Time
	onComplete   func(challenge DailyChallenge)
}

// NewDailyChallengeSystem создаёт новую систему ежедневных заданий
func NewDailyChallengeSystem() *DailyChallengeSystem {
	dcs := &DailyChallengeSystem{
		challenges:  make([]DailyChallenge, 0),
		currentDate: time.Now(),
	}
	
	dcs.generateDailyChallenges()
	
	return dcs
}

// generateDailyChallenges генерирует задания на сегодня
func (dcs *DailyChallengeSystem) generateDailyChallenges() {
	// Проверяем, нужно ли генерировать новые задания
	if len(dcs.challenges) == 0 || dcs.currentDate.Day() != time.Now().Day() {
		dcs.challenges = make([]DailyChallenge, 0)
		dcs.currentDate = time.Now()
		
		// Генерируем 3 случайных задания
		templates := getChallengeTemplates()
		
		// Выбираем 3 случайных
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 3 && i < len(templates); i++ {
			idx := rng.Intn(len(templates))
			challenge := templates[idx]
			challenge.Date = time.Now()
			challenge.Completed = false
			dcs.challenges = append(dcs.challenges, challenge)
		}
		
		fmt.Printf("📋 Сгенерировано %d ежедневных заданий\n", len(dcs.challenges))
	}
}

// GetChallenges возвращает текущие задания
func (dcs *DailyChallengeSystem) GetChallenges() []DailyChallenge {
	return dcs.challenges
}

// CheckProgress проверяет прогресс заданий
func (dcs *DailyChallengeSystem) CheckProgress(matches int, combo int, score int, bombs int) {
	for i := range dcs.challenges {
		if dcs.challenges[i].Completed {
			continue
		}
		
		challenge := &dcs.challenges[i]
		completed := false
		
		switch challenge.ID {
		case "match_10":
			completed = matches >= challenge.Target
		case "combo_5":
			completed = combo >= challenge.Target
		case "score_1000":
			completed = score >= challenge.Target
		case "bomb_3":
			completed = bombs >= challenge.Target
		}
		
		if completed && !challenge.Completed {
			challenge.Completed = true
			fmt.Printf("✅ Задание выполнено: %s (+%d очков)\n", challenge.Description, challenge.Reward)
			
			if dcs.onComplete != nil {
				dcs.onComplete(*challenge)
			}
		}
	}
}

// GetTotalReward возвращает общий бонус за выполненные задания
func (dcs *DailyChallengeSystem) GetTotalReward() int {
	total := 0
	for _, challenge := range dcs.challenges {
		if challenge.Completed {
			total += challenge.Reward
		}
	}
	return total
}

// SetOnCompleteCallback устанавливает callback при выполнении задания
func (dcs *DailyChallengeSystem) SetOnCompleteCallback(callback func(challenge DailyChallenge)) {
	dcs.onComplete = callback
}

// ChallengeTemplate шаблон задания
type ChallengeTemplate struct {
	ID          string
	Description string
	Target      int
	Reward      int
}

// getChallengeTemplates возвращает все шаблоны заданий
func getChallengeTemplates() []DailyChallenge {
	return []DailyChallenge{
		{
			ID:          "match_10",
			Description: "Соберите 10 матчей",
			Target:      10,
			Reward:      200,
		},
		{
			ID:          "match_25",
			Description: "Соберите 25 матчей",
			Target:      25,
			Reward:      500,
		},
		{
			ID:          "combo_5",
			Description: "Достигните комбо x5",
			Target:      5,
			Reward:      300,
		},
		{
			ID:          "score_1000",
			Description: "Наберите 1000 очков за игру",
			Target:      1000,
			Reward:      400,
		},
		{
			ID:          "bomb_3",
			Description: "Создайте 3 бомбы",
			Target:      3,
			Reward:      250,
		},
		{
			ID:          "score_3000",
			Description: "Наберите 3000 очков за игру",
			Target:      3000,
			Reward:      600,
		},
	}
}

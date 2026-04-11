package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveData представляет данные для сохранения
type SaveData struct {
	// Профиль игрока
	PlayerName    string    `json:"player_name"`
	CurrentLevel  int       `json:"current_level"`
	TotalScore    int       `json:"total_score"`
	CreatedAt     time.Time `json:"created_at"`
	LastPlayedAt  time.Time `json:"last_played_at"`
	
	// Рекорды уровней
	LevelRecords map[int]LevelRecord `json:"level_records"`
	
	// Статистика
	GamesPlayed    int `json:"games_played"`
	GamesWon       int `json:"games_won"`
	TotalMoves     int `json:"total_moves"`
	BestCombo      int `json:"best_combo"`
	TotalMatches   int `json:"total_matches"`
	
	// Настройки
	SoundVolume float64 `json:"sound_volume"`
	Muted       bool    `json:"muted"`
}

// LevelRecord представляет рекорд для одного уровня
type LevelRecord struct {
	LevelNumber  int       `json:"level_number"`
	BestScore    int       `json:"best_score"`
	BestMoves    int       `json:"best_moves"`
	FastestTime  int       `json:"fastest_time"` // в секундах
	Stars        int       `json:"stars"`        // 0-3 звезды
	CompletedAt  time.Time `json:"completed_at"`
	Attempts     int       `json:"attempts"`
}

// SaveManager управляет сохранением и загрузкой данных
type SaveManager struct {
	savePath string
	data     SaveData
}

// NewSaveManager создаёт менеджер сохранений
func NewSaveManager() *SaveManager {
	sm := &SaveManager{
		savePath: getSaveFilePath(),
		data:     createDefaultSave(),
	}
	
	// Загрузка существующего сохранения
	sm.Load()
	
	return sm
}

// getSaveFilePath возвращает путь к файлу сохранения
func getSaveFilePath() string {
	// Получаем директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback на текущую директорию
		return "save.json"
	}
	
	// Создаём скрытую папку для сохранений
	saveDir := filepath.Join(homeDir, ".match3_game")
	os.MkdirAll(saveDir, 0755)
	
	return filepath.Join(saveDir, "save.json")
}

// createDefaultSave создаёт стандартное сохранение
func createDefaultSave() SaveData {
	return SaveData{
		PlayerName:   "Player",
		CurrentLevel: 1,
		TotalScore:   0,
		CreatedAt:    time.Now(),
		LastPlayedAt: time.Now(),
		
		LevelRecords: make(map[int]LevelRecord),
		
		GamesPlayed:  0,
		GamesWon:     0,
		TotalMoves:   0,
		BestCombo:    0,
		TotalMatches: 0,
		
		SoundVolume: 0.5,
		Muted:       false,
	}
}

// Load загружает данные из файла
func (sm *SaveManager) Load() error {
	data, err := os.ReadFile(sm.savePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Файл не существует - это нормально для первого запуска
			fmt.Println("No save file found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read save file: %w", err)
	}

	err = json.Unmarshal(data, &sm.data)
	if err != nil {
		// Поврежденный файл сохранения - создать резервную копию
		fmt.Printf("Corrupted save file, creating backup: %v\n", err)
		sm.createBackupAndReset()
		return nil
	}

	fmt.Printf("Save loaded from %s\n", sm.savePath)
	return nil
}

// Save сохраняет данные в файл с обработкой ошибок
func (sm *SaveManager) Save() error {
	sm.data.LastPlayedAt = time.Now()

	data, err := json.MarshalIndent(sm.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Создать директорию если не существует
	dir := filepath.Dir(sm.savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Создать резервную копию перед записью
	sm.createBackup()

	err = os.WriteFile(sm.savePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	fmt.Printf("Save written to %s\n", sm.savePath)
	return nil
}

// createBackup создаёт резервную копию файла сохранения
func (sm *SaveManager) createBackup() {
	if _, err := os.Stat(sm.savePath); err == nil {
		backupPath := sm.savePath + ".bak"
		data, err := os.ReadFile(sm.savePath)
		if err != nil {
			fmt.Printf("Warning: Could not read save file for backup: %v\n", err)
			return
		}
		
		err = os.WriteFile(backupPath, data, 0644)
		if err != nil {
			fmt.Printf("Warning: Could not create backup: %v\n", err)
			return
		}
	}
}

// createBackupAndReset создаёт резервную копию и сбрасывает данные
func (sm *SaveManager) createBackupAndReset() {
	// Попытаться создать резервную копию поврежденного файла
	if _, err := os.Stat(sm.savePath); err == nil {
		backupPath := sm.savePath + ".corrupted"
		err := os.Rename(sm.savePath, backupPath)
		if err != nil {
			fmt.Printf("Warning: Could not backup corrupted save: %v\n", err)
		} else {
			fmt.Printf("Corrupted save backed up to %s\n", backupPath)
		}
	}
	
	// Сбросить данные
	sm.data = createDefaultSave()
}

// GetSaveData возвращает копию данных сохранения
func (sm *SaveManager) GetSaveData() SaveData {
	return sm.data
}

// UpdateLevelRecord обновляет рекорд уровня
func (sm *SaveManager) UpdateLevelRecord(level int, score int, moves int, timeSpent int) {
	record, exists := sm.data.LevelRecords[level]
	
	if !exists {
		// Новый рекорд
		record = LevelRecord{
			LevelNumber: level,
			BestScore:   score,
			BestMoves:   moves,
			FastestTime: timeSpent,
			CompletedAt: time.Now(),
			Attempts:    1,
		}
	} else {
		// Обновление существующего
		record.Attempts++
		
		if score > record.BestScore {
			record.BestScore = score
		}
		
		if moves < record.BestMoves || record.BestMoves == 0 {
			record.BestMoves = moves
		}
		
		if timeSpent > 0 && (timeSpent < record.FastestTime || record.FastestTime == 0) {
			record.FastestTime = timeSpent
		}
		
		record.CompletedAt = time.Now()
	}
	
	// Расчёт звёзд
	record.Stars = sm.calculateStars(level, score, moves)
	
	sm.data.LevelRecords[level] = record
}

// calculateStars рассчитывает количество звёзд (0-3)
func (sm *SaveManager) calculateStars(level int, score int, moves int) int {
	stars := 0
	
	// Звезда за прохождение
	if score > 0 {
		stars++
	}
	
	// Звезда за эффективность (меньше ходов = лучше)
	if moves > 0 && moves < 20 {
		stars++
	}
	
	// Звезда за высокий счёт
	if score > 1000 {
		stars++
	}
	
	return stars
}

// GetLevelRecord возвращает рекорд уровня
func (sm *SaveManager) GetLevelRecord(level int) (LevelRecord, bool) {
	record, exists := sm.data.LevelRecords[level]
	return record, exists
}

// GetTotalStars возвращает общее количество звёзд
func (sm *SaveManager) GetTotalStars() int {
	total := 0
	for _, record := range sm.data.LevelRecords {
		total += record.Stars
	}
	return total
}

// IncrementGamesPlayed увеличивает счётчик игр
func (sm *SaveManager) IncrementGamesPlayed() {
	sm.data.GamesPlayed++
}

// IncrementGamesWon увеличивает счётчик побед
func (sm *SaveManager) IncrementGamesWon() {
	sm.data.GamesWon++
}

// AddMoves добавляет ходы к статистике
func (sm *SaveManager) AddMoves(moves int) {
	sm.data.TotalMoves += moves
}

// IncrementMatches увеличивает счётчик матчей
func (sm *SaveManager) IncrementMatches() {
	sm.data.TotalMatches++
}

// UpdateBestCombo обновляет лучший комбо
func (sm *SaveManager) UpdateBestCombo(combo int) {
	if combo > sm.data.BestCombo {
		sm.data.BestCombo = combo
	}
}

// SetPlayerName устанавливает имя игрока
func (sm *SaveManager) SetPlayerName(name string) {
	sm.data.PlayerName = name
}

// SetCurrentLevel устанавливает текущий уровень
func (sm *SaveManager) SetCurrentLevel(level int) {
	sm.data.CurrentLevel = level
}

// SetSoundSettings устанавливает настройки звука
func (sm *SaveManager) SetSoundSettings(volume float64, muted bool) {
	sm.data.SoundVolume = volume
	sm.data.Muted = muted
}

// GetSaveFilePath returns save file path for external access
func (sm *SaveManager) SaveFilePath() string {
	return sm.savePath
}

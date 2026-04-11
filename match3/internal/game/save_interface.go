package game

// Compile-time проверка, что SaveManager реализует SaveStorage
var _ SaveStorage = (*SaveManager)(nil)

// SaveStorage определяет интерфейс для управления сохранениями
// Этот интерфейс позволяет легко заменять хранилище (файл, БД, облако)
// и упрощает тестирование через моки
type SaveStorage interface {
	// Load загружает данные из хранилища
	Load() error
	
	// Save сохраняет данные в хранилище
	Save() error
	
	// GetSaveData возвращает копию данных сохранения
	GetSaveData() SaveData
	
	// UpdateLevelRecord обновляет рекорд уровня
	UpdateLevelRecord(level int, score int, moves int, timeSpent int)
	
	// GetLevelRecord возвращает рекорд уровня
	GetLevelRecord(level int) (LevelRecord, bool)
	
	// GetTotalStars возвращает общее количество звёзд
	GetTotalStars() int
	
	// IncrementGamesPlayed увеличивает счётчик игр
	IncrementGamesPlayed()
	
	// IncrementGamesWon увеличивает счётчик побед
	IncrementGamesWon()
	
	// AddMoves добавляет ходы к статистике
	AddMoves(moves int)
	
	// IncrementMatches увеличивает счётчик матчей
	IncrementMatches()
	
	// UpdateBestCombo обновляет лучший комбо
	UpdateBestCombo(combo int)
	
	// SetPlayerName устанавливает имя игрока
	SetPlayerName(name string)
	
	// SetCurrentLevel устанавливает текущий уровень
	SetCurrentLevel(level int)
	
	// SetSoundSettings устанавливает настройки звука
	SetSoundSettings(volume float64, muted bool)
	
	// SaveFilePath возвращает путь к файлу сохранения (для отладки)
	SaveFilePath() string
}

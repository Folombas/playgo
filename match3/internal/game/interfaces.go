package game

// Compile-time проверка, что SoundManager реализует SoundPlayer
var _ SoundPlayer = (*SoundManager)(nil)

// SoundPlayer определяет интерфейс для воспроизведения звуков
// Этот интерфейс позволяет легко заменять реализации звуков
// и упрощает тестирование через моки
type SoundPlayer interface {
	// Play воспроизводит звук указанного типа
	Play(sound SoundType)
	
	// SetVolume устанавливает громкость (0.0 - 1.0)
	SetVolume(vol float64)
	
	// ToggleMute переключает режим отключения звука
	// Возвращает новое состояние
	ToggleMute() bool
	
	// IsMuted проверяет, отключен ли звук
	IsMuted() bool
	
	// GetVolume возвращает текущую громкость
	GetVolume() float64
}

// SoundEffect определяет интерфейс для генерации звуковых эффектов
// Может быть реализован через WAV, OGG, или процедурную генерацию
type SoundEffect interface {
	// Generate создаёт звуковой эффект и возвращает WAV данные
	Generate() ([]byte, error)
	
	// Duration возвращает длительность звука в секундах
	Duration() float64
}

// SoundRegistry определяет интерфейс для реестра звуков
type SoundRegistry interface {
	// Register регистрирует звук по типу
	Register(soundType SoundType, data []byte)
	
	// Get возвращает WAV данные для типа звука
	Get(soundType SoundType) ([]byte, bool)
	
	// Contains проверяет, зарегистрирован ли звук
	Contains(soundType SoundType) bool
}

package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundManager управляет звуками в игре
type SoundManager struct {
	audioContext *audio.Context
	swapSound    *audio.Player
	matchSound   *audio.Player
	errorSound   *audio.Player
	gameoverSound *audio.Player
}

// NewSoundManager создаёт новый менеджер звуков
func NewSoundManager() *SoundManager {
	sm := &SoundManager{}

	// Создаём аудио контекст (44100 Hz)
	sm.audioContext = audio.NewContext(44100)

	// Генерируем простые звуки программно
	sm.swapSound = sm.generateTone(440, 0.1)   // Ля первой октавы (короткий)
	sm.matchSound = sm.generateTone(880, 0.2)  // Ля второй октавы (повыше)
	sm.errorSound = sm.generateTone(220, 0.15) // Ля малой октавы (пониже)
	sm.gameoverSound = sm.generateTone(330, 0.5) // Ми первой октавы (длинный)

	return sm
}

// generateTone генерирует простой синусоидальный звук
// frequency - частота в Гц, duration - длительность в секундах
func (sm *SoundManager) generateTone(frequency float64, duration float64) *audio.Player {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	buf := make([]byte, numSamples*2) // 16-bit mono

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		// Синусоидальная волна с затуханием
		decay := 1.0 - (float64(i) / float64(numSamples))
		sample := math.Sin(2*math.Pi*frequency*t) * decay * 0.5

		// Преобразуем в 16-bit
		val := int16(sample * 32767)
		buf[i*2] = byte(val)
		buf[i*2+1] = byte(val >> 8)
	}

	player := sm.audioContext.NewPlayerFromBytes(buf)

	return player
}

// PlaySwap воспроизводит звук обмена
func (sm *SoundManager) PlaySwap() {
	if sm.swapSound != nil {
		sm.swapSound.Rewind()
		sm.swapSound.Play()
	}
}

// PlayMatch воспроизводит звук удаления комбинации
func (sm *SoundManager) PlayMatch() {
	if sm.matchSound != nil {
		sm.matchSound.Rewind()
		sm.matchSound.Play()
	}
}

// PlayError воспроизводит звук ошибки
func (sm *SoundManager) PlayError() {
	if sm.errorSound != nil {
		sm.errorSound.Rewind()
		sm.errorSound.Play()
	}
}

// PlayGameOver воспроизводит звук окончания игры
func (sm *SoundManager) PlayGameOver() {
	if sm.gameoverSound != nil {
		sm.gameoverSound.Rewind()
		sm.gameoverSound.Play()
	}
}

// Глобальный менеджер звуков
var soundManager *SoundManager

// initSounds инициализирует глобальный менеджер звуков
func initSounds() {
	soundManager = NewSoundManager()
	log.Println("Звуки инициализированы")
}

// PlaySwapWrapper обёртка для глобального доступа
func PlaySwap() {
	if soundManager != nil {
		soundManager.PlaySwap()
	}
}

// PlayMatchWrapper обёртка для глобального доступа
func PlayMatch() {
	if soundManager != nil {
		soundManager.PlayMatch()
	}
}

// PlayErrorWrapper обёртка для глобального доступа
func PlayError() {
	if soundManager != nil {
		soundManager.PlayError()
	}
}

// PlayGameOverWrapper обёртка для глобального доступа
func PlayGameOver() {
	if soundManager != nil {
		soundManager.PlayGameOver()
	}
}

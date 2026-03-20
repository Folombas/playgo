// Package audio содержит систему звуковых эффектов
package audio

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundType представляет тип звука
type SoundType int

const (
	SoundEatFood SoundType = iota
	SoundCollectCoin
	SoundCollectKey
	SoundOpenChest
	SoundShoot
	SoundExplosion
	SoundPowerUp
	SoundEnemyKill
	SoundGameOver
)

// AudioSystem управляет звуковыми эффектами
type AudioSystem struct {
	context    *audio.Context
	players    map[SoundType][]byte // Храним данные, создаём плееры при воспроизведении
	volume     float64
	enabled    bool
}

// NewAudioSystem создаёт новую аудиосистему
func NewAudioSystem() *AudioSystem {
	as := &AudioSystem{
		context: audio.NewContext(44100),
		players: make(map[SoundType][]byte),
		volume:  0.3,
		enabled: true,
	}
	
	// Предзагрузка звуков
	as.players[SoundEatFood] = as.generateEatFoodSound()
	as.players[SoundCollectCoin] = as.generateCoinSound()
	as.players[SoundCollectKey] = as.generateKeySound()
	as.players[SoundOpenChest] = as.generateChestSound()
	as.players[SoundShoot] = as.generateShootSound()
	as.players[SoundExplosion] = as.generateExplosionSound()
	as.players[SoundPowerUp] = as.generatePowerUpSound()
	as.players[SoundEnemyKill] = as.generateEnemyKillSound()
	as.players[SoundGameOver] = as.generateGameOverSound()
	
	return as
}

// Play воспроизводит звук
func (as *AudioSystem) Play(soundType SoundType) {
	if !as.enabled {
		return
	}

	data, ok := as.players[soundType]
	if !ok {
		return
	}

	player := as.context.NewPlayerFromBytes(data)
	player.SetVolume(as.volume)
	player.Rewind()
	player.Play()
}

// SetVolume устанавливает громкость
func (as *AudioSystem) SetVolume(volume float64) {
	as.volume = volume
}

// Enable включает/выключает звук
func (as *AudioSystem) Enable(enabled bool) {
	as.enabled = enabled
}

// generateEatFoodSound генерирует звук поедания еды
func (as *AudioSystem) generateEatFoodSound() []byte {
	// Короткий "чпок" (синусоида с затуханием)
	return as.generateSound(0.1, func(t float64) float64 {
		freq := 800.0 - 400.0*t // Частота падает с 800 до 400 Гц
		amp := 1.0 - t        // Затухание
		return amp * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateCoinSound генерирует звук сбора монеты
func (as *AudioSystem) generateCoinSound() []byte {
	// Высокий "динь" (синусоида 880 Гц с вибрато)
	return as.generateSound(0.2, func(t float64) float64 {
		freq := 880.0 + 50.0*math.Sin(2.0*math.Pi*10.0*t) // Вибрато 10 Гц
		amp := 1.0 - t
		return amp * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateKeySound генерирует звук сбора ключа
func (as *AudioSystem) generateKeySound() []byte {
	// Звонкий звук (две гармоники)
	return as.generateSound(0.15, func(t float64) float64 {
		freq1 := 1200.0
		freq2 := 1800.0
		amp := 1.0 - t
		return amp * (math.Sin(2.0*math.Pi*freq1*t) + 0.5*math.Sin(2.0*math.Pi*freq2*t))
	})
}

// generateChestSound генерирует звук открытия сундука
func (as *AudioSystem) generateChestSound() []byte {
	// Низкий "бум" с последующим звоном
	return as.generateSound(0.3, func(t float64) float64 {
		if t < 0.1 {
			// Удар
			return (1.0 - t*10.0) * math.Sin(2.0*math.Pi*200.0*t)
		}
		// Звон монет
		freq := 2000.0 + 500.0*math.Sin(2.0*math.Pi*20.0*t)
		return (1.0 - t) * 0.5 * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateShootSound генерирует звук выстрела
func (as *AudioSystem) generateShootSound() []byte {
	// Быстрый свист (высокая частота с быстрым падением)
	return as.generateSound(0.15, func(t float64) float64 {
		freq := 2000 - 1500*t
		amp := 1.0 - t
		return amp * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateExplosionSound генерирует звук взрыва
func (as *AudioSystem) generateExplosionSound() []byte {
	// Низкий грохот с шумом
	return as.generateSound(0.4, func(t float64) float64 {
		// Основной тон
		freq := 80 - 40*t
		// Добавляем "шум" через быструю модуляцию
		noise := 0.3 * math.Sin(2.0*math.Pi*10.00*t) * math.Sin(2.0*math.Pi*500.0*t)
		amp := 1.0 - t
		return amp * (math.Sin(2.0*math.Pi*freq*t) + noise)
	})
}

// generatePowerUpSound генерирует звук бонуса
func (as *AudioSystem) generatePowerUpSound() []byte {
	// Восходящий арпеджио
	return as.generateSound(0.3, func(t float64) float64 {
		freq := 400 + 800*t // От 400 до 1200 Гц
		amp := 1.0 - t
		return amp * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateEnemyKillSound генерирует звук убийства врага
func (as *AudioSystem) generateEnemyKillSound() []byte {
	// Неприятный скрежет
	return as.generateSound(0.2, func(t float64) float64 {
		freq1 := 300 + 200*math.Sin(2.0*math.Pi*30.0*t)
		freq2 := 450 + 300*math.Sin(2.0*math.Pi*25.0*t)
		amp := 1.0 - t
		return amp * (math.Sin(2.0*math.Pi*freq1*t) + 0.7*math.Sin(2.0*math.Pi*freq2*t))
	})
}

// generateGameOverSound генерирует звук проигрыша
func (as *AudioSystem) generateGameOverSound() []byte {
	// Нисходящий грустный звук
	return as.generateSound(0.5, func(t float64) float64 {
		freq := 600 - 400*t
		amp := 1.0 - t
		return amp * math.Sin(2.0*math.Pi*freq*t)
	})
}

// generateSound генерирует звук заданной длительности
func (as *AudioSystem) generateSound(duration float64, generator func(float64) float64) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	data := make([]byte, numSamples*4) // 16-bit stereo
	
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := generator(t)
		
		// Ограничиваем амплитуду
		if sample > 1 {
			sample = 1
		} else if sample < -1 {
			sample = -1
		}
		
		// Конвертируем в 16-bit
		sampleInt := int16(sample * 32767)
		
		// Stereo: левый и правый каналы
		data[i*4] = byte(sampleInt)
		data[i*4+1] = byte(sampleInt >> 8)
		data[i*4+2] = byte(sampleInt)
		data[i*4+3] = byte(sampleInt >> 8)
	}
	
	return data
}

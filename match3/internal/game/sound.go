package game

import (
	"bytes"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// SoundType определяет тип звука
type SoundType int

const (
	SoundMatch SoundType = iota
	SoundSwap
	SoundInvalid
	SoundCombo
	SoundGameOver
	SoundBomb      // Звук взрыва бомбы
	SoundFire      // Звук огненного камня
	SoundIce       // Звук разбивания льда
	SoundCount
)

// SoundManager управляет воспроизведением звуков
type SoundManager struct {
	audioContext *audio.Context
	players      map[SoundType][]*wav.Stream
	volume       float64
	muted        bool
}

// NewSoundManager создаёт новый менеджер звуков
func NewSoundManager() *SoundManager {
	sm := &SoundManager{
		audioContext: audio.NewContext(44100),
		players:      make(map[SoundType][]*wav.Stream),
		volume:       0.5,
		muted:        false,
	}
	
	// Генерируем простые звуки программно
	sm.generateSounds()
	
	return sm
}

// generateSounds создаёт программно простые WAV звуки
func (sm *SoundManager) generateSounds() {
	// Match sound - приятный восходящий звук
	sm.players[SoundMatch] = sm.generateTone(440, 0.15, "match")

	// Swap sound - короткий щелчок
	sm.players[SoundSwap] = sm.generateTone(330, 0.08, "swap")

	// Invalid sound - низкий тон
	sm.players[SoundInvalid] = sm.generateTone(220, 0.1, "invalid")

	// Combo sound - радостная мелодия
	sm.players[SoundCombo] = sm.generateTone(550, 0.2, "combo")

	// Game Over sound - грустный нисходящий
	sm.players[SoundGameOver] = sm.generateTone(330, 0.3, "gameover")
	
	// Bomb sound - низкий взрыв
	sm.players[SoundBomb] = sm.generateTone(150, 0.25, "bomb")
	
	// Fire sound - шипящий звук
	sm.players[SoundFire] = sm.generateTone(600, 0.2, "fire")
	
	// Ice sound - звонкий звук
	sm.players[SoundIce] = sm.generateTone(800, 0.15, "ice")
}

// generateTone генерирует простой тон заданной частоты и длительности
func (sm *SoundManager) generateTone(baseFreq float64, duration float64, soundName string) []*wav.Stream {
	// Создаём несколько вариантов звука
	count := 3
	streams := make([]*wav.Stream, 0, count)
	
	for i := 0; i < count; i++ {
		// Небольшая вариация частоты
		freq := baseFreq * (1 + rand.Float64()*0.1 - 0.05)
		
		// Генерация WAV данных
		data := sm.generateWavData(freq, duration, soundName)
		
		// Декодирование WAV
		stream, err := wav.Decode(sm.audioContext, bytes.NewReader(data))
		if err != nil {
			continue
		}
		
		streams = append(streams, stream)
	}
	
	return streams
}

// generateWavData создаёт простые WAV данные
func (sm *SoundManager) generateWavData(freq float64, duration float64, soundName string) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	
	// WAV header (44 bytes) + data
	headerSize := 44
	dataSize := numSamples * 2 // 16-bit mono
	fileSize := headerSize + dataSize - 8
	
	buf := make([]byte, headerSize+numSamples*2)
	
	// WAV header
	copy(buf[0:4], []byte("RIFF"))
	buf[4] = byte(fileSize)
	buf[5] = byte(fileSize >> 8)
	buf[6] = byte(fileSize >> 16)
	buf[7] = byte(fileSize >> 24)
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	buf[16] = 16  // Subchunk1Size
	buf[20] = 1   // AudioFormat (PCM)
	buf[22] = 1   // NumChannels (mono)
	buf[24] = byte(sampleRate)
	buf[25] = byte(sampleRate >> 8)
	buf[26] = byte(sampleRate >> 16)
	buf[27] = byte(sampleRate >> 24)
	buf[28] = byte(sampleRate * 2)     // ByteRate
	buf[29] = byte((sampleRate * 2) >> 8)
	buf[30] = byte((sampleRate * 2) >> 16)
	buf[31] = byte((sampleRate * 2) >> 24)
	buf[32] = 2   // BlockAlign
	buf[34] = 16  // BitsPerSample
	copy(buf[36:40], []byte("data"))
	buf[40] = byte(numSamples * 2)
	buf[41] = byte((numSamples * 2) >> 8)
	buf[42] = byte((numSamples * 2) >> 16)
	buf[43] = byte((numSamples * 2) >> 24)
	
	// Генерация синусоиды с затуханием
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		envelope := 1.0 - float64(i)/float64(numSamples) // Затухание
		
		// Синусоида
		sample := float64(32767) * 0.3 * envelope
		if soundName == "match" {
			// Восходящая частота
			currentFreq := freq * (1 + float64(i)/float64(numSamples)*0.5)
			sample *= math.Sin(2 * math.Pi * currentFreq * t)
		} else if soundName == "combo" {
			// Более сложная волна
			sample *= math.Sin(2*math.Pi*freq*t) * 0.5
			sample += math.Sin(2*math.Pi*freq*1.5*t) * 0.3
			sample += math.Sin(2*math.Pi*freq*2*t) * 0.2
		} else {
			sample *= math.Sin(2 * math.Pi * freq * t)
		}
		
		// Clipping protection
		if sample > 32767 {
			sample = 32767
		} else if sample < -32768 {
			sample = -32768
		}
		
		sampleInt := int16(sample)
		buf[headerSize+i*2] = byte(sampleInt)
		buf[headerSize+i*2+1] = byte(sampleInt >> 8)
	}
	
	return buf
}

// Play воспроизводит звук
func (sm *SoundManager) Play(sound SoundType) {
	if sm.muted {
		return
	}
	
	players, ok := sm.players[sound]
	if !ok || len(players) == 0 {
		return
	}
	
	// Случайный выбор варианта звука
	player := players[rand.Intn(len(players))]
	
	// Создание нового игрока
	p, err := sm.audioContext.NewPlayer(player)
	if err != nil {
		return
	}
	
	p.SetVolume(sm.volume)
	p.Play()
}

// SetVolume устанавливает громкость
func (sm *SoundManager) SetVolume(vol float64) {
	sm.volume = vol
}

// GetVolume возвращает текущую громкость
func (sm *SoundManager) GetVolume() float64 {
	return sm.volume
}

// ToggleMute переключает режим mute
func (sm *SoundManager) ToggleMute() bool {
	sm.muted = !sm.muted
	return sm.muted
}

// IsMuted проверяет, включён ли mute
func (sm *SoundManager) IsMuted() bool {
	return sm.muted
}

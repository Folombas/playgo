package audio

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundType типы звуков
type SoundType int

const (
	SoundPickUp SoundType = iota
	SoundDrop
	SoundSwap
	SoundMatch
	SoundCombo
	SoundFail
	SoundSnap
	SoundHint
	SoundGameOver
	SoundCount
)

// SoundManager менеджер звуков
type SoundManager struct {
	context *audio.Context
	sounds  map[SoundType][]*audio.Player
	volume  float64
	muted   bool
	rng     *rand.Rand
}

// NewSoundManager создаёт менеджер
func NewSoundManager() *SoundManager {
	sm := &SoundManager{
		context: audio.NewContext(44100),
		sounds:  make(map[SoundType][]*audio.Player),
		volume:  0.6,
		muted:   false,
		rng:     rand.New(rand.NewSource(42)),
	}

	sm.generateAllSounds()
	return sm
}

// generateAllSounds генерирует все звуки
func (sm *SoundManager) generateAllSounds() {
	// Звук поднятия - восходящий "пинг"
	sm.generateVariants(SoundPickUp, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 400,
			FreqEnd:   800,
			Duration:  0.12,
			WaveType:  WaveSine,
			Envelope:  EnvQuickAttack,
		}
	})

	// Звук опускания - нисходящий
	sm.generateVariants(SoundDrop, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 600,
			FreqEnd:   300,
			Duration:  0.15,
			WaveType:  WaveSine,
			Envelope:  EnvQuickDecay,
		}
	})

	// Звук обмена - короткий щелчок
	sm.generateVariants(SoundSwap, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 350,
			FreqEnd:   350,
			Duration:  0.08,
			WaveType:  WaveSquare,
			Envelope:  EnvClick,
		}
	})

	// Звук матча - приятная мелодия
	sm.generateVariants(SoundMatch, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 523, // C5
			FreqEnd:   784, // G5
			Duration:  0.3,
			WaveType:  WaveSine,
			Envelope:  EnvMelody,
			Melody:    true,
		}
	})

	// Звук комбо - радостный аккорд
	sm.generateVariants(SoundCombo, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 659, // E5
			FreqEnd:   1047, // C6
			Duration:  0.4,
			WaveType:  WaveSine,
			Envelope:  EnvFanfare,
			Melody:    true,
		}
	})

	// Звук ошибки - низкий буззер
	sm.generateVariants(SoundFail, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 150,
			FreqEnd:   100,
			Duration:  0.3,
			WaveType:  WaveSawtooth,
			Envelope:  EnvBuzz,
		}
	})

	// Звук прилипания - короткий клик
	sm.generateVariants(SoundSnap, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 1200,
			FreqEnd:   800,
			Duration:  0.05,
			WaveType:  WaveSine,
			Envelope:  EnvClick,
		}
	})

	// Звук подсказки - колокольчик
	sm.generateVariants(SoundHint, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 1047, // C6
			FreqEnd:   1319, // E6
			Duration:  0.25,
			WaveType:  WaveSine,
			Envelope:  EnvBell,
			Melody:    true,
		}
	})

	// Звук конца игры - грустная мелодия
	sm.generateVariants(SoundGameOver, 3, func(i int) SoundGenParams {
		return SoundGenParams{
			FreqStart: 440,
			FreqEnd:   220,
			Duration:  0.6,
			WaveType:  WaveSine,
			Envelope:  EnvSad,
			Melody:    true,
		}
	})
}

// SoundGenParams параметры генерации звука
type SoundGenParams struct {
	FreqStart float64
	FreqEnd   float64
	Duration  float64
	WaveType  WaveType
	Envelope  EnvelopeType
	Melody    bool
}

// WaveType тип волны
type WaveType int

const (
	WaveSine WaveType = iota
	WaveSquare
	WaveSawtooth
	WaveTriangle
)

// EnvelopeType тип огибающей
type EnvelopeType int

const (
	EnvQuickAttack EnvelopeType = iota
	EnvQuickDecay
	EnvClick
	EnvMelody
	EnvFanfare
	EnvBuzz
	EnvBell
	EnvSad
)

// SoundGen функция генерации одного звука
type SoundGen func(int) SoundGenParams

// generateVariants создаёт несколько вариантов звука
func (sm *SoundManager) generateVariants(soundType SoundType, count int, gen SoundGen) {
	sm.sounds[soundType] = make([]*audio.Player, 0, count)

	for i := 0; i < count; i++ {
		params := gen(i)
		wavData := sm.generateWav(params)

		player := sm.context.NewPlayerFromBytes(wavData)

		sm.sounds[soundType] = append(sm.sounds[soundType], player)
	}
}

// generateWav генерирует WAV данные
func (sm *SoundManager) generateWav(params SoundGenParams) []byte {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * params.Duration)

	// WAV header
	headerSize := 44
	dataSize := numSamples * 2 // 16-bit mono
	fileSize := headerSize + dataSize - 8

	buf := make([]byte, headerSize+numSamples*2)

	// Write header
	copy(buf[0:4], []byte("RIFF"))
	writeUint32(buf[4:], uint32(fileSize))
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	writeUint32(buf[16:], 16)             // Subchunk1Size
	writeUint16(buf[20:], 1)              // PCM
	writeUint16(buf[22:], 1)              // Mono
	writeUint32(buf[24:], uint32(44100)) // Sample rate
	writeUint32(buf[28:], uint32(88200)) // Byte rate
	writeUint16(buf[32:], 2)              // Block align
	writeUint16(buf[34:], 16)             // Bits per sample
	copy(buf[36:40], []byte("data"))
	writeUint32(buf[40:], uint32(dataSize))

	// Generate samples
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(numSamples)

		// Frequency
		freq := params.FreqStart + (params.FreqEnd-params.FreqStart)*progress

		// Wave
		var sample float64
		switch params.WaveType {
		case WaveSine:
			sample = math.Sin(2 * math.Pi * freq * t)
		case WaveSquare:
			if math.Sin(2*math.Pi*freq*t) > 0 {
				sample = 0.5
			} else {
				sample = -0.5
			}
		case WaveSawtooth:
			sample = 2*(freq*t-math.Floor(0.5+freq*t))
		case WaveTriangle:
			sample = 2 * math.Abs(2*(freq*t-math.Floor(0.5+freq*t))) - 1
		}

		// Envelope
		envelope := sm.getEnvelope(params.Envelope, progress, params.Duration)
		sample *= envelope

		// Melody - add harmonics
		if params.Melody {
			sample += 0.3 * math.Sin(2*math.Pi*freq*2*t) * envelope
			sample += 0.15 * math.Sin(2*math.Pi*freq*3*t) * envelope
		}

		// Convert to 16-bit
		value := int16(sample * 30000 * sm.volume)
		writeInt16(buf[headerSize+i*2:], value)
	}

	return buf
}

// getEnvelope возвращает значение огибающей
func (sm *SoundManager) getEnvelope(env EnvelopeType, progress, duration float64) float64 {
	switch env {
	case EnvQuickAttack:
		if progress < 0.1 {
			return progress / 0.1
		}
		return 1.0 - (progress-0.1)/0.9

	case EnvQuickDecay:
		return 1.0 - progress*progress

	case EnvClick:
		if progress < 0.05 {
			return 1.0
		}
		return math.Max(0, 1.0-(progress-0.05)/0.1)

	case EnvMelody:
		if progress < 0.05 {
			return progress / 0.05
		}
		return math.Pow(1.0-progress, 0.5)

	case EnvFanfare:
		if progress < 0.1 {
			return progress / 0.1
		}
		return math.Pow(1.0-progress, 0.7)

	case EnvBuzz:
		if progress < 0.2 {
			return 1.0
		}
		return 1.0 - (progress-0.2)/0.8

	case EnvBell:
		if progress < 0.02 {
			return progress / 0.02
		}
		return math.Pow(1.0-progress, 2)

	case EnvSad:
		return 1.0 - progress*progress

	default:
		return 1.0 - progress
	}
}

// Play воспроизводит звук
func (sm *SoundManager) Play(sound SoundType) {
	if sm.muted {
		return
	}

	players, ok := sm.sounds[sound]
	if !ok || len(players) == 0 {
		return
	}

	// Random variant
	player := players[sm.rng.Intn(len(players))]
	
	player.SetVolume(sm.volume)
	player.Play()
}

// SetVolume установка громкости
func (sm *SoundManager) SetVolume(vol float64) {
	sm.volume = math.Max(0, math.Min(1, vol))
}

// ToggleMute переключение mute
func (sm *SoundManager) ToggleMute() bool {
	sm.muted = !sm.muted
	return sm.muted
}

// Helper functions
func writeUint32(buf []byte, v uint32) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v >> 16)
	buf[3] = byte(v >> 24)
}

func writeUint16(buf []byte, v uint16) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
}

func writeInt16(buf []byte, v int16) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
}

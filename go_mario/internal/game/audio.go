package game

import (
	"io"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundEffect - структура звукового эффекта
type SoundEffect struct {
	frequency float64
	duration  int
	volume    float64
	waveform  int
	sliding   bool
}

// AudioStream - поток для воспроизведения звука
type AudioStream struct {
	buffer []float64
	pos    int
}

// AudioSystem - система воспроизведения звуков
type AudioSystem struct {
	enabled    bool
	audioCtx   *audio.Context
	soundQueue []SoundEffect
}

// Read читает данные из аудиопотока
func (a *AudioStream) Read(b []byte) (int, error) {
	if a.pos >= len(a.buffer) {
		return 0, io.EOF
	}
	n := 0
	for i := 0; i < len(b); i += 4 {
		if a.pos >= len(a.buffer) {
			break
		}
		sample := a.buffer[a.pos]
		a.pos++
		val := int16(sample * 32767)
		b[i] = byte(val)
		b[i+1] = byte(val >> 8)
		b[i+2] = 0
		b[i+3] = 0
		n += 4
	}
	return n, nil
}

// Seek перемещает позицию в потоке
func (a *AudioStream) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		a.pos = int(offset)
	case io.SeekCurrent:
		a.pos += int(offset)
	case io.SeekEnd:
		a.pos = len(a.buffer) + int(offset)
	}
	if a.pos < 0 {
		a.pos = 0
	}
	if a.pos > len(a.buffer) {
		a.pos = len(a.buffer)
	}
	return int64(a.pos), nil
}

// NewAudioSystem создаёт новую аудиосистему
func NewAudioSystem() *AudioSystem {
	audioCtx := audio.NewContext(44100)
	return &AudioSystem{
		enabled:    true,
		audioCtx:   audioCtx,
		soundQueue: make([]SoundEffect, 0),
	}
}

// generateSamples генерирует сэмплы для звука
func (as *AudioSystem) generateSamples(effect SoundEffect) []float64 {
	samples := make([]float64, effect.duration)
	phase := 0.0
	freq := effect.frequency

	for i := range samples {
		if effect.sliding {
			freq = effect.frequency * (1.0 - float64(i)/float64(effect.duration)*0.5)
		}

		var sample float64
		switch effect.waveform {
		case 0:
			sample = math.Sin(phase)
		case 1:
			if math.Sin(phase) >= 0 {
				sample = 1.0
			} else {
				sample = -1.0
			}
		case 2:
			sample = 2*(phase/(2*math.Pi)-math.Floor(phase/(2*math.Pi)+0.5))
		case 3:
			sample = (math.rand.Float64() - 0.5) * 2
		default:
			sample = math.Sin(phase)
		}

		envelope := 1.0
		if i < effect.duration/10 {
			envelope = float64(i) / float64(effect.duration/10)
		} else if i > effect.duration*7/10 {
			envelope = float64(effect.duration-i) / float64(effect.duration*3/10)
		}

		samples[i] = sample * envelope * effect.volume
		phase += 2 * math.Pi * freq / 44100
	}

	return samples
}

// playSound воспроизводит звук
func (as *AudioSystem) playSound(effect SoundEffect) {
	if !as.enabled {
		return
	}
	samples := as.generateSamples(effect)
	stream := &AudioStream{buffer: samples, pos: 0}
	player, err := as.audioCtx.NewPlayer(stream)
	if err != nil {
		return
	}
	player.Play()
}

// PlayJump - звук прыжка
func (as *AudioSystem) PlayJump() {
	as.playSound(SoundEffect{frequency: 300, duration: 250, volume: 0.25, waveform: 0, sliding: true})
}

// PlayCollect - звук сбора предмета
func (as *AudioSystem) PlayCollect() {
	as.playSound(SoundEffect{frequency: 1000, duration: 120, volume: 0.2, waveform: 0})
}

// PlayCoin - звук монеты
func (as *AudioSystem) PlayCoin() {
	as.playSound(SoundEffect{frequency: 1500, duration: 150, volume: 0.25, waveform: 0})
	as.playSound(SoundEffect{frequency: 2000, duration: 150, volume: 0.2, waveform: 0})
}

// PlayHit - звук получения урона
func (as *AudioSystem) PlayHit() {
	as.playSound(SoundEffect{frequency: 150, duration: 200, volume: 0.3, waveform: 2, sliding: true})
}

// PlayEnemyDefeat - звук победы над врагом
func (as *AudioSystem) PlayEnemyDefeat() {
	as.playSound(SoundEffect{frequency: 600, duration: 150, volume: 0.25, waveform: 0, sliding: true})
}

// PlayBlockMine - звук добычи блока
func (as *AudioSystem) PlayBlockMine() {
	as.playSound(SoundEffect{frequency: 200, duration: 80, volume: 0.2, waveform: 3})
}

// PlayBlockPlace - звук размещения блока
func (as *AudioSystem) PlayBlockPlace() {
	as.playSound(SoundEffect{frequency: 150, duration: 100, volume: 0.18, waveform: 1})
}

// PlayEnter - звук входа в дом
func (as *AudioSystem) PlayEnter() {
	as.playSound(SoundEffect{frequency: 400, duration: 200, volume: 0.2, waveform: 0})
}

// PlayJumpBump - звук при приземлении
func (as *AudioSystem) PlayJumpBump() {
	as.playSound(SoundEffect{frequency: 100, duration: 60, volume: 0.15, waveform: 3})
}

// PlayPowerup - звук бонуса
func (as *AudioSystem) PlayPowerup() {
	as.playSound(SoundEffect{frequency: 523, duration: 100, volume: 0.3, waveform: 0})
	as.playSound(SoundEffect{frequency: 659, duration: 100, volume: 0.3, waveform: 0})
	as.playSound(SoundEffect{frequency: 784, duration: 100, volume: 0.3, waveform: 0})
	as.playSound(SoundEffect{frequency: 1047, duration: 150, volume: 0.3, waveform: 0})
}

// PlayExtraLife - звук дополнительной жизни
func (as *AudioSystem) PlayExtraLife() {
	as.playSound(SoundEffect{frequency: 659, duration: 120, volume: 0.3, waveform: 0})
	as.playSound(SoundEffect{frequency: 988, duration: 120, volume: 0.3, waveform: 0})
	as.playSound(SoundEffect{frequency: 1319, duration: 200, volume: 0.3, waveform: 0})
}

// Simple Audio Manager - procedural sound effects
package main

import (
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

type AudioManager struct {
	volume float64
	ctx    *audio.Context
	mu     sync.Mutex
}

func NewAudioManager() *AudioManager {
	ctx := audio.NewContext(44100)
	return &AudioManager{
		volume: 0.4,
		ctx:    ctx,
	}
}

func (a *AudioManager) PlaySelect() {
	a.playTone(800, 0.08, 0.3, "sine")
}

func (a *AudioManager) PlaySwap() {
	a.playTone(400, 0.06, 0.2, "triangle")
}

func (a *AudioManager) PlayMatch(combo int) {
	baseFreq := 440.0 + float64(combo)*60
	a.playTone(baseFreq, 0.15, 0.4, "sine")
	a.playTone(baseFreq*1.5, 0.12, 0.25, "sine")
}

func (a *AudioManager) PlayCombo(combo int) {
	freqs := []float64{523.25, 659.25, 783.99, 1046.50}
	for i := 0; i < combo && i < len(freqs); i++ {
		freq := freqs[i%len(freqs)]
		a.playTone(freq, 0.12, 0.35, "sine")
	}
}

func (a *AudioManager) PlayNoMatch() {
	a.playTone(220, 0.12, 0.25, "sawtooth")
}

func (a *AudioManager) PlayLevelUp() {
	freqs := []float64{523.25, 659.25, 783.99, 1046.50, 1318.51}
	for _, f := range freqs {
		a.playTone(f, 0.1, 0.3, "sine")
	}
}

func (a *AudioManager) PlayMenuClick() {
	a.playTone(600, 0.04, 0.2, "sine")
}

func (a *AudioManager) PlayDrop() {
	a.playTone(350, 0.06, 0.2, "triangle")
}

func (a *AudioManager) playTone(freq, dur, vol float64, wave string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sampleRate := 44100
	samples := int(float64(sampleRate) * dur)
	buf := make([]int16, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(sampleRate)

		var val float64
		switch wave {
		case "sine":
			val = math.Sin(2 * math.Pi * freq * t)
		case "triangle":
			val = 2*math.Abs(2*(freq*t-math.Floor(freq*t+0.5))) - 1
		case "sawtooth":
			val = 2 * (freq*t - math.Floor(freq*t+0.5))
		default:
			val = math.Sin(2 * math.Pi * freq * t)
		}

		attack := 0.005
		release := 0.03
		env := 1.0
		if t < attack {
			env = t / attack
		} else if t > dur-release {
			env = (dur - t) / release
		}
		if env < 0 {
			env = 0
		}
		env *= vol * a.volume
		env *= 0.9 + rand.Float64()*0.2

		buf[i] = int16(val * env * 30000)
	}

	// Конвертируем в bytes
	byteBuf := make([]byte, len(buf)*2)
	for i, s := range buf {
		byteBuf[i*2] = byte(s)
		byteBuf[i*2+1] = byte(s >> 8)
	}

	player := a.ctx.NewPlayerFromBytes(byteBuf)
	player.Play()
}

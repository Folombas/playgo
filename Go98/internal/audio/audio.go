package audio

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	SampleRate = 44100
)

type SoundManager struct {
	context *audio.Context
	mu      sync.Mutex
}

func NewSoundManager() (*SoundManager, error) {
	ctx := audio.NewContext(SampleRate)
	return &SoundManager{context: ctx}, nil
}

// generateWave creates a simple waveform (square wave for retro sound effect)
func generateWave(frequency float64, durationSec float64, volume float64) []byte {
	samples := int(SampleRate * durationSec)
	buf := &bytes.Buffer{}

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate
		// Square wave
		val := 0.0
		if (t*frequency)-float64(int(t*frequency)) < 0.5 {
			val = volume
		} else {
			val = -volume
		}
		// Convert to 16-bit PCM
		sample := int16(val * 32767)
		binary.Write(buf, binary.LittleEndian, sample)
	}

	return buf.Bytes()
}

// PlaySound plays a procedural sound effect
func (sm *SoundManager) PlaySound(frequency float64, durationSec float64, volume float64) {
	if sm.context == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	data := generateWave(frequency, durationSec, volume)
	player := sm.context.NewPlayerFromBytes(data)
	player.Play()
}

// PlayStep plays step sound
func (sm *SoundManager) PlayStep() {
	sm.PlaySound(200, 0.05, 0.1)
}

// PlayAttack plays attack sound
func (sm *SoundManager) PlayAttack() {
	sm.PlaySound(440, 0.1, 0.3)
	sm.PlaySound(220, 0.15, 0.2)
}

// PlayHit plays hit sound
func (sm *SoundManager) PlayHit() {
	sm.PlaySound(150, 0.2, 0.4)
}

// PlayCoin plays coin pickup sound
func (sm *SoundManager) PlayCoin() {
	sm.PlaySound(880, 0.1, 0.3)
	sm.PlaySound(1100, 0.1, 0.3)
}

// PlayGem plays gem pickup sound
func (sm *SoundManager) PlayGem() {
	sm.PlaySound(660, 0.1, 0.3)
	sm.PlaySound(880, 0.1, 0.3)
	sm.PlaySound(1100, 0.15, 0.3)
}

// PlayHeal plays heal sound
func (sm *SoundManager) PlayHeal() {
	sm.PlaySound(440, 0.1, 0.2)
	sm.PlaySound(550, 0.1, 0.2)
	sm.PlaySound(660, 0.15, 0.2)
}

// PlayKey plays key pickup sound
func (sm *SoundManager) PlayKey() {
	sm.PlaySound(550, 0.1, 0.3)
	sm.PlaySound(770, 0.1, 0.3)
	sm.PlaySound(990, 0.2, 0.3)
}

// PlayDoor plays door open sound
func (sm *SoundManager) PlayDoor() {
	sm.PlaySound(330, 0.15, 0.3)
	sm.PlaySound(440, 0.2, 0.3)
}

// PlayStairs plays stairs sound (next floor)
func (sm *SoundManager) PlayStairs() {
	sm.PlaySound(440, 0.1, 0.3)
	sm.PlaySound(550, 0.1, 0.3)
	sm.PlaySound(660, 0.1, 0.3)
	sm.PlaySound(880, 0.2, 0.3)
}

// PlayDeath plays death sound
func (sm *SoundManager) PlayDeath() {
	sm.PlaySound(440, 0.2, 0.4)
	sm.PlaySound(330, 0.2, 0.4)
	sm.PlaySound(220, 0.3, 0.4)
	sm.PlaySound(110, 0.4, 0.4)
}

// PlayVictory plays victory fanfare
func (sm *SoundManager) PlayVictory() {
	notes := []float64{523, 659, 784, 1047}
	for _, freq := range notes {
		sm.PlaySound(freq, 0.2, 0.3)
	}
}

// PlayMenu plays menu selection sound
func (sm *SoundManager) PlayMenu() {
	sm.PlaySound(660, 0.05, 0.2)
}

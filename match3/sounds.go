package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundType defines available sound effects.
type SoundType int

const (
	SoundSwap SoundType = iota
	SoundMatch
	SoundError
	SoundGameOver
	SoundCascade
	SoundCombo
	SoundBonus
)

// SoundManager handles sound playback using programmatically generated waveforms.
type SoundManager struct {
	context *audio.Context
	players map[SoundType]*audio.Player
	data    map[SoundType][]byte
}

// NewSoundManager creates a new sound manager.
func NewSoundManager() *SoundManager {
	ctx := audio.NewContext(44100)
	sm := &SoundManager{
		context: ctx,
		players: make(map[SoundType]*audio.Player),
		data:    make(map[SoundType][]byte),
	}
	sm.generateSounds()
	return sm
}

// generateSounds creates simple beep/boop waveforms for each sound type.
func (sm *SoundManager) generateSounds() {
	// Swap sound: short ascending tone
	sm.data[SoundSwap] = sm.generateTone(0.1, 440, 660, "sine")

	// Match sound: pleasant chime
	sm.data[SoundMatch] = sm.generateChime(0.2, []float64{523.25, 659.25, 783.99})

	// Error sound: low buzz
	sm.data[SoundError] = sm.generateTone(0.15, 200, 150, "square")

	// Game Over sound: descending tone
	sm.data[SoundGameOver] = sm.generateTone(0.4, 660, 330, "sine")

	// Cascade sound: rising arpeggio for chain reactions
	sm.data[SoundCascade] = sm.generateChime(0.3, []float64{523.25, 659.25, 783.99, 1046.50})

	// Combo sound: exciting double chime
	sm.data[SoundCombo] = sm.generateChime(0.4, []float64{659.25, 783.99, 987.77, 1174.66, 1318.51})

	// Bonus sound: sparkle burst
	sm.data[SoundBonus] = sm.generateSparkle(0.25, 1200)

	// Create players
	for st, d := range sm.data {
		player := sm.context.NewPlayerFromBytes(d)
		if player != nil {
			sm.players[st] = player
		}
	}
}

// Play plays a sound.
func (sm *SoundManager) Play(st SoundType) {
	if player, ok := sm.players[st]; ok {
		player.SetVolume(0.5)
		player.Rewind()
		player.Play()
	}
}

// generateTone creates a simple tone sweep WAV-like bytes.
func (sm *SoundManager) generateTone(durationSec float64, freqStart, freqEnd float64, waveType string) []byte {
	sampleRate := 44100
	totalSamples := int(float64(sampleRate) * durationSec)
	data := make([]byte, totalSamples*2) // 16-bit mono

	for i := 0; i < totalSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(totalSamples)
		freq := freqStart + (freqEnd-freqStart)*progress

		var sample float64
		switch waveType {
		case "sine":
			sample = math.Sin(2 * math.Pi * freq * t)
		case "square":
			if math.Sin(2*math.Pi*freq*t) > 0 {
				sample = 0.5
			} else {
				sample = -0.5
			}
		}

		// Fade out
		envelope := 1.0 - progress*progress
		sample *= envelope * 0.3

		// Convert to 16-bit
		val := int16(sample * 32767)
		data[i*2] = byte(val)
		data[i*2+1] = byte(val >> 8)
	}

	return data
}

// generateChime creates a multi-note chime sound.
func (sm *SoundManager) generateChime(durationSec float64, freqs []float64) []byte {
	sampleRate := 44100
	totalSamples := int(float64(sampleRate) * durationSec)
	data := make([]byte, totalSamples*2)

	noteDuration := durationSec / float64(len(freqs))
	samplesPerNote := int(float64(sampleRate) * noteDuration)

	offset := 0
	for _, freq := range freqs {
		for i := 0; i < samplesPerNote && offset < totalSamples; i++ {
			t := float64(offset) / float64(sampleRate)
			sample := math.Sin(2*math.Pi*freq*t)

			// Quick attack, slow release
			progress := float64(i) / float64(samplesPerNote)
			envelope := 1.0
			if progress < 0.1 {
				envelope = progress / 0.1
			} else {
				envelope = 1.0 - (progress-0.1)/0.9*0.5
			}
			sample *= envelope * 0.25

			val := int16(sample * 32767)
			data[offset*2] = byte(val)
			data[offset*2+1] = byte(val >> 8)
			offset++
		}
	}

	return data
}

// generateSparkle creates a sparkling burst sound with high frequencies.
func (sm *SoundManager) generateSparkle(durationSec float64, baseFreq float64) []byte {
	sampleRate := 44100
	totalSamples := int(float64(sampleRate) * durationSec)
	data := make([]byte, totalSamples*2)

	for i := 0; i < totalSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(totalSamples)

		// Multiple harmonics for sparkle effect
		sample := 0.0
		sample += math.Sin(2*math.Pi*baseFreq*t) * 0.3
		sample += math.Sin(2*math.Pi*baseFreq*1.5*t) * 0.2
		sample += math.Sin(2*math.Pi*baseFreq*2*t) * 0.15
		sample += math.Sin(2*math.Pi*baseFreq*2.5*t) * 0.1

		// Quick attack, exponential decay
		envelope := 1.0 - math.Pow(progress, 0.5)
		sample *= envelope * 0.25

		val := int16(sample * 32767)
		data[i*2] = byte(val)
		data[i*2+1] = byte(val >> 8)
	}

	return data
}

package main

import (
	"bytes"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundType — тип звука
type SoundType int

const (
	SoundMove SoundType = iota
	SoundRotate
	SoundHardDrop
	SoundPlace
	SoundLineClear
	SoundGameOver
)

var audioContext *audio.Context

// initAudio инициализирует аудио
func initAudio() {
	audioContext = audio.NewContext(44100)
}

// generateBeep создаёт простой звуковой сигнал
func generateBeep(frequency float64, duration int, volume float32) []byte {
	buf := &bytes.Buffer{}
	sampleRate := 44100
	numSamples := sampleRate * duration / 1000

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		// Синусоидальная волна
		sample := math.Sin(2 * math.Pi * frequency * t)

		// Fade out
		envelope := 1.0 - float64(i)/float64(numSamples)
		sample *= envelope * float64(volume)

		// Конвертировать в PCM 16-bit
		val := int16(sample * 32767)
		buf.WriteByte(byte(val))
		buf.WriteByte(byte(val >> 8))
	}
	return buf.Bytes()
}

// SoundCache кэширует звуки
var soundCache = map[SoundType][]byte{}

// initSounds инициализирует звуки
func initSounds() {
	soundCache[SoundMove] = generateBeep(440, 50, 0.1)
	soundCache[SoundRotate] = generateBeep(660, 50, 0.1)
	soundCache[SoundHardDrop] = generateBeep(220, 100, 0.2)
	soundCache[SoundPlace] = generateBeep(330, 80, 0.15)
	soundCache[SoundLineClear] = generateBeep(880, 150, 0.2)
	soundCache[SoundGameOver] = generateBeep(110, 300, 0.3)
}

// PlaySound воспроизводит звук
func PlaySound(t SoundType) {
	if audioContext == nil {
		return
	}

	data, ok := soundCache[t]
	if !ok {
		return
	}

	// В простой реализации просто генерируем
	// Для полной реализации нужен audio.Player
	_ = data
}

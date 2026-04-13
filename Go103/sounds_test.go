package main

import (
	"sync"
	"testing"
)

var (
	soundTestOnce sync.Once
	soundTestMgr  *SoundManager
)

func getSoundManagerForTest() *SoundManager {
	soundTestOnce.Do(func() {
		soundTestMgr = NewSoundManager()
	})
	return soundTestMgr
}

func TestSoundManagerCreation(t *testing.T) {
	sm := getSoundManagerForTest()
	if sm == nil {
		t.Fatal("SoundManager should not be nil")
	}
	if sm.context == nil {
		t.Fatal("SoundManager context should not be nil")
	}
}

func TestSoundManagerHasAllSounds(t *testing.T) {
	sm := getSoundManagerForTest()

	expectedSounds := []SoundType{SoundSwap, SoundMatch, SoundError, SoundGameOver}
	for _, st := range expectedSounds {
		if _, ok := sm.players[st]; !ok {
			t.Errorf("SoundManager missing sound type: %v", st)
		}
	}
}

func TestGenerateTone(t *testing.T) {
	sm := getSoundManagerForTest()
	data := sm.generateTone(0.1, 440, 660, "sine")
	if len(data) == 0 {
		t.Error("Generated tone data should not be empty")
	}
	expectedLen := 4410 * 2
	if len(data) != expectedLen {
		t.Errorf("Tone length = %d, expected %d", len(data), expectedLen)
	}
}

func TestGenerateChime(t *testing.T) {
	sm := getSoundManagerForTest()
	data := sm.generateChime(0.2, []float64{523.25, 659.25, 783.99})
	if len(data) == 0 {
		t.Error("Generated chime data should not be empty")
	}
}

func TestPlaySound(t *testing.T) {
	sm := getSoundManagerForTest()
	sm.Play(SoundSwap)
	sm.Play(SoundMatch)
	sm.Play(SoundError)
	sm.Play(SoundGameOver)
}

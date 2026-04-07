package audio

import (
	"bytes"
	"embed"
	"io"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

//go:embed assets/sounds/*
var soundsFS embed.FS

// AudioManager управляет звуками и музыкой
type AudioManager struct {
	audioContext *audio.Context
	musicPlayer  *audio.Player
	sounds       map[string]*audio.Player
	mu           sync.RWMutex
	volume       float64
}

// NewAudioManager создает новый аудио менеджер
func NewAudioManager() *AudioManager {
	am := &AudioManager{
		sounds: make(map[string]*audio.Player),
		volume: 0.5,
	}

	// Инициализация аудио контекста (44100 Hz)
	context, err := audio.NewContext(44100)
	if err != nil {
		log.Printf("Warning: audio context error: %v", err)
		return am
	}

	am.audioContext = context
	return am
}

// Load загружает звуки
func (am *AudioManager) Load() error {
	if am.audioContext == nil {
		return nil
	}

	// Загрузка звуков совпадений
	for i := 1; i <= 5; i++ {
		filename := "assets/sounds/match_0" + string(rune('0'+i)) + ".wav"
		player, err := am.loadWAV(filename)
		if err != nil {
			log.Printf("Warning: could not load %s: %v", filename, err)
			continue
		}
		am.sounds["match_"+string(rune('0'+i))] = player
	}

	// Загрузка звука выбора
	selectPlayer, err := am.loadWAV("assets/sounds/select.wav")
	if err == nil {
		am.sounds["select"] = selectPlayer
	}

	// Загрузка фоновой музыки
	bgmPlayer, err := am.loadMP3("assets/sounds/bgm.mp3")
	if err == nil {
		am.musicPlayer = bgmPlayer
	}

	log.Printf("Loaded %d sounds", len(am.sounds))
	return nil
}

func (am *AudioManager) loadWAV(path string) (*audio.Player, error) {
	data, err := soundsFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	player, err := am.audioContext.NewPlayer(reader)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (am *AudioManager) loadMP3(path string) (*audio.Player, error) {
	data, err := soundsFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	player, err := am.audioContext.NewPlayerFromReader(reader)
	if err != nil {
		return nil, err
	}

	player.SetVolume(am.volume)
	return player, nil
}

// PlayMatchSound воспроизводит звук совпадения
func (am *AudioManager) PlayMatchSound(combo int) {
	if am.audioContext == nil {
		return
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	// Выбор звука в зависимости от комбо
	soundKey := "match_01"
	if combo > 1 && combo <= 5 {
		soundKey = "match_0" + string(rune('0'+combo))
	}

	if player, ok := am.sounds[soundKey]; ok {
		player.Rewind()
		player.Play()
	}
}

// PlaySelectSound воспроизводит звук выделения
func (am *AudioManager) PlaySelectSound() {
	if am.audioContext == nil {
		return
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	if player, ok := am.sounds["select"]; ok {
		player.Rewind()
		player.Play()
	}
}

// PlayBGM запускает фоновую музыку
func (am *AudioManager) PlayBGM() {
	if am.audioContext == nil || am.musicPlayer == nil {
		return
	}

	am.musicPlayer.SetVolume(am.volume)
	am.musicPlayer.Rewind()
	am.musicPlayer.Play()
}

// StopBGM останавливает фоновую музыку
func (am *AudioManager) StopBGM() {
	if am.musicPlayer != nil {
		am.musicPlayer.Pause()
	}
}

// SetVolume устанавливает громкость
func (am *AudioManager) SetVolume(vol float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.volume = vol
	
	if am.musicPlayer != nil {
		am.musicPlayer.SetVolume(vol)
	}
}

// Close закрывает аудио контекст
func (am *AudioManager) Close() {
	if am.audioContext != nil {
		am.audioContext.Close()
	}
}

// Ensure io import is used
var _ io.Reader

package audio

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const SampleRate = 44100

// Manager handles all audio
type Manager struct {
	ctx *audio.Context
	mu  sync.Mutex
}

// NewManager creates audio manager
func NewManager() (*Manager, error) {
	ctx := audio.NewContext(SampleRate)
	return &Manager{ctx: ctx}, nil
}

func (m *Manager) playTone(freq float64, duration float64, volume float64) {
	if m.ctx == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	samples := int(SampleRate * duration)
	buf := &bytes.Buffer{}

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate
		// Sine wave with fade out
		val := 0.0
		val = math.Sin(2*math.Pi*freq*t) * (1 - float64(i)/float64(samples))
		
		sample := int16(val * volume * 32767)
		binary.Write(buf, binary.LittleEndian, sample)
	}

	player := m.ctx.NewPlayerFromBytes(buf.Bytes())
	player.Play()
}

// PlayShoot plays shoot sound
func (m *Manager) PlayShoot() {
	m.playTone(800, 0.1, 0.2)
}

// PlayExplosion plays explosion sound
func (m *Manager) PlayExplosion() {
	m.playTone(150, 0.3, 0.4)
	m.playTone(100, 0.4, 0.3)
}

// PlayCoin plays coin sound
func (m *Manager) PlayCoin() {
	m.playTone(1200, 0.1, 0.2)
	m.playTone(1600, 0.15, 0.2)
}

// PlayPlace plays tower place sound
func (m *Manager) PlayPlace() {
	m.playTone(440, 0.1, 0.2)
	m.playTone(660, 0.15, 0.2)
}

// PlayUpgrade plays upgrade sound
func (m *Manager) PlayUpgrade() {
	m.playTone(523, 0.1, 0.2)
	m.playTone(659, 0.1, 0.2)
	m.playTone(784, 0.2, 0.2)
}

// PlayEnemyDeath plays enemy death sound
func (m *Manager) PlayEnemyDeath() {
	m.playTone(400, 0.15, 0.3)
	m.playTone(300, 0.2, 0.3)
}

// PlayWaveStart plays wave start sound
func (m *Manager) PlayWaveStart() {
	m.playTone(330, 0.1, 0.2)
	m.playTone(440, 0.1, 0.2)
	m.playTone(550, 0.2, 0.2)
}

// PlayGameOver plays game over sound
func (m *Manager) PlayGameOver() {
	m.playTone(440, 0.3, 0.3)
	m.playTone(330, 0.3, 0.3)
	m.playTone(220, 0.5, 0.3)
}

// PlayVictory plays victory sound
func (m *Manager) PlayVictory() {
	notes := []float64{523, 659, 784, 1047}
	for _, n := range notes {
		m.playTone(n, 0.3, 0.2)
	}
}

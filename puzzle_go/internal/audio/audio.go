// Package audio — процедурная генерация звуков для match-3.
package audio

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

// Manager хранит аудио контекст и все звуки.
type Manager struct {
	ctx    *audio.Context
	Match  *audio.Player
	Swap   *audio.Player
	Bad    *audio.Player
	Combo  *audio.Player
	Win    *audio.Player
}

// NewManager создаёт менеджер и генерирует все звуки.
func NewManager() *Manager {
	m := &Manager{}
	m.ctx = audio.NewContext(sampleRate)
	m.Match = m.newPlayer(arp(0.2, []float64{523, 659, 784}))
	m.Swap = m.newPlayer(tone(0.1, 440))
	m.Bad = m.newPlayer(noise(0.15))
	m.Combo = m.newPlayer(arp(0.3, []float64{523, 659, 784, 1047}))
	m.Win = m.newPlayer(arp(0.5, []float64{523, 659, 784, 1047, 1319}))
	return m
}

func (m *Manager) newPlayer(samples []int16) *audio.Player {
	p, _ := audio.NewPlayer(m.ctx, bytes.NewReader(mkWAV(samples)))
	return p
}

// Play перезапускает и играет звук. Безопасен при nil.
func Play(p *audio.Player) {
	if p == nil { return }
	p.Rewind()
	p.Play()
}

func mkWAV(s []int16) []byte {
	buf := &bytes.Buffer{}
	w := func(data interface{}) { binary.Write(buf, binary.LittleEndian, data) }
	w([]byte("RIFF"))
	w(uint32(36 + len(s)*2))
	w([]byte("WAVE"))
	w([]byte("fmt "))
	w(uint32(16))
	w(uint16(1))
	w(uint16(1))
	w(uint32(sampleRate))
	w(uint32(sampleRate * 2))
	w(uint16(2))
	w(uint16(16))
	w([]byte("data"))
	w(uint32(len(s) * 2))
	for _, v := range s { w(v) }
	return buf.Bytes()
}

func tone(dur float64, freq float64) []int16 {
	n := int(float64(sampleRate) * dur)
	s := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sampleRate)
		s[i] = int16(math.Sin(2*math.Pi*freq*t) * 32767 * 0.5)
	}
	return s
}

func noise(dur float64) []int16 {
	n := int(float64(sampleRate) * dur)
	s := make([]int16, n)
	for i := 0; i < n; i++ {
		s[i] = int16((rand.Float64()*2 - 1) * 32767 * 0.3)
	}
	return s
}

func arp(dur float64, fs []float64) []int16 {
	n := int(float64(sampleRate) * dur)
	s := make([]int16, n)
	seg := n / len(fs)
	for idx, f := range fs {
		start := idx * seg
		end := start + seg
		if idx == len(fs)-1 { end = n }
		for i := start; i < end; i++ {
			t := float64(i-start) / float64(seg)
			env := 1.0 - t
			s[i] = int16(math.Sin(2*math.Pi*f*float64(i)/float64(sampleRate)) * 32767 * 0.3 * env)
		}
	}
	return s
}

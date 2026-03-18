package music

import (
	"bytes"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

// --- Assets ---
//
//go:embed assets/songs/classroom_0_song.mp3
var classroom0 []byte

// Add others: //go:embed assets/battle.mp3 ...

type AudioManager struct {
	Context   *audio.Context
	Current   *audio.Player
	Next      *audio.Player
	MaxVolume float64
	FadeSpeed float64
	isPaused  bool
}

func NewAudioManager(ctx *audio.Context) *AudioManager {
	return &AudioManager{
		Context:   ctx,
		MaxVolume: 0.5,
		FadeSpeed: 0.005,
	}
}

func (m *AudioManager) Update() {
	if m.isPaused {
		return
	}

	// 1. Handle Fade Out & Swap
	if m.Current != nil && m.Next != nil {
		vol := m.Current.Volume()
		if vol > 0 {
			newVol := vol - m.FadeSpeed
			if newVol < 0 {
				newVol = 0
			}
			m.Current.SetVolume(newVol)
		} else {
			m.Current.Close()
			m.Current = m.Next
			m.Next = nil
		}
	}

	// 2. Handle Fade In
	if m.Current != nil && m.Next == nil {
		vol := m.Current.Volume()
		if vol < m.MaxVolume {
			newVol := vol + m.FadeSpeed
			if newVol > m.MaxVolume {
				newVol = m.MaxVolume
			}
			m.Current.SetVolume(newVol)
		}
	}
}

func (m *AudioManager) ChangeSong(name string) error {
	var data []byte
	switch name {
	case "classroom":
		data = classroom0
	default:
		return nil
	}

	stream, err := mp3.DecodeF32(bytes.NewReader(data))
	if err != nil {
		return err
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())
	p, err := m.Context.NewPlayerF32(loop)
	if err != nil {
		return err
	}

	p.SetVolume(0)
	p.Play()

	if m.Current == nil {
		m.Current = p
	} else {
		// Replace pending Next if one exists
		if m.Next != nil {
			m.Next.Close()
		}
		m.Next = p
	}
	return nil
}

// --- Pause / Resume Logic ---

func (m *AudioManager) Pause() {
	if m.Current != nil && m.Current.IsPlaying() {
		m.Current.Pause()
	}
	if m.Next != nil && m.Next.IsPlaying() {
		m.Next.Pause()
	}
	m.isPaused = true
}

func (m *AudioManager) Resume() {
	m.isPaused = false
	if m.Current != nil && !m.Current.IsPlaying() {
		m.Current.Play()
	}
	// We don't necessarily resume 'Next' here because it
	// only starts playing once the swap happens in Update()
}

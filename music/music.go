package music

import (
	"bytes"
	_ "embed"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// --- Assets ---
//
//go:embed assets/songs/classroom_0_song.mp3
var classroom0 []byte

//go:embed assets/songs/scene_change_riff.mp3
var scene_change_riff []byte

// Add others: //go:embed assets/battle.mp3 ...

// --- SFX Embeds ---
//
//go:embed assets/sfx/zoingggg.wav
var zoingWav []byte

//go:embed assets/sfx/bloop.wav
var bloopWav []byte

//go:embed assets/sfx/blip463.wav
var blipWav []byte

//go:embed assets/sfx/pickup496.wav
var pickupWav []byte

// sfxPool stores pre-decoded raw audio data
var sfxPool = make(map[string][]byte)

// PreloadSFX decodes all wav files into memory once.
// Call this in your NewGame() function.
func PreloadSFX(ctx *audio.Context) error {
	load := func(name string, b []byte) error {
		// Decode the WAV
		d, err := wav.DecodeF32(bytes.NewReader(b))
		if err != nil {
			return err
		}
		// Read the entire stream into a byte slice (the "Pool")
		raw, err := io.ReadAll(d)
		if err != nil {
			return err
		}
		sfxPool[name] = raw
		return nil
	}

	if err := load("zoing", zoingWav); err != nil {
		return err
	}
	if err := load("bloop", bloopWav); err != nil {
		return err
	}
	if err := load("blip", blipWav); err != nil {
		return err
	}
	if err := load("pickup", pickupWav); err != nil {
		return err
	}

	return nil
}

// PlaySE plays a sound effect from the pool immediately.
func (m *AudioManager) PlaySE(name string) {
	raw, ok := sfxPool[name]
	if !ok {
		return
	}
	// Create a player directly from the pre-decoded bytes
	// This is nearly instantaneous.
	sePlayer := m.Context.NewPlayerF32FromBytes(raw)
	// Apply the SFX specific volume
	sePlayer.SetVolume(m.SFXVolume)
	sePlayer.Play()
}

type AudioManager struct {
	Context   *audio.Context
	Current   *audio.Player
	Next      *audio.Player
	MaxVolume float64
	SFXVolume float64 // Sound Effect Volume (e.g., 0.2)
	FadeSpeed float64
	isPaused  bool
}

func NewAudioManager(ctx *audio.Context) *AudioManager {
	return &AudioManager{
		Context:   ctx,
		MaxVolume: 0.5,
		SFXVolume: 0.2,
		// FadeSpeed: 0.005,
		FadeSpeed: 0.05,
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
	case "scenechange":
		data = scene_change_riff
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

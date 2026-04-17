package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/music"
)

type Game struct {
	scene        Scene
	assets       *Assets
	player       *Player
	debug        bool
	audioManager *music.AudioManager
	gameplay     *GameplayState
}

func NewGame() *Game {
	// 1. Audio Setup - Catch the error instead of panicking
	audioContext := audio.NewContext(48000)

	var manager *music.AudioManager

	// Check if the Pi actually initialized an audio device
	if audioContext != nil {
		if err := music.PreloadSFX(audioContext); err != nil {
			// Print the error to terminal, but don't stop the game
			fmt.Printf("Warning: Audio hardware found but SFX failed: %v\n", err)
		} else {
			manager = music.NewAudioManager(audioContext)
			manager.SFXVolume = 0.15
		}
	} else {
		fmt.Println("Running in SILENT MODE: No audio device detected on Pi 5.")
	}

	// 2. Asset Setup
	assets := LoadAssets()
	player := NewPlayer()
	player.image = assets.DefaultPlayer

	// 3. Create the Game struct
	g := &Game{
		assets:       assets,
		player:       player,
		debug:        false,
		audioManager: manager, // This will be nil if audio failed
		gameplay:     NewGameplayState(assets.ObjectsTileset),
	}

	// 4. Initialize the starting scene
	g.scene = NewTitleScene(g)

	return g
}

func (g *Game) Update() error {
	if g.audioManager != nil {
		g.audioManager.Update()
	}
	if g.scene != nil {
		if err := g.scene.Update(); err != nil {
			return err
		}
	}
	if gpad.PressFullscreen() {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if gpad.PressDebug() {
		g.debug = !g.debug
		fmt.Println("Debug mode on: ", g.debug)
	}
	if g.debug {
		gpad.TestInputs()
		DebugJumpToLevel(g)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
	if g.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n\n\n\n\nTPS: %0.2f", ebiten.ActualTPS()))
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(sW*sX, sH*sX)
	gpad.Init(sW, sH)
	ebiten.SetWindowTitle("Play Station 41")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

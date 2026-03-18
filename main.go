package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/music"
)

type Game struct {
	scene        Scene
	assets       *Assets
	player       *Player
	debug        bool
	audioManager *music.AudioManager
}

func NewGame() *Game {
	// 1. Audio Setup
	audioContext := audio.NewContext(48000)
	if err := music.PreloadSFX(audioContext); err != nil {
		log.Fatal(err)
	}
	manager := music.NewAudioManager(audioContext)

	// 2. Asset Setup
	assets := LoadAssets()
	player := NewPlayer()
	player.image = assets.DefaultPlayer

	// 3. Create the Game struct pointer FIRST
	g := &Game{
		assets:       assets,
		player:       player,
		debug:        false,
		audioManager: manager,
	}
	g.audioManager.SFXVolume = 0.15

	// 4. NOW initialize the starting scene and assign it
	// This ensures g.scene is NOT nil when Update() runs
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
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// ebitenutil.DebugPrint(screen, "Play Station 41\nAKA PS41")
	g.scene.Draw(screen)
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

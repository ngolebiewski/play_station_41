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

	// 4. Create the Game struct pointer FIRST
	g := &Game{
		assets:       assets,
		player:       player,
		debug:        false,
		audioManager: manager,
		gameplay:     NewGameplayState(assets.ObjectsTileset),
	}
	g.audioManager.SFXVolume = 0.15

	// 4. NOW initialize the starting scene and assign it
	// This ensures g.scene is NOT nil when Update() runs
	g.scene = NewTitleScene(g)

	return g
}

func (g *Game) Update() error {
	// drawCounter++
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

// var drawCounter int

func (g *Game) Draw(screen *ebiten.Image) {

	// Only draw every 3rd frame (effectively 20 FPS)
	// but Update() still runs 60 times a second.
	// fmt.Println(drawCounter)
	// if drawCounter%7 != 0 {
	// 	fmt.Println(ebiten.ActualTPS())
	// 	return
	// }
	g.scene.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(sW*sX, sH*sX)
	ebiten.SetTPS(10)
	ebiten.SetVsyncEnabled(false)

	gpad.Init(sW, sH)
	ebiten.SetWindowTitle("Play Station 41")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

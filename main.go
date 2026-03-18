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
	///////////////////////SOUND///////////////////////////////
	// Create the Ebitengine audio context and audio managaer
	audioContext := audio.NewContext(48000)
	// Decode all SFX into RAM
	err := music.PreloadSFX(audioContext)
	if err != nil {
		log.Fatal(err)
	}
	manager := music.NewAudioManager(audioContext)
	///////////////////////////////////////////////////////////

	// Load embedded assets (spritesheets + tilemaps) and initialize a Player
	assets := LoadAssets()
	player := NewPlayer()
	player.image = assets.DefaultPlayer

	g := &Game{
		assets:       assets,
		player:       player,
		debug:        false,
		audioManager: manager,
	}
	g.scene = NewTitleScene(g)
	return g
}

func (g *Game) Update() error {
	g.audioManager.Update()
	g.scene.Update()
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

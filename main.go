package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
)

type Player struct{
	x,y float32
}

type Game struct{}

// State Machine for Scenes
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

func (g *Game) Update() error {
	if gpad.PressFullscreen() {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	gpad.TestInputs()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Play Station 41\nAKA PS41")

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(sW*sX,sH*sX)
	ebiten.SetWindowTitle("Play Station 41")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

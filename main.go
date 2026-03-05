package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
)


type Game struct{}

func (g *Game) Update() error {
	if gpad.MoveLeft(){fmt.Print("left")}
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

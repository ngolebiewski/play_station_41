package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	sW = 240
	sH = 160
	sX = 2
)
type Game struct{}

func (g *Game) Update() error {
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

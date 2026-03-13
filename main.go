package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
)

type Game struct{
	scene Scene
	assets *Assets 
}


func NewGame() *Game {
	assets := LoadAssets()
	g := &Game{
		assets: assets,
	}
	g.scene = NewTitleScene(g)
	return g
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
	g.scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(sW*sX,sH*sX)
	ebiten.SetWindowTitle("Play Station 41")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

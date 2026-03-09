package main

import (
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TitleScene struct {
	game *Game
}

var gameFont = text.NewGoXFace(bitmapfont.Face)

var textStr = "START"
var textW, textH = text.Measure(textStr, gameFont, 0)

func NewTitleScene(game *Game) *TitleScene {
	return &TitleScene{game: game}
}

func (s *TitleScene) Update() error {
	return nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	// --- Draw button under text ---
	buttonW := textW + 10 // padding
	buttonH := textH 
	buttonX := sW/2 - buttonW/2 
	buttonY := sH*7/8  - textH/2 

	red := color.RGBA{150, 0, 0, 255} // deep red
	ebitenutilDrawRect(screen, float64(buttonX), float64(buttonY), float64(buttonW), buttonH, red)

	// --- Draw text ---
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(
		sW/2-textW/2,
		sH*7/8-textH/2,
	)
	opts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, textStr, gameFont, opts)
}

// Simple helper to draw filled rectangle (like ebitenutil.DrawRect)
func ebitenutilDrawRect(screen *ebiten.Image, x, y, w, h float64, clr color.Color) {
	rect := ebiten.NewImage(int(w), int(h))
	rect.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(rect, op)
}
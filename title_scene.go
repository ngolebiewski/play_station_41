package main

import (
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type TitleScene struct {
	game *Game
	img  *ebiten.Image
}

var gameFont = text.NewGoXFace(bitmapfont.Face)

var textStr = "START"
var textW, textH = text.Measure(textStr, gameFont, 0)

func NewTitleScene(game *Game) *TitleScene {
	return &TitleScene{game: game,
		img: game.assets.TitleImage}
}

func (s *TitleScene) Update() error {
	return nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	// size := s.img.Bounds().Size()
	// op.GeoM.Translate(
	// 	float64((sW-size.X)/2),
	// 	float64((sH-size.Y)/2),
	// )
	screen.DrawImage(s.img, op)

	// --- Draw button under text ---
	buttonW := textW + 10 // padding
	buttonH := textH
	buttonX := sW/2 - buttonW/2
	buttonY := sH*7/8 - textH/2

	// red := color.RGBA{150, 0, 0, 255} // deep red

	vector.FillRect(screen, float32(buttonX), float32(buttonY), float32(buttonW), float32(buttonH), color.Black, false)

	// --- Draw text ---
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(
		sW/2-textW/2,
		sH*7/8-textH/2,
	)
	opts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, textStr, gameFont, opts)

}

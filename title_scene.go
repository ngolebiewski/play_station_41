package main

import (
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/play_station_41/gpad"
)

type TitleScene struct {
	game *Game
	img  *ebiten.Image
}

var gameFont = text.NewGoXFace(bitmapfont.Face)
var textStr = "START"
var textW, textH = text.Measure(textStr, gameFont, 0)

func NewTitleScene(game *Game) *TitleScene {

	return &TitleScene{game: game, img: game.assets.TitleImage}
}

func (s *TitleScene) Update() error {
	// Detect first touch — activates touch controls for the whole session
	if !gpad.TouchEnabled() && len(ebiten.AppendTouchIDs(nil)) > 0 {
		gpad.EnableTouch()
	}

	// Update touch d-pad state each tick
	gpad.UpdateTouch()

	// Tap anywhere on title screen OR press B to start
	touchTapped := gpad.TouchEnabled() && len(inpututil.AppendJustReleasedTouchIDs(nil)) > 0
	if gpad.PressB() || gpad.PressStart() || touchTapped {
		s.game.scene = NewCharacterSelectionScene(s.game)
	}

	return nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(s.img, op)

	// Button background
	buttonW := textW + 10
	buttonH := textH
	buttonX := sW/2 - buttonW/2
	buttonY := sH*7/8 - textH/2
	vector.FillRect(screen, float32(buttonX), float32(buttonY), float32(buttonW), float32(buttonH), color.Black, false)

	// START text
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(sW/2-textW/2, sH*7/8-textH/2)
	opts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, textStr, gameFont, opts)

	// Hint text when touch is active
	if gpad.TouchEnabled() {
		hint := "TAP TO START"
		hw, _ := text.Measure(hint, gameFont, 0)
		hopts := &text.DrawOptions{}
		hopts.GeoM.Translate(sW/2-hw/2, sH*7/8-textH/2-12)
		hopts.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, hint, gameFont, hopts)
	}
}

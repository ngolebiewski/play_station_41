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
	demoTimer int
	sparkleTimer int
	sparkleFlag bool
}

var gameFont = text.NewGoXFace(bitmapfont.Face)
var textStr = "START"
var textW, textH = text.Measure(textStr, gameFont, 0)

func NewTitleScene(game *Game) *TitleScene {

	return &TitleScene{game: game, img: game.assets.TitleImage, demoTimer: 0, sparkleTimer: 0, sparkleFlag: false}
}

func (s *TitleScene) Update() error {
	// Detect first touch — activates touch controls for the whole session
	if !gpad.TouchEnabled() && len(ebiten.AppendTouchIDs(nil)) > 0 {
		gpad.EnableTouch()
	}

	// Update touch d-pad state each tick
	gpad.UpdateTouch()

	// Increment demo timer
	s.demoTimer++

	// Increment sparkle timer
	s.sparkleTimer++
	if s.sparkleTimer%180 < 15 {
		s.sparkleFlag = true
	} else {
		s.sparkleFlag = false
	}

	// If demo timer > 600 (10 seconds at 60fps), switch to demo scene
	if s.demoTimer > 600 {
		s.game.scene = NewDemoScene(s.game)
		return nil
	}

	// Tap anywhere on title screen OR press B to start
	touchTapped := gpad.TouchEnabled() && len(inpututil.AppendJustReleasedTouchIDs(nil)) > 0
	if gpad.PressB() || gpad.PressStart() || touchTapped {
		s.demoTimer = 0 // Reset demo timer on input
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

	// Sparkle effect on start button
	if s.sparkleFlag {
		sparkleX := sW/2 + textW/2 + 5
		sparkleY := sH*7/8 - textH/2 - 5
		vector.FillRect(screen, float32(sparkleX), float32(sparkleY), 10, 10, color.RGBA{255, 255, 0, 255}, false)
	}

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

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ngolebiewski/play_station_41/gpad"
)

const creditsDuration = 5 * 60 // 1 seconds = 60fps

type CreditsScene struct {
	game         *Game
	framecounter int
}

func NewCreditsScene(game *Game) *CreditsScene {
	return &CreditsScene{
		game:         game,
		framecounter: 0,
	}
}

func (s *CreditsScene) Update() error {
	s.framecounter++
	gpad.UpdateTouch()

	// Advance after 2 seconds or on any button press
	if s.framecounter >= creditsDuration || gpad.PressA() || gpad.PressB() || gpad.PressStart() {
		s.game.scene = NewTitleScene(s.game)
	}
	return nil
}

func (s *CreditsScene) Draw(screen *ebiten.Image) {
	// Black background
	overlay := ebiten.NewImage(sW, sH)
	overlay.Fill(color.RGBA{0, 0, 0, 255})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	// "Credits" title
	titleOpt := &text.DrawOptions{}
	titleOpt.GeoM.Translate(100, 10)
	titleOpt.ColorScale.ScaleWithColor(color.RGBA{0, 225, 0, 255})
	text.Draw(screen, "Credits", transitionTextFace, titleOpt)

	// Pixel art credit line
	artOpt := &text.DrawOptions{}
	artOpt.GeoM.Translate(20, 40)
	artOpt.ColorScale.ScaleWithColor(color.RGBA{225, 225, 1, 225})
	text.Draw(screen, "Pixel Art Self Portraits and", transitionTextFace, artOpt)

	artOpt2 := &text.DrawOptions{}
	artOpt2.GeoM.Translate(20, 55)
	artOpt2.ColorScale.ScaleWithColor(color.RGBA{225, 225, 1, 225})
	text.Draw(screen, "Classroom Object Art:", transitionTextFace, artOpt2)

	artOpt3 := &text.DrawOptions{}
	artOpt3.GeoM.Translate(20, 70)
	artOpt3.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, "Class 3-410 Ms. Kim and Ms. G", transitionTextFace, artOpt3)

	// Game design credit line
	designOpt := &text.DrawOptions{}
	designOpt.GeoM.Translate(20, 90)
	designOpt.ColorScale.ScaleWithColor(color.RGBA{225, 225, 1, 225})
	text.Draw(screen, "Game Design, Code, Levels,", transitionTextFace, designOpt)

	designOpt2 := &text.DrawOptions{}
	designOpt2.GeoM.Translate(20, 105)
	designOpt2.ColorScale.ScaleWithColor(color.RGBA{225, 225, 1, 225})
	text.Draw(screen, "Classroom Art:", transitionTextFace, designOpt2)

	designOpt3 := &text.DrawOptions{}
	designOpt3.GeoM.Translate(20, 120)
	designOpt3.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, "Nick Golebiewski", transitionTextFace, designOpt3)

	designOpt4 := &text.DrawOptions{}
	designOpt4.GeoM.Translate(108, 140)
	designOpt4.ColorScale.ScaleWithColor(color.RGBA{0, 255, 0, 255})
	text.Draw(screen, "2026", transitionTextFace, designOpt4)
}

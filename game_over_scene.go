package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ngolebiewski/play_station_41/gpad"
)

// Create text face for game over screen
var gameOverTextFace = text.NewGoXFace(bitmapfont.Face)

const (
	gameOverDuration = 300 // frames before auto-transitioning
)

type GameOverScene struct {
	game           *Game
	framecounter   int
	isTimeUp       bool
	selectedOption int // 0 = Try Again, 1 = Start Over (when not GameOver)
}

func NewGameOverScene(game *Game, isTimeUp bool) *GameOverScene {
	/////////////////////////////////////////////////////
	// Start the music!
	if game.audioManager != nil {
		err := game.audioManager.ChangeSong("running")
		if err != nil {
			log.Printf("Audio Error: %v", err)
		}
	}
	/////////////////////////////////////////////////////

	return &GameOverScene{
		game:           game,
		framecounter:   0,
		isTimeUp:       isTimeUp,
		selectedOption: 0,
	}
}

func (s *GameOverScene) Update() error {
	s.framecounter++

	if s.framecounter > 20 {
		gpad.UpdateTouch()

		gp := s.game.gameplay

		// Menu navigation
		if gpad.MoveUp() {
			s.selectedOption = 0
		}
		if gpad.MoveDown() {
			if gp.GameOver {
				s.selectedOption = 0
			} else {
				s.selectedOption = 1
			}
		}

		// Select option
		if gpad.PressB() || gpad.PressStart() {
			if s.isTimeUp && gp.Lives > 0 {
				// Try Again or Start Over options
				if s.selectedOption == 0 {
					// Try Again - retry same level with same layout
					gp.IsRetryingLevel = true
					gp.TimerTriggered = false
					gp.ObjectsFound = 0
					s.game.scene = NewClassroomScene(s.game, gp.Level)
				} else if s.selectedOption == 1 {
					// Start Over - go to title
					gp.Level = 1
					gp.Lives = 3
					gp.Score = 0
					gp.Points = 0
					gp.TimerTriggered = false
					gp.GameOver = false
					gp.Lives = 3
					s.game.scene = NewTitleScene(s.game)
				}
			} else if gp.GameOver {
				// Game Over - only option is to go to title
				// TODO: In the future, this will link to high score scene
				gp.Level = 1
				gp.Lives = 3
				gp.Score = 0
				gp.Points = 0
				s.game.scene = NewTitleScene(s.game)
			}
		}
	}

	// Timeout, auto restart the game after 10 seconds
	if s.framecounter > 600 {
		gp := s.game.gameplay
		// Start Over - go to title
		gp.Level = 1
		gp.Lives = 3
		gp.Score = 0
		gp.Points = 0
		gp.TimerTriggered = false
		gp.GameOver = false
		gp.Lives = 3
		s.game.scene = NewTitleScene(s.game)
	}
	return nil
}

func (s *GameOverScene) Draw(screen *ebiten.Image) {
	// Draw black overlay
	overlay := ebiten.NewImage(sW, sH)
	overlay.Fill(color.RGBA{0, 0, 0, 255})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	gp := s.game.gameplay

	if gp.GameOver {
		// Game Over screen
		titleOpt := &text.DrawOptions{}
		titleOpt.GeoM.Translate(float64(sW)/2-50, float64(sH)/2-40)
		titleOpt.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255})
		text.Draw(screen, "GAME OVER", gameOverTextFace, titleOpt)

		msgOpt := &text.DrawOptions{}
		msgOpt.GeoM.Translate(float64(sW)/2-80, float64(sH)/2)
		msgOpt.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "PLAY AGAIN?", gameOverTextFace, msgOpt)

		hintOpt := &text.DrawOptions{}
		hintOpt.GeoM.Translate(float64(sW)/2-40, float64(sH)/2+40)
		hintOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
		text.Draw(screen, "Press B", gameOverTextFace, hintOpt)
	} else if s.isTimeUp {
		// Time's Up screen
		titleOpt := &text.DrawOptions{}
		titleOpt.GeoM.Translate(float64(sW)/2-60, float64(sH)/2-60)
		titleOpt.ColorScale.ScaleWithColor(color.RGBA{255, 200, 100, 255})
		text.Draw(screen, "TIME'S UP", gameOverTextFace, titleOpt)

		livesOpt := &text.DrawOptions{}
		livesOpt.GeoM.Translate(float64(sW)/2-100, float64(sH)/2-20)
		livesOpt.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, fmt.Sprintf("%d Lives Left", gp.Lives), gameOverTextFace, livesOpt)

		// Menu options
		tryAgainColor := color.RGBA{255, 255, 255, 255}
		startOverColor := color.RGBA{255, 255, 255, 255}

		if s.selectedOption == 0 {
			tryAgainColor = color.RGBA{255, 255, 100, 255}
		} else {
			startOverColor = color.RGBA{255, 255, 100, 255}
		}

		tryOpt := &text.DrawOptions{}
		tryOpt.GeoM.Translate(float64(sW)/2-40, float64(sH)/2+20)
		tryOpt.ColorScale.ScaleWithColor(tryAgainColor)
		text.Draw(screen, "> Try Again", gameOverTextFace, tryOpt)

		startOpt := &text.DrawOptions{}
		startOpt.GeoM.Translate(float64(sW)/2-40, float64(sH)/2+50)
		startOpt.ColorScale.ScaleWithColor(startOverColor)
		text.Draw(screen, "> Start Over", gameOverTextFace, startOpt)

		hintOpt := &text.DrawOptions{}
		hintOpt.GeoM.Translate(float64(sW)/2-60, float64(sH)/2+90)
		hintOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
		text.Draw(screen, "UP/DOWN to select", gameOverTextFace, hintOpt)
	}
}

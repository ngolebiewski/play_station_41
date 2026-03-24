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

// Letters available for initials entry
const availableLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var highScoreTextFace = text.NewGoXFace(bitmapfont.Face)

type HighScoreScene struct {
	game              *Game
	highScoreManager  *HighScoreManager
	scores            []HighScore
	framecounter      int
	currentScore      int
	
	// Initials entry state
	entryMode         bool // true = entering initials, false = showing results
	initials          [3]int // indices into availableLetters
	currentPosition   int // 0, 1, 2 for which letter being edited
	showNESMessage    bool
	nesMessageFrame   int
	confirmed         bool
}

func NewHighScoreScene(game *Game, currentScore int) *HighScoreScene {
	/////////////////////////////////////////////////////
	// Start the music!
	if game.audioManager != nil {
		err := game.audioManager.ChangeSong("scenechange")
		if err != nil {
			log.Printf("Audio Error: %v", err)
		}
	}
	/////////////////////////////////////////////////////

	hsm := NewHighScoreManager()
	scores, _ := hsm.LoadHighScores()

	return &HighScoreScene{
		game:             game,
		highScoreManager: hsm,
		scores:           scores,
		framecounter:     0,
		currentScore:     currentScore,
		entryMode:        true,
		initials:         [3]int{0, 0, 0}, // Start with "AAA"
		currentPosition:  0,
		showNESMessage:   false,
		nesMessageFrame:  0,
		confirmed:        false,
	}
}

func (s *HighScoreScene) Update() error {
	s.framecounter++

	gpad.UpdateTouch()

	if s.entryMode && !s.confirmed {
		// Initials entry mode
		// ─── Up/Down: Change current letter ────────────────────────────
		if gpad.MoveUp() {
			s.initials[s.currentPosition]--
			if s.initials[s.currentPosition] < 0 {
				s.initials[s.currentPosition] = len(availableLetters) - 1
			}
		}
		if gpad.MoveDown() {
			s.initials[s.currentPosition]++
			if s.initials[s.currentPosition] >= len(availableLetters) {
				s.initials[s.currentPosition] = 0
			}
		}

		// ─── Left/Right: Move between positions ────────────────────────
		if gpad.MoveLeft() {
			s.currentPosition--
			if s.currentPosition < 0 {
				s.currentPosition = 2
			}
		}
		if gpad.MoveRight() {
			s.currentPosition++
			if s.currentPosition > 2 {
				s.currentPosition = 0
			}
		}

		// ─── A/Start: Confirm initials ────────────────────────────────
		if gpad.PressA() || gpad.PressStart() {
			initialsStr := string([]rune{
				rune(availableLetters[s.initials[0]]),
				rune(availableLetters[s.initials[1]]),
				rune(availableLetters[s.initials[2]]),
			})

			// Check for NES easter egg
			if initialsStr == "NES" {
				s.showNESMessage = true
				s.nesMessageFrame = 180 // Show for 3 seconds at 60fps
			}

			// Save the high score
			s.highScoreManager.SaveHighScore(initialsStr, s.currentScore)
			s.scores, _ = s.highScoreManager.LoadHighScores()

			s.confirmed = true
			s.entryMode = false
		}
	} else {
		// Show results mode with "Press any button to continue"
		if gpad.PressA() || gpad.PressB() || gpad.PressStart() || gpad.PressSelect() {
			// Return to title scene
			gp := s.game.gameplay
			gp.Level = 1
			gp.Lives = 3
			gp.Score = 0
			gp.Points = 0
			gp.TimerTriggered = false
			gp.GameOver = false
			gp.LevelComplete = false
			s.game.scene = NewTitleScene(s.game)
		}

		// Update NES message frame
		if s.showNESMessage {
			s.nesMessageFrame--
			if s.nesMessageFrame <= 0 {
				s.showNESMessage = false
			}
		}
	}

	return nil
}

func (s *HighScoreScene) Draw(screen *ebiten.Image) {
	// Draw black overlay
	overlay := ebiten.NewImage(sW, sH)
	overlay.Fill(color.RGBA{0, 0, 0, 255})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	if s.entryMode && !s.confirmed {
		s.drawInitialsEntry(screen)
	} else {
		s.drawResults(screen)
	}

	// Draw NES easter egg message if active
	if s.showNESMessage {
		s.drawNESMessage(screen)
	}
}

func (s *HighScoreScene) drawInitialsEntry(screen *ebiten.Image) {
	// Title
	titleOpt := &text.DrawOptions{}
	titleOpt.GeoM.Translate(20, 30)
	titleOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "CONGRATULATIONS!", highScoreTextFace, titleOpt)

	// Subtitle
	subtitleOpt := &text.DrawOptions{}
	subtitleOpt.GeoM.Translate(40, 50)
	subtitleOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "YOU HAVE GRADUATED!", highScoreTextFace, subtitleOpt)

	// Score display
	scoreOpt := &text.DrawOptions{}
	scoreOpt.GeoM.Translate(60, 80)
	scoreOpt.ColorScale.ScaleWithColor(color.RGBA{255, 150, 100, 255})
	text.Draw(screen, fmt.Sprintf("SCORE: %d", s.currentScore), highScoreTextFace, scoreOpt)

	// "Enter Initials" prompt
	promptOpt := &text.DrawOptions{}
	promptOpt.GeoM.Translate(45, 120)
	promptOpt.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Enter Initials:", highScoreTextFace, promptOpt)

	// Initials with current position highlighted
	initialsX := 100.0
	initialsY := 140.0
	for i := 0; i < 3; i++ {
		ch := string(availableLetters[s.initials[i]])
		optColor := color.RGBA{200, 200, 200, 255}

		if i == s.currentPosition {
			optColor = color.RGBA{255, 255, 100, 255}
		}

		opt := &text.DrawOptions{}
		opt.GeoM.Translate(float64(initialsX+float64(i)*30), initialsY)
		opt.ColorScale.ScaleWithColor(optColor)
		text.Draw(screen, ch, highScoreTextFace, opt)
	}

	// Instructions
	instrOpt := &text.DrawOptions{}
	instrOpt.GeoM.Translate(30, 180)
	instrOpt.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
	text.Draw(screen, "UP/DOWN: Change  LEFT/RIGHT: Move", highScoreTextFace, instrOpt)

	instrOpt2 := &text.DrawOptions{}
	instrOpt2.GeoM.Translate(60, 195)
	instrOpt2.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
	text.Draw(screen, "A/START: Confirm", highScoreTextFace, instrOpt2)
}

func (s *HighScoreScene) drawResults(screen *ebiten.Image) {
	// Title
	titleOpt := &text.DrawOptions{}
	titleOpt.GeoM.Translate(20, 20)
	titleOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "HIGH SCORES", highScoreTextFace, titleOpt)

	// Top 7 scores table
	scoreStartX := 40.0
	scoreStartY := 50.0
	scoreLineHeight := 20.0

	// Header
	headerOpt := &text.DrawOptions{}
	headerOpt.GeoM.Translate(scoreStartX, scoreStartY)
	headerOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 100, 255})
	text.Draw(screen, "RANK  INITIALS  SCORE", highScoreTextFace, headerOpt)

	// Draw each score
	for i, score := range s.scores {
		yPos := scoreStartY + scoreLineHeight + float64(i)*scoreLineHeight
		optColor := color.RGBA{200, 200, 200, 255}

		// Highlight the newly entered score
		if score.Score == s.currentScore && i < 7 {
			optColor = color.RGBA{255, 255, 100, 255}
		}

		rankOpt := &text.DrawOptions{}
		rankOpt.GeoM.Translate(scoreStartX, yPos)
		rankOpt.ColorScale.ScaleWithColor(optColor)
		text.Draw(screen, fmt.Sprintf("%d.   %s       %5d", i+1, score.Initials, score.Score), highScoreTextFace, rankOpt)
	}

	// Continue prompt
	continueOpt := &text.DrawOptions{}
	continueOpt.GeoM.Translate(40, float64(sH)-40)
	continueOpt.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
	text.Draw(screen, "Press any key to continue", highScoreTextFace, continueOpt)
}

func (s *HighScoreScene) drawNESMessage(screen *ebiten.Image) {
	// Draw a large box with NES easter egg message
	messageBox := ebiten.NewImage(300, 100)
	messageBox.Fill(color.RGBA{40, 40, 40, 200})
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(sW)/2-150, float64(sH)/2-50)
	screen.DrawImage(messageBox, opt)

	// Draw border
	borderColor := color.RGBA{255, 100, 100, 255}
	borderImg := ebiten.NewImage(298, 2)
	borderImg.Fill(borderColor)
	borderOpt := &ebiten.DrawImageOptions{}
	borderOpt.GeoM.Translate(float64(sW)/2-149, float64(sH)/2-49)
	screen.DrawImage(borderImg, borderOpt)

	// Message text
	msgOpt := &text.DrawOptions{}
	msgOpt.GeoM.Translate(float64(sW)/2-120, float64(sH)/2-30)
	msgOpt.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255})
	text.Draw(screen, "NES EMULATOR UNLOCKED!", highScoreTextFace, msgOpt)

	subOpt := &text.DrawOptions{}
	subOpt.GeoM.Translate(float64(sW)/2-100, float64(sH)/2-10)
	subOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "RetroPI coming soon...", highScoreTextFace, subOpt)
}

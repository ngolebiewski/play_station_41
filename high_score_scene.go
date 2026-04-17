package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ngolebiewski/play_station_41/gpad"
)

// Letters available for initials entry
const availableLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var highScoreTextFace = text.NewGoXFace(bitmapfont.Face)

type HighScoreScene struct {
	game             *Game
	highScoreManager *HighScoreManager
	scores           []HighScore
	framecounter     int
	currentScore     int

	// Initials entry state
	entryMode       bool   // true = entering initials, false = showing results
	initials        [3]int // indices into availableLetters
	currentPosition int    // 0, 1, 2 for which letter being edited (or 3 = confirm prompt)
	showNESMessage  bool
	nesMessageFrame int
	confirmed       bool

	// Input debouncing
	lastInputFrame  int // Track last frame input was processed
	inputDebounceMs int // Milliseconds between inputs (10ms at 60fps = ~0.6 frames, use 10 frames for safety)
}

func NewHighScoreScene(game *Game, currentScore int) *HighScoreScene {
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
		lastInputFrame:   -100, // Ensure first input is accepted
		inputDebounceMs:  10,   // 10 frames at 60fps ~= 166ms
	}
}

func (s *HighScoreScene) Update() error {
	s.framecounter++

	gpad.UpdateTouch()

	if s.entryMode && !s.confirmed {
		// Initials entry mode - sequential single-letter entry

		// Check if enough frames have passed for next input (debounce)
		canInput := (s.framecounter - s.lastInputFrame) >= s.inputDebounceMs

		if s.currentPosition < 3 {
			// Editing one of the three letters

			// ─── Up/Down: Scroll through available letters ──────────────
			if canInput && (gpad.MoveUp() || gpad.MoveDown()) {
				if gpad.MoveUp() {
					s.initials[s.currentPosition]--
					if s.initials[s.currentPosition] < 0 {
						s.initials[s.currentPosition] = len(availableLetters) - 1
					}
				} else {
					s.initials[s.currentPosition]++
					if s.initials[s.currentPosition] >= len(availableLetters) {
						s.initials[s.currentPosition] = 0
					}
				}
				s.lastInputFrame = s.framecounter
			}

			// ─── A/Action button: Confirm letter and move to next ─────────────
			if gpad.PressB() || gpad.PressStart() {

				s.currentPosition++ // Move to next letter or confirmation prompt
				s.lastInputFrame = s.framecounter
			}

			// ─── B button: Go back to previous letter (if not first) ────
			if gpad.PressA() && s.currentPosition > 0 {

				s.currentPosition--
				s.lastInputFrame = s.framecounter
			}
		} else if s.currentPosition == 3 {
			// Confirmation prompt ("ENTER")

			// ─── A/Action: Confirm and save ──────────────────────────────────
			if gpad.PressB() || gpad.PressStart() {
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
				s.lastInputFrame = s.framecounter
			}

			// ─── B: Go back to edit third letter ──────────────────────
			if gpad.PressB() {
				s.currentPosition = 2
				s.lastInputFrame = s.framecounter
			}
		}
	} else if s.confirmed {
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
			s.game.scene = NewCreditsScene(s.game)
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
	// Bitmap font is ~6px wide per char. Screen is 240px wide.
	// Center formula: x = (240 - charCount*6) / 2

	// "CONGRATULATIONS!" = 16 chars = 96px → x=72
	titleOpt := &text.DrawOptions{}
	titleOpt.GeoM.Translate(72, 10)
	titleOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "CONGRATULATIONS!", highScoreTextFace, titleOpt)

	// "HIGH SCORE!" = 11 chars = 66px → x=87
	highScoreBannerOpt := &text.DrawOptions{}
	highScoreBannerOpt.GeoM.Translate(87, 22)
	highScoreBannerOpt.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255})
	text.Draw(screen, "HIGH SCORE!", highScoreTextFace, highScoreBannerOpt)

	// "YOU HAVE GRADUATED!" = 19 chars = 114px → x=63
	subtitleOpt := &text.DrawOptions{}
	subtitleOpt.GeoM.Translate(63, 34)
	subtitleOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "YOU HAVE GRADUATED!", highScoreTextFace, subtitleOpt)

	// Score — roughly centered
	scoreOpt := &text.DrawOptions{}
	scoreOpt.GeoM.Translate(70, 46)
	scoreOpt.ColorScale.ScaleWithColor(color.RGBA{255, 150, 100, 255})
	text.Draw(screen, fmt.Sprintf("SCORE: %d", s.currentScore), highScoreTextFace, scoreOpt)

	if s.currentPosition < 3 {
		// "Letter X of 3:" = 14 chars = 84px → x=78
		letterNumOpt := &text.DrawOptions{}
		letterNumOpt.GeoM.Translate(78, 62)
		letterNumOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
		text.Draw(screen, fmt.Sprintf("Letter %d of 3:", s.currentPosition+1), highScoreTextFace, letterNumOpt)

		// Three letters with 18px gap, total ~42px wide → center at x=99
		displayX := 99.0
		displayY := 78.0
		for i := 0; i < 3; i++ {
			ch := string(availableLetters[s.initials[i]])
			optColor := color.RGBA{100, 100, 100, 255} // unvisited: grey
			if i < s.currentPosition {
				optColor = color.RGBA{100, 220, 100, 255} // confirmed: green
			} else if i == s.currentPosition {
				optColor = color.RGBA{255, 255, 100, 255} // focused: yellow
			}

			opt := &text.DrawOptions{}
			opt.GeoM.Translate(displayX+float64(i)*18, displayY)
			opt.ColorScale.ScaleWithColor(optColor)
			text.Draw(screen, ch, highScoreTextFace, opt)
		}

		// Large current letter — single char → x=117
		currentLetterOpt := &text.DrawOptions{}
		currentLetterOpt.GeoM.Translate(117, 96)
		currentLetterOpt.ColorScale.ScaleWithColor(color.RGBA{255, 255, 100, 255})
		ch := string(availableLetters[s.initials[s.currentPosition]])
		text.Draw(screen, ch, highScoreTextFace, currentLetterOpt)

		// "UP/DOWN: Scroll" = 15 chars = 90px → x=75
		instrOpt := &text.DrawOptions{}
		instrOpt.GeoM.Translate(75, 118)
		instrOpt.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
		text.Draw(screen, "UP/DOWN: Scroll", highScoreTextFace, instrOpt)

		// "B: Set Letter" = 13 chars = 78px → x=81
		instrOpt2 := &text.DrawOptions{}
		instrOpt2.GeoM.Translate(81, 130)
		instrOpt2.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
		text.Draw(screen, "B: Set Letter", highScoreTextFace, instrOpt2)

		if s.currentPosition > 0 {
			// "A: Change prev" = 14 chars = 84px → x=78
			instrOpt3 := &text.DrawOptions{}
			instrOpt3.GeoM.Translate(78, 142)
			instrOpt3.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
			text.Draw(screen, "A: Change prev", highScoreTextFace, instrOpt3)
		}
	} else if s.currentPosition == 3 {
		// All three letters green, same layout as above
		displayX := 99.0
		displayY := 78.0
		for i := 0; i < 3; i++ {
			ch := string(availableLetters[s.initials[i]])
			opt := &text.DrawOptions{}
			opt.GeoM.Translate(displayX+float64(i)*18, displayY)
			opt.ColorScale.ScaleWithColor(color.RGBA{100, 220, 100, 255}) // all green
			text.Draw(screen, ch, highScoreTextFace, opt)
		}

		// "ENTER" = 5 chars = 30px → x=105
		enterOpt := &text.DrawOptions{}
		enterOpt.GeoM.Translate(105, 96)
		enterOpt.ColorScale.ScaleWithColor(color.RGBA{255, 255, 100, 255})
		text.Draw(screen, "ENTER", highScoreTextFace, enterOpt)

		// "A: Confirm and save" = 19 chars = 114px → x=63
		instrOpt := &text.DrawOptions{}
		instrOpt.GeoM.Translate(63, 118)
		instrOpt.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
		text.Draw(screen, "A: Confirm and save", highScoreTextFace, instrOpt)

		// "B: Edit last letter" = 19 chars = 114px → x=63
		instrOpt2 := &text.DrawOptions{}
		instrOpt2.GeoM.Translate(63, 130)
		instrOpt2.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
		text.Draw(screen, "B: Edit last letter", highScoreTextFace, instrOpt2)
	}
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
	scoreLineHeight := 10.0

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
	continueOpt.GeoM.Translate(40, float64(sH)-20)
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
	text.Draw(screen, "RetroPie coming soon...", highScoreTextFace, subOpt)
}

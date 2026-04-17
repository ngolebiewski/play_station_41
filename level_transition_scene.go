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

// Create text face for transition screen
var transitionTextFace = text.NewGoXFace(bitmapfont.Face)

type LevelTransitionScene struct {
	game         *Game
	framecounter int
	nextLevel    int
	timeLeft     int
}

func NewLevelTransitionScene(game *Game, timeLeft int) *LevelTransitionScene {
	/////////////////////////////////////////////////////
	// Start the music!
	// This will trigger the fade-in automatically.
	if game.audioManager != nil {
		err := game.audioManager.ChangeSong("scenechange")
		if err != nil {
			log.Printf("Audio Error: %v", err)
		}
	}
	/////////////////////////////////////////////////////

	return &LevelTransitionScene{
		game:         game,
		framecounter: 0,
		nextLevel:    game.gameplay.Level,
		timeLeft:     timeLeft,
	}
}

func (s *LevelTransitionScene) Update() error {
	s.framecounter++

	gpad.UpdateTouch()

	// Advance on any button press or after duration
	if s.framecounter >= levelTransitionDuration || gpad.PressA() || gpad.PressB() || gpad.PressStart() {
		// Transition to the next classroom level
		s.game.scene = NewClassroomScene(s.game, s.nextLevel)
	}

	return nil
}

func (s *LevelTransitionScene) Draw(screen *ebiten.Image) {
	// Draw black overlay
	overlay := ebiten.NewImage(sW, sH)
	overlay.Fill(color.RGBA{0, 0, 0, 255})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	gp := s.game.gameplay

	// "Good job" message
	goodJobOpt := &text.DrawOptions{}
	goodJobOpt.GeoM.Translate(100, 50)
	goodJobOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "Good job!", transitionTextFace, goodJobOpt)

	// "Time Bonus" message
	// totalTimeBonus := s.timeLeft * timeBonusPerSecond
	timeBonusString := fmt.Sprintf("Time Bonus!: %d sec x %d = %d Points", s.timeLeft/60, timeBonusPerSecond, (s.timeLeft/60)*timeBonusPerSecond)
	timeBonusOpt := &text.DrawOptions{}
	timeBonusOpt.GeoM.Translate(20, 65)
	timeBonusOpt.ColorScale.ScaleWithColor(color.RGBA{0, 255, 0, 255})
	text.Draw(screen, timeBonusString, transitionTextFace, timeBonusOpt)

	// Level name - "Level 2: Pre-K" format
	levelName := gp.GetLevelName()
	levelOpt := &text.DrawOptions{}
	levelOpt.GeoM.Translate(60, 80)
	levelOpt.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, fmt.Sprintf("Level %d: %s", s.nextLevel, levelName), transitionTextFace, levelOpt)

	// Lives display - "Lives: 3"
	livesOpt := &text.DrawOptions{}
	livesOpt.GeoM.Translate(90, 110)
	livesOpt.ColorScale.ScaleWithColor(color.RGBA{200, 100, 100, 255})
	text.Draw(screen, fmt.Sprintf("Lives: %d", gp.Lives), transitionTextFace, livesOpt)

	// Progress indicator - can just be a simple timeout bar or skip message
	if s.framecounter < levelTransitionDuration {
		skipOpt := &text.DrawOptions{}
		skipOpt.GeoM.Translate(50, 140)
		skipOpt.ColorScale.ScaleWithColor(color.RGBA{150, 150, 150, 255})
		text.Draw(screen, "Press any button to continue", transitionTextFace, skipOpt)
	}
}

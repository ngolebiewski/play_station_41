package main

import (
	"image/color"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	gotiled "github.com/lafriks/go-tiled"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/tiled"
)

var demoTextFace = text.NewGoXFace(bitmapfont.Face)

type DemoScene struct {
	game     *Game
	timer    int
	phase    int
	tilemap  *gotiled.Map
	renderer *tiled.Renderer
	floatingTexts []*floatingText
}

func NewDemoScene(game *Game) *DemoScene {
	m, err := tiled.LoadMapFS(embeddedAssets, "tiled_files/classroom_1.tmx")
	if err != nil {
		log.Fatal(err)
	}

	tileset := game.assets.ClassroomTileset_1

	renderer := tiled.NewRenderer(m, tileset, scale)

	return &DemoScene{
		game:          game,
		timer:         0,
		phase:         0,
		tilemap:       m,
		renderer:      renderer,
		floatingTexts: make([]*floatingText, 0),
	}
}

func (s *DemoScene) Update() error {
	s.timer++

	// Check for any input to exit demo
	if gpad.PressA() ||gpad.PressB() || gpad.PressStart() || len(ebiten.AppendTouchIDs(nil)) > 0 {
		s.game.scene = NewTitleScene(s.game)
		return nil
	}

	// Phases
	switch s.phase {
	case 0:
		if s.timer == 60 { // 1 second
			s.addFloatingText("Welcome to Play Station 41!", float64(sW)/2, float64(sH)/2-50, color.RGBA{255, 255, 255, 255}, false)
			s.phase++
		}
	case 1:
		if s.timer == 180 { // 3 seconds
			s.addFloatingText("Find objects in the classroom!", float64(sW)/2, float64(sH)/2, color.RGBA{255, 255, 0, 255}, false)
			s.phase++
		}
	case 2:
		if s.timer == 300 { // 5 seconds
			s.addFloatingText("Teachers and students are here!", float64(sW)/2, float64(sH)/2+50, color.RGBA{0, 255, 255, 255}, false)
			s.phase++
		}
	case 3:
		if s.timer == 420 { // 7 seconds
			// Animate student or something, but for simplicity, just text
			s.addFloatingText("Let's start!", float64(sW)/2, float64(sH)/2+100, color.RGBA{255, 0, 255, 255}, false)
			s.phase++
		}
	}

	// Update floating texts
	for i := 0; i < len(s.floatingTexts); i++ {
		ft := s.floatingTexts[i]
		ft.frame++
		if ft.frame >= ft.duration {
			s.floatingTexts = append(s.floatingTexts[:i], s.floatingTexts[i+1:]...)
			i--
		}
	}

	// After 30 seconds (1800 frames), switch back to title
	if s.timer > 1800 {
		s.game.scene = NewTitleScene(s.game)
	}

	return nil
}

func (s *DemoScene) Draw(screen *ebiten.Image) {
	// Draw title background
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(s.game.assets.TitleImage, op)

	// Draw tilemap
	s.renderer.Draw(screen, 0, 0)

	// Draw floating texts
	s.drawFloatingTexts(screen)
}

func (s *DemoScene) addFloatingText(text string, x, y float64, textColor color.RGBA, shake bool) {
	ft := &floatingText{
		text:     text,
		x:        x,
		y:        y,
		color:    textColor,
		frame:    0,
		duration: floatingTextDuration,
		shake:    shake,
	}
	s.floatingTexts = append(s.floatingTexts, ft)
}

func (s *DemoScene) drawFloatingTexts(screen *ebiten.Image) {
	for _, ft := range s.floatingTexts {
		// Calculate alpha for fade out in last 20 frames
		alpha := uint8(255)
		if ft.frame > ft.duration-20 {
			fadeProgress := float64(ft.frame-(ft.duration-20)) / 20.0
			alpha = uint8((1.0 - fadeProgress) * 255)
		}

		// Apply shake if enabled
		x := ft.x
		y := ft.y
		if ft.shake {
			// Random shake offset, decreasing over time
			shakeIntensity := 1.0 - float64(ft.frame)/float64(ft.duration)
			x += (rand.Float64() - 0.5) * 2 * shakeIntensity
			y += (rand.Float64() - 0.5) * 2 * shakeIntensity
		}

		// Gentle upward movement
		y -= float64(ft.frame) * 0.3

		opts := &text.DrawOptions{}
		opts.GeoM.Translate(x, y)
		ft.color.A = alpha
		opts.ColorScale.ScaleWithColor(ft.color)
		text.Draw(screen, ft.text, demoTextFace, opts)
	}
}
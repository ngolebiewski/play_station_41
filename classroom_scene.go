package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/play_station_41/gpad"
)

type ClassroomScene struct {
	game *Game
	img  *ebiten.Image
}

func NewClassroomScene(game *Game) *ClassroomScene {
	return &ClassroomScene{game: game,
		img: game.assets.DefaultPlayer}
}

func (s *ClassroomScene) Update() error {
	p := s.game.player
	if gpad.MoveUp() {
		p.y -= float32(p.speed)
	}
	if gpad.MoveDown() {
		p.y += float32(p.speed)
	}
	if gpad.MoveLeft() {
		p.x -= float32(p.speed)
		p.directionRight = false // Face left
	}
	if gpad.MoveRight() {
		p.x += float32(p.speed)
		p.directionRight = true // Face right
	}

	return nil
}

func (s *ClassroomScene) Draw(screen *ebiten.Image) {

	vector.FillRect(screen, 0, 0, sW, sH, color.Black, false)

	p := s.game.player
	if p.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)

		if !p.directionRight {
			// Flip the 16px sprite horizontally
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(tileSize*scale, 0)
		}

		// Move to player position
		op.GeoM.Translate(float64(p.x), float64(p.y))

		screen.DrawImage(p.image, op)
	}

	ebitenutil.DebugPrint(screen, "Classroom")
}

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/tiled"
)

type ClassroomScene struct {
	game     *Game
	renderer *tiled.Renderer
}

func NewClassroomScene(game *Game) *ClassroomScene {
	// Load the map from embedded assets
	m, err := tiled.LoadMapFS(embeddedAssets, "tiled_files/exports/classroom_1..tmj")
	if err != nil {
		log.Fatal(err)
	}

	return &ClassroomScene{
		game: game,
		// Initialize renderer with the map and the tileset asset
		renderer: tiled.NewRenderer(m, game.assets.ClassroomTileset_1, 1.0),
	}
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
		p.directionRight = false
	}
	if gpad.MoveRight() {
		p.x += float32(p.speed)
		p.directionRight = true
	}

	return nil
}

func (s *ClassroomScene) Draw(screen *ebiten.Image) {
	// 1. Draw the map. We pass (0, 0) for camera offset until we add the Camera struct.
	s.renderer.Draw(screen, 0, 0)

	// 2. Draw the player on top of the tilemap
	p := s.game.player
	if p.image != nil {
		op := &ebiten.DrawImageOptions{}

		// Apply sprite scale
		op.GeoM.Scale(scale, scale)

		if !p.directionRight {
			// Flip horizontal: Scale by -1, then shift back by width
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(tileSize*scale), 0)
		}

		// Final position translate
		op.GeoM.Translate(float64(p.x), float64(p.y))

		screen.DrawImage(p.image, op)
	}

	ebitenutil.DebugPrint(screen, "Classroom Scene Active")
}

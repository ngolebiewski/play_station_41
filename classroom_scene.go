package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/tiled"
)

type ClassroomScene struct {
	game      *Game
	renderer  *tiled.Renderer
	camera    Camera
	mapPixelW float64
	mapPixelH float64
}

func NewClassroomScene(game *Game) *ClassroomScene {
	m, err := tiled.LoadMapFS(embeddedAssets, "tiled_files/classroom_1.tmx")
	if err != nil {
		log.Fatal(err)
	}

	if game.debug {
		log.Printf("Map size: %dx%d tiles, tile size: %dx%d px",
			m.Width, m.Height, m.TileWidth, m.TileHeight)
		for i, ts := range m.Tilesets {
			log.Printf("  Tileset %d: name=%q columns=%d tilecount=%d firstgid=%d image=%q",
				i, ts.Name, ts.Columns, ts.TileCount, ts.FirstGID, ts.Image.Source)
		}
		for i, l := range m.Layers {
			log.Printf("  Layer %d: name=%q visible=%v tiles=%d",
				i, l.Name, l.Visible, len(l.Tiles))
		}
	}

	// Full map size in pixels at render scale
	mapPixelW := float64(m.Width*m.TileWidth) * scale
	mapPixelH := float64(m.Height*m.TileHeight) * scale

	return &ClassroomScene{
		game:      game,
		renderer:  tiled.NewRenderer(m, game.assets.ClassroomTileset_1, scale),
		mapPixelW: mapPixelW,
		mapPixelH: mapPixelH,
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

	s.camera.Update(
		float64(p.x), float64(p.y),
		float64(tileSize*scale), float64(tileSize*scale),
		s.mapPixelW, s.mapPixelH,
	)

	return nil
}

func (s *ClassroomScene) Draw(screen *ebiten.Image) {
	// 1. Draw map with camera offset
	s.renderer.Draw(screen, s.camera.X, s.camera.Y)

	// 2. Draw player at screen position (world pos minus camera offset)
	p := s.game.player
	if p.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		if !p.directionRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(tileSize*scale), 0)
		}
		screenX := float64(p.x) - s.camera.X
		screenY := float64(p.y) - s.camera.Y
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(p.image, op)
	}

	ebitenutil.DebugPrint(screen, "3rd Grade Classroom")
}

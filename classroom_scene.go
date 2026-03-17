package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/tiled"
)

type ClassroomScene struct {
	game          *Game
	renderer      *tiled.Renderer
	camera        Camera
	mapPixelW     float64
	mapPixelH     float64
	collisionGrid *tiled.CollisionGrid
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

	mapPixelW := float64(m.Width*m.TileWidth) * scale
	mapPixelH := float64(m.Height*m.TileHeight) * scale

	return &ClassroomScene{
		game:          game,
		renderer:      tiled.NewRenderer(m, game.assets.ClassroomTileset_1, scale),
		mapPixelW:     mapPixelW,
		mapPixelH:     mapPixelH,
		collisionGrid: tiled.BuildCollisionGrid(m),
	}
}

func (s *ClassroomScene) Update() error {
	p := s.game.player
	cg := s.collisionGrid

	// Player bounding box size (one tile at render scale)
	pw := float64(tileSize * scale)
	ph := float64(tileSize * scale)

	// Try each axis independently so the player slides along walls
	// rather than stopping dead on diagonal collisions.
	gpad.UpdateTouch()

	if gpad.MoveUp() {
		ny := float64(p.y) - float64(p.speed)
		if !collidesWithGrid(cg, float64(p.x), ny, pw, ph) {
			p.y = float32(ny)
		}
	}
	if gpad.MoveDown() {
		ny := float64(p.y) + float64(p.speed)
		if !collidesWithGrid(cg, float64(p.x), ny, pw, ph) {
			p.y = float32(ny)
		}
	}
	if gpad.MoveLeft() {
		nx := float64(p.x) - float64(p.speed)
		if !collidesWithGrid(cg, nx, float64(p.y), pw, ph) {
			p.x = float32(nx)
		}
		p.directionRight = false
	}
	if gpad.MoveRight() {
		nx := float64(p.x) + float64(p.speed)
		if !collidesWithGrid(cg, nx, float64(p.y), pw, ph) {
			p.x = float32(nx)
		}
		p.directionRight = true
	}

	if gpad.PressB() && s.game.debug {
		s.camera.Shake(20, 3.0)
	}

	s.camera.Update(
		float64(p.x), float64(p.y),
		pw, ph,
		s.mapPixelW, s.mapPixelH,
	)

	return nil
}

// collidesWithGrid checks all 4 corners of the player's bounding box
// against the collision grid. Using corners means the player can't clip
// through a tile by moving fast enough to skip over the center check.
func collidesWithGrid(cg *tiled.CollisionGrid, x, y, w, h float64) bool {
	// Inset corners by 1px to allow tight squeezing through 1-tile-wide gaps
	const inset = hitboxInset
	return cg.IsSolid(x+inset, y+inset) ||
		cg.IsSolid(x+w-inset, y+inset) ||
		cg.IsSolid(x+inset, y+h-inset) ||
		cg.IsSolid(x+w-inset, y+h-inset)
}

func (s *ClassroomScene) Draw(screen *ebiten.Image) {
	// 1. Draw map
	s.renderer.Draw(screen, s.camera.DrawX(), s.camera.DrawY())

	// 2. Debug: red overlay on collision tiles
	if s.game.debug {
		s.collisionGrid.DrawDebug(screen, s.camera.DrawX(), s.camera.DrawY(), scale)
	}

	// 3. Draw player
	p := s.game.player
	if p.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		if !p.directionRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(tileSize*scale), 0)
		}
		screenX := float64(p.x) - s.camera.DrawX()
		screenY := float64(p.y) - s.camera.DrawY()
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(p.image, op)
	}

	ebitenutil.DebugPrint(screen, "3rd Grade Classroom")
}

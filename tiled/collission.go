package tiled

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	gotiled "github.com/lafriks/go-tiled"
)

// CollisionGrid is a 2D boolean grid where true means the tile is solid.
// Index with grid.Solid[y][x].
type CollisionGrid struct {
	Width, Height int
	TileW, TileH  int // tile size in pixels (from the map)
	Solid         [][]bool
}

// BuildCollisionGrid scans all tile layers whose name contains "collide" or
// "moveable" (case-insensitive) and marks any non-empty tile as solid.
// "moveable" tiles block the player for now — a future issue will make them pushable.
func BuildCollisionGrid(m *gotiled.Map) *CollisionGrid {
	grid := &CollisionGrid{
		Width:  m.Width,
		Height: m.Height,
		TileW:  m.TileWidth,
		TileH:  m.TileHeight,
		Solid:  make([][]bool, m.Height),
	}
	for y := range m.Height {
		grid.Solid[y] = make([]bool, m.Width)
	}

	for _, layer := range m.Layers {
		name := strings.ToUpper(layer.Name)
		isCollide := strings.Contains(name, "COLLIDE")
		isMoveable := strings.Contains(name, "MOVEABLE")
		if !isCollide && !isMoveable {
			continue
		}
		for i, tile := range layer.Tiles {
			if tile.IsNil() {
				continue
			}
			x := i % m.Width
			y := i / m.Width
			grid.Solid[y][x] = true
		}
	}
	return grid
}

// IsSolid reports whether the world pixel position (wx, wy) is inside a solid tile.
// scale is the rendering scale factor - world coordinates are scaled, but the collision
// grid is based on unscaled tile positions, so we divide by scale to get the correct tile.
// Returns true (solid) if the position is outside the map bounds —
// treats the map edge as a wall.
func (g *CollisionGrid) IsSolid(wx, wy float64, scale float64) bool {
	// Unscale world coordinates to get tile-grid coordinates
	x := int(wx/scale) / g.TileW
	y := int(wy/scale) / g.TileH
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return true // out of bounds = solid
	}
	return g.Solid[y][x]
}

// DrawDebug draws a transparent red overlay over every solid tile.
// Call after renderer.Draw, gated by debug mode.
func (g *CollisionGrid) DrawDebug(screen *ebiten.Image, camX, camY, scale float64) {
	overlay := ebiten.NewImage(g.TileW, g.TileH)
	overlay.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 80})

	tw := float64(g.TileW) * scale
	th := float64(g.TileH) * scale

	for y, row := range g.Solid {
		for x, solid := range row {
			if !solid {
				continue
			}
			worldX := float64(x)*float64(g.TileW)*scale - camX
			worldY := float64(y)*float64(g.TileH)*scale - camY
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(tw/float64(g.TileW), th/float64(g.TileH))
			op.GeoM.Translate(worldX, worldY)
			screen.DrawImage(overlay, op)
		}
	}
}

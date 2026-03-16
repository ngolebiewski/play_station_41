package tiled

import (
	"strings"

	gotiled "github.com/lafriks/go-tiled"
)

// CollisionGrid is a 2D boolean grid where true means the tile is solid.
// Index with grid.Solid[y][x].
type CollisionGrid struct {
	Width, Height int
	Solid         [][]bool
}

// BuildCollisionGrid scans all tile layers whose name contains "collide"
// (case-insensitive) and marks any non-empty tile as solid.
//
// go-tiled decodes layer tile data into layer.Tiles ([]*LayerTile).
// tile.IsNil() is true for empty cells (GID 0). Any non-nil tile in a
// collision layer means that cell is solid.
func BuildCollisionGrid(m *gotiled.Map) *CollisionGrid {
	grid := &CollisionGrid{
		Width:  m.Width,
		Height: m.Height,
		Solid:  make([][]bool, m.Height),
	}
	for y := range m.Height {
		grid.Solid[y] = make([]bool, m.Width)
	}

	for _, layer := range m.Layers {
		if !strings.Contains(strings.ToUpper(layer.Name), "COLLIDE") {
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

// IsSolid reports whether the tile at world pixel position (wx, wy) is solid.
// tileW and tileH are the map's TileWidth and TileHeight in pixels.
// Returns false (not solid) if the position is outside the map bounds.
func (g *CollisionGrid) IsSolid(wx, wy float64, tileW, tileH int) bool {
	x := int(wx) / tileW
	y := int(wy) / tileH
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return false
	}
	return g.Solid[y][x]
}

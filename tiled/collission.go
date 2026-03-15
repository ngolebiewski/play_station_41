package tiled

import "strings"

type CollisionGrid struct {
	Width, Height int
	Solid         [][]bool
}

func BuildCollisionGrid(m *Map) *CollisionGrid {
	grid := &CollisionGrid{
		Width:  m.Width,
		Height: m.Height,
		Solid:  make([][]bool, m.Height),
	}

	for y := range m.Height {
		grid.Solid[y] = make([]bool, m.Width)
	}

	for _, layer := range m.Layers {
		// Only check layers with "COLLIDE" in the name [cite: 4]
		if layer.Type == "tilelayer" && strings.Contains(strings.ToUpper(layer.Name), "COLLIDE") {
			for i, gid := range layer.Data {
				if gid != 0 {
					x := i % m.Width
					y := i / m.Width
					grid.Solid[y][x] = true
				}
			}
		}
	}
	return grid
}

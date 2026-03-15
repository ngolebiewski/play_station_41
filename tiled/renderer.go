package tiled

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Tiled uses the high 3 bits for flipping
	flipH    = 0x80000000
	flipV    = 0x40000000
	flipD    = 0x20000000
	flipMask = flipH | flipV | flipD
)

type Renderer struct {
	Map     *Map
	Tileset *ebiten.Image
	Scale   float64
}

// NewRenderer initializes a new renderer with a scale factor for your sprites
func NewRenderer(m *Map, tileset *ebiten.Image, scale float64) *Renderer {
	b := tileset.Bounds()
	fmt.Printf("Tileset loaded: %dx%d\n", b.Dx(), b.Dy())
	return &Renderer{
		Map:     m,
		Tileset: tileset,
		Scale:   scale,
	}
}

// Draw renders the map layers relative to a camera position (camX, camY)
func (r *Renderer) Draw(screen *ebiten.Image, camX, camY float64) {
	tilesPerRow := r.Tileset.Bounds().Dx() / r.Map.TileWidth

	for _, layer := range r.Map.Layers {
		// Only draw visible tile layers
		if !layer.Visible || layer.Type != "tilelayer" {
			continue
		}

		for i, raw := range layer.Data {
			if raw == 0 {
				continue
			}

			// 1. Get the actual GID by clearing the flip bits
			gid := raw &^ flipMask
			tileIndex := int(gid - 1)

			// 2. Calculate source coordinates in the tileset
			sx := (tileIndex % tilesPerRow) * r.Map.TileWidth
			sy := (tileIndex / tilesPerRow) * r.Map.TileHeight

			tile := r.Tileset.SubImage(image.Rect(sx, sy, sx+r.Map.TileWidth, sy+r.Map.TileHeight)).(*ebiten.Image)

			// 3. Calculate screen position
			worldX := float64((i % r.Map.Width) * r.Map.TileWidth)
			worldY := float64((i / r.Map.Width) * r.Map.TileHeight)

			op := &ebiten.DrawImageOptions{}

			// 4. Apply Scale
			op.GeoM.Scale(r.Scale, r.Scale)

			// 5. Handle Tiled Flip Flags
			// Note: Flipping requires translating based on the scaled tile size
			sw := float64(r.Map.TileWidth)
			sh := float64(r.Map.TileHeight)

			if raw&flipH != 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(sw, 0)
			}
			if raw&flipV != 0 {
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(0, sh)
			}
			if raw&flipD != 0 {
				// Diagonal flip is a transpose (swap x/y)
				// In Ebitengine, this is usually handled via GeoM matrix math
				// This simple version assumes your tiles are square
				op.GeoM.Rotate(-90)
			}

			// 6. Final translation: apply camera and scale
			// We subtract camX/camY to move the world, not the screen
			op.GeoM.Translate((worldX-camX)*r.Scale, (worldY-camY)*r.Scale)

			screen.DrawImage(tile, op)
		}
	}
}

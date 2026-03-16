package tiled

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	gotiled "github.com/lafriks/go-tiled"
)

type Renderer struct {
	Map     *gotiled.Map
	Tileset *ebiten.Image
	Scale   float64
}

func NewRenderer(m *gotiled.Map, tileset *ebiten.Image, scale float64) *Renderer {
	b := tileset.Bounds()
	fmt.Printf("Tileset loaded: %dx%d\n", b.Dx(), b.Dy())
	return &Renderer{Map: m, Tileset: tileset, Scale: scale}
}

// Draw renders all visible tile layers to screen.
// camX, camY are the top-left camera position in world pixels.
func (r *Renderer) Draw(screen *ebiten.Image, camX, camY float64) {
	tw := float64(r.Map.TileWidth)
	th := float64(r.Map.TileHeight)
	for _, layer := range r.Map.Layers {
		if !layer.Visible {
			continue
		}
		r.drawLayer(screen, layer, tw, th, camX, camY)
	}
}

func (r *Renderer) drawLayer(screen *ebiten.Image, layer *gotiled.Layer, tw, th, camX, camY float64) {
	for i, tile := range layer.Tiles {
		if tile.IsNil() {
			continue
		}

		srcRect := tile.GetTileRect()
		src := r.Tileset.SubImage(srcRect).(*ebiten.Image)

		col := i % r.Map.Width
		row := i / r.Map.Width
		worldX := float64(col) * tw
		worldY := float64(row) * th

		op := &ebiten.DrawImageOptions{}

		h := tile.HorizontalFlip
		v := tile.VerticalFlip
		d := tile.DiagonalFlip

		// Tiled flip/rotation encoding, verified against actual tile data:
		//   H  V  D   raw example
		//   0  0  0   → normal
		//   1  0  0   → flip horizontal
		//   0  1  0   → flip vertical
		//   1  1  0   → rotate 180°
		//   1  0  1   → rotate 90° CW   (0xA0000000 series)
		//   0  1  1   → rotate 90° CCW  (0x60000000 series)
		//   0  0  1   → anti-diagonal transpose
		//   1  1  1   → anti-diagonal transpose + flip horizontal
		if d {
			if h && v {
				// Anti-diagonal transpose + flip horizontal
				op.GeoM.Rotate(math.Pi / 2)
				op.GeoM.Translate(th, 0)
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(tw, 0)
			} else if h {
				// Rotate 90° CW
				op.GeoM.Rotate(math.Pi / 2)
				op.GeoM.Translate(th, 0)
			} else if v {
				// Rotate 90° CCW
				op.GeoM.Rotate(-math.Pi / 2)
				op.GeoM.Translate(0, tw)
			} else {
				// Anti-diagonal transpose (pure)
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(tw, 0)
				op.GeoM.Rotate(math.Pi / 2)
				op.GeoM.Translate(th, 0)
			}
		} else {
			if h {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(tw, 0)
			}
			if v {
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(0, th)
			}
		}

		op.GeoM.Scale(r.Scale, r.Scale)
		op.GeoM.Translate((worldX-camX)*r.Scale, (worldY-camY)*r.Scale)

		screen.DrawImage(src, op)
	}
}

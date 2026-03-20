package main

import "math/rand"

// Camera tracks the top-left world position to pass to renderer.Draw.
// It keeps the player centered, clamps to map bounds, and supports screen shake.
type Camera struct {
	X, Y float64

	shakeTicks     int     // ticks remaining in shake
	shakeIntensity float64 // max pixel offset per tick
	shakeOffsetX   float64 // current frame's shake offset
	shakeOffsetY   float64
}

// Shake triggers a camera shake effect.
//   - ticks     — duration in game ticks (60 ticks = ~1 second at 60 TPS)
//   - intensity — max pixel offset (e.g. 2.0 for subtle, 5.0 for heavy)
//   - call like: `s.camera.Shake(20, 3.0)  // 20 ticks (~0.33s), 3px intensity“
func (c *Camera) Shake(ticks int, intensity float64) {
	c.shakeTicks = ticks
	c.shakeIntensity = intensity
}

// Update recenters the camera on the player, clamps to map edges, and
// advances the shake timer.
//   - playerX, playerY — player's world position in pixels (unscaled)
//   - playerW, playerH — player sprite size in pixels (scaled)
//   - mapW, mapH       — full map size in scaled pixels
//   - scale            — the render scale factor
func (c *Camera) Update(playerX, playerY, playerW, playerH, mapW, mapH, scale float64) {
	// Center camera on player
	c.X = playerX + playerW/2 - sW/2
	c.Y = playerY + playerH/2 - sH/2

	// Convert scaled map dimensions to unscaled world space for clamping
	mapW_unscaled := mapW / scale
	mapH_unscaled := mapH / scale

	// Clamp to map bounds so we never show black outside the map
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > mapW_unscaled-sW {
		c.X = mapW_unscaled - sW
	}
	if c.Y > mapH_unscaled-sH {
		c.Y = mapH_unscaled - sH
	}

	// Advance shake
	if c.shakeTicks > 0 {
		c.shakeTicks--
		// Random offset in both directions, scaled by intensity
		c.shakeOffsetX = (rand.Float64()*2 - 1) * c.shakeIntensity
		c.shakeOffsetY = (rand.Float64()*2 - 1) * c.shakeIntensity
	} else {
		c.shakeOffsetX = 0
		c.shakeOffsetY = 0
	}
}

// DrawX returns the final camera X to pass to renderer.Draw, including shake.
func (c *Camera) DrawX() float64 { return c.X + c.shakeOffsetX }

// DrawY returns the final camera Y to pass to renderer.Draw, including shake.
func (c *Camera) DrawY() float64 { return c.Y + c.shakeOffsetY }

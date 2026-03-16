package main

// Camera tracks the top-left world position to pass to renderer.Draw.
// It keeps the player centered and clamps to map bounds.
type Camera struct {
	X, Y float64
}

// Update recenters the camera on the player and clamps to map edges.
//   - playerX, playerY — player's world position in pixels
//   - playerW, playerH — player sprite size in pixels (for centering)
//   - mapW, mapH       — full map size in pixels (m.Width * tileSize * scale)
func (c *Camera) Update(playerX, playerY, playerW, playerH, mapW, mapH float64) {
	// Center camera on player
	c.X = playerX + playerW/2 - sW/2
	c.Y = playerY + playerH/2 - sH/2

	// Clamp to map bounds so we never show black outside the map
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > mapW-sW {
		c.X = mapW - sW
	}
	if c.Y > mapH-sH {
		c.Y = mapH - sH
	}
}

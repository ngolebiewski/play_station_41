package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ngolebiewski/play_station_41/gpad"
	"github.com/ngolebiewski/play_station_41/tiled"
)

type ClassroomScene struct {
	game              *Game
	renderer          *tiled.Renderer
	camera            Camera
	mapPixelW         float64
	mapPixelH         float64
	collisionGrid     *tiled.CollisionGrid
	spawns            []tiled.Spawn
	overlay           *ObjectFindOverlay
	foundMessage      *FoundObjectMessage
	levelHasStarted   bool
	playerSpawnX      float64
	playerSpawnY      float64
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

	/////////////////////////////////////////////////////
	// Start the music!
	// This will trigger the fade-in automatically.
	if game.audioManager != nil {
		err := game.audioManager.ChangeSong("classroom")
		if err != nil {
			log.Printf("Audio Error: %v", err)
		}
	}
	/////////////////////////////////////////////////////

	// Get all spawn points from the map
	spawns := tiled.GetSpawns(m)

	// Separate spawns by type
	var targetSpawns []tiled.SpawnPoint
	var otherSpawns []tiled.SpawnPoint
	var playerSpawnX, playerSpawnY float64

	for _, spawn := range spawns {
		sp := tiled.SpawnPoint{X: spawn.X, Y: spawn.Y, Type: spawn.Type}
		switch spawn.Type {
		case "find":
			targetSpawns = append(targetSpawns, sp)
		case "object":
			otherSpawns = append(otherSpawns, sp)
		case "student":
			// Use student spawn for player position
			playerSpawnX = spawn.X
			playerSpawnY = spawn.Y
		}
	}

	// Place objects on the map
	game.gameplay.PlaceObjects(targetSpawns, otherSpawns)

	// Set player spawn position
	if playerSpawnX > 0 || playerSpawnY > 0 {
		game.player.x = float32(playerSpawnX)
		game.player.y = float32(playerSpawnY)
	}

	scene := &ClassroomScene{
		game:            game,
		renderer:        tiled.NewRenderer(m, game.assets.ClassroomTileset_1, scale),
		mapPixelW:       mapPixelW,
		mapPixelH:       mapPixelH,
		collisionGrid:   tiled.BuildCollisionGrid(m),
		spawns:          spawns,
		overlay:         NewObjectFindOverlay(game.gameplay),
		foundMessage:    nil,
		levelHasStarted: false,
		playerSpawnX:    playerSpawnX,
		playerSpawnY:    playerSpawnY,
	}

	return scene
}

func (s *ClassroomScene) Update() error {
	p := s.game.player
	cg := s.collisionGrid
	gp := s.game.gameplay

	// Player bounding box size (one tile at render scale)
	pw := float64(tileSize * scale)
	ph := float64(tileSize * scale)

	// Update overlay
	if s.overlay != nil {
		if s.overlay.Update() {
			s.overlay = nil
			s.levelHasStarted = true
		}
	}

	// If overlay is showing, don't allow player movement yet
	if s.overlay != nil {
		s.camera.Update(
			float64(p.x), float64(p.y),
			pw, ph,
			s.mapPixelW, s.mapPixelH,
		)
		return nil
	}

	// Update gameplay state (timer, messages)
	gp.Update()

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

	// Check for object interaction
	if (gpad.PressB() || gpad.PressA()) && !gp.ObjectFound {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && s.checkPlayerObjectCollision(obj, pw, ph) {
				if obj.IsTarget {
					// Found the target object!
					gp.ObjectFound()
					obj.IsCollected = true
					s.foundMessage = NewFoundObjectMessage()
				}
			}
		}
	}

	// Debug shake
	if gpad.PressB() && s.game.debug {
		s.game.audioManager.PlaySE("zoing")
		s.camera.Shake(20, 3.0)
	}

	// Handle level progression
	if gp.LevelComplete && !gp.GameOver {
		// Reset for next level
		targetSpawns := s.getTargetSpawns()
		otherSpawns := s.getOtherSpawns()
		gp.PlaceObjects(targetSpawns, otherSpawns)
		gp.ObjectFound = false
		gp.LevelComplete = false
		gp.ShowingTargetOverlay = true
		gp.OverlayFrames = 0
		s.overlay = NewObjectFindOverlay(gp)
		s.foundMessage = nil
	}

	// Handle timer timeout (retry same level)
	if gp.TimerTriggered && gp.Lives > 0 {
		gp.TimerTriggered = false
		gp.ObjectFound = false
		// Reset collected flags on objects for retry
		for _, obj := range gp.PlacedObjects {
			obj.IsCollected = false
		}
		// Do NOT reset PlacedObjects - same layout repeats
		s.foundMessage = nil
	}

	// Update found message
	if s.foundMessage != nil {
		if s.foundMessage.Update() {
			s.foundMessage = nil
		}
	}

	s.camera.Update(
		float64(p.x), float64(p.y),
		pw, ph,
		s.mapPixelW, s.mapPixelH,
	)

	return nil
}

// checkPlayerObjectCollision checks if player is touching an object
func (s *ClassroomScene) checkPlayerObjectCollision(obj *ObjectInstance, pw, ph float64) bool {
	p := s.game.player

	playerX := float64(p.x)
	playerY := float64(p.y)

	// Simple AABB collision
	objW := float64(objectDisplaySize * scale)
	objH := float64(objectDisplaySize * scale)

	return playerX < obj.X+objW &&
		playerX+pw > obj.X &&
		playerY < obj.Y+objH &&
		playerY+ph > obj.Y
}

// getTargetSpawns returns all spawn points marked as "find"
func (s *ClassroomScene) getTargetSpawns() []tiled.SpawnPoint {
	var spawns []tiled.SpawnPoint
	for _, spawn := range s.spawns {
		if spawn.Type == "find" {
			spawns = append(spawns, tiled.SpawnPoint{
				X:    spawn.X,
				Y:    spawn.Y,
				Type: spawn.Type,
			})
		}
	}
	return spawns
}

// getOtherSpawns returns all spawn points marked as "object"
func (s *ClassroomScene) getOtherSpawns() []tiled.SpawnPoint {
	var spawns []tiled.SpawnPoint
	for _, spawn := range s.spawns {
		if spawn.Type == "object" {
			spawns = append(spawns, tiled.SpawnPoint{
				X:    spawn.X,
				Y:    spawn.Y,
				Type: spawn.Type,
			})
		}
	}
	return spawns
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

	// 3. Draw placed objects
	gp := s.game.gameplay
	for _, obj := range gp.PlacedObjects {
		if obj.Image != nil && !obj.IsCollected {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			screenX := obj.X - s.camera.DrawX()
			screenY := obj.Y - s.camera.DrawY()
			op.GeoM.Translate(screenX, screenY)
			screen.DrawImage(obj.Image, op)

			// Debug: highlight target object
			if s.game.debug && obj.IsTarget {
				// Draw a red outline
				w := float32(objectDisplaySize * scale)
				h := float32(objectDisplaySize * scale)
				sx := float32(screenX)
				sy := float32(screenY)
				// Manually draw outline by drawing 4 lines
				ebitenutil.DrawRect(screen, float64(sx)-2, float64(sy)-2, float64(w)+4, float64(h)+4, color.RGBA{255, 0, 0, 100})
			}
		}
	}

	// 4. Draw player
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

	// 5. Draw HUD (timer, lives, level)
	s.drawHUD(screen)

	// 6. Draw overlay if active
	if s.overlay != nil {
		s.overlay.Draw(screen)
	}

	// 7. Draw found message if active
	if s.foundMessage != nil {
		s.foundMessage.Draw(screen)
	}

	// Debug info
	ebitenutil.DebugPrint(screen, fmt.Sprintf("3rd Grade Classroom\nLevel: %d | Lives: %d", gp.Level, gp.Lives))
}

// drawHUD draws the heads-up display with timer, lives, and level
func (s *ClassroomScene) drawHUD(screen *ebiten.Image) {
	gp := s.game.gameplay

	// Draw semi-transparent background for HUD
	hudBg := ebiten.NewImage(sW, 20)
	hudBg.Fill(color.RGBA{0, 0, 0, 128})
	screen.DrawImage(hudBg, &ebiten.DrawImageOptions{})

	// Format timer display
	minutes := gp.RemainingTime / 3600
	seconds := (gp.RemainingTime % 3600) / 60
	timerText := fmt.Sprintf("%d:%02d", minutes, seconds)

	// Draw timer
	timerOpt := &text.DrawOptions{}
	timerOpt.GeoM.Translate(10, 5)
	text.Draw(screen, fmt.Sprintf("Time: %s", timerText), &text.GoTextFace{Face: bitmapfont.Face}, timerOpt)

	// Draw lives
	livesOpt := &text.DrawOptions{}
	livesOpt.GeoM.Translate(100, 5)
	text.Draw(screen, fmt.Sprintf("Lives: %d", gp.Lives), &text.GoTextFace{Face: bitmapfont.Face}, livesOpt)

	// Draw level
	levelOpt := &text.DrawOptions{}
	levelOpt.GeoM.Translate(180, 5)
	text.Draw(screen, fmt.Sprintf("Level: %d", gp.Level), &text.GoTextFace{Face: bitmapfont.Face}, levelOpt)

	// Draw target object indicator (greyed out version)
	if gp.TargetObjectImage != nil && !gp.ObjectFound {
		// Draw a small greyed out version
		const targetIndicatorScale = 1.0
		targetOp := &ebiten.DrawImageOptions{}
		targetOp.GeoM.Scale(targetIndicatorScale, targetIndicatorScale)
		targetOp.GeoM.Translate(220, 2)
		// Apply greyscale by reducing the color
		targetOp.ColorScale.SetR(0.5)
		targetOp.ColorScale.SetG(0.5)
		targetOp.ColorScale.SetB(0.5)
		targetOp.ColorScale.SetA(0.7)
		screen.DrawImage(gp.TargetObjectImage, targetOp)
	}
}

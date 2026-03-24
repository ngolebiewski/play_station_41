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

// How to switch tilemaps on each level:
//
// Each level can have its own tilemap and tileset by modifying:
//
// 1. getTilemapPath(level) function - maps level numbers to .tmx files
//    - Add a new case for your level and return the path
//    - Default (case default) is "tiled_files/classroom_1.tmx"
//    - Example: case 2: return "tiled_files/classroom_2.tmx"
//
// 2. getTileset(game, level) function - maps level numbers to image assets
//    - Add a new case for your level and return the tileset image
//    - Default (case default) is game.assets.ClassroomTileset_1
//    - Example: case 2: return game.assets.ClassroomTileset_2
//
// 3. Embed new tilemap files in embed.go
//    - Add //go:embed directives for new .tmx and .tsx files needed
//    - Also embed the referenced tileset PNG files if using new artwork
//
// 4. Load new image assets in embed.go LoadAssets()
//    - Add loading code for new tileset images
//    - Store in Assets struct if persistent, or load on demand

// Create text face for classroom HUD
var hudTextFace = text.NewGoXFace(bitmapfont.Face)

type ClassroomScene struct {
	game            *Game
	renderer        *tiled.Renderer
	camera          Camera
	mapPixelW       float64
	mapPixelH       float64
	collisionGrid   *tiled.CollisionGrid
	spawns          []tiled.Spawn
	overlay         *ObjectFindOverlay
	foundMessage    *FoundObjectMessage
	levelHasStarted bool
	playerSpawnX    float64
	playerSpawnY    float64
}

func NewClassroomScene(game *Game, level int) *ClassroomScene {
	tilemapPath := getTilemapPath(level)
	m, err := tiled.LoadMapFS(embeddedAssets, tilemapPath)
	if err != nil {
		log.Fatal(err)
	}

	tileset := getTileset(game, level)

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

	// Reset level-specific state for new level
	game.gameplay.HasFoundObject = false
	game.gameplay.LevelComplete = false
	game.gameplay.ShowingTargetOverlay = true
	game.gameplay.OverlayFrames = 0

	// Set player spawn position
	if playerSpawnX > 0 || playerSpawnY > 0 {
		game.player.x = float32(playerSpawnX)
		game.player.y = float32(playerSpawnY)
	}

	scene := &ClassroomScene{
		game:            game,
		renderer:        tiled.NewRenderer(m, tileset, scale),
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
	if (gpad.PressB() || gpad.PressA()) && !gp.HasFoundObject {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && s.checkPlayerObjectCollision(obj, pw, ph) {
				if obj.IsTarget {
					// Found the target object!
					s.game.audioManager.PlaySE("pickup")
					gp.ObjectFound()
					obj.IsCollected = true
					obj.CollectedFrame = 0
					obj.PickupProgress = 0.0
					s.foundMessage = NewFoundObjectMessage()
				}
			}
		}
	}

	// Update collected object animations (move toward player)
	for _, obj := range gp.PlacedObjects {
		if obj.IsCollected {
			obj.CollectedFrame++
			// Animate over 30 frames (0.5 seconds at 60fps)
			if obj.CollectedFrame < 30 {
				obj.PickupProgress = float64(obj.CollectedFrame) / 30.0
				// Move toward player center
				playerCenterX := float64(p.x) + pw/2
				playerCenterY := float64(p.y) + ph/2
				obj.X = obj.OrigX + (playerCenterX-obj.OrigX)*obj.PickupProgress
				obj.Y = obj.OrigY + (playerCenterY-obj.OrigY)*obj.PickupProgress
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
		// Transition to the level transition scene
		s.game.scene = NewLevelTransitionScene(s.game)
		return nil
	}

	// Handle timer timeout (retry same level)
	if gp.TimerTriggered && gp.Lives > 0 {
		gp.TimerTriggered = false
		gp.HasFoundObject = false
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
	if s.game.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("3rd Grade Classroom\nLevel: %d | Lives: %d", gp.Level, gp.Lives))
	}
}

// drawHUD draws the heads-up display with timer, lives, and level
func (s *ClassroomScene) drawHUD(screen *ebiten.Image) {
	gp := s.game.gameplay

	// Draw semi-transparent background for HUD (90% alpha = 230 alpha value)
	hudBg := ebiten.NewImage(sW, 20)
	hudBg.Fill(color.RGBA{0, 0, 0, 240})
	screen.DrawImage(hudBg, &ebiten.DrawImageOptions{})

	// Draw timer: just "T:" + seconds
	secondsRemaining := gp.RemainingTime / 60
	timerOpt := &text.DrawOptions{}
	timerOpt.GeoM.Translate(5, 5)
	text.Draw(screen, fmt.Sprintf("Time:%d", secondsRemaining), hudTextFace, timerOpt)

	// Removing Lives for now. Was not in original Super Mario brothers. on hud. Just shows when you lose a life?
	// // Draw lives: player image x count
	// if s.game.player.image != nil {
	// 	// Draw small player image
	// 	livesOp := &ebiten.DrawImageOptions{}
	// 	// livesOp.GeoM.Scale(0.5, 0.5)
	// 	livesOp.GeoM.Translate(50, 4)
	// 	livesOp.ColorScale.SetR(1.5)
	// 	livesOp.ColorScale.SetG(1.5)
	// 	livesOp.ColorScale.SetB(1.5)
	// 	screen.DrawImage(s.game.player.image, livesOp)

	// 	// Draw "x" count
	// 	livesTextOpt := &text.DrawOptions{}
	// 	livesTextOpt.GeoM.Translate(65, 5)
	// 	text.Draw(screen, fmt.Sprintf("x%d", gp.Lives), hudTextFace, livesTextOpt)
	// }

	// Draw level: "Lvl" + grade name
	levelName := gp.GetLevelName()
	levelOpt := &text.DrawOptions{}
	// levelOpt.GeoM.Translate(110, 5)
	levelOpt.GeoM.Translate(60, 5)
	levelOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, fmt.Sprintf("Lvl %d: %s", gp.Level, levelName), hudTextFace, levelOpt)

	// Draw target object indicator(s)
	// Show 1-3 small grayscale/colored objects depending on level
	if gp.TargetObjectImage != nil {
		const objIndicatorSize = 1
		const objSpace = 14 // Space between each object display

		// Determine how many to show based on level
		// was 3 before
		numToShow := min(1, gp.Lives)

		startX := 190.0

		for i := 0; i < numToShow; i++ {
			objOp := &ebiten.DrawImageOptions{}
			objOp.GeoM.Scale(objIndicatorSize, objIndicatorSize)
			objOp.GeoM.Translate(startX+float64(i*objSpace), 2)

			if gp.HasFoundObject {
				// Use full color when found
				objOp.ColorScale.SetA(1.0)
			} else {
				// True greyscale via luminance matrix
				var cm ebiten.ColorM
				cm.ChangeHSV(0, 0, 0.6) // saturation=0 kills all color, value=0.6 for brightness
				objOp.ColorM = cm
				objOp.ColorScale.SetA(0.7)
			}
			screen.DrawImage(gp.TargetObjectImage, objOp)
		}
		FindOpt := &text.DrawOptions{}
		// levelOpt.GeoM.Translate(110, 5)
		FindOpt.GeoM.Translate(165, 5)
		text.Draw(screen, "Find", hudTextFace, FindOpt)
	}
}

// getTilemapPath returns the path to the tilemap file for the given level.
// Level 1: classroom_1.tmx (default)
// Level 2: classroom_2.tmx
// Level 3+: classroom_1.tmx (default)
func getTilemapPath(level int) string {
	switch level {
	case 2:
		return "tiled_files/classroom_2.tmx"
	default:
		return "tiled_files/classroom_1.tmx"
	}
}

// getTileset returns the appropriate tileset image for the given level.
// Level 1: ClassroomTileset_1 (CLASSROOM.png)
// Level 2: ClassroomTileset_2 (CLASSROOM_MAX.png)
// Level 3+: ClassroomTileset_1 (CLASSROOM.png)
func getTileset(game *Game, level int) *ebiten.Image {
	switch level {
	case 2:
		return game.assets.ClassroomTileset_2
	default:
		return game.assets.ClassroomTileset_1
	}
}

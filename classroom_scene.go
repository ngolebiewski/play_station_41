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
//    - Store in Assets struct if needed, or load on demand

// Create text face for classroom HUD
var hudTextFace = text.NewGoXFace(bitmapfont.Face)

// hudIndicatorX/Y is where the object icon sits in the HUD — tween target.
const (
	hudIndicatorX = 190.0
	hudIndicatorY = 2.0
)

// objectTween animates a collected object flying from its world position
// to the HUD indicator slot.
type objectTween struct {
	// world-space start (adjusted for camera at collection time)
	startScreenX, startScreenY float64
	// progress 0→1 over tweenDuration frames
	frame    int
	duration int
	// the image to draw during the tween
	img *ebiten.Image
	// set true once the tween reaches the HUD
	done bool
}

const tweenDuration = 40 // frames (~0.67s at 60fps)

// easeOutQuad gives a snappy deceleration into the HUD slot.
func easeOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

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

	// Active tween (nil when idle)
	tween *objectTween
	// True once tween finishes — HUD icon draws full-color from this point
	hudLit bool
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
	if game.audioManager != nil {
		err := game.audioManager.ChangeSong("classroom")
		if err != nil {
			log.Printf("Audio Error: %v", err)
		}
	}
	/////////////////////////////////////////////////////

	spawns := tiled.GetSpawns(m)

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
			playerSpawnX = spawn.X
			playerSpawnY = spawn.Y
		}
	}

	game.gameplay.PlaceObjects(targetSpawns, otherSpawns)

	game.gameplay.HasFoundObject = false
	game.gameplay.LevelComplete = false
	game.gameplay.ShowingTargetOverlay = true
	game.gameplay.OverlayFrames = 0

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

	pw := float64(tileSize * scale)
	ph := float64(tileSize * scale)

	// Update overlay
	if s.overlay != nil {
		if s.overlay.Update() {
			s.overlay = nil
			s.levelHasStarted = true
		}
	}

	if s.overlay != nil {
		s.camera.Update(float64(p.x), float64(p.y), pw, ph, s.mapPixelW, s.mapPixelH)
		return nil
	}

	gp.Update()

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

	// ── Proximity-based object collection (no button press needed) ────────────
	if !gp.HasFoundObject {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && obj.IsTarget && s.checkPlayerObjectCollision(obj, pw, ph) {
				s.game.audioManager.PlaySE("pickup")

				// Capture screen-space position for the tween before marking collected
				screenX := obj.X - s.camera.DrawX()
				screenY := obj.Y - s.camera.DrawY()

				gp.ObjectFound()
				obj.IsCollected = true
				obj.CollectedFrame = 0
				obj.PickupProgress = 0.0

				// Kick off the HUD tween
				s.tween = &objectTween{
					startScreenX: screenX,
					startScreenY: screenY,
					frame:        0,
					duration:     tweenDuration,
					img:          obj.Image,
				}
				s.hudLit = false
				s.foundMessage = NewFoundObjectMessage()
				break
			}
		}
	}

	// ── Advance tween ─────────────────────────────────────────────────────────
	if s.tween != nil {
		s.tween.frame++
		if s.tween.frame >= s.tween.duration {
			s.tween.done = true
			s.tween = nil
			s.hudLit = true
		}
	}

	// Debug shake (keep for dev convenience)
	if gpad.PressB() && s.game.debug {
		s.game.audioManager.PlaySE("zoing")
		s.camera.Shake(20, 3.0)
	}

	// Handle level progression
	if gp.LevelComplete && !gp.GameOver {
		s.game.scene = NewLevelTransitionScene(s.game)
		return nil
	}

	// Handle timer timeout (retry same level)
	if gp.TimerTriggered && gp.Lives > 0 {
		gp.TimerTriggered = false
		gp.HasFoundObject = false
		s.tween = nil
		s.hudLit = false
		for _, obj := range gp.PlacedObjects {
			obj.IsCollected = false
		}
		s.foundMessage = nil
	}

	// Update found message
	if s.foundMessage != nil {
		if s.foundMessage.Update() {
			s.foundMessage = nil
		}
	}

	s.camera.Update(float64(p.x), float64(p.y), pw, ph, s.mapPixelW, s.mapPixelH)

	return nil
}

// checkPlayerObjectCollision checks if player is touching an object
func (s *ClassroomScene) checkPlayerObjectCollision(obj *ObjectInstance, pw, ph float64) bool {
	p := s.game.player
	playerX := float64(p.x)
	playerY := float64(p.y)

	objW := float64(objectDisplaySize * scale)
	objH := float64(objectDisplaySize * scale)

	return playerX < obj.X+objW &&
		playerX+pw > obj.X &&
		playerY < obj.Y+objH &&
		playerY+ph > obj.Y
}

func (s *ClassroomScene) getTargetSpawns() []tiled.SpawnPoint {
	var spawns []tiled.SpawnPoint
	for _, spawn := range s.spawns {
		if spawn.Type == "find" {
			spawns = append(spawns, tiled.SpawnPoint{X: spawn.X, Y: spawn.Y, Type: spawn.Type})
		}
	}
	return spawns
}

func (s *ClassroomScene) getOtherSpawns() []tiled.SpawnPoint {
	var spawns []tiled.SpawnPoint
	for _, spawn := range s.spawns {
		if spawn.Type == "object" {
			spawns = append(spawns, tiled.SpawnPoint{X: spawn.X, Y: spawn.Y, Type: spawn.Type})
		}
	}
	return spawns
}

func collidesWithGrid(cg *tiled.CollisionGrid, x, y, w, h float64) bool {
	const inset = hitboxInset
	return cg.IsSolid(x+inset, y+inset) ||
		cg.IsSolid(x+w-inset, y+inset) ||
		cg.IsSolid(x+inset, y+h-inset) ||
		cg.IsSolid(x+w-inset, y+h-inset)
}

func (s *ClassroomScene) Draw(screen *ebiten.Image) {
	// 1. Draw map
	s.renderer.Draw(screen, s.camera.DrawX(), s.camera.DrawY())

	// 2. Debug: collision overlay
	if s.game.debug {
		s.collisionGrid.DrawDebug(screen, s.camera.DrawX(), s.camera.DrawY(), scale)
	}

	// 3. Draw placed objects (skip collected ones — they fly via tween)
	gp := s.game.gameplay
	for _, obj := range gp.PlacedObjects {
		if obj.Image != nil && !obj.IsCollected {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			screenX := obj.X - s.camera.DrawX()
			screenY := obj.Y - s.camera.DrawY()
			op.GeoM.Translate(screenX, screenY)
			screen.DrawImage(obj.Image, op)

			if s.game.debug && obj.IsTarget {
				w := float32(objectDisplaySize * scale)
				h := float32(objectDisplaySize * scale)
				ebitenutil.DrawRect(screen, screenX-2, screenY-2, float64(w)+4, float64(h)+4, color.RGBA{255, 0, 0, 100})
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

	// 5. HUD
	s.drawHUD(screen)

	// 6. Flying tween — drawn on top of HUD so it lands visibly
	if s.tween != nil {
		s.drawTween(screen)
	}

	// 7. Overlay / found message
	if s.overlay != nil {
		s.overlay.Draw(screen)
	}
	if s.foundMessage != nil {
		s.foundMessage.Draw(screen)
	}

	if s.game.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("3rd Grade Classroom\nLevel: %d | Lives: %d", gp.Level, gp.Lives))
	}
}

// drawTween renders the object icon flying from world→HUD.
func (s *ClassroomScene) drawTween(screen *ebiten.Image) {
	t := s.tween
	progress := easeOutQuad(float64(t.frame) / float64(t.duration))

	// Interpolate position
	x := t.startScreenX + (hudIndicatorX-t.startScreenX)*progress
	y := t.startScreenY + (hudIndicatorY-t.startScreenY)*progress

	// Scale: shrink from scale→1 as it flies into the HUD
	spriteScale := scale - (scale-1)*progress

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(spriteScale, spriteScale)
	op.GeoM.Translate(x, y)

	// Fade from greyscale → full color as it approaches the HUD
	if progress < 0.5 {
		var cm ebiten.ColorM
		cm.ChangeHSV(0, progress*2, 0.8) // saturation ramps up
		op.ColorM = cm
	}
	// progress >= 0.5: full color, no ColorM needed

	screen.DrawImage(t.img, op)
}

// drawHUD draws the heads-up display with timer, lives, and level
func (s *ClassroomScene) drawHUD(screen *ebiten.Image) {
	gp := s.game.gameplay

	hudBg := ebiten.NewImage(sW, 20)
	hudBg.Fill(color.RGBA{0, 0, 0, 240})
	screen.DrawImage(hudBg, &ebiten.DrawImageOptions{})

	// Timer
	secondsRemaining := gp.RemainingTime / 60
	timerOpt := &text.DrawOptions{}
	timerOpt.GeoM.Translate(5, 5)
	text.Draw(screen, fmt.Sprintf("Time:%d", secondsRemaining), hudTextFace, timerOpt)

	// Level name
	levelName := gp.GetLevelName()
	levelOpt := &text.DrawOptions{}
	levelOpt.GeoM.Translate(60, 5)
	levelOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, fmt.Sprintf("Lvl %d: %s", gp.Level, levelName), hudTextFace, levelOpt)

	// Target object indicator
	if gp.TargetObjectImage != nil {
		const objIndicatorSize = 1
		const objSpace = 14

		numToShow := min(1, gp.Lives)
		startX := hudIndicatorX

		for i := 0; i < numToShow; i++ {
			objOp := &ebiten.DrawImageOptions{}
			objOp.GeoM.Scale(objIndicatorSize, objIndicatorSize)
			objOp.GeoM.Translate(startX+float64(i*objSpace), hudIndicatorY)

			if s.hudLit {
				// Full color — object was delivered to HUD
				objOp.ColorScale.SetA(1.0)
			} else if s.tween != nil {
				// Tween in progress: slot stays greyscale until icon lands
				var cm ebiten.ColorM
				cm.ChangeHSV(0, 0, 0.6)
				objOp.ColorM = cm
				objOp.ColorScale.SetA(0.7)
			} else {
				// Normal pre-collection greyscale
				var cm ebiten.ColorM
				cm.ChangeHSV(0, 0, 0.6)
				objOp.ColorM = cm
				objOp.ColorScale.SetA(0.7)
			}
			screen.DrawImage(gp.TargetObjectImage, objOp)
		}

		FindOpt := &text.DrawOptions{}
		FindOpt.GeoM.Translate(165, 5)
		text.Draw(screen, "Find", hudTextFace, FindOpt)
	}
}

// getTilemapPath returns the path to the tilemap file for the given level.
func getTilemapPath(level int) string {
	switch level {
	case 2:
		return "tiled_files/classroom_2.tmx"
	default:
		return "tiled_files/classroom_1.tmx"
	}
}

// getTileset returns the appropriate tileset image for the given level.
func getTileset(game *Game, level int) *ebiten.Image {
	switch level {
	case 2:
		return game.assets.ClassroomTileset_2
	default:
		return game.assets.ClassroomTileset_1
	}
}

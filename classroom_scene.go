package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand/v2"

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

// objectTweenOff animates a distractor object flying off-screen downward.
type objectTweenOff struct {
	startScreenX, startScreenY float64
	randomXOffset              float64 // Random horizontal drift
	frame                      int
	duration                   int
	img                        *ebiten.Image
}

const tweenOffDuration = 60 // frames for slower off-screen animation (~1 second at 60fps)

// pointsAnimation shows floating points text during level completion
type pointsAnimation struct {
	text     string
	points   int
	x, y     float64
	frame    int
	duration int
	alpha    float32
}

const pointsAnimationDuration = 120 // ~2 seconds

// floatingText shows temporary text messages at specific screen positions
type floatingText struct {
	text     string
	x, y     float64
	color    color.RGBA
	frame    int
	duration int
	shake    bool // If true, add random shake offset
}

const floatingTextDuration = 60 // ~1 second

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
	// Active off-screen tweens for dismissed distractors
	tweensOff []*objectTweenOff
	// Floating points animation during level completion
	pointsAnim *pointsAnimation
	// Floating text messages for point awards
	floatingTexts []*floatingText
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

	isRetrying := game.gameplay.IsRetryingLevel
	if game.gameplay.IsRetryingLevel {
		// Use stored objects for retry
		game.gameplay.PlacedObjects = make([]*ObjectInstance, len(game.gameplay.StoredPlacedObjects))
		for i, obj := range game.gameplay.StoredPlacedObjects {
			// Deep copy the stored object
			newObj := &ObjectInstance{
				X:              obj.X,
				Y:              obj.Y,
				OrigX:          obj.OrigX,
				OrigY:          obj.OrigY,
				ObjectIndex:    obj.ObjectIndex,
				Image:          obj.Image,
				IsTarget:       obj.IsTarget,
				IsCollected:    false,
				CountedAsFound: false,
				CollectedFrame: 0,
				PickupProgress: 0.0,
			}
			game.gameplay.PlacedObjects[i] = newObj
		}
		game.gameplay.TargetObjectIndex = game.gameplay.StoredTargetObjectIndex
		game.gameplay.TargetObjectImage = game.gameplay.Objects[game.gameplay.TargetObjectIndex]
		game.gameplay.ObjectsToFind = game.gameplay.StoredObjectsToFind
		game.gameplay.ObjectsFound = 0
		game.gameplay.IsRetryingLevel = false // Reset flag
	} else {
		game.gameplay.PlaceObjects(targetSpawns, otherSpawns)
	}

	// Set time limit for this level (in frames)
	game.gameplay.RemainingTime = GetLevelTimeLimit(game.gameplay.Level)

	// Reset level-specific state
	game.gameplay.TimerTriggered = false
	game.gameplay.HasFoundObject = false

	game.gameplay.LevelComplete = false
	game.gameplay.ShowingTargetOverlay = !isRetrying
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
		tweensOff:       make([]*objectTweenOff, 0),
		floatingTexts:   make([]*floatingText, 0),
	}

	return scene
}

func (s *ClassroomScene) Update() error {
	if s.game.debug && gpad.PressP() {
		s.game.scene = NewHighScoreScene(s.game, 1041)
	}

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
	// Use larger hitbox for object collection (more forgiving)
	getHitboxDim := float64(getHitboxSize * scale)

	// Only allow collection if not all objects are found yet
	if gp.ObjectsFound < gp.ObjectsToFind {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && obj.IsTarget && s.checkPlayerObjectCollision(obj, getHitboxDim, getHitboxDim) {
				s.game.audioManager.PlaySE("pickup")

				// Capture screen-space position for the tween before marking collected
				screenX := obj.X - s.camera.DrawX()
				screenY := obj.Y - s.camera.DrawY()

				obj.IsCollected = true
				obj.CollectedFrame = 0
				obj.PickupProgress = 0.0

				// Register this object as found immediately to prevent race conditions
				if !obj.CountedAsFound {
					obj.CountedAsFound = true
					gp.ObjectFound()
					// Show floating text for points earned
					s.addFloatingText("+41", obj.X-5, obj.Y-10, color.RGBA{255, 255, 0, 255}, true)
				}

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

	// ── Action button on target objects (expanded search area) ───────────────
	// Pressing action button expands search area by 16 pixels to help with tilemap placement
	if gpad.PressA() {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && obj.IsTarget && s.checkPlayerObjectCollisionWithRange(obj, getHitboxDim, getHitboxDim, 16) {
				s.game.audioManager.PlaySE("pickup")

				// Capture screen-space position for the tween before marking collected
				screenX := obj.X - s.camera.DrawX()
				screenY := obj.Y - s.camera.DrawY()

				obj.IsCollected = true
				obj.CollectedFrame = 0
				obj.PickupProgress = 0.0

				// Register this object as found immediately to prevent race conditions
				if !obj.CountedAsFound {
					obj.CountedAsFound = true
					gp.ObjectFound()
					// Show floating text for points earned
					s.addFloatingText("+41", obj.X, obj.Y-10, color.RGBA{255, 255, 0, 255}, true)
				}

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

	// ── Action button on distractor objects (dismiss them) ───────────────────
	if gpad.PressB() {
		for _, obj := range gp.PlacedObjects {
			if !obj.IsCollected && !obj.IsTarget && s.checkPlayerObjectCollision(obj, getHitboxDim, getHitboxDim) {
				s.game.audioManager.PlaySE("blip")
				if s.game.gameplay.Level == 7 || s.game.gameplay.Level == 8 {
					// s.camera.Shake(5, 0.1)
				} else {
					s.camera.Shake(15, 2.0)
				} // too aggressive makes me motion sick on the busy levels
				gp.Points++ // Award 1 point

				// Show floating text for points earned
				s.addFloatingText("+1", obj.X, obj.Y-10, color.RGBA{255, 255, 255, 255}, true)

				// Capture screen-space position for off-screen tween
				screenX := obj.X - s.camera.DrawX()
				screenY := obj.Y - s.camera.DrawY()

				// Randomize horizontal drift (-50 to 50 pixels)
				randomX := float64(rand.IntN(101)-50) * 0.5

				// Kick off tween to move object off-screen
				s.tweensOff = append(s.tweensOff, &objectTweenOff{
					startScreenX:  screenX,
					startScreenY:  screenY,
					randomXOffset: randomX,
					frame:         0,
					duration:      tweenOffDuration,
					img:           obj.Image,
				})

				// Mark as collected so it won't be drawn or checked again
				obj.IsCollected = true
				break
			}
		}
	}

	// ── Advance tweens off-screen ───────────────────────────────────────────
	for i := 0; i < len(s.tweensOff); i++ {
		toff := s.tweensOff[i]
		toff.frame++
		if toff.frame >= toff.duration {
			// Remove this tween from the list
			s.tweensOff = append(s.tweensOff[:i], s.tweensOff[i+1:]...)
			i--
		}
	}

	// ── Advance tween ─────────────────────────────────────────────────────────
	if s.tween != nil {
		s.tween.frame++
		if s.tween.frame >= s.tween.duration {
			s.tween.done = true
			s.tween = nil
			s.hudLit = true

			// If level complete, create points animation
			if gp.LevelComplete {
				secondsRemaining := gp.RemainingTime / 60
				timeBonus := secondsRemaining * 5
				timeBonusText := fmt.Sprintf("Time bonus %d sec x 5 points!", secondsRemaining)
				s.pointsAnim = &pointsAnimation{
					text:     timeBonusText,
					points:   timeBonus,
					x:        float64(sW) / 2,
					y:        float64(sH) / 2,
					frame:    0,
					duration: pointsAnimationDuration,
					alpha:    1.0,
				}
			}
		}
	}

	// ── Advance points animation ───────────────────────────────────────────
	if s.pointsAnim != nil {
		s.pointsAnim.frame++
		// Fade out after 60 frames
		if s.pointsAnim.frame > 60 {
			s.pointsAnim.alpha = float32(1.0 - float64(s.pointsAnim.frame-60)/float64(s.pointsAnim.duration-60))
		}
		if s.pointsAnim.frame >= s.pointsAnim.duration {
			s.pointsAnim = nil
		}
	}

	// ── Update floating texts ─────────────────────────────────────────────
	for i := 0; i < len(s.floatingTexts); i++ {
		ft := s.floatingTexts[i]
		ft.frame++
		if ft.frame >= ft.duration {
			// Remove this floating text
			s.floatingTexts = append(s.floatingTexts[:i], s.floatingTexts[i+1:]...)
			i--
		}
	}

	// Debug shake (keep for dev convenience)
	if gpad.PressB() && s.game.debug {
		s.game.audioManager.PlaySE("zoing")
		s.camera.Shake(20, 3.0)
	}

	// Handle level progression
	if gp.LevelComplete && !gp.GameOver {
		// Check if player just completed Level 8 (5th Grade) - graduation!
		if gp.Level == 9 {
			// Show high score scene after graduation
			s.game.scene = NewHighScoreScene(s.game, gp.Score)
		} else {
			// Otherwise show level transition scene
			timeLeft := int(gp.RemainingTime / 60)
			s.game.scene = NewLevelTransitionScene(s.game, timeLeft)
		}
		return nil
	}

	// Handle game over
	if gp.GameOver {
		s.game.scene = NewGameOverScene(s.game, false)
		return nil
	}

	// Handle timer timeout - show game over menu
	if gp.TimerTriggered && !gp.GameOver {
		s.game.scene = NewGameOverScene(s.game, true)
		return nil
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

// checkPlayerObjectCollisionWithRange checks collision with expanded range (in pixels)
func (s *ClassroomScene) checkPlayerObjectCollisionWithRange(obj *ObjectInstance, pw, ph float64, expandedRange float64) bool {
	p := s.game.player
	playerX := float64(p.x)
	playerY := float64(p.y)

	objW := float64(objectDisplaySize * scale)
	objH := float64(objectDisplaySize * scale)

	// Expand collision box by the given range
	return playerX < obj.X+objW+expandedRange &&
		playerX+pw > obj.X-expandedRange &&
		playerY < obj.Y+objH+expandedRange &&
		playerY+ph > obj.Y-expandedRange
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

// addFloatingText adds a temporary text message at the specified world position
func (s *ClassroomScene) addFloatingText(text string, worldX, worldY float64, textColor color.RGBA, shake bool) {
	// Convert world position to screen position
	screenX := worldX - s.camera.DrawX()
	screenY := worldY - s.camera.DrawY()

	ft := &floatingText{
		text:     text,
		x:        screenX,
		y:        screenY,
		color:    textColor,
		frame:    0,
		duration: floatingTextDuration,
		shake:    shake,
	}
	s.floatingTexts = append(s.floatingTexts, ft)
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

	// 6b. Distractors tweening off-screen
	for _, toff := range s.tweensOff {
		s.drawTweenOff(screen, toff)
	}

	// 6c. Points animation
	if s.pointsAnim != nil {
		s.drawPointsAnimation(screen)
	}

	// 6d. Floating texts
	s.drawFloatingTexts(screen)

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

// drawTweenOff renders a distractor object flying off-screen downward with random horizontal drift.
func (s *ClassroomScene) drawTweenOff(screen *ebiten.Image, toff *objectTweenOff) {
	progress := easeOutQuad(float64(toff.frame) / float64(toff.duration))

	// Move downward off-screen with randomized horizontal drift
	x := toff.startScreenX + toff.randomXOffset*progress
	y := toff.startScreenY + progress*float64(sH+100)

	// Scale: shrink as it falls
	spriteScale := scale * (1.0 - progress*0.3)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(spriteScale, spriteScale)
	op.GeoM.Translate(x, y)

	// Fade out as it leaves
	op.ColorScale.SetA(float32(1.0 - progress))

	screen.DrawImage(toff.img, op)
}

// drawPointsAnimation renders floating points text during level completion.
func (s *ClassroomScene) drawPointsAnimation(screen *ebiten.Image) {
	pa := s.pointsAnim
	if pa == nil {
		return
	}

	// Gentle upward movement
	y := pa.y - float64(pa.frame)*0.5

	opts := &text.DrawOptions{}
	opts.GeoM.Translate(pa.x, y)
	opts.ColorScale.ScaleWithColor(color.RGBA{255, 255, 100, uint8(pa.alpha * 255)})
	text.Draw(screen, pa.text, hudTextFace, opts)
}

// drawFloatingTexts renders all active floating text messages
func (s *ClassroomScene) drawFloatingTexts(screen *ebiten.Image) {
	for _, ft := range s.floatingTexts {
		// Calculate alpha for fade out in last 20 frames
		alpha := uint8(255)
		if ft.frame > ft.duration-20 {
			fadeProgress := float64(ft.frame-(ft.duration-20)) / 20.0
			alpha = uint8((1.0 - fadeProgress) * 255)
		}

		// Apply shake if enabled
		x := ft.x
		y := ft.y
		if ft.shake {
			// Random shake offset, decreasing over time
			shakeIntensity := 1.0 - float64(ft.frame)/float64(ft.duration)
			x += (rand.Float64() - 0.5) * .5 * shakeIntensity
			y += (rand.Float64() - 0.5) * .5 * shakeIntensity
		}

		// Gentle upward movement
		y -= float64(ft.frame) * 0.3

		opts := &text.DrawOptions{}
		opts.GeoM.Translate(x, y)
		ft.color.A = alpha
		opts.ColorScale.ScaleWithColor(ft.color)
		text.Draw(screen, ft.text, hudTextFace, opts)
	}
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
	levelOpt.GeoM.Translate(50, 5)
	levelOpt.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	// text.Draw(screen, fmt.Sprintf("Lvl %d: %s", gp.Level, levelName), hudTextFace, levelOpt)
	text.Draw(screen, fmt.Sprintf("%s", levelName), hudTextFace, levelOpt)

	// Target object indicators (one slot per object to find)
	if gp.TargetObjectImage != nil {
		const objIndicatorSize = 1
		const objSpace = 14

		startX := hudIndicatorX

		// Draw slots for all objects that need to be found
		for i := 0; i < gp.ObjectsToFind; i++ {
			objOp := &ebiten.DrawImageOptions{}
			objOp.GeoM.Scale(objIndicatorSize, objIndicatorSize)
			objOp.GeoM.Translate(startX+float64(i*objSpace), hudIndicatorY)

			// Check if this slot has been filled (ObjectsFound > current slot index)
			if i < gp.ObjectsFound {
				// Full color — object was delivered to HUD
				objOp.ColorScale.SetA(1.0)
			} else if s.tween != nil && i == gp.ObjectsFound {
				// Tween in progress on this slot: greyscale until icon lands
				var cm ebiten.ColorM
				cm.ChangeHSV(0, 0, 0.6)
				objOp.ColorM = cm
				objOp.ColorScale.SetA(0.7)
			} else {
				// Empty slot: greyscale
				var cm ebiten.ColorM
				cm.ChangeHSV(0, 0, 0.6)
				objOp.ColorM = cm
				objOp.ColorScale.SetA(0.7)
			}
			screen.DrawImage(gp.TargetObjectImage, objOp)
		}

		// Draw "Found X/Y" text
		FindOpt := &text.DrawOptions{}
		FindOpt.GeoM.Translate(140, 5)
		text.Draw(screen, fmt.Sprintf("Find %d/%d", gp.ObjectsFound, gp.ObjectsToFind), hudTextFace, FindOpt)
	}

	// Points row (second row)
	ptsRowBg := ebiten.NewImage(sW, 15)
	ptsRowBg.Fill(color.RGBA{20, 20, 20, 240})
	ptsOp := &ebiten.DrawImageOptions{}
	ptsOp.GeoM.Translate(0, 20)
	screen.DrawImage(ptsRowBg, ptsOp)

	ptsOpt := &text.DrawOptions{}
	ptsOpt.GeoM.Translate(5, 23)
	text.Draw(screen, fmt.Sprintf("Pts: %d", gp.Points), hudTextFace, ptsOpt)
}

// getTilemapPath returns the path to the tilemap file for the given level.
func getTilemapPath(level int) string {
	switch level {
	case 2:
		return "tiled_files/classroom_2.tmx"
	case 5:
		return "tiled_files/classroom_maze.tmx"
	case 7:
		return "tiled_files/classroom_busy2.tmx"
	case 8:
		return "tiled_files/classroom_final.tmx"
	default:
		return "tiled_files/classroom_1.tmx"
	}
}

// getTileset returns the appropriate tileset image for the given level.
func getTileset(game *Game, level int) *ebiten.Image {
	switch level {
	case 2, 5, 7:
		return game.assets.ClassroomTileset_2
	default:
		return game.assets.ClassroomTileset_1
	}
}

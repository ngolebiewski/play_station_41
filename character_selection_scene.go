package main

import (
	"image"
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ngolebiewski/play_station_41/gpad"
)

const (
	charGridX          = 8 // characters per row
	charTileSize       = 16
	charDisplaySize    = 32 // scaled up for display (16 * 2)
	selectionPadding   = 4
	autoSelectTimeout  = 300 // frames (about 5 seconds at 60fps)
	inactivityTimeout  = 1800 // frames (30 seconds)
)

type CharacterSelectionScene struct {
	game              *Game
	characters        []*ebiten.Image
	selectedIndex     int
	selectionX        int
	selectionY        int
	inputCooldown     int
	autoSelectCounter int
	lastInputFrame    int
}

func NewCharacterSelectionScene(game *Game) *CharacterSelectionScene {
	chars := extractCharacterSprites(game.assets.CharactersTileset)
	
	scene := &CharacterSelectionScene{
		game:           game,
		characters:     chars,
		selectedIndex:  0,
		selectionX:     0,
		selectionY:     0,
		inputCooldown:  0,
		autoSelectCounter: 0,
		lastInputFrame: 0,
	}
	
	return scene
}

// extractCharacterSprites extracts individual characters from the horizontal spritesheet
func extractCharacterSprites(spritesheet *ebiten.Image) []*ebiten.Image {
	var characters []*ebiten.Image
	
	// Get spritesheet dimensions
	bounds := spritesheet.Bounds()
	sheetWidth := bounds.Max.X
	
	// Calculate how many characters fit in the spritesheet
	totalChars := sheetWidth / charTileSize
	if totalChars > 30 {
		totalChars = 30 // cap at 30 for performance
	}
	
	// Extract each character sprite
	for i := 0; i < totalChars; i++ {
		x := i * charTileSize
		rect := image.Rect(x, 0, x+charTileSize, charTileSize)
		subImg := spritesheet.SubImage(rect).(*ebiten.Image)
		characters = append(characters, subImg)
	}
	
	return characters
}

func (s *CharacterSelectionScene) Update() error {
	// Detect and enable touch if needed
	if !gpad.TouchEnabled() && len(ebiten.AppendTouchIDs(nil)) > 0 {
		gpad.EnableTouch()
	}
	gpad.UpdateTouch()
	
	// Track inactivity for auto-select
	if s.inputCooldown > 0 {
		s.inputCooldown--
	}
	
	hasInput := false
	
	// Handle navigation input
	if s.inputCooldown == 0 {
		if gpad.MoveUp() {
			s.selectionY = max(0, s.selectionY-1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveDown() {
			maxY := (len(s.characters) - 1) / charGridX
			s.selectionY = min(maxY, s.selectionY+1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveLeft() {
			s.selectionX = max(0, s.selectionX-1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveRight() {
			maxX := charGridX - 1
			maxY := (len(s.characters) - 1) / charGridX
			if s.selectionY < maxY || s.selectionX < (len(s.characters)%charGridX)-1 {
				s.selectionX = min(maxX, s.selectionX+1)
				s.inputCooldown = 10
				hasInput = true
			}
		}
	}
	
	// Update selected index based on grid position
	s.selectedIndex = s.selectionY*charGridX + s.selectionX
	if s.selectedIndex >= len(s.characters) {
		s.selectedIndex = len(s.characters) - 1
		s.selectionX = s.selectedIndex % charGridX
		s.selectionY = s.selectedIndex / charGridX
	}
	
	// Handle random selection with ? button
	if gpad.PressSelect() {
		s.randomSelect()
		hasInput = true
	}
	
	// Reset inactivity timer on input
	if hasInput {
		s.lastInputFrame = 0
		s.autoSelectCounter = 0
	} else {
		s.lastInputFrame++
		s.autoSelectCounter++
	}
	
	// Auto select if inactivity timeout reached
	if s.autoSelectCounter >= inactivityTimeout {
		s.confirmSelection()
	}
	
	// Handle selection confirmation (Space, Enter, A button, or touch)
	touchTapped := gpad.TouchEnabled() && len(inpututil.AppendJustReleasedTouchIDs(nil)) > 0
	if gpad.PressB() || touchTapped {
		s.confirmSelection()
	}
	
	return nil
}

func (s *CharacterSelectionScene) Draw(screen *ebiten.Image) {
	// Draw background (dark)
	screen.Fill(color.RGBA{30, 30, 40, 255})
	
	// Draw title
	titleFont := text.NewGoXFace(bitmapfont.Face)
	titleOpts := &text.DrawOptions{}
	titleOpts.GeoM.Translate(sW/2-30, 10)
	titleOpts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "SELECT", titleFont, titleOpts)
	
	// Draw character grid
	startX := (sW - charGridX*charDisplaySize) / 2
	startY := 40
	
	for i, charSprite := range s.characters {
		x := i % charGridX
		y := i / charGridX
		
		screenX := float64(startX + x*charDisplaySize)
		screenY := float64(startY + y*charDisplaySize)
		
		// Draw character sprite
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2) // scale 2x (16->32)
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(charSprite, op)
		
		// Draw selection highlight if this is the selected character
		if i == s.selectedIndex {
			outlineSize := float32(charDisplaySize)
			outlineX := float32(screenX)
			outlineY := float32(screenY)
			
			// Draw blue outline
			vector.StrokeRect(screen, outlineX, outlineY, outlineSize, outlineSize, 2, color.RGBA{0, 100, 255, 255}, false)
		}
	}
	
	// Draw "1P" indicator at top-left
	p1Font := text.NewGoXFace(bitmapfont.Face)
	p1Opts := &text.DrawOptions{}
	p1Opts.GeoM.Translate(10, 10)
	p1Opts.ColorScale.ScaleWithColor(color.RGBA{0, 100, 255, 255})
	text.Draw(screen, "1P", p1Font, p1Opts)
	
	// Draw selected character index below grid
	indexFont := text.NewGoXFace(bitmapfont.Face)
	indexOpts := &text.DrawOptions{}
	indexOpts.GeoM.Translate(sW/2-20, sH-20)
	indexOpts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "START TO CONFIRM", indexFont, indexOpts)
	
	// Draw inactivity warning
	if s.autoSelectCounter > inactivityTimeout-60 {
		warningFont := text.NewGoXFace(bitmapfont.Face)
		warningOpts := &text.DrawOptions{}
		warningOpts.GeoM.Translate(sW/2-50, sH-40)
		warningOpts.ColorScale.ScaleWithColor(color.RGBA{255, 100, 0, 255})
		text.Draw(screen, "AUTO-SELECT SOON...", warningFont, warningOpts)
	}
}

func (s *CharacterSelectionScene) randomSelect() {
	s.selectedIndex = int(rand.IntN(len(s.characters)))
	s.selectionX = s.selectedIndex % charGridX
	s.selectionY = s.selectedIndex / charGridX
}

func (s *CharacterSelectionScene) confirmSelection() {
	// Set the selected character in the player
	if s.selectedIndex < len(s.characters) {
		s.game.player.characterIndex = s.selectedIndex
		s.game.player.image = s.characters[s.selectedIndex]
	}
	
	// Transition to classroom scene
	s.game.scene = NewClassroomScene(s.game)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

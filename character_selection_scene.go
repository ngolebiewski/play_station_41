package main

import (
	_ "embed"
	"fmt"
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

//go:embed portrait_bg.kage
var portraitBgKage []byte

const (
	charGridX        = 8
	charTileSize     = 16
	charDisplaySize  = 16 // 1:1 display, no scaling
	selectionPadding = 2

	inactivityTimeout = 600 // frames (10s at 60fps)

	// Layout for 240x160 screen
	gridMarginLeft = 68                                     // space for portrait panel on left
	gridMarginTop  = 18                                     // below title
	gridCellSize   = charDisplaySize + selectionPadding + 1 // 19px per cell

	portraitX     = 2
	portraitY     = 24
	portraitScale = 3
	portraitSize  = charTileSize * portraitScale // 48px

	// Touch START button (bottom strip) — raised to stay inside 160px
	startBtnX = 150
	startBtnY = 143
	startBtnW = 86
	startBtnH = 10
)

type CharacterSelectionScene struct {
	game              *Game
	characters        []*ebiten.Image
	selectedIndex     int
	selectionX        int
	selectionY        int
	inputCooldown     int
	autoSelectCounter int
	portraitShader    *ebiten.Shader
	shaderTime        float32
	// portraitCanvas is a reusable offscreen sized to the portrait area
	portraitCanvas *ebiten.Image
}

func NewCharacterSelectionScene(game *Game) *CharacterSelectionScene {
	chars := extractCharacterSprites(game.assets.CharactersTileset)

	shader, err := ebiten.NewShader(portraitBgKage)
	if err != nil {
		// Non-fatal: fall back to plain background if shader fails
		shader = nil
	}

	canvas := ebiten.NewImage(portraitSize, portraitSize)

	return &CharacterSelectionScene{
		game:           game,
		characters:     chars,
		portraitShader: shader,
		portraitCanvas: canvas,
	}
}

// extractCharacterSprites extracts individual characters from the horizontal spritesheet.
// Starts at index 1 — index 0 is blank (Aseprite convention).
func extractCharacterSprites(spritesheet *ebiten.Image) []*ebiten.Image {
	var characters []*ebiten.Image
	bounds := spritesheet.Bounds()
	totalChars := bounds.Max.X / charTileSize
	for i := 1; i < totalChars; i++ {
		x := i * charTileSize
		rect := image.Rect(x, 0, x+charTileSize, charTileSize)
		subImg := spritesheet.SubImage(rect).(*ebiten.Image)
		characters = append(characters, subImg)
	}
	return characters
}

// totalTiles = characters + 1 "?" random tile at the end
func (s *CharacterSelectionScene) totalTiles() int {
	return len(s.characters) + 1
}

func (s *CharacterSelectionScene) isRandomTile(i int) bool {
	return i == len(s.characters)
}

// tileScreenRect returns the top-left pixel of tile i in screen space
func (s *CharacterSelectionScene) tileScreenPos(i int) (x, y int) {
	col := i % charGridX
	row := i / charGridX
	return gridMarginLeft + col*gridCellSize, gridMarginTop + 4 + row*gridCellSize
}

func (s *CharacterSelectionScene) Update() error {
	s.shaderTime += 1.0 / 60.0
	gpad.UpdateTouch()

	if s.inputCooldown > 0 {
		s.inputCooldown--
	}

	hasInput := false
	total := s.totalTiles()
	maxIdx := total - 1

	// Keyboard / gamepad navigation
	if s.inputCooldown == 0 {
		if gpad.MoveUp() {
			s.game.audioManager.PlaySE("blip")
			s.selectionY = max(0, s.selectionY-1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveDown() {
			s.game.audioManager.PlaySE("blip")
			maxY := maxIdx / charGridX
			s.selectionY = min(maxY, s.selectionY+1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveLeft() {
			s.game.audioManager.PlaySE("blip")
			s.selectionX = max(0, s.selectionX-1)
			s.inputCooldown = 10
			hasInput = true
		} else if gpad.MoveRight() {
			s.game.audioManager.PlaySE("blip")
			maxY := maxIdx / charGridX
			colMax := charGridX - 1
			if s.selectionY < maxY || s.selectionX < (total%charGridX)-1 {
				s.selectionX = min(colMax, s.selectionX+1)
				s.inputCooldown = 10
				hasInput = true
			}
		}
	}

	// Clamp index
	idx := s.selectionY*charGridX + s.selectionX
	if idx > maxIdx {
		idx = maxIdx
		s.selectionX = idx % charGridX
		s.selectionY = idx / charGridX
	}
	s.selectedIndex = idx

	// Keyboard/gamepad confirm
	if gpad.PressB() || gpad.PressStart() {
		hasInput = true
		if s.isRandomTile(s.selectedIndex) {
			s.randomSelect()
		} else {
			s.confirmSelection()
			return nil
		}
	}

	// Touch: detect released taps
	justReleased := inpututil.AppendJustReleasedTouchIDs(nil)
	for _, tid := range justReleased {
		tx, ty := inpututil.TouchPositionInPreviousTick(tid)

		// Check START button tap
		if tx >= startBtnX && tx <= startBtnX+startBtnW &&
			ty >= startBtnY && ty <= startBtnY+startBtnH {
			hasInput = true
			s.confirmSelection()
			return nil
		}

		// Check character tile tap
		for i := 0; i < total; i++ {
			cx, cy := s.tileScreenPos(i)
			if tx >= cx && tx < cx+gridCellSize && ty >= cy && ty < cy+gridCellSize {
				hasInput = true
				s.game.audioManager.PlaySE("blip")
				if s.isRandomTile(i) {
					s.randomSelect()
				} else if i == s.selectedIndex {
					// Second tap on already-highlighted = confirm
					s.confirmSelection()
					return nil
				} else {
					s.selectedIndex = i
					s.selectionX = i % charGridX
					s.selectionY = i / charGridX
				}
				break
			}
		}
	}

	if hasInput {
		s.autoSelectCounter = 0
	} else {
		s.autoSelectCounter++
	}

	if s.autoSelectCounter >= inactivityTimeout {
		s.randomSelect()
		s.confirmSelection()
	}

	return nil
}

func (s *CharacterSelectionScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 32, 255})

	f := text.NewGoXFace(bitmapfont.Face)

	// Title
	titleOpts := &text.DrawOptions{}
	titleOpts.GeoM.Translate(gridMarginLeft, 3)
	titleOpts.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "SELECT PLAYER", f, titleOpts)

	// Portrait panel — shader draws noisy background + sprite composited on top
	if s.portraitShader != nil && s.selectedIndex < len(s.characters) {
		s.portraitCanvas.Clear()
		// Draw sprite onto canvas at 3x scale
		spriteOp := &ebiten.DrawImageOptions{}
		spriteOp.GeoM.Scale(portraitScale, portraitScale)
		s.portraitCanvas.DrawImage(s.characters[s.selectedIndex], spriteOp)

		// DrawRectShader: shader samples canvas as Src0, fills background with noise,
		// and passes through opaque sprite pixels
		shaderOpts := &ebiten.DrawRectShaderOptions{}
		shaderOpts.GeoM.Translate(portraitX, portraitY)
		shaderOpts.Uniforms = map[string]any{
			"Time": s.shaderTime,
		}
		shaderOpts.Images[0] = s.portraitCanvas
		screen.DrawRectShader(portraitSize, portraitSize, s.portraitShader, shaderOpts)
	} else {
		// Fallback: plain dark background + raw sprite draw
		vector.DrawFilledRect(screen,
			float32(portraitX), float32(portraitY),
			float32(portraitSize), float32(portraitSize),
			color.RGBA{10, 10, 20, 255}, false)
		if s.selectedIndex < len(s.characters) {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(portraitScale, portraitScale)
			op.GeoM.Translate(portraitX, portraitY)
			screen.DrawImage(s.characters[s.selectedIndex], op)
		}
	}

	// Portrait border (drawn over shader output)
	vector.StrokeRect(screen,
		float32(portraitX-1), float32(portraitY-1),
		float32(portraitSize+2), float32(portraitSize+2),
		1, color.RGBA{0, 180, 255, 255}, false)

	// 1P label
	p1Opts := &text.DrawOptions{}
	p1Opts.GeoM.Translate(float64(portraitX+portraitSize/2-4), float64(portraitY+portraitSize+2))
	p1Opts.ColorScale.ScaleWithColor(color.RGBA{0, 180, 255, 255})
	text.Draw(screen, "1P", f, p1Opts)

	// Countdown seconds (below 1P)
	secondsLeft := (inactivityTimeout - s.autoSelectCounter) / 60
	if secondsLeft < 0 {
		secondsLeft = 0
	}
	cdOpts := &text.DrawOptions{}
	cdOpts.GeoM.Translate(float64(portraitX), float64(portraitY+portraitSize+14))
	if secondsLeft <= 5 {
		cdOpts.ColorScale.ScaleWithColor(color.RGBA{255, 80, 0, 255})
	} else {
		cdOpts.ColorScale.ScaleWithColor(color.RGBA{160, 160, 160, 255})
	}
	text.Draw(screen, fmt.Sprintf("%2ds", secondsLeft), f, cdOpts)

	// Character grid
	total := s.totalTiles()
	for i := 0; i < total; i++ {
		col := i % charGridX
		row := i / charGridX
		sx := float32(gridMarginLeft + col*gridCellSize)
		sy := float32(gridMarginTop + 4 + row*gridCellSize)

		// White background tile
		vector.DrawFilledRect(screen, sx, sy,
			float32(charDisplaySize), float32(charDisplaySize),
			color.RGBA{240, 236, 228, 255}, false)

		if s.isRandomTile(i) {
			// "?" tile
			qOpts := &text.DrawOptions{}
			qOpts.GeoM.Translate(float64(sx+4), float64(sy+4))
			qOpts.ColorScale.ScaleWithColor(color.RGBA{60, 60, 200, 255})
			text.Draw(screen, "?", f, qOpts)
		} else {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx), float64(sy))
			screen.DrawImage(s.characters[i], op)
		}

		// Selection highlight
		if i == s.selectedIndex {
			vector.StrokeRect(screen,
				sx-1, sy-1,
				float32(charDisplaySize+2), float32(charDisplaySize+2),
				1.5, color.RGBA{0, 180, 255, 255}, false)
		}
	}

	// Touch START button
	vector.DrawFilledRect(screen,
		float32(startBtnX), float32(startBtnY),
		float32(startBtnW), float32(startBtnH),
		color.RGBA{0, 120, 200, 255}, false)
	vector.StrokeRect(screen,
		float32(startBtnX), float32(startBtnY),
		float32(startBtnW), float32(startBtnH),
		1, color.RGBA{0, 220, 255, 255}, false)
	startTxtOpts := &text.DrawOptions{}
	startTxtOpts.GeoM.Translate(float64(startBtnX+20), float64(startBtnY+1))
	startTxtOpts.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "START", f, startTxtOpts)

	// Small keyboard hint (left of START button)
	hintOpts := &text.DrawOptions{}
	hintOpts.GeoM.Translate(float64(gridMarginLeft), float64(startBtnY+1))
	hintOpts.ColorScale.ScaleWithColor(color.RGBA{100, 100, 100, 255})
	text.Draw(screen, "B:PICK", f, hintOpts)
}

func (s *CharacterSelectionScene) randomSelect() {
	s.game.audioManager.PlaySE("blip")
	s.selectedIndex = int(rand.IntN(len(s.characters)))
	s.selectionX = s.selectedIndex % charGridX
	s.selectionY = s.selectedIndex / charGridX
}

func (s *CharacterSelectionScene) confirmSelection() {
	if s.selectedIndex < len(s.characters) {
		s.game.player.characterIndex = s.selectedIndex
		s.game.player.image = s.characters[s.selectedIndex]
	}
	s.game.scene = NewClassroomScene(s.game)
}

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

// shufflePlayers controls whether the character grid is randomised on each load.
// Set to true for variety, false to keep the fixed spritesheet order (à la Street Fighter II).
var shufflePlayers = true

// studentNames maps character slice index (0-based after extraction) to a display name.
// Slice index 0 = spritesheet index 1 (first real sprite), etc.
// Any index beyond the map falls back to "Student N" — safe to have more sprites than names.
var studentNames = map[int]string{
	0:  "Reece",
	1:  "Adeline",
	2:  "Lennon",
	3:  "Lennon",
	4:  "Dylan",
	5:  "Uma",
	6:  "Ansel",
	7:  "Sylvie",
	8:  "Bella",
	9:  "Hudson",
	10: "Teddy",
	11: "Teddy",
	12: "Teddy",
	13: "Sena",
	14: "Sena",
	15: "Camile",
	16: "Camile",
	17: "Camile",
	18: "Calder",
	19: "Calder",
	20: "Marlo",
	21: "Marlo",
	22: "Peter",
	23: "Bodhi",
	24: "Hudson",
	25: "Mr. J",
	26: "Ms. G",
	27: "Ms. Kim",
	28: "Jack",
	29: "Liam",
	30: "Ms. C",
}

// studentName returns the display name for character slice index i.
// Falls back gracefully to "Student N" if the index isn't in the map.
func studentName(i int) string {
	if name, ok := studentNames[i]; ok {
		return name
	}
	return fmt.Sprintf("Student %d", i+1)
}

const (
	charGridX        = 8
	charTileSize     = 16
	charDisplaySize  = 16
	selectionPadding = 2

	inactivityTimeout = 540 // frames (9s at 60fps)

	// Left panel — portrait fills most of it
	panelW = 66
	panelH = 160

	portraitScale = 4                            // 4x → 64×64px
	portraitSize  = charTileSize * portraitScale // 64px
	portraitX     = (panelW - portraitSize) / 2  // horizontally centered
	portraitY     = 22                           // below "1P" label

	// Grid area starts after left panel
	gridMarginLeft = panelW + 2
	gridMarginTop  = 18
	gridCellSize   = charDisplaySize + selectionPadding + 1 // 19px per cell

	// Touch START button
	startBtnX = 152
	startBtnY = 143
	startBtnW = 86
	startBtnH = 10

	// Frames to ignore touch on scene entry (prevents title-tap bleed-through)
	touchEntryCooldown = 20

	// Slot machine
	slotFastFrames   = 80 // total frames at/near full speed
	slotSlowFrames   = 60 // deceleration frames
	slotLandFrames   = 90 // hold on winner before auto-confirming
	slotFastInterval = 2  // advance every N frames (fast)
	slotSlowInterval = 12 // advance every N frames (slow)
)

type slotState int

const (
	slotIdle     slotState = iota
	slotSpinning           // scrolling through characters
	slotLanding            // stopped on winner, counting down to confirm
)

type CharacterSelectionScene struct {
	game              *Game
	characters        []*ebiten.Image
	charOrder         []int // display order; values are original spritesheet indices
	selectedIndex     int
	selectionX        int
	selectionY        int
	inputCooldown     int
	autoSelectCounter int
	entryCooldown     int
	portraitShader    *ebiten.Shader
	shaderTime        float32
	portraitCanvas    *ebiten.Image

	// Slot machine
	slot         slotState
	slotFrame    int
	slotTarget   int
	slotTick     int
	slotInterval int
}

func NewCharacterSelectionScene(game *Game) *CharacterSelectionScene {
	chars := extractCharacterSprites(game.assets.CharactersTileset)

	// Build the display order — shuffled or fixed depending on shufflePlayers.
	order := make([]int, len(chars))
	for i := range order {
		order[i] = i
	}
	if shufflePlayers {
		rand.Shuffle(len(order), func(i, j int) {
			order[i], order[j] = order[j], order[i]
		})
	}

	shader, err := ebiten.NewShader(portraitBgKage)
	if err != nil {
		shader = nil
	}

	canvas := ebiten.NewImage(portraitSize, portraitSize)

	return &CharacterSelectionScene{
		game:           game,
		characters:     chars,
		charOrder:      order,
		portraitShader: shader,
		portraitCanvas: canvas,
		entryCooldown:  touchEntryCooldown,
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

func (s *CharacterSelectionScene) totalTiles() int {
	return len(s.characters) + 1
}

func (s *CharacterSelectionScene) isRandomTile(i int) bool {
	return i == len(s.characters)
}

func (s *CharacterSelectionScene) tileScreenPos(i int) (x, y int) {
	col := i % charGridX
	row := i / charGridX
	return gridMarginLeft + col*gridCellSize, gridMarginTop + 4 + row*gridCellSize
}

func (s *CharacterSelectionScene) startSlot() {
	s.slotTarget = rand.IntN(len(s.characters))
	s.slot = slotSpinning
	s.slotFrame = 0
	s.slotTick = 0
	s.slotInterval = slotFastInterval
}

func (s *CharacterSelectionScene) updateSlot() {
	s.slotFrame++
	s.slotTick++

	totalSpin := slotFastFrames + slotSlowFrames

	switch s.slot {
	case slotSpinning:
		t := float64(s.slotFrame) / float64(totalSpin)
		if t > 1 {
			t = 1
		}
		s.slotInterval = slotFastInterval + int(t*t*float64(slotSlowInterval-slotFastInterval))

		if s.slotTick >= s.slotInterval {
			s.slotTick = 0
			next := (s.selectedIndex + 1) % len(s.characters)
			s.selectedIndex = next
			s.selectionX = next % charGridX
			s.selectionY = next / charGridX
			s.game.audioManager.PlaySE("blip")
		}

		if s.slotFrame >= totalSpin {
			s.selectedIndex = s.slotTarget
			s.selectionX = s.slotTarget % charGridX
			s.selectionY = s.slotTarget / charGridX
			s.slot = slotLanding
			s.slotFrame = 0
			s.game.audioManager.PlaySE("bloop")
		}

	case slotLanding:
		if s.slotFrame >= slotLandFrames {
			s.slot = slotIdle
			s.confirmSelection()
		}
	}
}

func (s *CharacterSelectionScene) Update() error {
	s.shaderTime += 1.0 / 60.0

	if s.entryCooldown > 0 {
		s.entryCooldown--
		gpad.UpdateTouch()
		return nil
	}

	if s.slot != slotIdle {
		s.updateSlot()
		return nil
	}

	gpad.UpdateTouch()

	if s.inputCooldown > 0 {
		s.inputCooldown--
	}

	hasInput := false
	total := s.totalTiles()
	maxIdx := total - 1

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
			newIdx := s.selectionY*charGridX + s.selectionX + 1
			if newIdx <= maxIdx {
				s.selectionX = min(charGridX-1, s.selectionX+1)
				s.inputCooldown = 10
				hasInput = true
			}
		}
	}

	idx := s.selectionY*charGridX + s.selectionX
	if idx > maxIdx {
		idx = maxIdx
		s.selectionX = idx % charGridX
		s.selectionY = idx / charGridX
	}
	s.selectedIndex = idx

	if gpad.PressB() || gpad.PressStart() {
		hasInput = true
		if s.isRandomTile(s.selectedIndex) {
			s.startSlot()
		} else {
			s.confirmSelection()
			return nil
		}
	}

	justReleased := inpututil.AppendJustReleasedTouchIDs(nil)
	for _, tid := range justReleased {
		tx, ty := inpututil.TouchPositionInPreviousTick(tid)

		if tx >= startBtnX && tx <= startBtnX+startBtnW &&
			ty >= startBtnY && ty <= startBtnY+startBtnH {
			hasInput = true
			s.confirmSelection()
			return nil
		}

		for i := 0; i < total; i++ {
			cx, cy := s.tileScreenPos(i)
			if tx >= cx && tx < cx+gridCellSize && ty >= cy && ty < cy+gridCellSize {
				hasInput = true
				s.game.audioManager.PlaySE("blip")
				if s.isRandomTile(i) {
					s.startSlot()
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
		s.startSlot()
	}

	return nil
}

func (s *CharacterSelectionScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 32, 255})

	f := text.NewGoXFace(bitmapfont.Face)

	// Left panel divider
	vector.DrawFilledRect(screen,
		float32(panelW), 0, 1, float32(panelH),
		color.RGBA{0, 80, 120, 255}, false)

	// "1P" label
	p1Opts := &text.DrawOptions{}
	p1Opts.GeoM.Translate(float64(panelW/2-4), 8)
	p1Opts.ColorScale.ScaleWithColor(color.RGBA{0, 180, 255, 255})
	text.Draw(screen, "1P", f, p1Opts)

	// Portrait panel
	onRandomTile := s.isRandomTile(s.selectedIndex)

	// Resolve the original spritesheet index for the current selection
	origIdx := -1
	if !onRandomTile && s.selectedIndex < len(s.charOrder) {
		origIdx = s.charOrder[s.selectedIndex]
	}

	if s.portraitShader != nil {
		s.portraitCanvas.Clear()
		if origIdx >= 0 {
			spriteOp := &ebiten.DrawImageOptions{}
			spriteOp.GeoM.Scale(portraitScale, portraitScale)
			s.portraitCanvas.DrawImage(s.characters[origIdx], spriteOp)
		}
		shaderOpts := &ebiten.DrawRectShaderOptions{}
		shaderOpts.GeoM.Translate(portraitX, portraitY)
		shaderOpts.Uniforms = map[string]any{
			"Time": s.shaderTime,
		}
		shaderOpts.Images[0] = s.portraitCanvas
		screen.DrawRectShader(portraitSize, portraitSize, s.portraitShader, shaderOpts)
	} else {
		vector.DrawFilledRect(screen,
			float32(portraitX), float32(portraitY),
			float32(portraitSize), float32(portraitSize),
			color.RGBA{10, 10, 20, 255}, false)
		if origIdx >= 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(portraitScale, portraitScale)
			op.GeoM.Translate(portraitX, portraitY)
			screen.DrawImage(s.characters[origIdx], op)
		}
	}

	// Large "?" centered in portrait when on the mystery tile
	if onRandomTile {
		qw, qh := text.Measure("?", f, 0)
		const qScale = 3.0
		qOpts := &text.DrawOptions{}
		qOpts.GeoM.Scale(qScale, qScale)
		qOpts.GeoM.Translate(
			float64(portraitX)+float64(portraitSize)/2-(qw*qScale)/2,
			float64(portraitY)+float64(portraitSize)/2-(qh*qScale)/2,
		)
		qOpts.ColorScale.ScaleWithColor(color.RGBA{80, 80, 220, 255})
		text.Draw(screen, "?", f, qOpts)
	}

	// Portrait border
	vector.StrokeRect(screen,
		float32(portraitX-1), float32(portraitY-1),
		float32(portraitSize+2), float32(portraitSize+2),
		1, color.RGBA{0, 180, 255, 255}, false)

	// Name below portrait — "Mystery" (purple) or student name (yellow)
	nameY := float64(portraitY + portraitSize + 4)
	var name string
	var nameColor color.RGBA
	if onRandomTile {
		name = "Mystery"
		nameColor = color.RGBA{180, 140, 255, 255}
	} else {
		name = studentName(origIdx)
		nameColor = color.RGBA{255, 220, 60, 255}
	}
	nw, _ := text.Measure(name, f, 0)
	nameOpts := &text.DrawOptions{}
	nameOpts.GeoM.Translate(float64(panelW)/2-nw/2, nameY)
	nameOpts.ColorScale.ScaleWithColor(nameColor)
	text.Draw(screen, name, f, nameOpts)

	// Status line: countdown or GO!
	statusY := nameY + 12
	switch s.slot {
	case slotIdle:
		secondsLeft := (inactivityTimeout - s.autoSelectCounter) / 60
		if secondsLeft < 0 {
			secondsLeft = 0
		}
		cdOpts := &text.DrawOptions{}
		cdOpts.GeoM.Translate(float64(panelW/2-6), statusY)
		if secondsLeft <= 5 {
			cdOpts.ColorScale.ScaleWithColor(color.RGBA{255, 80, 0, 255})
		} else {
			cdOpts.ColorScale.ScaleWithColor(color.RGBA{130, 130, 130, 255})
		}
		text.Draw(screen, fmt.Sprintf("%2ds", secondsLeft), f, cdOpts)

	case slotLanding:
		if (s.slotFrame/8)%2 == 0 {
			goOpts := &text.DrawOptions{}
			goOpts.GeoM.Translate(float64(panelW/2-8), statusY)
			goOpts.ColorScale.ScaleWithColor(color.RGBA{80, 255, 120, 255})
			text.Draw(screen, "GO!", f, goOpts)
		}
	}

	// Title above grid
	titleOpts := &text.DrawOptions{}
	titleOpts.GeoM.Translate(float64(gridMarginLeft), 3)
	titleOpts.ColorScale.ScaleWithColor(color.RGBA{255, 220, 60, 255})
	text.Draw(screen, "SELECT STUDENT", f, titleOpts)

	// Character grid
	total := s.totalTiles()
	for i := 0; i < total; i++ {
		col := i % charGridX
		row := i / charGridX
		sx := float32(gridMarginLeft + col*gridCellSize)
		sy := float32(gridMarginTop + 4 + row*gridCellSize)

		// Tile background — flash yellow on active slot tile
		bgColor := color.RGBA{240, 236, 228, 255}
		if s.slot == slotSpinning && i == s.selectedIndex {
			bgColor = color.RGBA{255, 240, 160, 255}
		}
		vector.DrawFilledRect(screen, sx, sy,
			float32(charDisplaySize), float32(charDisplaySize),
			bgColor, false)

		if s.isRandomTile(i) {
			qOpts := &text.DrawOptions{}
			qOpts.GeoM.Translate(float64(sx+4), float64(sy+4))
			qOpts.ColorScale.ScaleWithColor(color.RGBA{60, 60, 200, 255})
			text.Draw(screen, "?", f, qOpts)
		} else {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx), float64(sy))
			screen.DrawImage(s.characters[s.charOrder[i]], op)
		}

		// Selection highlight — pulses gold when landing
		if i == s.selectedIndex {
			hlColor := color.RGBA{0, 180, 255, 255}
			if s.slot == slotLanding && (s.slotFrame/6)%2 == 0 {
				hlColor = color.RGBA{255, 200, 0, 255}
			}
			vector.StrokeRect(screen,
				sx-1, sy-1,
				float32(charDisplaySize+2), float32(charDisplaySize+2),
				1.5, hlColor, false)
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

	// Keyboard hint
	hintOpts := &text.DrawOptions{}
	hintOpts.GeoM.Translate(float64(gridMarginLeft), float64(startBtnY+1))
	hintOpts.ColorScale.ScaleWithColor(color.RGBA{100, 100, 100, 255})
	text.Draw(screen, "B:PICK", f, hintOpts)
}

func (s *CharacterSelectionScene) confirmSelection() {
	if s.selectedIndex < len(s.characters) {
		origIdx := s.charOrder[s.selectedIndex]
		s.game.player.characterIndex = origIdx
		s.game.player.image = s.characters[origIdx]
	}
	s.game.audioManager.PlaySE("bloop")
	s.game.scene = NewClassroomScene(s.game, 1)
}

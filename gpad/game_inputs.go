package gpad

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	deadZone = .2
	id       = 1

	// Touch tuning
	dragThreshold    = 12  // pixels of movement before a drag direction locks in
	doubleTapMs      = 300 // milliseconds between taps to count as double-tap
	doubleTapMaxMove = 20  // max pixel drift between the two taps
)

// touchEnabled is set true the first time a touch is detected on the title screen.
var touchEnabled bool

// screenW/H are the logical game screen dimensions, set via Init.
var screenW, screenH int

// Init sets the logical screen size for touch input mapping.
// Call this once at startup before any input is read.
func Init(w, h int) {
	screenW = w
	screenH = h
}

// TouchEnabled returns whether touch mode is active.
func TouchEnabled() bool { return touchEnabled }

// EnableTouch activates touch input. Called from the title screen on first touch.
func EnableTouch() { touchEnabled = true }

// ── per-touch tracking ───────────────────────────────────────────────────────

type touchInfo struct {
	startX, startY int  // position when finger went down
	locked         bool // direction has been committed
}

var activeTouches = map[ebiten.TouchID]*touchInfo{}

// ── double-tap detection ─────────────────────────────────────────────────────

type tapRecord struct {
	x, y    int
	ticksAt int // ebiten tick count approximated via frame counter
}

var (
	lastTap        *tapRecord
	frameCounter   int
	doubleTapFired bool // consumed once per double-tap event
)

// ticksToMs converts our frame counter delta to milliseconds (assumes 60 fps).
func ticksToMs(delta int) int { return delta * 1000 / 60 }

// ── computed state (set once per frame by UpdateTouch) ───────────────────────

var dpadTouch struct{ up, down, left, right bool }
var bButtonTouch bool // double-tap fired this frame
var aButtonMouse bool // right-click fired this frame

// UpdateTouch must be called once per game tick.
// Full screen drag = D-pad direction; double-tap anywhere = B button.
// Mouse: left-click = B, right-click = A. No movement mapping.
func UpdateTouch() {
	// Reset frame state
	dpadTouch.up = false
	dpadTouch.down = false
	dpadTouch.left = false
	dpadTouch.right = false
	doubleTapFired = false
	bButtonTouch = false
	aButtonMouse = false

	frameCounter++

	// ── mouse clicks ─────────────────────────────────────────────────────────
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		bButtonTouch = true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		aButtonMouse = true
	}

	if !touchEnabled {
		return
	}

	// ── track newly started touches ──────────────────────────────────────────
	newIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, t := range newIDs {
		x, y := ebiten.TouchPosition(t)
		activeTouches[t] = &touchInfo{startX: x, startY: y}
	}

	// ── detect double-tap on just-released touches ───────────────────────────
	releasedIDs := inpututil.AppendJustReleasedTouchIDs(nil)
	for _, t := range releasedIDs {
		info, ok := activeTouches[t]
		if !ok {
			continue
		}
		rx, ry := inpututil.TouchPositionInPreviousTick(t)
		dx := rx - info.startX
		dy := ry - info.startY
		moved := dx*dx+dy*dy > doubleTapMaxMove*doubleTapMaxMove

		if !moved && !info.locked {
			// This was a tap — check for double-tap
			if lastTap != nil &&
				ticksToMs(frameCounter-lastTap.ticksAt) <= doubleTapMs {
				// Close enough in time — fire B
				doubleTapFired = true
				lastTap = nil
			} else {
				lastTap = &tapRecord{x: rx, y: ry, ticksAt: frameCounter}
			}
		}
		delete(activeTouches, t)
	}

	// Expire stale last-tap record
	if lastTap != nil && ticksToMs(frameCounter-lastTap.ticksAt) > doubleTapMs {
		lastTap = nil
	}

	// ── derive D-pad directions from active drags ────────────────────────────
	for t, info := range activeTouches {
		x, y := ebiten.TouchPosition(t)
		dx := x - info.startX
		dy := y - info.startY
		dist2 := dx*dx + dy*dy

		if dist2 < dragThreshold*dragThreshold {
			continue // finger hasn't moved enough yet
		}

		info.locked = true // mark as a drag, not a tap

		// Pick the dominant axis
		if abs(dx) >= abs(dy) {
			if dx > 0 {
				dpadTouch.right = true
			} else {
				dpadTouch.left = true
			}
		} else {
			if dy > 0 {
				dpadTouch.down = true
			} else {
				dpadTouch.up = true
			}
		}
	}

	bButtonTouch = bButtonTouch || doubleTapFired
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ── public input API ─────────────────────────────────────────────────────────

func MoveUp() bool {
	return dpadTouch.up ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyW) ||
		ebiten.IsGamepadButtonPressed(id, 12) ||
		ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftTop) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) < -deadZone
}

func MoveDown() bool {
	return dpadTouch.down ||
		ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyS) ||
		ebiten.IsGamepadButtonPressed(id, 13) ||
		ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftBottom) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) > deadZone
}

func MoveLeft() bool {
	return dpadTouch.left ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyA) ||
		ebiten.IsGamepadButtonPressed(id, 14) ||
		ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) < -deadZone
}

func MoveRight() bool {
	return dpadTouch.right ||
		ebiten.IsKeyPressed(ebiten.KeyRight) ||
		ebiten.IsKeyPressed(ebiten.KeyD) ||
		ebiten.IsGamepadButtonPressed(id, 15) ||
		ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) > deadZone
}

func PressStart() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, 9)
}

func PressSelect() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyShift) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, 8)
}

func PressB() bool {
	return bButtonTouch ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, 0)
}

func PressA() bool {
	return aButtonMouse ||
		inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeyDelete) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) ||
		inpututil.IsStandardGamepadButtonJustPressed(id, 1)
}

func PressFullscreen() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyF)
}

func PressQuit() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}

func PressDebug() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyH)
}

func PressP() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP)
}

func PressDigits() int {
	for i := 0; i <= 9; i++ {
		key := ebiten.KeyDigit0 + ebiten.Key(i)
		if inpututil.IsKeyJustPressed(key) {
			return i
		}
	}
	return -1
}

func TestInputs() {
	if MoveUp() {
		fmt.Println("Up")
	}
	if MoveDown() {
		fmt.Println("Down")
	}
	if MoveLeft() {
		fmt.Println("Left")
	}
	if MoveRight() {
		fmt.Println("Right")
	}
	if PressStart() {
		fmt.Println("Pause")
	}
	if PressSelect() {
		fmt.Println("Select")
	}
	if PressB() {
		fmt.Println("B Button")
	}
	if PressA() {
		fmt.Println("A Button")
	}
	if PressFullscreen() {
		fmt.Println("Fullscreen")
	}
	if PressQuit() {
		fmt.Println("Quit")
	}
	if PressDebug() {
		fmt.Println("H for Debug")
	}
	if PressP() {
		fmt.Println("P for skip to game over")
	}
	p := PressDigits()
	if p != -1 {
		fmt.Println("Press Digit: ", p)
	}
}

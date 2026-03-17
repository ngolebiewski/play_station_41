package gpad

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	deadZone = .2
	id       = 0
)

// touchEnabled is set true the first time a touch is detected on the title screen.
var touchEnabled bool

// screenW/H are the logical game screen dimensions, set via Init.
var screenW, screenH int

// Init sets the logical screen size for touch input mapping.
// Call this once at startup before any input is read.
// e.g. gpad.Init(sW, sH)
func Init(w, h int) {
	screenW = w
	screenH = h
}

// TouchEnabled returns whether touch mode is active.
func TouchEnabled() bool { return touchEnabled }

// EnableTouch activates touch input. Called from the title screen on first touch.
func EnableTouch() { touchEnabled = true }

// dpadTouch holds the current touch directions, computed once per frame.
var dpadTouch struct{ up, down, left, right bool }

// UpdateTouch should be called once per game tick.
// Left half of screen = D-pad, right half = action button.
// Touch coordinates are in logical game pixels, matching screenW/H from Init.
func UpdateTouch() {
	dpadTouch.up = false
	dpadTouch.down = false
	dpadTouch.left = false
	dpadTouch.right = false

	if !touchEnabled || screenW == 0 {
		return
	}

	for _, t := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(t)
		if x > screenW/2 {
			continue // right half = action buttons
		}
		// Center of the left half d-pad zone
		cx := screenW / 4
		cy := screenH / 2
		dx := x - cx
		dy := y - cy
		if dy < -10 {
			dpadTouch.up = true
		}
		if dy > 10 {
			dpadTouch.down = true
		}
		if dx < -10 {
			dpadTouch.left = true
		}
		if dx > 10 {
			dpadTouch.right = true
		}
	}
}

// isTouchingRight returns true if any touch is in the right half of the screen.
func isTouchingRight() bool {
	if !touchEnabled || screenW == 0 {
		return false
	}
	for _, t := range ebiten.AppendTouchIDs(nil) {
		x, _ := ebiten.TouchPosition(t)
		if x > screenW/2 {
			return true
		}
	}
	return false
}

// old school NES-like controller where D-Pad is:
//
//	   12
//	14    15.    8.      9.     0  1
//	   13.       select. start  B  A
func MoveUp() bool {
	return dpadTouch.up ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyW) ||
		ebiten.IsGamepadButtonPressed(0, 12) ||
		ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftTop) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) < -deadZone
}

func MoveDown() bool {
	return dpadTouch.down ||
		ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyS) ||
		ebiten.IsGamepadButtonPressed(0, 13) ||
		ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftBottom) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) > deadZone
}

func MoveLeft() bool {
	return dpadTouch.left ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyA) ||
		ebiten.IsGamepadButtonPressed(0, 14) ||
		ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftLeft) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) < -deadZone
}

func MoveRight() bool {
	return dpadTouch.right ||
		ebiten.IsKeyPressed(ebiten.KeyRight) ||
		ebiten.IsKeyPressed(ebiten.KeyD) ||
		ebiten.IsGamepadButtonPressed(0, 15) ||
		ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftRight) ||
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) > deadZone
}

func PressPause() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, 9)
}

func PressSelect() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyShift) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, 8)
}

func PressB() bool {
	return isTouchingRight() ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightBottom) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, 0)
}

func PressA() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeyDelete) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightRight) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, 1)
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
	if PressPause() {
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
}

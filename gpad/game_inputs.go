package gpad

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const(
	deadZone = .2
	id = 0
)

// old school NES-like controller where D-Pad is:
//.     12
//.  14    15.    8.      9.     0  1
//.     13.       select. start  B  A


func MoveUp()bool{
	return ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) || 
	ebiten.IsGamepadButtonPressed(0, 12) || 
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) < -deadZone
}

func MoveDown()bool{
	return ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) || 
	ebiten.IsGamepadButtonPressed(0, 13)  || 
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) > deadZone
}

func MoveLeft()bool{
	return ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) ||
	ebiten.IsGamepadButtonPressed(0, 14) || 
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) < -deadZone
}

func MoveRight()bool{
	return ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) ||
	ebiten.IsGamepadButtonPressed(0, 14) || 
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) > deadZone
}

func PressPause()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsStandardGamepadButtonJustPressed(0, 9)
}

func PressSelect()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeyShift) || inpututil.IsStandardGamepadButtonJustPressed(0, 8)
}

func PressB()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
	inpututil.IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightBottom) ||
	inpututil.IsStandardGamepadButtonJustPressed(0, 0)
}

func PressA()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeyDelete) ||
		inpututil.IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightRight)||
	inpututil.IsStandardGamepadButtonJustPressed(0, 1)
}

func PressFullscreen()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyF)
}

func PressQuit()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) 
}

func TestInputs(){
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
}

// func GetGamepadState() GamepadInput {
// 	var in GamepadInput
// 	ids := ebiten.AppendGamepadIDs(nil)
// 	if len(ids) == 0 {
// 		return in
// 	}

// 	id := ids[0]

// 	if ebiten.IsStandardGamepadLayoutAvailable(id) {
// 		// 1. Axis Movement
// 		rawX := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
// 		rawY := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)

// 		const deadzone = 0.2
// 		if math.Abs(rawX) > deadzone {
// 			in.GX = rawX
// 		}
// 		if math.Abs(rawY) > deadzone {
// 			in.GY = rawY
// 		}

// 		// 2. Buttons
// 		in.ButtonA = ebiten.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom)
// 		in.ButtonB = ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonRightRight)

// 		// 3. Just Pressed Actions
// 		in.AJustPressed = inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom)
// 		in.StartJustPressed = inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight)

// 		// 4. D-Pad (Standard Layout - D-Pad is on the "Left" side)
// 		in.UpJustPressed = inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop)
// 		in.DownJustPressed = inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom)
// 	} else {
// 		// Fallback for your NES-like controller mapping
// 		in.ButtonA = ebiten.IsGamepadButtonPressed(id, 1)
// 		in.ButtonB = ebiten.IsGamepadButtonPressed(id, 0)
// 		in.AJustPressed = inpututil.IsGamepadButtonJustPressed(id, 0)
// 		in.StartJustPressed = inpututil.IsGamepadButtonJustPressed(id, 9)

// 		// NES D-Pad Just Pressed
// 		in.UpJustPressed = inpututil.IsGamepadButtonJustPressed(id, 12)
// 		in.DownJustPressed = inpututil.IsGamepadButtonJustPressed(id, 13)

// 		// Map NES D-pad to Axis if held
// 		if ebiten.IsGamepadButtonPressed(id, 12) {
// 			in.GY = -1
// 		}
// 		if ebiten.IsGamepadButtonPressed(id, 13) {
// 			in.GY = 1
// 		}
// 	}

// 	return in
// }
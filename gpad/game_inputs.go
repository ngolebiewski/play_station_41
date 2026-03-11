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
	return ebiten.IsKeyPressed(ebiten.KeyUp) || 
	ebiten.IsKeyPressed(ebiten.KeyW) || 
	ebiten.IsGamepadButtonPressed(0, 12) || 
	ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftTop) ||
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) < -deadZone
}

func MoveDown()bool{
	return ebiten.IsKeyPressed(ebiten.KeyDown) || 
	ebiten.IsKeyPressed(ebiten.KeyS) || 
	ebiten.IsGamepadButtonPressed(0, 13)  || 
	ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftBottom) ||
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) > deadZone
}

func MoveLeft()bool{
	return ebiten.IsKeyPressed(ebiten.KeyLeft) || 
	ebiten.IsKeyPressed(ebiten.KeyA) ||
	ebiten.IsGamepadButtonPressed(0, 14) || 
	ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftLeft) ||
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) < -deadZone
}

func MoveRight()bool{
	return ebiten.IsKeyPressed(ebiten.KeyRight) || 
	ebiten.IsKeyPressed(ebiten.KeyD) ||
	ebiten.IsGamepadButtonPressed(0, 15) || 
	ebiten.IsStandardGamepadButtonPressed(0, ebiten.StandardGamepadButtonLeftRight) ||
	ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal) > deadZone
}

func PressPause()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeyP) || 
	inpututil.IsStandardGamepadButtonJustPressed(0, 9)
}

func PressSelect()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeyShift) || 
	inpututil.IsStandardGamepadButtonJustPressed(0, 8)
}

func PressB()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) || 
	inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
	inpututil.IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightBottom) ||
	inpututil.IsStandardGamepadButtonJustPressed(0, 0)
}

func PressA()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyZ) || 
		inpututil.IsKeyJustPressed(ebiten.KeyDelete) ||
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
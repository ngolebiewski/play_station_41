package gpad

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func MoveUp()bool{
	return ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
}

func MoveDown()bool{
	return ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
}

func MoveLeft()bool{
	return ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
}

func MoveRight()bool{
	return ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
}

func PressPause()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeyP)
}

func PressInteract()bool{
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func PressUndo()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyDelete)
}

func PressFullscreen()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyF)
}

func PressQuit()bool{
		return inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) 
}
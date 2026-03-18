package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Create text face once at package level
var textFace = text.NewGoXFace(bitmapfont.Face)

const (
	// Overlay display constants
	overlayDarkAlpha = 230 // Alpha for the dark overlay behind (90%)
	overlayBoxW      = 100
	overlayBoxH      = 80
	overlayBoxX      = (sW - overlayBoxW) / 2
	overlayBoxY      = (sH - overlayBoxH) / 2
	overlayPadding   = 5
)

// ObjectFindOverlay shows the target object to find at the start of a level
type ObjectFindOverlay struct {
	gameplay *GameplayState
	frames   int
}

// NewObjectFindOverlay creates a new overlay for showing the target object
func NewObjectFindOverlay(gameplay *GameplayState) *ObjectFindOverlay {
	return &ObjectFindOverlay{
		gameplay: gameplay,
		frames:   0,
	}
}

// Update updates the overlay state (showing for a fixed duration)
func (o *ObjectFindOverlay) Update() bool {
	o.frames++
	// Show for 3 seconds (180 frames at 60fps)
	if o.frames > 180 {
		return true // Overlay is done
	}
	return false
}

// Draw draws the overlay with the target object
func (o *ObjectFindOverlay) Draw(screen *ebiten.Image) {
	// Draw dark overlay behind
	darkImg := ebiten.NewImage(sW, sH)
	darkImg.Fill(color.RGBA{0, 0, 0, overlayDarkAlpha})
	screen.DrawImage(darkImg, &ebiten.DrawImageOptions{})

	// Draw box background
	boxImg := ebiten.NewImage(overlayBoxW, overlayBoxH)
	boxImg.Fill(color.RGBA{40, 40, 60, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(overlayBoxX), float64(overlayBoxY))
	screen.DrawImage(boxImg, op)

	// Draw border
	vector.StrokeRect(screen, float32(overlayBoxX), float32(overlayBoxY),
		float32(overlayBoxW), float32(overlayBoxH), 2, color.RGBA{100, 150, 255, 255}, false)

	// Draw target object image (centered in box)
	if o.gameplay.TargetObjectImage != nil {
		imgW := float64(o.gameplay.TargetObjectImage.Bounds().Dx())
		// imgH := float64(o.gameplay.TargetObjectImage.Bounds().Dy())

		// Center the image in box, with scaling
		imgScale := 2.0
		scaledW := imgW * imgScale
		// scaledH := imgH * imgScale

		imgX := float64(overlayBoxX) + (float64(overlayBoxW)-scaledW)/2
		imgY := float64(overlayBoxY) + overlayPadding*float64(scale)

		imgOp := &ebiten.DrawImageOptions{}
		imgOp.GeoM.Scale(imgScale, imgScale)
		imgOp.GeoM.Translate(imgX, imgY)
		screen.DrawImage(o.gameplay.TargetObjectImage, imgOp)
	}

	// Draw "Find:" label
	textOpt := &text.DrawOptions{}
	textOpt.GeoM.Translate(
		float64(overlayBoxX)+overlayPadding,
		float64(overlayBoxY)+overlayPadding+25,
	)
	text.Draw(screen, "Find:", textFace, textOpt)

	// Draw level name
	levelName := o.gameplay.GetLevelName()
	textOpt2 := &text.DrawOptions{}
	textOpt2.GeoM.Translate(
		float64(overlayBoxX)+overlayPadding,
		float64(overlayBoxY)+overlayPadding+35,
	)
	text.Draw(screen, levelName, textFace, textOpt2)

	// Draw countdown/message
	secondsLeft := (180 - o.frames) / 60
	countText := fmt.Sprintf("%d...", secondsLeft)
	textOpt3 := &text.DrawOptions{}
	textOpt3.GeoM.Translate(
		float64(overlayBoxX)+overlayBoxW-30,
		float64(overlayBoxY)+overlayBoxH-15,
	)
	text.Draw(screen, countText, textFace, textOpt3)
}

// FoundObjectMessage shows a message when the object is found
type FoundObjectMessage struct {
	frames int
}

// NewFoundObjectMessage creates a new found object message
func NewFoundObjectMessage() *FoundObjectMessage {
	return &FoundObjectMessage{
		frames: 0,
	}
}

// Update updates the message state
func (m *FoundObjectMessage) Update() bool {
	m.frames++
	if m.frames > 120 { // 2 seconds at 60fps
		return true // Message is done
	}
	return false
}

// Draw draws the found message overlay
func (m *FoundObjectMessage) Draw(screen *ebiten.Image) {
	// Draw dark overlay
	darkImg := ebiten.NewImage(sW, sH)
	darkImg.Fill(color.RGBA{0, 0, 0, 100})
	screen.DrawImage(darkImg, &ebiten.DrawImageOptions{})

	// Draw message box
	msgW := 120
	msgH := 40
	msgX := (sW - msgW) / 2
	msgY := (sH - msgH) / 2

	msgImg := ebiten.NewImage(msgW, msgH)
	msgImg.Fill(color.RGBA{50, 100, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(msgX), float64(msgY))
	screen.DrawImage(msgImg, op)

	// Draw border
	vector.StrokeRect(screen, float32(msgX), float32(msgY),
		float32(msgW), float32(msgH), 2, color.RGBA{100, 255, 100, 255}, false)

	// Draw text
	textOpt := &text.DrawOptions{}
	textOpt.GeoM.Translate(float64(msgX+5), float64(msgY+8))
	text.Draw(screen, "Object Found!", textFace, textOpt)
}

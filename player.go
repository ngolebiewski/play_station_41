package main

import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	x              float32
	y              float32
	health         int
	score          int
	speed          int
	directionRight bool
	image          *ebiten.Image
	characterIndex int
}

func NewPlayer() *Player {
	p := &Player{
		x:              sW / 2,
		y:              sH / 2,
		health:         100,
		score:          0,
		speed:          1, // World pixels per frame
		directionRight: true,
		characterIndex: 0,
	}
	return p
}

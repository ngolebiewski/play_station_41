package main

import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	x, y           float32
	health         int
	score          int
	speed          int
	directionRight bool
	image          *ebiten.Image
}

func NewPlayer() *Player {
	p := &Player{
		x:              sW / 2,
		y:              sH / 2,
		health:         100,
		score:          0,
		speed:          1, // Seems like it should be 1xscale of player. i.e. 2 if player doubled
		directionRight: true,
	}
	return p
}

package main

import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	x, y   float32
	health int
	score  int
	image  *ebiten.Image
}

func NewPlayer() *Player {
	p := &Player{
		x:      sW / 2,
		y:      sX / 2,
		health: 100,
		score:  0,
	}
	return p
}

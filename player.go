package main

import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	x,y float32
	health int
	score int
	image *ebiten.Image
}


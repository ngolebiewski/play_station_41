package main

const (
	//TFT SCREEN IS 480 x 320 so the scale will automatically go up by 2 to fit the screen.
	sW          = 240
	sH          = 160
	sX          = 2
	tileSize    = 16
	scale       = 1 //i.e. the player scale. not sure if the whole tileset and player scale should be x1 or x2
	hitboxInset = 3 * scale
)

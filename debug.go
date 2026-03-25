package main

import "github.com/ngolebiewski/play_station_41/gpad"

func DebugJumpToLevel(g *Game) {
	d := gpad.PressDigits()
	if d != -1 {
		g.scene = NewClassroomScene(g, d)
	}

}

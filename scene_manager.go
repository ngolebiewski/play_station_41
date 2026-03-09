package main

//scene_manager.go

import "github.com/hajimehoshi/ebiten/v2"

// State Machine for Scenes
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type SceneID int

// Register each scene here -- probably don't even need, but as a placeholder for organizing them.
const (
	Start SceneID = iota
	Demo
	SelectPlayer
	Objective
	Classroom // game lives here
	HUD // to overlay UI, does this make sense
	GameOver
	HighScore // inititals.
	LevelCutScene
)

//CutScene?
//MenuScene?
//TitleScene?
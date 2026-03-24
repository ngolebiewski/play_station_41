package main

import (
	"image"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ngolebiewski/play_station_41/tiled"
)

const (
	objectTileSize    = 16
	objectDisplaySize = 16
)

// extractObjectSprites extracts individual objects from the horizontal spritesheet.
// Starts at index 1 — index 0 is blank (Aseprite convention).
func extractObjectSprites(spritesheet *ebiten.Image) []*ebiten.Image {
	var objects []*ebiten.Image
	bounds := spritesheet.Bounds()
	totalObjects := bounds.Max.X / objectTileSize

	for i := 1; i < totalObjects; i++ {
		x := i * objectTileSize
		rect := image.Rect(x, 0, x+objectTileSize, objectTileSize)
		subImg := spritesheet.SubImage(rect).(*ebiten.Image)
		objects = append(objects, subImg)
	}
	return objects
}

// ObjectInstance represents a single object placed in the world
type ObjectInstance struct {
	X              float64
	Y              float64
	OrigX          float64 // Original X for animation
	OrigY          float64 // Original Y for animation
	ObjectIndex    int     // Index into the objects slice
	Image          *ebiten.Image
	IsTarget       bool    // True if this is the object to find
	IsCollected    bool    // True if player has found this object
	CollectedFrame int     // Frame when collected (for animation)
	PickupProgress float64 // 0.0 to 1.0 for pickup animation
}

// GameplayState tracks the object-finding game state
type GameplayState struct {
	// Game progression
	Level              int
	Lives              int
	Score              int
	GameOver           bool
	LevelComplete      bool
	HasFoundObject     bool
	OverlayActive      bool
	OverlayFrames      int
	FoundMessageFrames int

	// Objects
	Objects           []*ebiten.Image   // All available object sprites
	PlacedObjects     []*ObjectInstance // Objects placed in the current level
	TargetObjectIndex int               // Index of the object to find
	UsedObjectIndices []int             // Track which objects have been used as targets
	DistractorIndices []int             // Indices of distractors placed this level

	// Timer
	TimePerLevel   int  // Frames available per level
	RemainingTime  int  // Current remaining time
	TimerTriggered bool // True if timer reached 0

	// Overlay
	ShowingTargetOverlay bool
	TargetObjectImage    *ebiten.Image
}

// NewGameplayState creates a new gameplay state
func NewGameplayState(objectsImage *ebiten.Image) *GameplayState {
	objects := extractObjectSprites(objectsImage)

	return &GameplayState{
		Level:                1,
		Lives:                3,
		Score:                0,
		GameOver:             false,
		LevelComplete:        false,
		HasFoundObject:       false,
		OverlayActive:        false,
		OverlayFrames:        0,
		FoundMessageFrames:   0,
		Objects:              objects,
		PlacedObjects:        make([]*ObjectInstance, 0),
		TargetObjectIndex:    0,
		UsedObjectIndices:    make([]int, 0),
		DistractorIndices:    make([]int, 0),
		TimePerLevel:         3600, // 60 seconds at 60fps
		RemainingTime:        3600,
		TimerTriggered:       false,
		ShowingTargetOverlay: true,
	}
}

// SelectRandomObject selects an object that hasn't been used yet, tracking history
func (gs *GameplayState) SelectRandomObject() int {
	if len(gs.Objects) == 0 {
		return 0
	}

	// If we've used all objects, reset the history
	if len(gs.UsedObjectIndices) >= len(gs.Objects) {
		gs.UsedObjectIndices = make([]int, 0)
	}

	// Find an object that hasn't been used
	for {
		idx := rand.IntN(len(gs.Objects))
		found := false
		for _, used := range gs.UsedObjectIndices {
			if used == idx {
				found = true
				break
			}
		}
		if !found {
			gs.UsedObjectIndices = append(gs.UsedObjectIndices, idx)
			return idx
		}
	}
}

// PlaceObjects places objects on spawn points
// targetSpawns: spawn points marked as "find" for target placement
// otherSpawns: spawn points for placing distractors
func (gs *GameplayState) PlaceObjects(targetSpawns []tiled.SpawnPoint, otherSpawns []tiled.SpawnPoint) {
	gs.PlacedObjects = make([]*ObjectInstance, 0)
	gs.DistractorIndices = make([]int, 0)

	// Select target object
	gs.TargetObjectIndex = gs.SelectRandomObject()
	gs.TargetObjectImage = gs.Objects[gs.TargetObjectIndex]

	// Place target object on a random spawn point from targetSpawns
	if len(targetSpawns) > 0 {
		targetSpawn := targetSpawns[rand.IntN(len(targetSpawns))]
		// Add small random offset to prevent perfect overlap
		offsetX := float64(rand.IntN(8) - 4)
		offsetY := float64(rand.IntN(8) - 4)
		targetObj := &ObjectInstance{
			X:           targetSpawn.X + offsetX,
			Y:           targetSpawn.Y + offsetY,
			OrigX:       targetSpawn.X + offsetX,
			OrigY:       targetSpawn.Y + offsetY,
			ObjectIndex: gs.TargetObjectIndex,
			Image:       gs.Objects[gs.TargetObjectIndex],
			IsTarget:    true,
			IsCollected: false,
		}
		gs.PlacedObjects = append(gs.PlacedObjects, targetObj)
	}

	// Place distractors on other spawn points
	// Increase distractors: 2-3 per level, up to available spawn points
	numDistractors := min(len(otherSpawns), max(2, gs.Level+1))

	for i := 0; i < numDistractors && i < len(otherSpawns); i++ {
		distractorIdx := gs.SelectRandomObject()
		for distractorIdx == gs.TargetObjectIndex {
			distractorIdx = gs.SelectRandomObject()
		}

		spawn := otherSpawns[i]
		// Add small random offset to prevent perfect overlap
		offsetX := float64(rand.IntN(8) - 4)
		offsetY := float64(rand.IntN(8) - 4)
		distractorObj := &ObjectInstance{
			X:           spawn.X + offsetX,
			Y:           spawn.Y + offsetY,
			OrigX:       spawn.X + offsetX,
			OrigY:       spawn.Y + offsetY,
			ObjectIndex: distractorIdx,
			Image:       gs.Objects[distractorIdx],
			IsTarget:    false,
			IsCollected: false,
		}
		gs.PlacedObjects = append(gs.PlacedObjects, distractorObj)
		gs.DistractorIndices = append(gs.DistractorIndices, distractorIdx)
	}
}

// Update updates the gameplay state (timer, messages)
func (gs *GameplayState) Update() {
	if gs.GameOver || gs.LevelComplete {
		return
	}

	if gs.ShowingTargetOverlay {
		gs.OverlayFrames++
		if gs.OverlayFrames > 180 { // 3 seconds at 60fps
			gs.ShowingTargetOverlay = false
			gs.OverlayFrames = 0
		}
		return
	}

	// Update timer
	if gs.RemainingTime > 0 {
		gs.RemainingTime--
		if gs.RemainingTime <= 0 {
			gs.RemainingTime = 0
			gs.TimerTriggered = true
			gs.Lives--
			if gs.Lives <= 0 {
				gs.GameOver = true
			}
			// Lives > 0: Reset timer for retry, level stays the same
			if gs.Lives > 0 {
				gs.RemainingTime = gs.TimePerLevel
				// Do NOT call PlaceObjects - same layout repeats
			}
		}
	}

	// Update found message display
	if gs.FoundMessageFrames > 0 {
		gs.FoundMessageFrames--
		if gs.FoundMessageFrames <= 0 {
			gs.LevelComplete = true
		}
	}
}

// ObjectFound should be called when the player collects the target object
func (gs *GameplayState) ObjectFound() {
	if gs.HasFoundObject {
		return
	}

	gs.HasFoundObject = true
	gs.FoundMessageFrames = 120 // 2 seconds at 60fps
	gs.Score += calculateLevelScore(gs.Level, gs.RemainingTime)
	gs.Level++
	gs.RemainingTime = gs.TimePerLevel
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func calculateLevelScore(level int, remainingTime int) int {
	// Score based on level and time remaining
	baseScore := level * 100
	timeBonus := (remainingTime / 60) * 10 // Bonus for time left
	return baseScore + timeBonus
}

// GetLevelName returns the grade level name for the current level
func (gs *GameplayState) GetLevelName() string {
	levelNames := []string{
		"",          // 0 - unused
		"3K",        // 1
		"Pre-K",     // 2
		"K",         // 3
		"1st Grade", // 4
		"2nd Grade", // 5
		"3rd Grade", // 6
		"4th Grade", // 7
		"5th Grade", // 8
	}

	if gs.Level < 1 || gs.Level >= len(levelNames) {
		return "5th" // Cap at 5th grade
	}
	return levelNames[gs.Level]
}

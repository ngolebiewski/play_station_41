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
	CountedAsFound bool    // True if this object has been registered in ObjectsFound count
	CollectedFrame int     // Frame when collected (for animation)
	PickupProgress float64 // 0.0 to 1.0 for pickup animation
}

// GameplayState tracks the object-finding game state
type GameplayState struct {
	// Game progression
	Level              int
	Lives              int
	Score              int
	Points             int // Bonus points from dismissing distractors
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
	ObjectsToFind     int               // Number of target objects to find this level
	ObjectsFound      int               // Number of target objects already found

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
		Points:               0,
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
		ObjectsToFind:        1,
		ObjectsFound:         0,
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

	// Reset object counting
	gs.ObjectsFound = 0

	// Select target object
	gs.TargetObjectIndex = gs.SelectRandomObject()
	gs.TargetObjectImage = gs.Objects[gs.TargetObjectIndex]

	// Place multiple target objects (same type) at different spawn points
	// Level 1: 1 object, Level 2: 2 objects, Level 3+: 3 objects
	numTargets := 1
	if gs.Level >= 2 && gs.Level < 3 {
		numTargets = 2
	} else if gs.Level >= 3 {
		numTargets = 3
	}
	gs.ObjectsToFind = min(numTargets, len(targetSpawns))

	// Place target objects at different spawn points
	usedTargetIdx := make(map[int]bool)
	for i := 0; i < gs.ObjectsToFind && len(targetSpawns) > 0; i++ {
		var spawnIdx int
		// Find a spawn point we haven't used yet
		for {
			spawnIdx = rand.IntN(len(targetSpawns))
			if !usedTargetIdx[spawnIdx] {
				usedTargetIdx[spawnIdx] = true
				break
			}
		}
		targetSpawn := targetSpawns[spawnIdx]
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

	// Place distractors on other spawn points (prioritize otherSpawns, then use unused targetSpawns)
	// SKIP distractors on Level 1 (easy mode)
	if gs.Level > 1 {
		// Increase distractors: 2-3 per level, up to available spawn points
		numDistractors := max(2, gs.Level+1)

		// First use otherSpawns if available
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
			numDistractors--
		}

		// If we need more distractors, use unused targetSpawns
		if numDistractors > 0 && len(targetSpawns) > gs.ObjectsToFind {
			for i := 0; i < numDistractors; i++ {
				var spawnIdx int
				// Find an unused target spawn point
				for {
					spawnIdx = rand.IntN(len(targetSpawns))
					if !usedTargetIdx[spawnIdx] {
						usedTargetIdx[spawnIdx] = true
						break
					}
				}

				distractorIdx := gs.SelectRandomObject()
				for distractorIdx == gs.TargetObjectIndex {
					distractorIdx = gs.SelectRandomObject()
				}

				spawn := targetSpawns[spawnIdx]
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

// GetLevelTimeLimit returns the time limit in frames for the given level.
// Configure per-level timing here:
// - Level 1 (3K): 30 seconds, easy, no distractors
// - Level 2 (Pre-K): 25 seconds, medium difficulty
// - Level 3 (K): 10 seconds, hard
// - Level 4 (1st Grade): 60 seconds, challenging with many objects
// - Level 5 (2nd Grade): 20 seconds, increasingly difficult
// - Level 6 (3rd Grade): 15 seconds, very challenging
// - Level 7 (4th Grade): 25 seconds, medium pace
// - Level 8 (5th Grade): 30 seconds, final level before graduation
// Default: 30 seconds (1800 frames at 60fps)
func GetLevelTimeLimit(level int) int {
	switch level {
	case 1:
		return 20 * 60 // 30 seconds - 3K, easy, no distractors
	case 2:
		return 25 * 60 // 25 seconds - Pre-K
	case 3:
		return 10 * 60 // 10 seconds - K, hard
	case 4:
		return 60 * 60 // 60 seconds - 1st Grade
	case 5:
		return 100 * 60 // 100 seconds - 2nd Grade
	case 6:
		return 15 * 60 // 15 seconds - 3rd Grade
	case 7:
		return 25 * 60 // 25 seconds - 4th Grade
	case 8:
		return 30 * 60 // 30 seconds - 5th Grade (final level)
	default:
		return 30 * 60 // 30 seconds default
	}
}

// ObjectFound should be called when the player collects a target object
func (gs *GameplayState) ObjectFound() {
	gs.ObjectsFound++
	gs.Points += 41 // Award 41 points per object found

	// Only mark level complete when ALL objects are found
	if gs.ObjectsFound >= gs.ObjectsToFind {
		gs.HasFoundObject = true
		gs.FoundMessageFrames = 30 // 1 second at 60fps

		// Calculate time bonus: 5 points per second remaining
		secondsRemaining := gs.RemainingTime / 60
		timeBonus := secondsRemaining * 5
		gs.Points += timeBonus

		gs.Score += calculateLevelScore(gs.Level, gs.RemainingTime)
		gs.Level++
		gs.RemainingTime = GetLevelTimeLimit(gs.Level)
	}
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

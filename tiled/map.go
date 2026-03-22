// Package tiled wraps github.com/lafriks/go-tiled for use with Ebitengine.
// The go-tiled library provides the Map, Layer, Tileset, and Object types.
// This package adds only what go-tiled doesn't: a Spawn value type and
// Ebitengine-specific rendering/collision helpers.
package tiled

// Spawn represents a single point object from the "spawn" object layer.
// Type is the object's Name field in Tiled ("student", "teacher", "find").
// X and Y are world pixel coordinates.
type Spawn struct {
	X, Y float64
	Type string // "student", "teacher", "find"
}

// SpawnPoint represents a single spawn point with its type.
// Used for object placement in gameplay.
type SpawnPoint struct {
	X    float64
	Y    float64
	Type string
}


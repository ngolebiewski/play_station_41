package tiled

import (
	"io/fs"
	"path"
	"strings"

	gotiled "github.com/lafriks/go-tiled"
)

// LoadMapFS loads a Tiled TMX map file from an fs.FS (e.g. an embed.FS).
//
// It uses fs.Sub to give go-tiled a filesystem rooted at the same directory
// as the map file. This means go-tiled can open "CLASSROOM.tsx" directly
// without needing to know the full path inside the embed.
//
// Example:
//
//	//go:embed tiled_files/classroom_1.tmx
//	//go:embed tiled_files/CLASSROOM.tsx
//	var embeddedAssets embed.FS
//
//	m, err := tiled.LoadMapFS(embeddedAssets, "tiled_files/classroom_1.tmx")
func LoadMapFS(fsys fs.FS, filePath string) (*gotiled.Map, error) {
	dir := path.Dir(filePath)
	base := path.Base(filePath)

	// Sub creates a new FS rooted at dir (e.g. "tiled_files").
	// go-tiled will then open "CLASSROOM.tsx" directly within that root —
	// matching exactly what the .tmx's <tileset source="CLASSROOM.tsx"/> says.
	subFS, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil, err
	}

	f, err := subFS.Open(base)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// baseDir is "." because subFS is already rooted at the map's directory.
	return gotiled.LoadReader(".", f, gotiled.WithFileSystem(subFS))
}

// GetSpawns returns all point objects from the layer named "spawn".
// Objects are matched case-insensitively so "Spawn" and "SPAWN" also work.
//
// In Tiled, spawn points are stored as point objects in an Object Layer.
// The object's Name field becomes Spawn.Type ("student", "teacher", "find").
func GetSpawns(m *gotiled.Map) []Spawn {
	var spawns []Spawn
	for _, og := range m.ObjectGroups {
		if !strings.EqualFold(og.Name, "spawn") {
			continue
		}
		for _, obj := range og.Objects {
			spawns = append(spawns, Spawn{
				X:    obj.X,
				Y:    obj.Y,
				Type: obj.Name,
			})
		}
	}
	return spawns
}

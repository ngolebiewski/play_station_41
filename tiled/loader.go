package tiled

import (
	"encoding/json"
	"io/fs"
)

func LoadMapFS(fsys fs.FS, path string) (*Map, error) {
	b, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	var m Map
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// GetSpawns extracts the slice of objects from the "spawn" layer
func (m *Map) GetSpawns() []Spawn {
	var spawns []Spawn
	for _, layer := range m.Layers {
		if layer.Name == "spawn" {
			for _, obj := range layer.Objects {
				spawns = append(spawns, Spawn{
					X:    obj.X,
					Y:    obj.Y,
					Type: obj.Name,
				})
			}
		}
	}
	return spawns
}

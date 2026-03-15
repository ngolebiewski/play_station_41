package tiled

type Map struct {
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	TileWidth  int       `json:"tilewidth"`
	TileHeight int       `json:"tileheight"`
	Layers     []Layer   `json:"layers"`
	Tilesets   []Tileset `json:"tilesets"`
}

type Layer struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`   // "tilelayer", "objectgroup", "group"
	Data    []uint32 `json:"data"`   // For tile layers
	Layers  []Layer  `json:"layers"` // For groups
	Visible bool     `json:"visible"`
	Objects []Object `json:"objects,omitempty"`
}

type Tileset struct {
	FirstGID int `json:"firstgid"`
}

type Object struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type Spawn struct {
	X, Y float64
	Type string // "student", "teacher", "find"
}

package main

// Embed images into the GO Binary. Essential for WASM the web, and all around useful.
// TBD on how to save a local file for high score

import (
	"bytes"
	"embed"
	"encoding/json"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed aseprite_art/title_ps41_c.png
//go:embed aseprite_art/characters.png
//go:embed aseprite_art/objects.png
//go:embed aseprite_art/default_player.png
//go:embed aseprite_art/CLASSROOM.png
//go:embed aseprite_art/CLASSROOM_MAX.png

//go:embed tiled_files/classroom_1.tmx
//go:embed tiled_files/CLASSROOM.tsx
//go:embed tiled_files/classroom_1.png

//go:embed tiled_files/classroom_2.tmx
//go:embed tiled_files/CLASSROOM_MAX.tsx

//go:embed tiled_files/CLASSROOM_busy2.tmx

//go:embed tiled_files/classroom_maze.tmx

//go:embed tiled_files/classroom_final.tmx

// could embed the entire directory with 'art/**' but there are files I don't want in there to keep the build small.
// For example. ASEPRITE files with layers, and unused artworks or test files.

var embeddedAssets embed.FS

func loadImage(path string) (*ebiten.Image, error) {
	data, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

func loadJSON(path string, v any) error {
	data, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type Assets struct {
	TitleImage         *ebiten.Image
	ClassroomTileset_1 *ebiten.Image
	ClassroomTileset_2 *ebiten.Image
	CharactersTileset  *ebiten.Image
	ObjectsTileset     *ebiten.Image
	DefaultPlayer      *ebiten.Image
	// StickersTileset  *ebiten.Image

}

func LoadAssets() *Assets {
	title, err := loadImage("aseprite_art/title_ps41_c.png")
	if err != nil {
		log.Fatal(err)
	}

	classroomTiles_1, err := loadImage("aseprite_art/CLASSROOM.png")
	if err != nil {
		log.Fatal(err)
	}
	classroomTiles_2, err := loadImage("aseprite_art/CLASSROOM_MAX.png")
	if err != nil {
		log.Fatal(err)
	}

	char, err := loadImage("aseprite_art/characters.png")
	if err != nil {
		log.Fatal(err)
	}

	obj, err := loadImage("aseprite_art/objects.png")
	if err != nil {
		log.Fatal(err)
	}

	d, err := loadImage("aseprite_art/default_player.png")
	if err != nil {
		log.Fatal(err)
	}

	// stick, err := loadImage("aseprite_art/stickers.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return &Assets{
		TitleImage:         title,
		ClassroomTileset_1: classroomTiles_1,
		ClassroomTileset_2: classroomTiles_2,
		CharactersTileset:  char,
		ObjectsTileset:     obj,
		DefaultPlayer:      d,
		// StickersTileset: stick,
	}
}

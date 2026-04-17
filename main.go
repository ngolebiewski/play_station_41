package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ngolebiewski/play_station_41/gpad"
)

type Game struct {
	scene      Scene
	assets     *Assets
	player     *Player
	debug      bool
	gameplay   *GameplayState
	quitFrames int
}

func NewGame() *Game {

	// 2. Asset Setup
	assets := LoadAssets()
	player := NewPlayer()
	player.image = assets.DefaultPlayer

	// 3. Create the Game struct
	g := &Game{
		assets:   assets,
		player:   player,
		debug:    false,
		gameplay: NewGameplayState(assets.ObjectsTileset),
	}

	// 4. Initialize the starting scene
	g.scene = NewTitleScene(g)

	return g
}

func (g *Game) Update() error {

	if g.scene != nil {
		if err := g.scene.Update(); err != nil {
			return err
		}
	}
	if gpad.PressFullscreen() {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if gpad.PressToQuit() {
		g.quitFrames++

		// 120 frames = 2 seconds at 60fps
		if g.quitFrames >= 120 {
			return fmt.Errorf("intentional arcade exit")
		}
	} else {
		// Reset the timer if either button is released
		g.quitFrames = 0
	}

	if gpad.PressDebug() {
		g.debug = !g.debug
		fmt.Println("Debug mode on: ", g.debug)
	}
	if g.debug {
		gpad.TestInputs()
		DebugJumpToLevel(g)
		// ids := ebiten.AppendGamepadIDs(nil)
		// for _, id := range ids {
		// 	fmt.Printf("Device Found: %s | ID: %d\n", ebiten.GamepadName(id), id)
		// }
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
	if g.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n\n\n\n\nTPS: %0.2f", ebiten.ActualTPS()))
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(sW*sX, sH*sX)
	op := &ebiten.RunGameOptions{}

	// Default is false, which means it starts FOCUSED.
	op.InitUnfocused = false

	// This ensures it stays running even if the DSI touch driver
	// flickers focus for a split second.
	ebiten.SetRunnableOnUnfocused(true)

	// This is safe for both Pi 5 and WASM
	// Run with DISPLAY=:0 ARCADE_MODE=1 ./playstation41_pi
	if os.Getenv("ARCADE_MODE") == "1" {
		ebiten.SetFullscreen(true)
		ebiten.SetCursorMode(ebiten.CursorModeHidden) // Also hides mouse for arcade
	}

	gpad.Init(sW, sH)
	ebiten.SetWindowTitle("Play Station 41")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

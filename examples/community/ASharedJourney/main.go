package main

import (
	"time"

	"github.com/gandrin/ASharedJourney/supervisor"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/gandrin/ASharedJourney/tiles"
	"golang.org/x/image/colornames"

	"github.com/gandrin/ASharedJourney/mechanics"
	"github.com/gandrin/ASharedJourney/menu"
	"github.com/gandrin/ASharedJourney/music"
	"github.com/gandrin/ASharedJourney/shared"
)

const frameRate = 60

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "A Shared Journey",
		Bounds: pixel.R(0, 0, 800, 800),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	shared.Win = win

	menu.Menu(menu.MainMenuImage, "    Loading ...",pixel.V(200,150), false, music.SOUND_NONE)

	music.Music.Start()

	<-music.GameMusicLoader
	menu.Menu(menu.MainMenuImage, "Press ENTER to PLAY ...", pixel.V(180,150),true, music.SOUND_EFFECT_START_GAME)
	menu.Menu(menu.RulesMenuImage, "Press ENTER to START", pixel.V(180,150),true, music.SOUND_EFFECT_START_GAME)

	world := tiles.NextLevel()


	fps := time.Tick(time.Second / frameRate)

	gameEventsChannel := supervisor.Start()

	newWorldChannel := mechanics.Start(gameEventsChannel, world)

	for !win.Closed() {
		win.Clear(colornames.Black)
		supervisor.Sup.Play()
		mechanics.Mecha.Play()
		upToDateWorld := <-newWorldChannel
		tiles.DrawMap(upToDateWorld.BackgroundTiles)
		tiles.DrawMap(upToDateWorld.Obstacles)
		tiles.DrawMap(upToDateWorld.WinStars)
		tiles.DrawMap(upToDateWorld.Water)
		tiles.DrawMap(upToDateWorld.Movables)
		tiles.DrawMap(upToDateWorld.Players)
		tiles.DrawMap(upToDateWorld.Holes)
		win.Update()
		<-fps
	}
}

func main() {

	pixelgl.Run(run)
}

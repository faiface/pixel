package main

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Relay",
		Bounds: pixel.R(0, 0, 2048, 512),
		VSync:  true,
	}

	w, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	s := newState()
	colors := teamColors(s.teams)

	start := time.Now()

	for !w.Closed() && !s.gameOver {
		if w.JustPressed(pixelgl.KeySpace) {
			w.Clear(colornames.Peru)
			s = updateState(s)
			render(s, w, time.Since(start), colors)
		}
		w.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

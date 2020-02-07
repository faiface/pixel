package main

import (
	"math/rand"
	"relay/game"
	"relay/gfx"
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

	rand.Seed(time.Now().UnixNano())

	s := game.NewState()

	start := time.Now()

	for !w.Closed() && !s.GameOver {
		w.Clear(colornames.Peru)

		//sOld := s

		switch {
		case w.JustPressed(pixelgl.KeyQ):
			return
		case w.JustPressed(pixelgl.KeySpace):
			s = game.UpdateState(s)
		}
		rs := gfx.RenderState{
			Animating: true,
			Frames:    10,
		}
		for rs.Animating {
			rs = gfx.Render(rs, s, w, time.Since(start))
		}
		w.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

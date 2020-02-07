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

	w.Clear(colornames.Peachpuff)
	for !w.Closed() && !s.GameOver {
		sOld := s

		rs := gfx.RenderState{
			Animating: false,
			Frames:    20,
		}

		switch {
		case w.Pressed(pixelgl.KeyQ):
			return
		case w.JustPressed(pixelgl.KeySpace) || true:
			rs.Animating = true
			s = game.UpdateState(s, sOld)
			if s.GameOver {
				s = game.NewState()
				sOld = s
			}
		}
		for rs.Animating {
			rs = gfx.Render(rs, sOld, s, w, time.Since(start))
			w.Update()
		}
		w.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

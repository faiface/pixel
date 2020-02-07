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

	w.Clear(colornames.Peachpuff)

	rs := gfx.RenderState{
		Animating: false,
		Frames:    20,
	}
	sOld := s

	for !w.Closed() && !s.GameOver {
		switch {
		case w.Pressed(pixelgl.KeyQ):
			return
		case rs.Animating:
			rs = gfx.Render(rs, sOld, s, w)
			if !rs.Animating {
				sOld = s
			}
		case w.Pressed(pixelgl.KeySpace):
			rs.Animating = true
			rs.Frame = 0
			s = game.UpdateState(s, sOld)
			if s.GameOver {
				s = game.NewState()
				sOld = s
			}
		}

		w.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

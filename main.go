package main

import (
	"fmt"
	"log"
	"math/rand"
	"relay/game"
	"relay/gfx"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() error {
	cfg := pixelgl.WindowConfig{
		Title:  "Relay",
		Bounds: pixel.R(0, 0, 2400, 1024),
		VSync:  true,
	}

	w, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())

	s := game.NewState()

	rs := gfx.RenderState{
		Animating: true,
		Frames:    20,
	}
	sb, err := gfx.NewSpriteBank()
	if err != nil {
		return err
	}
	sOld := s
	turn := 1

	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	for !w.Closed() && !s.GameOver {
		switch {
		case w.Pressed(pixelgl.KeyQ):
			return nil
		case rs.Animating:
			rs = gfx.Render(rs, sOld, s, w, *sb)
			if !rs.Animating {
				sOld = s
			}
		case w.Pressed(pixelgl.KeySpace):
			log.Printf("TURN %d", turn)
			rs.Animating = true
			rs.Frame = 0
			s = game.UpdateState(s, sOld)
			turn++
			if s.GameOver {
				s = game.NewState()
				sOld = s
				turn = 1
			}
		}

		w.Update()
		frames++
		select {
		case <-second:
			w.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
	return nil
}

func pixelRun() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	pixelgl.Run(pixelRun)
}

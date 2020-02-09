package main

import (
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
		Bounds: pixel.R(0, 0, 2400, 512),
		VSync:  true,
	}

	w, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())

	s := game.NewState()

	sb, err := gfx.NewSpriteBank()
	if err != nil {
		return err
	}
	sOld := s
	turn := 1

	cmdC := make(chan []game.Command)
	go func() { cmdC <- game.PollCommands(s) }()

	stateCA := make(chan game.State)
	stateCB := make(chan game.State)

	go gfx.RenderLoop(w, s, sOld, stateCA, sb)

	for !w.Closed() {
		switch {
		case w.Pressed(pixelgl.KeyQ):
			return nil
		case w.JustPressed(pixelgl.KeySpace) || true:
			cmds := <-cmdC
			s = game.UpdateState(s, sOld, cmds)
			turn++
			if s.GameOver {
				s = game.NewState()
				sOld = s
				turn = 1
			}
			go func() {
				s := <-stateCB
				cmdC <- game.PollCommands(s)
			}()
			stateCA <- s
			stateCB <- s
		}

		w.UpdateInput()
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

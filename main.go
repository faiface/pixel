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

	stateC := make(chan game.State)

	go gfx.RenderLoop(w, s, stateC, sb)
	go game.CommandLoop(w, s, stateC)

	for !w.Closed() {
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

package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type scene int

const (
	start scene = iota
	menu
	game
	credits
	end
)

func (s *scene) nextScene() {
	*s = (*s + 1) % 5
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Click On The Screen!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	s := start

	for !win.Closed() {
		switch s {
		case start:
			win.Clear(colornames.Lavender)
		case menu:
			win.Clear(colornames.Turquoise)
		case game:
			win.Clear(colornames.Lightyellow)
		case credits:
			win.Clear(colornames.Lightpink)
		case end:
			win.Clear(colornames.Sandybrown)
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			s.nextScene()
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

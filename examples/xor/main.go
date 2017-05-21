package main

import (
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:     "Xor",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
		VSync:     true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(nil)

	canvas := pixelgl.NewCanvas(win.Bounds())

	start := time.Now()
	for !win.Closed() {
		// in case window got resized, we also need to resize our canvas
		canvas.SetBounds(win.Bounds())

		offset := math.Sin(time.Since(start).Seconds()) * 300

		// clear the canvas to be totally transparent and set the xor compose method
		canvas.Clear(pixel.Alpha(0))
		canvas.SetComposeMethod(pixel.ComposeXor)

		// red circle
		imd.Clear()
		imd.Color = pixel.RGB(1, 0, 0)
		imd.Push(win.Bounds().Center().Add(pixel.V(-offset, 0)))
		imd.Circle(200, 0)
		imd.Draw(canvas)

		// blue circle
		imd.Clear()
		imd.Color = pixel.RGB(0, 0, 1)
		imd.Push(win.Bounds().Center().Add(pixel.V(offset, 0)))
		imd.Circle(150, 0)
		imd.Draw(canvas)

		// yellow circle
		imd.Clear()
		imd.Color = pixel.RGB(1, 1, 0)
		imd.Push(win.Bounds().Center().Add(pixel.V(0, -offset)))
		imd.Circle(100, 0)
		imd.Draw(canvas)

		// magenta circle
		imd.Clear()
		imd.Color = pixel.RGB(1, 0, 1)
		imd.Push(win.Bounds().Center().Add(pixel.V(0, offset)))
		imd.Circle(50, 0)
		imd.Draw(canvas)

		win.Clear(colornames.Green)
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

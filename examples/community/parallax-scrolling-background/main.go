package main

import (
	"image"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

const (
	windowWidth      = 600
	windowHeight     = 450
	foregroundHeight = 149
	// This is the scrolling speed (pixels per second)
	// Negative values will make background to scroll to the left,
	// positive to the right.
	backgroundSpeed = -60
	foregroundSpeed = -120
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Parallax scrolling demo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Pic must have double the width of the window, as it will scroll to the left or right
	picBackground, err := loadPicture("background.png")
	if err != nil {
		panic(err)
	}
	picForeground, err := loadPicture("foreground.png")
	if err != nil {
		panic(err)
	}

	background := NewScrollingBackground(picBackground, windowWidth, windowHeight, backgroundSpeed)
	foreground := NewScrollingBackground(picForeground, windowWidth, foregroundHeight, foregroundSpeed)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		background.Update(win, dt)
		foreground.Update(win, dt)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

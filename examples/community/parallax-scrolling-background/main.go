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
	backgroundSpeed = 60
	foregroundSpeed = 120
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Scrolling background demo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Pic must have double the width of the window, as it will scroll to the left
	picBackground, err := loadPicture("background.png")
	if err != nil {
		panic(err)
	}
	picForeground, err := loadPicture("foreground.png")
	if err != nil {
		panic(err)
	}

	background := newScrollingBackground(picBackground, windowWidth, windowHeight, windowWidth)
	foreground := newScrollingBackground(picForeground, windowWidth, foregroundHeight, windowWidth)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		background.update(win, backgroundSpeed, dt)
		foreground.update(win, foregroundSpeed, dt)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

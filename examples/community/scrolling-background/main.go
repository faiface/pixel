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
	windowWidth  = 600
	windowHeight = 450
	// This is the scrolling speed
	linesPerSecond = 60
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
	pic, err := loadPicture("gamebackground.png")
	if err != nil {
		panic(err)
	}

	// Backgrounds are made taking the left and right halves of the image
	background1 := pixel.NewSprite(pic, pixel.R(0, 0, windowWidth, windowHeight))
	background2 := pixel.NewSprite(pic, pixel.R(windowWidth, 0, windowWidth*2, windowHeight))

	// In the beginning, vector1 will put background1 filling the whole window, while vector2 will
	// put background2 just at the right side of the window, out of view
	vector1 := pixel.V(windowWidth/2, windowHeight/2)
	vector2 := pixel.V(windowWidth+(windowWidth/2), windowHeight/2)

	i := float64(0)
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		// When one of the backgrounds has completely scrolled, we swap displacement vectors,
		// so the backgrounds will swap positions too regarding the previous iteration,
		// thus making the background endless.
		if i <= -windowWidth {
			i = 0
			vector1, vector2 = vector2, vector1
		}
		// This delta vector will move the backgrounds to the left
		d := pixel.V(-i, 0)
		background1.Draw(win, pixel.IM.Moved(vector1.Sub(d)))
		background2.Draw(win, pixel.IM.Moved(vector2.Sub(d)))
		i -= linesPerSecond * dt
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

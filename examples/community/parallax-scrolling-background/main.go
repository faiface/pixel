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

type scrollingBackground struct {
	Pic   pixel.Picture
	speed int
}

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

	// Backgrounds are made taking the left and right halves of the image
	background1 := pixel.NewSprite(picBackground, pixel.R(0, 0, windowWidth, windowHeight))
	background2 := pixel.NewSprite(picBackground, pixel.R(windowWidth, 0, windowWidth*2, windowHeight))
	foreground1 := pixel.NewSprite(picForeground, pixel.R(0, 0, windowWidth, foregroundHeight))
	foreground2 := pixel.NewSprite(picForeground, pixel.R(windowWidth, 0, windowWidth*2, foregroundHeight))

	// In the beginning, vector1 will put background1 filling the whole window, while vector2 will
	// put background2 just at the right side of the window, out of view
	backgroundVector1 := pixel.V(windowWidth/2, (windowHeight/2)+1)
	backgroundVector2 := pixel.V(windowWidth+(windowWidth/2), (windowHeight/2)+1)

	foregroundVector1 := pixel.V(windowWidth/2, (foregroundHeight/2)+1)
	foregroundVector2 := pixel.V(windowWidth+(windowWidth/2), (foregroundHeight/2)+1)

	bi, fi := 0., 0.
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		// When one of the backgrounds has completely scrolled, we swap displacement vectors,
		// so the backgrounds will swap positions too regarding the previous iteration,
		// thus making the background endless.
		if bi <= -windowWidth {
			bi = 0
			backgroundVector1, backgroundVector2 = backgroundVector2, backgroundVector1
		}
		if fi <= -windowWidth {
			fi = 0
			foregroundVector1, foregroundVector2 = foregroundVector2, foregroundVector1
		}
		// This delta vector will move the backgrounds to the left
		db := pixel.V(-bi, 0)
		df := pixel.V(-fi, 0)
		background1.Draw(win, pixel.IM.Moved(backgroundVector1.Sub(db)))
		background2.Draw(win, pixel.IM.Moved(backgroundVector2.Sub(db)))
		foreground1.Draw(win, pixel.IM.Moved(foregroundVector1.Sub(df)))
		foreground2.Draw(win, pixel.IM.Moved(foregroundVector2.Sub(df)))
		bi -= backgroundSpeed * dt
		fi -= foregroundSpeed * dt
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

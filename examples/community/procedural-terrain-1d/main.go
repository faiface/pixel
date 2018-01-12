package main

import (
	"image"
	"math/rand"
	"os"
	"time"

	_ "image/jpeg"

	perlin "github.com/aquilax/go-perlin"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	width  = 800
	height = 600
	// Top of the mountain must be around the half of the screen height
	verticalOffset = height / 2
	// Perlin noise provides variations in values between -1 and 1,
	// we multiply those so they're visible on screen
	scale            = 100
	waveLength       = 100
	alpha            = 2.
	beta             = 2.
	n                = 3
	maximumSeedValue = 100
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Procedural terrain 1D",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	pic, err := loadPicture("stone.jpg")
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(pic)

	drawTerrain(win, imd)

	for !win.Closed() {
		if win.JustPressed(pixelgl.KeySpace) {
			drawTerrain(win, imd)
		}
		win.Update()
	}
}

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

func drawTerrain(win *pixelgl.Window, imd *imdraw.IMDraw) {
	var seed = rand.Int63n(maximumSeedValue)
	p := perlin.NewPerlin(alpha, beta, n, seed)

	imd.Clear()
	win.Clear(colornames.Skyblue)
	for x := 0.; x < width; x++ {
		y := p.Noise1D(x/waveLength)*scale + verticalOffset
		renderTexturedLine(x, y, imd)
	}
	imd.Draw(win)
}

// Render a textured line in position x with a height y.
// Note that the textured line is just a 1 px width rectangle.
// We push the opposite vertices of that rectangle and specify the points of the
// texture we want to apply to them. Pixel will fill the rest of the rectangle interpolating the texture.
func renderTexturedLine(x, y float64, imd *imdraw.IMDraw) {
	imd.Intensity = 1.
	imd.Picture = pixel.V(x, 0)
	imd.Push(pixel.V(x, 0))
	imd.Picture = pixel.V(x+1, y)
	imd.Push(pixel.V(x+1, y))
	imd.Rectangle(0)
}

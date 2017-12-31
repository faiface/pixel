package main

import (
	"image"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	_ "image/png"
)

const (
	windowWidth  = 800
	windowHeight = 800
	// sprite tiles are squared, 64x64 size
	tileSize = 64
	f        = 0 // floor identifier
	w        = 1 // wall identifier
)

var levelData = [][]uint{
	{f, f, f, f, f, f}, // This row will be rendered in the lower left part of the screen (closer to the viewer)
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, w, w, w, w, w}, // And this in the upper right
}
var win *pixelgl.Window
var offset = pixel.V(400, 325)
var floorTile, wallTile *pixel.Sprite

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

func run() {
	var err error

	cfg := pixelgl.WindowConfig{
		Title:  "Isometric demo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic, err := loadPicture("castle.png")
	if err != nil {
		panic(err)
	}

	wallTile = pixel.NewSprite(pic, pixel.R(0, 448, tileSize, 512))
	floorTile = pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192))

	depthSort()

	for !win.Closed() {
		win.Update()
	}
}

// Draw level data tiles to window, from farthest to closest.
// In order to achieve the depth effect, we need to render tiles up to down, being lower
// closer to the viewer (see painter's algorithm). To do that, we need to process levelData in reverse order,
// so its first row is rendered last, as OpenGL considers its origin to be in the lower left corner of the display.
func depthSort() {
	for x := len(levelData) - 1; x >= 0; x-- {
		for y := len(levelData[x]) - 1; y >= 0; y-- {
			isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
			mat := pixel.IM.Moved(offset.Add(isoCoords))
			// Not really needed, just put to show bigger blocks
			mat = mat.ScaledXY(win.Bounds().Center(), pixel.V(2, 2))
			tileType := levelData[x][y]
			if tileType == f {
				floorTile.Draw(win, mat)
			} else {
				wallTile.Draw(win, mat)
			}
		}
	}
}

func cartesianToIso(pt pixel.Vec) pixel.Vec {
	return pixel.V((pt.X-pt.Y)*(tileSize/2), (pt.X+pt.Y)*(tileSize/4))
}

func main() {
	pixelgl.Run(run)
}

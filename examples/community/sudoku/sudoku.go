package main

import (
	"image"
	"os"
	"strconv"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	width = 900
)

var input bool
var x, y int

var board = [9][9]int{}
var puzzle = [9][9]int{
	{6, 7, 2, 3, 4, 1, 5, 8, 9},
	{5, 3, 4, 9, 6, 8, 1, 2, 7},
	{8, 9, 1, 7, 5, 2, 6, 3, 4},
	{3, 5, 6, 8, 2, 9, 4, 7, 1},
	{7, 2, 8, 4, 1, 5, 3, 9, 6},
	{4, 1, 9, 6, 7, 3, 8, 5, 2},
	{1, 8, 3, 2, 9, 6, 7, 4, 5},
	{9, 6, 7, 5, 3, 4, 2, 1, 8},
	{2, 4, 5, 1, 8, 7, 9, 6, 3},
}

var mask = [9][9]bool{
	{false, false, true, false, false, true, true, true, false},
	{false, false, true, true, true, false, false, false, true},
	{true, false, false, false, false, false, true, true, false},
	{false, true, true, false, true, false, true, true, false},
	{false, true, false, false, false, true, false, true, false},
	{false, true, true, false, true, false, true, true, false},
	{false, true, true, false, false, false, false, false, true},
	{true, false, false, false, true, true, true, false, false},
	{false, true, true, true, false, false, true, false, false},
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

func updateBoard(value int, imd *imdraw.IMDraw) {
	board[x][y] = value
	imd.Clear()
	input = false
}

func run() {
	// initialize window
	cfg := pixelgl.WindowConfig{
		Title:  "Sudoku",
		Bounds: pixel.R(0, 0, width, width),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// initialize imdraw
	imd := imdraw.New(nil)

	// initialize font
	ttf, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size: 64 * width / 900,
	})
	atlas := text.NewAtlas(face, text.ASCII)

	// initialize batch
	batch := pixel.NewBatch(&pixel.TrianglesData{}, atlas.Picture())

	// setup game
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if mask[i][j] {
				board[i][j] = puzzle[i][j]
			}
		}
	}

	for !win.Closed() {
		// win condition
		if board == puzzle {
			win.SetClosed(true)
		}

		// select a box
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			if input {
				imd.Clear()
			}
			pos := win.MousePosition()
			x = int(pos.X) / (width / 9)
			y = int(pos.Y) / (width / 9)

			if !mask[x][y] {
				imd.Color = colornames.Paleturquoise
				imd.Push(pixel.V(float64(x*width/9)+1, float64(y*width/9)+1),
					pixel.V(float64((x+1)*width/9)-1, float64((y+1)*width/9)-1))
				imd.Rectangle(0)
				input = true
			}
		}
		// act on user input
		if input && !mask[x][y] {
			if win.JustPressed(pixelgl.Key1) || win.JustPressed(pixelgl.KeyKP1) {
				updateBoard(1, imd)
			}
			if win.JustPressed(pixelgl.Key2) || win.JustPressed(pixelgl.KeyKP2) {
				updateBoard(2, imd)
			}
			if win.JustPressed(pixelgl.Key3) || win.JustPressed(pixelgl.KeyKP3) {
				updateBoard(3, imd)
			}
			if win.JustPressed(pixelgl.Key4) || win.JustPressed(pixelgl.KeyKP4) {
				updateBoard(4, imd)
			}
			if win.JustPressed(pixelgl.Key5) || win.JustPressed(pixelgl.KeyKP5) {
				updateBoard(5, imd)
			}
			if win.JustPressed(pixelgl.Key6) || win.JustPressed(pixelgl.KeyKP6) {
				updateBoard(6, imd)
			}
			if win.JustPressed(pixelgl.Key7) || win.JustPressed(pixelgl.KeyKP7) {
				updateBoard(7, imd)
			}
			if win.JustPressed(pixelgl.Key8) || win.JustPressed(pixelgl.KeyKP8) {
				updateBoard(8, imd)
			}
			if win.JustPressed(pixelgl.Key9) || win.JustPressed(pixelgl.KeyKP9) {
				updateBoard(9, imd)
			}
			if win.JustPressed(pixelgl.Key0) || win.JustPressed(pixelgl.KeyKP0) ||
				win.JustPressed(pixelgl.KeyBackspace) || win.JustPressed(pixelgl.KeySpace) {
				updateBoard(0, imd)
			}
		}

		// set up lines for drawing
		for i := 1; i < 9; i++ {
			imd.Color = colornames.Black
			if i%3 == 0 {
				imd.Push(pixel.V(float64(i*width/9), 0), pixel.V(float64(i*width/9), width))
				imd.Line(6)

				imd.Push(pixel.V(0, float64(i*width/9)), pixel.V(width, float64(i*width/9)))
				imd.Line(6)
			} else {
				imd.Push(pixel.V(float64(i*width/9), 0), pixel.V(float64(i*width/9), width))
				imd.Line(3)

				imd.Push(pixel.V(0, float64(i*width/9)), pixel.V(width, float64(i*width/9)))
				imd.Line(3)
			}
		}

		// set up numbers for drawing
		batch.Clear()
		for a, sa := range board {
			for b, sb := range sa {
				if sb != 0 {
					num := text.New(pixel.ZV, atlas)
					num.WriteString(strconv.Itoa(sb))
					num.DrawColorMask(batch,
						pixel.IM.
							Scaled(
								pixel.ZV, float64(width)/900).
							Moved(
								pixel.V((float64(a)+0.3)*width/9,
									(float64(b)+0.25)*width/9)),
						colornames.Black)
				}
			}
		}

		// draw the scene to the window
		win.Clear(colornames.Snow)
		batch.Draw(win)
		imd.Draw(win)

		// update the window
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

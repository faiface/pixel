package main

// Code based on the Recursive backtracker algorithm.
// https://en.wikipedia.org/wiki/Maze_generation_algorithm#Recursive_backtracker
// See https://youtu.be/HyK_Q5rrcr4 as an example
// YouTube example ported to Go for the Pixel library.

// Created by Stephen Chavez

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/examples/community/maze/stack"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"

	"github.com/pkg/profile"
	"golang.org/x/image/colornames"
)

var visitedColor = pixel.RGB(0.5, 0, 1).Mul(pixel.Alpha(0.35))
var hightlightColor = pixel.RGB(0.3, 0, 0).Mul(pixel.Alpha(0.45))
var debug = false

type cell struct {
	walls [4]bool // Wall order: top, right, bottom, left

	row     int
	col     int
	visited bool
}

func (c *cell) Draw(imd *imdraw.IMDraw, wallSize int) {
	drawCol := c.col * wallSize // x
	drawRow := c.row * wallSize // y

	imd.Color = colornames.White
	if c.walls[0] {
		// top line
		imd.Push(pixel.V(float64(drawCol), float64(drawRow)), pixel.V(float64(drawCol+wallSize), float64(drawRow)))
		imd.Line(3)
	}
	if c.walls[1] {
		// right Line
		imd.Push(pixel.V(float64(drawCol+wallSize), float64(drawRow)), pixel.V(float64(drawCol+wallSize), float64(drawRow+wallSize)))
		imd.Line(3)
	}
	if c.walls[2] {
		// bottom line
		imd.Push(pixel.V(float64(drawCol+wallSize), float64(drawRow+wallSize)), pixel.V(float64(drawCol), float64(drawRow+wallSize)))
		imd.Line(3)
	}
	if c.walls[3] {
		// left line
		imd.Push(pixel.V(float64(drawCol), float64(drawRow+wallSize)), pixel.V(float64(drawCol), float64(drawRow)))
		imd.Line(3)
	}
	imd.EndShape = imdraw.SharpEndShape

	if c.visited {
		imd.Color = visitedColor
		imd.Push(pixel.V(float64(drawCol), (float64(drawRow))), pixel.V(float64(drawCol+wallSize), float64(drawRow+wallSize)))
		imd.Rectangle(0)
	}
}

func (c *cell) GetNeighbors(grid []*cell, cols int, rows int) ([]*cell, error) {
	neighbors := []*cell{}
	j := c.row
	i := c.col

	top, _ := getCellAt(i, j-1, cols, rows, grid)
	right, _ := getCellAt(i+1, j, cols, rows, grid)
	bottom, _ := getCellAt(i, j+1, cols, rows, grid)
	left, _ := getCellAt(i-1, j, cols, rows, grid)

	if top != nil && !top.visited {
		neighbors = append(neighbors, top)
	}
	if right != nil && !right.visited {
		neighbors = append(neighbors, right)
	}
	if bottom != nil && !bottom.visited {
		neighbors = append(neighbors, bottom)
	}
	if left != nil && !left.visited {
		neighbors = append(neighbors, left)
	}

	if len(neighbors) == 0 {
		return nil, errors.New("We checked all cells...")
	}
	return neighbors, nil
}

func (c *cell) GetRandomNeighbor(grid []*cell, cols int, rows int) (*cell, error) {
	neighbors, err := c.GetNeighbors(grid, cols, rows)
	if neighbors == nil {
		return nil, err
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(neighbors))))
	if err != nil {
		panic(err)
	}
	randomIndex := nBig.Int64()
	return neighbors[randomIndex], nil
}

func (c *cell) hightlight(imd *imdraw.IMDraw, wallSize int) {
	x := c.col * wallSize
	y := c.row * wallSize

	imd.Color = hightlightColor
	imd.Push(pixel.V(float64(x), float64(y)), pixel.V(float64(x+wallSize), float64(y+wallSize)))
	imd.Rectangle(0)
}

func newCell(col int, row int) *cell {
	newCell := new(cell)
	newCell.row = row
	newCell.col = col

	for i := range newCell.walls {
		newCell.walls[i] = true
	}
	return newCell
}

// Creates the inital maze slice for use.
func initGrid(cols, rows int) []*cell {
	grid := []*cell{}
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			newCell := newCell(i, j)
			grid = append(grid, newCell)
		}
	}
	return grid
}

func setupMaze(cols, rows int) ([]*cell, *stack.Stack, *cell) {
	// Make an empty grid
	grid := initGrid(cols, rows)
	backTrackStack := stack.NewStack(len(grid))
	currentCell := grid[0]

	return grid, backTrackStack, currentCell
}

func cellIndex(i, j, cols, rows int) int {
	if i < 0 || j < 0 || i > cols-1 || j > rows-1 {
		return -1
	}
	return i + j*cols
}

func getCellAt(i int, j int, cols int, rows int, grid []*cell) (*cell, error) {
	possibleIndex := cellIndex(i, j, cols, rows)

	if possibleIndex == -1 {
		return nil, fmt.Errorf("cellIndex: CellIndex is a negative number %d", possibleIndex)
	}
	return grid[possibleIndex], nil
}

func removeWalls(a *cell, b *cell) {
	x := a.col - b.col

	if x == 1 {
		a.walls[3] = false
		b.walls[1] = false
	} else if x == -1 {
		a.walls[1] = false
		b.walls[3] = false
	}

	y := a.row - b.row

	if y == 1 {
		a.walls[0] = false
		b.walls[2] = false
	} else if y == -1 {
		a.walls[2] = false
		b.walls[0] = false
	}
}

func run() {
	// unsiged integers, because easier parsing error checks.
	// We must convert these to intergers, as done below...
	uScreenWidth, uScreenHeight, uWallSize := parseArgs()

	var (
		// In pixels
		// Defualt is 800x800x40 = 20x20 wallgrid
		screenWidth  = int(uScreenWidth)
		screenHeight = int(uScreenHeight)
		wallSize     = int(uWallSize)

		frames = 0
		second = time.Tick(time.Second)

		grid           = []*cell{}
		cols           = screenWidth / wallSize
		rows           = screenHeight / wallSize
		currentCell    = new(cell)
		backTrackStack = stack.NewStack(1)
	)

	// Set game FPS manually
	fps := time.Tick(time.Second / 60)

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks! - Maze example",
		Bounds: pixel.R(0, 0, float64(screenHeight), float64(screenWidth)),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	grid, backTrackStack, currentCell = setupMaze(cols, rows)

	gridIMDraw := imdraw.New(nil)

	for !win.Closed() {
		if win.JustReleased(pixelgl.KeyR) {
			fmt.Println("R pressed")
			grid, backTrackStack, currentCell = setupMaze(cols, rows)
		}

		win.Clear(colornames.Gray)
		gridIMDraw.Clear()

		for i := range grid {
			grid[i].Draw(gridIMDraw, wallSize)
		}

		// step 1
		// Make the initial cell the current cell and mark it as visited
		currentCell.visited = true
		currentCell.hightlight(gridIMDraw, wallSize)

		// step 2.1
		// If the current cell has any neighbours which have not been visited
		// Choose a random unvisited cell
		nextCell, _ := currentCell.GetRandomNeighbor(grid, cols, rows)
		if nextCell != nil && !nextCell.visited {
			// step 2.2
			// Push the current cell to the stack
			backTrackStack.Push(currentCell)

			// step 2.3
			// Remove the wall between the current cell and the chosen cell

			removeWalls(currentCell, nextCell)

			// step 2.4
			// Make the chosen cell the current cell and mark it as visited
			nextCell.visited = true
			currentCell = nextCell
		} else if backTrackStack.Len() > 0 {
			currentCell = backTrackStack.Pop().(*cell)
		}

		gridIMDraw.Draw(win)
		win.Update()
		<-fps
		updateFPSDisplay(win, &cfg, &frames, grid, second)
	}
}

// Parses the maze arguments, all of them are optional.
// Uses uint as implicit error checking :)
func parseArgs() (uint, uint, uint) {
	var mazeWidthPtr = flag.Uint("w", 800, "w sets the maze's width in pixels.")
	var mazeHeightPtr = flag.Uint("h", 800, "h sets the maze's height in pixels.")
	var wallSizePtr = flag.Uint("c", 40, "c sets the maze cell's size in pixels.")

	flag.Parse()

	// If these aren't default values AND if they're not the same values.
	// We should warn the user that the maze will look funny.
	if *mazeWidthPtr != 800 || *mazeHeightPtr != 800 {
		if *mazeWidthPtr != *mazeHeightPtr {
			fmt.Printf("WARNING: maze width: %d and maze height: %d don't match. \n", *mazeWidthPtr, *mazeHeightPtr)
			fmt.Println("Maze will look funny because the maze size is bond to the window size!")
		}
	}

	return *mazeWidthPtr, *mazeHeightPtr, *wallSizePtr
}

func updateFPSDisplay(win *pixelgl.Window, cfg *pixelgl.WindowConfig, frames *int, grid []*cell, second <-chan time.Time) {
	*frames++
	select {
	case <-second:
		win.SetTitle(fmt.Sprintf("%s | FPS: %d with %d Cells", cfg.Title, *frames, len(grid)))
		*frames = 0
	default:
	}

}

func main() {
	if debug {
		defer profile.Start().Stop()
	}
	pixelgl.Run(run)
}

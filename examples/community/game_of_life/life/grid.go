package life

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

// Shamelessly taken/inspired by https://golang.org/doc/play/life.go
// Grid is the structure in which the cellular automota live
type Grid struct {
	h        int
	cellSize int
	Cells    [][]bool
}

// NewGrid constructs a new Grid
func NewGrid(h, size int) *Grid {
	cells := make([][]bool, h)
	for i := 0; i < h; i++ {
		cells[i] = make([]bool, h)
	}
	return &Grid{h: h, cellSize: size, Cells: cells}
}

// Alive returns whether the specified position is alive
func (g *Grid) Alive(x, y int) bool {
	x += g.h
	x %= g.h
	y += g.h
	y %= g.h
	return g.Cells[y][x]
}

// Set sets the state of a specific location
func (g *Grid) Set(x, y int, state bool) {
	g.Cells[y][x] = state
}

// Draw draws the grid
func (g *Grid) Draw(imd *imdraw.IMDraw) {
	for i := 0; i < g.h; i++ {
		for j := 0; j < g.h; j++ {
			if g.Alive(i, j) {
				imd.Color = colornames.Black
			} else {
				imd.Color = colornames.White
			}
			imd.Push(
				pixel.V(float64(i*g.cellSize), float64(j*g.cellSize)),
				pixel.V(float64(i*g.cellSize+g.cellSize), float64(j*g.cellSize+g.cellSize)),
			)
			imd.Rectangle(0)
		}
	}
}

// Next returns the next state
func (g *Grid) Next(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && g.Alive(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && g.Alive(x, y)
}

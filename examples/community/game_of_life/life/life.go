// Package life manages the "game" state
// Shamelessly taken from https://golang.org/doc/play/life.go
package life

import "math/rand"

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	A, b *Grid
	h    int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(h, size int) *Life {
	a := NewGrid(h, size)
	for i := 0; i < (h * h / 2); i++ {
		a.Set(rand.Intn(h), rand.Intn(h), true)
	}
	return &Life{
		A: a, b: NewGrid(h, size),
		h: h,
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.h; x++ {
			l.b.Set(x, y, l.A.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.A, l.b = l.b, l.A
}

package main

import (
	"image/color"
	"sync"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

// Scalpel is a surgical knife for dissection and surgery. And for debugging purposes sometimes.
type Scalpel struct {
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
}

// Draw guarantees the thread safety, though it's not a necessary condition.
// It is quite dangerous to access this struct's member (imdraw) directly from outside these methods.
func (s *Scalpel) Draw(t pixel.Target) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.imd == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	s.imd.Draw(t)
}

// Update dissects a ladder. The anatomy of a ladder is drawn on an imdraw.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (s *Scalpel) Update(l Ladder) {
	ptsEnd := l.PtsAtLevelOfPrizes()

	// lock shared imdraw access
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// imdraw (a state machine)
	if s.imd == nil { // lazy creation
		s.imd = imdraw.New(nil)
	}
	imd := s.imd
	imd.Clear()

	// draw bounds
	imd.Color = colornames.Black
	imd.EndShape = imdraw.NoEndShape
	imd.Push(gg.VerticesOfRect(l.bound)...)
	imd.Polygon(4)

	// draw end points
	imd.Color = colornames.Blueviolet
	imd.Push(ptsEnd...)
	imd.Circle(10, 0)

	// draw grid
	imd.Color = colornames.Red
	for nrow, row := range l.grid {
		for ncol := range row {
			imd.Push(l.grid[nrow][ncol])
		}
	}
	imd.Circle(5, 0)
}

// -------------------------------------------------------------------------

// UpdateDrawProjekt has nothing to do with scalpel.
func UpdateDrawProjekt(t pixel.Target, rekt pixel.Rect, color color.Color, matrix pixel.Matrix) {
	imd := imdraw.New(nil)

	imd.Color = color
	imd.EndShape = imdraw.NoEndShape
	vertices := gg.VerticesOfRect(rekt)
	for i, v := range vertices {
		vertices[i] = matrix.Project(v)
	}
	imd.Push(vertices...)
	imd.Polygon(10)

	imd.Draw(t)
}

// UpdateDrawUnprojekt has nothing to do with scalpel.
func UpdateDrawUnprojekt(t pixel.Target, rekt pixel.Rect, color color.Color, matrix pixel.Matrix) {
	imd := imdraw.New(nil)

	imd.Color = color
	imd.EndShape = imdraw.NoEndShape
	vertices := gg.VerticesOfRect(rekt)
	for i, v := range vertices {
		vertices[i] = matrix.Unproject(v)
	}
	imd.Push(vertices...)
	imd.Polygon(10)

	imd.Draw(t)
}

// UpdateDrawUnprojekt2 has nothing to do with scalpel.
func UpdateDrawUnprojekt2(t pixel.Target, rekt pixel.Rect, color color.Color, camera gg.Camera) {
	imd := imdraw.New(nil)

	imd.Color = color
	imd.EndShape = imdraw.NoEndShape
	vertices := gg.VerticesOfRect(rekt)
	for i, v := range vertices {
		vertices[i] = camera.Unproject(v)
	}
	imd.Push(vertices...)
	imd.Polygon(10)

	imd.Draw(t)
}

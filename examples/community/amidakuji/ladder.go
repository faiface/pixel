package main

import (
	"math/rand"
	"sync"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

// Ladder is an imdraw that does not animate at all,
// hence does not need to be modified every frame.
// Something kinda static and bone-like.
type Ladder struct {
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
	//
	bound         pixel.Rect
	grid          [][]pixel.Vec
	bridges       [][]bool
	nParticipants int
	nLevel        int
	paddingTop    float64
	paddingRight  float64
	paddingBottom float64
	paddingLeft   float64
	colors        []pixel.RGBA
}

// NewLadder is a constructor.
func NewLadder(_nParticipants, _nLevel int,
	_width, _height,
	_paddingTop, _paddingRight,
	_paddingBottom, _paddingLeft float64) *Ladder {

	// get random colors
	colors := []pixel.RGBA{}
	for i := 0; i < _nParticipants; i++ {
		colors = append(colors, gg.RandomNiceColor())
	}

	// new grid
	newGrid := func(nRow, nCol int) [][]pixel.Vec {
		arr := make([][]pixel.Vec, nRow)
		for i := range arr {
			arr[i] = make([]pixel.Vec, nCol)
		}
		// log.Println(arr) //
		return arr
	}

	// new bridges
	newBridges := func(nParticipants, nLevel int) [][]bool {
		nRow := nParticipants - 1
		nCol := nLevel
		arr := make([][]bool, nRow)
		for i := range arr {
			arr[i] = make([]bool, nCol)
		}
		return arr
	}

	// init ladder
	l := Ladder{
		imd:           imdraw.New(nil),
		bound:         pixel.R(0, 0, _width, _height),
		grid:          newGrid(_nParticipants, _nLevel),
		bridges:       newBridges(_nParticipants, _nLevel),
		nParticipants: _nParticipants,
		nLevel:        _nLevel,
		paddingTop:    _paddingTop,
		paddingBottom: _paddingBottom,
		paddingRight:  _paddingRight,
		paddingLeft:   _paddingLeft,
		colors:        colors,
	}

	// init grid
	updateGrid := func(l *Ladder) {
		for participant := range l.grid { // row
			for level := range l.grid[participant] { // col
				y := l.Height() - (float64(participant) * l.DistParticipant()) // reverse
				x := float64(level) * l.DistLevel()
				y -= l.paddingTop
				x += l.paddingLeft
				l.grid[participant][level] = pixel.V(x, y)
			}
		}
	} // Indices would not be aligned with the screen coordinates. (Reverse Y - Rows)
	updateGrid(&l)

	// log.Println(l.grid) //
	return &l
}

// -------------------------------------------------------------------------
// Important methods

// Draw guarantees the thread safety, though it's not a necessary condition.
// It is quite dangerous to access this struct's member (imdraw) directly from outside these methods.
func (l *Ladder) Draw(t pixel.Target) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.imd == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	l.imd.Draw(t)
}

// Update draws a ladder on an imdraw.
func (l *Ladder) Update() {
	ptsMovedAbout := func(sub pixel.Vec, pts ...pixel.Vec) (ptsMoved []pixel.Vec) {
		ptsMoved = make([]pixel.Vec, len(pts))
		copy(ptsMoved, pts)
		for i, vec := range ptsMoved {
			ptsMoved[i] = vec.Sub(sub)
		}
		// log.Println(ptsMoved) // debug
		// log.Println(pts)      // debug
		return ptsMoved
	}

	ptsStart := l.PtsAtLevelOfPicks()
	ptsEnd := l.PtsAtLevelOfPrizes()

	circleRadius := 20.0
	circlePts := ptsMovedAbout(pixel.V(circleRadius+10, 0), ptsStart...)

	// lock shared imdraw access
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// imdraw (a state machine)
	if l.imd == nil { // lazy creation
		l.imd = imdraw.New(nil)
	}
	imd := l.imd
	imd.Clear()

	// draw lanes
	imd.Color = colornames.White
	imd.EndShape = imdraw.RoundEndShape
	for i := range ptsStart {
		imd.Push(ptsStart[i], ptsEnd[i])
		imd.Line(13)
	}

	// draw bridges
	imd.Color = colornames.White
	imd.EndShape = imdraw.RoundEndShape
	for nrow, row := range l.bridges {
		for ncol, e := range row {
			if e {
				imd.Push(l.grid[nrow][ncol], l.grid[nrow+1][ncol])
				imd.Line(13)
			}
		}
	}

	// draw start points
	imd.EndShape = imdraw.RoundEndShape
	for i := range ptsStart {
		imd.Color = l.colors[i]
		imd.Push(circlePts[i])
	}
	imd.Circle(circleRadius, 0)
}

// -------------------------------------------------------------------------
// Read only methods

// Height returns the height of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) Height() float64 {
	return l.bound.H()
}

// Width returns the width of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) Width() float64 {
	return l.bound.W()
}

// DistLevel returns the distance between each level.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) DistLevel() float64 {
	return (l.Width() - (l.paddingLeft + l.paddingRight)) / float64(l.nLevel-1)
}

// DistParticipant returns the distance between each lane.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) DistParticipant() float64 {
	return (l.Height() - (l.paddingTop + l.paddingBottom)) / float64(l.nParticipants-1)
}

// PtsAtLevelOfPicks returns all starting points of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) PtsAtLevelOfPicks() (ret []pixel.Vec) {
	const levelOfDraw int = 0 // where it starts
	ret = make([]pixel.Vec, l.nParticipants, l.nParticipants)

	for participant := range l.grid { // row
		ret[participant] = l.grid[participant][levelOfDraw]
	}

	// log.Println(len(ret), ret) //
	return ret //[:l.nParticipants] //
}

// PtAtLevelOfPicks returns a starting point of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) PtAtLevelOfPicks(participant int) pixel.Vec {
	const levelOfDraw int = 0 // where it starts
	return l.grid[participant][levelOfDraw]
}

// PtsAtLevelOfPrizes returns all end points of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) PtsAtLevelOfPrizes() (ret []pixel.Vec) {
	levelOfPrize := l.nLevel - 1 // where it ends
	ret = make([]pixel.Vec, l.nParticipants, l.nParticipants)

	for participant := range l.grid { // row
		ret[participant] = l.grid[participant][levelOfPrize]
	}

	// log.Println(len(ret), ret) //
	return ret //[:l.nParticipants] //
}

// PtAtLevelOfPrizes returns an end point of a ladder.
// A non-ptr Ladder as a read only argument passes lock by value within itself but that seems totally fine.
func (l Ladder) PtAtLevelOfPrizes(participant int) pixel.Vec {
	levelOfPrize := l.nLevel - 1 // where it ends
	return l.grid[participant][levelOfPrize]
}

// -------------------------------------------------------------------
// Methods that write to itself

// ClearBridges of a ladder.
// Only values are changed, not the pointers.
func (l *Ladder) ClearBridges() {
	for _, row := range l.bridges {
		for i := range row {
			row[i] = false
		}
	}
}

// GenerateRandomBridges of an approximate amount.
// Only values are changed, not the pointers.
func (l *Ladder) GenerateRandomBridges(amountApprox int) {
	pickOneBridgeInRandom := func(l *Ladder) {
		nRow := len(l.bridges)
		nCol := l.nLevel
		row := rand.Intn(int(nRow)) // participant
		col := rand.Intn(int(nCol)) // level
		// check right
		isOkRight := func(rowRight, col int) bool {
			includeLowerBound := func(rowRight int) bool {
				return rowRight >= 0
			}
			if !includeLowerBound(rowRight) { // out of bound
				return true
			}
			if !l.bridges[rowRight][col] {
				return true
			}
			return false
		}
		// check left
		isOkLeft := func(rowLeft, col int) bool {
			excludeUpperBound := func(rowLeft int) bool {
				return rowLeft < len(l.bridges)
			}
			if !excludeUpperBound(rowLeft) { // out of bound
				return true
			}
			if !l.bridges[rowLeft][col] {
				return true
			}
			return false
		}
		rowRight := row - 1
		rowLeft := row + 1
		if isOkRight(rowRight, col) && isOkLeft(rowLeft, col) {
			l.bridges[row][col] = true
		}
	} // func

	// repeat
	for i := 0; i < amountApprox; i++ {
		pickOneBridgeInRandom(l)
	}
} // method

// RegenerateRandomBridges clears out all bridges and then GenerateRandomBridges() an approximate amount.
// Only values are changed, not the pointers.
func (l *Ladder) RegenerateRandomBridges(amountApprox int) {
	l.ClearBridges()
	l.GenerateRandomBridges(amountApprox)
}

// RegenerateRandomColors sets all colors of a ladder random for each.
// Only values are changed, not the pointers.
func (l *Ladder) RegenerateRandomColors() {
	for i := range l.colors {
		l.colors[i] = gg.RandomNiceColor()
	}
}

// Reset all bridges and colors.
// Only values are changed, not the pointers.
func (l *Ladder) Reset() {
	aboutTwo := l.nParticipants * (l.nLevel - 1) * 2
	aboutOne := l.nParticipants * (l.nLevel - 1)
	aboutHalf := (l.nParticipants * (l.nLevel - 1)) / 2
	//
	var pick [4]int
	pick[0] = aboutTwo
	pick[1] = aboutOne
	pick[2] = aboutHalf
	i := rand.Intn(3)
	// log.Println(i) //
	//
	l.RegenerateRandomBridges(pick[i])
	l.RegenerateRandomColors()
}

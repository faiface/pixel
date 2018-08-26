package main

import (
	"math"
	"sync"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

// Path is for animating a path to the prize in a ladder.
type Path struct {
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
	//
	roads  []pixel.Vec // A list of vectors - each vector for a position where a road starts.
	prize  *int
	tip    *pixel.Vec
	tipDir pixel.Vec
	iroad  int
	//
	watchAnim          gg.DtWatch // When it started to animate.
	timeLimitAnimInSec float64
	animateInTime      bool
	isAnimating        bool

	// -----------------------------------------------------------
	// exported callbacks(listeners) regarding animation

	// callback on reaching the prize level of a ladder.
	OnFinishedAnimation func()

	// callback when the animating 'tip' passes a point of a road.
	// pt: a point(road) just passed.
	// dir: ...
	// dir is a normalized vector. (pixel.ZV) is passed if the direction can't be found.
	// dir can be different depending on how fast this Path is updated.
	OnPassedEachPoint func(pt pixel.Vec, dir pixel.Vec)
}

// NewPath is a contructor.
func NewPath(_roads []pixel.Vec, _prize *int) *Path {
	newTip := func() *pixel.Vec {
		if _roads != nil {
			if len(_roads) > 0 {
				v := _roads[0]
				return &v
			}
		}
		return nil
	}
	return &Path{
		roads:  _roads,
		prize:  _prize,
		tip:    newTip(),
		tipDir: pixel.ZV,
	}
}

// NewPathEmpty is a contructor.
func NewPathEmpty() *Path {
	return &Path{}
}

// -------------------------------------------------------------------------
// Important methods

// Draw guarantees the thread safety, though it's not a necessary condition.
// It is quite dangerous to access this struct's member (imdraw) directly from outside these methods.
func (path *Path) Draw(t pixel.Target) {
	path.mutex.Lock()
	defer path.mutex.Unlock()

	if path.imd == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	path.imd.Draw(t)
}

// Update animates a path. A path is drawn on an imdraw.
func (path *Path) Update(color pixel.RGBA) {
	var (
		iroad = len(path.roads) - 1
		from  = path.roads[len(path.roads)-1]
		to    = path.roads[len(path.roads)-1]
		dir   = pixel.ZV
	)

	if path.isAnimating {
		// get where it is abstract // get a scalar
		dt := path.watchAnim.DtSinceStart()
		const distPerSec = 500
		scalarProgress := dt * distPerSec
		// log.Println(dt, path.Len(), scalarProgress) //

		if path.animateInTime { // overwrite scalarProgress
			fromLot := pixel.V(0, 0)
			toPrize := pixel.V(path.Len(), 0)
			percentagePointPerSec := 1 / path.timeLimitAnimInSec
			scalarProgress = pixel.Lerp(fromLot, toPrize, dt*percentagePointPerSec).X
			// log.Println(dt, path.Len(), scalarProgress, dt*percentagePointPerSec) //
		}

		// get where it is concrete // turn a scalar into a set of vectors
		iroad, from, to, dir = path.FindRoadByDist(scalarProgress)
		if iroad > path.iroad {
			if path.OnPassedEachPoint != nil {
				go path.OnPassedEachPoint(from, dir)
			}
		}
		path.iroad = iroad
		// log.Println(iroad, len(path.roads), iroad == len(path.roads)-1, path.isAnimating) //
		if iroad >= len(path.roads)-1 { // the end
			path.isAnimating = false
			// log.Println(path.Len(), dt) //
			if path.OnFinishedAnimation != nil {
				go path.OnFinishedAnimation()
			} // callback
		}
	}

	// lock before imdraw update
	path.mutex.Lock()
	defer path.mutex.Unlock()

	// imdraw (a state machine)
	if path.imd == nil { // lazy creation
		path.imd = imdraw.New(nil)
	}
	imd := path.imd
	imd.Clear()

	// draw path
	imd.Color = color
	imd.EndShape = imdraw.RoundEndShape
	for i := 0; i < iroad; i++ {
		imd.Push(path.roads[i], path.roads[i+1])
		imd.Line(9)
	}
	imd.Push(from, to)
	imd.Line(9)

	// save where the tip is
	path.tip = &to
}

// -------------------------------------------------------------------------
// Read only methods

// IsAnimating determines whether this Path is about to be updated or not.
// Pass lock by value warning from (path Path) should be ignored,
// because a Path here is just passed as a read only argument.
func (path Path) IsAnimating() bool {
	return path.isAnimating
}

// GetPrize is just an average getter.
// It returns -1 if the receiver is not initialized with that member(prize).
// Pass lock by value warning from (path Path) should be ignored,
// because a Path here is just passed as a read only argument.
func (path Path) GetPrize() int {
	if path.prize == nil {
		return -1
	}
	return *path.prize
}

// PosTip returns a vector that tells you how far the animating path currently has reached.
// A non-ptr Path as a read only argument passes lock by value within itself but that seems totally fine.
func (path Path) PosTip() (v pixel.Vec) {
	return *path.tip
}

// Len returns the total length of all roads.
// A non-ptr Path as a read only argument passes lock by value within itself but that seems totally fine.
func (path Path) Len() (sum float64) {
	for i := 0; i < len(path.roads)-1; i++ {
		sum += math.Abs(path.roads[i].Sub(path.roads[i+1]).Len())
	}
	return
}

// FindRoadByDist converts a scalar into a set of vectors.
// A non-ptr Path as a read only argument passes lock by value within itself but that seems totally fine.
//
// Returns
// iroad: The index of a road found.
// road: The vector representation of a road found. A road is a line from pt A to B, and that vector points to where pt A is.
// pos: A position(point) found which is in the middle of that found road(line).
// dirVecNormalized: A direction as a normalized vector. This vector always has a length of 1.
func (path Path) FindRoadByDist(distProgress float64) (iroad int, road pixel.Vec, pos pixel.Vec, dirVecNormalized pixel.Vec) {
	lengthOfTraveledRoads := float64(0.0)
	for iroad = 0; iroad < len(path.roads)-1; iroad++ {
		var lengthOfThisRoad float64
		iroadNext := iroad + 1
		lengthOfThisRoad = math.Abs(path.roads[iroad].Sub(path.roads[iroadNext]).Len())
		lengthOfTraveledRoads += lengthOfThisRoad
		// For loop breaker: distProgress is somewhere between the total length of a path.
		if lengthOfTraveledRoads > distProgress {
			scalar := lengthOfThisRoad - (lengthOfTraveledRoads - distProgress)
			if path.roads[iroad].Y == path.roads[iroadNext].Y &&
				path.roads[iroad].X < path.roads[iroadNext].X { // to the bottom (east)
				pos = path.roads[iroad]
				pos.X += scalar
				dirVecNormalized = pixel.V(1, 0)
			} else if path.roads[iroad].X == path.roads[iroadNext].X &&
				path.roads[iroad].Y > path.roads[iroadNext].Y { // to the left (south)
				pos = path.roads[iroad]
				pos.Y -= scalar
				dirVecNormalized = pixel.V(0, -1)
			} else if path.roads[iroad].X == path.roads[iroadNext].X &&
				path.roads[iroad].Y < path.roads[iroadNext].Y { // to the right (north)
				pos = path.roads[iroad]
				pos.Y += scalar
				dirVecNormalized = pixel.V(0, 1)
			} else if path.roads[iroad].Y == path.roads[iroadNext].Y &&
				path.roads[iroad].X > path.roads[iroadNext].X { // to the top (west)
				// Placed at the end of an elif statement,
				// since this case is of no possibility unless the path finding is going reverse.
				pos = path.roads[iroad]
				pos.X -= scalar
				dirVecNormalized = pixel.V(-1, 0)
			} else {
				panic("unhandled exception: it may be a diagonal bridge")
			} // elif
			return iroad, path.roads[iroad], pos, dirVecNormalized
		} // if - for loop breaker
	} // for

	// coming down to here means that the case is (road == pos)
	from := iroad - 1
	to := iroad
	if iroad == 0 {
		from = iroad
		to = iroad + 1
	}
	if from < 0 || to >= len(path.roads) {
		dirVecNormalized = pixel.ZV
	} else {
		dirVecNormalized = gg.Direction(path.roads[from], path.roads[to])
	}
	return iroad, path.roads[iroad], path.roads[iroad], dirVecNormalized
}

// -------------------------------------------------------------------------
// Methods that write to itself

// Animate a path.
func (path *Path) Animate() {
	path.watchAnim.Start()
	path.animateInTime = false
	path.isAnimating = true
}

// AnimateInTime animates a path in given time.
func (path *Path) AnimateInTime(sec float64) {
	path.watchAnim.Start()
	path.timeLimitAnimInSec = sec
	path.animateInTime = true
	path.isAnimating = true
}

// Pause a path's clock.
func (path *Path) Pause() {
	if path.watchAnim.IsStarted() {
		path.watchAnim.Dt()
	}
}

// Resume after pause.
func (path *Path) Resume() {
	if path.watchAnim.IsStarted() {
		started := path.watchAnim.GetTimeStarted()
		dtPause := path.watchAnim.DtNano()
		path.watchAnim.SetTimeStarted(started.Add(dtPause))
		// log.Println(dtPause, started, path.watchAnim.GetTimeStarted()) //
	}
}

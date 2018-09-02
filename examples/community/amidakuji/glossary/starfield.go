package glossary

import (
	"image/color"
	"math/rand"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

// -------------------------------------------------------------------------
// Reusable modified starfiled
// - Original: "github.com/faiface/pixel/examples/community/starfield"
// - Encapsulated by nanitefactory

// -------------------------------------------------------------------------
// Galaxy

type star struct {
	Pos pixel.Vec  // x, y
	Z   float64    // z
	P   float64    // prev z
	C   color.RGBA // color
}

// Galaxy is an imd of stars.
type Galaxy struct {
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
	//
	width  float64
	height float64
	speed  float64
	stars  [1024]*star
}

// NewGalaxy is a constructor.
func NewGalaxy(_width, _height, _speed float64) *Galaxy {
	return &Galaxy{
		width:  _width,
		height: _height,
		speed:  _speed,
	}
}

// Speed is a getter.
// Pass lock by value warning from (galaxy Galaxy) should be ignored,
// because a galaxy here is just passed as a read only argument.
func (galaxy Galaxy) Speed() float64 {
	return galaxy.speed
}

// SetSpeed is a setter.
func (galaxy *Galaxy) SetSpeed(_speed float64) {
	galaxy.speed = _speed
}

// Draw guarantees the thread safety, though it's not a necessary condition.
// It is quite dangerous to access this struct's member (imdraw) directly from outside these methods.
func (galaxy *Galaxy) Draw(t pixel.Target) {
	galaxy.mutex.Lock()
	defer galaxy.mutex.Unlock()

	if galaxy.imd == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	galaxy.imd.Draw(t)
}

// Update animates a galaxy.
func (galaxy *Galaxy) Update(dt float64) {
	// random()
	random := func(min, max float64) float64 {
		return rand.Float64()*(max-min) + min
	}

	// newStar()
	newStar := func() *star {
		starColors := []color.RGBA{
			color.RGBA{157, 180, 255, 255},
			color.RGBA{162, 185, 255, 255},
			color.RGBA{167, 188, 255, 255},
			color.RGBA{170, 191, 255, 255},
			color.RGBA{175, 195, 255, 255},
			color.RGBA{186, 204, 255, 255},
			color.RGBA{192, 209, 255, 255},
			color.RGBA{202, 216, 255, 255},
			color.RGBA{228, 232, 255, 255},
			color.RGBA{237, 238, 255, 255},
			color.RGBA{251, 248, 255, 255},
			color.RGBA{255, 249, 249, 255},
			color.RGBA{255, 245, 236, 255},
			color.RGBA{255, 244, 232, 255},
			color.RGBA{255, 241, 223, 255},
			color.RGBA{255, 235, 209, 255},
			color.RGBA{255, 215, 174, 255},
			color.RGBA{255, 198, 144, 255},
			color.RGBA{255, 190, 127, 255},
			color.RGBA{255, 187, 123, 255},
			color.RGBA{255, 187, 123, 255},
		} // Colors based on stellar types listed at // http://www.vendian.org/mncharity/dir3/starcolor/
		return &star{
			Pos: pixel.V(random(-galaxy.width, galaxy.width), random(-galaxy.height, galaxy.height)),
			Z:   random(0, galaxy.width),
			P:   0,
			C:   starColors[rand.Intn(len(starColors))],
		}
	}

	// lock before imdraw update
	galaxy.mutex.Lock()
	defer galaxy.mutex.Unlock()

	// imdraw (a state machine)
	if galaxy.imd == nil { // lazy creation
		galaxy.imd = imdraw.New(nil)
		galaxy.imd.SetMatrix(pixel.IM.Moved(pixel.V(galaxy.width/2, galaxy.height/2)))
	}
	imd := galaxy.imd
	imd.Clear()
	imd.Precision = 7

	// now update all stars in this galaxy
	for i, s := range galaxy.stars {
		if s == nil {
			galaxy.stars[i] = newStar()
			s = galaxy.stars[i]
		}

		scale := func(unscaledNum, min, max, minAllowed, maxAllowed float64) float64 {
			return (maxAllowed-minAllowed)*(unscaledNum-min)/(max-min) + minAllowed
		}

		s.P = s.Z
		s.Z -= dt * galaxy.speed

		if s.Z < 0 {
			s.Pos.X = random(-galaxy.width, galaxy.width)
			s.Pos.Y = random(-galaxy.height, galaxy.height)
			s.Z = galaxy.width
			s.P = s.Z
		}

		p := pixel.V(
			scale(s.Pos.X/s.Z, 0, 1, 0, galaxy.width),
			scale(s.Pos.Y/s.Z, 0, 1, 0, galaxy.height),
		)

		o := pixel.V(
			scale(s.Pos.X/s.P, 0, 1, 0, galaxy.width),
			scale(s.Pos.Y/s.P, 0, 1, 0, galaxy.height),
		)

		r := scale(s.Z, 0, galaxy.width, 11, 0)

		galaxy.imd.Color = s.C
		if p.Sub(o).Len() > 6 {
			galaxy.imd.Push(p, o)
			galaxy.imd.Line(r)
		}
		galaxy.imd.Push(p)
		galaxy.imd.Circle(r, 0)
	}
}

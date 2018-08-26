package glossary

import (
	"image/color"
	"math/rand"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

// -------------------------------------------------------------------------
// explosive.go
// - Original idea: "github.com/faiface/pixel/examples/community/bouncing"

// --------------------------------------------------------------------

// Explosions is an imdraw and a manager of all particles.
type Explosions struct {
	imd   *imdraw.IMDraw
	mutex sync.Mutex // It is unsafe to access any refd; ptrd object without a critical section.
	//
	*colorPicker
	width     float64
	height    float64
	particles []*particle
	precision int
}

// NewExplosions is a constructor.
// The 3rd argument colors can be nil. Then it will use its default value of a color set.
func NewExplosions(width, height float64, colors []color.Color, precision int) *Explosions {
	return &Explosions{
		nil, sync.Mutex{},
		newColorPicker(colors),
		width, height, nil,
		precision,
	}
}

// SetBound of particles. All particles bounce when they meet this bound.
func (e *Explosions) SetBound(width, height float64) {
	e.width = width
	e.height = height
}

// IsExploding determines whether this Explosions is about to be updated or not.
// Pass lock by value warning from (e Explosions) should be ignored,
// because an Explosions here is just passed as a read only argument.
func (e Explosions) IsExploding() bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.particles != nil
}

// Draw guarantees the thread safety, though it's not a necessary condition.
// It is quite dangerous to access this struct's member (imdraw) directly from outside these methods.
func (e *Explosions) Draw(t pixel.Target) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.imd == nil || len(e.particles) <= 0 { // isInvisible set to true.
		return // An empty image is drawn.
	}

	e.imd.Draw(t)
}

// Update animates an Explosions. An Explosions is drawn on an imdraw.
func (e *Explosions) Update(dt float64) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// physics
	aliveParticles := []*particle{}
	for _, particle := range e.particles {
		particle.update(dt, e.width, e.height)
		if particle.life > 0 {
			aliveParticles = append(aliveParticles, particle)
		}
	}
	e.particles = aliveParticles

	// imdraw (a state machine)
	if e.imd == nil { // lazy creation
		e.imd = imdraw.New(nil)
		e.imd.EndShape = imdraw.RoundEndShape
		e.imd.Precision = e.precision
	}
	imd := e.imd
	imd.Clear()

	// draw
	for _, particle := range e.particles {
		imd.Color = particle.color
		imd.Push(particle.pos)
		imd.Circle(16*particle.life, 0)
	}
}

// ExplodeAt generates an explosion at given point.
func (e *Explosions) ExplodeAt(pos, vel pixel.Vec) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.next()
	e.particles = append(e.particles,
		newParticleAt(pos, vel.Rotated(1).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(2).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(3).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(4).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(5).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(6).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(7).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(8).Scaled(rand.Float64()), e.here()),
		newParticleAt(pos, vel.Rotated(9).Scaled(rand.Float64()), e.here()),

		newParticleAt(pos, vel.Rotated(10).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(20).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(30).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(40).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(50).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(60).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(70).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(80).Scaled(rand.Float64()+1), e.here()),
		newParticleAt(pos, vel.Rotated(90).Scaled(rand.Float64()+1), e.here()),
	)
}

// --------------------------------------------------------------------

type particle struct {
	pos   pixel.Vec
	vel   pixel.Vec
	color color.RGBA
	life  float64
}

func newParticleAt(pos, vel pixel.Vec, color color.RGBA) *particle {
	color.A = 5
	return &particle{pos, vel, color, rand.Float64() * 1.5}
}

func (p *particle) update(dt, width, height float64) {
	p.pos = p.pos.Add(p.vel)
	p.life -= 3 * dt
	switch {
	case p.pos.Y < 0 || p.pos.Y >= height:
		p.vel.Y *= (-10 * dt)
	case p.pos.X < 0 || p.pos.X >= width:
		p.vel.X *= (-10 * dt)
	}
}

// --------------------------------------------------------------------

type colorPicker struct {
	colors []color.RGBA
	index  int
}

func newColorPicker(_colors []color.Color) *colorPicker {
	if _colors == nil {
		_colors = []color.Color{
			color.RGBA{190, 38, 51, 255},
			color.RGBA{224, 111, 139, 255},
			color.RGBA{73, 60, 43, 255},
			color.RGBA{164, 100, 34, 255},
			color.RGBA{235, 137, 49, 255},
			color.RGBA{247, 226, 107, 255},
			color.RGBA{47, 72, 78, 255},
			color.RGBA{68, 137, 26, 255},
			color.RGBA{163, 206, 39, 255},
			color.RGBA{0, 87, 132, 255},
			color.RGBA{49, 162, 242, 255},
			color.RGBA{178, 220, 239, 255},
		}
	}
	colors := []color.RGBA{}
	for _, v := range _colors {
		if c, ok := v.(color.RGBA); ok {
			colors = append(colors, c)
		}
	}
	return &colorPicker{colors, 0}
}

func (colorPicker *colorPicker) next() color.RGBA {
	if colorPicker.index++; colorPicker.index >= len(colorPicker.colors) {
		colorPicker.index = 0
	}
	return colorPicker.colors[colorPicker.index]
}

func (colorPicker *colorPicker) here() color.RGBA {
	return colorPicker.colors[colorPicker.index]
}

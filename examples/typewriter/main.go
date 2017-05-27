package main

import (
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"
	"unicode"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

func ttfFromBytesMust(b []byte, size float64) font.Face {
	ttf, err := truetype.Parse(b)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(ttf, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})
}

type typewriter struct {
	mu sync.Mutex

	regular *text.Text
	bold    *text.Text
	italic  *text.Text

	offset   pixel.Vec
	position pixel.Vec
	move     pixel.Vec
}

func newTypewriter(c color.Color, regular, bold, italic *text.Atlas) *typewriter {
	tw := &typewriter{
		regular: text.New(pixel.ZV, regular),
		bold:    text.New(pixel.ZV, bold),
		italic:  text.New(pixel.ZV, italic),
	}
	tw.regular.Color = c
	tw.bold.Color = c
	tw.italic.Color = c
	return tw
}

func (tw *typewriter) Ribbon(r rune) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	dice := rand.Intn(21)
	switch {
	case 0 <= dice && dice <= 18:
		tw.regular.WriteRune(r)
	case dice == 19:
		tw.bold.Dot = tw.regular.Dot
		tw.bold.WriteRune(r)
		tw.regular.Dot = tw.bold.Dot
	case dice == 20:
		tw.italic.Dot = tw.regular.Dot
		tw.italic.WriteRune(r)
		tw.regular.Dot = tw.italic.Dot
	}
}

func (tw *typewriter) Back() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.regular.Dot = tw.regular.Dot.Sub(pixel.V(tw.regular.Atlas().Glyph(' ').Advance, 0))
}

func (tw *typewriter) Offset(off pixel.Vec) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.offset = tw.offset.Add(off)
}

func (tw *typewriter) Position() pixel.Vec {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	return tw.position
}

func (tw *typewriter) Move(vel pixel.Vec) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.move = vel
}

func (tw *typewriter) Dot() pixel.Vec {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	return tw.regular.Dot
}

func (tw *typewriter) Update(dt float64) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.position = tw.position.Add(tw.move.Scaled(dt))
}

func (tw *typewriter) Draw(t pixel.Target, m pixel.Matrix) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	m = pixel.IM.Moved(tw.position.Add(tw.offset)).Chained(m)
	tw.regular.Draw(t, m)
	tw.bold.Draw(t, m)
	tw.italic.Draw(t, m)
}

func typeRune(tw *typewriter, r rune) {
	tw.Ribbon(r)
	if !unicode.IsSpace(r) {
		go shake(tw, 3, 17)
	}
}

func back(tw *typewriter) {
	tw.Back()
}

func shake(tw *typewriter, intensity, friction float64) {
	const (
		freq = 24
		dt   = 1.0 / freq
	)
	ticker := time.NewTicker(time.Second / freq)
	defer ticker.Stop()

	off := pixel.ZV

	for range ticker.C {
		tw.Offset(off.Scaled(-1))

		if intensity < 0.01*dt {
			break
		}

		off = pixel.V((rand.Float64()-0.5)*intensity*2, (rand.Float64()-0.5)*intensity*2)
		intensity -= friction * dt

		tw.Offset(off)
	}
}

func scroll(tw *typewriter, intensity, speedUp float64) {
	const (
		freq = 120
		dt   = 1.0 / freq
	)
	ticker := time.NewTicker(time.Second / freq)
	defer ticker.Stop()

	speed := 0.0

	for range ticker.C {
		if math.Abs(tw.Dot().Y+tw.Position().Y) < 0.01 {
			break
		}

		targetSpeed := -(tw.Dot().Y + tw.Position().Y) * intensity
		if speed < targetSpeed {
			speed += speedUp * dt
		} else {
			speed = targetSpeed
		}

		tw.Move(pixel.V(0, speed))
	}
}

type dotlight struct {
	tw           *typewriter
	color        color.Color
	radius       float64
	intensity    float64
	acceleration float64
	maxSpeed     float64

	pos pixel.Vec
	vel pixel.Vec

	imd *imdraw.IMDraw
}

func newDotlight(tw *typewriter, c color.Color, radius, intensity, acceleration, maxSpeed float64) *dotlight {
	return &dotlight{
		tw:           tw,
		color:        c,
		radius:       radius,
		intensity:    intensity,
		acceleration: acceleration,
		maxSpeed:     maxSpeed,
		pos:          tw.Dot(),
		vel:          pixel.ZV,
		imd:          imdraw.New(nil),
	}
}

func (dl *dotlight) Update(dt float64) {
	targetVel := dl.tw.Dot().Add(dl.tw.Position()).Sub(dl.pos).Scaled(dl.intensity)
	acc := targetVel.Sub(dl.vel).Scaled(dl.acceleration)
	dl.vel = dl.vel.Add(acc.Scaled(dt))
	if dl.vel.Len() > dl.maxSpeed {
		dl.vel = dl.vel.Unit().Scaled(dl.maxSpeed)
	}
	dl.pos = dl.pos.Add(dl.vel.Scaled(dt))
}

func (dl *dotlight) Draw(t pixel.Target, m pixel.Matrix) {
	dl.imd.Clear()
	dl.imd.SetMatrix(m)
	dl.imd.Color = dl.color
	dl.imd.Push(dl.pos)
	dl.imd.Color = pixel.Alpha(0)
	for i := 0.0; i <= 32; i++ {
		angle := i * 2 * math.Pi / 32
		dl.imd.Push(dl.pos.Add(pixel.V(dl.radius, 0).Rotated(angle)))
	}
	dl.imd.Polygon(0)
	dl.imd.Draw(t)
}

func run() {
	rand.Seed(time.Now().UnixNano())

	cfg := pixelgl.WindowConfig{
		Title:     "Typewriter",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)

	var (
		regular = text.NewAtlas(
			ttfFromBytesMust(goregular.TTF, 42),
			text.ASCII, text.RangeTable(unicode.Latin),
		)
		bold = text.NewAtlas(
			ttfFromBytesMust(gobold.TTF, 42),
			text.ASCII, text.RangeTable(unicode.Latin),
		)
		italic = text.NewAtlas(
			ttfFromBytesMust(goitalic.TTF, 42),
			text.ASCII, text.RangeTable(unicode.Latin),
		)

		bgColor = color.RGBA{
			R: 241,
			G: 241,
			B: 212,
			A: 255,
		}
		fgColor = color.RGBA{
			R: 0,
			G: 15,
			B: 85,
			A: 255,
		}

		tw = newTypewriter(pixel.ToRGBA(fgColor).Scaled(0.9), regular, bold, italic)
		dl = newDotlight(tw, colornames.Red, 6, 30, 20, 1600)
	)

	fps := time.Tick(time.Second / 120)
	last := time.Now()
	for !win.Closed() {
		for _, r := range win.Typed() {
			go typeRune(tw, r)
		}
		if win.JustPressed(pixelgl.KeyTab) || win.Repeated(pixelgl.KeyTab) {
			go typeRune(tw, '\t')
		}
		if win.JustPressed(pixelgl.KeyEnter) || win.Repeated(pixelgl.KeyEnter) {
			go typeRune(tw, '\n')
			go scroll(tw, 20, 6400)
		}
		if win.JustPressed(pixelgl.KeyBackspace) || win.Repeated(pixelgl.KeyBackspace) {
			go back(tw)
		}

		dt := time.Since(last).Seconds()
		last = time.Now()

		tw.Update(dt)
		dl.Update(dt)

		win.Clear(bgColor)

		m := pixel.IM.Moved(pixel.V(32, 32))
		tw.Draw(win, m)
		dl.Draw(win, m)

		win.Update()
		<-fps
	}
}

func main() {
	pixelgl.Run(run)
}

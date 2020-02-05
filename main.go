package main

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Relay",
		Bounds: pixel.R(0, 0, 2048, 512),
		VSync:  true,
	}

	w, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	s := newState()

	start := time.Now()

	for !w.Closed() {
		w.Clear(colornames.Peru)
		s = updateState(s)
		render(s, w, time.Since(start))
		w.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

func render(s state, w *pixelgl.Window, d time.Duration) {
	b := w.Bounds()
	i := imdraw.New(nil)
	offset := b.Size().X / steps

	for _, bot := range s.bots {
		if bot.active {
			i.Color = pixel.RGB(0, 1, 0)
		} else {
			i.Color = pixel.RGB(1, 0, 0)
		}
		from := pixel.V(b.Min.X+25, b.Center().Y)
		pos := from.Add(pixel.V(float64(bot.pos)*offset, 0))

		i.Push(pos)

		i.Clear()
		i.Circle(50, 0)

		i.Draw(w)
	}
}

type state struct {
	bots []bot
}

func newState() state {
	return state{
		bots: []bot{
			{pos: 0, active: true},
			{pos: steps / 2},
		},
	}
}

type bot struct {
	pos    int
	active bool
}

func updateState(sOld state) state {
	s := sOld

	for i := range s.bots {
		updateBot(&s.bots[i], sOld)
	}
	return s
}

func updateBot(b *bot, s state) {
	if !b.active {
		return
	}

	b.pos++
	maybePassBaton(b, s)
}

func maybePassBaton(b *bot, s state) {
	for i, bb := range s.bots {
		if b.pos == bb.pos {
			continue // same bot
		}
		if bb.pos-b.pos == 1 {
			b.active = false
			s.bots[i].active = true
			return
		}
	}
}

const steps = 500

package main

import (
	"log"
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

	for !w.Closed() && !s.won {
		w.Clear(colornames.Peru)
		s = updateState(s)
		render(s, w, time.Since(start))
		w.Update()
		if s.won {
			log.Println("You win!")
		}
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
	won  bool
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

	var active *bot
	for i := range s.bots {
		if !s.bots[i].active {
			continue
		}
		active = &s.bots[i]
	}

	active.pos++
	maybePassBaton(active, &s)
	if won(*active, s) {
		s.won = true
	}

	return s
}

func maybePassBaton(b *bot, s *state) {
	for i, bb := range s.bots {
		if b == &bb {
			continue
		}
		if bb.pos-b.pos == 1 {
			b.active = false
			s.bots[i].active = true
			return
		}
	}
}

func won(b bot, s state) bool {
	return b.pos == steps
}

const steps = 150

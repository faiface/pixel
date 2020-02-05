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

	for !w.Closed() && !s.gameOver {
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
	im := imdraw.New(nil)
	hOffset := b.Size().X / steps
	vOffset := b.Size().Y / (numTeams + 1)

	for i, t := range s.teams {
		for j, bot := range t.bots {
			if &t.bots[j] == t.baton.holder {
				im.Color = pixel.RGB(0, 1, 0)
			} else {
				im.Color = pixel.RGB(1, 0, 0)
			}
			from := pixel.V(b.Min.X+25, b.Min.Y+float64(i+1)*vOffset)
			pos := from.Add(pixel.V(float64(bot.pos)*hOffset, 0))

			im.Push(pos)

			im.Clear()
			im.Circle(50, 0)

			im.Draw(w)
		}
	}
}

type state struct {
	teams    []team
	gameOver bool
}

func newState() state {
	var teams []team
	for i := 0; i < numTeams; i++ {
		var bots []bot
		for j := 0; j < numBots; j++ {
			bots = append(bots, bot{pos: j * (steps / numBots)})
		}
		teams = append(teams, team{
			bots:  bots,
			baton: baton{holder: &bots[0]},
		})
	}

	return state{
		teams: teams,
	}
}

type team struct {
	bots  []bot
	baton baton
	won   bool
}

type bot struct {
	pos int
}

type baton struct {
	holder *bot
}

func updateState(sOld state) state {
	s := sOld

	for _, t := range s.teams {
		b := t.baton.holder
		b.pos++
		maybePassBaton(t)
	}

	for _, t := range s.teams {
		if won(*t.baton.holder, s) {
			s.gameOver = true
		}
	}

	return s
}

func maybePassBaton(t team) {
	for i, b := range t.bots {
		h := t.baton.holder
		if h == &b {
			continue
		}
		if b.pos-h.pos == 1 {
			t.baton.holder = &t.bots[i]
			log.Println("pass!")
			return
		}
	}
}

func won(b bot, s state) bool {
	return b.pos == steps
}

func gameOver(s state) bool {
	for _, t := range s.teams {
		if t.won {
			return true
		}
	}
	return false
}

const (
	steps    = 150
	numBots  = 5
	numTeams = 2
)

package main

import (
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func render(s state, w *pixelgl.Window, d time.Duration, colors map[*team]pixel.RGBA) {
	renderBots(s, w, d, colors)
	renderObstacles(s, w)
}

func renderBots(s state, w *pixelgl.Window, d time.Duration, colors map[*team]pixel.RGBA) {
	b := w.Bounds()
	im := imdraw.New(nil)

	for i, t := range s.teams {
		for j, bot := range t.bots {
			if &t.bots[j] == t.baton.holder {
				im.Color = pixel.RGB(0, 1, 0)
			} else {
				im.Color = colors[&s.teams[i]]
			}

			pos := lanePos(bot.pos, i, botWidth, b)

			im.Push(pos)

			im.Clear()
			im.Circle(float64(botWidth), 0)

			im.Draw(w)
		}
	}
}

func lanePos(pos, lane int, width float64, bounds pixel.Rect) pixel.Vec {
	hOffset := bounds.Size().X / steps
	vOffset := bounds.Size().Y / (numTeams + 1)

	return pixel.V(bounds.Min.X+width/2+float64(pos)*hOffset,
		bounds.Min.Y+float64(lane+1)*vOffset)
}

func renderObstacles(s state, w *pixelgl.Window) {
	b := w.Bounds()
	im := imdraw.New(nil)

	for _, o := range s.obstacles {
		im.Color = pixel.RGB(1, 0, 1)

		pos := lanePos(o.pos, o.lane, botWidth, b)

		im.Push(pos)

		im.Clear()
		im.Circle(float64(botWidth), 0)

		im.Draw(w)
	}
}

func teamColors(ts []team) map[*team]pixel.RGBA {
	m := make(map[*team]pixel.RGBA)
	for i := range ts {
		var c color.RGBA
		switch i {
		case 0:
			c = colornames.Cyan
		case 1:
			c = colornames.Gold
		case 2:
			c = colornames.Lavender
		case 3:
			c = colornames.Indigo
		}
		m[&ts[i]] = pixel.ToRGBA(c)
	}
	return m
}

const (
	botWidth float64 = 20
)

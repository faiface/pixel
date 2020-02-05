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
	b := w.Bounds()
	im := imdraw.New(nil)
	hOffset := b.Size().X / steps
	vOffset := b.Size().Y / (numTeams + 1)
	width := 20

	for i, t := range s.teams {
		for j, bot := range t.bots {
			if &t.bots[j] == t.baton.holder {
				im.Color = pixel.RGB(0, 1, 0)
			} else {
				im.Color = colors[&s.teams[i]]
			}
			from := pixel.V(b.Min.X+float64(width)/2, b.Min.Y+float64(i+1)*vOffset)
			pos := from.Add(pixel.V(float64(bot.pos)*hOffset, 0))

			im.Push(pos)

			im.Clear()
			im.Circle(float64(width), 0)

			im.Draw(w)
		}
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

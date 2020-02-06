package gfx

import (
	"image/color"
	"relay/game"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func Render(s game.State, w *pixelgl.Window, d time.Duration) {
	colors := teamColors(s.Teams)
	renderBots(s, w, d, colors)
	renderObstacles(s, w)
}

func renderBots(s game.State, w *pixelgl.Window, d time.Duration, colors map[*game.Team]pixel.RGBA) {
	bounds := w.Bounds()
	im := imdraw.New(nil)

	for i, t := range s.Teams {
		for j, bot := range t.Bots {
			c := colors[&s.Teams[i]]
			c.R += 0.2 * float64(j)
			c.G -= 0.1 * float64(j)
			im.Color = c

			pos := lanePos(bot.Pos, bot.Lane, botWidth, bounds)

			im.Push(pos)

			im.Clear()
			im.Circle(botWidth, 0)

			im.Draw(w)
			if &t.Bots[j] == t.Baton.Holder {
				renderBaton(pos, w)
			}
		}
	}
}

func renderBaton(pos pixel.Vec, w *pixelgl.Window) {
	im := imdraw.New(nil)
	im.Color = pixel.RGB(0, 0, 0)
	im.Push(pos)
	im.Clear()
	im.Circle(batonWidth, 3)
	im.Draw(w)
}

func lanePos(pos, lane int, width float64, bounds pixel.Rect) pixel.Vec {
	hOffset := bounds.Size().X / game.Steps
	vOffset := bounds.Size().Y / (game.NumLanes + 1)

	return pixel.V(bounds.Min.X+width/2+float64(pos)*hOffset,
		bounds.Min.Y+float64(lane+1)*vOffset)
}

func renderObstacles(s game.State, w *pixelgl.Window) {
	b := w.Bounds()
	im := imdraw.New(nil)

	for _, o := range s.Obstacles {
		im.Color = pixel.RGB(1, 0, 1)

		pos := lanePos(o.Pos, o.Lane, botWidth, b)

		im.Push(pos)

		im.Clear()
		im.Circle(float64(botWidth), 0)

		im.Draw(w)
	}
}

func teamColors(ts []game.Team) map[*game.Team]pixel.RGBA {
	m := make(map[*game.Team]pixel.RGBA)
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
	botWidth   float64 = 20
	batonWidth float64 = 12
)

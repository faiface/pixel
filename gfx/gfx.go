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
	b := w.Bounds()
	im := imdraw.New(nil)

	for i, t := range s.Teams {
		for j, bot := range t.Bots {
			if &t.Bots[j] == t.Baton.Holder {
				im.Color = pixel.RGB(0, 1, 0)
			} else {
				im.Color = colors[&s.Teams[i]]
			}

			pos := lanePos(bot.Pos, i, botWidth, b)

			im.Push(pos)

			im.Clear()
			im.Circle(float64(botWidth), 0)

			im.Draw(w)
		}
	}
}

func lanePos(pos, lane int, width float64, bounds pixel.Rect) pixel.Vec {
	hOffset := bounds.Size().X / game.Steps
	vOffset := bounds.Size().Y / (game.NumTeams + 1)

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
	botWidth float64 = 20
)

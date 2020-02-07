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

type RenderState struct {
	Animating bool
	Frames    int
	frame     int
}

func Render(rs RenderState, sOld, sNew game.State, w *pixelgl.Window, d time.Duration) RenderState {
	//log.Println("render")
	w.Clear(colornames.Peru)

	tween := float64(rs.frame) / float64(rs.Frames)

	colors := teamColors(sNew.Teams)
	renderBots(sOld, sNew, tween, w, d, colors)
	renderObstacles(sNew, w)

	rs.frame++
	//log.Println("frame", rs.frame)
	if rs.frame >= rs.Frames {
		rs.Animating = false
	}
	return rs
}

func renderBots(sOld, sNew game.State, tween float64, w *pixelgl.Window, d time.Duration, colors map[*game.Team]pixel.RGBA) {
	bounds := w.Bounds()
	im := imdraw.New(nil)

	//log.Println("sOld.Teams:", sOld.Teams)

	for i, t := range sNew.Teams {
		for j, bot := range t.Bots {
			c := colors[&sNew.Teams[i]]
			c.R += 0.2 * float64(j)
			c.G -= 0.1 * float64(j)
			im.Color = c

			oldBot := sOld.Teams[i].Bots[j]
			// log.Println("oldBot:", oldBot)
			// log.Println("bot:", bot)
			oldPos := lanePos(oldBot.Pos, oldBot.Lane, botWidth, bounds)
			newPos := lanePos(bot.Pos, bot.Lane, botWidth, bounds)

			// log.Println("oldPos:", oldPos)
			// log.Println("newPos:", newPos)

			pos := pixel.Vec{
				X: oldPos.X + tween*(newPos.X-oldPos.X),
				Y: oldPos.Y + tween*(newPos.Y-oldPos.Y),
			}

			im.Push(pos)

			im.Clear()
			im.Circle(botWidth, 0)

			im.Draw(w)
			if t.Bots[j].ID == t.Baton.HolderID {
				renderBaton(pos, w)
			}
			//log.Println("sOld.Teams[i].Bots[j]:", sOld.Teams[i].Bots[j])
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
		im.Color = pixel.RGB(0.1, 0.1, 0.2)

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

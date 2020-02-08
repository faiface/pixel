package gfx

import (
	"image/color"
	"relay/game"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type RenderState struct {
	Animating bool
	Frames    int
	Frame     int
}

type context struct {
	sOld  game.State
	sNew  game.State
	tween float64
	w     *pixelgl.Window
}

func Render(rs RenderState, sOld, sNew game.State, w *pixelgl.Window) RenderState {
	w.Clear(colornames.Olivedrab)

	colors := teamColors(sNew.Teams)
	ctx := context{
		sOld:  sOld,
		sNew:  sNew,
		tween: float64(rs.Frame) / float64(rs.Frames),
		w:     w,
	}
	renderBots(ctx, colors)
	renderObstacles(sNew, w)

	rs.Frame++
	if rs.Frame > rs.Frames {
		rs.Animating = false
	}
	return rs
}

func renderBots(ctx context, colors map[*game.Team]pixel.RGBA) {
	for i, t := range ctx.sNew.Teams {
		c := colors[&ctx.sNew.Teams[i]]
		for j, bot := range t.Bots {
			oldBot := ctx.sOld.Teams[i].Bots[j]
			renderBot(oldBot, bot, ctx.sOld, ctx.sNew, ctx.w, c, ctx.tween)
		}

		oldHolder, newHolder := game.ActiveBot(ctx.sOld.Teams[i]), game.ActiveBot(ctx.sNew.Teams[i])
		oldPos := lanePos(oldHolder.Position.Pos, oldHolder.Position.Lane, botWidth, ctx.w.Bounds())
		newPos := lanePos(newHolder.Position.Pos, newHolder.Position.Lane, botWidth, ctx.w.Bounds())

		pos := pixel.Vec{
			X: oldPos.X + ctx.tween*(newPos.X-oldPos.X),
			Y: oldPos.Y + ctx.tween*(newPos.Y-oldPos.Y),
		}
		renderBaton(pos, ctx.w)
	}
}

func renderBot(oldBot, bot game.Bot, sOld, sNew game.State, w *pixelgl.Window, c pixel.RGBA, tween float64) {
	im := imdraw.New(nil)
	im.Color = c

	oldPos := lanePos(oldBot.Position.Pos, oldBot.Position.Lane, botWidth, w.Bounds())
	newPos := lanePos(bot.Position.Pos, bot.Position.Lane, botWidth, w.Bounds())

	pos := pixel.Vec{
		X: oldPos.X + tween*(newPos.X-oldPos.X),
		Y: oldPos.Y + tween*(newPos.Y-oldPos.Y),
	}

	im.Push(pos)
	im.Clear()
	im.Circle(botWidth, 0)
	im.Draw(w)
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

		pos := lanePos(o.Position.Pos, o.Position.Lane, botWidth, b)

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
			c = colornames.Red
		case 1:
			c = colornames.Green
		case 2:
			c = colornames.Blue
		case 3:
			c = colornames.Magenta
		case 4:
			c = colornames.Cyan
		case 5:
			c = colornames.Yellow
		}
		m[&ts[i]] = pixel.ToRGBA(c)
	}
	return m
}

const (
	botWidth   float64 = 17
	batonWidth float64 = 12
)

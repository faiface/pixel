package gfx

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math/rand"
	"os"
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

type spriteBank struct {
	bot      pixel.Picture
	obstacle pixel.Picture
}

func NewSpriteBank() (*spriteBank, error) {
	bot, err := loadPicture("shuttle.png")
	if err != nil {
		return nil, fmt.Errorf("load picture: %w", err)
	}

	ob, err := loadPicture("rock.png")
	if err != nil {
		return nil, fmt.Errorf("load picture: %w", err)
	}

	return &spriteBank{
		bot:      bot,
		obstacle: ob,
	}, nil
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func Render(rs RenderState, sOld, sNew game.State, w *pixelgl.Window, sb spriteBank) RenderState {
	renderBackground(w)

	colors := teamColors(sNew.Teams)
	ctx := context{
		sOld:  sOld,
		sNew:  sNew,
		tween: float64(rs.Frame) / float64(rs.Frames),
		w:     w,
	}
	renderBots(ctx, colors, sb.bot)
	renderObstacles(sNew, w, sb.obstacle)

	rs.Frame++
	if rs.Frame > rs.Frames {
		rs.Animating = false
	}
	return rs
}

var stars []pixel.Vec

func renderBackground(w *pixelgl.Window) {
	w.Clear(colornames.Black)

	if len(stars) == 0 {
		const numStars = 100
		for i := 0; i < numStars; i++ {
			stars = append(stars, pixel.Vec{
				X: rand.Float64() * w.Bounds().W(),
				Y: rand.Float64() * w.Bounds().H(),
			})
		}
	}

	for _, star := range stars {
		im := imdraw.New(nil)
		im.Color = colornames.White
		im.Push(star)
		im.Clear()
		im.Circle(2, 0)
		im.Draw(w)
	}
}

func renderBots(ctx context, colors map[*game.Team]pixel.RGBA, pic pixel.Picture) {
	for i, t := range ctx.sNew.Teams {
		c := colors[&ctx.sNew.Teams[i]]
		for j, bot := range t.Bots {
			oldBot := ctx.sOld.Teams[i].Bots[j]
			renderBot(ctx, oldBot, bot, c, pic)
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

func renderBot(ctx context, oldBot, bot game.Bot, c pixel.RGBA, pic pixel.Picture) {
	im := imdraw.New(nil)
	im.Color = c

	oldPos := lanePos(oldBot.Position.Pos, oldBot.Position.Lane, botWidth, ctx.w.Bounds())
	newPos := lanePos(bot.Position.Pos, bot.Position.Lane, botWidth, ctx.w.Bounds())

	pos := pixel.Vec{
		X: oldPos.X + ctx.tween*(newPos.X-oldPos.X),
		Y: oldPos.Y + ctx.tween*(newPos.Y-oldPos.Y),
	}

	im.Push(pos)
	im.Clear()
	im.Draw(ctx.w)
	bounds := pic.Bounds()
	sprite := pixel.NewSprite(pic, bounds)
	sprite.DrawColorMask(ctx.w, pixel.IM.Moved(pos).ScaledXY(pos, pixel.Vec{2, 2}), c)
}

func renderBaton(pos pixel.Vec, w *pixelgl.Window) {
	im := imdraw.New(nil)
	im.Color = colornames.Bisque
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

func renderObstacles(s game.State, w *pixelgl.Window, pic pixel.Picture) {
	b := w.Bounds()
	im := imdraw.New(nil)

	for _, o := range s.Obstacles {
		//im.Color = colornames.Slategray

		pos := lanePos(o.Position.Pos, o.Position.Lane, botWidth, b)

		im.Push(pos)

		im.Clear()
		sprite := pixel.NewSprite(pic, pic.Bounds())
		sprite.Draw(w, pixel.IM.Moved(pos))
		//im.Circle(float64(botWidth), 0)

		im.Draw(w)
	}
}

func teamColors(ts []game.Team) map[*game.Team]pixel.RGBA {
	m := make(map[*game.Team]pixel.RGBA)
	for i := range ts {
		var c color.RGBA
		switch i {
		case 0:
			c = colornames.Palevioletred
		case 1:
			c = colornames.Lime
		case 2:
			c = colornames.Cornflowerblue
		case 3:
			c = colornames.Magenta
		case 4:
			c = colornames.Cyan
		case 5:
			c = colornames.Yellow
		case 6:
			c = colornames.Blueviolet
		case 7:
			c = colornames.Orange
		case 8:
			c = colornames.Coral

		}
		m[&ts[i]] = pixel.ToRGBA(c)
	}
	return m
}

const (
	botWidth   float64 = 17
	batonWidth float64 = 12
)

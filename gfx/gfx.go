package gfx

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
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
	racer    pixel.Picture
	obstacle pixel.Picture
}

func NewSpriteBank() (*spriteBank, error) {
	racer, err := loadPicture("shuttle.png")
	if err != nil {
		return nil, fmt.Errorf("load picture: %w", err)
	}

	ob, err := loadPicture("rock.png")
	if err != nil {
		return nil, fmt.Errorf("load picture: %w", err)
	}

	return &spriteBank{
		racer:    racer,
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
	bgBatch := pixel.NewBatch(new(pixel.TrianglesData), nil)
	renderBackground(w, bgBatch)

	colors := teamColors(sNew.Teams)
	ctx := context{
		sOld:  sOld,
		sNew:  sNew,
		tween: float64(rs.Frame) / float64(rs.Frames),
		w:     w,
	}
	rBatch := pixel.NewBatch(new(pixel.TrianglesData), sb.racer)
	renderRacers(ctx, rBatch, colors, sb.racer)
	rBatch.Draw(w)

	oBatch := pixel.NewBatch(new(pixel.TrianglesData), sb.obstacle)
	renderObstacles(sNew, w, oBatch, sb.obstacle)
	oBatch.Draw(w)

	sBatch := pixel.NewBatch(new(pixel.TrianglesData), nil)
	renderSpawnPoints(sBatch, sNew.SpawnPoints, w.Bounds(), colors)
	sBatch.Draw(w)

	rs.Frame++
	if rs.Frame > rs.Frames {
		rs.Animating = false
	}
	return rs
}

var stars []pixel.Vec

func renderBackground(w *pixelgl.Window, batch *pixel.Batch) {
	w.Clear(colornames.Black)

	batch.Clear()

	if len(stars) == 0 {
		const numStars = 100
		for i := 0; i < numStars; i++ {
			stars = append(stars, pixel.Vec{
				X: rand.Float64() * w.Bounds().W(),
				Y: rand.Float64() * w.Bounds().H(),
			})
		}
	}

	im := imdraw.New(nil)
	im.Color = colornames.White
	for _, star := range stars {
		im.Push(star)
		im.Clear()
	}
	im.Circle(2, 0)
	im.Draw(batch)

	batch.Draw(w)
}

func renderRacers(ctx context, batch *pixel.Batch, colors map[int]pixel.RGBA, pic pixel.Picture) {
	for i, t := range ctx.sNew.Teams {
		c := colors[i]
		for j, racer := range t.Racers {
			oldRacer := ctx.sOld.Teams[i].Racers[j]
			renderRacer(ctx, batch, oldRacer, racer, racer.ID == ctx.sOld.Teams[i].Baton.HolderID, c, pic)
		}

		oldHolder, newHolder := game.ActiveRacer(ctx.sOld.Teams[i]), game.ActiveRacer(ctx.sNew.Teams[i])
		oldPos := lanePos(oldHolder.Position.Pos, oldHolder.Position.Lane, racerWidth, ctx.w.Bounds())
		newPos := lanePos(newHolder.Position.Pos, newHolder.Position.Lane, racerWidth, ctx.w.Bounds())

		pos := pixel.Vec{
			X: oldPos.X + ctx.tween*(newPos.X-oldPos.X),
			Y: oldPos.Y + ctx.tween*(newPos.Y-oldPos.Y),
		}

		renderBaton(pos, batch)
	}
}

func renderRacer(ctx context, batch *pixel.Batch, oldRacer, racer game.Racer, active bool, c pixel.RGBA, pic pixel.Picture) {
	oldPos := lanePos(oldRacer.Position.Pos, oldRacer.Position.Lane, racerWidth, ctx.w.Bounds())
	newPos := lanePos(racer.Position.Pos, racer.Position.Lane, racerWidth, ctx.w.Bounds())
	pos := pixel.Vec{
		X: oldPos.X + ctx.tween*(newPos.X-oldPos.X),
		Y: oldPos.Y + ctx.tween*(newPos.Y-oldPos.Y),
	}

	bounds := pic.Bounds()
	if active {
		renderProjection(ctx, batch, c, bounds, racer.Kinetics, pos, newPos)
	}

	sprite := pixel.NewSprite(pic, bounds)
	sprite.DrawColorMask(batch, pixel.IM.Moved(pos).ScaledXY(pos, pixel.Vec{1.7, 1.7}), c)

	renderFuelGuage(batch, pos, racer.Battery)
}

func renderFuelGuage(b *pixel.Batch, pos pixel.Vec, batt game.Battery) {
	w := 3.0

	im1, im2 := imdraw.New(nil), imdraw.New(nil)
	im1.Color = colornames.Yellow
	im2.Color = colornames.Yellow
	for i := 0; i < batt.Capacity; i++ {
		pos := pos
		pos.X -= racerWidth
		pos.Y -= racerWidth + w*2
		pos.X += (w * 2) * float64(i)
		if i >= batt.Charge {
			im2.Push(pos)
		} else {
			im1.Push(pos)
		}
	}
	im1.Circle(w, 0)
	im2.Circle(w, 1)
	im1.Draw(b)
	im2.Draw(b)
}

func renderProjection(ctx context, b *pixel.Batch, c pixel.RGBA, bounds pixel.Rect, k game.Kinetics, pos, newPos pixel.Vec) {
	im := imdraw.New(nil)
	projC := c
	alpha := 0.25
	projC.R *= alpha
	projC.G *= alpha
	projC.B *= alpha
	projC.A = alpha
	im.Color = projC
	w := bounds.W() * 0.65

	ll := pixel.Vec{
		X: pos.X + w,
		Y: pos.Y - w,
	}
	ur := pixel.Vec{
		X: pos.X + w*float64(k.V+1),
		Y: pos.Y + w,
	}
	if ctx.tween < 1 {
		ur.X = math.Min(ur.X, newPos.X+racerWidth)
	}
	ur.X = math.Max(ur.X, ll.X)

	im.Push(ll)
	im.Push(ur)
	im.Rectangle(0)
	im.Draw(b)
}

func renderBaton(pos pixel.Vec, b *pixel.Batch) {
	im := imdraw.New(nil)
	im.Color = colornames.Bisque
	im.Push(pos)
	im.Clear()
	im.Circle(batonWidth, 3)
	im.Draw(b)
}

func lanePos(pos, lane int, width float64, bounds pixel.Rect) pixel.Vec {
	hOffset := bounds.Size().X / game.Steps
	vOffset := bounds.Size().Y / (game.NumLanes + 1)

	return pixel.V(bounds.Min.X+width/2+float64(pos)*hOffset,
		bounds.Min.Y+float64(lane+1)*vOffset)
}

func renderObstacles(s game.State, w *pixelgl.Window, batch *pixel.Batch, pic pixel.Picture) {
	b := w.Bounds()
	im := imdraw.New(nil)

	for _, o := range s.Obstacles {
		pos := lanePos(o.Position.Pos, o.Position.Lane, racerWidth, b)

		im.Push(pos)
		sprite := pixel.NewSprite(pic, pic.Bounds())
		sprite.Draw(batch, pixel.IM.Moved(pos))

	}
	im.Draw(batch)
}

func renderSpawnPoints(b *pixel.Batch, sps map[int]game.SpawnPoint, bounds pixel.Rect, colors map[int]pixel.RGBA) {
	im := imdraw.New(nil)

	for _, sp := range sps {
		c := colors[sp.TeamID]
		c.R *= 0.5
		c.G *= 0.5
		c.B *= 0.5
		im.Color = c

		pos := lanePos(sp.Pos.Pos, sp.Pos.Lane, racerWidth, bounds)

		im.Push(pos)
		im.Circle(float64(racerWidth)*1.5, 3)
	}
	im.Draw(b)
}

func teamColors(ts []game.Team) map[int]pixel.RGBA {
	m := make(map[int]pixel.RGBA)
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
		m[i] = pixel.ToRGBA(c)
	}
	return m
}

const (
	racerWidth float64 = 17
	batonWidth float64 = 12
)

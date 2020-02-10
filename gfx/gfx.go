package gfx

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"github.com/snargleplax/relay/game"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func RenderLoop(w *pixelgl.Window, s game.State, stateC <-chan game.State, sb *SpriteBank) {
	sOld := s
	var (
		frames = 0
		second = time.Tick(time.Second)
		rs     = renderState{
			frames: 10,
		}
	)

	for !w.Closed() {
		if rs.frame == rs.frames {
			select {
			case ss := <-stateC:
				sOld = s
				s = ss
				rs.frame = 0
				rs.timeFlowing = true
			default:
			}
		}

		if rs.frame == rs.frames && rs.timeFlowing {
			rs.timeFlowing = false
		}
		rs = render(rs, sOld, s, w, *sb)
		w.SwapBuffers()
		frames++

		select {
		case <-second:
			w.SetTitle(fmt.Sprintf("%s | FPS: %d", "Relay", frames))
			frames = 0
		default:
		}
	}
}

type context struct {
	sOld game.State
	sNew game.State
	rs   renderState
	w    *pixelgl.Window
}

type renderState struct {
	frames      int
	frame       int
	timeFlowing bool
}

func (rs renderState) tween() float64 { return float64(rs.frame) / float64(rs.frames) }

func render(rs renderState, sOld, sNew game.State, w *pixelgl.Window, sb SpriteBank) renderState {
	w.Clear(colornames.Black)

	bgBatch := pixel.NewBatch(new(pixel.TrianglesData), nil)
	renderBackground(w.Bounds(), bgBatch)
	bgBatch.Draw(w)

	oBatch := pixel.NewBatch(new(pixel.TrianglesData), sb.obstacle)
	renderObstacles(sNew.Obstacles, w.Bounds(), oBatch, sb.obstacle)
	oBatch.Draw(w)

	dBatch := pixel.NewBatch(new(pixel.TrianglesData), sb.derelict)
	renderObstacles(sNew.Derelicts, w.Bounds(), dBatch, sb.derelict)
	dBatch.Draw(w)

	sBatch := pixel.NewBatch(new(pixel.TrianglesData), nil)
	renderSpawnPoints(sBatch, sNew.SpawnPoints, w.Bounds())
	sBatch.Draw(w)

	ctx := context{
		sOld: sOld,
		sNew: sNew,
		rs:   rs,
		w:    w,
	}
	rBatch := pixel.NewBatch(new(pixel.TrianglesData), sb.racer)
	renderRacers(ctx, rBatch, sb.racer)
	rBatch.Draw(w)

	if rs.frame != rs.frames {
		rs.frame++
	}
	return rs
}

type SpriteBank struct {
	racer    pixel.Picture
	obstacle pixel.Picture
	derelict pixel.Picture
}

func NewSpriteBank() (*SpriteBank, error) {
	var sb SpriteBank
	for file, field := range map[string]*pixel.Picture{
		"shuttle.png":  &sb.racer,
		"rock.png":     &sb.obstacle,
		"derelict.png": &sb.derelict,
	} {
		p, err := loadPicture(file)
		if err != nil {
			return nil, fmt.Errorf("load picture %q: %w", file, err)
		}
		*field = p
	}
	return &sb, nil
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

var stars []pixel.Vec

func renderBackground(bounds pixel.Rect, batch *pixel.Batch) {
	if len(stars) == 0 {
		const numStars = 100
		for i := 0; i < numStars; i++ {
			stars = append(stars, pixel.Vec{
				X: rand.Float64() * bounds.W(),
				Y: rand.Float64() * bounds.H(),
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
}

func renderRacers(ctx context, batch *pixel.Batch, pic pixel.Picture) {
	for i, t := range ctx.sNew.Teams {
		c := teamColors[i]
		for j, racer := range t.Racers {
			oldRacer := ctx.sOld.Teams[i].Racers[j]
			renderRacer(ctx, batch, oldRacer, racer, racer.ID == ctx.sOld.Teams[i].Baton.HolderID, c, pic)
		}

		oldHolder, newHolder := game.ActiveRacer(ctx.sOld.Teams[i]), game.ActiveRacer(ctx.sNew.Teams[i])
		oldPos := lanePos(oldHolder.Position, ctx.w.Bounds())
		newPos := lanePos(newHolder.Position, ctx.w.Bounds())

		pos := pixel.Vec{
			X: oldPos.X + ctx.rs.tween()*(newPos.X-oldPos.X),
			Y: oldPos.Y + ctx.rs.tween()*(newPos.Y-oldPos.Y),
		}

		renderBaton(pos, batch)
	}
}

func renderRacer(ctx context, batch *pixel.Batch, oldRacer, racer game.Racer, active bool, c pixel.RGBA, pic pixel.Picture) {
	oldPos := lanePos(oldRacer.Position, ctx.w.Bounds())
	newPos := lanePos(racer.Position, ctx.w.Bounds())
	pos := pixel.Vec{
		X: oldPos.X + ctx.rs.tween()*(newPos.X-oldPos.X),
		Y: oldPos.Y + ctx.rs.tween()*(newPos.Y-oldPos.Y),
	}

	bounds := pic.Bounds()
	if active {
		renderProjection(ctx, batch, c, bounds, racer.Position, racer.Kinetics.VX, pos)
	}

	sprite := pixel.NewSprite(pic, bounds)
	sprite.DrawColorMask(batch, pixel.IM.Moved(pos).ScaledXY(pos, pixel.Vec{scale, scale}), c)

	//renderFuelGuage(batch, pos, racer.Battery)
}

func renderProjection(ctx context, b *pixel.Batch, c pixel.RGBA, bounds pixel.Rect, p game.Position, vx int, pos pixel.Vec) {
	if vx == 0 {
		return
	}

	im := imdraw.New(nil)
	projC := c
	alpha := 0.25
	projC.R *= alpha
	projC.G *= alpha
	projC.B *= alpha
	projC.A = alpha
	im.Color = projC

	w := bounds.W() * scale / 2

	ll := pixel.Vec{
		X: pos.X + w,
		Y: pos.Y - w,
	}

	nextPos := p
	if !ctx.rs.timeFlowing {
		nextPos.Pos += vx
	}
	vNext := lanePos(nextPos, ctx.w.Bounds())
	vNext.X += w

	ur := pixel.Vec{
		X: math.Max(vNext.X+w, ll.X),
		Y: pos.Y + w,
	}

	im.Push(ll)
	im.Push(ur)
	im.Rectangle(0)
	im.Draw(b)
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

func renderBaton(pos pixel.Vec, b *pixel.Batch) {
	im := imdraw.New(nil)
	im.Color = colornames.Bisque
	im.Push(pos)
	im.Clear()
	im.Circle(batonWidth, 3)
	im.Draw(b)
}

func lanePos(pos game.Position, bounds pixel.Rect) pixel.Vec {
	hOffset := bounds.Size().X / game.Steps
	vOffset := bounds.Size().Y / (game.NumLanes + 1)

	return pixel.V(bounds.Min.X+racerWidth+float64(pos.Pos)*hOffset,
		bounds.Min.Y+float64(pos.Lane+1)*vOffset)
}

func renderObstacles(os []game.Obstacle, bounds pixel.Rect, batch *pixel.Batch, pic pixel.Picture) {
	im := imdraw.New(nil)

	for _, o := range os {
		pos := lanePos(o.Position, bounds)

		im.Push(pos)
		sprite := pixel.NewSprite(pic, pic.Bounds())
		sprite.Draw(batch, pixel.IM.Moved(pos).ScaledXY(pos, pixel.Vec{scale, scale}))

	}
	im.Draw(batch)
}

func renderSpawnPoints(b *pixel.Batch, sps map[int]game.SpawnPoint, bounds pixel.Rect) {
	im := imdraw.New(nil)

	for _, sp := range sps {
		c := teamColors[sp.TeamID]
		c.R *= 0.5
		c.G *= 0.5
		c.B *= 0.5
		im.Color = c

		pos := lanePos(sp.Position, bounds)

		im.Push(pos)
		im.Circle(float64(racerWidth)*1.5, 3)
	}
	im.Draw(b)
}

var teamColors = []pixel.RGBA{
	pixel.ToRGBA(colornames.Palevioletred),
	pixel.ToRGBA(colornames.Lime),
	pixel.ToRGBA(colornames.Cornflowerblue),
	pixel.ToRGBA(colornames.Magenta),
	pixel.ToRGBA(colornames.Cyan),
	pixel.ToRGBA(colornames.Yellow),
	pixel.ToRGBA(colornames.Blueviolet),
	pixel.ToRGBA(colornames.Orange),
}

const (
	racerWidth float64 = 17
	batonWidth float64 = 12
	scale              = 1.5
)

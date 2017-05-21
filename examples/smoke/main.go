package main

import (
	"container/list"
	"encoding/csv"
	"image"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type particle struct {
	Sprite     *pixel.Sprite
	Pos        pixel.Vec
	Rot, Scale float64
	Mask       pixel.RGBA
	Data       interface{}
}

type particles struct {
	Generate            func() *particle
	Update              func(dt float64, p *particle) bool
	SpawnAvg, SpawnDist float64

	parts     list.List
	spawnTime float64
}

func (p *particles) UpdateAll(dt float64) {
	p.spawnTime -= dt
	for p.spawnTime <= 0 {
		p.parts.PushFront(p.Generate())
		p.spawnTime += math.Max(0, p.SpawnAvg+rand.NormFloat64()*p.SpawnDist)
	}

	for e := p.parts.Front(); e != nil; e = e.Next() {
		part := e.Value.(*particle)
		if !p.Update(dt, part) {
			defer p.parts.Remove(e)
		}
	}
}

func (p *particles) DrawAll(t pixel.Target) {
	for e := p.parts.Front(); e != nil; e = e.Next() {
		part := e.Value.(*particle)

		part.Sprite.DrawColorMask(
			t,
			pixel.IM.
				Scaled(pixel.ZV, part.Scale).
				Rotated(pixel.ZV, part.Rot).
				Moved(part.Pos),
			part.Mask,
		)
	}
}

type smokeData struct {
	Vel  pixel.Vec
	Time float64
	Life float64
}

type smokeSystem struct {
	Sheet pixel.Picture
	Rects []pixel.Rect
	Orig  pixel.Vec

	VelBasis []pixel.Vec
	VelDist  float64

	LifeAvg, LifeDist float64
}

func (ss *smokeSystem) Generate() *particle {
	sd := new(smokeData)
	for _, base := range ss.VelBasis {
		c := math.Max(0, 1+rand.NormFloat64()*ss.VelDist)
		sd.Vel = sd.Vel.Add(base.Scaled(c))
	}
	sd.Vel = sd.Vel.Scaled(1 / float64(len(ss.VelBasis)))
	sd.Life = math.Max(0, ss.LifeAvg+rand.NormFloat64()*ss.LifeDist)

	p := new(particle)
	p.Data = sd

	p.Pos = ss.Orig
	p.Scale = 1
	p.Mask = pixel.Alpha(1)
	p.Sprite = pixel.NewSprite(ss.Sheet, ss.Rects[rand.Intn(len(ss.Rects))])

	return p
}

func (ss *smokeSystem) Update(dt float64, p *particle) bool {
	sd := p.Data.(*smokeData)
	sd.Time += dt

	frac := sd.Time / sd.Life

	p.Pos = p.Pos.Add(sd.Vel.Scaled(dt))
	p.Scale = 0.5 + frac*1.5

	const (
		fadeIn  = 0.2
		fadeOut = 0.4
	)
	if frac < fadeIn {
		p.Mask = pixel.Alpha(math.Pow(frac/fadeIn, 0.75))
	} else if frac >= fadeOut {
		p.Mask = pixel.Alpha(math.Pow(1-(frac-fadeOut)/(1-fadeOut), 1.5))
	} else {
		p.Mask = pixel.Alpha(1)
	}

	return sd.Time < sd.Life
}

func loadSpriteSheet(sheetPath, descriptionPath string) (sheet pixel.Picture, rects []pixel.Rect, err error) {
	sheetFile, err := os.Open(sheetPath)
	if err != nil {
		return nil, nil, err
	}
	defer sheetFile.Close()

	sheetImg, _, err := image.Decode(sheetFile)
	if err != nil {
		return nil, nil, err
	}

	sheet = pixel.PictureDataFromImage(sheetImg)

	descriptionFile, err := os.Open(descriptionPath)
	if err != nil {
		return nil, nil, err
	}
	defer descriptionFile.Close()

	description := csv.NewReader(descriptionFile)
	for {
		record, err := description.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		x, _ := strconv.ParseFloat(record[0], 64)
		y, _ := strconv.ParseFloat(record[1], 64)
		w, _ := strconv.ParseFloat(record[2], 64)
		h, _ := strconv.ParseFloat(record[3], 64)

		y = sheet.Bounds().H() - y - h

		rects = append(rects, pixel.R(x, y, x+w, y+h))
	}

	return sheet, rects, nil
}

func run() {
	sheet, rects, err := loadSpriteSheet("blackSmoke.png", "blackSmoke.csv")
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Title:     "Smoke",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
		VSync:     true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	ss := &smokeSystem{
		Rects:    rects,
		Orig:     pixel.ZV,
		VelBasis: []pixel.Vec{pixel.V(-100, 100), pixel.V(100, 100), pixel.V(0, 100)},
		VelDist:  0.1,
		LifeAvg:  7,
		LifeDist: 0.5,
	}

	p := &particles{
		Generate:  ss.Generate,
		Update:    ss.Update,
		SpawnAvg:  0.3,
		SpawnDist: 0.1,
	}

	batch := pixel.NewBatch(&pixel.TrianglesData{}, sheet)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		p.UpdateAll(dt)

		win.Clear(colornames.Aliceblue)

		orig := win.Bounds().Center()
		orig.Y -= win.Bounds().H() / 2
		win.SetMatrix(pixel.IM.Moved(orig))

		batch.Clear()
		p.DrawAll(batch)
		batch.Draw(win)

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

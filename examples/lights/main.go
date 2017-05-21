package main

import (
	"image"
	"math"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

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

type colorlight struct {
	color  pixel.RGBA
	point  pixel.Vec
	angle  float64
	radius float64
	dust   float64

	spread float64

	imd *imdraw.IMDraw
}

func (cl *colorlight) apply(dst pixel.ComposeTarget, center pixel.Vec, src, noise *pixel.Sprite) {
	// create the light arc if not created already
	if cl.imd == nil {
		imd := imdraw.New(nil)
		imd.Color = pixel.Alpha(1)
		imd.Push(pixel.ZV)
		imd.Color = pixel.Alpha(0)
		for angle := -cl.spread / 2; angle <= cl.spread/2; angle += cl.spread / 64 {
			imd.Push(pixel.V(1, 0).Rotated(angle))
		}
		imd.Polygon(0)
		cl.imd = imd
	}

	// draw the light arc
	dst.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	dst.SetColorMask(pixel.Alpha(1))
	dst.SetComposeMethod(pixel.ComposePlus)
	cl.imd.Draw(dst)

	// draw the noise inside the light
	dst.SetMatrix(pixel.IM)
	dst.SetComposeMethod(pixel.ComposeIn)
	noise.Draw(dst, pixel.IM.Moved(center))

	// draw an image inside the noisy light
	dst.SetColorMask(cl.color)
	dst.SetComposeMethod(pixel.ComposeIn)
	src.Draw(dst, pixel.IM.Moved(center))

	// draw the light reflected from the dust
	dst.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	dst.SetColorMask(cl.color.Mul(pixel.Alpha(cl.dust)))
	dst.SetComposeMethod(pixel.ComposePlus)
	cl.imd.Draw(dst)
}

func run() {
	pandaPic, err := loadPicture("panda.png")
	if err != nil {
		panic(err)
	}
	noisePic, err := loadPicture("noise.png")
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Title:  "Lights",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	panda := pixel.NewSprite(pandaPic, pandaPic.Bounds())
	noise := pixel.NewSprite(noisePic, noisePic.Bounds())

	colors := []pixel.RGBA{
		pixel.RGB(1, 0, 0),
		pixel.RGB(0, 1, 0),
		pixel.RGB(0, 0, 1),
		pixel.RGB(1/math.Sqrt2, 1/math.Sqrt2, 0),
	}

	points := []pixel.Vec{
		{X: win.Bounds().Min.X, Y: win.Bounds().Min.Y},
		{X: win.Bounds().Max.X, Y: win.Bounds().Min.Y},
		{X: win.Bounds().Max.X, Y: win.Bounds().Max.Y},
		{X: win.Bounds().Min.X, Y: win.Bounds().Max.Y},
	}

	angles := []float64{
		math.Pi / 4,
		math.Pi/4 + math.Pi/2,
		math.Pi/4 + 2*math.Pi/2,
		math.Pi/4 + 3*math.Pi/2,
	}

	lights := make([]colorlight, 4)
	for i := range lights {
		lights[i] = colorlight{
			color:  colors[i],
			point:  points[i],
			angle:  angles[i],
			radius: 800,
			dust:   0.3,
			spread: math.Pi / math.E,
		}
	}

	speed := []float64{11.0 / 23, 13.0 / 23, 17.0 / 23, 19.0 / 23}

	oneLight := pixelgl.NewCanvas(win.Bounds())
	allLight := pixelgl.NewCanvas(win.Bounds())

	fps30 := time.Tick(time.Second / 30)

	start := time.Now()
	for !win.Closed() {
		if win.Pressed(pixelgl.KeyW) {
			for i := range lights {
				lights[i].dust += 0.05
				if lights[i].dust > 1 {
					lights[i].dust = 1
				}
			}
		}
		if win.Pressed(pixelgl.KeyS) {
			for i := range lights {
				lights[i].dust -= 0.05
				if lights[i].dust < 0 {
					lights[i].dust = 0
				}
			}
		}

		since := time.Since(start).Seconds()
		for i := range lights {
			lights[i].angle = angles[i] + math.Sin(since*speed[i])*math.Pi/8
		}

		win.Clear(pixel.RGB(0, 0, 0))

		// draw the panda visible outside the light
		win.SetColorMask(pixel.Alpha(0.4))
		win.SetComposeMethod(pixel.ComposePlus)
		panda.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		allLight.Clear(pixel.Alpha(0))
		allLight.SetComposeMethod(pixel.ComposePlus)

		// accumulate all the lights
		for i := range lights {
			oneLight.Clear(pixel.Alpha(0))
			lights[i].apply(oneLight, oneLight.Bounds().Center(), panda, noise)
			oneLight.Draw(allLight, pixel.IM.Moved(allLight.Bounds().Center()))
		}

		// compose the final result
		win.SetColorMask(pixel.Alpha(1))
		allLight.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()

		<-fps30 // maintain 30 fps, because my computer couldn't handle 60 here
	}
}

func main() {
	pixelgl.Run(run)
}

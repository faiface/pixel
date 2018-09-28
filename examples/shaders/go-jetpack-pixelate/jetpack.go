package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	"github.com/go-gl/mathgl/mgl32"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
)

// InstallShader ...
func InstallShader(w *pixelgl.Window, uAmount *float32, uResolution, uMouse *mgl32.Vec2) {
	wc := w.GetCanvas()
	wc.BindUniform("u_resolution", uResolution)
	wc.BindUniform("u_mouse", uMouse)
	wc.BindUniform("u_amount", uAmount)
	wc.SetFragmentShader(pixelateFragShader)
	wc.UpdateShader()
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

func loadSprite(path string) (pixel.Sprite, error) {
	pic, err := loadPicture(path)
	if err != nil {
		return *pixel.NewSprite(pic, pic.Bounds()), err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return *sprite, nil
}

func loadTTF(path string, size float64, origin pixel.Vec) *text.Text {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	face := truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})

	atlas := text.NewAtlas(face, text.ASCII)

	txt := text.New(origin, atlas)

	return txt

}

func run() {
	// Set up window configs
	cfg := pixelgl.WindowConfig{ // Default: 1024 x 768
		Title:  "Golang Jetpack!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Importantr variables
	var jetX, jetY, velX, velY, radians float64
	flipped := 1.0
	jetpackOn := false
	gravity := 0.6 // Default: 0.004
	jetAcc := 0.9  // Default: 0.008
	tilt := 0.01   // Default: 0.001
	whichOn := false
	onNumber := 0
	jetpackOffName := "jetpack.png"
	jetpackOn1Name := "jetpack-on.png"
	jetpackOn2Name := "jetpack-on2.png"
	camVector := win.Bounds().Center()

	var uAmount float32
	var uMouse, uResolution mgl32.Vec2

	InstallShader(win, &uAmount, &uResolution, &uMouse)

	bg, _ := loadSprite("sky.png")

	// Jetpack - Rendering
	jetpackOff, err := loadSprite(jetpackOffName)
	if err != nil {
		panic(err)
	}
	jetpackOn1, err := loadSprite(jetpackOn1Name)
	if err != nil {
		panic(err)
	}
	jetpackOn2, err := loadSprite(jetpackOn2Name)
	if err != nil {
		panic(err)
	}

	// Tutorial Text
	txt := loadTTF("intuitive.ttf", 50, pixel.V(win.Bounds().Center().X-450, win.Bounds().Center().Y-200))
	fmt.Fprintf(txt, "Explore the Skies with WASD or Arrow Keys!")

	currentSprite := jetpackOff

	canvas := pixelgl.NewCanvas(win.Bounds())
	uResolution[0] = float32(win.Bounds().W())
	uResolution[1] = float32(win.Bounds().H())
	uAmount = 300.0
	// Game Loop
	for !win.Closed() {

		mpos := win.MousePosition()
		uMouse[0] = float32(mpos.X)
		uMouse[1] = float32(mpos.Y)
		win.SetTitle(fmt.Sprint(uAmount))
		win.Clear(colornames.Blue)
		canvas.Clear(colornames.Green)

		// Jetpack - Controls
		jetpackOn = win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW)

		if win.Pressed(pixelgl.KeyRight) || win.Pressed(pixelgl.KeyD) {
			jetpackOn = true
			flipped = -1
			radians -= tilt
			velX += tilt * 30
		} else if win.Pressed(pixelgl.KeyLeft) || win.Pressed(pixelgl.KeyA) {
			jetpackOn = true
			flipped = 1
			radians += tilt
			velX -= tilt * 30
		} else {
			if velX < 0 {
				radians -= tilt / 3
				velX += tilt * 10
			} else if velX > 0 {
				radians += tilt / 3
				velX -= tilt * 10
			}
		}
		if jetY < 0 {
			jetY = 0
			velY = -0.3 * velY
		}

		if jetpackOn {
			velY += jetAcc
			whichOn = !whichOn
			onNumber++
			if onNumber == 5 { // every 5 frames, toggle anijetMation
				onNumber = 0
				if whichOn {
					currentSprite = jetpackOn1
				} else {
					currentSprite = jetpackOn2
				}
			}
		} else {
			currentSprite = jetpackOff
			velY -= gravity
		}

		if win.Pressed(pixelgl.KeyEqual) {
			uAmount += 10
		}
		if win.Pressed(pixelgl.KeyMinus) {
			uAmount -= 10
		}

		positionVector := pixel.V(win.Bounds().Center().X+jetX, win.Bounds().Center().Y+jetY-372)
		jetMat := pixel.IM
		jetMat = jetMat.Scaled(pixel.ZV, 4)
		jetMat = jetMat.Moved(positionVector)
		jetMat = jetMat.ScaledXY(positionVector, pixel.V(flipped, 1))
		jetMat = jetMat.Rotated(positionVector, radians)

		jetX += velX
		jetY += velY

		// Camera
		camVector.X += (positionVector.X - camVector.X) * 0.2
		camVector.Y += (positionVector.Y - camVector.Y) * 0.2

		if camVector.X > 25085 {
			camVector.X = 25085
		} else if camVector.X < -14843 {
			camVector.X = -14843
		}

		if camVector.Y > 22500 {
			camVector.Y = 22500
		}

		cam := pixel.IM.Moved(win.Bounds().Center().Sub(camVector))

		canvas.SetMatrix(cam)

		// Drawing to the screen
		win.SetSmooth(true)
		bg.Draw(canvas, pixel.IM.Moved(pixel.V(win.Bounds().Center().X, win.Bounds().Center().Y+766)).Scaled(pixel.ZV, 10))
		txt.Draw(canvas, pixel.IM)
		win.SetSmooth(false)
		currentSprite.Draw(canvas, jetMat)
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}

}

func main() {
	pixelgl.Run(run)
}

var pixelateFragShader = `
#version 330 core
#ifdef GL_ES
precision mediump float;
precision mediump int;
#endif

in vec4 Color;
in vec2 texcoords;
in vec2 glpos;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform float u_amount;
uniform vec2 u_mouse;
uniform vec2 u_resolution;

void main(void)
{
	vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
	
	float d = 1.0 / u_amount;
	float ar = u_resolution.x / u_resolution.y;
	float u = floor( t.x / d ) * d;
	d = ar / u_amount;
	float v = floor( t.y / d ) * d;
	fragColor = texture( u_texture, vec2( u, v ) );
}
`

/*
	//fragColor = vec4(1.0,0.0,0.0,1.0);
	// float d = 1.0 / u_amount;
	// float ar = u_resolution.x / u_resolution.y;
	// float u = floor( texcoords.x / d ) * d;
	// d = ar / u_amount;
	// float v = floor( texcoords.y / d ) * d;
	// fragColor = texture( u_texture, vec2( u, v ) );

  	// vec2 p = t.st;
	// p.x -= mod(t.x / glpos.x, t.x / glpos.x + 0.1);
	// p.y -= mod(t.y / glpos.y, t.y / glpos.y + 0.1);

	// fragColor = texture(u_texture, p).rgba;
*/
// varying vec2 vUv;
// uniform sampler2D tInput;
// uniform vec2 resolution;
// uniform float amount;

// void main() {

// 	float d = 1.0 / amount;
// 	float ar = resolution.x / resolution.y;
// 	float u = floor( vUv.x / d ) * d;
// 	d = ar / amount;
// 	float v = floor( vUv.y / d ) * d;
// 	gl_FragColor = texture2D( tInput, vec2( u, v ) );

// }

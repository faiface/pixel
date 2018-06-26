package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
)

var fragmentShader = `// afl_ext @ 2016
 
#ifdef GL_ES
precision highp float;
#endif
 
#extension GL_OES_standard_derivatives : enable
 
#define HOW_CLOUDY 0.4
#define SHADOW_THRESHOLD 0.2
#define SHADOW 0.2
#define SUBSURFACE 1.0
#define WIND_DIRECTION 5.0
#define TIME_SCALE 1.7
#define SCALE 0.3
#define ENABLE_SHAFTS
 
#define iGlobalTime time
#define iMouse (mouse.xy * resolution.xy)
#define iResolution resolution
 
mat2 RM = mat2(cos(WIND_DIRECTION), -sin(WIND_DIRECTION), sin(WIND_DIRECTION), cos(WIND_DIRECTION));
uniform float time;
uniform vec2 mouse;
uniform vec2 resolution;
 
float hash( float n )
{
    return fract(sin(n)*758.5453);
}
 
float noise( in vec3 x )
{
    vec3 p = floor(x);
    vec3 f = fract(x);
    float n = p.x + p.y*57.0 + p.z*800.0;
    float res = mix(mix(mix( hash(n+  0.0), hash(n+  1.0),f.x), mix( hash(n+ 57.0), hash(n+ 58.0),f.x),f.y),
            mix(mix( hash(n+800.0), hash(n+801.0),f.x), mix( hash(n+857.0), hash(n+858.0),f.x),f.y),f.z);
    return res;
}
 
float fbm( vec3 p )
{
    float f = 0.0;
    f += 0.50000*noise( p ); p = p*2.02;
    f -= 0.25000*noise( p ); p = p*2.03;
    f += 0.12500*noise( p ); p = p*3.01;
    f += 0.06250*noise( p ); p = p*3.04;
    f += 0.03500*noise( p ); p = p*4.01;
    f += 0.01250*noise( p ); p = p*4.04;
    f -= 0.00125*noise( p );
    return f/0.984375;
}
 
float cloud(vec3 p)
{
    p-=fbm(vec3(p.x,p.y,0.0)*0.5)*1.25;
    float a = min((fbm(p*3.0)*2.2-1.1), 0.0);
    return a*a;
}
 
float shadow = 1.0;
 
 
float clouds(vec2 p){
    float ic = cloud(vec3(p * 2.0, iGlobalTime*0.01 * TIME_SCALE)) / HOW_CLOUDY;
    float init = smoothstep(0.1, 1.0, ic) * 10.0;
    shadow = smoothstep(0.0, SHADOW_THRESHOLD, ic) * SHADOW + (1.0 - SHADOW);
    init = (init * cloud(vec3(p * (6.0), iGlobalTime*0.01 * TIME_SCALE)) * ic);
    init = (init * (cloud(vec3(p * (11.0), iGlobalTime*0.01 * TIME_SCALE))*0.5 + 0.4) * init);
    return min(1.0, init);
}
uniform sampler2D bb;
float cloudslowres(vec2 p){
    return 1.0 - (texture2D(bb, p).a - 0.9) * 10.0;
}
 
vec2 ratio = vec2(1.0, 1.0);
 
vec4 getresult(){
    vec2 surfacePosition = ((( gl_FragCoord.xy / iResolution.xy ) * vec2(iResolution.x / iResolution.y, 1.0)) * 2.0 - 1.0)*SCALE;
    vec2 position = ( surfacePosition);
    vec2 sun = ((iMouse.xy/ iResolution.xy)* vec2(iResolution.x / iResolution.y, 1.0)*2.0-1.0) * SCALE;
    float dst = distance(sun * ratio, position * ratio);
    float suni = pow(dst + 1.0, -10.0);
    float shaft =0.0;
    float st = 0.05;
    float w = 1.0;
    vec2 dir = sun - position;
    float c = clouds(position);
    #ifdef ENABLE_SHAFTS
    for(int i=0;i<50;i++){
        float occl = cloudslowres(clamp((gl_FragCoord.xy / iResolution.xy) + dir * st, 0.0, 1.0));
        w *= 0.99;
        st *= 1.05;
        shaft += max(0.0, (1.0 - occl)) * w;
    }
    #endif
    shadow = min(1.0, shadow + suni * suni * 0.2 * SUBSURFACE);
    suni *= (shaft * 0.03);
    return vec4(pow(mix(vec3(shadow), pow(vec3(0.23, 0.33, 0.48), vec3(2.2)) + suni, c), vec3(1.0/2.2)), c*0.1 + 0.9);     
}
 
void main( void ) {
    gl_FragColor = getresult().rgba;
}
`

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

	b := pixel.R(0, 0, 1024, 768)

	cfg := pixelgl.WindowConfig{ // Default: 1024 x 768
		Title:  "Golang Jetpack!",
		Bounds: b,
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
	jetpackName := "jetpack.png"
	camVector := win.Bounds().Center()

	bg, _ := loadSprite("sky.png")

	c := pixelgl.NewCanvas(b)

	var mouse mgl32.Vec2
	var resolution mgl32.Vec2
	var t float32
	c.BindUniform("mouse", &mouse)
	c.BindUniform("resolution", &resolution)
	c.BindUniform("time", &t)
	c.SetFragmentShader(fragmentShader)
	c.UpdateShader()

	// Tutorial Text
	txt := loadTTF("intuitive.ttf", 50, pixel.V(win.Bounds().Center().X-450, win.Bounds().Center().Y-200))
	fmt.Fprintf(txt, "Explore the Skies with WASD or Arrow Keys!")

	// Game Loop
	for !win.Closed() {
		win.Update()
		win.Clear(colornames.Green)

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
			onNumber += 1
			if onNumber == 5 { // every 5 frames, toggle anijetMation
				onNumber = 0
				if whichOn {
					jetpackName = "jetpack-on.png"
				} else {
					jetpackName = "jetpack-on2.png"
				}
			}
		} else {
			jetpackName = "jetpack.png"
			velY -= gravity
		}

		// Jetpack - Rendering
		jetpack, err := loadSprite(jetpackName)
		if err != nil {
			panic(err)
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

		win.SetMatrix(cam)

		// Drawing to the screen
		win.SetSmooth(true)
		bg.Draw(win, pixel.IM.Moved(pixel.V(win.Bounds().Center().X, win.Bounds().Center().Y+766)).Scaled(pixel.ZV, 10))
		txt.Draw(win, pixel.IM)
		win.SetSmooth(false)
		c.Draw(win, pixel.IM)
		jetpack.Draw(win, jetMat)

	}

}

func main() {
	pixelgl.Run(run)
}

package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
)

// InstallShader ...
func InstallShader(c *pixelgl.Canvas, uTime *float32, uMouse, uGopherPos *mgl32.Vec2) {
	c.BindUniform("u_time", uTime)
	c.BindUniform("u_mouse", uMouse)
	c.BindUniform("u_gopherpos", uGopherPos)
	c.SetFragmentShader(cloudsFragmentShader)
	c.UpdateShader()
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

	var uTime float32
	var uMouse, uGopherPos mgl32.Vec2

	bounds := win.Bounds()
	bounds.Max = bounds.Max.ScaledXY(pixel.V(1.0, 1.0))

	canvas := pixelgl.NewCanvas(bounds)

	InstallShader(canvas, &uTime, &uMouse, &uGopherPos)

	bg, _ := loadSprite("sky.png")
	bg.Draw(canvas, pixel.IM)

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
	start := time.Now()

	// Game Loop
	for !win.Closed() {
		uTime = float32(time.Since(start).Seconds())
		mpos := win.MousePosition()
		uMouse[0] = float32(mpos.X)
		uMouse[1] = float32(mpos.Y)

		win.SetTitle(fmt.Sprint(uGopherPos))

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

		positionVector := pixel.V(win.Bounds().Center().X+jetX, win.Bounds().Center().Y+jetY-372)
		jetMat := pixel.IM
		jetMat = jetMat.Scaled(pixel.ZV, 4)
		jetMat = jetMat.Moved(positionVector)
		jetMat = jetMat.ScaledXY(positionVector, pixel.V(flipped, 1))
		jetMat = jetMat.Rotated(positionVector, radians)

		jetX += velX
		jetY += velY
		uGopherPos[0] = float32(jetX*0.01 + 1.0)
		uGopherPos[1] = float32(jetY*0.01 + 1.0)
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

		//bg.Draw(canvas, pixel.IM.Moved(pixel.V(win.Bounds().Center().X, win.Bounds().Center().Y+766)))
		bg.Draw(canvas, pixel.IM)

		canvas.Draw(win, pixel.IM.Moved(camVector))
		txt.Draw(win, pixel.IM)
		win.SetSmooth(false)
		currentSprite.Draw(win, jetMat)

		win.Update()
	}

}

func main() {
	pixelgl.Run(run)
}

var cloudsFragmentShader = `
#version 330 core

#ifdef GL_ES
precision highp float;
#endif

#define HOW_CLOUDY 0.4
#define SHADOW_THRESHOLD 0.2
#define SHADOW 0.2
#define SUBSURFACE 1.0
#define WIND_DIRECTION 5.0
#define TIME_SCALE 1.7
#define SCALE 0.3
#define ENABLE_SHAFTS
in vec2 texcoords;
out vec4 fragColor;
mat2 RM = mat2(cos(WIND_DIRECTION), -sin(WIND_DIRECTION), sin(WIND_DIRECTION), cos(WIND_DIRECTION));
uniform float u_time;
uniform vec2 u_mouse;
//uniform vec2 u_resolution;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform vec2 u_gopherpos;
 
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
    float ic = cloud(vec3(p * 2.0, u_time*0.01 * TIME_SCALE)) / HOW_CLOUDY;
    float init = smoothstep(0.1, 1.0, ic) * 10.0;
    shadow = smoothstep(0.0, SHADOW_THRESHOLD, ic) * SHADOW + (1.0 - SHADOW);
    init = (init * cloud(vec3(p * (6.0), u_time*0.01 * TIME_SCALE)) * ic);
    init = (init * (cloud(vec3(p * (11.0), u_time*0.01 * TIME_SCALE))*0.5 + 0.4) * init);
    return min(1.0, init);
}
//uniform sampler2D bb;
float cloudslowres(vec2 p){
    return 1.0 - (texture(u_texture, p).a - 0.9) * 10.0;
}
 
vec2 ratio = vec2(1.0, 1.0);
 
vec4 getresult(){
	vec2 uvmouse = (u_mouse/(texcoords - u_texbounds.xy));
	vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
    vec2 surfacePosition = ((( t ) * vec2(u_gopherpos.x , u_gopherpos.y)) * 2.0 - 1.0)*SCALE;
	vec2 position = ( surfacePosition);
	vec2 sun = (uvmouse.xy * vec2(texcoords.x / texcoords.y+1.1, 1.0)*2.0-1.0) * SCALE;
    
    float dst = distance(sun * ratio, position * ratio);
    float suni = pow(dst + 1.0, -10.0);
    float shaft =0.0;
    float st = 0.05;
    float w = 1.0;
    vec2 dir = sun - position;
    float c = clouds(position);
    #ifdef ENABLE_SHAFTS
    for(int i=0;i<50;i++){
        float occl = cloudslowres(clamp((t) + dir * st, 0.0, 1.0));
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
    fragColor = getresult().rgba;
}
`

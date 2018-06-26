package main

import (
	"image"
	"os"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	_ "image/png"
)

const (
	windowWidth  = 800
	windowHeight = 800
	// sprite tiles are squared, 64x64 size
	tileSize = 64
	f        = 0 // floor identifier
	w        = 1 // wall identifier
)

var levelData = [][]uint{
	{f, f, f, f, f, f}, // This row will be rendered in the lower left part of the screen (closer to the viewer)
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, f, f, f, f, w},
	{w, w, w, w, w, w}, // And this in the upper right
}
var win *pixelgl.Window
var canvas *pixelgl.Canvas
var offset = pixel.V(400, 325)
var floorTile, wallTile *pixel.Sprite

func installshader(c *pixelgl.Canvas, src string, uniformNameAndVar ...interface{}) {
	c.SetFragmentShader(src)
	for i := 0; i < len(uniformNameAndVar); i += 2 {
		c.BindUniform(
			uniformNameAndVar[i+0].(string),
			uniformNameAndVar[i+1],
		)
	}
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

func run() {
	var err error

	cfg := pixelgl.WindowConfig{
		Title:  "Isometric demo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic, err := loadPicture("castle.png")
	if err != nil {
		panic(err)
	}

	wallTile = pixel.NewSprite(pic, pixel.R(0, 448, tileSize, 512))
	floorTile = pixel.NewSprite(pic, pixel.R(0, 128, tileSize, 192))

	uResolution := mgl32.Vec2{float32(win.Bounds().W()), float32(win.Bounds().H())}
	var uTime float32
	var uMouse mgl32.Vec4

	canvas = pixelgl.NewCanvas(win.Bounds())
	wc := win.GetCanvas()

	installshader(wc, embossFragmentShader,
		"u_resolution", &uResolution,
		"u_time", &uTime,
		"u_mouse", &uMouse,
	)
	start := time.Now()
	for !win.Closed() {
		uTime = float32(time.Since(start).Seconds())

		uMouse[0] = float32(win.MousePosition().X)
		uMouse[1] = float32(win.MousePosition().Y)
		if win.Pressed(pixelgl.MouseButton1) {
			uMouse[2] = 1.0
		} else {
			uMouse[2] = 0.0
		}
		if win.Pressed(pixelgl.MouseButton2) {
			uMouse[3] = 1.0
		} else {
			uMouse[3] = 0.0
		}

		canvas.Clear(pixel.RGBA{R: 0.09, G: 0.05, B: 0.09, A: 1.0})
		depthSort()
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()
	}
}

// Draw level data tiles to window, from farthest to closest.
// In order to achieve the depth effect, we need to render tiles up to down, being lower
// closer to the viewer (see painter's algorithm). To do that, we need to process levelData in reverse order,
// so its first row is rendered last, as OpenGL considers its origin to be in the lower left corner of the display.
func depthSort() {
	for x := len(levelData) - 1; x >= 0; x-- {
		for y := len(levelData[x]) - 1; y >= 0; y-- {
			isoCoords := cartesianToIso(pixel.V(float64(x), float64(y)))
			mat := pixel.IM.Moved(offset.Add(isoCoords))
			// Not really needed, just put to show bigger blocks
			mat = mat.ScaledXY(win.Bounds().Center(), pixel.V(2, 2))
			tileType := levelData[x][y]
			if tileType == f {
				floorTile.Draw(canvas, mat)
			} else {
				wallTile.Draw(canvas, mat)
			}
		}
	}
}

func cartesianToIso(pt pixel.Vec) pixel.Vec {
	return pixel.V((pt.X-pt.Y)*(tileSize/2), (pt.X+pt.Y)*(tileSize/4))
}

func main() {
	pixelgl.Run(run)
}

var embossFragmentShader = `
#version 330 core

// Using a sobel filter to create a normal map and then applying simple lighting.

// This makes the darker areas less bumpy but I like it
#define USE_LINEAR_FOR_BUMPMAP

//#define SHOW_NORMAL_MAP
//#define SHOW_ALBEDO

in vec2 texcoords;

uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform float u_time;
uniform vec4 u_mouse;

out vec4 fragColor;

struct C_Sample
{
	vec3 vAlbedo;
	vec3 vNormal;
};
	
C_Sample SampleMaterial(const in vec2 uv, sampler2D sampler,  const in vec2 vTextureSize, const in float fNormalScale)
{
	C_Sample result;
	
	vec2 vInvTextureSize = vec2(1.0) / vTextureSize;
	
	vec3 cSampleNegXNegY = texture(sampler, uv + (vec2(-1.0, -1.0)) * vInvTextureSize.xy).rgb;
	vec3 cSampleZerXNegY = texture(sampler, uv + (vec2( 0.0, -1.0)) * vInvTextureSize.xy).rgb;
	vec3 cSamplePosXNegY = texture(sampler, uv + (vec2( 1.0, -1.0)) * vInvTextureSize.xy).rgb;
	
	vec3 cSampleNegXZerY = texture(sampler, uv + (vec2(-1.0, 0.0)) * vInvTextureSize.xy).rgb;
	vec3 cSampleZerXZerY = texture(sampler, uv + (vec2( 0.0, 0.0)) * vInvTextureSize.xy).rgb;
	vec3 cSamplePosXZerY = texture(sampler, uv + (vec2( 1.0, 0.0)) * vInvTextureSize.xy).rgb;
	
	vec3 cSampleNegXPosY = texture(sampler, uv + (vec2(-1.0,  1.0)) * vInvTextureSize.xy).rgb;
	vec3 cSampleZerXPosY = texture(sampler, uv + (vec2( 0.0,  1.0)) * vInvTextureSize.xy).rgb;
	vec3 cSamplePosXPosY = texture(sampler, uv + (vec2( 1.0,  1.0)) * vInvTextureSize.xy).rgb;

	// convert to linear	
	vec3 cLSampleNegXNegY = cSampleNegXNegY * cSampleNegXNegY;
	vec3 cLSampleZerXNegY = cSampleZerXNegY * cSampleZerXNegY;
	vec3 cLSamplePosXNegY = cSamplePosXNegY * cSamplePosXNegY;

	vec3 cLSampleNegXZerY = cSampleNegXZerY * cSampleNegXZerY;
	vec3 cLSampleZerXZerY = cSampleZerXZerY * cSampleZerXZerY;
	vec3 cLSamplePosXZerY = cSamplePosXZerY * cSamplePosXZerY;

	vec3 cLSampleNegXPosY = cSampleNegXPosY * cSampleNegXPosY;
	vec3 cLSampleZerXPosY = cSampleZerXPosY * cSampleZerXPosY;
	vec3 cLSamplePosXPosY = cSamplePosXPosY * cSamplePosXPosY;

	// Average samples to get albdeo colour
	result.vAlbedo = ( cLSampleNegXNegY + cLSampleZerXNegY + cLSamplePosXNegY 
		    	     + cLSampleNegXZerY + cLSampleZerXZerY + cLSamplePosXZerY
		    	     + cLSampleNegXPosY + cLSampleZerXPosY + cLSamplePosXPosY ) / 9.0;	
	
	vec3 vScale = vec3(0.3333);
	
	#ifdef USE_LINEAR_FOR_BUMPMAP
		
		float fSampleNegXNegY = dot(cLSampleNegXNegY, vScale);
		float fSampleZerXNegY = dot(cLSampleZerXNegY, vScale);
		float fSamplePosXNegY = dot(cLSamplePosXNegY, vScale);
		
		float fSampleNegXZerY = dot(cLSampleNegXZerY, vScale);
		float fSampleZerXZerY = dot(cLSampleZerXZerY, vScale);
		float fSamplePosXZerY = dot(cLSamplePosXZerY, vScale);
		
		float fSampleNegXPosY = dot(cLSampleNegXPosY, vScale);
		float fSampleZerXPosY = dot(cLSampleZerXPosY, vScale);
		float fSamplePosXPosY = dot(cLSamplePosXPosY, vScale);
	
	#else
	
		float fSampleNegXNegY = dot(cSampleNegXNegY, vScale);
		float fSampleZerXNegY = dot(cSampleZerXNegY, vScale);
		float fSamplePosXNegY = dot(cSamplePosXNegY, vScale);
		
		float fSampleNegXZerY = dot(cSampleNegXZerY, vScale);
		float fSampleZerXZerY = dot(cSampleZerXZerY, vScale);
		float fSamplePosXZerY = dot(cSamplePosXZerY, vScale);
		
		float fSampleNegXPosY = dot(cSampleNegXPosY, vScale);
		float fSampleZerXPosY = dot(cSampleZerXPosY, vScale);
		float fSamplePosXPosY = dot(cSamplePosXPosY, vScale);	
	
	#endif
	
	// Sobel operator - http://en.wikipedia.org/wiki/Sobel_operator
	
	vec2 vEdge;
	vEdge.x = (fSampleNegXNegY - fSamplePosXNegY) * 0.25 
			+ (fSampleNegXZerY - fSamplePosXZerY) * 0.5
			+ (fSampleNegXPosY - fSamplePosXPosY) * 0.25;

	vEdge.y = (fSampleNegXNegY - fSampleNegXPosY) * 0.25 
			+ (fSampleZerXNegY - fSampleZerXPosY) * 0.5
			+ (fSamplePosXNegY - fSamplePosXPosY) * 0.25;

	result.vNormal = normalize(vec3(vEdge * fNormalScale, 1.0));	
	
	return result;
}

void main()
{	
	vec2 uv = gl_FragCoord.xy / u_resolution.xy;
	
	C_Sample materialSample;
		
	float fNormalScale = 5.0;
	materialSample = SampleMaterial( uv, u_texture, u_resolution.xy, fNormalScale );
	
	// Random Lighting...
	
	float fLightHeight = 0.2;
	float fViewHeight = 1.0;
	
	vec3 vSurfacePos = vec3(uv, 0.0);
	
	vec3 vViewPos = vec3(0.5, 0.5, fViewHeight);
			
	vec3 vLightPos = vec3( vec2(sin(u_time),cos(u_time)) * 0.25 + 0.5 , fLightHeight);
		
	if( u_mouse.z > 0.0 )
	{
		vLightPos = vec3(u_mouse.xy / u_resolution.xy, fLightHeight);
	}
	
	vec3 vDirToView = normalize( vViewPos - vSurfacePos );
	vec3 vDirToLight = normalize( vLightPos - vSurfacePos );
		
	float fNDotL = clamp( dot(materialSample.vNormal, vDirToLight), 0.0, 1.0);
	float fDiffuse = fNDotL;
	
	vec3 vHalf = normalize( vDirToView + vDirToLight );
	float fNDotH = clamp( dot(materialSample.vNormal, vHalf), 0.0, 1.0);
	float fSpec = pow(fNDotH, 10.0) * fNDotL * 0.5;
	
	vec3 vResult = materialSample.vAlbedo * fDiffuse + fSpec;
	
	vResult = sqrt(vResult);
	
	#ifdef SHOW_NORMAL_MAP
	vResult = materialSample.vNormal * 0.5 + 0.5;
	#endif
	
	#ifdef SHOW_ALBEDO
	vResult = sqrt(materialSample.vAlbedo);
	#endif
	
	fragColor = vec4(vResult,1.0);
}
`

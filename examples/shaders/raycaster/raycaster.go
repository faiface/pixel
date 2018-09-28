package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const texSize = 64

var (
	fullscreen   = false
	showMap      = true
	width        = 640
	height       = 400
	scale        = 1.0
	wallDistance = 8.0

	as actionSquare

	pos, dir, plane pixel.Vec

	textures = loadTextures()
)

func setup() {
	pos = pixel.V(12.0, 14.5)
	dir = pixel.V(-1.0, 0.0)
	plane = pixel.V(0.0, 0.66)
}

var world = [24][24]int{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 0, 0, 0, 0, 3, 0, 3, 0, 3, 0, 0, 0, 1},
	{1, 0, 0, 0, 2, 7, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 2, 0, 2, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, 7, 0, 3, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 2, 2, 2, 2, 0, 2, 2, 0, 0, 0, 0, 3, 0, 3, 0, 3, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 4, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 6, 0, 4, 0, 0, 0, 4, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 5, 0, 0, 0, 1},
	{1, 0, 6, 0, 4, 0, 7, 0, 4, 0, 0, 0, 0, 0, 5, 0, 0, 0, 5, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 4, 0, 0, 0, 4, 0, 0, 0, 0, 5, 5, 5, 5, 5, 5, 5, 0, 0, 0, 1},
	{1, 4, 4, 4, 4, 4, 4, 0, 4, 0, 0, 0, 5, 5, 0, 5, 5, 5, 0, 5, 5, 0, 0, 1},
	{1, 4, 0, 0, 0, 0, 0, 0, 4, 0, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 0, 1},
	{1, 4, 0, 4, 0, 0, 0, 0, 4, 0, 0, 5, 0, 5, 5, 5, 5, 5, 5, 5, 0, 5, 0, 1},
	{1, 4, 0, 4, 4, 4, 4, 4, 4, 0, 0, 5, 0, 5, 0, 0, 0, 0, 0, 5, 0, 5, 0, 1},
	{1, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 5, 0, 5, 5, 0, 0, 0, 0, 1},
	{1, 4, 4, 4, 4, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

func loadTextures() *image.RGBA {
	p, err := png.Decode(bytes.NewReader(textureData))
	if err != nil {
		panic(err)
	}

	m := image.NewRGBA(p.Bounds())

	draw.Draw(m, m.Bounds(), p, image.ZP, draw.Src)

	return m
}

func getTexNum(x, y int) int {
	return world[x][y]
}

func getColor(x, y int) color.RGBA {
	switch world[x][y] {
	case 0:
		return color.RGBA{43, 30, 24, 255}
	case 1:
		return color.RGBA{100, 89, 73, 255}
	case 2:
		return color.RGBA{110, 23, 0, 255}
	case 3:
		return color.RGBA{45, 103, 171, 255}
	case 4:
		return color.RGBA{123, 84, 33, 255}
	case 5:
		return color.RGBA{158, 148, 130, 255}
	case 6:
		return color.RGBA{203, 161, 47, 255}
	case 7:
		return color.RGBA{255, 107, 0, 255}
	case 9:
		return color.RGBA{0, 0, 0, 0}
	default:
		return color.RGBA{255, 194, 32, 255}
	}
}

func frame() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		var (
			step         image.Point
			sideDist     pixel.Vec
			perpWallDist float64
			hit, side    bool

			rayPos, worldX, worldY = pos, int(pos.X), int(pos.Y)

			cameraX = 2*float64(x)/float64(width) - 1

			rayDir = pixel.V(
				dir.X+plane.X*cameraX,
				dir.Y+plane.Y*cameraX,
			)

			deltaDist = pixel.V(
				math.Sqrt(1.0+(rayDir.Y*rayDir.Y)/(rayDir.X*rayDir.X)),
				math.Sqrt(1.0+(rayDir.X*rayDir.X)/(rayDir.Y*rayDir.Y)),
			)
		)

		if rayDir.X < 0 {
			step.X = -1
			sideDist.X = (rayPos.X - float64(worldX)) * deltaDist.X
		} else {
			step.X = 1
			sideDist.X = (float64(worldX) + 1.0 - rayPos.X) * deltaDist.X
		}

		if rayDir.Y < 0 {
			step.Y = -1
			sideDist.Y = (rayPos.Y - float64(worldY)) * deltaDist.Y
		} else {
			step.Y = 1
			sideDist.Y = (float64(worldY) + 1.0 - rayPos.Y) * deltaDist.Y
		}

		for !hit {
			if sideDist.X < sideDist.Y {
				sideDist.X += deltaDist.X
				worldX += step.X
				side = false
			} else {
				sideDist.Y += deltaDist.Y
				worldY += step.Y
				side = true
			}

			if world[worldX][worldY] > 0 {
				hit = true
			}
		}

		var wallX float64

		if side {
			perpWallDist = (float64(worldY) - rayPos.Y + (1-float64(step.Y))/2) / rayDir.Y
			wallX = rayPos.X + perpWallDist*rayDir.X
		} else {
			perpWallDist = (float64(worldX) - rayPos.X + (1-float64(step.X))/2) / rayDir.X
			wallX = rayPos.Y + perpWallDist*rayDir.Y
		}

		if x == width/2 {
			wallDistance = perpWallDist
		}

		wallX -= math.Floor(wallX)

		texX := int(wallX * float64(texSize))

		lineHeight := int(float64(height) / perpWallDist)

		if lineHeight < 1 {
			lineHeight = 1
		}

		drawStart := -lineHeight/2 + height/2
		if drawStart < 0 {
			drawStart = 0
		}

		drawEnd := lineHeight/2 + height/2
		if drawEnd >= height {
			drawEnd = height - 1
		}

		if !side && rayDir.X > 0 {
			texX = texSize - texX - 1
		}

		if side && rayDir.Y < 0 {
			texX = texSize - texX - 1
		}

		texNum := getTexNum(worldX, worldY)

		for y := drawStart; y < drawEnd+1; y++ {
			d := y*256 - height*128 + lineHeight*128
			texY := ((d * texSize) / lineHeight) / 256

			c := textures.RGBAAt(
				texX+texSize*(texNum),
				texY%texSize,
			)

			if side {
				c.R = c.R / 2
				c.G = c.G / 2
				c.B = c.B / 2
			}

			m.Set(x, y, c)
		}

		var floorWall pixel.Vec

		if !side && rayDir.X > 0 {
			floorWall.X = float64(worldX)
			floorWall.Y = float64(worldY) + wallX
		} else if !side && rayDir.X < 0 {
			floorWall.X = float64(worldX) + 1.0
			floorWall.Y = float64(worldY) + wallX
		} else if side && rayDir.Y > 0 {
			floorWall.X = float64(worldX) + wallX
			floorWall.Y = float64(worldY)
		} else {
			floorWall.X = float64(worldX) + wallX
			floorWall.Y = float64(worldY) + 1.0
		}

		distWall, distPlayer := perpWallDist, 0.0

		for y := drawEnd + 1; y < height; y++ {
			currentDist := float64(height) / (2.0*float64(y) - float64(height))

			weight := (currentDist - distPlayer) / (distWall - distPlayer)

			currentFloor := pixel.V(
				weight*floorWall.X+(1.0-weight)*pos.X,
				weight*floorWall.Y+(1.0-weight)*pos.Y,
			)

			fx := int(currentFloor.X*float64(texSize)) % texSize
			fy := int(currentFloor.Y*float64(texSize)) % texSize

			m.Set(x, y, textures.At(fx, fy))

			m.Set(x, height-y-1, textures.At(fx+(4*texSize), fy))
			m.Set(x, height-y, textures.At(fx+(4*texSize), fy))
		}
	}

	return m
}

func minimap() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, 24, 26))

	for x, row := range world {
		for y := range row {
			c := getColor(x, y)
			if c.A == 255 {
				c.A = 96
			}
			m.Set(x, y, c)
		}
	}

	m.Set(int(pos.X), int(pos.Y), color.RGBA{255, 0, 0, 255})

	if as.active {
		m.Set(as.X, as.Y, color.RGBA{255, 255, 255, 255})
	} else {
		m.Set(as.X, as.Y, color.RGBA{64, 64, 64, 255})
	}

	return m
}

func getActionSquare() actionSquare {
	pt := image.Pt(int(pos.X)+1, int(pos.Y))

	a := dir.Angle()

	switch {
	case a > 2.8 || a < -2.8:
		pt = image.Pt(int(pos.X)-1, int(pos.Y))
	case a > -2.8 && a < -2.2:
		pt = image.Pt(int(pos.X)-1, int(pos.Y)-1)
	case a > -2.2 && a < -1.4:
		pt = image.Pt(int(pos.X), int(pos.Y)-1)
	case a > -1.4 && a < -0.7:
		pt = image.Pt(int(pos.X)+1, int(pos.Y)-1)
	case a > 0.4 && a < 1.0:
		pt = image.Pt(int(pos.X)+1, int(pos.Y)+1)
	case a > 1.0 && a < 1.7:
		pt = image.Pt(int(pos.X), int(pos.Y)+1)
	case a > 1.7:
		pt = image.Pt(int(pos.X)-1, int(pos.Y)+1)
	}

	block := -1
	active := pt.X > 0 && pt.X < 23 && pt.Y > 0 && pt.Y < 23

	if active {
		block = world[pt.X][pt.Y]
	}

	return actionSquare{
		X:      pt.X,
		Y:      pt.Y,
		active: active,
		block:  block,
	}
}

type actionSquare struct {
	X      int
	Y      int
	block  int
	active bool
}

func (as actionSquare) toggle(n int) {
	if as.active {
		if world[as.X][as.Y] == 0 {
			world[as.X][as.Y] = n
		} else {
			world[as.X][as.Y] = 0
		}
	}
}

func (as actionSquare) set(n int) {
	if as.active {
		world[as.X][as.Y] = n
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Bounds:      pixel.R(0, 0, float64(width)*scale, float64(height)*scale),
		VSync:       true,
		Undecorated: false,
	}

	if fullscreen {
		cfg.Monitor = pixelgl.PrimaryMonitor()
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	centerScreenPos(win)

	c := win.Bounds().Center()

	last := time.Now()

	mapRot := -1.6683362599999894

	canvas := pixelgl.NewCanvas(win.Bounds())

	wc := win.GetCanvas()
	wc.SetFragmentShader(fxaaFragShader)

	var uAmount, uTime float32
	var uResolution, uMouse mgl32.Vec2
	var uParams mgl32.Vec3

	uParams[0] = 10.0
	uParams[1] = 0.8
	uParams[2] = 0.1

	uAmount = 0

	wc.BindUniform("u_amount", &uAmount)
	wc.BindUniform("u_resolution", &uResolution)
	wc.BindUniform("u_mouse", &uMouse)
	wc.BindUniform("u_time", &uTime)
	wc.BindUniform("u_params", &uParams)

	wc.UpdateShader()

	uResolution[0] = float32(win.Bounds().W())
	uResolution[1] = float32(win.Bounds().H())

	start := time.Now()
	for !win.Closed() {
		if win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ) {
			return
		}
		uTime = float32(time.Since(start).Seconds())
		uMouse[0] = float32(win.MousePosition().X)
		uMouse[1] = float32(win.MousePosition().Y)

		win.Clear(color.Black)
		canvas.Clear(color.Black)

		dt := time.Since(last).Seconds()
		last = time.Now()

		as = getActionSquare()
		if win.Pressed(pixelgl.KeyEqual) {
			uAmount++
		}
		if win.Pressed(pixelgl.KeyMinus) {
			uAmount = 0
		}
		if win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW) {
			moveForward(3.5 * dt)
		}

		if win.Pressed(pixelgl.KeyA) {
			moveLeft(3.5 * dt)
		}

		if win.Pressed(pixelgl.KeyDown) || win.Pressed(pixelgl.KeyS) {
			moveBackwards(3.5 * dt)
		}

		if win.Pressed(pixelgl.KeyD) {
			moveRight(3.5 * dt)
		}

		if win.Pressed(pixelgl.KeyRight) {
			turnRight(1.2 * dt)
		}

		if win.Pressed(pixelgl.KeyLeft) {
			turnLeft(1.2 * dt)
		}

		if win.JustPressed(pixelgl.KeyM) {
			showMap = !showMap
		}

		if win.JustPressed(pixelgl.Key1) {
			as.set(1)
		}

		if win.JustPressed(pixelgl.Key2) {
			as.set(2)
		}

		if win.JustPressed(pixelgl.Key3) {
			as.set(3)
		}

		if win.JustPressed(pixelgl.Key4) {
			as.set(4)
		}

		if win.JustPressed(pixelgl.Key5) {
			as.set(5)
		}

		if win.JustPressed(pixelgl.Key6) {
			as.set(6)
		}

		if win.JustPressed(pixelgl.Key7) {
			as.set(7)
		}

		if win.JustPressed(pixelgl.Key0) {
			as.set(0)
		}

		if win.JustPressed(pixelgl.KeySpace) {
			as.toggle(3)
		}

		p := pixel.PictureDataFromImage(frame())

		pixel.NewSprite(p, p.Bounds()).
			Draw(canvas, pixel.IM.Moved(c).Scaled(c, scale))

		if showMap {
			m := pixel.PictureDataFromImage(minimap())

			mc := m.Bounds().Min.Add(pixel.V(-m.Rect.W(), m.Rect.H()))

			pixel.NewSprite(m, m.Bounds()).
				Draw(canvas, pixel.IM.
					Moved(mc).
					Rotated(mc, mapRot).
					ScaledXY(pixel.ZV, pixel.V(-scale*2, scale*2)))
		}

		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()
	}
}

func moveForward(s float64) {
	if wallDistance > 0.3 {
		if world[int(pos.X+dir.X*s)][int(pos.Y)] == 0 {
			pos.X += dir.X * s
		}

		if world[int(pos.X)][int(pos.Y+dir.Y*s)] == 0 {
			pos.Y += dir.Y * s
		}
	}
}

func moveLeft(s float64) {
	if world[int(pos.X-plane.X*s)][int(pos.Y)] == 0 {
		pos.X -= plane.X * s
	}

	if world[int(pos.X)][int(pos.Y-plane.Y*s)] == 0 {
		pos.Y -= plane.Y * s
	}
}

func moveBackwards(s float64) {
	if world[int(pos.X-dir.X*s)][int(pos.Y)] == 0 {
		pos.X -= dir.X * s
	}

	if world[int(pos.X)][int(pos.Y-dir.Y*s)] == 0 {
		pos.Y -= dir.Y * s
	}
}

func moveRight(s float64) {
	if world[int(pos.X+plane.X*s)][int(pos.Y)] == 0 {
		pos.X += plane.X * s
	}

	if world[int(pos.X)][int(pos.Y+plane.Y*s)] == 0 {
		pos.Y += plane.Y * s
	}
}

func turnRight(s float64) {
	oldDirX := dir.X

	dir.X = dir.X*math.Cos(-s) - dir.Y*math.Sin(-s)
	dir.Y = oldDirX*math.Sin(-s) + dir.Y*math.Cos(-s)

	oldPlaneX := plane.X

	plane.X = plane.X*math.Cos(-s) - plane.Y*math.Sin(-s)
	plane.Y = oldPlaneX*math.Sin(-s) + plane.Y*math.Cos(-s)
}

func turnLeft(s float64) {
	oldDirX := dir.X

	dir.X = dir.X*math.Cos(s) - dir.Y*math.Sin(s)
	dir.Y = oldDirX*math.Sin(s) + dir.Y*math.Cos(s)

	oldPlaneX := plane.X

	plane.X = plane.X*math.Cos(s) - plane.Y*math.Sin(s)
	plane.Y = oldPlaneX*math.Sin(s) + plane.Y*math.Cos(s)
}

func main() {
	flag.BoolVar(&fullscreen, "f", fullscreen, "fullscreen")
	flag.IntVar(&width, "w", width, "width")
	flag.IntVar(&height, "h", height, "height")
	flag.Float64Var(&scale, "s", scale, "scale")
	flag.Parse()

	setup()

	pixelgl.Run(run)
}

func centerScreenPos(window *pixelgl.Window) {
	width, height := pixelgl.PrimaryMonitor().Size()
	window.SetPos(
		pixel.V(
			width/2-(window.Bounds().W()/2),
			height/2-(window.Bounds().H()/2),
		))
}

var shockwaveFragShader = `
#version 330 core

in vec2 texcoords;

uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform float u_amount;
uniform float u_time;
uniform vec2 u_mouse;
uniform vec3 u_params; // 10.0, 0.8, 0.1

out vec4 fragColor;

vec2 getMx() {
	vec2 mx = u_mouse / u_resolution.xy;
    // correct aspect ratio
	mx.y *= u_resolution.y / u_resolution.x;
	mx.y += (u_resolution.x - u_resolution.y) / u_resolution.x / 2.0;
    // centering
    // mx -= 0.5;
    // mx *= vec2(1.0, -1.0);
	return mx;
}

void main() 
{ 
  vec2 uv = (texcoords - u_texbounds.xy) / u_texbounds.zw;
  vec2 texCoord = uv;
  vec2 um = getMx();
  float distance = distance(uv, um);
  if ( (distance <= (u_time + u_params.z)) && 
       (distance >= (u_time - u_params.z)) ) 
  {
    float diff = (distance - u_time); 
    float powDiff = 1.0 - pow(abs(diff*u_params.x), 
                                u_params.y); 
    float diffu_time = diff  * powDiff; 
    vec2 diffUV = normalize(uv - um); 
    texCoord = uv + (diffUV * diffu_time);
  } 
  fragColor = texture(u_texture, texCoord);
}
`

var fxaaFragShader = `
#version 330 core

in vec2 texcoords;

uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform vec2 u_resolution;
uniform float u_amount;
uniform float u_time;
uniform vec2 u_mouse;

out vec4 fragColor;

#define FXAA_REDUCE_MIN   (1.0/128.0)
#define FXAA_REDUCE_MUL   (1.0/8.0)
#define FXAA_SPAN_MAX     8.0

void main() {
    vec2 uv = (texcoords - u_texbounds.xy) / u_texbounds.zw;
	vec2 res = 1. / u_resolution;
	
	if (u_amount < 1 ) {
		fragColor = texture(u_texture, uv);
		return;
	}

    vec3 rgbNW = texture( u_texture, ( uv.xy + vec2( -1.0, -1.0 ) * res ) ).xyz;
    vec3 rgbNE = texture( u_texture, ( uv.xy + vec2( 1.0, -1.0 ) * res ) ).xyz;
    vec3 rgbSW = texture( u_texture, ( uv.xy + vec2( -1.0, 1.0 ) * res ) ).xyz;
    vec3 rgbSE = texture( u_texture, ( uv.xy + vec2( 1.0, 1.0 ) * res ) ).xyz;
    vec4 rgbaM  = texture( u_texture,  uv.xy  * res );
    vec3 rgbM  = rgbaM.xyz;
    vec3 luma = vec3( 0.299, 0.587, 0.114 );

    float lumaNW = dot( rgbNW, luma );
    float lumaNE = dot( rgbNE, luma );
    float lumaSW = dot( rgbSW, luma );
    float lumaSE = dot( rgbSE, luma );
    float lumaM  = dot( rgbM,  luma );
    float lumaMin = min( lumaM, min( min( lumaNW, lumaNE ), min( lumaSW, lumaSE ) ) );
    float lumaMax = max( lumaM, max( max( lumaNW, lumaNE) , max( lumaSW, lumaSE ) ) );

    vec2 dir;
    dir.x = -((lumaNW + lumaNE) - (lumaSW + lumaSE));
    dir.y =  ((lumaNW + lumaSW) - (lumaNE + lumaSE));

    float dirReduce = max( ( lumaNW + lumaNE + lumaSW + lumaSE ) * ( 0.25 * FXAA_REDUCE_MUL ), FXAA_REDUCE_MIN );

    float rcpDirMin = 1.0 / ( min( abs( dir.x ), abs( dir.y ) ) + dirReduce );
    dir = min( vec2( FXAA_SPAN_MAX,  FXAA_SPAN_MAX),
          max( vec2(-FXAA_SPAN_MAX, -FXAA_SPAN_MAX),
                dir * rcpDirMin)) * res;
    vec4 rgbA = (1.0/2.0) * (
    texture(u_texture,  uv.xy + dir * (1.0/3.0 - 0.5)) +
    texture(u_texture,  uv.xy + dir * (2.0/3.0 - 0.5)));
    vec4 rgbB = rgbA * (1.0/2.0) + (1.0/4.0) * (
    texture(u_texture,  uv.xy + dir * (0.0/3.0 - 0.5)) +
    texture(u_texture,  uv.xy + dir * (3.0/3.0 - 0.5)));
    float lumaB = dot(rgbB, vec4(luma, 0.0));

    if ( ( lumaB < lumaMin ) || ( lumaB > lumaMax ) ) {
        fragColor = rgbA;
    } else {
        fragColor = rgbB;
    }

    //fragColor = vec4( texture( u_texture,uv ).xyz, 1. );
}
`

var toonFragShader = `
#version 330 core

in vec4 Color;
in vec2 texcoords;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform float u_amount;
uniform float u_time;
uniform vec2 u_mouse;
uniform vec2 u_resolution;

#define HueLevCount 6
#define SatLevCount 7
#define ValLevCount 4

float HueLevels[HueLevCount];
float SatLevels[SatLevCount];
float ValLevels[ValLevCount];

vec3 RGBtoHSV( float r, float g, float b) {
   float minv, maxv, delta;
   vec3 res;

   minv = min(min(r, g), b);
   maxv = max(max(r, g), b);
   res.z = maxv;            // v

   delta = maxv - minv;

   if( maxv != 0.0 )
      res.y = delta / maxv;      // s
   else {
      // r = g = b = 0      // s = 0, v is undefined
      res.y = 0.0;
      res.x = -1.0;
      return res;
   }

   if( r == maxv )
      res.x = ( g - b ) / delta;      // between yellow & magenta
   else if( g == maxv )
      res.x = 2.0 + ( b - r ) / delta;   // between cyan & yellow
   else
      res.x = 4.0 + ( r - g ) / delta;   // between magenta & cyan

   res.x = res.x * 60.0;            // degrees
   if( res.x < 0.0 )
      res.x = res.x + 360.0;

   return res;
}

vec3 HSVtoRGB(float h, float s, float v ) {
   int i;
   float f, p, q, t;
   vec3 res;

   if( s == 0.0 ) {
      // achromatic (grey)
      res.x = v;
      res.y = v;
      res.z = v;
      return res;
   }

   h /= 60.0;         // sector 0 to 5
   i = int(floor( h ));
   f = h - float(i);         // factorial part of h
   p = v * ( 1.0 - s );
   q = v * ( 1.0 - s * f );
   t = v * ( 1.0 - s * ( 1.0 - f ) );

   if (i==0) {
		res.x = v;
		res.y = t;
		res.z = p;
   	} else if (i==1) {
         res.x = q;
         res.y = v;
         res.z = p;
	} else if (i==2) {
         res.x = p;
         res.y = v;
         res.z = t;
	} else if (i==3) {
         res.x = p;
         res.y = q;
         res.z = v;
	} else if (i==4) {
         res.x = t;
         res.y = p;
         res.z = v;
	} else if (i==5) {
         res.x = v;
         res.y = p;
         res.z = q;
   }

   return res;
}

float nearestLevel(float col, int mode) {

   if (mode==0) {
   		for (int i =0; i<HueLevCount-1; i++ ) {
		    if (col >= HueLevels[i] && col <= HueLevels[i+1]) {
		      return HueLevels[i+1];
		    }
		}
	 }

	if (mode==1) {
		for (int i =0; i<SatLevCount-1; i++ ) {
			if (col >= SatLevels[i] && col <= SatLevels[i+1]) {
	          return SatLevels[i+1];
	        }
		}
	}


	if (mode==2) {
		for (int i =0; i<ValLevCount-1; i++ ) {
			if (col >= ValLevels[i] && col <= ValLevels[i+1]) {
	          return ValLevels[i+1];
	        }
		}
	}


}

// averaged pixel intensity from 3 color channels
float avg_intensity(vec4 pix) {
 return (pix.r + pix.g + pix.b)/3.;
}

vec4 get_pixel(vec2 coords, float dx, float dy) {
 return texture(u_texture, coords + vec2(dx, dy));
}

// returns pixel color
float IsEdge(in vec2 coords){
  float dxtex = 1.0 / u_resolution.x ;
  float dytex = 1.0 / u_resolution.y ;

  float pix[9];

  int k = -1;
  float delta;

  // read neighboring pixel intensities
float pix0 = avg_intensity(get_pixel(coords,-1.0*dxtex, -1.0*dytex));
float pix1 = avg_intensity(get_pixel(coords,-1.0*dxtex, 0.0*dytex));
float pix2 = avg_intensity(get_pixel(coords,-1.0*dxtex, 1.0*dytex));
float pix3 = avg_intensity(get_pixel(coords,0.0*dxtex, -1.0*dytex));
float pix4 = avg_intensity(get_pixel(coords,0.0*dxtex, 0.0*dytex));
float pix5 = avg_intensity(get_pixel(coords,0.0*dxtex, 1.0*dytex));
float pix6 = avg_intensity(get_pixel(coords,1.0*dxtex, -1.0*dytex));
float pix7 = avg_intensity(get_pixel(coords,1.0*dxtex, 0.0*dytex));
float pix8 = avg_intensity(get_pixel(coords,1.0*dxtex, 1.0*dytex));
  // average color differences around neighboring pixels
  delta = (abs(pix1-pix7)+
          abs(pix5-pix3) +
          abs(pix0-pix8)+
          abs(pix2-pix6)
           )/4.;

  return clamp(5.5*delta,0.0,1.0);
}

void main(void)
{
    vec2 uv = (texcoords - u_texbounds.xy) / u_texbounds.zw;

	HueLevels[0] = 0.0;
	HueLevels[1] = 80.0;
	HueLevels[2] = 160.0;
	HueLevels[3] = 240.0;
	HueLevels[4] = 320.0;
	HueLevels[5] = 360.0;

	SatLevels[0] = 0.0;
	SatLevels[1] = 0.1;
	SatLevels[2] = 0.3;
	SatLevels[3] = 0.5;
	SatLevels[4] = 0.6;
	SatLevels[5] = 0.8;
	SatLevels[6] = 1.0;

	ValLevels[0] = 0.0;
	ValLevels[1] = 0.3;
	ValLevels[2] = 0.6;
	ValLevels[3] = 1.0;

    vec4 colorOrg = texture( u_texture, uv );
    vec3 vHSV =  RGBtoHSV(colorOrg.r,colorOrg.g,colorOrg.b);
    vHSV.x = nearestLevel(vHSV.x, 0);
    vHSV.y = nearestLevel(vHSV.y, 1);
    vHSV.z = nearestLevel(vHSV.z, 2);
    float edg = IsEdge(uv);
    vec3 vRGB = (edg >= 0.3)? vec3(0.0,0.0,0.0):HSVtoRGB(vHSV.x,vHSV.y,vHSV.z);
    fragColor = vec4(vRGB.x,vRGB.y,vRGB.z,1.0);
}
`

var textureData = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 2, 0, 0, 0, 0, 64, 8, 3, 0, 0, 0, 91, 97, 63, 141, 0, 0, 2, 100, 80, 76, 84, 69, 182, 182, 182, 48, 38, 29, 33, 23, 15, 191, 146, 37, 7, 3, 1, 72, 64, 49, 40, 28, 20, 37, 26, 19, 42, 31, 22, 110, 100, 84, 64, 54, 40, 55, 46, 32, 142, 131, 114, 74, 40, 10, 22, 14, 7, 97, 86, 69, 66, 35, 6, 99, 60, 20, 102, 91, 75, 16, 9, 3, 120, 27, 1, 27, 18, 11, 184, 140, 33, 83, 72, 57, 131, 119, 103, 114, 73, 27, 128, 87, 34, 48, 23, 2, 159, 149, 134, 185, 185, 185, 6, 37, 82, 68, 59, 45, 130, 188, 209, 111, 72, 8, 129, 32, 4, 134, 93, 38, 150, 139, 123, 46, 33, 27, 80, 68, 54, 199, 156, 43, 87, 75, 61, 43, 43, 43, 123, 111, 95, 60, 50, 38, 92, 55, 17, 119, 79, 31, 170, 159, 142, 82, 46, 14, 170, 125, 23, 146, 135, 118, 164, 119, 20, 189, 189, 189, 31, 80, 138, 115, 104, 89, 203, 162, 47, 1, 0, 0, 127, 115, 99, 87, 49, 15, 193, 193, 193, 1, 27, 64, 40, 96, 165, 83, 83, 83, 54, 42, 36, 42, 100, 171, 154, 143, 128, 118, 108, 92, 57, 57, 56, 109, 69, 25, 137, 127, 108, 206, 167, 50, 89, 80, 63, 224, 211, 196, 0, 9, 27, 255, 255, 255, 79, 44, 1, 3, 31, 71, 92, 82, 66, 71, 71, 71, 22, 61, 110, 19, 54, 100, 198, 198, 198, 189, 176, 161, 1, 22, 57, 117, 78, 9, 93, 19, 0, 29, 76, 132, 99, 22, 1, 111, 26, 2, 35, 87, 150, 27, 71, 126, 78, 69, 57, 127, 127, 127, 160, 114, 18, 0, 5, 16, 1, 18, 49, 0, 12, 38, 134, 124, 105, 178, 133, 28, 34, 33, 33, 105, 25, 2, 139, 123, 109, 135, 38, 7, 27, 67, 116, 105, 66, 24, 37, 91, 157, 72, 58, 46, 174, 129, 26, 59, 30, 5, 33, 83, 144, 132, 125, 112, 124, 83, 32, 148, 130, 114, 45, 104, 178, 78, 62, 51, 195, 152, 40, 23, 54, 90, 153, 137, 123, 81, 20, 2, 38, 15, 1, 184, 173, 157, 175, 163, 148, 9, 44, 95, 172, 119, 47, 28, 63, 102, 210, 171, 54, 95, 56, 3, 94, 84, 72, 83, 73, 66, 207, 195, 180, 201, 189, 173, 179, 168, 152, 28, 10, 0, 155, 155, 155, 97, 87, 80, 139, 100, 42, 152, 55, 18, 144, 46, 11, 164, 155, 138, 123, 123, 123, 55, 27, 4, 203, 203, 203, 214, 203, 187, 87, 86, 86, 164, 112, 46, 155, 103, 39, 50, 116, 197, 106, 106, 106, 57, 133, 225, 44, 107, 183, 52, 121, 205, 86, 49, 3, 54, 126, 214, 147, 103, 16, 147, 95, 36, 68, 16, 1, 140, 98, 16, 153, 109, 18, 145, 102, 45, 169, 150, 136, 194, 181, 166, 48, 112, 190, 169, 73, 36, 84, 7, 0, 19, 47, 78, 169, 120, 61, 114, 114, 114, 131, 91, 14, 123, 84, 13, 169, 170, 170, 64, 63, 62, 137, 137, 137, 100, 64, 7, 56, 11, 0, 100, 100, 100, 209, 209, 209, 178, 130, 76, 50, 50, 50, 43, 87, 171, 95, 95, 95, 9, 51, 110, 87, 79, 70, 110, 159, 225, 105, 146, 167, 15, 39, 62, 13, 31, 46, 174, 175, 175, 213, 144, 111, 95, 143, 212, 189, 90, 48, 164, 164, 163, 66, 148, 245, 70, 114, 191, 116, 160, 179, 141, 143, 144, 220, 155, 122, 222, 224, 225, 154, 198, 214, 60, 139, 234, 125, 53, 24, 237, 242, 244, 140, 175, 189, 188, 217, 227, 191, 150, 127, 218, 172, 148, 46, 0, 54, 145, 0, 0, 51, 110, 73, 68, 65, 84, 120, 218, 164, 89, 251, 83, 27, 85, 20, 206, 46, 44, 201, 238, 38, 89, 8, 132, 141, 36, 146, 9, 133, 52, 36, 72, 0, 27, 11, 104, 90, 94, 166, 140, 6, 104, 10, 82, 26, 106, 11, 229, 209, 2, 125, 88, 30, 133, 14, 211, 240, 40, 200, 163, 128, 14, 227, 116, 176, 160, 72, 145, 74, 45, 48, 90, 113, 208, 81, 199, 225, 23, 255, 46, 191, 187, 187, 233, 110, 124, 160, 213, 143, 189, 143, 239, 158, 123, 218, 41, 223, 185, 231, 62, 170, 243, 165, 228, 229, 81, 255, 21, 167, 242, 104, 49, 192, 222, 200, 189, 113, 195, 27, 53, 68, 121, 171, 143, 97, 246, 99, 99, 226, 24, 195, 39, 235, 157, 252, 254, 244, 244, 62, 195, 241, 62, 234, 20, 179, 63, 54, 54, 166, 181, 171, 94, 86, 43, 173, 167, 173, 86, 243, 149, 83, 85, 81, 96, 63, 202, 36, 96, 76, 237, 70, 85, 251, 55, 249, 87, 136, 159, 222, 106, 77, 206, 126, 73, 188, 171, 224, 174, 140, 236, 251, 5, 5, 5, 254, 50, 2, 211, 253, 234, 172, 99, 159, 2, 31, 43, 64, 55, 1, 202, 208, 177, 172, 106, 141, 143, 115, 194, 49, 60, 62, 49, 190, 120, 102, 113, 124, 124, 98, 113, 157, 61, 179, 56, 49, 49, 62, 129, 114, 102, 241, 204, 153, 51, 109, 158, 197, 197, 241, 137, 245, 8, 161, 139, 139, 145, 145, 54, 46, 178, 56, 62, 57, 82, 56, 17, 89, 95, 92, 95, 159, 24, 55, 114, 95, 74, 248, 232, 163, 143, 200, 223, 248, 145, 68, 18, 1, 19, 16, 111, 49, 231, 51, 9, 241, 169, 85, 255, 19, 186, 165, 105, 223, 41, 21, 223, 160, 188, 12, 168, 60, 166, 162, 116, 110, 178, 210, 102, 140, 25, 56, 222, 202, 240, 220, 216, 24, 148, 226, 248, 116, 74, 220, 143, 137, 49, 72, 102, 153, 142, 49, 188, 24, 139, 49, 90, 187, 234, 53, 182, 191, 63, 22, 165, 104, 243, 253, 59, 59, 143, 30, 237, 236, 176, 101, 143, 100, 160, 143, 129, 59, 211, 126, 149, 171, 246, 29, 251, 253, 177, 125, 56, 82, 116, 114, 195, 79, 223, 127, 255, 61, 10, 169, 127, 66, 139, 250, 104, 46, 23, 16, 224, 215, 95, 179, 239, 103, 125, 248, 225, 103, 50, 238, 167, 166, 166, 165, 93, 173, 174, 126, 5, 120, 237, 53, 148, 155, 248, 209, 0, 195, 213, 213, 87, 211, 210, 82, 83, 53, 62, 12, 116, 214, 192, 114, 230, 140, 102, 96, 125, 164, 84, 37, 139, 227, 195, 195, 94, 126, 60, 97, 186, 145, 73, 125, 144, 154, 154, 250, 14, 240, 16, 197, 132, 90, 234, 43, 13, 62, 66, 240, 17, 200, 163, 169, 18, 250, 0, 169, 19, 172, 10, 214, 17, 160, 150, 155, 170, 58, 165, 39, 65, 33, 160, 42, 135, 236, 224, 178, 9, 1, 96, 188, 17, 219, 143, 97, 173, 142, 253, 39, 232, 153, 66, 129, 189, 97, 9, 59, 68, 202, 160, 183, 242, 20, 63, 54, 77, 81, 62, 67, 138, 89, 79, 100, 158, 118, 242, 34, 195, 51, 92, 148, 139, 25, 124, 20, 111, 240, 48, 162, 45, 23, 43, 216, 160, 122, 77, 199, 162, 136, 9, 189, 21, 250, 19, 4, 126, 27, 210, 2, 130, 43, 189, 71, 248, 100, 59, 33, 59, 119, 224, 183, 31, 139, 166, 36, 119, 127, 241, 197, 7, 31, 124, 251, 43, 193, 183, 10, 254, 13, 143, 147, 236, 59, 159, 29, 251, 236, 216, 177, 215, 142, 65, 235, 59, 208, 63, 173, 250, 149, 106, 72, 159, 168, 187, 82, 163, 192, 138, 57, 169, 26, 31, 102, 92, 18, 23, 24, 39, 203, 220, 178, 168, 13, 8, 71, 164, 141, 52, 234, 80, 49, 191, 190, 24, 1, 85, 3, 64, 209, 81, 137, 2, 192, 100, 66, 165, 69, 60, 58, 228, 16, 48, 201, 242, 163, 178, 219, 31, 60, 232, 123, 80, 181, 181, 181, 28, 71, 16, 186, 170, 84, 5, 162, 34, 56, 15, 108, 1, 138, 125, 23, 120, 254, 252, 121, 176, 78, 119, 170, 210, 123, 195, 230, 245, 86, 178, 211, 255, 9, 86, 206, 88, 105, 97, 5, 79, 79, 152, 162, 173, 233, 148, 79, 180, 148, 10, 30, 31, 79, 165, 135, 44, 172, 200, 26, 133, 64, 192, 226, 102, 5, 193, 194, 82, 110, 142, 167, 67, 30, 161, 178, 130, 181, 80, 180, 234, 229, 141, 145, 221, 193, 64, 219, 119, 134, 154, 134, 134, 154, 111, 252, 134, 166, 169, 185, 163, 185, 185, 3, 165, 25, 188, 137, 160, 153, 84, 170, 125, 232, 145, 29, 126, 216, 77, 242, 146, 107, 71, 7, 186, 71, 15, 14, 126, 216, 28, 29, 237, 30, 64, 51, 128, 246, 223, 240, 129, 129, 205, 1, 240, 218, 251, 166, 84, 19, 126, 169, 166, 212, 62, 251, 253, 212, 178, 130, 52, 41, 7, 92, 149, 161, 116, 48, 132, 31, 212, 0, 154, 130, 178, 84, 141, 15, 179, 62, 92, 81, 225, 112, 204, 21, 70, 34, 142, 26, 135, 145, 26, 25, 25, 30, 71, 52, 160, 76, 172, 23, 214, 204, 89, 10, 209, 147, 182, 132, 241, 241, 245, 138, 92, 129, 202, 109, 179, 140, 140, 71, 38, 71, 34, 19, 35, 133, 21, 21, 172, 83, 13, 0, 89, 120, 40, 44, 11, 158, 136, 63, 5, 0, 254, 110, 187, 29, 1, 80, 215, 186, 187, 180, 180, 92, 23, 92, 174, 219, 13, 126, 147, 95, 23, 156, 127, 94, 39, 253, 60, 149, 107, 252, 236, 6, 171, 242, 235, 170, 150, 131, 187, 75, 152, 185, 91, 181, 251, 20, 157, 186, 165, 224, 211, 93, 120, 180, 230, 235, 204, 255, 19, 124, 160, 82, 16, 66, 34, 19, 165, 172, 52, 205, 244, 88, 194, 149, 158, 74, 142, 231, 24, 159, 197, 139, 0, 176, 180, 149, 178, 150, 128, 69, 240, 4, 68, 55, 207, 113, 54, 145, 216, 45, 41, 86, 213, 75, 202, 63, 30, 206, 192, 236, 79, 87, 6, 216, 202, 82, 155, 173, 20, 165, 179, 194, 104, 171, 100, 3, 149, 168, 88, 12, 130, 119, 150, 86, 170, 118, 118, 127, 236, 70, 44, 102, 97, 99, 92, 242, 251, 181, 181, 181, 151, 128, 203, 151, 47, 215, 202, 229, 40, 14, 114, 249, 125, 137, 159, 59, 7, 2, 103, 59, 196, 244, 151, 153, 164, 95, 104, 159, 223, 143, 16, 144, 164, 127, 69, 2, 89, 243, 36, 10, 80, 161, 6, 136, 252, 126, 127, 159, 198, 135, 105, 99, 56, 39, 199, 57, 25, 39, 231, 227, 125, 12, 21, 98, 24, 143, 32, 33, 44, 8, 30, 175, 219, 35, 132, 67, 97, 1, 85, 72, 228, 56, 134, 50, 184, 25, 198, 45, 138, 162, 39, 132, 127, 190, 155, 9, 113, 174, 39, 171, 171, 95, 185, 238, 61, 112, 173, 174, 60, 121, 242, 213, 179, 212, 70, 240, 39, 141, 207, 100, 222, 120, 175, 207, 37, 243, 190, 198, 191, 178, 131, 111, 237, 62, 213, 253, 103, 236, 238, 86, 253, 239, 0, 160, 60, 30, 81, 8, 185, 163, 110, 138, 114, 154, 249, 168, 133, 245, 134, 4, 78, 44, 181, 58, 3, 149, 33, 193, 98, 9, 4, 88, 139, 197, 34, 160, 43, 138, 94, 67, 56, 4, 187, 232, 164, 41, 213, 43, 151, 21, 88, 214, 115, 155, 50, 13, 73, 203, 188, 227, 252, 249, 140, 140, 243, 77, 143, 7, 55, 58, 154, 208, 5, 233, 192, 88, 198, 208, 200, 76, 198, 121, 213, 62, 180, 19, 173, 96, 241, 103, 135, 111, 39, 95, 254, 46, 251, 174, 244, 243, 29, 78, 120, 119, 191, 187, 219, 112, 33, 145, 223, 37, 60, 251, 207, 252, 46, 56, 234, 75, 118, 127, 170, 223, 95, 224, 247, 155, 250, 72, 74, 45, 240, 23, 164, 41, 144, 115, 0, 169, 226, 49, 32, 13, 99, 198, 131, 7, 26, 31, 78, 100, 56, 25, 140, 211, 233, 228, 41, 193, 45, 69, 130, 52, 224, 230, 121, 61, 35, 50, 0, 226, 3, 96, 24, 189, 30, 29, 229, 67, 32, 8, 97, 206, 229, 106, 116, 61, 123, 120, 207, 238, 250, 234, 137, 171, 145, 8, 236, 250, 170, 209, 117, 15, 1, 1, 238, 114, 61, 235, 131, 93, 226, 8, 4, 151, 11, 118, 112, 213, 14, 190, 21, 204, 223, 218, 154, 143, 163, 245, 95, 1, 19, 229, 13, 161, 234, 48, 95, 151, 158, 158, 140, 207, 140, 26, 45, 122, 201, 47, 201, 245, 28, 206, 249, 78, 220, 8, 104, 3, 109, 165, 66, 130, 80, 105, 177, 244, 120, 156, 41, 140, 133, 13, 67, 122, 236, 45, 97, 132, 0, 139, 24, 16, 57, 42, 100, 129, 93, 228, 245, 148, 234, 229, 240, 178, 149, 129, 24, 9, 128, 14, 8, 159, 145, 145, 209, 149, 209, 132, 212, 63, 184, 129, 222, 108, 23, 88, 7, 198, 114, 154, 34, 51, 25, 57, 170, 125, 232, 81, 212, 81, 26, 240, 86, 10, 183, 79, 94, 110, 105, 32, 168, 175, 255, 246, 110, 67, 125, 67, 119, 67, 195, 39, 32, 13, 223, 105, 185, 98, 207, 38, 205, 128, 196, 97, 191, 208, 114, 151, 180, 231, 236, 5, 84, 138, 191, 192, 68, 209, 169, 200, 0, 56, 215, 23, 16, 145, 53, 72, 83, 1, 130, 10, 115, 250, 52, 62, 124, 136, 40, 204, 57, 157, 156, 19, 133, 55, 184, 57, 31, 50, 2, 35, 131, 231, 105, 158, 72, 205, 160, 0, 28, 167, 215, 187, 221, 140, 27, 93, 20, 192, 34, 248, 86, 92, 237, 107, 171, 247, 86, 83, 87, 93, 73, 171, 43, 207, 86, 30, 174, 185, 50, 215, 86, 238, 173, 152, 8, 199, 248, 59, 43, 42, 95, 89, 147, 236, 237, 50, 87, 230, 43, 219, 251, 212, 75, 99, 30, 97, 147, 191, 133, 0, 72, 134, 166, 178, 170, 68, 222, 147, 201, 47, 201, 205, 86, 61, 77, 211, 86, 228, 2, 43, 206, 0, 76, 72, 8, 139, 97, 103, 20, 155, 128, 192, 122, 16, 0, 1, 228, 127, 168, 47, 90, 60, 76, 15, 167, 247, 32, 243, 137, 148, 234, 194, 83, 158, 65, 91, 101, 169, 87, 72, 161, 252, 77, 57, 93, 57, 93, 179, 179, 179, 51, 57, 232, 204, 124, 190, 129, 222, 204, 236, 108, 87, 87, 23, 104, 70, 78, 100, 163, 107, 86, 181, 119, 116, 140, 41, 126, 39, 107, 175, 95, 192, 166, 94, 63, 58, 122, 26, 5, 130, 143, 94, 123, 79, 226, 13, 42, 255, 43, 123, 253, 133, 107, 3, 164, 189, 100, 47, 208, 211, 125, 101, 38, 218, 218, 247, 192, 126, 135, 236, 170, 242, 126, 140, 61, 62, 14, 144, 120, 13, 96, 154, 253, 142, 198, 135, 103, 201, 138, 231, 121, 31, 89, 244, 62, 159, 158, 113, 115, 42, 120, 138, 230, 69, 228, 128, 56, 245, 233, 245, 88, 252, 10, 67, 196, 184, 25, 254, 98, 111, 251, 197, 149, 139, 43, 166, 181, 146, 162, 181, 139, 171, 79, 238, 149, 180, 103, 150, 172, 148, 172, 188, 179, 86, 210, 187, 182, 246, 228, 171, 135, 176, 203, 188, 168, 253, 98, 9, 177, 131, 175, 41, 118, 50, 31, 251, 252, 212, 252, 97, 240, 74, 21, 17, 116, 11, 135, 187, 229, 249, 214, 175, 255, 6, 173, 243, 203, 207, 151, 150, 182, 90, 167, 14, 15, 171, 174, 92, 153, 47, 223, 202, 191, 162, 102, 0, 179, 57, 57, 221, 156, 158, 14, 129, 101, 158, 174, 242, 184, 221, 44, 241, 116, 140, 19, 110, 77, 54, 91, 193, 227, 55, 122, 124, 60, 109, 16, 69, 228, 250, 16, 199, 196, 104, 103, 216, 226, 193, 102, 136, 213, 143, 172, 32, 138, 177, 48, 227, 102, 169, 104, 24, 89, 143, 210, 184, 248, 144, 7, 110, 139, 162, 143, 227, 153, 105, 239, 224, 224, 173, 193, 137, 207, 111, 117, 118, 222, 186, 117, 203, 214, 217, 217, 137, 205, 190, 50, 226, 0, 233, 36, 95, 241, 45, 213, 238, 157, 102, 20, 191, 228, 218, 238, 235, 45, 45, 163, 192, 137, 235, 39, 186, 71, 71, 7, 208, 130, 163, 115, 226, 68, 156, 95, 79, 228, 45, 45, 215, 200, 252, 22, 240, 129, 131, 90, 123, 25, 46, 84, 5, 38, 200, 218, 103, 127, 37, 75, 125, 4, 192, 141, 59, 17, 234, 235, 64, 214, 43, 26, 31, 30, 65, 45, 98, 201, 139, 162, 91, 228, 13, 28, 13, 149, 17, 253, 60, 15, 125, 221, 34, 99, 160, 121, 100, 123, 66, 25, 172, 124, 222, 96, 48, 200, 210, 147, 13, 131, 241, 132, 60, 78, 190, 8, 88, 91, 93, 121, 88, 212, 158, 89, 212, 139, 181, 93, 212, 219, 94, 116, 113, 69, 230, 69, 43, 171, 15, 139, 138, 122, 101, 158, 169, 216, 21, 174, 204, 71, 0, 4, 183, 231, 131, 175, 255, 178, 173, 83, 240, 234, 223, 162, 92, 170, 229, 89, 219, 135, 83, 135, 243, 82, 0, 144, 68, 110, 77, 182, 146, 245, 107, 166, 205, 68, 96, 104, 175, 112, 43, 120, 162, 61, 206, 105, 179, 149, 78, 183, 18, 206, 84, 120, 219, 110, 224, 70, 47, 93, 234, 41, 39, 34, 0, 91, 128, 208, 147, 231, 20, 88, 40, 111, 9, 176, 216, 9, 48, 132, 220, 47, 244, 232, 61, 56, 254, 68, 79, 105, 93, 242, 10, 39, 135, 39, 11, 5, 31, 191, 243, 168, 233, 124, 70, 115, 211, 240, 70, 78, 14, 22, 124, 14, 54, 122, 130, 156, 73, 194, 113, 22, 0, 237, 80, 237, 56, 4, 40, 126, 201, 181, 242, 105, 254, 135, 3, 180, 155, 104, 54, 7, 128, 31, 84, 142, 90, 229, 90, 251, 230, 193, 230, 230, 15, 151, 223, 183, 23, 152, 252, 105, 5, 105, 38, 156, 233, 237, 126, 211, 85, 156, 251, 94, 147, 145, 117, 243, 38, 105, 110, 222, 204, 202, 66, 151, 212, 132, 194, 126, 213, 228, 215, 248, 80, 16, 156, 50, 240, 50, 240, 172, 133, 46, 118, 54, 3, 249, 0, 90, 178, 36, 218, 125, 232, 40, 118, 158, 242, 101, 102, 38, 65, 209, 181, 123, 189, 73, 73, 144, 248, 226, 51, 240, 246, 162, 146, 139, 132, 183, 99, 173, 223, 83, 248, 179, 246, 63, 216, 227, 188, 238, 233, 55, 186, 169, 43, 191, 232, 18, 244, 127, 253, 236, 235, 114, 171, 69, 249, 130, 60, 168, 76, 44, 15, 110, 205, 95, 145, 3, 0, 43, 153, 54, 211, 16, 84, 111, 85, 4, 142, 115, 243, 63, 115, 114, 163, 183, 225, 70, 47, 93, 234, 245, 56, 4, 135, 221, 183, 123, 98, 28, 229, 20, 4, 162, 60, 138, 24, 102, 5, 247, 109, 150, 139, 113, 86, 198, 25, 198, 35, 64, 130, 203, 164, 195, 88, 225, 176, 113, 252, 206, 80, 51, 78, 129, 29, 35, 63, 99, 207, 7, 114, 200, 70, 128, 244, 255, 249, 12, 186, 224, 228, 56, 168, 218, 113, 8, 80, 252, 210, 229, 91, 0, 78, 244, 42, 112, 218, 63, 138, 191, 255, 62, 225, 240, 33, 221, 90, 251, 213, 178, 52, 192, 95, 80, 102, 178, 151, 73, 234, 103, 73, 56, 70, 144, 165, 66, 137, 2, 41, 6, 202, 52, 62, 124, 219, 196, 25, 64, 170, 164, 183, 0, 0, 111, 1, 120, 19, 136, 68, 34, 197, 33, 202, 56, 124, 148, 93, 100, 136, 160, 237, 69, 16, 56, 51, 41, 41, 169, 168, 228, 89, 102, 210, 113, 8, 91, 2, 30, 23, 188, 63, 179, 189, 168, 8, 227, 224, 170, 29, 173, 60, 191, 174, 46, 127, 111, 187, 234, 71, 141, 254, 144, 125, 123, 65, 17, 29, 93, 93, 249, 139, 0, 64, 165, 6, 192, 118, 213, 252, 148, 28, 0, 102, 4, 128, 222, 172, 215, 167, 91, 13, 86, 51, 4, 198, 226, 78, 224, 176, 255, 13, 79, 1, 143, 223, 232, 241, 133, 17, 210, 140, 59, 196, 136, 62, 168, 236, 36, 73, 49, 20, 134, 252, 140, 7, 3, 34, 75, 225, 12, 236, 193, 24, 147, 146, 224, 50, 108, 180, 25, 139, 75, 123, 248, 161, 230, 28, 236, 248, 51, 145, 159, 103, 103, 102, 103, 54, 54, 102, 102, 54, 102, 80, 205, 68, 54, 200, 182, 47, 67, 181, 255, 246, 219, 144, 226, 151, 124, 176, 121, 238, 0, 153, 29, 184, 180, 137, 106, 115, 115, 243, 224, 135, 209, 134, 1, 101, 0, 252, 104, 251, 57, 123, 181, 116, 199, 55, 225, 122, 103, 55, 225, 29, 72, 9, 128, 155, 47, 64, 98, 65, 142, 8, 168, 15, 84, 99, 178, 198, 135, 119, 68, 134, 231, 230, 10, 71, 114, 35, 197, 14, 199, 100, 161, 99, 206, 49, 55, 50, 60, 236, 40, 44, 116, 44, 142, 84, 56, 194, 6, 227, 226, 81, 118, 193, 167, 8, 235, 130, 176, 73, 88, 219, 174, 36, 210, 246, 190, 224, 104, 251, 147, 50, 123, 209, 246, 199, 185, 212, 130, 191, 33, 181, 193, 170, 170, 61, 221, 188, 70, 127, 162, 242, 246, 94, 121, 28, 191, 252, 24, 239, 189, 186, 176, 151, 16, 1, 243, 135, 219, 87, 242, 241, 14, 0, 193, 137, 160, 6, 125, 58, 157, 39, 11, 12, 158, 114, 4, 87, 231, 27, 192, 227, 55, 122, 233, 82, 175, 79, 17, 57, 134, 245, 240, 12, 142, 249, 56, 248, 246, 68, 125, 12, 242, 29, 118, 5, 214, 227, 140, 50, 20, 50, 64, 8, 123, 100, 130, 203, 122, 167, 173, 120, 208, 38, 224, 5, 49, 80, 234, 197, 45, 255, 22, 128, 109, 126, 189, 176, 211, 6, 122, 203, 139, 55, 42, 140, 13, 174, 99, 239, 87, 237, 129, 216, 180, 226, 151, 254, 230, 192, 193, 102, 119, 125, 253, 232, 64, 195, 165, 205, 6, 73, 232, 131, 131, 238, 122, 72, 46, 13, 64, 233, 163, 237, 231, 236, 210, 219, 14, 17, 179, 0, 1, 128, 45, 0, 73, 63, 158, 1, 226, 145, 160, 52, 202, 195, 224, 213, 52, 147, 198, 135, 155, 200, 205, 205, 157, 112, 20, 230, 26, 141, 72, 74, 70, 71, 133, 99, 114, 125, 110, 206, 49, 57, 89, 17, 41, 54, 230, 218, 124, 198, 163, 237, 92, 102, 127, 63, 66, 160, 68, 18, 180, 95, 106, 143, 203, 252, 56, 105, 33, 52, 177, 183, 19, 254, 6, 214, 190, 198, 254, 134, 60, 31, 15, 123, 186, 189, 229, 132, 0, 40, 255, 229, 117, 18, 0, 103, 37, 28, 78, 157, 85, 80, 254, 234, 217, 5, 100, 131, 23, 1, 176, 60, 181, 157, 159, 191, 172, 35, 91, 185, 213, 156, 98, 165, 176, 168, 243, 104, 235, 201, 100, 112, 218, 108, 80, 185, 106, 87, 185, 108, 167, 9, 143, 223, 232, 165, 75, 189, 149, 134, 222, 140, 207, 71, 78, 191, 228, 234, 195, 249, 242, 24, 188, 215, 136, 12, 206, 108, 56, 18, 91, 41, 222, 201, 167, 232, 19, 92, 214, 177, 146, 141, 1, 129, 55, 13, 53, 159, 199, 219, 159, 148, 253, 207, 55, 63, 254, 124, 227, 124, 51, 217, 237, 113, 3, 232, 192, 80, 211, 250, 70, 78, 134, 106, 31, 218, 25, 83, 252, 146, 55, 187, 27, 54, 175, 159, 56, 241, 222, 123, 23, 26, 186, 47, 92, 184, 208, 221, 221, 61, 122, 185, 69, 226, 239, 97, 0, 252, 104, 251, 166, 253, 21, 136, 89, 150, 102, 74, 131, 152, 126, 146, 2, 228, 55, 95, 41, 16, 0, 116, 37, 34, 63, 9, 161, 75, 18, 128, 95, 227, 195, 219, 32, 112, 33, 86, 180, 17, 176, 121, 109, 165, 197, 19, 197, 54, 163, 163, 112, 124, 206, 104, 204, 117, 176, 60, 91, 124, 164, 221, 167, 4, 64, 99, 251, 113, 109, 0, 128, 191, 241, 103, 126, 252, 47, 248, 82, 93, 157, 110, 175, 85, 13, 0, 232, 191, 189, 189, 32, 75, 190, 176, 160, 107, 221, 221, 154, 90, 88, 144, 249, 143, 103, 23, 206, 106, 82, 192, 214, 182, 46, 63, 127, 73, 151, 46, 9, 14, 129, 83, 172, 6, 73, 224, 191, 227, 250, 63, 241, 20, 194, 227, 55, 122, 233, 82, 143, 164, 160, 55, 32, 52, 204, 228, 85, 152, 210, 227, 168, 232, 99, 56, 3, 71, 46, 254, 100, 144, 51, 248, 120, 218, 108, 78, 112, 137, 224, 215, 209, 201, 122, 168, 212, 33, 236, 242, 205, 242, 61, 191, 163, 185, 99, 16, 135, 189, 89, 114, 241, 195, 230, 15, 144, 119, 128, 46, 213, 142, 119, 0, 197, 239, 100, 195, 233, 19, 245, 95, 156, 126, 235, 90, 75, 203, 245, 107, 31, 124, 114, 250, 244, 181, 211, 215, 47, 17, 126, 173, 229, 52, 6, 62, 0, 63, 210, 222, 253, 34, 0, 10, 144, 1, 208, 81, 30, 1, 101, 189, 81, 200, 27, 144, 2, 105, 244, 42, 153, 172, 241, 225, 45, 53, 70, 99, 77, 155, 49, 183, 6, 1, 137, 82, 83, 83, 72, 66, 179, 56, 183, 56, 23, 6, 11, 229, 245, 30, 101, 23, 248, 204, 227, 100, 133, 95, 108, 236, 85, 2, 0, 43, 95, 230, 68, 224, 164, 56, 47, 105, 44, 58, 14, 46, 165, 254, 23, 246, 76, 98, 175, 11, 34, 0, 14, 53, 9, 160, 252, 245, 41, 221, 222, 130, 132, 195, 165, 186, 231, 243, 243, 203, 75, 173, 58, 137, 150, 255, 120, 118, 79, 147, 2, 90, 73, 0, 212, 225, 12, 64, 67, 80, 61, 125, 202, 64, 231, 17, 65, 95, 150, 199, 47, 245, 248, 112, 77, 212, 83, 228, 137, 131, 227, 83, 12, 6, 250, 20, 165, 103, 98, 177, 168, 65, 207, 112, 24, 226, 156, 24, 211, 51, 61, 62, 67, 130, 71, 172, 112, 176, 179, 216, 24, 72, 161, 30, 237, 12, 61, 126, 252, 184, 9, 39, 127, 212, 143, 187, 6, 103, 155, 177, 242, 59, 240, 248, 135, 15, 134, 225, 217, 199, 205, 170, 253, 81, 83, 84, 241, 59, 121, 238, 221, 55, 235, 223, 235, 238, 190, 208, 221, 112, 105, 160, 1, 200, 126, 247, 220, 165, 19, 18, 207, 190, 156, 77, 248, 209, 246, 115, 74, 0, 244, 21, 64, 76, 123, 31, 185, 239, 43, 240, 227, 35, 144, 250, 104, 20, 96, 70, 159, 214, 135, 43, 245, 122, 219, 218, 108, 228, 137, 186, 173, 56, 183, 173, 166, 198, 152, 107, 68, 85, 92, 152, 91, 83, 147, 107, 20, 168, 26, 227, 81, 246, 48, 15, 129, 251, 33, 168, 11, 130, 246, 247, 35, 229, 107, 56, 2, 64, 226, 73, 132, 23, 29, 127, 59, 169, 191, 55, 193, 158, 73, 236, 85, 193, 160, 78, 247, 181, 54, 1, 252, 56, 117, 118, 79, 146, 159, 252, 127, 223, 242, 210, 214, 242, 210, 210, 210, 124, 57, 134, 94, 253, 229, 207, 1, 240, 148, 4, 192, 239, 148, 91, 255, 79, 27, 101, 28, 230, 128, 187, 246, 174, 135, 133, 170, 244, 148, 163, 53, 218, 150, 210, 98, 241, 228, 59, 165, 114, 96, 169, 89, 104, 17, 106, 80, 70, 19, 86, 167, 2, 155, 108, 50, 82, 45, 243, 7, 148, 57, 77, 96, 166, 6, 28, 33, 154, 232, 22, 13, 9, 26, 75, 8, 89, 76, 150, 184, 68, 51, 141, 250, 95, 249, 188, 239, 189, 208, 183, 106, 78, 251, 82, 202, 61, 247, 185, 207, 15, 203, 243, 220, 251, 126, 190, 205, 129, 222, 157, 83, 144, 234, 5, 103, 61, 206, 248, 26, 49, 73, 234, 175, 200, 154, 106, 117, 247, 197, 143, 3, 87, 60, 234, 160, 44, 155, 249, 88, 108, 80, 20, 34, 38, 58, 30, 201, 100, 114, 236, 138, 199, 149, 188, 99, 170, 249, 128, 71, 172, 246, 184, 116, 205, 125, 41, 170, 92, 18, 62, 144, 223, 252, 26, 235, 78, 76, 47, 20, 116, 19, 95, 102, 193, 192, 105, 127, 227, 235, 2, 150, 142, 5, 227, 153, 93, 147, 231, 153, 223, 0, 66, 252, 56, 214, 100, 60, 187, 144, 197, 53, 141, 248, 1, 129, 179, 36, 206, 7, 182, 183, 191, 7, 50, 151, 186, 187, 218, 186, 90, 186, 222, 123, 18, 233, 255, 251, 220, 250, 140, 45, 122, 113, 106, 65, 129, 224, 73, 206, 7, 2, 232, 195, 230, 78, 90, 20, 70, 76, 199, 85, 32, 144, 143, 200, 201, 64, 113, 139, 188, 227, 41, 151, 166, 218, 217, 195, 174, 6, 16, 223, 64, 130, 64, 38, 128, 6, 16, 14, 252, 74, 67, 7, 9, 242, 24, 134, 253, 245, 53, 224, 231, 56, 59, 4, 210, 10, 12, 150, 193, 101, 181, 0, 122, 166, 143, 142, 142, 118, 209, 26, 44, 101, 14, 118, 183, 15, 74, 100, 29, 30, 148, 111, 182, 255, 155, 0, 6, 240, 70, 59, 31, 113, 72, 167, 132, 214, 138, 69, 35, 178, 101, 140, 169, 102, 108, 76, 211, 54, 156, 226, 188, 17, 209, 12, 83, 71, 9, 120, 115, 204, 163, 8, 90, 90, 147, 175, 120, 81, 7, 210, 60, 138, 126, 37, 47, 107, 145, 13, 177, 218, 67, 91, 95, 159, 89, 127, 211, 35, 160, 25, 72, 234, 254, 51, 56, 235, 73, 6, 248, 24, 78, 252, 55, 158, 126, 236, 27, 228, 253, 0, 88, 184, 81, 177, 35, 8, 96, 126, 3, 241, 222, 222, 149, 151, 125, 215, 227, 96, 149, 142, 123, 156, 199, 129, 255, 242, 228, 245, 108, 54, 158, 157, 36, 120, 209, 222, 190, 208, 246, 68, 87, 203, 139, 77, 67, 205, 221, 52, 6, 120, 246, 9, 26, 255, 177, 245, 213, 39, 95, 125, 242, 204, 39, 103, 31, 43, 44, 124, 226, 217, 166, 110, 206, 71, 201, 231, 101, 53, 96, 170, 88, 122, 204, 12, 168, 129, 216, 58, 244, 175, 229, 147, 6, 174, 229, 84, 52, 25, 176, 179, 19, 1, 76, 172, 53, 92, 166, 132, 174, 89, 132, 247, 87, 4, 128, 55, 158, 18, 13, 220, 218, 64, 132, 82, 193, 107, 244, 57, 178, 3, 28, 35, 158, 227, 5, 80, 119, 243, 193, 244, 46, 153, 18, 0, 237, 153, 220, 201, 1, 21, 64, 230, 24, 92, 79, 67, 0, 195, 103, 2, 216, 174, 59, 59, 2, 130, 140, 80, 18, 213, 215, 138, 69, 195, 35, 107, 248, 12, 122, 220, 121, 175, 224, 23, 55, 81, 250, 217, 28, 67, 17, 72, 115, 167, 20, 65, 246, 186, 195, 170, 223, 31, 37, 229, 208, 188, 199, 99, 70, 198, 174, 40, 213, 30, 51, 235, 133, 130, 49, 230, 14, 254, 240, 238, 27, 175, 62, 246, 6, 242, 124, 196, 125, 8, 254, 240, 75, 170, 192, 223, 160, 14, 96, 97, 40, 160, 98, 71, 16, 192, 252, 26, 227, 35, 43, 189, 8, 238, 38, 23, 124, 83, 83, 100, 95, 31, 177, 112, 22, 251, 251, 212, 212, 226, 34, 176, 173, 61, 14, 1, 224, 109, 238, 198, 46, 15, 1, 160, 193, 79, 233, 255, 132, 172, 175, 176, 200, 95, 64, 246, 77, 51, 3, 60, 211, 205, 249, 40, 121, 85, 150, 85, 35, 169, 235, 224, 87, 27, 84, 205, 128, 174, 27, 114, 36, 57, 24, 145, 251, 140, 136, 219, 159, 87, 237, 236, 41, 215, 90, 255, 4, 61, 227, 59, 58, 250, 59, 24, 225, 32, 158, 17, 12, 188, 198, 112, 3, 47, 0, 96, 60, 79, 241, 120, 9, 2, 40, 85, 4, 128, 232, 255, 230, 205, 195, 196, 241, 49, 222, 255, 82, 34, 179, 157, 40, 67, 0, 184, 232, 60, 30, 127, 48, 141, 40, 176, 34, 128, 93, 34, 128, 113, 8, 64, 2, 161, 78, 74, 168, 244, 56, 206, 248, 26, 177, 162, 33, 169, 199, 71, 118, 167, 69, 20, 2, 162, 145, 141, 52, 184, 79, 123, 82, 125, 110, 12, 0, 132, 231, 195, 200, 0, 253, 225, 13, 76, 3, 68, 60, 152, 21, 72, 223, 19, 170, 61, 54, 117, 181, 96, 70, 220, 206, 47, 223, 37, 111, 248, 91, 51, 63, 189, 6, 182, 73, 221, 223, 234, 3, 96, 7, 56, 91, 103, 118, 164, 9, 95, 50, 191, 198, 201, 145, 145, 101, 20, 123, 125, 241, 217, 89, 196, 248, 83, 190, 81, 134, 39, 103, 87, 200, 13, 96, 91, 123, 150, 8, 160, 133, 30, 249, 45, 109, 221, 75, 100, 12, 136, 203, 1, 177, 174, 158, 150, 4, 172, 2, 1, 73, 11, 150, 186, 57, 31, 37, 34, 203, 131, 42, 50, 58, 44, 109, 48, 105, 152, 6, 94, 237, 193, 136, 38, 203, 30, 57, 137, 127, 187, 166, 217, 217, 189, 46, 164, 119, 52, 173, 3, 225, 253, 132, 112, 14, 83, 130, 25, 134, 0, 58, 58, 128, 121, 59, 197, 23, 47, 94, 172, 171, 251, 174, 135, 23, 64, 251, 118, 230, 24, 2, 32, 188, 39, 114, 219, 185, 109, 114, 65, 238, 140, 151, 14, 167, 57, 1, 32, 118, 184, 216, 121, 17, 71, 128, 228, 8, 62, 18, 116, 16, 66, 29, 32, 180, 86, 76, 147, 122, 124, 172, 238, 190, 226, 239, 195, 85, 159, 59, 148, 74, 141, 133, 253, 94, 69, 67, 237, 95, 150, 177, 15, 160, 237, 49, 56, 239, 198, 191, 250, 158, 171, 218, 227, 142, 170, 22, 244, 65, 183, 243, 158, 28, 33, 85, 126, 83, 215, 205, 2, 14, 252, 153, 152, 78, 32, 249, 210, 222, 49, 129, 191, 54, 223, 169, 216, 53, 249, 99, 230, 215, 232, 187, 189, 58, 58, 58, 50, 218, 187, 56, 55, 135, 205, 126, 118, 97, 231, 54, 66, 255, 145, 85, 130, 151, 113, 99, 225, 174, 189, 253, 66, 219, 139, 32, 179, 185, 9, 239, 51, 201, 2, 80, 8, 160, 60, 179, 93, 128, 189, 244, 156, 32, 80, 6, 0, 247, 156, 143, 130, 169, 133, 136, 150, 207, 35, 196, 87, 53, 36, 118, 102, 192, 72, 98, 192, 134, 200, 98, 44, 34, 135, 132, 100, 196, 206, 46, 10, 148, 208, 142, 231, 176, 229, 227, 130, 198, 2, 192, 13, 28, 38, 246, 53, 224, 53, 8, 224, 121, 138, 217, 243, 120, 16, 184, 19, 181, 156, 186, 241, 170, 29, 160, 174, 19, 107, 60, 81, 42, 239, 150, 15, 114, 88, 100, 8, 40, 65, 53, 241, 41, 39, 128, 251, 112, 66, 37, 16, 2, 112, 56, 130, 143, 7, 29, 65, 16, 90, 255, 120, 253, 64, 173, 152, 38, 245, 248, 88, 221, 253, 168, 216, 151, 10, 167, 250, 194, 161, 116, 200, 19, 246, 250, 157, 41, 24, 32, 246, 107, 27, 233, 49, 151, 28, 118, 147, 182, 176, 80, 237, 49, 163, 26, 58, 217, 1, 186, 81, 7, 192, 20, 16, 246, 122, 164, 126, 164, 14, 240, 24, 174, 201, 194, 206, 207, 234, 0, 21, 251, 187, 63, 220, 179, 252, 32, 128, 85, 74, 240, 236, 212, 202, 202, 236, 108, 239, 236, 133, 157, 51, 220, 75, 110, 92, 184, 251, 31, 246, 54, 228, 114, 75, 205, 205, 148, 204, 230, 230, 39, 90, 172, 172, 31, 139, 85, 128, 168, 28, 172, 111, 90, 13, 192, 19, 205, 205, 156, 143, 226, 65, 20, 40, 39, 147, 154, 22, 145, 35, 129, 88, 204, 68, 206, 167, 17, 134, 7, 17, 11, 123, 68, 193, 144, 237, 237, 173, 19, 253, 173, 148, 224, 254, 142, 9, 26, 236, 157, 225, 53, 75, 0, 12, 67, 16, 148, 112, 238, 121, 138, 49, 219, 247, 119, 1, 124, 250, 203, 241, 113, 2, 212, 147, 117, 82, 42, 179, 33, 1, 164, 131, 200, 4, 254, 38, 128, 78, 34, 0, 39, 8, 21, 28, 65, 39, 8, 173, 7, 193, 53, 98, 154, 212, 227, 99, 117, 247, 21, 66, 109, 202, 29, 14, 207, 167, 210, 94, 175, 75, 192, 32, 72, 74, 69, 79, 8, 91, 128, 203, 157, 38, 205, 193, 176, 80, 237, 113, 199, 48, 76, 21, 125, 221, 90, 231, 1, 152, 95, 227, 203, 123, 119, 111, 223, 94, 221, 25, 237, 221, 255, 124, 127, 103, 111, 100, 225, 243, 189, 29, 224, 189, 209, 149, 253, 223, 62, 255, 124, 103, 15, 216, 214, 126, 225, 86, 11, 206, 243, 38, 144, 217, 60, 116, 203, 154, 9, 68, 178, 143, 197, 74, 0, 86, 93, 200, 250, 38, 192, 154, 9, 228, 124, 92, 164, 217, 21, 198, 129, 70, 6, 95, 240, 86, 147, 191, 214, 239, 88, 31, 8, 118, 169, 246, 118, 1, 233, 220, 235, 56, 228, 95, 105, 157, 176, 4, 240, 47, 184, 149, 225, 254, 42, 59, 195, 24, 2, 28, 30, 230, 142, 128, 30, 8, 96, 248, 4, 172, 111, 39, 208, 248, 61, 192, 65, 80, 134, 12, 18, 248, 194, 226, 5, 112, 8, 1, 116, 82, 1, 4, 81, 226, 125, 193, 225, 4, 161, 14, 16, 90, 43, 166, 89, 189, 163, 30, 139, 140, 4, 68, 189, 96, 219, 35, 162, 229, 67, 218, 0, 46, 12, 0, 132, 61, 233, 116, 122, 67, 76, 137, 65, 244, 133, 241, 147, 86, 170, 61, 76, 67, 213, 213, 136, 71, 170, 117, 30, 128, 249, 53, 198, 241, 26, 159, 159, 122, 121, 225, 194, 36, 82, 189, 184, 111, 97, 110, 118, 118, 133, 97, 68, 250, 241, 133, 21, 123, 59, 221, 1, 90, 192, 100, 87, 243, 16, 29, 178, 163, 117, 128, 74, 254, 223, 132, 31, 235, 155, 221, 34, 3, 1, 188, 143, 32, 186, 49, 241, 64, 150, 7, 139, 94, 184, 233, 31, 210, 23, 71, 228, 163, 216, 219, 133, 203, 120, 163, 17, 253, 145, 55, 27, 127, 144, 223, 243, 248, 185, 106, 188, 134, 146, 48, 135, 251, 241, 60, 4, 144, 57, 58, 202, 124, 218, 206, 11, 96, 151, 16, 94, 78, 148, 118, 115, 101, 75, 0, 136, 7, 115, 219, 128, 15, 42, 2, 184, 127, 136, 81, 50, 171, 29, 236, 4, 161, 130, 36, 5, 165, 160, 163, 94, 26, 168, 21, 179, 230, 62, 25, 7, 144, 208, 11, 194, 209, 238, 13, 161, 27, 24, 37, 100, 187, 232, 0, 128, 12, 17, 96, 250, 201, 149, 118, 211, 21, 18, 156, 156, 143, 32, 72, 161, 208, 181, 107, 215, 164, 90, 231, 1, 152, 95, 227, 133, 5, 148, 130, 72, 98, 63, 153, 197, 66, 178, 143, 121, 63, 223, 36, 193, 132, 97, 130, 236, 237, 183, 90, 134, 40, 153, 221, 205, 93, 183, 174, 98, 26, 0, 249, 190, 221, 34, 147, 249, 95, 92, 229, 124, 156, 81, 50, 239, 19, 10, 133, 188, 33, 186, 232, 37, 22, 253, 163, 144, 166, 175, 141, 93, 112, 56, 159, 155, 152, 56, 45, 248, 80, 1, 0, 183, 242, 152, 183, 175, 33, 24, 252, 251, 243, 227, 7, 165, 225, 225, 76, 207, 253, 202, 17, 208, 158, 40, 157, 36, 78, 64, 248, 9, 217, 6, 200, 62, 144, 43, 131, 252, 93, 136, 224, 32, 247, 175, 2, 112, 224, 101, 150, 176, 167, 59, 37, 16, 90, 43, 102, 19, 254, 116, 28, 192, 245, 130, 55, 21, 86, 196, 176, 40, 138, 130, 55, 37, 94, 18, 104, 251, 15, 196, 123, 194, 110, 209, 137, 214, 112, 218, 157, 34, 35, 195, 188, 79, 145, 252, 119, 138, 80, 84, 170, 117, 30, 128, 249, 53, 46, 248, 226, 83, 139, 189, 200, 237, 125, 113, 36, 122, 83, 190, 73, 95, 252, 91, 11, 95, 207, 250, 8, 182, 183, 199, 111, 117, 49, 50, 155, 134, 110, 13, 53, 33, 4, 188, 202, 181, 131, 89, 78, 200, 190, 105, 22, 128, 48, 176, 105, 136, 243, 193, 140, 135, 224, 114, 249, 21, 58, 20, 134, 206, 63, 126, 48, 2, 2, 12, 222, 21, 1, 86, 135, 195, 214, 126, 185, 159, 16, 254, 252, 41, 161, 175, 216, 224, 53, 34, 128, 203, 56, 25, 78, 49, 73, 31, 51, 39, 199, 200, 2, 234, 56, 1, 220, 204, 36, 202, 224, 158, 45, 118, 241, 61, 84, 176, 157, 43, 149, 218, 153, 0, 224, 144, 169, 8, 128, 100, 116, 82, 48, 40, 73, 14, 231, 64, 173, 152, 77, 248, 211, 222, 190, 224, 10, 123, 67, 50, 233, 6, 186, 132, 144, 232, 143, 10, 100, 0, 64, 241, 98, 26, 68, 220, 80, 36, 247, 252, 198, 124, 26, 246, 123, 2, 239, 83, 140, 233, 102, 44, 114, 73, 170, 117, 30, 128, 249, 53, 102, 123, 95, 158, 26, 165, 181, 126, 223, 249, 185, 229, 57, 66, 237, 245, 213, 145, 145, 145, 185, 185, 185, 201, 243, 203, 203, 192, 182, 246, 169, 182, 46, 244, 246, 201, 14, 143, 237, 188, 11, 13, 95, 139, 122, 16, 78, 170, 0, 214, 21, 126, 153, 28, 240, 75, 30, 233, 226, 124, 132, 36, 198, 192, 139, 31, 190, 132, 86, 191, 22, 9, 123, 157, 169, 245, 34, 254, 219, 207, 157, 153, 98, 113, 125, 253, 70, 44, 86, 196, 178, 177, 135, 157, 151, 241, 198, 247, 91, 77, 158, 14, 66, 40, 193, 29, 167, 248, 178, 133, 215, 56, 194, 33, 8, 8, 128, 217, 241, 92, 230, 164, 68, 130, 192, 67, 46, 8, 60, 57, 41, 151, 183, 31, 244, 28, 254, 241, 227, 207, 116, 253, 248, 115, 251, 195, 115, 231, 126, 111, 207, 229, 118, 15, 206, 118, 128, 195, 158, 204, 240, 176, 37, 0, 201, 65, 42, 251, 82, 80, 144, 160, 215, 129, 90, 49, 107, 238, 211, 222, 190, 43, 26, 158, 247, 134, 201, 128, 180, 232, 199, 169, 175, 208, 29, 64, 196, 64, 152, 172, 136, 116, 7, 240, 166, 96, 247, 59, 120, 159, 27, 5, 67, 143, 245, 133, 164, 90, 231, 1, 152, 95, 227, 228, 114, 239, 212, 14, 77, 245, 102, 151, 193, 235, 50, 146, 187, 184, 133, 71, 113, 3, 216, 222, 190, 216, 134, 168, 110, 105, 168, 137, 205, 3, 44, 209, 58, 0, 227, 219, 90, 213, 91, 192, 85, 218, 14, 230, 124, 4, 217, 64, 108, 31, 211, 241, 83, 220, 52, 20, 101, 230, 134, 169, 227, 198, 86, 12, 121, 191, 138, 101, 154, 118, 246, 148, 179, 181, 18, 245, 119, 240, 81, 63, 240, 68, 53, 238, 176, 48, 13, 10, 25, 38, 89, 192, 201, 248, 33, 25, 8, 225, 4, 240, 75, 121, 247, 1, 24, 7, 231, 191, 62, 69, 214, 175, 211, 191, 159, 35, 235, 225, 247, 185, 7, 149, 32, 112, 26, 177, 3, 19, 128, 228, 168, 23, 112, 162, 35, 178, 7, 161, 181, 98, 214, 220, 167, 57, 61, 98, 0, 28, 0, 243, 34, 102, 1, 200, 176, 31, 219, 1, 208, 8, 130, 40, 4, 178, 3, 136, 196, 46, 10, 188, 79, 209, 40, 80, 1, 212, 58, 15, 192, 252, 26, 227, 203, 179, 127, 39, 56, 187, 179, 202, 9, 192, 222, 190, 72, 231, 1, 134, 134, 104, 111, 159, 9, 128, 74, 128, 105, 128, 95, 168, 5, 49, 1, 112, 62, 46, 36, 247, 235, 49, 29, 33, 105, 192, 48, 84, 57, 90, 4, 175, 1, 253, 198, 250, 86, 192, 140, 21, 205, 88, 44, 96, 216, 217, 35, 2, 31, 213, 243, 81, 63, 143, 237, 178, 132, 76, 57, 113, 120, 120, 204, 9, 0, 169, 126, 174, 231, 232, 156, 181, 168, 2, 134, 31, 50, 244, 48, 87, 87, 17, 0, 191, 3, 212, 215, 35, 22, 11, 58, 41, 161, 181, 98, 214, 220, 167, 57, 189, 243, 5, 17, 9, 158, 31, 140, 147, 227, 205, 47, 90, 3, 0, 136, 146, 188, 225, 168, 224, 220, 240, 207, 207, 83, 187, 139, 247, 153, 49, 140, 194, 86, 228, 154, 84, 235, 60, 0, 243, 107, 156, 4, 193, 86, 174, 127, 126, 142, 240, 11, 194, 23, 24, 193, 184, 1, 108, 111, 159, 162, 157, 61, 196, 243, 132, 204, 238, 102, 212, 248, 185, 113, 128, 39, 173, 185, 64, 130, 172, 44, 144, 124, 183, 116, 53, 119, 115, 62, 81, 85, 207, 231, 245, 226, 140, 142, 247, 89, 203, 143, 41, 49, 195, 76, 198, 102, 2, 250, 86, 236, 195, 117, 20, 127, 243, 73, 221, 206, 30, 240, 219, 100, 1, 253, 255, 7, 163, 29, 124, 19, 51, 129, 188, 0, 208, 248, 61, 199, 214, 239, 132, 255, 105, 6, 160, 0, 46, 13, 108, 175, 8, 64, 66, 123, 231, 5, 73, 64, 125, 223, 41, 53, 214, 138, 89, 115, 159, 230, 244, 24, 29, 165, 211, 0, 104, 245, 10, 103, 179, 0, 78, 88, 8, 66, 214, 40, 56, 168, 253, 81, 222, 167, 88, 48, 212, 66, 223, 37, 169, 214, 121, 0, 230, 247, 145, 111, 111, 100, 150, 230, 246, 183, 71, 239, 238, 239, 239, 237, 220, 221, 27, 93, 32, 120, 21, 185, 254, 237, 253, 125, 96, 91, 123, 175, 37, 0, 164, 246, 221, 40, 4, 225, 162, 133, 155, 7, 160, 223, 75, 184, 176, 190, 233, 93, 250, 48, 231, 163, 12, 170, 121, 82, 225, 75, 170, 168, 240, 229, 35, 174, 205, 0, 42, 84, 51, 91, 201, 27, 91, 96, 220, 32, 245, 63, 59, 123, 94, 249, 71, 22, 192, 227, 215, 255, 27, 103, 50, 56, 2, 58, 239, 87, 9, 96, 216, 218, 0, 142, 238, 99, 11, 32, 27, 192, 209, 193, 233, 142, 192, 247, 2, 144, 5, 92, 60, 219, 1, 164, 23, 28, 65, 66, 168, 163, 177, 86, 204, 186, 251, 180, 185, 31, 85, 68, 54, 12, 224, 16, 156, 252, 44, 64, 148, 220, 20, 211, 40, 18, 17, 187, 131, 247, 209, 99, 42, 136, 84, 164, 90, 231, 1, 152, 223, 64, 214, 151, 93, 156, 59, 143, 92, 127, 113, 225, 58, 162, 124, 52, 252, 227, 217, 57, 150, 251, 95, 39, 216, 222, 30, 103, 59, 192, 219, 93, 67, 116, 30, 128, 27, 8, 224, 230, 1, 184, 113, 0, 172, 183, 121, 31, 37, 111, 208, 254, 190, 25, 72, 230, 141, 124, 196, 191, 137, 126, 159, 161, 175, 231, 147, 120, 241, 49, 9, 148, 12, 152, 118, 118, 67, 225, 162, 124, 58, 232, 81, 43, 78, 36, 42, 71, 64, 59, 21, 192, 167, 71, 108, 207, 239, 76, 252, 114, 238, 207, 167, 126, 29, 126, 56, 126, 156, 251, 23, 1, 176, 74, 160, 69, 168, 227, 47, 82, 174, 173, 181, 137, 32, 10, 179, 105, 179, 217, 75, 109, 116, 75, 204, 66, 215, 198, 135, 90, 98, 42, 45, 193, 174, 198, 214, 224, 214, 75, 65, 98, 212, 70, 124, 240, 66, 141, 82, 170, 130, 138, 55, 52, 130, 136, 86, 95, 42, 34, 168, 136, 111, 234, 139, 160, 62, 84, 17, 41, 24, 16, 188, 65, 189, 252, 41, 191, 179, 59, 201, 206, 54, 237, 212, 169, 75, 141, 158, 156, 158, 7, 57, 103, 103, 230, 204, 247, 157, 15, 111, 180, 138, 159, 88, 187, 172, 205, 233, 3, 244, 107, 31, 71, 207, 19, 230, 233, 85, 173, 206, 120, 206, 2, 241, 205, 57, 239, 98, 66, 40, 121, 222, 241, 60, 175, 215, 234, 215, 109, 242, 231, 245, 120, 24, 5, 92, 63, 123, 201, 52, 84, 89, 62, 0, 139, 219, 10, 72, 159, 250, 124, 252, 169, 156, 240, 31, 80, 0, 96, 7, 120, 63, 61, 98, 63, 227, 3, 12, 118, 35, 153, 215, 232, 30, 96, 89, 62, 0, 221, 3, 112, 49, 113, 199, 241, 241, 253, 113, 218, 237, 171, 122, 124, 58, 111, 23, 192, 247, 177, 193, 5, 204, 15, 76, 15, 228, 167, 175, 139, 252, 158, 193, 117, 1, 62, 232, 35, 107, 99, 196, 23, 185, 228, 47, 2, 81, 0, 243, 65, 1, 196, 222, 238, 253, 78, 43, 192, 135, 216, 99, 86, 0, 60, 24, 228, 223, 4, 126, 160, 2, 160, 53, 157, 18, 74, 247, 251, 237, 178, 118, 220, 195, 173, 54, 211, 7, 208, 244, 164, 219, 235, 129, 240, 52, 122, 43, 145, 236, 197, 14, 55, 128, 69, 112, 252, 250, 64, 213, 234, 173, 213, 236, 209, 91, 253, 5, 139, 252, 187, 117, 35, 140, 114, 157, 90, 173, 54, 153, 84, 101, 249, 0, 44, 142, 248, 0, 67, 77, 188, 191, 4, 188, 159, 217, 19, 165, 146, 255, 197, 148, 200, 207, 248, 0, 96, 250, 119, 251, 96, 16, 248, 0, 56, 4, 178, 222, 111, 81, 62, 192, 5, 252, 198, 170, 85, 92, 76, 198, 9, 240, 125, 192, 125, 96, 125, 88, 153, 129, 14, 61, 192, 255, 123, 209, 165, 56, 14, 72, 0, 34, 127, 135, 113, 242, 230, 158, 182, 61, 109, 199, 155, 9, 149, 181, 139, 15, 139, 11, 11, 224, 1, 91, 1, 230, 135, 247, 98, 5, 88, 191, 217, 183, 130, 111, 120, 48, 232, 61, 182, 142, 141, 69, 20, 64, 34, 161, 166, 98, 236, 141, 78, 181, 203, 218, 241, 14, 96, 187, 129, 62, 128, 166, 38, 211, 73, 111, 58, 155, 213, 251, 226, 74, 102, 119, 213, 205, 213, 70, 45, 140, 67, 219, 192, 61, 157, 248, 45, 0, 130, 86, 150, 252, 24, 152, 9, 163, 6, 28, 59, 63, 57, 185, 79, 149, 229, 3, 176, 184, 173, 1, 31, 96, 106, 228, 68, 15, 224, 125, 194, 251, 153, 61, 17, 242, 1, 68, 254, 138, 95, 0, 24, 243, 68, 50, 193, 7, 96, 132, 16, 142, 14, 16, 229, 3, 192, 139, 147, 224, 32, 23, 99, 17, 190, 159, 3, 210, 87, 208, 93, 61, 87, 48, 243, 85, 93, 207, 3, 241, 155, 25, 96, 248, 191, 208, 175, 163, 0, 110, 182, 221, 68, 66, 219, 46, 19, 10, 136, 4, 75, 218, 197, 55, 101, 190, 0, 134, 185, 45, 128, 181, 1, 63, 30, 132, 135, 64, 174, 0, 94, 111, 188, 119, 47, 224, 3, 36, 84, 36, 20, 61, 29, 238, 118, 144, 80, 89, 219, 176, 155, 250, 0, 177, 78, 220, 248, 230, 60, 64, 28, 25, 179, 211, 114, 11, 73, 219, 209, 171, 118, 110, 95, 78, 7, 100, 106, 166, 51, 164, 15, 64, 254, 164, 18, 11, 163, 176, 72, 92, 237, 200, 101, 85, 89, 62, 0, 139, 219, 218, 224, 3, 84, 90, 240, 254, 33, 198, 7, 16, 250, 25, 31, 128, 144, 29, 172, 0, 212, 3, 240, 55, 129, 212, 251, 81, 247, 199, 62, 169, 41, 160, 62, 96, 21, 23, 163, 19, 190, 111, 131, 225, 131, 55, 59, 169, 187, 74, 85, 183, 114, 227, 104, 249, 166, 241, 193, 240, 127, 129, 127, 52, 67, 9, 221, 179, 150, 222, 232, 224, 112, 39, 107, 151, 223, 20, 35, 5, 240, 231, 246, 237, 217, 159, 7, 155, 25, 95, 143, 103, 243, 60, 179, 126, 225, 8, 208, 44, 128, 207, 167, 190, 126, 13, 10, 32, 22, 38, 52, 129, 4, 75, 218, 156, 62, 64, 66, 209, 251, 120, 125, 128, 172, 158, 183, 170, 118, 171, 62, 128, 158, 72, 133, 81, 51, 147, 29, 184, 245, 207, 170, 178, 124, 0, 22, 71, 124, 128, 139, 28, 222, 95, 225, 240, 127, 250, 2, 182, 208, 207, 248, 0, 116, 206, 227, 249, 0, 148, 255, 86, 62, 0, 10, 128, 241, 1, 194, 24, 61, 192, 247, 171, 142, 99, 143, 2, 239, 81, 242, 86, 210, 173, 226, 233, 245, 176, 253, 17, 254, 47, 244, 219, 153, 181, 91, 252, 132, 130, 10, 118, 179, 141, 246, 118, 89, 187, 88, 198, 105, 30, 63, 236, 12, 248, 187, 254, 173, 254, 109, 238, 211, 247, 70, 23, 56, 183, 179, 200, 42, 128, 182, 3, 20, 0, 183, 5, 220, 187, 199, 40, 97, 154, 138, 220, 169, 90, 130, 208, 157, 118, 89, 155, 211, 7, 136, 167, 250, 178, 255, 160, 15, 144, 61, 202, 233, 3, 52, 11, 64, 150, 15, 192, 226, 120, 62, 192, 16, 94, 234, 9, 14, 255, 39, 32, 16, 182, 208, 31, 160, 129, 171, 41, 155, 196, 7, 216, 181, 166, 161, 16, 180, 24, 31, 128, 92, 248, 141, 213, 171, 185, 152, 49, 123, 55, 210, 89, 160, 126, 175, 160, 143, 101, 180, 188, 110, 141, 234, 132, 255, 87, 109, 175, 134, 131, 191, 215, 33, 242, 187, 6, 37, 114, 15, 18, 185, 178, 191, 89, 1, 20, 155, 5, 48, 254, 174, 94, 175, 127, 157, 251, 50, 27, 100, 124, 125, 57, 118, 234, 39, 118, 129, 95, 243, 243, 191, 176, 27, 132, 132, 16, 60, 101, 182, 5, 16, 151, 63, 69, 239, 51, 75, 176, 164, 205, 235, 3, 196, 50, 201, 127, 208, 7, 24, 51, 34, 250, 0, 147, 118, 193, 117, 211, 210, 124, 0, 22, 119, 247, 134, 223, 211, 63, 218, 62, 244, 244, 201, 179, 151, 143, 206, 140, 80, 143, 207, 240, 127, 34, 0, 140, 60, 19, 251, 129, 6, 130, 219, 75, 171, 249, 96, 192, 7, 160, 134, 127, 73, 62, 0, 125, 71, 124, 0, 46, 198, 212, 221, 28, 246, 119, 59, 231, 230, 64, 118, 48, 18, 186, 78, 76, 151, 36, 150, 123, 11, 158, 42, 216, 192, 17, 191, 22, 245, 91, 166, 252, 155, 31, 181, 33, 248, 4, 165, 143, 176, 11, 124, 241, 238, 29, 10, 96, 110, 238, 203, 15, 186, 5, 244, 234, 40, 8, 84, 0, 123, 34, 5, 240, 16, 135, 64, 182, 5, 36, 192, 244, 6, 180, 151, 160, 132, 202, 218, 188, 62, 64, 103, 38, 189, 188, 62, 64, 214, 92, 160, 15, 80, 152, 68, 34, 165, 249, 0, 44, 174, 139, 227, 3, 140, 148, 64, 250, 15, 241, 254, 160, 217, 59, 39, 246, 79, 208, 10, 0, 181, 151, 109, 144, 125, 235, 190, 230, 203, 3, 8, 249, 0, 190, 64, 192, 53, 46, 70, 209, 173, 130, 149, 14, 30, 160, 31, 169, 180, 197, 224, 222, 116, 54, 75, 63, 174, 205, 249, 113, 12, 234, 75, 242, 126, 67, 249, 199, 189, 254, 208, 66, 123, 45, 179, 125, 78, 224, 43, 254, 16, 248, 251, 221, 28, 21, 128, 159, 113, 239, 247, 159, 111, 191, 215, 135, 5, 208, 66, 10, 61, 229, 23, 0, 46, 118, 124, 112, 159, 18, 42, 107, 71, 244, 1, 250, 247, 45, 175, 15, 144, 140, 234, 3, 116, 93, 50, 20, 165, 75, 149, 230, 3, 176, 184, 46, 226, 3, 148, 24, 222, 207, 16, 255, 10, 108, 116, 124, 61, 48, 48, 13, 32, 246, 19, 31, 128, 146, 73, 178, 127, 18, 124, 128, 48, 70, 27, 195, 28, 124, 156, 233, 3, 96, 90, 82, 33, 185, 160, 52, 73, 128, 208, 252, 127, 31, 254, 191, 161, 95, 73, 65, 47, 64, 49, 56, 191, 162, 201, 159, 254, 163, 54, 72, 61, 195, 195, 199, 162, 109, 224, 23, 20, 0, 75, 249, 44, 195, 3, 2, 84, 40, 40, 0, 86, 1, 247, 217, 77, 32, 45, 233, 248, 81, 240, 129, 127, 180, 203, 218, 192, 246, 29, 167, 161, 15, 96, 30, 93, 94, 31, 32, 109, 242, 33, 27, 156, 235, 211, 215, 199, 167, 115, 210, 124, 0, 22, 215, 21, 240, 1, 14, 55, 240, 126, 232, 255, 140, 4, 189, 126, 233, 6, 190, 128, 45, 246, 87, 192, 7, 128, 238, 15, 65, 252, 18, 124, 0, 46, 70, 91, 160, 15, 144, 208, 120, 125, 0, 56, 20, 206, 15, 22, 173, 166, 68, 252, 154, 124, 255, 31, 181, 139, 229, 226, 236, 236, 91, 86, 0, 236, 42, 152, 43, 128, 7, 84, 0, 209, 21, 160, 177, 4, 28, 99, 55, 129, 116, 168, 195, 15, 186, 56, 192, 59, 137, 118, 89, 59, 62, 147, 212, 11, 141, 97, 127, 205, 92, 94, 31, 160, 79, 139, 132, 140, 15, 92, 157, 169, 117, 200, 243, 1, 88, 92, 215, 200, 84, 105, 106, 251, 14, 244, 246, 135, 43, 83, 212, 215, 29, 153, 42, 53, 122, 253, 202, 254, 195, 176, 197, 254, 82, 247, 233, 0, 219, 39, 104, 119, 19, 154, 252, 86, 62, 0, 30, 158, 15, 128, 171, 130, 77, 92, 140, 38, 154, 255, 47, 100, 148, 177, 170, 192, 63, 166, 253, 239, 10, 0, 70, 208, 227, 225, 97, 20, 0, 191, 4, 252, 164, 29, 96, 145, 231, 117, 253, 15, 87, 0, 111, 1, 6, 81, 1, 116, 166, 40, 161, 154, 66, 24, 47, 18, 42, 107, 27, 182, 43, 167, 15, 208, 167, 69, 66, 94, 228, 39, 49, 230, 125, 86, 154, 15, 16, 196, 161, 0, 24, 225, 227, 92, 72, 248, 184, 2, 194, 7, 108, 255, 11, 216, 66, 127, 143, 175, 15, 0, 233, 55, 188, 205, 62, 28, 124, 97, 161, 54, 4, 221, 10, 242, 75, 0, 193, 193, 92, 140, 34, 156, 255, 31, 19, 235, 3, 120, 255, 191, 5, 20, 31, 71, 102, 3, 215, 137, 10, 224, 115, 189, 254, 237, 78, 164, 0, 78, 81, 1, 80, 91, 183, 97, 3, 37, 148, 222, 104, 89, 219, 116, 221, 112, 216, 95, 85, 143, 46, 171, 15, 96, 24, 8, 105, 209, 7, 80, 87, 172, 15, 48, 116, 99, 191, 127, 185, 3, 130, 15, 38, 128, 49, 247, 125, 224, 138, 111, 251, 95, 32, 225, 98, 127, 15, 227, 3, 12, 250, 112, 48, 154, 64, 26, 4, 65, 158, 151, 224, 3, 224, 3, 141, 224, 32, 23, 163, 8, 231, 255, 119, 199, 59, 132, 126, 67, 249, 223, 45, 160, 204, 10, 32, 50, 25, 18, 110, 1, 145, 231, 121, 29, 21, 112, 187, 89, 0, 179, 179, 97, 1, 104, 44, 161, 177, 118, 89, 27, 204, 111, 57, 125, 0, 83, 67, 72, 139, 62, 128, 186, 82, 125, 128, 30, 194, 251, 67, 194, 7, 141, 123, 76, 132, 132, 15, 216, 98, 255, 1, 134, 236, 53, 244, 1, 78, 55, 228, 1, 168, 10, 22, 227, 3, 208, 2, 48, 200, 197, 152, 121, 79, 48, 255, 159, 203, 136, 245, 1, 210, 166, 52, 8, 20, 181, 81, 0, 101, 42, 128, 232, 18, 176, 68, 1, 140, 35, 255, 88, 2, 22, 91, 1, 212, 32, 161, 24, 247, 150, 181, 241, 170, 135, 195, 254, 157, 137, 168, 62, 64, 87, 103, 127, 188, 63, 209, 111, 210, 29, 1, 193, 197, 208, 7, 128, 196, 55, 66, 90, 244, 1, 212, 149, 234, 3, 236, 111, 224, 253, 23, 183, 55, 240, 254, 74, 131, 251, 143, 47, 96, 11, 253, 71, 120, 129, 8, 94, 31, 128, 191, 12, 96, 159, 173, 250, 0, 84, 0, 85, 79, 164, 15, 144, 17, 235, 3, 196, 77, 121, 24, 56, 106, 67, 3, 150, 77, 7, 115, 140, 0, 236, 1, 97, 243, 31, 105, 2, 238, 32, 255, 45, 5, 160, 242, 9, 149, 181, 163, 250, 0, 154, 73, 162, 137, 88, 0, 144, 103, 208, 63, 250, 97, 65, 62, 50, 56, 246, 166, 240, 123, 153, 179, 40, 136, 72, 8, 155, 243, 87, 87, 172, 15, 208, 83, 154, 66, 175, 79, 120, 63, 208, 29, 28, 243, 9, 239, 63, 18, 244, 254, 7, 200, 22, 251, 71, 186, 79, 179, 100, 134, 250, 0, 225, 211, 194, 7, 96, 98, 129, 156, 64, 132, 49, 227, 216, 75, 207, 255, 91, 134, 88, 31, 32, 110, 182, 16, 61, 36, 237, 226, 99, 240, 1, 238, 243, 5, 64, 21, 176, 232, 33, 224, 151, 127, 15, 196, 183, 129, 141, 21, 32, 1, 130, 199, 223, 214, 206, 199, 167, 173, 42, 138, 227, 150, 12, 74, 95, 169, 212, 86, 4, 3, 182, 153, 210, 6, 91, 173, 147, 64, 134, 115, 45, 63, 92, 212, 48, 166, 2, 50, 55, 134, 48, 149, 76, 42, 152, 208, 68, 11, 196, 200, 6, 101, 26, 135, 46, 97, 252, 170, 3, 29, 137, 10, 235, 252, 181, 200, 18, 93, 226, 22, 53, 97, 49, 250, 79, 249, 57, 247, 93, 250, 218, 78, 48, 168, 95, 224, 221, 251, 189, 231, 222, 232, 114, 206, 189, 239, 254, 122, 231, 200, 151, 62, 244, 211, 131, 123, 229, 106, 81, 207, 175, 190, 19, 224, 112, 76, 22, 123, 110, 20, 247, 58, 58, 90, 91, 111, 216, 138, 42, 29, 108, 128, 79, 78, 78, 114, 29, 208, 62, 249, 182, 163, 124, 50, 212, 116, 214, 106, 241, 191, 220, 7, 24, 232, 247, 1, 181, 188, 27, 232, 234, 87, 128, 194, 125, 79, 203, 129, 191, 162, 64, 203, 133, 247, 227, 39, 86, 85, 160, 250, 64, 242, 49, 83, 153, 204, 233, 146, 15, 178, 15, 160, 207, 255, 183, 111, 3, 228, 248, 12, 84, 229, 106, 31, 128, 53, 192, 118, 155, 106, 163, 38, 210, 180, 243, 247, 255, 229, 254, 221, 253, 3, 20, 59, 191, 87, 208, 129, 1, 190, 84, 164, 0, 59, 4, 16, 208, 210, 54, 185, 19, 200, 158, 206, 93, 67, 192, 221, 3, 64, 174, 254, 193, 165, 66, 3, 168, 136, 149, 160, 208, 189, 114, 78, 246, 57, 212, 58, 125, 164, 86, 109, 4, 16, 16, 162, 247, 236, 121, 7, 55, 62, 122, 223, 238, 245, 151, 22, 213, 133, 111, 20, 223, 96, 94, 96, 220, 240, 151, 58, 216, 12, 58, 91, 121, 190, 216, 106, 241, 127, 220, 7, 224, 44, 96, 161, 7, 7, 192, 93, 120, 3, 222, 124, 97, 98, 115, 116, 110, 110, 173, 1, 46, 254, 130, 55, 209, 243, 208, 194, 232, 194, 201, 137, 23, 124, 152, 193, 166, 240, 230, 185, 197, 197, 145, 77, 228, 88, 193, 201, 205, 77, 95, 242, 9, 137, 254, 240, 152, 56, 4, 73, 178, 15, 32, 103, 0, 214, 252, 159, 5, 64, 14, 244, 105, 208, 131, 46, 215, 83, 217, 54, 213, 70, 100, 172, 113, 231, 239, 255, 203, 253, 187, 251, 7, 240, 59, 93, 196, 8, 0, 237, 60, 223, 127, 127, 137, 199, 82, 251, 251, 187, 97, 41, 235, 82, 222, 173, 16, 141, 170, 207, 195, 11, 135, 128, 95, 238, 30, 2, 10, 244, 79, 35, 181, 17, 228, 197, 215, 19, 203, 186, 88, 32, 32, 10, 181, 31, 220, 43, 47, 174, 149, 69, 61, 255, 34, 181, 170, 247, 251, 27, 249, 224, 161, 170, 88, 34, 67, 132, 217, 23, 241, 219, 194, 206, 27, 108, 24, 201, 75, 224, 217, 98, 191, 39, 88, 28, 50, 242, 91, 232, 115, 253, 192, 191, 189, 15, 224, 91, 94, 144, 85, 157, 239, 232, 230, 200, 80, 115, 115, 195, 230, 214, 252, 218, 130, 240, 254, 205, 134, 6, 150, 125, 19, 139, 139, 163, 205, 204, 254, 125, 79, 79, 52, 112, 7, 180, 103, 115, 126, 126, 109, 237, 112, 243, 144, 175, 107, 243, 196, 208, 225, 209, 209, 137, 164, 75, 182, 117, 159, 16, 79, 176, 201, 167, 120, 201, 179, 10, 176, 124, 68, 160, 127, 8, 80, 41, 127, 72, 31, 20, 15, 17, 217, 54, 213, 206, 200, 149, 157, 191, 255, 63, 226, 40, 62, 93, 91, 230, 168, 250, 104, 7, 121, 77, 240, 31, 12, 224, 235, 127, 52, 128, 182, 99, 242, 117, 240, 213, 252, 33, 64, 89, 64, 225, 11, 160, 64, 255, 87, 219, 244, 78, 96, 5, 218, 177, 7, 194, 1, 38, 238, 156, 239, 28, 244, 86, 176, 190, 207, 229, 200, 53, 15, 51, 195, 203, 231, 37, 129, 131, 70, 72, 127, 237, 111, 110, 4, 216, 88, 247, 133, 43, 195, 78, 63, 139, 129, 146, 34, 175, 147, 121, 63, 219, 93, 78, 153, 17, 114, 55, 176, 55, 204, 10, 49, 191, 133, 62, 215, 183, 97, 0, 247, 51, 241, 159, 254, 153, 132, 3, 63, 209, 255, 195, 96, 156, 141, 32, 245, 3, 178, 114, 78, 133, 178, 247, 1, 38, 150, 23, 23, 215, 214, 214, 54, 125, 205, 163, 178, 220, 223, 156, 87, 167, 63, 11, 155, 190, 195, 176, 19, 35, 19, 8, 49, 136, 230, 9, 223, 144, 56, 134, 106, 120, 17, 49, 7, 130, 163, 47, 250, 154, 23, 164, 254, 209, 164, 139, 251, 221, 166, 247, 255, 228, 19, 247, 153, 171, 0, 150, 0, 86, 180, 8, 126, 53, 20, 87, 199, 193, 86, 27, 12, 96, 186, 241, 52, 19, 253, 22, 32, 135, 127, 53, 100, 107, 78, 55, 29, 225, 18, 68, 232, 73, 92, 128, 28, 233, 12, 53, 181, 68, 180, 156, 185, 160, 200, 107, 77, 121, 211, 147, 211, 117, 225, 190, 190, 3, 111, 213, 167, 211, 251, 87, 113, 1, 91, 95, 159, 238, 174, 175, 63, 240, 218, 129, 213, 213, 238, 190, 183, 20, 223, 168, 63, 32, 60, 221, 221, 71, 162, 228, 84, 232, 75, 175, 236, 95, 173, 175, 95, 93, 93, 93, 193, 79, 160, 220, 7, 80, 203, 128, 252, 237, 192, 223, 127, 217, 101, 252, 7, 215, 182, 239, 4, 210, 155, 99, 37, 40, 52, 198, 61, 63, 165, 224, 18, 188, 59, 9, 47, 178, 120, 73, 192, 176, 197, 236, 225, 146, 146, 109, 142, 28, 94, 90, 26, 56, 232, 116, 228, 249, 7, 112, 26, 49, 89, 8, 218, 120, 216, 88, 15, 84, 218, 176, 5, 25, 15, 152, 239, 247, 218, 156, 204, 15, 13, 103, 126, 11, 125, 172, 107, 235, 157, 116, 52, 225, 55, 161, 12, 12, 50, 73, 142, 180, 176, 102, 226, 181, 41, 177, 44, 40, 124, 39, 242, 78, 89, 149, 37, 111, 61, 159, 61, 14, 158, 152, 95, 94, 198, 6, 134, 122, 112, 0, 198, 178, 174, 127, 75, 52, 188, 176, 54, 114, 102, 232, 132, 172, 245, 251, 151, 231, 100, 198, 191, 48, 114, 82, 45, 1, 79, 244, 207, 207, 97, 17, 205, 163, 13, 47, 8, 27, 57, 49, 144, 148, 175, 188, 221, 195, 68, 110, 25, 78, 50, 180, 99, 1, 10, 218, 93, 32, 107, 127, 160, 85, 15, 208, 63, 51, 63, 87, 78, 27, 231, 88, 231, 248, 71, 31, 161, 244, 105, 217, 220, 225, 123, 31, 118, 122, 106, 142, 112, 220, 171, 38, 252, 14, 127, 217, 149, 178, 233, 200, 71, 45, 17, 185, 42, 201, 37, 48, 242, 53, 45, 79, 34, 231, 170, 240, 88, 237, 116, 185, 13, 253, 62, 128, 130, 191, 150, 228, 64, 125, 122, 255, 91, 111, 9, 39, 125, 64, 248, 106, 150, 31, 208, 60, 167, 126, 125, 125, 95, 122, 63, 209, 63, 152, 207, 181, 93, 187, 90, 248, 18, 248, 248, 251, 207, 175, 154, 75, 129, 223, 175, 142, 255, 38, 250, 63, 148, 219, 255, 175, 181, 93, 208, 119, 2, 75, 208, 81, 169, 205, 176, 133, 75, 194, 165, 40, 152, 17, 62, 28, 211, 28, 5, 239, 187, 215, 110, 241, 0, 10, 183, 120, 169, 17, 8, 4, 14, 218, 202, 243, 252, 3, 208, 225, 217, 14, 116, 146, 97, 59, 192, 86, 97, 160, 254, 73, 217, 254, 247, 215, 217, 123, 157, 108, 2, 97, 10, 249, 45, 90, 154, 228, 147, 233, 160, 211, 125, 89, 66, 129, 16, 43, 132, 136, 0, 236, 244, 150, 125, 120, 252, 101, 178, 138, 28, 127, 227, 149, 119, 107, 238, 176, 34, 200, 202, 47, 127, 83, 172, 219, 165, 206, 220, 222, 162, 211, 111, 205, 13, 221, 190, 125, 123, 126, 126, 161, 235, 246, 60, 42, 159, 155, 95, 27, 218, 82, 5, 112, 177, 136, 229, 53, 37, 223, 90, 235, 186, 189, 44, 99, 2, 159, 135, 207, 223, 222, 218, 90, 94, 30, 72, 13, 187, 92, 195, 195, 213, 242, 147, 114, 15, 51, 181, 83, 241, 65, 248, 85, 142, 225, 242, 128, 196, 188, 14, 238, 202, 105, 19, 30, 251, 228, 92, 103, 231, 116, 231, 149, 113, 44, 97, 140, 192, 81, 132, 12, 153, 142, 140, 211, 207, 177, 225, 113, 135, 179, 108, 170, 165, 83, 226, 74, 77, 215, 140, 71, 58, 63, 145, 56, 34, 227, 17, 172, 36, 50, 70, 155, 72, 167, 199, 142, 171, 88, 92, 193, 138, 107, 88, 128, 79, 224, 189, 242, 76, 244, 209, 11, 98, 1, 153, 31, 102, 110, 230, 91, 192, 31, 172, 250, 191, 19, 112, 34, 252, 103, 158, 254, 111, 222, 250, 33, 131, 163, 216, 123, 46, 168, 227, 224, 64, 44, 78, 119, 46, 142, 25, 129, 176, 45, 16, 216, 87, 17, 176, 25, 89, 110, 179, 221, 205, 205, 250, 137, 152, 97, 51, 224, 5, 254, 1, 194, 226, 42, 25, 18, 150, 206, 206, 43, 67, 190, 18, 57, 239, 247, 179, 15, 228, 180, 171, 235, 225, 200, 243, 91, 12, 134, 56, 21, 247, 36, 156, 179, 24, 0, 26, 38, 36, 196, 203, 151, 137, 10, 20, 186, 67, 238, 21, 54, 4, 148, 214, 143, 95, 174, 189, 35, 235, 65, 45, 39, 102, 144, 161, 219, 121, 125, 244, 123, 206, 119, 241, 249, 115, 84, 157, 240, 53, 143, 208, 243, 225, 253, 62, 117, 6, 216, 213, 124, 98, 8, 46, 114, 133, 174, 195, 120, 19, 239, 57, 99, 126, 30, 142, 188, 63, 89, 173, 206, 120, 221, 213, 238, 234, 100, 94, 188, 0, 243, 9, 149, 130, 252, 120, 1, 46, 171, 77, 181, 209, 121, 14, 240, 233, 159, 108, 244, 127, 50, 21, 97, 159, 183, 179, 147, 79, 255, 166, 106, 216, 255, 189, 210, 234, 12, 49, 40, 76, 213, 16, 84, 108, 106, 44, 50, 134, 250, 207, 145, 155, 138, 68, 58, 167, 169, 55, 117, 206, 83, 132, 119, 112, 229, 44, 90, 58, 182, 164, 123, 229, 209, 76, 70, 44, 224, 158, 153, 11, 109, 199, 248, 34, 48, 251, 18, 56, 164, 12, 224, 99, 129, 24, 128, 165, 255, 91, 23, 190, 61, 214, 246, 235, 12, 57, 244, 47, 199, 193, 182, 88, 34, 204, 159, 51, 78, 175, 229, 128, 181, 72, 243, 112, 156, 241, 220, 22, 219, 39, 220, 208, 60, 102, 139, 33, 119, 90, 60, 22, 43, 240, 15, 128, 246, 65, 41, 126, 209, 73, 216, 24, 130, 43, 103, 250, 54, 182, 1, 253, 12, 9, 130, 252, 22, 94, 62, 25, 8, 120, 139, 156, 195, 151, 229, 30, 200, 27, 116, 112, 146, 151, 95, 9, 221, 161, 167, 179, 16, 160, 72, 248, 241, 178, 59, 199, 95, 177, 228, 239, 190, 219, 166, 219, 121, 7, 186, 228, 114, 175, 90, 230, 9, 186, 228, 71, 21, 240, 16, 222, 133, 28, 136, 92, 113, 41, 33, 111, 54, 128, 36, 171, 121, 147, 51, 152, 211, 155, 147, 73, 229, 33, 128, 128, 60, 59, 1, 217, 44, 117, 134, 115, 218, 216, 90, 85, 212, 168, 115, 18, 40, 72, 146, 41, 117, 224, 67, 22, 141, 79, 227, 249, 164, 238, 180, 216, 198, 115, 231, 68, 38, 143, 200, 21, 6, 12, 228, 87, 198, 58, 35, 145, 42, 163, 168, 15, 53, 226, 247, 127, 169, 175, 15, 117, 174, 172, 44, 173, 226, 10, 188, 47, 13, 175, 135, 167, 225, 34, 39, 173, 87, 60, 189, 68, 125, 166, 2, 22, 143, 102, 80, 99, 20, 19, 48, 141, 32, 122, 233, 214, 182, 5, 76, 177, 237, 247, 146, 224, 143, 223, 126, 59, 119, 200, 60, 4, 186, 117, 41, 154, 57, 133, 242, 193, 133, 40, 13, 219, 184, 18, 22, 11, 87, 26, 177, 112, 208, 136, 199, 226, 232, 202, 91, 18, 11, 215, 197, 133, 39, 98, 113, 186, 233, 189, 37, 78, 75, 254, 55, 92, 237, 3, 16, 186, 65, 223, 242, 119, 168, 148, 178, 93, 185, 78, 201, 113, 35, 4, 255, 17, 229, 29, 193, 112, 172, 248, 172, 7, 240, 29, 1, 193, 128, 60, 231, 111, 176, 121, 80, 87, 215, 81, 201, 237, 63, 207, 249, 114, 66, 207, 157, 71, 152, 149, 159, 45, 182, 5, 1, 71, 75, 206, 167, 39, 124, 61, 35, 236, 241, 158, 97, 217, 215, 211, 211, 35, 1, 100, 94, 20, 126, 114, 98, 243, 105, 10, 100, 193, 223, 160, 56, 242, 134, 172, 124, 136, 250, 34, 63, 233, 75, 37, 103, 209, 43, 202, 172, 78, 166, 158, 225, 74, 184, 21, 34, 112, 7, 112, 53, 252, 205, 156, 54, 88, 248, 179, 193, 58, 78, 185, 234, 120, 4, 129, 250, 223, 226, 40, 148, 253, 111, 58, 20, 93, 66, 197, 10, 0, 42, 68, 4, 96, 0, 212, 160, 87, 84, 164, 87, 81, 251, 70, 183, 107, 101, 37, 189, 178, 190, 177, 225, 90, 217, 230, 105, 139, 147, 105, 79, 167, 87, 211, 235, 27, 235, 237, 105, 204, 33, 13, 95, 5, 43, 235, 235, 237, 153, 99, 232, 81, 155, 0, 152, 57, 165, 108, 64, 92, 70, 31, 250, 195, 50, 128, 67, 240, 123, 110, 125, 21, 109, 179, 180, 47, 184, 206, 12, 114, 159, 97, 224, 113, 49, 94, 25, 79, 132, 19, 113, 195, 184, 55, 96, 196, 225, 137, 186, 120, 194, 8, 194, 43, 108, 166, 188, 78, 228, 134, 200, 225, 134, 211, 172, 79, 132, 12, 229, 31, 160, 234, 237, 166, 150, 113, 181, 170, 55, 163, 1, 82, 70, 186, 51, 215, 169, 202, 77, 141, 117, 142, 77, 143, 51, 108, 146, 180, 212, 114, 76, 82, 195, 65, 9, 140, 192, 90, 128, 155, 245, 60, 248, 150, 86, 194, 3, 80, 137, 193, 85, 106, 146, 163, 100, 154, 242, 218, 163, 205, 39, 27, 150, 101, 230, 191, 54, 49, 196, 22, 0, 75, 129, 147, 47, 64, 101, 37, 48, 49, 180, 54, 215, 220, 44, 242, 69, 37, 63, 156, 39, 167, 0, 121, 79, 42, 69, 63, 158, 173, 118, 39, 83, 73, 2, 136, 233, 176, 129, 130, 199, 115, 38, 255, 230, 19, 145, 10, 28, 232, 118, 229, 180, 225, 92, 84, 190, 255, 7, 232, 88, 158, 26, 100, 252, 6, 195, 157, 157, 10, 40, 94, 75, 178, 82, 49, 2, 26, 218, 239, 93, 39, 248, 19, 209, 127, 220, 235, 233, 213, 149, 141, 245, 13, 247, 6, 234, 134, 187, 254, 150, 175, 108, 184, 165, 62, 81, 166, 218, 87, 86, 251, 200, 172, 187, 174, 71, 51, 68, 11, 187, 110, 25, 129, 178, 129, 175, 110, 137, 9, 252, 121, 72, 129, 155, 194, 56, 16, 255, 234, 98, 94, 223, 7, 109, 143, 70, 197, 0, 226, 9, 15, 186, 174, 75, 4, 227, 28, 228, 199, 43, 98, 22, 175, 180, 120, 161, 60, 1, 55, 132, 43, 255, 0, 158, 211, 102, 0, 128, 74, 155, 25, 13, 144, 50, 210, 157, 185, 78, 85, 174, 134, 143, 73, 6, 67, 161, 198, 166, 42, 153, 240, 135, 66, 85, 131, 85, 4, 7, 227, 89, 213, 216, 88, 214, 212, 52, 88, 91, 251, 228, 160, 200, 28, 34, 104, 172, 226, 89, 214, 216, 216, 244, 228, 96, 107, 35, 194, 170, 198, 142, 9, 174, 247, 49, 217, 195, 6, 154, 23, 231, 73, 89, 247, 251, 76, 190, 220, 60, 39, 150, 177, 60, 138, 92, 48, 186, 108, 202, 23, 124, 204, 27, 201, 205, 139, 252, 4, 6, 144, 100, 236, 159, 77, 1, 215, 51, 89, 3, 208, 139, 192, 66, 40, 3, 120, 198, 149, 211, 6, 253, 2, 195, 95, 160, 99, 58, 58, 160, 139, 23, 97, 30, 90, 106, 150, 233, 90, 70, 24, 56, 237, 222, 238, 247, 87, 187, 187, 219, 247, 187, 187, 151, 164, 223, 119, 15, 239, 39, 42, 208, 221, 188, 27, 158, 222, 232, 110, 135, 83, 127, 99, 169, 219, 189, 177, 36, 227, 192, 134, 27, 3, 200, 92, 63, 118, 12, 11, 200, 179, 129, 76, 244, 210, 181, 155, 24, 129, 224, 161, 155, 215, 10, 181, 15, 218, 136, 38, 119, 61, 202, 28, 32, 17, 244, 4, 19, 193, 58, 84, 92, 135, 207, 149, 34, 99, 111, 92, 249, 7, 120, 182, 213, 163, 111, 249, 183, 170, 148, 178, 93, 185, 78, 85, 110, 176, 163, 188, 53, 36, 31, 77, 52, 54, 242, 96, 143, 76, 230, 118, 141, 33, 168, 199, 81, 54, 216, 138, 242, 165, 120, 48, 228, 160, 140, 121, 95, 83, 43, 41, 92, 174, 84, 242, 233, 65, 71, 220, 183, 118, 120, 219, 0, 150, 183, 182, 72, 151, 23, 142, 106, 3, 192, 34, 48, 128, 197, 209, 17, 20, 14, 22, 76, 3, 16, 57, 21, 193, 40, 22, 115, 194, 155, 74, 86, 164, 120, 249, 23, 121, 83, 100, 201, 88, 51, 0, 94, 247, 20, 152, 247, 4, 249, 81, 37, 112, 144, 211, 198, 110, 23, 253, 131, 60, 19, 32, 67, 36, 33, 80, 97, 119, 50, 181, 114, 234, 10, 74, 68, 86, 191, 8, 196, 0, 232, 225, 221, 221, 42, 26, 216, 6, 209, 227, 190, 150, 48, 113, 58, 122, 88, 46, 199, 32, 186, 225, 95, 19, 62, 110, 157, 250, 22, 119, 243, 10, 16, 175, 144, 215, 51, 25, 173, 217, 215, 163, 167, 212, 68, 112, 230, 90, 230, 226, 165, 153, 155, 55, 103, 46, 93, 204, 92, 187, 165, 38, 128, 167, 46, 178, 240, 87, 136, 210, 249, 249, 185, 30, 109, 195, 0, 18, 229, 137, 120, 34, 40, 61, 60, 24, 79, 20, 133, 247, 198, 149, 127, 128, 160, 199, 163, 87, 245, 173, 42, 165, 108, 87, 174, 83, 149, 11, 213, 149, 103, 253, 102, 121, 120, 86, 181, 134, 176, 4, 178, 173, 194, 26, 171, 228, 225, 41, 231, 143, 255, 6, 74, 39, 21, 211, 8, 57, 160, 100, 59, 140, 129, 247, 190, 248, 241, 140, 138, 0, 33, 49, 33, 37, 40, 208, 23, 207, 159, 81, 113, 1, 94, 164, 0, 254, 247, 114, 168, 146, 255, 36, 6, 112, 47, 58, 173, 16, 3, 160, 87, 231, 30, 250, 128, 44, 211, 5, 230, 181, 208, 156, 54, 69, 118, 153, 228, 26, 162, 214, 48, 31, 199, 102, 161, 95, 244, 94, 59, 115, 107, 13, 170, 25, 218, 8, 4, 42, 150, 140, 247, 145, 207, 4, 159, 202, 151, 71, 159, 126, 250, 1, 89, 13, 38, 27, 121, 96, 246, 33, 197, 212, 51, 119, 37, 117, 181, 225, 215, 193, 163, 252, 170, 132, 31, 197, 47, 98, 3, 168, 156, 129, 0, 156, 154, 121, 8, 156, 186, 136, 222, 145, 169, 39, 80, 79, 181, 19, 152, 143, 131, 123, 229, 114, 195, 191, 152, 57, 80, 135, 90, 213, 155, 183, 253, 41, 35, 221, 153, 235, 84, 114, 24, 128, 196, 89, 228, 12, 9, 59, 0, 114, 95, 26, 2, 115, 192, 5, 50, 54, 76, 78, 162, 110, 126, 41, 147, 186, 173, 173, 114, 165, 100, 146, 186, 65, 227, 85, 162, 63, 154, 32, 229, 23, 110, 21, 0, 75, 78, 185, 78, 7, 44, 249, 171, 104, 157, 142, 47, 125, 61, 229, 37, 157, 117, 23, 90, 0, 5, 36, 195, 10, 228, 213, 192, 144, 211, 70, 69, 9, 66, 177, 168, 214, 82, 46, 48, 204, 46, 142, 1, 144, 132, 195, 166, 148, 98, 176, 253, 122, 0, 118, 239, 146, 107, 22, 84, 43, 204, 110, 155, 28, 15, 65, 123, 33, 68, 164, 99, 204, 82, 71, 118, 143, 83, 143, 222, 141, 215, 183, 141, 224, 33, 13, 180, 159, 85, 122, 65, 173, 255, 110, 0, 118, 249, 167, 48, 220, 105, 63, 129, 42, 165, 108, 87, 174, 83, 149, 243, 116, 4, 131, 76, 160, 179, 110, 180, 128, 204, 168, 9, 183, 201, 165, 217, 14, 36, 20, 231, 1, 161, 44, 28, 38, 69, 22, 244, 59, 95, 5, 207, 235, 63, 129, 230, 104, 55, 159, 91, 117, 178, 6, 0, 207, 49, 128, 20, 35, 0, 169, 219, 236, 239, 59, 2, 83, 192, 0, 172, 54, 56, 62, 48, 167, 1, 244, 231, 44, 32, 138, 150, 86, 176, 115, 14, 41, 0, 18, 17, 130, 10, 47, 74, 5, 238, 66, 48, 204, 40, 96, 116, 40, 155, 74, 121, 34, 140, 69, 100, 8, 83, 50, 146, 171, 104, 208, 64, 50, 109, 185, 250, 197, 8, 24, 247, 181, 174, 53, 164, 110, 84, 106, 103, 218, 50, 60, 239, 57, 88, 160, 208, 189, 114, 89, 213, 19, 202, 93, 95, 244, 183, 171, 148, 178, 93, 185, 78, 85, 174, 4, 87, 99, 252, 89, 80, 212, 42, 83, 137, 37, 207, 103, 84, 123, 117, 239, 216, 151, 75, 188, 251, 80, 188, 130, 87, 1, 205, 90, 16, 69, 23, 112, 133, 156, 54, 242, 175, 43, 178, 155, 40, 42, 64, 133, 132, 90, 99, 136, 64, 215, 26, 118, 160, 40, 41, 98, 26, 75, 247, 255, 247, 24, 78, 69, 239, 134, 132, 7, 199, 2, 10, 129, 234, 69, 247, 5, 248, 11, 191, 251, 223, 84, 153, 36, 87, 100, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}

package main

import (
	"image"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic, err := loadPicture("thegopherproject.png")
	if err != nil {
		panic(err)
	}

	start := time.Now()
	var utime float32

	sprite := pixel.NewSprite(pic, pic.Bounds())
	sc := pixelgl.NewCanvas(win.Bounds())
	sc.SetFragmentShader(customFragShader)
	sc.BindUniform("u_time", &utime)
	sc.RecompileShader()

	sprite.Draw(sc, pixel.IM.Moved(win.Bounds().Center()))
	win.Clear(colornames.Greenyellow)

	for !win.Closed() {
		utime = float32(time.Since(start).Seconds())
		sprite.Draw(sc, pixel.IM.Moved(win.Bounds().Center()))
		sc.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

var customFragShader = `
#version 330 core

in vec4 Color;
in vec2 TexCoords;
in float Intensity;

out vec4 color;

uniform vec4 colorMask;
uniform vec4 texBounds;
uniform sampler2D tex;
uniform float u_time;

void main() {
	if (Intensity == 0) {
		color = colorMask * Color;
	} else {
		color = vec4(0, 0, 0, 0);
		color += (1 - Intensity) * Color;
		vec2 t = (TexCoords - texBounds.xy) / texBounds.zw;
		color += Intensity * Color * texture(tex, t);
		color *= colorMask;
	}
	color.rgb *= cos(u_time*5);
}
`

// var umouse = mgl32.Vec2{}
// umouse[0] = float32(win.MousePosition().X) / 1024
// umouse[1] = float32(win.MousePosition().Y) / 768

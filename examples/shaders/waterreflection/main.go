package main

import (
	"image"
	"math"
	"os"
	"time"

	"golang.org/x/image/colornames"

	_ "image/png"

	"github.com/faiface/pixel"
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

func run() {

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	pic, err := loadPicture("thegopherproject.png")
	if err != nil {
		panic(err)
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	var utime float32

	sprite := pixel.NewSprite(pic, pic.Bounds())

	sc := pixelgl.NewCanvas(pic.Bounds())

	wc := win.GetCanvas()
	wc.SetFragmentShader(reflectionShader)
	wc.BindUniform("u_time", &utime)
	wc.UpdateShader()

	curpos := pixel.V(sc.Bounds().Center().X, -25)
	tgtpos := pixel.V(sc.Bounds().Center().X, 316)
	last := start

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		curpos = pixel.Lerp(curpos, tgtpos, 1-math.Pow(1.0/16, dt*0.5))
		sc.Clear(colornames.Black)

		utime = float32(time.Since(start).Seconds())
		sprite.Draw(sc, pixel.IM.Moved(curpos))
		sc.Draw(wc, pixel.IM.Moved(wc.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

var reflectionShader = `
#version 330 core

in vec4 Color;
in vec2 texcoords;
in float Intensity;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform float u_time;

void main() {
	if (Intensity == 0) {
		fragColor = u_colormask * Color;
	} else {
		fragColor = vec4(0, 0, 0, 0);
		fragColor += (1 - Intensity) * Color;
		vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
		fragColor += Intensity * Color * texture(u_texture, t);
		fragColor *= u_colormask;
		vec2 uv =  t;
		vec3 overlayColor = vec3(0.1,0.1,1);
		float sepoffset = 0.005*cos(u_time*3.0);
		
		if (t.y < 0.3 + sepoffset)
		{
			float xoffset = 0.005*cos(u_time*3.0+200.0*t.y);
			float yoffset = ((0.3 - t.y)/0.3) * 0.05*(1.0+cos(u_time*3.0+50.0*t.y));
			fragColor = texture(u_texture, vec2(t.x+xoffset,t.y+yoffset));
		}
	}
}
`

// void main() {
// 	fragColor = vec4(0, 0, 0, 0);
// 	fragColor += (1 - Intensity) * incolor;
// 	vec2 t = (TexCoords) / texBounds.zw;
// 	fragColor += Intensity * incolor * texture(tex, t);
// 	fragColor *= colorMask;

// 	vec2 uv =  t;
// 	vec3 overlayColor = vec3(0.1,0.1,1);
// 	float sepoffset = 0.005*cos(u_time*3.0);

// 	if (t.y < 0.3 + sepoffset)
// 	{
//         float xoffset = 0.005*cos(u_time*3.0+200.0*t.y);
// 		float yoffset = ((0.3 - t.y)/0.3) * 0.05*(1.0+cos(u_time*3.0+50.0*t.y));
//         fragColor = texture(tex, vec2(t.x+xoffset,t.y+yoffset));
// 	}
// }
// `

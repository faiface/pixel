package main

import (
	"fmt"
	"image"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// InstallShader ...
func InstallShader(w *pixelgl.Window, weight *float32) {
	wc := w.GetCanvas()
	wc.SetFragmentShader(edgeDetectionFragShader)
	wc.BindUniform("u_weight", weight)
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

const (
	windowWidth      = 600
	windowHeight     = 450
	foregroundHeight = 149
	// This is the scrolling speed (pixels per second)
	// Negative values will make background to scroll to the left,
	// positive to the right.
	backgroundSpeed = -60
	foregroundSpeed = -120
)

func run() {
	fmt.Println("Use +/- to adjust weight")

	cfg := pixelgl.WindowConfig{
		Title:  "Parallax scrolling demo",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	var uWeight float32

	uWeight = 4.0

	InstallShader(win, &uWeight)

	// Pic must have double the width of the window, as it will scroll to the left or right
	picBackground, err := loadPicture("background.png")
	if err != nil {
		panic(err)
	}
	picForeground, err := loadPicture("foreground.png")
	if err != nil {
		panic(err)
	}

	background := NewScrollingBackground(picBackground, windowWidth, windowHeight, backgroundSpeed)
	foreground := NewScrollingBackground(picForeground, windowWidth, foregroundHeight, foregroundSpeed)
	win.SetTitle(fmt.Sprint("Weight: ", uWeight))
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		background.Update(win, dt)
		foreground.Update(win, dt)

		if win.Pressed(pixelgl.KeyEqual) {
			uWeight += 0.1
			win.SetTitle(fmt.Sprint("Weight: ", uWeight))
		}
		if win.Pressed(pixelgl.KeyMinus) {
			uWeight -= 0.1
			win.SetTitle(fmt.Sprint("Weight: ", uWeight))
		}

		win.Update()

	}
}

func main() {
	pixelgl.Run(run)
}

var edgeDetectionFragShader = `
#version 330 core

#ifdef GL_ES
precision mediump float;
precision mediump int;
#endif

in vec2 texcoords;

out vec4 fragColor;

uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform float u_weight;

void main(void)
{
	vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
	vec2 onePixel = vec2(1.0, 1.0) / u_texbounds.zw;
	vec4 colorSum =
	  texture(u_texture, t + onePixel * vec2(-1, -1)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2( 0, -1)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2( 1, -1)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2(-1,  0)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2( 0,  0)) *  8.0 +
	  texture(u_texture, t + onePixel * vec2( 1,  0)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2(-1,  1)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2( 0,  1)) * -0.1 +
	  texture(u_texture, t + onePixel * vec2( 1,  1)) * -0.1 ;
  
	fragColor = (colorSum / u_weight).rgba;
}
`

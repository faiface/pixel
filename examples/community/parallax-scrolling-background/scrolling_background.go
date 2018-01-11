package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type scrollingBackground struct {
	width               float64
	height              float64
	windowWidth         float64
	displacementCounter float64
	backgrounds         [2]*pixel.Sprite
	positions           [2]pixel.Vec
}

func newScrollingBackground(pic pixel.Picture, width, height, windowWidth float64) *scrollingBackground {
	return &scrollingBackground{
		width:       width,
		height:      height,
		windowWidth: windowWidth,
		backgrounds: [2]*pixel.Sprite{
			pixel.NewSprite(pic, pixel.R(0, 0, width, height)),
			pixel.NewSprite(pic, pixel.R(width, 0, width*2, height)),
		},
		positions: [2]pixel.Vec{
			pixel.V(width/2, (height/2)+1),
			pixel.V(width+(width/2), (height/2)+1),
		},
	}
}

func (sb *scrollingBackground) update(win *pixelgl.Window, speed, dt float64) {
	if sb.displacementCounter <= -sb.windowWidth {
		sb.displacementCounter = 0
		sb.positions[0], sb.positions[1] = sb.positions[1], sb.positions[0]
	}
	d := pixel.V(-sb.displacementCounter, 0)
	sb.backgrounds[0].Draw(win, pixel.IM.Moved(sb.positions[0].Sub(d)))
	sb.backgrounds[1].Draw(win, pixel.IM.Moved(sb.positions[1].Sub(d)))
	sb.displacementCounter -= speed * dt
}

package main

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// ScrollingBackground stores all needed information to scroll a background
// to the left or right
type ScrollingBackground struct {
	width        float64
	height       float64
	displacement float64
	speed        float64
	backgrounds  [2]*pixel.Sprite
	positions    [2]pixel.Vec
}

// NewScrollingBackground construct and returns a new instance of scrollingBackground,
// positioning the background images according to the speed value
func NewScrollingBackground(pic pixel.Picture, width, height, speed float64) *ScrollingBackground {
	sb := &ScrollingBackground{
		width:  width,
		height: height,
		speed:  speed,
		backgrounds: [2]*pixel.Sprite{
			pixel.NewSprite(pic, pixel.R(0, 0, width, height)),
			pixel.NewSprite(pic, pixel.R(width, 0, width*2, height)),
		},
	}

	sb.positionImages()
	return sb
}

// If scrolling speed > 0, put second background image ouside the screen,
// at the left side, otherwise put it at the right side.
func (sb *ScrollingBackground) positionImages() {
	if sb.speed > 0 {
		sb.positions = [2]pixel.Vec{
			pixel.V(sb.width/2, sb.height/2),
			pixel.V((sb.width/2)-sb.width, sb.height/2),
		}
	} else {
		sb.positions = [2]pixel.Vec{
			pixel.V(sb.width/2, sb.height/2),
			pixel.V(sb.width+(sb.width/2), sb.height/2),
		}
	}
}

// Update will move backgrounds certain pixels, depending of the amount of time passed
func (sb *ScrollingBackground) Update(win *pixelgl.Window, dt float64) {
	if math.Abs(sb.displacement) >= sb.width {
		sb.displacement = 0
		sb.positions[0], sb.positions[1] = sb.positions[1], sb.positions[0]
	}
	d := pixel.V(sb.displacement, 0)
	sb.backgrounds[0].Draw(win, pixel.IM.Moved(sb.positions[0].Add(d)))
	sb.backgrounds[1].Draw(win, pixel.IM.Moved(sb.positions[1].Add(d)))
	sb.displacement += sb.speed * dt
}

package pixel

import "image/color"

// colorToRGBA converts a color from image/color to RGBA components in interval [0, 1].
func colorToRGBA(c color.Color) (r, g, b, a float32) {
	ri, gi, bi, ai := c.RGBA()
	r = float32(ri) / 0xffff
	g = float32(gi) / 0xffff
	b = float32(bi) / 0xffff
	a = float32(ai) / 0xffff
	return
}

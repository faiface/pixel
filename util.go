package pixel

import "image/color"

// colorToRGBA converts a color from image/color to RGBA components in interval [0, 1)
func colorToRGBA(c color.Color) (r, g, b, a float64) {
	ri, gi, bi, ai := c.RGBA()
	r = float64(ri) / 0xffff
	g = float64(gi) / 0xffff
	b = float64(bi) / 0xffff
	a = float64(ai) / 0xffff
	return
}

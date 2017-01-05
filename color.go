package pixel

import "image/color"

// NRGBA represents a non-alpha-premultiplied RGBA color with components within range [0, 1].
//
// The difference between color.NRGBA is that the value range is [0, 1] and the values are floats.
type NRGBA struct {
	R, G, B, A float64
}

// Add adds color d to color c component-wise and returns the result (the components are not
// clamped).
func (c NRGBA) Add(d NRGBA) NRGBA {
	return NRGBA{
		R: c.R + d.R,
		G: c.G + d.G,
		B: c.B + d.B,
		A: c.A + d.A,
	}
}

// Sub subtracts color d from color c component-wise and returns the result (the components
// are not clamped).
func (c NRGBA) Sub(d NRGBA) NRGBA {
	return NRGBA{
		R: c.R - d.R,
		G: c.G - d.G,
		B: c.B - d.B,
		A: c.A - d.A,
	}
}

// Mul multiplies color c by color d component-wise (the components are not clamped).
func (c NRGBA) Mul(d NRGBA) NRGBA {
	return NRGBA{
		R: c.R * d.R,
		G: c.G * d.G,
		B: c.B * d.B,
		A: c.A * d.A,
	}
}

// Scaled multiplies each component of color c by scale and returns the result (the components
// are not clamped).
func (c NRGBA) Scaled(scale float64) NRGBA {
	return NRGBA{
		R: c.R * scale,
		G: c.G * scale,
		B: c.B * scale,
		A: c.A * scale,
	}
}

// RGBA returns alpha-premultiplied red, green, blue and alpha components of a color.
func (c NRGBA) RGBA() (r, g, b, a uint32) {
	c.R = clamp(c.R, 0, 1)
	c.G = clamp(c.G, 0, 1)
	c.B = clamp(c.B, 0, 1)
	c.A = clamp(c.A, 0, 1)
	r = uint32(0xffff * c.R * c.A)
	g = uint32(0xffff * c.G * c.A)
	b = uint32(0xffff * c.B * c.A)
	a = uint32(0xffff * c.A)
	return
}

// NRGBAModel converts colors to NRGBA format.
var NRGBAModel = color.ModelFunc(nrgbaModel)

func nrgbaModel(c color.Color) color.Color {
	if c, ok := c.(NRGBA); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return NRGBA{0, 0, 0, 0}
	}
	return NRGBA{
		float64(r) / float64(a),
		float64(g) / float64(a),
		float64(b) / float64(a),
		float64(a) / 0xffff,
	}
}

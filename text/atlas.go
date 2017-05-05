package text

import (
	"image"
	"image/draw"
	"unicode"

	"github.com/faiface/pixel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	Orig    pixel.Vec
	Frame   pixel.Rect
	Advance float64
}

type Atlas struct {
	pic        pixel.Picture
	mapping    map[rune]Glyph
	kern       map[struct{ r0, r1 rune }]float64
	ascent     float64
	descent    float64
	lineHeight float64
}

func NewAtlas(face font.Face, runes []rune) *Atlas {
	//FIXME: don't put glyphs in just one row, make a square

	width := fixed.Int26_6(0)
	for _, r := range runes {
		b, _, ok := face.GlyphBounds(r)
		if !ok && r != unicode.ReplacementChar {
			continue
		}
		width += b.Max.X - b.Min.X

		// padding to avoid filtering artifacts
		width = fixed.I(width.Ceil())
		width += fixed.I(2)
	}

	atlasImg := image.NewRGBA(image.Rect(
		0, 0,
		width.Ceil(), (face.Metrics().Ascent + face.Metrics().Descent).Ceil(),
	))
	atlasHeight := float64(atlasImg.Bounds().Dy())

	mapping := make(map[rune]Glyph)

	dot := fixed.Point26_6{
		X: 0,
		Y: face.Metrics().Ascent,
	}

	for _, r := range runes {
		b, _, ok := face.GlyphBounds(r)
		if !ok && r != unicode.ReplacementChar {
			continue
		}

		dot.X -= b.Min.X

		dr, mask, maskp, _, _ := face.Glyph(dot, r)
		draw.Draw(atlasImg, dr, mask, maskp, draw.Src)

		orig := pixel.V(
			float64(dot.X)/(1<<6),
			atlasHeight-float64(dot.Y)/(1<<6),
		)

		frame := pixel.R(
			float64(dr.Min.X),
			atlasHeight-float64(dr.Min.Y),
			float64(dr.Max.X),
			atlasHeight-float64(dr.Max.Y),
		).Norm()

		adv, _ := face.GlyphAdvance(r)
		advance := float64(adv) / (1 << 6)

		mapping[r] = Glyph{orig, frame, advance}

		dot.X += b.Max.X

		// padding
		dot.X = fixed.I(dot.X.Ceil())
		dot.X += fixed.I(2)
	}

	kern := make(map[struct{ r0, r1 rune }]float64)
	for _, r0 := range runes {
		for _, r1 := range runes {
			kern[struct{ r0, r1 rune }{r0, r1}] = float64(face.Kern(r0, r1)) / (1 << 6)
		}
	}

	return &Atlas{
		pic:        pixel.PictureDataFromImage(atlasImg),
		mapping:    mapping,
		kern:       kern,
		ascent:     float64(face.Metrics().Ascent) / (1 << 6),
		descent:    float64(face.Metrics().Descent) / (1 << 6),
		lineHeight: float64(face.Metrics().Height) / (1 << 6),
	}
}

func (a *Atlas) Picture() pixel.Picture {
	return a.pic
}

func (a *Atlas) Contains(r rune) bool {
	_, ok := a.mapping[r]
	return ok
}

func (a *Atlas) Glyph(r rune) Glyph {
	return a.mapping[r]
}

func (a *Atlas) Kern(r0, r1 rune) float64 {
	return a.kern[struct{ r0, r1 rune }{r0, r1}]
}

func (a *Atlas) Ascent() float64 {
	return a.ascent
}

func (a *Atlas) Descent() float64 {
	return a.descent
}

func (a *Atlas) LineHeight() float64 {
	return a.lineHeight
}

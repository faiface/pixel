package text

import (
	"image"
	"image/draw"
	"sort"
	"unicode"

	"github.com/faiface/pixel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	Dot     pixel.Vec
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
	fixedMapping, fixedBounds := makeSquareMapping(face, runes, fixed.I(2))

	atlasImg := image.NewRGBA(image.Rect(
		fixedBounds.Min.X.Floor(),
		fixedBounds.Min.Y.Floor(),
		fixedBounds.Max.X.Ceil(),
		fixedBounds.Max.Y.Ceil(),
	))

	for r, fg := range fixedMapping {
		dr, mask, maskp, _, _ := face.Glyph(fg.dot, r)
		draw.Draw(atlasImg, dr, mask, maskp, draw.Src)
	}

	bounds := pixel.Rect{}
	for _, fg := range fixedMapping {
		b := pixel.R(
			i2f(fg.frame.Min.X),
			i2f(fg.frame.Min.Y),
			i2f(fg.frame.Max.X),
			i2f(fg.frame.Max.Y),
		)
		bounds = bounds.Union(b)
	}

	mapping := make(map[rune]Glyph)
	for r, fg := range fixedMapping {
		mapping[r] = Glyph{
			Dot: pixel.V(
				i2f(fg.dot.X),
				bounds.Max.Y()-(i2f(fg.dot.Y)-bounds.Min.Y()),
			),
			Frame: pixel.R(
				i2f(fg.frame.Min.X),
				bounds.Max.Y()-(i2f(fg.frame.Min.Y)-bounds.Min.Y()),
				i2f(fg.frame.Max.X),
				bounds.Max.Y()-(i2f(fg.frame.Max.Y)-bounds.Min.Y()),
			).Norm(),
			Advance: i2f(fg.advance),
		}
	}

	kern := make(map[struct{ r0, r1 rune }]float64)
	for _, r0 := range runes {
		for _, r1 := range runes {
			kern[struct{ r0, r1 rune }{r0, r1}] = i2f(face.Kern(r0, r1))
		}
	}

	return &Atlas{
		pic:        pixel.PictureDataFromImage(atlasImg),
		mapping:    mapping,
		kern:       kern,
		ascent:     i2f(face.Metrics().Ascent),
		descent:    i2f(face.Metrics().Descent),
		lineHeight: i2f(face.Metrics().Height),
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

func (a *Atlas) DrawRune(prevR, r rune, dot pixel.Vec) (rect, frame, bounds pixel.Rect, newDot pixel.Vec) {
	if !a.Contains(r) {
		r = unicode.ReplacementChar
	}
	if !a.Contains(unicode.ReplacementChar) {
		return pixel.Rect{}, pixel.Rect{}, pixel.Rect{}, dot
	}
	if !a.Contains(prevR) {
		prevR = unicode.ReplacementChar
	}

	if prevR >= 0 {
		dot += pixel.X(a.Kern(prevR, r))
	}

	glyph := a.Glyph(r)

	rect = glyph.Frame.Moved(dot - glyph.Dot)
	bounds = rect

	if bounds.W()*bounds.H() != 0 {
		bounds = pixel.R(
			bounds.Min.X(),
			dot.Y()-a.Descent(),
			bounds.Max.X(),
			dot.Y()+a.Ascent(),
		)
	}

	dot += pixel.X(glyph.Advance)

	return rect, glyph.Frame, bounds, dot
}

type fixedGlyph struct {
	dot     fixed.Point26_6
	frame   fixed.Rectangle26_6
	advance fixed.Int26_6
}

func makeSquareMapping(face font.Face, runes []rune, padding fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	width := sort.Search(int(fixed.I(1024*1024)), func(i int) bool {
		width := fixed.Int26_6(i)
		_, bounds := makeMapping(face, runes, padding, width)
		return bounds.Max.X-bounds.Min.X >= bounds.Max.Y-bounds.Min.Y
	})
	return makeMapping(face, runes, padding, fixed.Int26_6(width))
}

func makeMapping(face font.Face, runes []rune, padding, width fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	mapping := make(map[rune]fixedGlyph)
	bounds := fixed.Rectangle26_6{}

	dot := fixed.P(0, 0)

	for _, r := range runes {
		// face.Glyph gives more useful results for drawing than face.GlyphBounds
		dr, _, _, advance, ok := face.Glyph(fixed.P(0, 0), r)
		if !ok {
			continue
		}

		frame := fixed.Rectangle26_6{
			Min: fixed.P(dr.Min.X, dr.Min.Y),
			Max: fixed.P(dr.Max.X, dr.Max.Y),
		}

		dot.X -= frame.Min.X
		frame = frame.Add(dot)

		mapping[r] = fixedGlyph{
			dot:     dot,
			frame:   frame,
			advance: advance,
		}
		bounds = bounds.Union(frame)

		dot.X = frame.Max.X

		// padding + align to integer
		dot.X += padding
		dot.X = fixed.I(dot.X.Ceil())

		// width exceeded, new row
		if frame.Max.X >= width {
			dot.X = 0
			dot.Y += face.Metrics().Ascent + face.Metrics().Descent

			// padding + align to integer
			dot.Y += padding
			dot.Y = fixed.I(dot.Y.Ceil())
		}
	}

	return mapping, bounds
}

func i2f(i fixed.Int26_6) float64 {
	return float64(i) / (1 << 6)
}

func f2i(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * (1 << 6))
}

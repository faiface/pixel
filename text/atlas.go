package text

import (
	"fmt"
	"image"
	"image/draw"
	"sort"
	"unicode"

	"github.com/faiface/pixel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Glyph describes one glyph in an Atlas.
type Glyph struct {
	Dot     pixel.Vec
	Frame   pixel.Rect
	Advance float64
}

// Atlas is a set of pre-drawn glyphs of a fixed set of runes. This allows for efficient text drawing.
type Atlas struct {
	face       font.Face
	pic        pixel.Picture
	mapping    map[rune]Glyph
	ascent     float64
	descent    float64
	lineHeight float64
}

// NewAtlas creates a new Atlas containing glyphs of the union of the given sets of runes (plus
// unicode.ReplacementChar) from the given font face.
//
// Creating an Atlas is rather expensive, do not create a new Atlas each frame.
//
// Do not destroy or close the font.Face after creating the Atlas. Atlas still uses it.
func NewAtlas(face font.Face, runeSets ...[]rune) *Atlas {
	seen := make(map[rune]bool)
	runes := []rune{unicode.ReplacementChar}
	for _, set := range runeSets {
		for _, r := range set {
			if !seen[r] {
				runes = append(runes, r)
				seen[r] = true
			}
		}
	}

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

	bounds := pixel.R(
		i2f(fixedBounds.Min.X),
		i2f(fixedBounds.Min.Y),
		i2f(fixedBounds.Max.X),
		i2f(fixedBounds.Max.Y),
	)

	mapping := make(map[rune]Glyph)
	for r, fg := range fixedMapping {
		mapping[r] = Glyph{
			Dot: pixel.V(
				i2f(fg.dot.X),
				bounds.Max.Y-(i2f(fg.dot.Y)-bounds.Min.Y),
			),
			Frame: pixel.R(
				i2f(fg.frame.Min.X),
				bounds.Max.Y-(i2f(fg.frame.Min.Y)-bounds.Min.Y),
				i2f(fg.frame.Max.X),
				bounds.Max.Y-(i2f(fg.frame.Max.Y)-bounds.Min.Y),
			).Norm(),
			Advance: i2f(fg.advance),
		}
	}

	return &Atlas{
		face:       face,
		pic:        pixel.PictureDataFromImage(atlasImg),
		mapping:    mapping,
		ascent:     i2f(face.Metrics().Ascent),
		descent:    i2f(face.Metrics().Descent),
		lineHeight: i2f(face.Metrics().Height),
	}
}

// Picture returns the underlying Picture containing an arrangement of all the glyphs contained
// within the Atlas.
func (a *Atlas) Picture() pixel.Picture {
	return a.pic
}

// Contains reports wheter r in contained within the Atlas.
func (a *Atlas) Contains(r rune) bool {
	_, ok := a.mapping[r]
	return ok
}

// Glyph returns the description of r within the Atlas.
func (a *Atlas) Glyph(r rune) Glyph {
	return a.mapping[r]
}

// Kern returns the kerning distance between runes r0 and r1. Positive distance means that the
// glyphs should be further apart.
func (a *Atlas) Kern(r0, r1 rune) float64 {
	return i2f(a.face.Kern(r0, r1))
}

// Ascent returns the distance from the top of the line to the baseline.
func (a *Atlas) Ascent() float64 {
	return a.ascent
}

// Descent returns the distance from the baseline to the bottom of the line.
func (a *Atlas) Descent() float64 {
	return a.descent
}

// LineHeight returns the recommended vertical distance between two lines of text.
func (a *Atlas) LineHeight() float64 {
	return a.lineHeight
}

// DrawRune returns parameters necessary for drawing a rune glyph.
//
// Rect is a rectangle where the glyph should be positioned. Frame is the glyph frame inside the
// Atlas's Picture. NewDot is the new position of the dot.
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
		dot.X += a.Kern(prevR, r)
	}

	glyph := a.Glyph(r)

	rect = glyph.Frame.Moved(dot.Sub(glyph.Dot))
	bounds = rect

	if bounds.W()*bounds.H() != 0 {
		bounds = pixel.R(
			bounds.Min.X,
			dot.Y-a.Descent(),
			bounds.Max.X,
			dot.Y+a.Ascent(),
		)
	}

	dot.X += glyph.Advance

	return rect, glyph.Frame, bounds, dot
}

type fixedGlyph struct {
	dot     fixed.Point26_6
	frame   fixed.Rectangle26_6
	advance fixed.Int26_6
}

// makeSquareMapping finds an optimal glyph arrangement of the given runes, so that their common
// bounding box is as square as possible.
func makeSquareMapping(face font.Face, runes []rune, padding fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	width := sort.Search(int(fixed.I(1024*1024)), func(i int) bool {
		width := fixed.Int26_6(i)
		_, bounds := makeMapping(face, runes, padding, width)
		return bounds.Max.X-bounds.Min.X >= bounds.Max.Y-bounds.Min.Y
	})
	return makeMapping(face, runes, padding, fixed.Int26_6(width))
}

// makeMapping arranges glyphs of the given runes into rows in such a way, that no glyph is located
// fully to the right of the specified width. Specifically, it places glyphs in a row one by one and
// once it reaches the specified width, it starts a new row.
func makeMapping(face font.Face, runes []rune, padding, width fixed.Int26_6) (map[rune]fixedGlyph, fixed.Rectangle26_6) {
	mapping := make(map[rune]fixedGlyph)
	bounds := fixed.Rectangle26_6{}

	dot := fixed.P(0, 0)

	for _, r := range runes {
		b, advance, ok := face.GlyphBounds(r)
		if !ok {
			fmt.Println(r)
			continue
		}

		// this is important for drawing, artifacts arise otherwise
		frame := fixed.Rectangle26_6{
			Min: fixed.P(b.Min.X.Floor(), b.Min.Y.Floor()),
			Max: fixed.P(b.Max.X.Ceil(), b.Max.Y.Ceil()),
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

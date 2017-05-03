package text

import (
	"image"
	"image/color"
	"image/draw"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/pixel"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var ASCII []rune

func init() {
	ASCII = make([]rune, unicode.MaxASCII-32)
	for i := range ASCII {
		ASCII[i] = rune(32 + i)
	}
}

func RangeTable(table *unicode.RangeTable) []rune {
	var runes []rune
	for _, rng := range table.R16 {
		for r := rng.Lo; r <= rng.Hi; r += rng.Stride {
			runes = append(runes, rune(r))
		}
	}
	for _, rng := range table.R32 {
		for r := rng.Lo; r <= rng.Hi; r += rng.Stride {
			runes = append(runes, rune(r))
		}
	}
	return runes
}

type Text struct {
	Orig pixel.Vec
	Dot  pixel.Vec

	color pixel.RGBA

	prevR rune
	atlas atlas
	glyph pixel.TrianglesData
	tris  pixel.TrianglesData
	d     pixel.Drawer
}

func New(face font.Face, runeSets ...[]rune) *Text {
	runes := []rune{unicode.ReplacementChar}
	for _, set := range runeSets {
		runes = append(runes, set...)
	}

	txt := &Text{
		color: pixel.Alpha(1),
		atlas: makeAtlas(face, runes),
	}
	txt.glyph.SetLen(6)
	txt.d.Picture = txt.atlas.pic
	txt.d.Triangles = &txt.tris

	txt.Clear()

	return txt
}

func (txt *Text) Color(c color.Color) {
	txt.color = pixel.ToRGBA(c)
}

func (txt *Text) Clear() {
	txt.prevR = -1
	txt.tris.SetLen(0)
	txt.d.Dirty()
}

func (txt *Text) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	n = len(p)

	for i := range txt.glyph {
		txt.glyph[i].Color = txt.color
		txt.glyph[i].Intensity = 1
	}

	for len(p) > 0 {
		r, size := utf8.DecodeRune(p)
		p = p[size:]

		switch r {
		case '\n':
			txt.Dot -= pixel.Y(txt.atlas.lineHeight)
			txt.Dot = txt.Dot.WithX(txt.Orig.X())
			continue
		case '\r':
			txt.Dot = txt.Dot.WithX(txt.Orig.X())
			continue
		case '\t':
			//TODO
			continue
		}

		glyph, ok := txt.atlas.mapping[r]
		if !ok {
			glyph = txt.atlas.mapping[unicode.ReplacementChar]
		}

		if txt.prevR >= 0 {
			txt.Dot += pixel.X(txt.atlas.kern[struct{ r0, r1 rune }{txt.prevR, r}])
		}

		a := pixel.V(glyph.frame.Min.X(), glyph.frame.Min.Y())
		b := pixel.V(glyph.frame.Max.X(), glyph.frame.Min.Y())
		c := pixel.V(glyph.frame.Max.X(), glyph.frame.Max.Y())
		d := pixel.V(glyph.frame.Min.X(), glyph.frame.Max.Y())

		for i, v := range []pixel.Vec{a, b, c, a, c, d} {
			txt.glyph[i].Position = v - glyph.orig + txt.Dot
			txt.glyph[i].Picture = v
		}

		txt.tris = append(txt.tris, txt.glyph...)

		txt.Dot += pixel.X(glyph.advance)
		txt.prevR = r
	}

	txt.d.Dirty()

	return n, nil
}

func (txt *Text) Draw(t pixel.Target) {
	txt.d.Draw(t)
}

type atlas struct {
	pic        pixel.Picture
	mapping    map[rune]glyph
	kern       map[struct{ r0, r1 rune }]float64
	lineHeight float64
}

type glyph struct {
	orig    pixel.Vec
	frame   pixel.Rect
	advance float64
}

func makeAtlas(face font.Face, runes []rune) atlas {
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

	mapping := make(map[rune]glyph)

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

		mapping[r] = glyph{orig, frame, advance}

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

	return atlas{
		pixel.PictureDataFromImage(atlasImg),
		mapping,
		kern,
		float64(face.Metrics().Height) / (1 << 6),
	}
}

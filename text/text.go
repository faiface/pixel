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

	atlas *Atlas

	color      pixel.RGBA
	lineHeight float64
	tabWidth   float64

	prevR rune
	glyph pixel.TrianglesData
	tris  pixel.TrianglesData
	d     pixel.Drawer
	trans *pixel.Batch
}

func New(face font.Face, runeSets ...[]rune) *Text {
	runes := []rune{unicode.ReplacementChar}
	for _, set := range runeSets {
		runes = append(runes, set...)
	}

	atlas := NewAtlas(face, runes)

	txt := &Text{
		atlas:      atlas,
		color:      pixel.Alpha(1),
		lineHeight: atlas.LineHeight(),
		tabWidth:   atlas.Glyph(' ').Advance * 4,
	}
	txt.glyph.SetLen(6)
	txt.d.Picture = txt.atlas.pic
	txt.d.Triangles = &txt.tris
	txt.trans = pixel.NewBatch(&pixel.TrianglesData{}, atlas.pic)

	txt.Clear()

	return txt
}

func (txt *Text) Atlas() *Atlas {
	return txt.atlas
}

func (txt *Text) SetMatrix(m pixel.Matrix) {
	txt.trans.SetMatrix(m)
}

func (txt *Text) SetColorMask(c color.Color) {
	txt.trans.SetColorMask(c)
}

func (txt *Text) Color(c color.Color) {
	txt.color = pixel.ToRGBA(c)
}

func (txt *Text) LineHeight(scale float64) {
	txt.lineHeight = scale
}

func (txt *Text) TabWidth(width float64) {
	txt.tabWidth = width
}

func (txt *Text) Clear() {
	txt.prevR = -1
	txt.tris.SetLen(0)
	txt.d.Dirty()
}

func (txt *Text) Write(p []byte) (n int, err error) {
	n, err = len(p), nil // always returns this

	if len(p) == 0 {
		return
	}

	for i := range txt.glyph {
		txt.glyph[i].Color = txt.color
		txt.glyph[i].Intensity = 1
	}

	for len(p) > 0 {
		r, size := utf8.DecodeRune(p)
		p = p[size:]
		txt.WriteRune(r)
	}

	return
}

func (txt *Text) WriteString(s string) (n int, err error) {
	if len(s) == 0 {
		return
	}

	for i := range txt.glyph {
		txt.glyph[i].Color = txt.color
		txt.glyph[i].Intensity = 1
	}

	for _, r := range s {
		txt.WriteRune(r)
	}

	return len(s), nil
}

func (txt *Text) WriteByte(c byte) error {
	_, err := txt.WriteRune(rune(c))
	return err
}

func (txt *Text) WriteRune(r rune) (n int, err error) {
	n, err = utf8.RuneLen(r), nil // always returns this

	switch r {
	case '\n':
		txt.Dot -= pixel.Y(txt.lineHeight)
		txt.Dot = txt.Dot.WithX(txt.Orig.X())
		return
	case '\r':
		txt.Dot = txt.Dot.WithX(txt.Orig.X())
		return
	case '\t':
		//TODO: properly align tab
		txt.Dot += pixel.X(txt.tabWidth)
		return
	}

	if !txt.atlas.Contains(r) {
		r = unicode.ReplacementChar
	}
	if !txt.atlas.Contains(unicode.ReplacementChar) {
		return
	}

	glyph := txt.atlas.Glyph(r)

	if txt.prevR >= 0 {
		txt.Dot += pixel.X(txt.atlas.Kern(txt.prevR, r))
	}

	a := pixel.V(glyph.Frame.Min.X(), glyph.Frame.Min.Y())
	b := pixel.V(glyph.Frame.Max.X(), glyph.Frame.Min.Y())
	c := pixel.V(glyph.Frame.Max.X(), glyph.Frame.Max.Y())
	d := pixel.V(glyph.Frame.Min.X(), glyph.Frame.Max.Y())

	for i, v := range []pixel.Vec{a, b, c, a, c, d} {
		txt.glyph[i].Position = v - glyph.Orig + txt.Dot
		txt.glyph[i].Picture = v
	}

	txt.tris = append(txt.tris, txt.glyph...)

	txt.Dot += pixel.X(glyph.Advance)
	txt.prevR = r

	txt.d.Dirty()

	return
}

func (txt *Text) Draw(t pixel.Target) {
	txt.trans.Clear()
	txt.d.Draw(txt.trans)
	txt.trans.Draw(t)
}

type Glyph struct {
	Orig    pixel.Vec
	Frame   pixel.Rect
	Advance float64
}

type Atlas struct {
	pic        pixel.Picture
	mapping    map[rune]Glyph
	kern       map[struct{ r0, r1 rune }]float64
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
		pixel.PictureDataFromImage(atlasImg),
		mapping,
		kern,
		float64(face.Metrics().Height) / (1 << 6),
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

func (a *Atlas) LineHeight() float64 {
	return a.lineHeight
}

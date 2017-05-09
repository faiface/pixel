package text

import (
	"image/color"
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/pixel"
	"golang.org/x/image/font"
)

// ASCII is a set of all ASCII runes. These runes are codepoints from 32 to 127 inclusive.
var ASCII []rune

func init() {
	ASCII = make([]rune, unicode.MaxASCII-32)
	for i := range ASCII {
		ASCII[i] = rune(32 + i)
	}
}

// RangeTable takes a *unicode.RangeTable and generates a set of runes contained within that
// RangeTable.
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

	lineHeight float64
	tabWidth   float64

	buf    []byte
	prevR  rune
	bounds pixel.Rect
	glyph  pixel.TrianglesData
	tris   pixel.TrianglesData

	mat    pixel.Matrix
	col    pixel.RGBA
	trans  pixel.TrianglesData
	transD pixel.Drawer
	dirty  bool
}

func New(face font.Face, runeSets ...[]rune) *Text {
	runes := []rune{unicode.ReplacementChar}
	for _, set := range runeSets {
		runes = append(runes, set...)
	}

	atlas := NewAtlas(face, runes)

	txt := &Text{
		atlas:      atlas,
		lineHeight: atlas.LineHeight(),
		tabWidth:   atlas.Glyph(' ').Advance * 4,
		mat:        pixel.IM,
		col:        pixel.Alpha(1),
	}

	txt.glyph.SetLen(6)
	for i := range txt.glyph {
		txt.glyph[i].Color = pixel.Alpha(1)
		txt.glyph[i].Intensity = 1
	}

	txt.transD.Picture = txt.atlas.pic
	txt.transD.Triangles = &txt.trans

	txt.Clear()

	return txt
}

func (txt *Text) Atlas() *Atlas {
	return txt.atlas
}

func (txt *Text) SetMatrix(m pixel.Matrix) {
	if txt.mat != m {
		txt.mat = m
		txt.dirty = true
	}
}

func (txt *Text) SetColorMask(c color.Color) {
	rgba := pixel.ToRGBA(c)
	if txt.col != rgba {
		txt.col = rgba
		txt.dirty = true
	}
}

func (txt *Text) Bounds() pixel.Rect {
	return txt.bounds
}

func (txt *Text) BoundsOf(s string) pixel.Rect {
	dot := txt.Dot
	prevR := txt.prevR
	bounds := pixel.Rect{}

	for _, r := range s {
		var control bool
		dot, control = txt.controlRune(r, dot)
		if control {
			continue
		}

		var b pixel.Rect
		_, _, b, dot = txt.Atlas().DrawRune(prevR, r, dot)

		if bounds.W()*bounds.H() == 0 {
			bounds = b
		} else {
			bounds = bounds.Union(b)
		}

		prevR = r
	}

	return bounds
}

func (txt *Text) Color(c color.Color) {
	rgba := pixel.ToRGBA(c)
	for i := range txt.glyph {
		txt.glyph[i].Color = rgba
	}
}

func (txt *Text) LineHeight(height float64) {
	txt.lineHeight = height
}

func (txt *Text) TabWidth(width float64) {
	txt.tabWidth = width
}

func (txt *Text) Clear() {
	txt.prevR = -1
	txt.bounds = pixel.Rect{}
	txt.tris.SetLen(0)
	txt.dirty = true
}

func (txt *Text) Write(p []byte) (n int, err error) {
	txt.buf = append(txt.buf, p...)
	txt.drawBuf()
	return len(p), nil
}

func (txt *Text) WriteString(s string) (n int, err error) {
	txt.buf = append(txt.buf, s...)
	txt.drawBuf()
	return len(s), nil
}

func (txt *Text) WriteByte(c byte) error {
	txt.buf = append(txt.buf, c)
	txt.drawBuf()
	return nil
}

func (txt *Text) WriteRune(r rune) (n int, err error) {
	var b [4]byte
	n = utf8.EncodeRune(b[:], r)
	txt.buf = append(txt.buf, b[:n]...)
	txt.drawBuf()
	return n, nil
}

// controlRune checks if r is a control rune (newline, tab, ...). If it is, a new dot position and
// true is returned. If r is not a control rune, the original dot and false is returned.
func (txt *Text) controlRune(r rune, dot pixel.Vec) (newDot pixel.Vec, control bool) {
	switch r {
	case '\n':
		dot -= pixel.Y(txt.lineHeight)
		dot = dot.WithX(txt.Orig.X())
	case '\r':
		dot = dot.WithX(txt.Orig.X())
	case '\t':
		rem := math.Mod(dot.X()-txt.Orig.X(), txt.tabWidth)
		rem = math.Mod(rem, rem+txt.tabWidth)
		if rem == 0 {
			rem = txt.tabWidth
		}
		dot += pixel.X(rem)
	default:
		return dot, false
	}
	return dot, true
}

func (txt *Text) drawBuf() {
	for utf8.FullRune(txt.buf) {
		r, size := utf8.DecodeRune(txt.buf)
		txt.buf = txt.buf[size:]

		var control bool
		txt.Dot, control = txt.controlRune(r, txt.Dot)
		if control {
			continue
		}

		var rect, frame, bounds pixel.Rect
		rect, frame, bounds, txt.Dot = txt.Atlas().DrawRune(txt.prevR, r, txt.Dot)

		txt.prevR = r

		rv := [...]pixel.Vec{pixel.V(rect.Min.X(), rect.Min.Y()),
			pixel.V(rect.Max.X(), rect.Min.Y()),
			pixel.V(rect.Max.X(), rect.Max.Y()),
			pixel.V(rect.Min.X(), rect.Max.Y()),
		}

		fv := [...]pixel.Vec{pixel.V(frame.Min.X(), frame.Min.Y()),
			pixel.V(frame.Max.X(), frame.Min.Y()),
			pixel.V(frame.Max.X(), frame.Max.Y()),
			pixel.V(frame.Min.X(), frame.Max.Y()),
		}

		for i, j := range [...]int{0, 1, 2, 0, 2, 3} {
			txt.glyph[i].Position = rv[j]
			txt.glyph[i].Picture = fv[j]
		}

		txt.tris = append(txt.tris, txt.glyph...)
		txt.dirty = true

		if txt.bounds.W()*txt.bounds.H() == 0 {
			txt.bounds = bounds
		} else {
			txt.bounds = txt.bounds.Union(bounds)
		}
	}
}

func (txt *Text) Draw(t pixel.Target) {
	if txt.dirty {
		txt.trans.SetLen(txt.tris.Len())
		txt.trans.Update(&txt.tris)

		for i := range txt.trans {
			txt.trans[i].Position = txt.mat.Project(txt.trans[i].Position)
			txt.trans[i].Color = txt.trans[i].Color.Mul(txt.col)
		}

		txt.transD.Dirty()
		txt.dirty = false
	}

	txt.transD.Draw(t)
}

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

// Text allows for effiecient and convenient text drawing.
//
// To create a Text object, use the New constructor:
//   txt := text.New(face, text.ASCII)
//
// As suggested by the constructor, a Text object is always associated with one font face and a
// fixed set of runes. For example, the Text we create above can draw text using the font face
// contained in the `face` variable and is capable of drawing ASCII characters.
//
// Here we create a Text object which can draw ASCII and Katakana characters:
//   txt := text.New(face, text.ASCII, text.RangeTable(unicode.Katakana))
//
// Similarly to IMDraw, Text functions as a buffer. It implements io.Writer interface, so writing
// text to it is really simple:
//   fmt.Print(txt, "Hello, world!")
//
// Finally, if we want the written text to show up on some other Target, we can draw it:
//   txt.Draw(target)
//
// Text exports two important fields: Orig and Dot. Dot is the position where the next character
// will be written. Dot is automatically moved when writing to a Text object, but you can also
// manipulate it manually. Orig specifies the text origin, usually the top-left dot position. Dot is
// always aligned to Orig when writing newlines.
//
// To reset the Dot to the Orig, just assign it:
//   txt.Dot = txt.Orig
type Text struct {
	// Orig specifies the text origin, usually the top-left dot position. Dot is always aligned
	// to Orig when writing newlines.
	Orig pixel.Vec

	// Dot is the position where the next character will be written. Dot is automatically moved
	// when writing to a Text object, but you can also manipulate it manually
	Dot pixel.Vec

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

// New creates a new Text capable of drawing runes contained in the provided rune sets, plus
// unicode.ReplacementChar using the provided font.Face. New automatically generates an Atlas for
// the Text.
//
// Do not destroy or close the font.Face after creating a Text. Although Text caches most of the
// stuff (pre-drawn glyphs, etc.), it still uses the face for a few things.
//
// Here we create a Text capable of drawing ASCII characters using the Go Regular font.
//   ttf, err := truetype.Parse(goregular.TTF)
//   if err != nil {
//       panic(err)
//   }
//   face := truetype.NewFace(ttf, &truetype.Options{
//       Size: 14,
//   })
//   txt := text.New(face, text.ASCII)
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

// Atlas returns the underlying Text's Atlas containing all of the pre-drawn glyphs. The Atlas is
// also useful for getting values such as the recommended line height.
func (txt *Text) Atlas() *Atlas {
	return txt.atlas
}

// SetMatrix sets a Matrix by which the text will be transformed before drawing to another Target.
func (txt *Text) SetMatrix(m pixel.Matrix) {
	if txt.mat != m {
		txt.mat = m
		txt.dirty = true
	}
}

// SetColorMask sets a color by which the text will be masked before drawingto another Target.
func (txt *Text) SetColorMask(c color.Color) {
	rgba := pixel.ToRGBA(c)
	if txt.col != rgba {
		txt.col = rgba
		txt.dirty = true
	}
}

// Bounds returns the bounding box of the text currently written to the Text excluding whitespace.
//
// If the Text is empty, a zero rectangle is returned.
func (txt *Text) Bounds() pixel.Rect {
	return txt.bounds
}

// BoundsOf returns the bounding box of s if it was to be written to the Text right now.
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

// Color sets the text color. This does not affect any previously written text.
func (txt *Text) Color(c color.Color) {
	rgba := pixel.ToRGBA(c)
	for i := range txt.glyph {
		txt.glyph[i].Color = rgba
	}
}

// LineHeight sets the vertical distance between two lines of text. This does not affect any
// previously written text.
func (txt *Text) LineHeight(height float64) {
	txt.lineHeight = height
}

// TabWidth sets the horizontal tab width. Tab characters will align to the multiples of this width.
func (txt *Text) TabWidth(width float64) {
	txt.tabWidth = width
}

// Clear removes all written text from the Text.
func (txt *Text) Clear() {
	txt.prevR = -1
	txt.bounds = pixel.Rect{}
	txt.tris.SetLen(0)
	txt.dirty = true
}

// Write writes a slice of bytes to the Text. This method never fails, always returns len(p), nil.
func (txt *Text) Write(p []byte) (n int, err error) {
	txt.buf = append(txt.buf, p...)
	txt.drawBuf()
	return len(p), nil
}

// WriteString writes a string to the Text. This method never fails, always returns len(s), nil.
func (txt *Text) WriteString(s string) (n int, err error) {
	txt.buf = append(txt.buf, s...)
	txt.drawBuf()
	return len(s), nil
}

// WriteByte writes a byte to the Text. This method never fails, always returns nil.
//
// Writing a multi-byte rune byte-by-byte is perfectly supported.
func (txt *Text) WriteByte(c byte) error {
	txt.buf = append(txt.buf, c)
	txt.drawBuf()
	return nil
}

// WriteRune writes a rune to the Text. This method never fails, always returns utf8.RuneLen(r), nil.
func (txt *Text) WriteRune(r rune) (n int, err error) {
	var b [4]byte
	n = utf8.EncodeRune(b[:], r)
	txt.buf = append(txt.buf, b[:n]...)
	txt.drawBuf()
	return n, nil
}

// Draw draws all text written to the Text to the provided Target. The text is transformed by the
// Text's matrix and color mask.
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

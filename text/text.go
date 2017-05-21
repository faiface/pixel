package text

import (
	"image/color"
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/pixel"
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
//   txt := text.New(pixel.ZV, text.NewAtlas(face, text.ASCII))
//
// As suggested by the constructor, a Text object is always associated with one font face and a
// fixed set of runes. For example, the Text we created above can draw text using the font face
// contained in the face variable and is capable of drawing ASCII characters.
//
// Here we create a Text object which can draw ASCII and Katakana characters:
//   txt := text.New(0, text.NewAtlas(face, text.ASCII, text.RangeTable(unicode.Katakana)))
//
// Similarly to IMDraw, Text functions as a buffer. It implements io.Writer interface, so writing
// text to it is really simple:
//   fmt.Print(txt, "Hello, world!")
//
// Newlines, tabs and carriage returns are supported.
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

	// Color is the color of the text that is to be written. Defaults to white.
	Color color.Color

	// LineHeight is the vertical distance between two lines of text.
	//
	// Example:
	//   txt.LineHeight = 1.5 * txt.Atlas().LineHeight()
	LineHeight float64

	// TabWidth is the horizontal tab width. Tab characters will align to the multiples of this
	// width.
	//
	// Example:
	//   txt.TabWidth = 8 * txt.Atlas().Glyph(' ').Advance
	TabWidth float64

	atlas *Atlas

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

// New creates a new Text capable of drawing runes contained in the provided Atlas. Orig and Dot
// will be initially set to orig.
//
// Here we create a Text capable of drawing ASCII characters using the Go Regular font.
//   ttf, err := truetype.Parse(goregular.TTF)
//   if err != nil {
//       panic(err)
//   }
//   face := truetype.NewFace(ttf, &truetype.Options{
//       Size: 14,
//   })
//   txt := text.New(orig, text.NewAtlas(face, text.ASCII))
func New(orig pixel.Vec, atlas *Atlas) *Text {
	txt := &Text{
		Orig:       orig,
		Dot:        orig,
		Color:      pixel.Alpha(1),
		LineHeight: atlas.LineHeight(),
		TabWidth:   atlas.Glyph(' ').Advance * 4,
		atlas:      atlas,
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
// provided Matrix.
//
// This method is equivalent to calling DrawColorMask with nil color mask.
//
// If there's a lot of text written to the Text, changing a matrix or a color mask often might hurt
// performance. Consider using your Target's SetMatrix or SetColorMask methods if available.
func (txt *Text) Draw(t pixel.Target, matrix pixel.Matrix) {
	txt.DrawColorMask(t, matrix, nil)
}

// DrawColorMask draws all text written to the Text to the provided Target. The text is transformed
// by the provided Matrix and masked by the provided color mask.
//
// If there's a lot of text written to the Text, changing a matrix or a color mask often might hurt
// performance. Consider using your Target's SetMatrix or SetColorMask methods if available.
func (txt *Text) DrawColorMask(t pixel.Target, matrix pixel.Matrix, mask color.Color) {
	if matrix != txt.mat {
		txt.mat = matrix
		txt.dirty = true
	}
	if mask == nil {
		mask = pixel.Alpha(1)
	}
	rgba := pixel.ToRGBA(mask)
	if rgba != txt.col {
		txt.col = rgba
		txt.dirty = true
	}

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
		dot.X = txt.Orig.X
		dot.Y -= txt.LineHeight
	case '\r':
		dot.X = txt.Orig.X
	case '\t':
		rem := math.Mod(dot.X-txt.Orig.X, txt.TabWidth)
		rem = math.Mod(rem, rem+txt.TabWidth)
		if rem == 0 {
			rem = txt.TabWidth
		}
		dot.X += rem
	default:
		return dot, false
	}
	return dot, true
}

func (txt *Text) drawBuf() {
	if !utf8.FullRune(txt.buf) {
		return
	}

	rgba := pixel.ToRGBA(txt.Color)
	for i := range txt.glyph {
		txt.glyph[i].Color = rgba
	}

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

		rv := [...]pixel.Vec{
			{X: rect.Min.X, Y: rect.Min.Y},
			{X: rect.Max.X, Y: rect.Min.Y},
			{X: rect.Max.X, Y: rect.Max.Y},
			{X: rect.Min.X, Y: rect.Max.Y},
		}

		fv := [...]pixel.Vec{
			{X: frame.Min.X, Y: frame.Min.Y},
			{X: frame.Max.X, Y: frame.Min.Y},
			{X: frame.Max.X, Y: frame.Max.Y},
			{X: frame.Min.X, Y: frame.Max.Y},
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

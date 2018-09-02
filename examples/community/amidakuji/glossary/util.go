package glossary

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"math/rand"
	"os"

	// Relevant packages of target format for a decoder must be initialized to register.
	_ "image/gif"
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

var atlasASCII *text.Atlas

func init() {
	atlasASCII = NewAtlas("", 18, nil)
}

// AtlasASCII returns an atlas which allows you to draw only ASCII characters.
// Atlas is a set of generated textures for glyphs in a specific font.
func AtlasASCII() *text.Atlas {
	return atlasASCII
}

// NewAtlas newly loads and prepares a set of images of characters or symbols to be drawn.
// Arg runeSet would be set to nil if non-ASCII characters are not in use.
func NewAtlas(nameAssetTTF string, size float64, runeSet []rune) *text.Atlas {
	if nameAssetTTF == "" {
		nameAssetTTF = "NanumBarunGothic.ttf"
	}

	var face font.Face
	asset, err := Asset(nameAssetTTF)
	if err == nil {
		face, err = LoadTrueTypeFont(asset, size)
	}
	if err != nil {
		face = basicfont.Face7x13
	}
	return text.NewAtlas(face, text.ASCII, runeSet)
}

// NewSprite converts an asset (resource) into a sprite. Returns nil if there is an error.
// AssetNames() or AssetDir() might be helpful when utilizing this function.
func NewSprite(nameAsset string) *pixel.Sprite {
	asset, err := Asset(nameAsset)
	if err != nil {
		// log.Println("1", err) //
		return nil
	}
	pic, err := LoadPicture(asset)
	if err != nil {
		// log.Println("2", err) //
		return nil
	}
	// log.Println("3", "success yay") //
	return pixel.NewSprite(pic, pic.Bounds())
}

// LoadTrueTypeFontFromFile creates and returns a font face.
func LoadTrueTypeFontFromFile(path string, size float64) (font.Face, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	face, err := LoadTrueTypeFont(bytes, size)
	if err != nil {
		return nil, err
	}

	return face, nil
}

// LoadPictureFromFile decodes an image that has been encoded in a registered format. (png, jpg, etc.)
// Format registration is typically done by an init function in the codec-specific package. (with underscore import)
func LoadPictureFromFile(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// LoadTrueTypeFont creates and returns a font face.
func LoadTrueTypeFont(bytes []byte, size float64) (font.Face, error) {
	font, err := truetype.Parse(bytes)
	if err != nil {
		return nil, err
	}
	return truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	}), nil
}

// LoadPicture decodes an image that has been encoded in a registered format. (png, jpg, etc.)
// Format registration is typically done by an init function in the codec-specific package. (with underscore import)
func LoadPicture(_bytes []byte) (pixel.Picture, error) {
	img, _, err := image.Decode(bytes.NewReader(_bytes))
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// RandomNiceColor from Platformer.
// Is not completely random without rand.Seed().
func RandomNiceColor() pixel.RGBA {
again:
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	len := math.Sqrt(r*r + g*g + b*b)
	if len == 0 {
		goto again
	}
	return pixel.RGB(r/len, g/len, b/len)
}

// VerticesOfRect returns 4 vertices of a rectangle in a form of a slice of vectors.
func VerticesOfRect(r pixel.Rect) []pixel.Vec {
	return []pixel.Vec{
		r.Min,
		pixel.V(r.Max.X, r.Min.Y),
		r.Max,
		pixel.V(r.Min.X, r.Max.Y),
	}
}

// ItfsToStrs converts []interface{} to []string.
func ItfsToStrs(itfs []interface{}) (strs []string) {
	strs = make([]string, len(itfs))
	for i, v := range itfs {
		strs[i] = fmt.Sprint(v)
	}
	return strs
}

// Direction returns a direction as a normalized vector. This vector always has a length of 1.
func Direction(from, to pixel.Vec) (dirVecNormalized pixel.Vec) {
	vec := to.Sub(from)
	if vec.X == 0 && vec.Y == 0 {
		return vec
	}
	return vec.Unit()
}

// -------------------------------------------------------------------------
// Anchors

// AnchorY - Top, Middle, Bottom
type AnchorY int

// enum AnchorY
const (
	Top AnchorY = 1 + iota
	Middle
	Bottom
)

// AnchorX - Left, Center, Right
type AnchorX int

// enum AnchorX
const (
	Left AnchorX = 1 + iota
	Center
	Right
)

// AnchorTxt positions a text.Text label with an anchor alignment.
func AnchorTxt(txt *text.Text, pos pixel.Vec, anchorX AnchorX, anchorY AnchorY, desc string) {
	txt.Orig = pos
	txt.Dot = pos
	switch anchorX {
	case Left:
		txt.Dot.X -= 0
	case Center:
		txt.Dot.X -= (txt.BoundsOf(desc).W() / 2)
	case Right:
		txt.Dot.X -= txt.BoundsOf(desc).W()
	}
	switch anchorY {
	case Top:
		txt.Dot.Y -= txt.BoundsOf(desc).H()
	case Middle:
		txt.Dot.Y -= (txt.BoundsOf(desc).H() / 2)
	case Bottom:
		txt.Dot.Y -= 0
	}
	txt.Dot.X += 0
	txt.Dot.Y += 0
}

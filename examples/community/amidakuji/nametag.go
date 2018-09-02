package main

import (
	"image/color"
	"sync"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

// Nametags is a list(slice) of nametags.
type Nametags []Nametag

// Update nametags.
func (updaters Nametags) Update() {
	for i := range updaters {
		updaters[i].Update()
	}
}

// Draw nametags.
func (updaters Nametags) Draw(t pixel.Target) {
	for i := range updaters {
		updaters[i].Draw(t)
	}
}

// Nametag for each nametag.
type Nametag struct {
	txt   *text.Text     // shared variable
	atlas *text.Atlas    // borrowed atlas for txt
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
	//
	desc     string
	pos      pixel.Vec
	anchorX  gg.AnchorX
	anchorY  gg.AnchorY
	colorBg  color.Color
	colorTxt color.Color
}

// NewNametag is a constructor.
func NewNametag(
	_atlas *text.Atlas,
	_desc string, _pos pixel.Vec,
	_anchorY gg.AnchorY, // This is because the order is usually Y then X in spoken language.
	_anchorX gg.AnchorX,
	_colorBg, _colorTxt color.Color) *Nametag {
	atlas := _atlas
	if atlas == nil {
		atlas = gg.AtlasASCII()
	}
	return &Nametag{
		atlas:    atlas,
		desc:     _desc,
		pos:      _pos,
		anchorX:  _anchorX,
		anchorY:  _anchorY,
		colorBg:  _colorBg,
		colorTxt: _colorTxt,
	}
}

// NewNametagSimple is a constructor.
func NewNametagSimple(
	_atlas *text.Atlas,
	_desc string, _pos pixel.Vec,
	_anchorY gg.AnchorY,
	_anchorX gg.AnchorX,
) *Nametag {
	return NewNametag(_atlas, _desc, _pos, _anchorY, _anchorX, colornames.Wheat, colornames.Black)
}

// String of a nametag.
// A getter and a callback which allows a nametag to be passed to a function as a string.
// A non-ptr Nametag as a read only argument passes lock by value within itself but that seems totally fine.
func (n Nametag) String() string {
	return n.desc
}

// Draw a nametag.
func (n *Nametag) Draw(t pixel.Target) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.imd == nil && n.txt == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	n.imd.Draw(t)
	n.txt.Draw(t, pixel.IM)
}

// Update a nametag.
func (n *Nametag) Update() {
	// lock before txt & imdraw update
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// text label (a state machine)
	if n.txt == nil { // lazy creation
		n.txt = text.New(pixel.ZV, n.atlas)
	}
	txt := n.txt
	txt.Clear()

	gg.AnchorTxt(txt, n.pos, n.anchorX, n.anchorY, n.desc)
	txt.Color = n.colorTxt
	txt.WriteString(n.desc)

	// imdraw (a state machine)
	if n.imd == nil { // lazy creation
		n.imd = imdraw.New(nil)
	}
	imd := n.imd
	imd.Clear()

	imd.Color = n.colorBg
	imd.Push(gg.VerticesOfRect(txt.Bounds())...)
	imd.Polygon(0)
}

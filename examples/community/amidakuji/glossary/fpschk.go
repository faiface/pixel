package glossary

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

// FPSWatch measures the real-time frame rates and displays it on a target canvas.
type FPSWatch struct {
	txt   *text.Text     // shared variable
	atlas *text.Atlas    // borrowed atlas for txt
	imd   *imdraw.IMDraw // shared variable
	mutex sync.Mutex     // synchronize
	//
	fps    int              // The FPS evaluated every second.
	frames int              // Frames count before the FPS update.
	seccer <-chan time.Time // Ticks time every second.
	//
	desc     string
	pos      pixel.Vec
	anchorX  AnchorX
	anchorY  AnchorY
	colorBg  color.Color
	colorTxt color.Color
}

// NewFPSWatch is a constructor.
func NewFPSWatch(
	additionalCaption string, _pos pixel.Vec,
	_anchorY AnchorY, _anchorX AnchorX, // This is because the order is usually Y then X in spoken language.
	_colorBg, _colorTxt color.Color,
) (watch *FPSWatch) {
	return &FPSWatch{
		atlas:    AtlasASCII(),
		fps:      0,
		frames:   0,
		seccer:   nil,
		desc:     additionalCaption,
		pos:      _pos,
		anchorX:  _anchorX,
		anchorY:  _anchorY,
		colorBg:  _colorBg,
		colorTxt: _colorTxt,
	}
}

// NewFPSWatchSimple is a constructor.
func NewFPSWatchSimple(_pos pixel.Vec, _anchorY AnchorY, _anchorX AnchorX) *FPSWatch {
	return NewFPSWatch("", _pos, _anchorY, _anchorX, colornames.Black, colornames.White)
}

// Start ticking every second.
func (watch *FPSWatch) Start() {
	watch.seccer = time.Tick(time.Second)
}

// Poll () should be called only once and in every single frame. (Obligatory)
// This is an extended behavior of Update() like funcs.
func (watch *FPSWatch) Poll() {
	watch.frames++
	select {
	case <-watch.seccer:
		watch.fps = watch.frames
		watch.frames = 0
		go watch._Update()
	default:
	}
}

// SetPos to a position in screen coords.
func (watch *FPSWatch) SetPos(pos pixel.Vec, anchorY AnchorY, anchorX AnchorX) {
	watch.pos = pos
	watch.anchorX = anchorX
	watch.anchorY = anchorY
}

// GetFPS returns the most recent FPS recorded.
// A non-ptr FPSWatch as a read only argument passes lock by value within itself but that seems totally fine.
func (watch FPSWatch) GetFPS() int {
	return watch.fps
}

// Draw FPSWatch.
func (watch *FPSWatch) Draw(t pixel.Target) {
	// lock before accessing txt & imdraw
	watch.mutex.Lock()
	defer watch.mutex.Unlock()

	if watch.imd == nil && watch.txt == nil { // isInvisible set to true.
		return // An empty image is drawn.
	}

	watch.imd.Draw(t)
	watch.txt.Draw(t, pixel.IM)
}

// unexported
func (watch *FPSWatch) _Update() {
	// lock before txt & imdraw update
	watch.mutex.Lock()
	defer watch.mutex.Unlock()

	// text label (a state machine)
	if watch.txt == nil { // lazy creation
		watch.txt = text.New(pixel.ZV, watch.atlas)
	}
	txt := watch.txt
	txt.Clear()

	str := fmt.Sprint("FPS: ", watch.fps, " ", watch.desc)
	AnchorTxt(txt, watch.pos, watch.anchorX, watch.anchorY, str)
	txt.Color = watch.colorTxt
	txt.Dot.X -= 1.0
	txt.Dot.Y += 5.0
	txt.WriteString(str)

	// imdraw (a state machine)
	if watch.imd == nil { // lazy creation
		watch.imd = imdraw.New(nil)
	}
	imd := watch.imd
	imd.Clear()

	imd.Color = watch.colorBg
	imd.Push(VerticesOfRect(txt.Bounds())...)
	imd.Polygon(0)
}

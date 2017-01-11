package pixel

import (
	"image/color"

	"runtime"

	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// WindowConfig is a structure for specifying all possible properties of a window. Properties are
// chosen in such a way, that you usually only need to set a few of them - defaults (zeros) should
// usually be sensible.
//
// Note that you always need to set the width and the height of a window.
type WindowConfig struct {
	// Title at the top of a window.
	Title string

	// Width of a window in pixels.
	Width float64

	// Height of a window in pixels.
	Height float64

	// If set to nil, a window will be windowed. Otherwise it will be fullscreen on the
	// specified monitor.
	Fullscreen *Monitor

	// Whether a window is resizable.
	Resizable bool

	// If set to true, the window will be initially invisible.
	Hidden bool

	// Undecorated window ommits the borders and decorations (close button, etc.).
	Undecorated bool

	// If set to true, a window will not get focused upon showing up.
	Unfocused bool

	// Whether a window is maximized.
	Maximized bool

	// VSync (vertical synchronization) synchronizes window's framerate with the framerate
	// of the monitor.
	VSync bool

	// Number of samples for multi-sample anti-aliasing (edge-smoothing).  Usual values
	// are 0, 2, 4, 8 (powers of 2 and not much more than this).
	MSAASamples int
}

// Window is a window handler. Use this type to manipulate a window (input, drawing, ...).
type Window struct {
	enabled bool
	window  *glfw.Window
	config  WindowConfig
	shader  *pixelgl.Shader

	// Target stuff, Picture, transformation matrix and color
	pic *Picture
	mat mgl32.Mat3
	col mgl32.Vec4

	// need to save these to correctly restore a fullscreen window
	restore struct {
		xpos, ypos, width, height int
	}

	prevInp, tempInp, currInp struct {
		buttons [KeyLast + 1]bool
		scroll  Vec
	}
}

var currentWindow *Window

// NewWindow creates a new window with it's properties specified in the provided config.
//
// If window creation fails, an error is returned.
func NewWindow(config WindowConfig) (*Window, error) {
	bool2int := map[bool]int{
		true:  glfw.True,
		false: glfw.False,
	}

	w := &Window{config: config}

	err := pixelgl.DoErr(func() error {
		err := glfw.Init()
		if err != nil {
			return err
		}

		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

		glfw.WindowHint(glfw.Resizable, bool2int[config.Resizable])
		glfw.WindowHint(glfw.Visible, bool2int[!config.Hidden])
		glfw.WindowHint(glfw.Decorated, bool2int[!config.Undecorated])
		glfw.WindowHint(glfw.Focused, bool2int[!config.Unfocused])
		glfw.WindowHint(glfw.Maximized, bool2int[config.Maximized])
		glfw.WindowHint(glfw.Samples, config.MSAASamples)

		var share *glfw.Window
		if currentWindow != nil {
			share = currentWindow.window
		}
		w.window, err = glfw.CreateWindow(int(config.Width), int(config.Height), config.Title, nil, share)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating window failed")
	}

	pixelgl.Do(func() {
		w.begin()
		w.end()

		w.shader, err = pixelgl.NewShader(
			defaultVertexFormat,
			defaultUniformFormat,
			defaultVertexShader,
			defaultFragmentShader,
		)
		if err != nil {
			panic(errors.Wrap(err, "NewWindow: failed to create shader"))
		}
	})
	if err != nil {
		w.Destroy()
		return nil, errors.Wrap(err, "creating window failed")
	}

	w.initInput()
	w.SetFullscreen(config.Fullscreen)

	w.SetPicture(nil)
	w.SetTransform()
	w.SetMaskColor(NRGBA{1, 1, 1, 1})

	runtime.SetFinalizer(w, (*Window).Destroy)

	return w, nil
}

// Destroy destroys a window. The window can't be used any further.
func (w *Window) Destroy() {
	pixelgl.Do(func() {
		w.window.Destroy()
	})
}

// Clear clears the window with a color.
func (w *Window) Clear(c color.Color) {
	pixelgl.DoNoBlock(func() {
		w.begin()
		defer w.end()

		c := NRGBAModel.Convert(c).(NRGBA)
		gl.ClearColor(float32(c.R), float32(c.G), float32(c.B), float32(c.A))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}

// Update swaps buffers and polls events.
func (w *Window) Update() {
	pixelgl.Do(func() {
		w.begin()
		defer w.end()

		if w.config.VSync {
			glfw.SwapInterval(1)
		}
		w.window.SwapBuffers()
	})

	w.updateInput()

	pixelgl.Do(func() {
		w.begin()
		defer w.end()

		w, h := w.window.GetSize()
		gl.Viewport(0, 0, int32(w), int32(h))
	})
}

// SetClosed sets the closed flag of a window.
//
// This is usefull when overriding the user's attempt to close a window, or just to close a
// window from within a program.
func (w *Window) SetClosed(closed bool) {
	pixelgl.Do(func() {
		w.window.SetShouldClose(closed)
	})
}

// Closed returns the closed flag of a window, which reports whether the window should be closed.
//
// The closed flag is automatically set when a user attempts to close a window.
func (w *Window) Closed() bool {
	return pixelgl.DoVal(func() interface{} {
		return w.window.ShouldClose()
	}).(bool)
}

// SetTitle changes the title of a window.
func (w *Window) SetTitle(title string) {
	pixelgl.Do(func() {
		w.window.SetTitle(title)
	})
}

// SetSize resizes a window to the specified size in pixels.  In case of a fullscreen window,
// it changes the resolution of that window.
func (w *Window) SetSize(width, height float64) {
	pixelgl.Do(func() {
		w.window.SetSize(int(width), int(height))
	})
}

// Size returns the size of the client area of a window (the part you can draw on).
func (w *Window) Size() (width, height float64) {
	pixelgl.Do(func() {
		wi, hi := w.window.GetSize()
		width = float64(wi)
		height = float64(hi)
	})
	return width, height
}

// Show makes a window visible if it was hidden.
func (w *Window) Show() {
	pixelgl.Do(func() {
		w.window.Show()
	})
}

// Hide hides a window if it was visible.
func (w *Window) Hide() {
	pixelgl.Do(func() {
		w.window.Hide()
	})
}

// SetFullscreen sets a window fullscreen on a given monitor. If the monitor is nil, the window
// will be resored to windowed instead.
//
// Note, that there is nothing about the resolution of the fullscreen window. The window is
// automatically set to the monitor's resolution. If you want a different resolution, you need
// to set it manually with SetSize method.
func (w *Window) SetFullscreen(monitor *Monitor) {
	if w.Monitor() != monitor {
		if monitor == nil {
			pixelgl.Do(func() {
				w.window.SetMonitor(
					nil,
					w.restore.xpos,
					w.restore.ypos,
					w.restore.width,
					w.restore.height,
					0,
				)
			})
		} else {
			pixelgl.Do(func() {
				w.restore.xpos, w.restore.ypos = w.window.GetPos()
				w.restore.width, w.restore.height = w.window.GetSize()

				width, height := monitor.Size()
				refreshRate := monitor.RefreshRate()
				w.window.SetMonitor(
					monitor.monitor,
					0,
					0,
					int(width),
					int(height),
					int(refreshRate),
				)
			})
		}
	}
}

// IsFullscreen returns true if the window is in the fullscreen mode.
func (w *Window) IsFullscreen() bool {
	return w.Monitor() != nil
}

// Monitor returns a monitor a fullscreen window is on. If the window is not fullscreen, this
// function returns nil.
func (w *Window) Monitor() *Monitor {
	monitor := pixelgl.DoVal(func() interface{} {
		return w.window.GetMonitor()
	}).(*glfw.Monitor)
	if monitor == nil {
		return nil
	}
	return &Monitor{
		monitor: monitor,
	}
}

// Focus brings a window to the front and sets input focus.
func (w *Window) Focus() {
	pixelgl.Do(func() {
		w.window.Focus()
	})
}

// Focused returns true if a window has input focus.
func (w *Window) Focused() bool {
	return pixelgl.DoVal(func() interface{} {
		return w.window.GetAttrib(glfw.Focused) == glfw.True
	}).(bool)
}

// Maximize puts a windowed window to a maximized state.
func (w *Window) Maximize() {
	pixelgl.Do(func() {
		w.window.Maximize()
	})
}

// Restore restores a windowed window from a maximized state.
func (w *Window) Restore() {
	pixelgl.Do(func() {
		w.window.Restore()
	})
}

// Note: must be called inside the main thread.
func (w *Window) begin() {
	if currentWindow != w {
		w.window.MakeContextCurrent()
		pixelgl.Init()
		currentWindow = w
	}
	if w.shader != nil {
		w.shader.Begin()
	}
}

// Note: must be called inside the main thread.
func (w *Window) end() {
	if w.shader != nil {
		w.shader.End()
	}
}

type windowTriangles struct {
	w    *Window
	vs   *pixelgl.VertexSlice
	data []pixelgl.VertexData
}

func (wt *windowTriangles) Len() int {
	return wt.vs.Len()
}

func (wt *windowTriangles) Draw() {
	pic := wt.w.pic // avoid
	mat := wt.w.mat // race
	col := wt.w.col // condition

	pixelgl.DoNoBlock(func() {
		wt.w.begin()

		wt.w.shader.SetUniformAttr(transformMat3, mat)
		wt.w.shader.SetUniformAttr(maskColorVec4, col)

		if pic != nil {
			pic.Texture().Begin()
		}
		wt.vs.Begin()
		wt.vs.Draw()
		wt.vs.End()
		if pic != nil {
			pic.Texture().End()
		}

		wt.w.end()
	})
}

func (wt *windowTriangles) updateData(offset int, t Triangles) {
	if t, ok := t.(TrianglesPosition); ok {
		for i := offset; i < offset+t.Len(); i++ {
			pos := t.Position(i)
			wt.data[i][positionVec2] = mgl32.Vec2{
				float32(pos.X()),
				float32(pos.Y()),
			}
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := offset; i < offset+t.Len(); i++ {
			col := NRGBAModel.Convert(t.Color(i)).(NRGBA)
			wt.data[i][colorVec4] = mgl32.Vec4{
				float32(col.R),
				float32(col.G),
				float32(col.B),
				float32(col.A),
			}
		}
	}
	if t, ok := t.(TrianglesTexture); ok {
		for i := offset; i < offset+t.Len(); i++ {
			tex := t.Texture(i)
			wt.data[i][textureVec2] = mgl32.Vec2{
				float32(tex.X()),
				float32(tex.Y()),
			}
		}
	}
}

func (wt *windowTriangles) resize(len int) {
	if len > wt.Len() {
		newData := make([]pixelgl.VertexData, len-wt.Len())
		// default values
		for i := range newData {
			newData[i] = make(pixelgl.VertexData)
			newData[i][colorVec4] = mgl32.Vec4{1, 1, 1, 1}
			newData[i][textureVec2] = mgl32.Vec2{-1, -1}
		}
		wt.data = append(wt.data, newData...)
	}
	if len < wt.Len() {
		wt.data = wt.data[:len]
	}
}

func (wt *windowTriangles) submitData() {
	data := wt.data // avoid race condition
	pixelgl.DoNoBlock(func() {
		wt.vs.Begin()
		if len(wt.data) > wt.vs.Len() {
			wt.vs.Append(make([]pixelgl.VertexData, len(data)-wt.vs.Len())...)
		}
		if len(wt.data) < wt.vs.Len() {
			wt.vs = wt.vs.Slice(0, len(wt.data))
		}
		wt.vs.SetVertexData(wt.data)
		wt.vs.End()
	})
}

func (wt *windowTriangles) Update(t Triangles) {
	wt.resize(t.Len())
	wt.updateData(0, t)
	wt.submitData()
}

func (wt *windowTriangles) Append(t Triangles) {
	wt.resize(wt.Len() + t.Len())
	wt.updateData(wt.Len()-t.Len(), t)
	wt.submitData()
}

func (wt *windowTriangles) Position(i int) Vec {
	v := wt.data[i][positionVec2].(mgl32.Vec2)
	return V(float64(v.X()), float64(v.Y()))
}

func (wt *windowTriangles) Color(i int) color.Color {
	c := wt.data[i][colorVec4].(mgl32.Vec4)
	return NRGBA{
		R: float64(c.X()),
		G: float64(c.Y()),
		B: float64(c.Z()),
		A: float64(c.W()),
	}
}

func (wt *windowTriangles) Texture(i int) Vec {
	t := wt.data[i][textureVec2].(mgl32.Vec2)
	return V(float64(t.X()), float64(t.Y()))
}

// MakeTriangles generates a specialized copy of the supplied triangles that will draw onto this
// Window.
//
// Window supports TrianglesPosition, TrianglesColor and TrianglesTexture.
func (w *Window) MakeTriangles(t Triangles) Triangles {
	wt := &windowTriangles{
		w:  w,
		vs: pixelgl.MakeVertexSlice(w.shader, 0, 0),
	}
	wt.Update(t)
	return wt
}

// SetPicture sets a Picture that will be used in subsequent drawings onto the window.
func (w *Window) SetPicture(p *Picture) {
	w.pic = p
}

// SetTransform sets a global transformation matrix for the Window.
//
// Transforms are applied right-to-left.
func (w *Window) SetTransform(t ...Transform) {
	w.mat = mgl32.Ident3()
	for i := range t {
		w.mat = w.mat.Mul3(t[i].Mat())
	}
}

// SetMaskColor sets a global mask color for the Window.
func (w *Window) SetMaskColor(c color.Color) {
	if c == nil {
		c = NRGBA{1, 1, 1, 1}
	}
	nrgba := NRGBAModel.Convert(c).(NRGBA)
	r := float32(nrgba.R)
	g := float32(nrgba.G)
	b := float32(nrgba.B)
	a := float32(nrgba.A)
	w.col = mgl32.Vec4{r, g, b, a}
}

var defaultVertexFormat = pixelgl.AttrFormat{
	"position": pixelgl.Vec2,
	"color":    pixelgl.Vec4,
	"texture":  pixelgl.Vec2,
}

var defaultUniformFormat = pixelgl.AttrFormat{
	"maskColor": pixelgl.Vec4,
	"transform": pixelgl.Mat3,
}

var defaultVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texture;

out vec4 Color;
out vec2 Texture;

uniform mat3 transform;

void main() {
	gl_Position = vec4((transform * vec3(position.x, position.y, 1.0)).xy, 0.0, 1.0);
	Color = color;
	Texture = texture;
}
`

var defaultFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 Texture;

out vec4 color;

uniform vec4 maskColor;
uniform sampler2D tex;

void main() {
	if (Texture == vec2(-1, -1)) {
		color = maskColor * Color;
	} else {
		color = maskColor * Color * texture(tex, vec2(Texture.x, 1 - Texture.y));
	}
}
`

var (
	positionVec2 = pixelgl.Attr{
		Name: "position",
		Type: pixelgl.Vec2,
	}
	colorVec4 = pixelgl.Attr{
		Name: "color",
		Type: pixelgl.Vec4,
	}
	textureVec2 = pixelgl.Attr{
		Name: "texture",
		Type: pixelgl.Vec2,
	}
	maskColorVec4 = pixelgl.Attr{
		Name: "maskColor",
		Type: pixelgl.Vec4,
	}
	transformMat3 = pixelgl.Attr{
		Name: "transform",
		Type: pixelgl.Mat3,
	}
)

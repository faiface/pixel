package pixel

import (
	"image/color"

	"runtime"

	"github.com/faiface/mainthread"
	"github.com/faiface/pixel/pixelgl"
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

	// cache
	width, height float64

	// Target stuff, Picture, transformation matrix and color
	pic *Picture
	mat mgl32.Mat3
	col mgl32.Vec4
	bnd mgl32.Vec4

	// need to save these to correctly restore a fullscreen window
	restore struct {
		xpos, ypos, width, height int
	}

	prevInp, tempInp, currInp struct {
		buttons [KeyLast + 1]bool
		scroll  Vec
	}

	//DEBUG
	Frame *pixelgl.Frame
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

	err := mainthread.CallErr(func() error {
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

	mainthread.Call(func() {
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
	mainthread.Call(func() {
		w.window.Destroy()
	})
}

// Clear clears the window with a color.
func (w *Window) Clear(c color.Color) {
	mainthread.CallNonBlock(func() {
		w.begin()
		c := NRGBAModel.Convert(c).(NRGBA)
		pixelgl.Clear(float32(c.R), float32(c.G), float32(c.B), float32(c.A))
		w.end()
	})
}

// Update swaps buffers and polls events.
func (w *Window) Update() {
	mainthread.Call(func() {
		w.begin()
		if w.config.VSync {
			glfw.SwapInterval(1)
		}
		w.window.SwapBuffers()
		w.end()
	})

	w.updateInput()

	w.width, w.height = w.Size()
}

// SetClosed sets the closed flag of a window.
//
// This is usefull when overriding the user's attempt to close a window, or just to close a
// window from within a program.
func (w *Window) SetClosed(closed bool) {
	mainthread.Call(func() {
		w.window.SetShouldClose(closed)
	})
}

// Closed returns the closed flag of a window, which reports whether the window should be closed.
//
// The closed flag is automatically set when a user attempts to close a window.
func (w *Window) Closed() bool {
	return mainthread.CallVal(func() interface{} {
		return w.window.ShouldClose()
	}).(bool)
}

// SetTitle changes the title of a window.
func (w *Window) SetTitle(title string) {
	mainthread.Call(func() {
		w.window.SetTitle(title)
	})
}

// SetSize resizes a window to the specified size in pixels.  In case of a fullscreen window,
// it changes the resolution of that window.
func (w *Window) SetSize(width, height float64) {
	mainthread.Call(func() {
		w.window.SetSize(int(width), int(height))
	})
}

// Size returns the size of the client area of a window (the part you can draw on).
func (w *Window) Size() (width, height float64) {
	mainthread.Call(func() {
		wi, hi := w.window.GetSize()
		width = float64(wi)
		height = float64(hi)
	})
	return width, height
}

// Show makes a window visible if it was hidden.
func (w *Window) Show() {
	mainthread.Call(func() {
		w.window.Show()
	})
}

// Hide hides a window if it was visible.
func (w *Window) Hide() {
	mainthread.Call(func() {
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
			mainthread.Call(func() {
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
			mainthread.Call(func() {
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
	monitor := mainthread.CallVal(func() interface{} {
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
	mainthread.Call(func() {
		w.window.Focus()
	})
}

// Focused returns true if a window has input focus.
func (w *Window) Focused() bool {
	return mainthread.CallVal(func() interface{} {
		return w.window.GetAttrib(glfw.Focused) == glfw.True
	}).(bool)
}

// Maximize puts a windowed window to a maximized state.
func (w *Window) Maximize() {
	mainthread.Call(func() {
		w.window.Maximize()
	})
}

// Restore restores a windowed window from a maximized state.
func (w *Window) Restore() {
	mainthread.Call(func() {
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
	if w.Frame != nil {
		w.Frame.Begin()
		pixelgl.Viewport(0, 0, int32(w.Frame.Width()), int32(w.Frame.Height()))
	} else {
		pixelgl.Viewport(0, 0, int32(w.width), int32(w.height))
	}
}

// Note: must be called inside the main thread.
func (w *Window) end() {
	if w.Frame != nil {
		w.Frame.End()
	}
	if w.shader != nil {
		w.shader.End()
	}
}

type trianglesPositionColorTexture interface {
	Triangles
	Position(i int) Vec
	Color(i int) NRGBA
	Texture(i int) Vec
}

type windowTriangles struct {
	w *Window
	trianglesPositionColorTexture
}

func (wt *windowTriangles) Draw() {
	// avoid possible race condition
	pic := wt.w.pic
	mat := wt.w.mat
	col := wt.w.col
	bnd := wt.w.bnd

	mainthread.CallNonBlock(func() {
		wt.w.begin()

		wt.w.shader.SetUniformAttr(transformMat3, mat)
		wt.w.shader.SetUniformAttr(maskColorVec4, col)
		wt.w.shader.SetUniformAttr(boundsVec4, bnd)

		if pic != nil {
			pic.Texture().Begin()
		}
		wt.trianglesPositionColorTexture.Draw()
		if pic != nil {
			pic.Texture().End()
		}

		wt.w.end()
	})
}

// MakeTriangles generates a specialized copy of the supplied triangles that will draw onto this
// Window.
//
// Window supports TrianglesPosition, TrianglesColor and TrianglesTexture.
func (w *Window) MakeTriangles(t Triangles) Triangles {
	tpcs := NewGLTriangles(w.shader, t).(trianglesPositionColorTexture)
	wt := &windowTriangles{
		w: w,
		trianglesPositionColorTexture: tpcs,
	}
	return wt
}

// SetPicture sets a Picture that will be used in subsequent drawings onto the window.
func (w *Window) SetPicture(p *Picture) {
	if p != nil {
		min := pictureBounds(p, V(0, 0))
		max := pictureBounds(p, V(1, 1))
		w.bnd = mgl32.Vec4{
			float32(min.X()), float32(min.Y()),
			float32(max.X()), float32(max.Y()),
		}
	}
	w.pic = p
}

// SetTransform sets a global transformation matrix for the Window.
//
// Transforms are applied right-to-left.
func (w *Window) SetTransform(t ...Transform) {
	w.mat = transformToMat(t...)
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

const (
	positionVec2 int = iota
	colorVec4
	textureVec2
)

var defaultVertexFormat = pixelgl.AttrFormat{
	positionVec2: {Name: "position", Type: pixelgl.Vec2},
	colorVec4:    {Name: "color", Type: pixelgl.Vec4},
	textureVec2:  {Name: "texture", Type: pixelgl.Vec2},
}

const (
	maskColorVec4 int = iota
	transformMat3
	boundsVec4
)

var defaultUniformFormat = pixelgl.AttrFormat{
	{Name: "maskColor", Type: pixelgl.Vec4},
	{Name: "transform", Type: pixelgl.Mat3},
	{Name: "bounds", Type: pixelgl.Vec4},
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
uniform vec4 bounds;
uniform sampler2D tex;

void main() {
	vec2 boundsMin = bounds.xy;
	vec2 boundsMax = bounds.zw;

	if (Texture == vec2(-1, -1)) {
		color = maskColor * Color;
	} else {
		float tx = boundsMin.x * (1 - Texture.x) + boundsMax.x * Texture.x;
		float ty = boundsMin.y * (1 - Texture.y) + boundsMax.y * Texture.y;
		color = maskColor * Color * texture(tex, vec2(tx, 1 - ty));
	}
}
`

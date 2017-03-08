package pixelgl

import (
	"image/color"
	"math"
	"runtime"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// WindowConfig is a structure for specifying all possible properties of a window. Properties are
// chosen in such a way, that you usually only need to set a few of them - defaults (zeros) should
// usually be sensible.
//
// Note that you always need to set the Bounds of the Window.
type WindowConfig struct {
	// Title at the top of the Window
	Title string

	// Bounds specify the bounds of the Window in pixels.
	Bounds pixel.Rect

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

	// VSync (vertical synchronization) synchronizes window's framerate with the framerate of
	// the monitor.
	VSync bool
}

// Window is a window handler. Use this type to manipulate a window (input, drawing, ...).
type Window struct {
	window *glfw.Window

	bounds pixel.Rect
	canvas *Canvas
	vsync  bool

	// need to save these to correctly restore a fullscreen window
	restore struct {
		xpos, ypos, width, height int
	}

	prevInp, tempInp, currInp struct {
		mouse   pixel.Vec
		buttons [KeyLast + 1]bool
		scroll  pixel.Vec
	}
}

var currentWindow *Window

// NewWindow creates a new Window with it's properties specified in the provided config.
//
// If Window creation fails, an error is returned (e.g. due to unavailable graphics device).
func NewWindow(cfg WindowConfig) (*Window, error) {
	bool2int := map[bool]int{
		true:  glfw.True,
		false: glfw.False,
	}

	w := &Window{
		bounds: cfg.Bounds,
	}

	err := mainthread.CallErr(func() error {
		var err error

		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

		glfw.WindowHint(glfw.Resizable, bool2int[cfg.Resizable])
		glfw.WindowHint(glfw.Visible, bool2int[!cfg.Hidden])
		glfw.WindowHint(glfw.Decorated, bool2int[!cfg.Undecorated])
		glfw.WindowHint(glfw.Focused, bool2int[!cfg.Unfocused])
		glfw.WindowHint(glfw.Maximized, bool2int[cfg.Maximized])

		var share *glfw.Window
		if currentWindow != nil {
			share = currentWindow.window
		}
		_, _, width, height := intBounds(cfg.Bounds)
		w.window, err = glfw.CreateWindow(
			width,
			height,
			cfg.Title,
			nil,
			share,
		)
		if err != nil {
			return err
		}

		// enter the OpenGL context
		w.begin()
		w.end()

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating window failed")
	}

	w.SetVSync(cfg.VSync)

	w.initInput()
	w.SetMonitor(cfg.Fullscreen)

	w.canvas = NewCanvas(cfg.Bounds, false)
	w.Update()

	runtime.SetFinalizer(w, (*Window).Destroy)

	return w, nil
}

// Destroy destroys the Window. The Window can't be used any further.
func (w *Window) Destroy() {
	mainthread.Call(func() {
		w.window.Destroy()
	})
}

// Update swaps buffers and polls events.
func (w *Window) Update() {
	mainthread.Call(func() {
		wi, hi := w.window.GetSize()
		w.bounds.Size = pixel.V(float64(wi), float64(hi))
		// fractional positions end up covering more pixels with less size
		if w.bounds.X() != math.Floor(w.bounds.X()) {
			w.bounds.Size -= pixel.V(1, 0)
		}
		if w.bounds.Y() != math.Floor(w.bounds.Y()) {
			w.bounds.Size -= pixel.V(0, 1)
		}
	})

	w.canvas.SetBounds(w.Bounds())

	mainthread.Call(func() {
		w.begin()

		glhf.Bounds(0, 0, w.canvas.f.Texture().Width(), w.canvas.f.Texture().Height())

		glhf.Clear(0, 0, 0, 0)
		w.canvas.f.Begin()
		w.canvas.f.Blit(
			nil,
			0, 0, w.canvas.f.Texture().Width(), w.canvas.f.Texture().Height(),
			0, 0, w.canvas.f.Texture().Width(), w.canvas.f.Texture().Height(),
		)
		w.canvas.f.End()

		if w.vsync {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}
		w.window.SwapBuffers()
		w.end()
	})

	w.updateInput()
}

// SetClosed sets the closed flag of the Window.
//
// This is usefull when overriding the user's attempt to close the Window, or just to close the
// Window from within the program.
func (w *Window) SetClosed(closed bool) {
	mainthread.Call(func() {
		w.window.SetShouldClose(closed)
	})
}

// Closed returns the closed flag of the Window, which reports whether the Window should be closed.
//
// The closed flag is automatically set when a user attempts to close the Window.
func (w *Window) Closed() bool {
	return mainthread.CallVal(func() interface{} {
		return w.window.ShouldClose()
	}).(bool)
}

// SetTitle changes the title of the Window.
func (w *Window) SetTitle(title string) {
	mainthread.Call(func() {
		w.window.SetTitle(title)
	})
}

// SetBounds sets the bounds of the Window in pixels. Bounds can be fractional, but the size will be
// changed in the next Update to a real possible size of the Window.
func (w *Window) SetBounds(bounds pixel.Rect) {
	w.bounds = bounds
	mainthread.Call(func() {
		_, _, width, height := intBounds(bounds)
		w.window.SetSize(width, height)
	})
}

// Bounds returns the current bounds of the Window.
func (w *Window) Bounds() pixel.Rect {
	return w.bounds
}

// Show makes the Window visible if it was hidden.
func (w *Window) Show() {
	mainthread.Call(func() {
		w.window.Show()
	})
}

// Hide hides the Window if it was visible.
func (w *Window) Hide() {
	mainthread.Call(func() {
		w.window.Hide()
	})
}

func (w *Window) setFullscreen(monitor *Monitor) {
	mainthread.Call(func() {
		w.restore.xpos, w.restore.ypos = w.window.GetPos()
		w.restore.width, w.restore.height = w.window.GetSize()

		mode := monitor.monitor.GetVideoMode()

		w.window.SetMonitor(
			monitor.monitor,
			0,
			0,
			mode.Width,
			mode.Height,
			mode.RefreshRate,
		)
	})
}

func (w *Window) setWindowed() {
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
}

// SetMonitor sets the Window fullscreen on the given Monitor. If the Monitor is nil, the Window
// will be resored to windowed state instead.
//
// Note, that there is nothing about the resolution of the fullscreen Window. The Window is
// automatically set to the Monitor's resolution. If you want a different resolution, you need
// to set it manually with SetSize method.
func (w *Window) SetMonitor(monitor *Monitor) {
	if w.Monitor() != monitor {
		if monitor != nil {
			w.setFullscreen(monitor)
		} else {
			w.setWindowed()
		}
	}
}

// IsFullscreen returns true if the Window is in fullscreen mode.
func (w *Window) IsFullscreen() bool {
	return w.Monitor() != nil
}

// Monitor returns a monitor the Window is fullscreen is on. If the Window is not fullscreen, this
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

// Focus brings the Window to the front and sets input focus.
func (w *Window) Focus() {
	mainthread.Call(func() {
		w.window.Focus()
	})
}

// Focused returns true if the Window has input focus.
func (w *Window) Focused() bool {
	return mainthread.CallVal(func() interface{} {
		return w.window.GetAttrib(glfw.Focused) == glfw.True
	}).(bool)
}

// Maximize puts the Window window to the maximized state.
func (w *Window) Maximize() {
	mainthread.Call(func() {
		w.window.Maximize()
	})
}

// Restore restores the Window window from the maximized state.
func (w *Window) Restore() {
	mainthread.Call(func() {
		w.window.Restore()
	})
}

// SetVSync sets whether the Window should synchronize with the monitor refresh rate.
func (w *Window) SetVSync(vsync bool) {
	w.vsync = vsync
}

// VSync returns whether the Window is set to synchronize with the monitor refresh rate.
func (w *Window) VSync() bool {
	return w.vsync
}

// Note: must be called inside the main thread.
func (w *Window) begin() {
	if currentWindow != w {
		w.window.MakeContextCurrent()
		glhf.Init()
		currentWindow = w
	}
}

// Note: must be called inside the main thread.
func (w *Window) end() {
	// nothing, really
}

// MakeTriangles generates a specialized copy of the supplied Triangles that will draw onto this
// Window.
//
// Window supports TrianglesPosition, TrianglesColor and TrianglesTexture.
func (w *Window) MakeTriangles(t pixel.Triangles) pixel.TargetTriangles {
	return w.canvas.MakeTriangles(t)
}

// MakePicture generates a specialized copy of the supplied Picture that will draw onto this Window.
//
// Window support PictureColor.
func (w *Window) MakePicture(p pixel.Picture) pixel.TargetPicture {
	return w.canvas.MakePicture(p)
}

// SetMatrix sets a Matrix that every point will be projected by.
func (w *Window) SetMatrix(m pixel.Matrix) {
	w.canvas.SetMatrix(m)
}

// SetColorMask sets a global color mask for the Window.
func (w *Window) SetColorMask(c color.Color) {
	w.canvas.SetColorMask(c)
}

// SetSmooth sets whether the stretched Pictures drawn onto this Window should be drawn smooth or
// pixely.
func (w *Window) SetSmooth(smooth bool) {
	w.canvas.SetSmooth(smooth)
}

// Smooth returns whether the stretched Pictures drawn onto this Window are set to be drawn smooth
// or pixely.
func (w *Window) Smooth() bool {
	return w.canvas.Smooth()
}

// Clear clears the Window with a color.
func (w *Window) Clear(c color.Color) {
	w.canvas.Clear(c)
}

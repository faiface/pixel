package pixel

import (
	"image/color"
	"sync"

	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// WindowConfig is convenience structure for specifying all possible properties of a window.
// Properties are chosen in such a way, that you usually only need to set a few of them - defaults
// (zeros) should usually be sensible.
//
// Note that you always need to set the width and the height of a window.
type WindowConfig struct {
	// Title at the top of a window.
	Title string

	// Width of a window in pixels.
	Width float64

	// Height of a window in pixels.
	Height float64

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

	// VSync (vertical synchronization) synchronizes window's framerate with the framerate of the monitor.
	VSync bool

	// Number of samples for multi-sample anti-aliasing (edge-smoothing).
	// Usual values are 0, 2, 4, 8 (powers of 2 and not much more than this).
	MSAASamples int
}

// Window is a window handler. Use this type to manipulate a window (input, drawing, ...).
type Window struct {
	window *glfw.Window
	config WindowConfig
}

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

		var err error
		w.window, err = glfw.CreateWindow(int(config.Width), int(config.Height), config.Title, nil, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating window failed")
	}

	return w, nil
}

// Clear clears the window with a color.
func (w *Window) Clear(c color.Color) {
	w.Begin()
	pixelgl.Clear(colorToRGBA(c))
	w.End()
}

// Update swaps buffers and polls events.
func (w *Window) Update() {
	w.Begin()
	pixelgl.Do(func() {
		if w.config.VSync {
			glfw.SwapInterval(1)
		}
		w.window.SwapBuffers()
		glfw.PollEvents()
	})
	w.End()
}

var currentWindow struct {
	sync.Mutex
	handler *Window
}

// Begin makes the context of this window current.
//
// Note that you only need to use this function if you're designing a low-level technical plugin (such as an effect).
func (w *Window) Begin() {
	currentWindow.Lock()
	if currentWindow.handler != w {
		pixelgl.Do(func() {
			w.window.MakeContextCurrent()
			pixelgl.Init()
		})
		currentWindow.handler = w
	}
}

// End makes it possible for other windows to make their context current.
//
// Note that you only need to use this function if you're designing a low-level technical plugin (such as an effect).
func (w *Window) End() {
	currentWindow.Unlock()
}

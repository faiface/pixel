package pixel

import (
	"sync"

	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

//TODO: better doc

// WindowConfig is convenience structure for specifying all possible properties of a window.
// Properties are chosen in such a way, that you usually only need to set a few of them - defaults
// (zeros) should usually be sensible.
//
// Note that you always need to set the width and the height of a window.
type WindowConfig struct {
	Title       string
	Width       float64
	Height      float64
	Resizable   bool
	Hidden      bool
	Undecorated bool
	Unfocused   bool
	Maximized   bool
	VSync       bool
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
func (w *Window) Clear(r, g, b, a float64) {
	w.Begin()
	pixelgl.Clear(r, g, b, a)
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

var (
	windowMutex   sync.Mutex
	currentWindow *Window
)

// Begin makes the context of this window current.
func (w *Window) Begin() {
	windowMutex.Lock()
	if currentWindow != w {
		pixelgl.Do(func() {
			w.window.MakeContextCurrent()
			pixelgl.Init()
		})
		currentWindow = w
	}
}

// End makes it possible for other windows to make their context current.
func (w *Window) End() {
	windowMutex.Unlock()
}

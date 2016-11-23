package pixel

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

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

type Window struct {
	window *glfw.Window
	config WindowConfig
}

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

func (w *Window) Clear(r, g, b, a float64) {
	w.Begin()
	pixelgl.Do(func() {
		gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
	w.End()
}

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

var currentWindow *Window = nil

func (w *Window) Begin() {
	pixelgl.Do(func() {
		if currentWindow != w {
			w.window.MakeContextCurrent()
			pixelgl.Init()
			currentWindow = w
		}
	})

}

func (w *Window) End() {
	// nothing really
}

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

	// If set to nil, a window will be windowed. Otherwise it will be fullscreen on the specified monitor.
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

	// VSync (vertical synchronization) synchronizes window's framerate with the framerate of the monitor.
	VSync bool

	// Number of samples for multi-sample anti-aliasing (edge-smoothing).
	// Usual values are 0, 2, 4, 8 (powers of 2 and not much more than this).
	MSAASamples int
}

// Window is a window handler. Use this type to manipulate a window (input, drawing, ...).
type Window struct {
	window        *glfw.Window
	config        WindowConfig
	contextHolder pixelgl.ContextHolder

	// need to save these to correctly restore a fullscreen window
	restore struct {
		xpos, ypos, width, height int
	}
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

		w.window, err = glfw.CreateWindow(int(config.Width), int(config.Height), config.Title, nil, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating window failed")
	}

	w.SetFullscreen(config.Fullscreen)

	defaultShader, err := pixelgl.NewShader(w, defaultVertexFormat, defaultUniformFormat, defaultVertexShader, defaultFragmentShader)
	if err != nil {
		w.Delete()
		return nil, errors.Wrap(err, "creating window failed")
	}

	w.contextHolder.Context = w.contextHolder.Context.WithShader(defaultShader)

	return w, nil
}

// Delete destroys a window. The window can't be used any further.
func (w *Window) Delete() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Destroy()
		})
	})
}

// Clear clears the window with a color.
func (w *Window) Clear(c color.Color) {
	w.Do(func(pixelgl.Context) {
		pixelgl.Clear(colorToRGBA(c))
	})
}

// Update swaps buffers and polls events.
func (w *Window) Update() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			if w.config.VSync {
				glfw.SwapInterval(1)
			}
			w.window.SwapBuffers()
			glfw.PollEvents()
		})
	})
}

// SetTitle changes the title of a window.
func (w *Window) SetTitle(title string) {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.SetTitle(title)
		})
	})
}

// SetSize resizes a window to the specified size in pixels.
// In case of a fullscreen window, it changes the resolution of that window.
func (w *Window) SetSize(width, height float64) {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.SetSize(int(width), int(height))
		})
	})
}

// Size returns the size of the client area of a window (the part you can draw on).
func (w *Window) Size() (width, height float64) {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			wi, hi := w.window.GetSize()
			width = float64(wi)
			height = float64(hi)
		})
	})
	return width, height
}

// Show makes a window visible if it was hidden.
func (w *Window) Show() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Show()
		})
	})
}

// Hide hides a window if it was visible.
func (w *Window) Hide() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Hide()
		})
	})
}

// SetFullscreen sets a window fullscreen on a given monitor. If the monitor is nil, the window will be resored to windowed instead.
//
// Note, that there is nothing about the resolution of the fullscreen window. The window is automatically set to the monitor's
// resolution. If you want a different resolution, you need to set it manually with SetSize method.
func (w *Window) SetFullscreen(monitor *Monitor) {
	if w.Monitor() != monitor {
		if monitor == nil {
			w.Do(func(pixelgl.Context) {
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
			})
		} else {
			w.Do(func(pixelgl.Context) {
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
			})
		}
	}
}

// IsFullscreen returns true if the window is in the fullscreen mode.
func (w *Window) IsFullscreen() bool {
	return w.Monitor() != nil
}

// Monitor returns a monitor a fullscreen window is on. If the window is not fullscreen, this function returns nil.
func (w *Window) Monitor() *Monitor {
	var monitor *glfw.Monitor
	w.Do(func(pixelgl.Context) {
		monitor = pixelgl.DoVal(func() interface{} {
			return w.window.GetMonitor()
		}).(*glfw.Monitor)
	})
	if monitor == nil {
		return nil
	}
	return &Monitor{
		monitor: monitor,
	}
}

// Focus brings a window to the front and sets input focus.
func (w *Window) Focus() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Focus()
		})
	})
}

// Focused returns true if a window has input focus.
func (w *Window) Focused() bool {
	var focused bool
	w.Do(func(pixelgl.Context) {
		focused = pixelgl.DoVal(func() interface{} {
			return w.window.GetAttrib(glfw.Focused) == glfw.True
		}).(bool)
	})
	return focused
}

// Maximize puts a windowed window to a maximized state.
func (w *Window) Maximize() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Maximize()
		})
	})
}

// Restore restores a windowed window from a maximized state.
func (w *Window) Restore() {
	w.Do(func(pixelgl.Context) {
		pixelgl.Do(func() {
			w.window.Restore()
		})
	})
}

var currentWindow struct {
	sync.Mutex
	handler *Window
}

// Do makes the context of this window current, if it's not already, and executes sub.
func (w *Window) Do(sub func(pixelgl.Context)) {
	currentWindow.Lock()
	defer currentWindow.Unlock()

	if currentWindow.handler != w {
		pixelgl.Do(func() {
			w.window.MakeContextCurrent()
			pixelgl.Init()
		})
		currentWindow.handler = w
	}

	w.contextHolder.Do(sub)
}

var defaultVertexFormat = pixelgl.VertexFormat{
	"position": {Purpose: pixelgl.Position, Type: pixelgl.Vec2},
	"color":    {Purpose: pixelgl.Color, Type: pixelgl.Vec4},
	"texCoord": {Purpose: pixelgl.TexCoord, Type: pixelgl.Vec2},
}

var defaultUniformFormat = pixelgl.UniformFormat{
	"transform": {Purpose: pixelgl.Transform, Type: pixelgl.Mat3},
	"isTexture": {Purpose: pixelgl.IsTexture, Type: pixelgl.Int},
}

var defaultVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texCoord;

out vec4 Color;
out vec2 TexCoord;

uniform mat3 transform;

void main() {
	gl_Position = vec4((transform * vec3(position.x, position.y, 1.0)).xy, 0.0, 1.0);
	Color = color;
	TexCoord = texCoord;
}
`

var defaultFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 TexCoord;

out vec4 color;

uniform int isTexture;
uniform sampler2D tex;

void main() {
	if (isTexture != 0) {
		color = Color * texture(tex, vec2(TexCoord.x, 1 - TexCoord.y));
	} else {
		color = Color;
	}
}
`

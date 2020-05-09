package pixelgl

import (
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Monitor represents a physical display attached to your computer.
type Monitor struct {
	monitor *glfw.Monitor
}

// VideoMode represents all properties of a video mode and is
// associated with a monitor if it is used in fullscreen mode.
type VideoMode struct {
	// Width is the width of the vide mode in pixels.
	Width int
	// Height is the height of the video mode in pixels.
	Height int
	// RefreshRate holds the refresh rate of the associated monitor in Hz.
	RefreshRate int
}

// PrimaryMonitor returns the main monitor (usually the one with the taskbar and stuff).
func PrimaryMonitor() *Monitor {
	var monitor *glfw.Monitor
	mainthread.Call(func() {
		monitor = glfw.GetPrimaryMonitor()
	})
	return &Monitor{
		monitor: monitor,
	}
}

// Monitors returns a slice of all currently available monitors.
func Monitors() []*Monitor {
	var monitors []*Monitor
	mainthread.Call(func() {
		for _, monitor := range glfw.GetMonitors() {
			monitors = append(monitors, &Monitor{monitor: monitor})
		}
	})
	return monitors
}

// Name returns a human-readable name of the Monitor.
func (m *Monitor) Name() string {
	var name string
	mainthread.Call(func() {
		name = m.monitor.GetName()
	})
	return name
}

// PhysicalSize returns the size of the display area of the Monitor in millimeters.
func (m *Monitor) PhysicalSize() (width, height float64) {
	var wi, hi int
	mainthread.Call(func() {
		wi, hi = m.monitor.GetPhysicalSize()
	})
	width = float64(wi)
	height = float64(hi)
	return
}

// Position returns the position of the upper-left corner of the Monitor in screen coordinates.
func (m *Monitor) Position() (x, y float64) {
	var xi, yi int
	mainthread.Call(func() {
		xi, yi = m.monitor.GetPos()
	})
	x = float64(xi)
	y = float64(yi)
	return
}

// Size returns the resolution of the Monitor in pixels.
func (m *Monitor) Size() (width, height float64) {
	var mode *glfw.VidMode
	mainthread.Call(func() {
		mode = m.monitor.GetVideoMode()
	})
	width = float64(mode.Width)
	height = float64(mode.Height)
	return
}

// BitDepth returns the number of bits per color of the Monitor.
func (m *Monitor) BitDepth() (red, green, blue int) {
	var mode *glfw.VidMode
	mainthread.Call(func() {
		mode = m.monitor.GetVideoMode()
	})
	red = mode.RedBits
	green = mode.GreenBits
	blue = mode.BlueBits
	return
}

// RefreshRate returns the refresh frequency of the Monitor in Hz (refreshes/second).
func (m *Monitor) RefreshRate() (rate float64) {
	var mode *glfw.VidMode
	mainthread.Call(func() {
		mode = m.monitor.GetVideoMode()
	})
	rate = float64(mode.RefreshRate)
	return
}

// VideoModes returns all available video modes for the monitor.
func (m *Monitor) VideoModes() (vmodes []VideoMode) {
	var modes []*glfw.VidMode
	mainthread.Call(func() {
		modes = m.monitor.GetVideoModes()
	})
	for _, mode := range modes {
		vmodes = append(vmodes, VideoMode{
			Width:       mode.Width,
			Height:      mode.Height,
			RefreshRate: mode.RefreshRate,
		})
	}
	return
}

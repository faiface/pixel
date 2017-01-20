package pixel

import (
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Monitor represents a physical display attached to your computer.
type Monitor struct {
	monitor *glfw.Monitor
}

// PrimaryMonitor returns the main monitor (usually the one with the taskbar and stuff).
func PrimaryMonitor() *Monitor {
	monitor := mainthread.CallVal(func() interface{} {
		return glfw.GetPrimaryMonitor()
	}).(*glfw.Monitor)
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

// Name returns a human-readable name of a monitor.
func (m *Monitor) Name() string {
	name := mainthread.CallVal(func() interface{} {
		return m.monitor.GetName()
	}).(string)
	return name
}

// PhysicalSize returns the size of the display area of a monitor in millimeters.
func (m *Monitor) PhysicalSize() (width, height float64) {
	var wi, hi int
	mainthread.Call(func() {
		wi, hi = m.monitor.GetPhysicalSize()
	})
	width = float64(wi)
	height = float64(hi)
	return
}

// Position returns the position of the upper-left corner of a monitor in screen coordinates.
func (m *Monitor) Position() (x, y float64) {
	var xi, yi int
	mainthread.Call(func() {
		xi, yi = m.monitor.GetPos()
	})
	x = float64(xi)
	y = float64(yi)
	return
}

// Size returns the resolution of a monitor in pixels.
func (m *Monitor) Size() (width, height float64) {
	mode := mainthread.CallVal(func() interface{} {
		return m.monitor.GetVideoMode()
	}).(*glfw.VidMode)
	width = float64(mode.Width)
	height = float64(mode.Height)
	return
}

// BitDepth returns the number of bits per color of a monitor.
func (m *Monitor) BitDepth() (red, green, blue int) {
	mode := mainthread.CallVal(func() interface{} {
		return m.monitor.GetVideoMode()
	}).(*glfw.VidMode)
	red = mode.RedBits
	green = mode.GreenBits
	blue = mode.BlueBits
	return
}

// RefreshRate returns the refresh frequency of a monitor in Hz (refreshes/second).
func (m *Monitor) RefreshRate() (rate float64) {
	mode := mainthread.CallVal(func() interface{} {
		return m.monitor.GetVideoMode()
	}).(*glfw.VidMode)
	rate = float64(mode.RefreshRate)
	return
}

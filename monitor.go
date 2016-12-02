package pixel

import "github.com/go-gl/glfw/v3.2/glfw"

// Monitor represents a physical display attached to your computer.
type Monitor struct {
	monitor *glfw.Monitor
}

// PrimaryMonitor returns the main monitor (usually the one with the taskbar and stuff).
func PrimaryMonitor() *Monitor {
	return &Monitor{
		monitor: glfw.GetPrimaryMonitor(),
	}
}

// Monitors returns a slice of all currently available monitors.
func Monitors() []*Monitor {
	var monitors []*Monitor
	for _, monitor := range glfw.GetMonitors() {
		monitors = append(monitors, &Monitor{monitor: monitor})
	}
	return monitors
}

// Name returns a human-readable name of a monitor.
func (m *Monitor) Name() string {
	return m.monitor.GetName()
}

// PhysicalSize returns the size of the display area of a monitor in millimeters.
func (m *Monitor) PhysicalSize() (width, height float64) {
	wi, hi := m.monitor.GetPhysicalSize()
	width = float64(wi)
	height = float64(hi)
	return
}

// Position returns the position of the upper-left corner of a monitor in screen coordinates.
func (m *Monitor) Position() (x, y float64) {
	xi, yi := m.monitor.GetPos()
	x = float64(xi)
	y = float64(yi)
	return
}

// Size returns the resolution of a monitor in pixels.
func (m *Monitor) Size() (width, height float64) {
	mode := m.monitor.GetVideoMode()
	width = float64(mode.Width)
	height = float64(mode.Height)
	return
}

// BitDepth returns the number of bits per color of a monitor.
func (m *Monitor) BitDepth() (red, green, blue int) {
	mode := m.monitor.GetVideoMode()
	red = mode.RedBits
	green = mode.GreenBits
	blue = mode.BlueBits
	return
}

// RefreshRate returns the refresh frequency of a monitor in Hz (refreshes/second).
func (m *Monitor) RefreshRate() (rate float64) {
	mode := m.monitor.GetVideoMode()
	rate = float64(mode.RefreshRate)
	return
}

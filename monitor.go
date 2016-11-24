package pixel

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Monitor represents a physical display attached to your computer.
type Monitor struct {
	monitor *glfw.Monitor
}

// PrimaryMonitor returns the main monitor (usually the one with the taskbar and stuff).
func PrimaryMonitor() Monitor {
	m := Monitor{}
	pixelgl.Do(func() {
		m.monitor = glfw.GetPrimaryMonitor()
	})
	return m
}

// Monitors returns a slice of all currently attached monitors.
func Monitors() []Monitor {
	var monitors []Monitor
	pixelgl.Do(func() {
		for _, monitor := range glfw.GetMonitors() {
			monitors = append(monitors, Monitor{monitor: monitor})
		}
	})
	return monitors
}

// Name returns a human-readable name of a monitor.
func (m Monitor) Name() string {
	return pixelgl.DoVal(func() interface{} {
		return m.monitor.GetName()
	}).(string)
}

// PhysicalSize returns the size of the display are of a monitor in millimeters.
func (m Monitor) PhysicalSize() (width, height float64) {
	var w, h float64
	pixelgl.Do(func() {
		wi, hi := m.monitor.GetPhysicalSize()
		w = float64(wi)
		h = float64(hi)
	})
	return w, h
}

// Position returns the position of the upper-left corner of a monitor in screen coordinates.
func (m Monitor) Position() (x, y float64) {
	pixelgl.Do(func() {
		xi, yi := m.monitor.GetPos()
		x = float64(xi)
		y = float64(yi)
	})
	return x, y
}

// Size returns the resolution of a monitor in pixels.
func (m Monitor) Size() (width, height float64) {
	var w, h float64
	pixelgl.Do(func() {
		mode := m.monitor.GetVideoMode()
		w = float64(mode.Width)
		h = float64(mode.Height)
	})
	return w, h
}

// BitDepth returns the number of bits per color of a monitor.
func (m Monitor) BitDepth() (red, green, blue int) {
	var r, g, b int
	pixelgl.Do(func() {
		mode := m.monitor.GetVideoMode()
		r = mode.RedBits
		g = mode.GreenBits
		b = mode.BlueBits
	})
	return r, g, b
}

// RefreshRate returns the refresh frequency of a monitor in Hz (refreshes/second).
func (m Monitor) RefreshRate() float64 {
	var rate float64
	pixelgl.Do(func() {
		mode := m.monitor.GetVideoMode()
		rate = float64(mode.RefreshRate)
	})
	return rate
}

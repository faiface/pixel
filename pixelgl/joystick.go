package pixelgl

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Joystick is a joystick or controller.
type Joystick int

// List all of the joysticks.
const (
	Joystick1  = Joystick(glfw.Joystick1)
	Joystick2  = Joystick(glfw.Joystick2)
	Joystick3  = Joystick(glfw.Joystick3)
	Joystick4  = Joystick(glfw.Joystick4)
	Joystick5  = Joystick(glfw.Joystick5)
	Joystick6  = Joystick(glfw.Joystick6)
	Joystick7  = Joystick(glfw.Joystick7)
	Joystick8  = Joystick(glfw.Joystick8)
	Joystick9  = Joystick(glfw.Joystick9)
	Joystick10 = Joystick(glfw.Joystick10)
	Joystick11 = Joystick(glfw.Joystick11)
	Joystick12 = Joystick(glfw.Joystick12)
	Joystick13 = Joystick(glfw.Joystick13)
	Joystick14 = Joystick(glfw.Joystick14)
	Joystick15 = Joystick(glfw.Joystick15)
	Joystick16 = Joystick(glfw.Joystick16)

	JoystickLast = Joystick(glfw.JoystickLast)
)

// JoystickPresent returns if the joystick is currently connected.
//
// This API is experimental.
func (w *Window) JoystickPresent(js Joystick) bool {
	return w.currJoy.connected[js]
}

// JoystickName returns the name of the joystick. A disconnected joystick will return an
// empty string.
//
// This API is experimental.
func (w *Window) JoystickName(js Joystick) string {
	return w.currJoy.name[js]
}

// JoystickButtonCount returns the number of buttons a connected joystick has.
//
// This API is experimental.
func (w *Window) JoystickButtonCount(js Joystick) int {
	return len(w.currJoy.buttons[js])
}

// JoystickAxisCount returns the number of axes a connected joystick has.
//
// This API is experimental.
func (w *Window) JoystickAxisCount(js Joystick) int {
	return len(w.currJoy.axis[js])
}

// JoystickPressed returns whether the joystick Button is currently pressed down.
// If the button index is out of range, this will return false.
//
// This API is experimental.
func (w *Window) JoystickPressed(js Joystick, button int) bool {
	return w.currJoy.getButton(js, button)
}

// JoystickJustPressed returns whether the joystick Button has just been pressed down.
// If the button index is out of range, this will return false.
//
// This API is experimental.
func (w *Window) JoystickJustPressed(js Joystick, button int) bool {
	return w.currJoy.getButton(js, button) && !w.prevJoy.getButton(js, button)
}

// JoystickJustReleased returns whether the joystick Button has just been released up.
// If the button index is out of range, this will return false.
//
// This API is experimental.
func (w *Window) JoystickJustReleased(js Joystick, button int) bool {
	return !w.currJoy.getButton(js, button) && w.prevJoy.getButton(js, button)
}

// JoystickAxis returns the value of a joystick axis at the last call to Window.Update.
// If the axis index is out of range, this will return 0.
//
// This API is experimental.
func (w *Window) JoystickAxis(js Joystick, axis int) float64 {
	return w.currJoy.getAxis(js, axis)
}

// Used internally during Window.UpdateInput to update the state of the joysticks.
func (w *Window) updateJoystickInput() {
	for js := Joystick1; js <= JoystickLast; js++ {
		// Determine and store if the joystick was connected
		joystickPresent := glfw.JoystickPresent(glfw.Joystick(js))
		w.tempJoy.connected[js] = joystickPresent

		if joystickPresent {
			w.tempJoy.buttons[js] = glfw.GetJoystickButtons(glfw.Joystick(js))
			w.tempJoy.axis[js] = glfw.GetJoystickAxes(glfw.Joystick(js))

			if !w.currJoy.connected[js] {
				// The joystick was recently connected, we get the name
				w.tempJoy.name[js] = glfw.GetJoystickName(glfw.Joystick(js))
			} else {
				// Use the name from the previous one
				w.tempJoy.name[js] = w.currJoy.name[js]
			}
		} else {
			w.tempJoy.buttons[js] = []byte{}
			w.tempJoy.axis[js] = []float32{}
			w.tempJoy.name[js] = ""
		}
	}

	w.prevJoy = w.currJoy
	w.currJoy = w.tempJoy
}

type joystickState struct {
	connected [JoystickLast + 1]bool
	name      [JoystickLast + 1]string
	buttons   [JoystickLast + 1][]byte
	axis      [JoystickLast + 1][]float32
}

// Returns if a button on a joystick is down, returning false if the button or joystick is invalid.
func (js *joystickState) getButton(joystick Joystick, button int) bool {
	// Check that the joystick and button is valid, return false by default
	if js.buttons[joystick] == nil || button >= len(js.buttons[joystick]) || button < 0 {
		return false
	}
	return js.buttons[joystick][byte(button)] == 1
}

// Returns the value of a joystick axis, returning 0 if the button or joystick is invalid.
func (js *joystickState) getAxis(joystick Joystick, axis int) float64 {
	// Check that the joystick and axis is valid, return 0 by default.
	if js.axis[joystick] == nil || axis >= len(js.axis[joystick]) || axis < 0 {
		return 0
	}
	return float64(js.axis[joystick][axis])
}

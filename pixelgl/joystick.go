package pixelgl

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Joystick is a joystick or controller (gamepad).
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

// GamepadAxis corresponds to a gamepad axis.
type GamepadAxis int

// Gamepad axis IDs.
const (
	AxisLeftX        = GamepadAxis(glfw.AxisLeftX)
	AxisLeftY        = GamepadAxis(glfw.AxisLeftY)
	AxisRightX       = GamepadAxis(glfw.AxisRightX)
	AxisRightY       = GamepadAxis(glfw.AxisRightY)
	AxisLeftTrigger  = GamepadAxis(glfw.AxisLeftTrigger)
	AxisRightTrigger = GamepadAxis(glfw.AxisRightTrigger)
	AxisLast         = GamepadAxis(glfw.AxisLast)
)

// GamepadButton corresponds to a gamepad button.
type GamepadButton int

// Gamepad button IDs.
const (
	ButtonA           = GamepadButton(glfw.ButtonA)
	ButtonB           = GamepadButton(glfw.ButtonB)
	ButtonX           = GamepadButton(glfw.ButtonX)
	ButtonY           = GamepadButton(glfw.ButtonY)
	ButtonLeftBumper  = GamepadButton(glfw.ButtonLeftBumper)
	ButtonRightBumper = GamepadButton(glfw.ButtonRightBumper)
	ButtonBack        = GamepadButton(glfw.ButtonBack)
	ButtonStart       = GamepadButton(glfw.ButtonStart)
	ButtonGuide       = GamepadButton(glfw.ButtonGuide)
	ButtonLeftThumb   = GamepadButton(glfw.ButtonLeftThumb)
	ButtonRightThumb  = GamepadButton(glfw.ButtonRightThumb)
	ButtonDpadUp      = GamepadButton(glfw.ButtonDpadUp)
	ButtonDpadRight   = GamepadButton(glfw.ButtonDpadRight)
	ButtonDpadDown    = GamepadButton(glfw.ButtonDpadDown)
	ButtonDpadLeft    = GamepadButton(glfw.ButtonDpadLeft)
	ButtonLast        = GamepadButton(glfw.ButtonLast)
	ButtonCross       = GamepadButton(glfw.ButtonCross)
	ButtonCircle      = GamepadButton(glfw.ButtonCircle)
	ButtonSquare      = GamepadButton(glfw.ButtonSquare)
	ButtonTriangle    = GamepadButton(glfw.ButtonTriangle)
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
func (w *Window) JoystickPressed(js Joystick, button GamepadButton) bool {
	return w.currJoy.getButton(js, int(button))
}

// JoystickJustPressed returns whether the joystick Button has just been pressed down.
// If the button index is out of range, this will return false.
//
// This API is experimental.
func (w *Window) JoystickJustPressed(js Joystick, button GamepadButton) bool {
	return w.currJoy.getButton(js, int(button)) && !w.prevJoy.getButton(js, int(button))
}

// JoystickJustReleased returns whether the joystick Button has just been released up.
// If the button index is out of range, this will return false.
//
// This API is experimental.
func (w *Window) JoystickJustReleased(js Joystick, button GamepadButton) bool {
	return !w.currJoy.getButton(js, int(button)) && w.prevJoy.getButton(js, int(button))
}

// JoystickAxis returns the value of a joystick axis at the last call to Window.Update.
// If the axis index is out of range, this will return 0.
//
// This API is experimental.
func (w *Window) JoystickAxis(js Joystick, axis GamepadAxis) float64 {
	return w.currJoy.getAxis(js, int(axis))
}

// Used internally during Window.UpdateInput to update the state of the joysticks.
func (w *Window) updateJoystickInput() {
	for js := Joystick1; js <= JoystickLast; js++ {
		// Determine and store if the joystick was connected
		joystickPresent := glfw.Joystick(js).Present()
		w.tempJoy.connected[js] = joystickPresent

		if joystickPresent {
			if glfw.Joystick(js).IsGamepad() {
				gamepadInputs := glfw.Joystick(js).GetGamepadState()

				w.tempJoy.buttons[js] = gamepadInputs.Buttons[:]
				w.tempJoy.axis[js] = gamepadInputs.Axes[:]
			} else {
				w.tempJoy.buttons[js] = glfw.Joystick(js).GetButtons()
				w.tempJoy.axis[js] = glfw.Joystick(js).GetAxes()
			}

			if !w.currJoy.connected[js] {
				// The joystick was recently connected, we get the name
				w.tempJoy.name[js] = glfw.Joystick(js).GetName()
			} else {
				// Use the name from the previous one
				w.tempJoy.name[js] = w.currJoy.name[js]
			}
		} else {
			w.tempJoy.buttons[js] = []glfw.Action{}
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
	buttons   [JoystickLast + 1][]glfw.Action
	axis      [JoystickLast + 1][]float32
}

// Returns if a button on a joystick is down, returning false if the button or joystick is invalid.
func (js *joystickState) getButton(joystick Joystick, button int) bool {
	// Check that the joystick and button is valid, return false by default
	if js.buttons[joystick] == nil || button >= len(js.buttons[joystick]) || button < 0 {
		return false
	}
	return js.buttons[joystick][byte(button)] == glfw.Press
}

// Returns the value of a joystick axis, returning 0 if the button or joystick is invalid.
func (js *joystickState) getAxis(joystick Joystick, axis int) float64 {
	// Check that the joystick and axis is valid, return 0 by default.
	if js.axis[joystick] == nil || axis >= len(js.axis[joystick]) || axis < 0 {
		return 0
	}
	return float64(js.axis[joystick][axis])
}

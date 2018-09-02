package glossary

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

// Camera is a tool to get the screen center to be able to follow a certain point on a plane.
type Camera struct {
	anglePhysic    float64   // Angle in radians (math.Pi)
	angleFollow    float64   // Angle expected to be in the near future.
	zoomPosPhysic  float64   // Z
	zoomPosFollow  float64   // Z expected to be in the near future.
	planePosPhysic pixel.Vec // X, Y
	planePosFollow pixel.Vec // X, Y expected to be in the near future.
	screenBound    pixel.Rect
	moveSmooth     bool
}

// NewCamera is a constructor.
func NewCamera(_pos pixel.Vec, _screenBound pixel.Rect) *Camera {
	return &Camera{
		anglePhysic:    0,
		angleFollow:    0,
		zoomPosPhysic:  1.0,
		zoomPosFollow:  1.0,
		planePosPhysic: _pos,
		planePosFollow: _pos,
		screenBound:    _screenBound,
		moveSmooth:     true,
	}
}

// -------------------------------------------------------------------------
// Read only

// Transform returns a transformation matrix of a camera.
// Use Transform().Project() to convert a game position to a screen position.
// To do the inverse operation, it is recommended to use Camera#Unproject() rather than Transform().Unproject()
func (camera Camera) Transform() pixel.Matrix {
	return pixel.IM. // This transformation order is significant.
		// ScaledXY(camera.planePos, pixel.V(camera.zoomPos, camera.zoomPos)).
		Scaled(camera.planePosPhysic, camera.zoomPosPhysic).          // Scaling
		Rotated(camera.planePosPhysic, camera.anglePhysic).           // Rotatation
		Moved(camera.screenBound.Center().Sub(camera.planePosPhysic)) // Translation
}

// Unproject converts a screen position to a game position.
// This method is a replacement of Transform().Unproject() which might return a bit off position.
func (camera Camera) Unproject(screenPosition pixel.Vec) (gamePosition pixel.Vec) {
	matrix1 := pixel.IM.
		Scaled(camera.planePosPhysic, camera.zoomPosPhysic).          // Scaling
		Moved(camera.screenBound.Center().Sub(camera.planePosPhysic)) // Translation
	matrix2 := pixel.IM.
		Rotated(camera.planePosPhysic, -camera.anglePhysic) // Rotatation
	return matrix2.Project(matrix1.Unproject(screenPosition))
}

// Angle returns the angle of a camera in radians.
func (camera Camera) Angle() float64 {
	return camera.anglePhysic
}

// XYZ returns a camera's coordinates value X, Y, and Z in a current physical state.
func (camera Camera) XYZ() (float64, float64, float64) {
	return camera.planePosPhysic.X, camera.planePosPhysic.Y, camera.zoomPosPhysic
}

// XY returns the X and Y of a camera as a vector.
func (camera Camera) XY() pixel.Vec {
	return camera.planePosPhysic
}

// Z returns the zoom depth of a camera.
func (camera Camera) Z() float64 {
	return camera.zoomPosPhysic
}

// -------------------------------------------------------------------------
// Read and Write

// Update a camera's current physical state (physics)
// by calculating coordinates X, Y, Z and its angle after delta time in seconds.
func (camera *Camera) Update(dt float64) {
	if camera.moveSmooth { // lerp the camera position towards the target
		angle := pixel.V(camera.anglePhysic, 0)
		angleFollow := pixel.V(camera.angleFollow, 0)
		angle = pixel.Lerp(angle, angleFollow, 1-math.Pow(1.0/128, dt))
		camera.anglePhysic = angle.X
		camera.planePosPhysic = pixel.Lerp(camera.planePosPhysic, camera.planePosFollow, 1-math.Pow(1.0/128, dt))
		zoomPos := pixel.V(camera.zoomPosPhysic, 0)
		zoomFollow := pixel.V(camera.zoomPosFollow, 0)
		zoomPos = pixel.Lerp(zoomPos, zoomFollow, 1-math.Pow(1.0/128, dt))
		camera.zoomPosPhysic = zoomPos.X
	} else {
		camera.anglePhysic = camera.angleFollow
		camera.planePosPhysic = camera.planePosFollow
		camera.zoomPosPhysic = camera.zoomPosFollow
	}
}

// Rotate a camera by certain degrees.
// + ) Counterclockwise
// - ) Clockwise
func (camera *Camera) Rotate(degree float64) {
	camera.angleFollow += degree * math.Pi / 180
}

// Zoom in and out with a camera by certain levels.
// + ) Zoom in
// - ) Zoom out
func (camera *Camera) Zoom(byLevel float64) {
	const zoomAmount = 1.2
	camera.zoomPosFollow *= math.Pow(zoomAmount, byLevel)
}

// Move camera a specified distance.
func (camera *Camera) Move(distance pixel.Vec) {
	camera.planePosFollow = camera.planePosFollow.Add(distance)
}

// MoveTo () moves a camera to a point on a plane.
func (camera *Camera) MoveTo(posAim pixel.Vec) {
	camera.planePosFollow = posAim
}

// SetScreenBound of a camera.
func (camera *Camera) SetScreenBound(screenBound pixel.Rect) {
	camera.screenBound = screenBound
}

// -------------------------------------------------------------------
// Unnecessary

// Aim for experiments.
type Aim struct {
	pos pixel.Vec
}

// Draw aim as a dot.
func (aim Aim) Draw(t pixel.Target) {
	imd := imdraw.New(nil)
	imd.Color = colornames.Red
	imd.Push(aim.pos)
	imd.Circle(10, 0)
	imd.Draw(t)
}

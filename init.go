package pixel

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// Init initializes Pixel library. Call this function before using any of Pixel's functionality.
//
// If the initialization fails, an error is returned.
func Init() error {
	err := pixelgl.DoErr(func() error {
		return glfw.Init()
	})
	if err != nil {
		return errors.Wrap(err, "initializing GLFW failed")
	}
	return nil
}

// MustInit initializes Pixel library and panics when the initialization fails.
func MustInit() {
	err := Init()
	if err != nil {
		panic(err)
	}
}

// Quit terminates Pixel library. Call this function when you're done with Pixel.
func Quit() {
	pixelgl.Do(func() {
		glfw.Terminate()
	})
}

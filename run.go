package pixel

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/pkg/errors"
)

// Run is essentialy the "main" function of Pixel. It exists mainly due to the technical limitations of OpenGL and operating systems,
// in short, all graphics and window manipulating calls must be done from the main thread. Run makes this possible.
//
// Call this function from the main function of your application. This is necessary, so that Run runs on the main thread.
//
//   func run() {
//       window := pixel.NewWindow(...)
//       for {
//           // your game's main loop
//       }
//   }
//
//   func main() {
//       pixel.Run(run)
//   }
//
// You can spawn any number of goroutines from you run function and interact with Pixel concurrently.
// The only condition is that the Run function must be called from your main function.
func Run(run func()) error {
	err := glfw.Init()
	if err != nil {
		return errors.Wrap(err, "failed to initialize GLFW")
	}
	defer glfw.Terminate()

	pixelgl.Run(run)

	return nil
}

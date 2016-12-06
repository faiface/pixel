package pixelgl

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// Due to the limitations of OpenGL and operating systems, all OpenGL related calls must be done from the main thread.

var callQueue = make(chan func(), 32)

func init() {
	runtime.LockOSThread()
}

// Run is essentialy the "main" function of the pixelgl package.
// Run this function from the main function (because that's guaranteed to run in the main thread).
//
// This function reserves the main thread for the OpenGL stuff and runs a supplied run function in a
// separate goroutine.
//
// Run returns when the provided run function finishes.
func Run(run func()) {
	done := make(chan struct{})

	go func() {
		run()
		close(done)
	}()

loop:
	for {
		select {
		case f := <-callQueue:
			f()
		case <-done:
			break loop
		}
	}
}

// Init initializes OpenGL by loading the function pointers from the active OpenGL context.
// This function must be manually run inside the main thread (Do, DoErr, DoVal, etc.).
//
// It must be called under the presence of an active OpenGL context, e.g., always after calling window.MakeContextCurrent().
// Also, always call this function when switching contexts.
func Init() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

// DoNoBlock executes a function inside the main OpenGL thread.
// DoNoBlock does not wait until the function finishes.
func DoNoBlock(f func()) {
	callQueue <- f
}

// Do executes a function inside the main OpenGL thread.
// Do blocks until the function finishes.
//
// All OpenGL calls must be done in the dedicated thread.
func Do(f func()) {
	done := make(chan bool)
	callQueue <- func() {
		f()
		done <- true
	}
	<-done
}

// DoErr executes a function inside the main OpenGL thread and returns an error to the called.
// DoErr blocks until the function finishes.
//
// All OpenGL calls must be done in the dedicated thread.
func DoErr(f func() error) error {
	err := make(chan error)
	callQueue <- func() {
		err <- f()
	}
	return <-err
}

// DoVal executes a function inside the main OpenGL thread and returns a value to the caller.
// DoVal blocks until the function finishes.
//
// All OpenGL calls must be done in the main thread.
func DoVal(f func() interface{}) interface{} {
	val := make(chan interface{})
	callQueue <- func() {
		val <- f()
	}
	return <-val
}

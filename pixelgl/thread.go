package pixelgl

import "runtime"

// Due to the existance and usage of thread-local variables by OpenGL, it's recommended to
// execute all OpenGL calls from a single dedicated thread. This file defines functions to make
// it possible.

var (
	callQueue = make(chan func())

	//TODO: some OpenGL state variables will be here
)

func init() {
	go func() {
		runtime.LockOSThread()
		for f := range callQueue {
			f()
		}
	}()
}

// Do executes a function inside a dedicated OpenGL thread.
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

// DoErr executes a function inside a dedicated OpenGL thread and returns an error to the called.
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

// DoVal executes a function inside a dedicated OpenGL thread and returns a value to the caller.
// DoVal blocks until the function finishes.
//
// All OpenGL calls must be done in the dedicated thread.
func DoVal(f func() interface{}) interface{} {
	val := make(chan interface{})
	callQueue <- func() {
		val <- f()
	}
	return <-val
}

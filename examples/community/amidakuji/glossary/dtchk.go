package glossary

import (
	"errors"
	"time"
)

// DtWatch is a delta time checker.
type DtWatch struct {
	init *time.Time
	last *time.Time
}

// Start () is required in order to call other methods of a DtWatch.
func (watch *DtWatch) Start() {
	byVal1 := time.Now()
	byVal2 := byVal1
	watch.init = &byVal1
	watch.last = &byVal2
}

// IsStarted () determines whether it has started or not.
// .Start() is <not> required in order to call this method.
func (watch DtWatch) IsStarted() bool {
	if watch.init == nil {
		return false
	} else if watch.init != nil {
		return true
	} else {
		panic(errors.New("It might be thread"))
	}
}

// GetTimeStarted gets the time it started.
// .Start() must be called prior to calling this method.
func (watch DtWatch) GetTimeStarted() time.Time {
	return *watch.init
}

// SetTimeStarted sets the time it started.
// .Start() must be called prior to calling this method.
func (watch *DtWatch) SetTimeStarted(t time.Time) {
	*watch.init = t
}

// Dt since last Dt() or DtNano().
// .Start() must be called prior to calling this method.
func (watch *DtWatch) Dt() (deltaTimeInSeconds float64) {
	deltaTimeInSeconds = time.Since(time.Time(*watch.last)).Seconds()
	*watch.last = time.Now()
	return
}

// DtNano since last Dt() or DtNano().
// It returns a time instance with nanosecond precision.
// .Start() must be called prior to calling this method.
func (watch *DtWatch) DtNano() (deltaTimeInNanosec time.Duration) {
	deltaTimeInNanosec = time.Since(time.Time(*watch.last))
	*watch.last = time.Now()
	return
}

// DtSinceStart is dt since last Start().
// .Start() must be called prior to calling this method.
func (watch DtWatch) DtSinceStart() (deltaTimeInSeconds float64) {
	return time.Since(time.Time(*watch.init)).Seconds()
}

package audio

import "time"

// Ctrl allows for pausing and tracking a Streamer.
//
// Wrap a Streamer in a Ctrl.
//
//   ctrl := &audio.Ctrl{Streamer: s}
//
// Then, we can pause the streaming (this will cause Ctrl to stream silence).
//
//   ctrl.Paused = true
//
// And we can check how much has already been streamed. Position is not incremented when the Ctrl is
// paused.
//
//   fmt.Println(ctrl.Position)
//
// To completely stop a Ctrl before the wrapped Streamer is drained, just set the wrapped Streamer
// to nil.
//
//   ctrl.Streamer = nil
//
// If you're playing a Streamer wrapped in a Ctrl through the speaker, you need to lock and unlock
// the speaker when modifying the Ctrl to avoid race conditions.
//
//   speaker.Play(ctrl)
//   // ...
//   speaker.Lock()
//   ctrl.Paused = true
//   speaker.Unlock()
//   // ...
//   speaker.Lock()
//   fmt.Println(ctrl.Position)
//   speaker.Unlock()
type Ctrl struct {
	Streamer Streamer
	Paused   bool
	Position time.Duration
}

// Stream streams the wrapped Streamer, if not nil. If the Streamer is nil, Ctrl acts as drained.
// When paused, Ctrl streams silence.
func (c *Ctrl) Stream(samples [][2]float64) (n int, ok bool) {
	if c.Streamer == nil {
		return 0, false
	}
	if c.Paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}
	n, ok = c.Streamer.Stream(samples)
	c.Position += time.Duration(n) * time.Second / time.Duration(SampleRate)
	return n, ok
}

// Err returns the error of the wrapped Streamer, if not nil.
func (c *Ctrl) Err() error {
	if c.Streamer == nil {
		return nil
	}
	return c.Err()
}

package audio

import "time"

type Ctrl struct {
	Streamer Streamer
	Paused   bool
	Duration time.Duration
}

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
	c.Duration += time.Duration(n) * time.Second / time.Duration(SampleRate)
	return n, ok
}

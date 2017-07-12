package audio

// Gain amplifies the wrapped Streamer. The output of the wrapped Streamer gets multiplied by
// 1+Gain.
//
// Note that gain is not equivalent to the human perception of volume. Human perception of volume is
// roughly exponential, while gain only amplifies linearly.
type Gain struct {
	Streamer Streamer
	Gain     float64
}

// Stream streams the wrapped Streamer amplified by Gain.
func (g *Gain) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = g.Streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0] *= 1 + g.Gain
		samples[i][1] *= 1 + g.Gain
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (g *Gain) Err() error {
	return g.Streamer.Err()
}

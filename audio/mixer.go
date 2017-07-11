package audio

// Mixer allows for dynamic mixing of arbitrary number of Streamers. Mixer automatically removes
// drained Streamers. Mixer's stream never drains, when empty, Mixer streams silence.
type Mixer struct {
	streamers []Streamer
}

// Len returns the number of Streamers currently playing in the Mixer.
func (m *Mixer) Len() int {
	return len(m.streamers)
}

// Play adds Streamers to the Mixer.
func (m *Mixer) Play(s ...Streamer) {
	m.streamers = append(m.streamers, s...)
}

// Stream streams all Streamers currently in the Mixer mixed together. This method always returns
// len(samples), true. If there are no Streamers available, this methods streams silence.
func (m *Mixer) Stream(samples [][2]float64) (n int, ok bool) {
	var tmp [512][2]float64

	for len(samples) > 0 {
		toStream := len(tmp)
		if toStream > len(samples) {
			toStream = len(samples)
		}

		// clear the samples
		for i := range samples[:toStream] {
			samples[i] = [2]float64{}
		}

		for si := 0; si < len(m.streamers); si++ {
			// mix the stream
			sn, sok := m.streamers[si].Stream(tmp[:toStream])
			for i := range tmp[:sn] {
				samples[i][0] += tmp[i][0]
				samples[i][1] += tmp[i][1]
			}
			if !sok {
				// remove drained streamer
				sj := len(m.streamers) - 1
				m.streamers[si], m.streamers[sj] = m.streamers[sj], m.streamers[si]
				m.streamers = m.streamers[:sj]
				si--
			}
		}

		samples = samples[toStream:]
		n += toStream
	}

	return n, true
}

// Err always returns nil for Mixer.
//
// There are two reasons. The first one is that erroring Streamers are immediately drained and
// removed from the Mixer. The second one is that one Streamer shouldn't break the whole Mixer and
// you should handle the errors right where they can happen.
func (m *Mixer) Err() error {
	return nil
}

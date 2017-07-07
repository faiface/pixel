package audio

type Mixer struct {
	streamers []Streamer
}

func (m *Mixer) Play(s ...Streamer) {
	m.streamers = append(m.streamers, s...)
}

func (m *Mixer) Stream(samples [][2]float64) (n int, ok bool) {
	var tmp, mix [512][2]float64

	for len(samples) > 0 {
		toStream := len(mix)
		if toStream > len(samples) {
			toStream = len(samples)
		}

		// clear the mix buffer
		for i := range mix[:toStream] {
			mix[i] = [2]float64{}
		}

		for si := 0; si < len(m.streamers); si++ {
			// mix the stream
			sn, sok := m.streamers[si].Stream(tmp[:toStream])
			for i := range tmp[:sn] {
				mix[i][0] += tmp[i][0]
				mix[i][1] += tmp[i][1]
			}
			if !sok {
				// remove drained streamer
				sj := len(m.streamers) - 1
				m.streamers[si], m.streamers[sj] = m.streamers[sj], m.streamers[si]
				m.streamers = m.streamers[:sj]
				si--
			}
		}

		// copy mix buffer into samples
		for i := range mix[:toStream] {
			samples[i] = mix[i]
		}

		samples = samples[toStream:]
		n += toStream
	}

	return n, true
}

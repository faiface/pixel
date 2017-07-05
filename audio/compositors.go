package audio

import (
	"math"
	"time"
)

// Take returns a Streamer which streams s for at most d duration.
func Take(d time.Duration, s Streamer) Streamer {
	currSample := 0
	numSamples := int(math.Ceil(d.Seconds() * SampleRate))
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if currSample >= numSamples {
			return 0, false
		}
		toStream := numSamples - currSample
		if len(samples) < toStream {
			toStream = len(samples)
		}
		sn, sok := s.Stream(samples[:toStream])
		currSample += sn
		return sn, sok
	})
}

// Seq takes zero or more Streamers and returns a Streamer which streams them one by one without pauses.
func Seq(s ...Streamer) Streamer {
	i := 0
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i < len(s) && len(samples) > 0 {
			sn, sok := s[i].Stream(samples)
			samples = samples[sn:]
			n, ok = n+sn, ok || sok
			if !sok {
				i++
			}
		}
		return n, ok
	})
}

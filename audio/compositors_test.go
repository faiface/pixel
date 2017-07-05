package audio_test

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/faiface/pixel/audio"
)

// randomDataStreamer generates random samples of duration d and returns a Streamer which streams
// them and the data itself.
func randomDataStreamer(d time.Duration) (s audio.Streamer, data [][2]float64) {
	numSamples := int(math.Ceil(d.Seconds() * audio.SampleRate))
	data = make([][2]float64, numSamples)
	for i := range data {
		data[i][0] = rand.Float64()*2 - 1
		data[i][1] = rand.Float64()*2 - 1
	}
	return audio.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if len(data) == 0 {
			return 0, false
		}
		n = copy(samples, data)
		data = data[n:]
		return n, true
	}), data
}

// collect drains Streamer s and returns all of the samples it streamed.
func collect(s audio.Streamer) [][2]float64 {
	var (
		result [][2]float64
		buf    [512][2]float64
	)
	for {
		n, ok := s.Stream(buf[:])
		if !ok {
			return result
		}
		result = append(result, buf[:n]...)
	}
}

func TestSeq(t *testing.T) {
	var (
		s    = make([]audio.Streamer, 7)
		want [][2]float64
	)
	for i := range s {
		var data [][2]float64
		s[i], data = randomDataStreamer(time.Nanosecond * time.Duration(1e8+rand.Intn(1e9)))
		want = append(want, data...)
	}

	got := collect(audio.Seq(s...))

	if !reflect.DeepEqual(want, got) {
		t.Error("Seq not working correctly")
	}
}

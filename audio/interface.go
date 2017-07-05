package audio

// SampleRate is the number of audio samples a Streamer should produce per one second of audio.
//
// This value should be set at most once before using audio package. It is safe to assume that this
// value does not change during runtime.
var SampleRate float64 = 48000

// Streamer is able to stream a finite or infinite sequence of audio samples.
type Streamer interface {
	// Stream copies at most len(samples) next audio samples to the samples slice.
	//
	// The sample rate of the samples is specified by the global SampleRate variable/constant.
	// The value at samples[i][0] is the value of the left channel of the i-th sample.
	// Similarly, samples[i][1] is the value of the right channel of the i-th sample.
	//
	// Stream returns the number of streamed samples. If the Streamer is drained and no more
	// samples will be produced, it returns 0 and false. Stream must not touch any samples
	// outside samples[:n].
	//
	// There are 3 valid return pattterns of the Stream method:
	//
	//   1. n == len(samples) && ok
	//
	// Stream streamed all of the requested samples. Cases 1, 2 and 3 may occur in the following
	// calls.
	//
	//   2. 0 < n && n < len(samples) && ok
	//
	// Stream streamed n samples and drained the Streamer. Only case 3 may occur in the
	// following calls.
	//
	//   3. n == 0 && !ok
	//
	// The Streamer is drained and no more samples will come. Only this case may occur in the
	// following calls.
	Stream(samples [][2]float64) (n int, ok bool)
}

// StreamerFunc is a Streamer created by simply wrapping a streaming function (usually a closure,
// which encloses a time tracking variable). This sometimes simplifies creating new streamers.
//
// Example:
//
//   noise := StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
//       for i := range samples {
//           samples[i][0] = rand.Float64()*2 - 1
//           samples[i][1] = rand.Float64()*2 - 1
//       }
//       return len(samples), true
//   })
type StreamerFunc func(samples [][2]float64) (n int, ok bool)

// Stream calls the wrapped streaming function.
func (sf StreamerFunc) Stream(samples [][2]float64) (n int, ok bool) {
	return sf(samples)
}

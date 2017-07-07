package speaker

import (
	"math"
	"sync"
	"time"

	"github.com/faiface/pixel/audio"
	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
)

var (
	mu      sync.Mutex
	mixer   audio.Mixer
	samples [][2]float64
	buf     []byte
	player  *oto.Player
)

// Init initializes audio playback through speaker. Must be called before using this package. The
// value of audio.SampleRate must be set (or left to the default) before calling this function.
//
// The bufferSize argument specifies the length of the speaker's buffer. Bigger bufferSize means
// lower CPU usage and more reliable playback. Lower bufferSize means better responsiveness and less
// delay.
func Init(bufferSize time.Duration) error {
	mu.Lock()
	defer mu.Unlock()

	if player != nil {
		panic("already called Init")
	}

	numSamples := int(math.Ceil(bufferSize.Seconds() * audio.SampleRate))
	numBytes := numSamples * 4

	var err error
	player, err = oto.NewPlayer(int(audio.SampleRate), 2, 2, numBytes)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}

	samples = make([][2]float64, numSamples)
	buf = make([]byte, numBytes)

	go func() {
		for {
			update()
		}
	}()

	return nil
}

// Lock locks the speaker. While locked, speaker won't pull new data from the playing Stramers. Lock
// if you want to modify any currently playing Streamers to avoid race conditions.
func Lock() {
	mu.Lock()
}

// Unlock unlocks the speaker. Call after modifying any currently playing Streamer.
func Unlock() {
	mu.Unlock()
}

// Play starts playing all provided Streamers through the speaker.
func Play(s ...audio.Streamer) {
	mu.Lock()
	mixer.Play(s...)
	mu.Unlock()
}

// update pulls new data from the playing Streamers and sends it to the speaker. Blocks until the
// data is sent and started playing.
func update() {
	mu.Lock()
	mixer.Stream(samples)
	mu.Unlock()

	for i := range samples {
		for c := range samples[i] {
			val := samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16 % (1 << 8))
			high := byte(valInt16 / (1 << 8))
			buf[i*4+c*2+0] = low
			buf[i*4+c*2+1] = high
		}
	}

	player.Write(buf)
}

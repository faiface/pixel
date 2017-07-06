package playback

import (
	"errors"

	"io"

	"github.com/faiface/pixel/audio"
	"github.com/hajimehoshi/oto"
)

// Speaker is the interface used for playing back audio.Streamers.
type Speaker interface {
	// Play tells the Speaker that it is ready for playback and handles preparing the Streamer.
	Play(audio.Streamer)
	// Update is called once per game loop and handles pulling samples from the Streamer and writing them to Speaker's
	// player.
	Update() error
}

// DefaultSpeaker is a default implementation of speaker capable of playing back samples to the default output device.
type DefaultSpeaker struct {
	// audio.Streamer is the Streamer to pull samples from. It is passed in and set with Speaker.Play(audio.Streamer)
	audio.Streamer
	// isPlaying informs the update loop about whether or not this Speaker is playing
	isPlaying bool
	// samples is the internal buffer of samples that read() and readSample() fill and drain, respectively
	// samples' length is the total buffer size / 2
	samples [][2]float64
	// player is the underlying *oto.Player, which uses os specific APIs for audio playback
	player *oto.Player
	// buf is the buffer of samples converted to bytes that is written to player
	buf []uint8
	// bufferSize is the size in bytes of the total buffer in bytes. bufferSize must be a power of 2.
	bufferSize int
}

// NewDefaultSpeaker returns a *DefaultSpeaker ready to read samples and write to the underlying player for playback
func NewDefaultSpeaker(bufferSize int) (*DefaultSpeaker, error) {
	p, err := oto.NewPlayer(int(audio.SampleRate), 2, 2, bufferSize)
	if err != nil {
		return nil, err
	}
	return &DefaultSpeaker{
		player:     p,
		samples:    make([][2]float64, bufferSize/2),
		buf:        make([]uint8, bufferSize),
		bufferSize: bufferSize,
	}, nil
}

var (
	// ErrBufferMustBePowerOf2 should be returned when the buffer passed in to read is not a power of 2, as a sample is
	// 2 bytes, the buffer must have the capacity to handle all samples.
	ErrBufferMustBePowerOf2 = errors.New("Buffer passed to Read must be a power of 2")
)

// read reads up to len(dst) / 2 samples into dst
func (s *DefaultSpeaker) read(dst []byte) (n int, err error) {
	if !s.isPlaying || s.eof() {
		return 0, io.EOF
	}
	// we need dst to be a power of two in order for us to write samples cleanly
	if len(dst)%2 != 0 {
		return 0, ErrBufferMustBePowerOf2
	}

	if l := len(dst); l > 1 {
		for n < l-1 {
			sample := s.readSample()
			dst[n] = byte(sample[0])
			dst[n+1] = byte(sample[1])
			if s.eof() {
				s.samples = make([][2]float64, s.bufferSize/2)
				break
			}
			n += 2
		}
	}
	return n, nil
}

// eof returns whether or not we have read all samples currently in the samples buffer
func (s *DefaultSpeaker) eof() bool {
	return len(s.samples) == 0
}

// Sample is a single sample stored as an array of [2]float64, with Sample[0] being the left channel and Sample[1] being the right channel
type Sample [2]float64

// readSample reads a single sample from s.samples and truncates it from the buffer
func (s *DefaultSpeaker) readSample() Sample {
	sample := s.samples[0]
	s.samples = s.samples[1:]
	return sample
}

// Play initializes the Streamer and sets s.isPlaying to true
func (ds *DefaultSpeaker) Play(s audio.Streamer) {
	ds.isPlaying = true
	ds.Streamer = s
}

// streamToPlayer Streams up to len(s.samples) into s.samples, converts those into bytes for s.buf, and writes s.buf
// to the underlying player
func (s *DefaultSpeaker) streamToPlayer() error {
	n, ok := s.Stream(s.samples)
	if (n == len(s.samples) || 0 < n && n < len(s.samples)) && ok {
		r, err := s.read(s.buf)
		if err != nil {
			return err
		}
		s.buf = s.buf[:r]
		_, err = s.player.Write(s.buf)
		if err != nil {
			return err
		}
		// we drained the streamer while while reading,
		if n < len(s.samples) {
			s.isPlaying = false
			return nil
		}
		s.buf = make([]byte, s.bufferSize)
	}
	// this stream is already drained, set isPlaying to false
	if n == 0 && !ok {
		s.isPlaying = false
	}
	return nil
}

// Update should be called during the main update loop in order to handle synchronization
// If s.isPlaying, Update will stream all available samples to the underlying player once per update.
func (s *DefaultSpeaker) Update() error {
	if s.isPlaying {
		return s.streamToPlayer()
	}
	return nil
}

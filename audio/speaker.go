package audio

import (
	"errors"

	"io"

	"github.com/hajimehoshi/oto"
)

type Speaker interface {
	Play(Streamer)
	Update() error
}

type DefaultSpeaker struct {
	Streamer
	isPlaying  bool
	samples    [][2]float64
	player     *oto.Player
	buf        []uint8
	bufferSize int
}

var (
	ErrBufferMustBePowerOf2 = errors.New("Buffer passed to Read must be a power of 2")
)

func (s *DefaultSpeaker) Read(dst []byte) (n int, err error) {
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

func (s *DefaultSpeaker) eof() bool {
	return len(s.samples) == 0
}

type Sample [2]float64

func (s *DefaultSpeaker) readSample() Sample {
	sample := s.samples[0]
	s.samples = s.samples[1:]
	return sample
}

func NewDefaultSpeaker(bufferSize int) (*DefaultSpeaker, error) {
	p, err := oto.NewPlayer(int(SampleRate), 2, 2, bufferSize)
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

func (bs *DefaultSpeaker) Play(s Streamer) {
	bs.isPlaying = true
	bs.Streamer = s
}

func (b *DefaultSpeaker) Update() error {

	if b.isPlaying {
		n, ok := b.Stream(b.samples)
		if n == len(b.samples) && ok {
			r, err := b.Read(b.buf)
			if err != nil {
				return err
			}
			b.buf = b.buf[:r]
			_, err = b.player.Write(b.buf)
			if err != nil {
				return err
			}
			b.buf = make([]byte, b.bufferSize)
		}
		// we're read bytes but drained the streamer, so copy data and stop playing
		if n > 0 && n < len(b.samples) && ok {
			r, err := b.Read(b.buf)
			if err != nil {
				return err
			}
			b.buf = b.buf[:r]
			_, err = b.player.Write(b.buf)
			if err != nil {
				return err
			}
			b.isPlaying = false
		}
		// this stream is already drained, set isPlaying to false
		if n == 0 && !ok {
			b.isPlaying = false
		}
	}
	return nil
}

package wav

import (
	"io"
	"time"

	"github.com/pkg/errors"
)

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type Streamer struct {
	rsc ReadSeekCloser
	h   header
	pos int32
	err error
}

func NewStreamer(rsc ReadSeekCloser) (s *Streamer, err error) {
	var (
		d    Streamer
		herr error
	)
	d.rsc = rsc
	defer func() { // hacky way to always close rsc if an error occured
		if err != nil {
			d.rsc.Close()
		}
	}()
	d.h, herr = readHeader(rsc)
	if herr != nil {
		return nil, errors.Wrap(herr, "wav")
	}
	if d.h.formatType != 1 {
		return nil, errors.New("wav: unsupported format type")
	}
	if d.h.numChans <= 0 {
		return nil, errors.New("wav: invalid number of channels (less than 1)")
	}
	if d.h.bitsPerSample != 8 && d.h.bitsPerSample != 16 {
		return nil, errors.New("wav: unsupported number of bits per sample, 8 or 16 are supported")
	}
	return &d, nil
}

func (s *Streamer) Err() error {
	return s.err
}

func (s *Streamer) Duration() time.Duration {
	numBytes := time.Duration(s.h.dataSize)
	perFrame := time.Duration(s.h.bytesPerFrame)
	sampRate := time.Duration(s.h.sampleRate)
	return numBytes / perFrame * time.Second / sampRate
}

func (s *Streamer) Position() time.Duration {
	frameIndex := time.Duration(s.pos / int32(s.h.bytesPerFrame))
	frameTime := time.Second / time.Duration(s.h.sampleRate)
	return frameIndex * frameTime
}

func (s *Streamer) Seek(d time.Duration) {
	if d < 0 || s.Duration() < d {
		panic("wav: seek duration out of range")
	}
	frame := int32(d / (time.Second / time.Duration(s.h.sampleRate)))
	pos := frame * int32(s.h.bytesPerFrame)
	_, err := s.rsc.Seek(int64(pos), io.SeekStart)
	if err != nil {
		s.err = err
		return
	}
	s.pos = pos
}

func (s *Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.pos >= s.h.dataSize {
		return 0, false
	}
	switch {
	case s.h.bitsPerSample == 8 && s.h.numChans == 1:
		width := 1
		p := make([]byte, len(samples)*width)
		n, err := s.rsc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			val := float64(p[i])/(1<<8-1)*2 - 1
			samples[j][0] = val
			samples[j][1] = val
		}
		if err != nil {
			s.err = err
		}
		s.pos += int32(n)
		return n / width, true
	case s.h.bitsPerSample == 8 && s.h.numChans >= 2:
		width := int(s.h.numChans)
		p := make([]byte, len(samples)*width)
		n, err := s.rsc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			samples[j][0] = float64(p[i+0])/(1<<8-1)*2 - 1
			samples[j][1] = float64(p[i+1])/(1<<8-1)*2 - 1
		}
		if err != nil {
			s.err = err
		}
		s.pos += int32(n)
		return n / width, true
	case s.h.bitsPerSample == 16 && s.h.numChans == 1:
		width := 2
		p := make([]byte, len(samples)*width)
		n, err := s.rsc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			val := float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][0] = val
			samples[j][1] = val
		}
		if err != nil {
			s.err = err
		}
		s.pos += int32(n)
		return n / width, true
	case s.h.bitsPerSample == 16 && s.h.numChans >= 2:
		width := int(s.h.numChans) * 2
		p := make([]byte, len(samples)*width)
		n, err := s.rsc.Read(p)
		for i, j := 0, 0; i <= n-width; i, j = i+width, j+1 {
			samples[j][0] = float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][1] = float64(int16(p[i+2])+int16(p[i+3])*(1<<8)) / (1<<15 - 1)
		}
		if err != nil {
			s.err = err
		}
		s.pos += int32(n)
		return n / width, true
	}
	panic("unreachable")
}

func (s *Streamer) Close() error {
	err := s.rsc.Close()
	if err != nil {
		return errors.Wrap(err, "wav")
	}
	return nil
}

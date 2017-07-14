// Package wav implements audio data decoding in WAVE format through an audio.StreamSeekCloser.
package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/faiface/pixel/audio"
	"github.com/pkg/errors"
)

// ReadSeekCloser is a union of io.Reader, io.Seeker and io.Closer.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Decode takes a ReadSeekCloser containing audio data in WAVE format and returns a
// StreamSeekCloser, which streams that audio.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(rsc ReadSeekCloser) (s audio.StreamSeekCloser, err error) {
	d := decoder{rsc: rsc}
	defer func() { // hacky way to always close rsc if an error occured
		if err != nil {
			d.rsc.Close()
		}
	}()
	herr := binary.Read(rsc, binary.LittleEndian, &d.h)
	if herr != nil {
		return nil, errors.Wrap(herr, "wav")
	}
	if string(d.h.RiffMark[:]) != "RIFF" {
		return nil, errors.New("wav: missing RIFF at the beginning")
	}
	if string(d.h.WaveMark[:]) != "WAVE" {
		return nil, errors.New("wav: unsupported file type")
	}
	if string(d.h.FmtMark[:]) != "fmt " {
		return nil, errors.New("wav: missing format chunk marker")
	}
	if string(d.h.DataMark[:]) != "data" {
		return nil, errors.New("wav: missing data chunk marker")
	}
	if d.h.FormatType != 1 {
		return nil, errors.New("wav: unsupported format type")
	}
	if d.h.NumChans <= 0 {
		return nil, errors.New("wav: invalid number of channels (less than 1)")
	}
	if d.h.BitsPerSample != 8 && d.h.BitsPerSample != 16 {
		return nil, errors.New("wav: unsupported number of bits per sample, 8 or 16 are supported")
	}
	return &d, nil
}

type header struct {
	RiffMark      [4]byte
	FileSize      int32
	WaveMark      [4]byte
	FmtMark       [4]byte
	FormatSize    int32
	FormatType    int16
	NumChans      int16
	SampleRate    int32
	ByteRate      int32
	BytesPerFrame int16
	BitsPerSample int16
	DataMark      [4]byte
	DataSize      int32
}

type decoder struct {
	rsc ReadSeekCloser
	h   header
	pos int32
	err error
}

func (s *decoder) Err() error {
	return s.err
}

func (s *decoder) Duration() time.Duration {
	numBytes := time.Duration(s.h.DataSize)
	perFrame := time.Duration(s.h.BytesPerFrame)
	sampRate := time.Duration(s.h.SampleRate)
	return numBytes / perFrame * time.Second / sampRate
}

func (s *decoder) Position() time.Duration {
	frameIndex := time.Duration(s.pos / int32(s.h.BytesPerFrame))
	frameTime := time.Second / time.Duration(s.h.SampleRate)
	return frameIndex * frameTime
}

func (s *decoder) Seek(d time.Duration) error {
	if d < 0 || s.Duration() < d {
		return fmt.Errorf("wav: seek duration %v out of range [%v, %v]", d, 0, s.Duration())
	}
	frame := int32(d / (time.Second / time.Duration(s.h.SampleRate)))
	pos := frame * int32(s.h.BytesPerFrame)
	_, err := s.rsc.Seek(int64(pos)+44, io.SeekStart) // 44 is the size of the header
	if err != nil {
		return errors.Wrap(err, "wav: seek error")
	}
	s.pos = pos
	return nil
}

func (s *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if s.err != nil || s.pos >= s.h.DataSize {
		return 0, false
	}
	bytesPerFrame := int(s.h.BytesPerFrame)
	p := make([]byte, len(samples)*bytesPerFrame)
	n, err := s.rsc.Read(p)
	if err != nil {
		s.err = err
	}
	switch {
	case s.h.BitsPerSample == 8 && s.h.NumChans == 1:
		for i, j := 0, 0; i < n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			val := float64(p[i])/(1<<8-1)*2 - 1
			samples[j][0] = val
			samples[j][1] = val
		}
	case s.h.BitsPerSample == 8 && s.h.NumChans >= 2:
		for i, j := 0, 0; i < n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			samples[j][0] = float64(p[i+0])/(1<<8-1)*2 - 1
			samples[j][1] = float64(p[i+1])/(1<<8-1)*2 - 1
		}
	case s.h.BitsPerSample == 16 && s.h.NumChans == 1:
		for i, j := 0, 0; i < n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			val := float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][0] = val
			samples[j][1] = val
		}
	case s.h.BitsPerSample == 16 && s.h.NumChans >= 2:
		for i, j := 0, 0; i <= n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			samples[j][0] = float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][1] = float64(int16(p[i+2])+int16(p[i+3])*(1<<8)) / (1<<15 - 1)
		}
	}
	return n / bytesPerFrame, true
}

func (s *decoder) Close() error {
	err := s.rsc.Close()
	if err != nil {
		return errors.Wrap(err, "wav")
	}
	return nil
}

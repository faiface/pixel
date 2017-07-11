package wav

import (
	"io"
	"time"

	"github.com/pkg/errors"
)

type Streamer struct {
	rc  io.ReadCloser
	h   header
	pos int32
	err error
}

func NewStreamer(rc io.ReadCloser) (*Streamer, error) {
	var (
		d   Streamer
		err error
	)
	d.rc = rc
	d.h, err = readHeader(rc)
	if err != nil {
		rc.Close()
		return nil, errors.Wrap(err, "wav")
	}
	if d.h.formatType != 1 {
		rc.Close()
		return nil, errors.New("wav: unsupported format type")
	}
	if d.h.numChans <= 0 {
		rc.Close()
		return nil, errors.New("wav: invalid number of channels (less than 1)")
	}
	if d.h.bitsPerSample != 8 && d.h.bitsPerSample != 16 {
		rc.Close()
		return nil, errors.New("wav: unsupported number of bits per sample, 8 or 16 are supported")
	}
	return &d, nil
}

func (d *Streamer) Err() error {
	return d.err
}

func (d *Streamer) Duration() time.Duration {
	numBytes := time.Duration(d.h.dataSize)
	perFrame := time.Duration(d.h.bytesPerFrame)
	sampRate := time.Duration(d.h.sampleRate)
	return numBytes / perFrame * time.Second / sampRate
}

func (d *Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	if d.pos >= d.h.dataSize {
		return 0, false
	}
	switch {
	case d.h.bitsPerSample == 8 && d.h.numChans == 1:
		width := 1
		p := make([]byte, len(samples)*width)
		n, err := d.rc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			val := float64(p[i])/255*2 - 1
			samples[j][0] = val
			samples[j][1] = val
		}
		if err != nil {
			d.err = err
		}
		d.pos += int32(n)
		return n / width, true
	case d.h.bitsPerSample == 8 && d.h.numChans >= 2:
		width := int(d.h.numChans)
		p := make([]byte, len(samples)*width)
		n, err := d.rc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			samples[j][0] = float64(p[i+0])/255*2 - 1
			samples[j][1] = float64(p[i+1])/255*2 - 1
		}
		if err != nil {
			d.err = err
		}
		d.pos += int32(n)
		return n / width, true
	case d.h.bitsPerSample == 16 && d.h.numChans == 1:
		width := 2
		p := make([]byte, len(samples)*width)
		n, err := d.rc.Read(p)
		for i, j := 0, 0; i < n-width; i, j = i+width, j+1 {
			val := float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][0] = val
			samples[j][1] = val
		}
		if err != nil {
			d.err = err
		}
		d.pos += int32(n)
		return n / width, true
	case d.h.bitsPerSample == 16 && d.h.numChans >= 2:
		width := int(d.h.numChans) * 2
		p := make([]byte, len(samples)*width)
		n, err := d.rc.Read(p)
		for i, j := 0, 0; i <= n-width; i, j = i+width, j+1 {
			samples[j][0] = float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][1] = float64(int16(p[i+2])+int16(p[i+3])*(1<<8)) / (1<<15 - 1)
		}
		if err != nil {
			d.err = err
		}
		d.pos += int32(n)
		return n / width, true
	}
	panic("unreachable")
}

func (d *Streamer) Close() error {
	err := d.rc.Close()
	if err != nil {
		return errors.Wrap(err, "wav")
	}
	return nil
}

package audio

import (
	"errors"
	"io"
	"time"

	"fmt"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

type State int

const (
	StatePlaying State = iota
	StatePaused
	StateEnded
	StateNotStarted
)

var (
	ErrAlreadyPlaying = errors.New("This stream is already playing")
)

type MP3Player struct {
	Time      time.Duration
	State     State
	decoded   *mp3.Decoded
	stateChan chan State
	player    *oto.Player
	buffer    []byte
	pos       int64
}

func NewMP3Player(r io.ReadCloser) (*MP3Player, error) {
	decoded, err := mp3.Decode(r)
	if err != nil {
		return nil, err
	}

	player, err := oto.NewPlayer(decoded.SampleRate(), 2, 2, 65536)
	if err != nil {
		return nil, err
	}

	return &MP3Player{
		stateChan: make(chan State),
		player:    player,
		decoded:   decoded,
		State:     StateNotStarted,
	}, nil
}

func (m *MP3Player) Play() error {
	errChan := make(chan error)
	go func() {
		for {
			s := <-m.stateChan

			switch s {
			case StatePlaying:
				if m.State == StatePlaying {
					errChan <- ErrAlreadyPlaying
					return
				}

				m.State = StatePlaying
				go func() {
					for m.State == StatePlaying {
						br, err := m.Advance()
						if err != nil && err == io.EOF {
							m.stateChan <- StateEnded
						}

						dr := time.Duration(int64(m.decoded.SampleRate()) * br / 8)
						// with mp3's this will drift because frame size is variable, but it's pretty close
						m.Time += dr
						m.pos += br
					}
				}()

			case StateEnded:
				errChan <- io.EOF
				return
			case StatePaused:
				m.State = StatePaused
			}
		}
	}()

	m.stateChan <- StatePlaying
	return <-errChan
}

func (m *MP3Player) Start() {
	m.stateChan <- StatePlaying
}

func (m *MP3Player) Pause() error {
	fmt.Println(m.Time)
	fmt.Println(m.pos)
	m.stateChan <- StatePaused
	return nil
}

func (m *MP3Player) Stop() error {
	m.stateChan <- StateEnded
	return m.player.Close()
}

func (m *MP3Player) Seek(f float64, whence int) error {
	_, err := m.decoded.Seek(0, io.SeekStart)
	return err
}

func (m *MP3Player) Advance() (n int64, err error) {

	if m.buffer == nil {
		m.buffer = make([]byte, 32*1024)
	}
	nr, er := m.decoded.Read(m.buffer)
	if nr > 0 {
		nw, ew := m.player.Write(m.buffer[0:nr])
		if nw > 0 {
			n += int64(nw)
		}
		if ew != nil {
			err = ew
		}
		if nr != nw {
			err = errors.New("Short write")
		}
	}
	if er != nil {
		if er != io.EOF {
			err = er
		}
	}
	return
}

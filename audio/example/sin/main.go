package main

import (
	"math"

	"fmt"

	"os"

	"github.com/faiface/pixel/audio"
)

type sine struct {
	freq float64
	rate float64
	time float64
}

func (s *sine) Stream(samples [][2]float64) (n int, ok bool) {
	if len(samples) == 0 {
		os.Exit(-1)
	}
	fmt.Println(len(samples))
	for i := 0; i < len(samples)-2; i += 2 {
		val := math.Sin(math.Pi*s.time*s.freq) / 1.1
		s.time += 1 / s.rate
		valI := int16((1 << 15) * val)
		low := float64(valI % (1 << 8))
		high := float64(valI / (1 << 8))
		samples[i][0] = low
		samples[i][1] = high
		samples[i+1][0] = low
		samples[i+1][1] = high
	}
	fmt.Println(samples)
	return len(samples), true
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
	}
}

func run() error {
	audio.SampleRate = 44100
	const bufSize = 1 << 13

	speaker, err := audio.NewDefaultSpeaker(bufSize)
	if err != nil {
		return err
	}

	s := &sine{freq: 440, rate: audio.SampleRate, time: 0}

	speaker.Play(s)

	for {
		err := speaker.Update()
		if err != nil {
			return err
		}
	}

	return nil
}

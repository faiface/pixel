package main

import (
	"os"

	"fmt"

	"bufio"

	"github.com/faiface/pixel/audio"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	f, err := os.Open("Keyscape.mp3")
	if err != nil {
		return err
	}

	mp3player, err := audio.NewMP3Player(f)
	if err != nil {
		return err
	}
	exitChan := make(chan error)

	go func() {
		exitChan <- mp3player.Play()
	}()

	go func() {
		scan := bufio.NewScanner(os.Stdin)
		for {
			scan.Scan()
			switch scan.Text() {
			case "s":
				mp3player.Stop()
			case "p":
				mp3player.Pause()
			case "S":
				mp3player.Start()
			}
		}
	}()

	return <-exitChan
}

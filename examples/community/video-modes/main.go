package main

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	texts         []*text.Text
	staticText    *text.Text
	vmodes        []*pixelgl.VideoMode
	activeMode    pixelgl.VideoMode
	activeMonitor *pixelgl.Monitor
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Video Modes",
		Bounds: pixel.R(0, 0, 800, 600),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Retrieve all monitors.
	monitors := pixelgl.Monitors()

	texts = make([]*text.Text, len(monitors))
	key := byte('0')
	for i := 0; i < len(monitors); i++ {
		// Retrieve all video modes for a specific monitor.
		modes := monitors[i].VideoModes()

		vmodes = append(vmodes, modes...)
		texts[i] = text.New(pixel.V(10+250*float64(i), -20), atlas)
		texts[i].Color = colornames.Black
		texts[i].WriteString(fmt.Sprintf("MONITOR %s\n\n", monitors[i].Name()))

		for _, v := range modes {
			texts[i].WriteString(fmt.Sprintf("(%c) %dx%d @ %d hz, %d bpp\n", key, v.Width, v.Height, v.RefreshRate, v.BlueBits+v.RedBits+v.GreenBits))
			key++
		}
	}

	staticText = text.New(pixel.V(10, 30), atlas)
	staticText.Color = colornames.Black
	staticText.WriteString("ESC to exit\nW toggles windowed/fullscreen")

	for !win.Closed() {
		win.Clear(colornames.Antiquewhite)

		for _, txt := range texts {
			txt.Draw(win, pixel.IM.Moved(pixel.V(0, win.Bounds().H())))
		}
		staticText.Draw(win, pixel.IM)

		if win.JustPressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
		}

		if win.JustPressed(pixelgl.KeyW) {
			if activeMode.Monitor != nil {
				// Switch to windowed and backup the correct monitor.
				activeMonitor = activeMode.Monitor
				activeMode.Monitor = nil
				win.SetVideoMode(activeMode)
			} else {
				// Switch to fullscreen.
				activeMode.Monitor = activeMonitor
				win.SetVideoMode(activeMode)
			}
		}

		input := win.Typed()
		if len(input) > 0 {
			key := int(input[0]) - 48
			fmt.Println(key)
			if key >= 0 && key < len(vmodes) {
				activeMode = *vmodes[key]
				activeMonitor = activeMode.Monitor
				fmt.Println("change to:", activeMode.Width, activeMode.Height)
				win.SetVideoMode(activeMode)
			}
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

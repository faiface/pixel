package music

import (
	"bytes"
	"io"
	"log"
	"path/filepath"
	"time"

	"github.com/gandrin/ASharedJourney/assets_manager"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

//global synchronization channel
var GameMusicLoader chan int

func init() {
	GameMusicLoader = make(chan int, 0)
}

const musicThemeFileName = "MainThemeMiroir.mp3"

type musicStreamers struct {
	//list of loaded musics ( streamer )
	mainThemeStreamer beep.Streamer
	backgroundMusic   *beep.Buffer
	gameEffects       map[SoundEffect]*beep.Buffer
	streamControl     beep.Ctrl
}

var Music musicStreamers

//called when package is called into scope the first time
func (m *musicStreamers) Start() {
	var format beep.Format

	go func() {
		format = m.loadMainTheme()

		err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			log.Fatal(err)
		}
		m.streamControl.Streamer = m.mainThemeStreamer
		m.playMainTheme()
	}()

	go m.loadEffects()

	log.Print("Music loaded")
}

func (m *musicStreamers) loadMainTheme() beep.Format {
	var format beep.Format
	m.mainThemeStreamer, format = getStream(musicThemeFileName)
	m.backgroundMusic = beep.NewBuffer(format)
	m.backgroundMusic.Append(m.mainThemeStreamer)
	m.streamControl.Paused = false
	return format
}

func (m *musicStreamers) playMainTheme() {

	log.Print("Starting music")
	var streamer = m.backgroundMusic.Streamer(0, m.backgroundMusic.Len())
	LoopAudio := beep.Loop(5, streamer)
	go speaker.Play(beep.Seq(LoopAudio))

	log.Print("Music finished")

	GameMusicLoader <- 1

}

func getFileName(fileName string) string {
	return "assets/" + fileName
}

func getStream(filename string) (beep.StreamCloser, beep.Format) {

	absFilePath := getFileName(filename)
	var newStreamer beep.StreamCloser
	var format beep.Format
	var err error

	byteSound, err := assetsManager.Asset(absFilePath)
	if err != nil {
		log.Fatal("Music file  ", err)
	}
	ext := filepath.Ext(filename)
	nopCloser := nopCloser{bytes.NewReader(byteSound)}
	if ext == ".mp3" {
		newStreamer, format, err = mp3.Decode(nopCloser)
	} else if ext == ".wav" {
		newStreamer, format, err = wav.Decode(nopCloser)
	}

	if err != nil {
		log.Fatal("Decorer error on file ", absFilePath, " ", err)
	}
	return newStreamer, format
}

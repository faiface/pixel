package jukebox

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
)

// -------------------------------------------------------------------------

const nMusics = 2

var (
	mutex     sync.Mutex
	isPlaying bool
	musics    [nMusics]*_Music
)

// -------------------------------------------------------------------------

// singleton
func init() {
	// city pop favorites
	musics[0] = _NewMusicFromAsset("nighttempo-purepresent1", "karaoke/kikuchimomoko-nightcruising.ogg")
	musics[1] = _NewMusicFromAsset("nighttempo-purepresent2", "karaoke/takeuchimariya-plasticlove.ogg")

	// speaker on
	speaker.Init(musics[0].format.SampleRate, musics[0].format.SampleRate.N(time.Second))
	speaker.Play(beep.Iterate(func() (soundtrack beep.Streamer) {
		musics[0].stream.Seek(0)
		musics[1].stream.Seek(0)
		return beep.Seq(musics[0].stream, musics[1].stream)
	}))
	speaker.Lock()
}

// IsPlaying determines whether the soundtrack is currently playing or not.
func IsPlaying() bool {
	return isPlaying
}

// Play unlocks the speaker.
func Play() {
	mutex.Lock()
	defer mutex.Unlock()
	if !isPlaying {
		isPlaying = true
		speaker.Unlock()
	}
	return
}

// Pause locks the speaker.
func Pause() {
	mutex.Lock()
	defer mutex.Unlock()
	if isPlaying {
		isPlaying = false
		speaker.Lock()
	}
	return
}

// Finalize should be called on program exit.
// This function deletes the temporary music file its package generates.
func Finalize() error {
	errs := ""
	for _, music := range musics {
		music.Close()
		err := music._Destory()
		if err != nil {
			errs += " " + err.Error()
		}
	}
	if errs != "" {
		return errors.New(errs)
	}
	return nil
}

// -------------------------------------------------------------------------

// NewMusicFromAsset is a constructor.
func _NewMusicFromAsset(nameMusic, nameAsset string) *_Music {
	asset, err := gg.Asset(nameAsset)
	if err != nil {
		// log.Fatal(err) //
	}
	return _NewMusic(nameMusic, asset)
}

// Music is a temporary file to play a single background music. It should be destroyed on program exit.
type _Music struct {
	os.File
	stream beep.StreamSeekCloser
	format beep.Format
}

// NewMusic creates an instance of Music, a temporary file from which the speaker plays a music.
// speaker.Lock() to pause.
// speaker.Unlock() to resume/play.
func _NewMusic(name string, asset []byte) *_Music {
	tmpfile, err := ioutil.TempFile("", name)
	if err != nil {
		// log.Fatal(err) //
	}
	// log.Println(tmpfile.Name()) //
	_, err = tmpfile.Write(asset)
	if err != nil {
		// log.Fatal(err) //
	}
	stream, format, err := vorbis.Decode(tmpfile)
	if err != nil {
		// log.Fatal(err) //
	}
	return &_Music{*tmpfile, stream, format}
}

// Destory deletes the temporary music file.
func (music *_Music) _Destory() error {
	return os.Remove(music.Name())
}

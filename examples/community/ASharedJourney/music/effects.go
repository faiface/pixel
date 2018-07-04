package music

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

type SoundEffect string

const (
	SOUND_EFFECT_START_GAME SoundEffect = "MenuEffect.wav"

	SOUND_EFFECT_WIN_GAME SoundEffect = "win2.mp3"

	SOUND_EFFECT_WIN_FINAL_GAME SoundEffect = "win.mp3"

	SOUND_EFFECT_SNORE SoundEffect = "snoring.mp3"

	SOUND_EFFECT_LOSE_GAME SoundEffect = "lose2.mp3"

	SOUND_EFFECT_WATER SoundEffect = "Acid_Bubble.mp3"

	SOUND_NONE SoundEffect = "Acid_Bubble.mp3"
)

func (m *musicStreamers) PlayEffect(effectType SoundEffect) {

	es, ok := m.gameEffects[effectType]
	if ok {
		speaker.Lock()
		m.streamControl.Paused = true
		speaker.Unlock()
		//log.Print("Creating new stream entry")
		LoopAudio := beep.Loop(1, es.Streamer(0, es.Len()))
		speaker.Play(beep.Seq(LoopAudio)) //effect exists -> play
		//log.Print("finished playing effect ")
		speaker.Lock()
		m.streamControl.Paused = false
		speaker.Unlock()
		//log.Print("stream of sound finished")
	}

}

func (m *musicStreamers) loadEffects() {
	m.gameEffects = make(map[SoundEffect]*beep.Buffer, 0)
	//make new buffers and add to buffer
	stream1, format1 := getStream(string(SOUND_EFFECT_START_GAME))
	m.gameEffects[SOUND_EFFECT_START_GAME] = beep.NewBuffer(format1)
	m.gameEffects[SOUND_EFFECT_START_GAME].Append(stream1)

	stream2, format2 := getStream(string(SOUND_EFFECT_WIN_GAME))
	m.gameEffects[SOUND_EFFECT_WIN_GAME] = beep.NewBuffer(format2)
	m.gameEffects[SOUND_EFFECT_WIN_GAME].Append(stream2)

	stream3, format3 := getStream(string(SOUND_EFFECT_WATER))
	m.gameEffects[SOUND_EFFECT_WATER] = beep.NewBuffer(format3)
	m.gameEffects[SOUND_EFFECT_WATER].Append(stream3)

	stream4, format4 := getStream(string(SOUND_EFFECT_WIN_FINAL_GAME))
	m.gameEffects[SOUND_EFFECT_WIN_FINAL_GAME] = beep.NewBuffer(format4)
	m.gameEffects[SOUND_EFFECT_WIN_FINAL_GAME].Append(stream4)

	stream5, format5 := getStream(string(SOUND_EFFECT_SNORE))
	m.gameEffects[SOUND_EFFECT_SNORE] = beep.NewBuffer(format5)
	m.gameEffects[SOUND_EFFECT_SNORE].Append(stream5)
}

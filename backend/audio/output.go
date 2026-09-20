package audio

import (
	"errors"

	beep "github.com/gopxl/beep/v2"
)

type audioOutput interface {
	Init(sampleRate beep.SampleRate, bufferSize int) error
	Play(streamer beep.Streamer) error
	Close()
	Err() error
}

var errOutputNotInitialized = errors.New("audio output is not initialized")

const sampleRate48000 = beep.SampleRate(48000)

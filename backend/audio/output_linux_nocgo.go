//go:build linux && !cgo

package audio

import (
	"errors"

	beep "github.com/gopxl/beep/v2"
)

var errPulseAudioRequiresCGO = errors.New("Linux PulseAudio output requires CGO")

type unavailableAudioOutput struct{}

func newAudioOutput() audioOutput {
	return &unavailableAudioOutput{}
}

func (o *unavailableAudioOutput) Init(beep.SampleRate, int) error {
	return errPulseAudioRequiresCGO
}

func (o *unavailableAudioOutput) Play(beep.Streamer) error {
	return errOutputNotInitialized
}

func (o *unavailableAudioOutput) Close() {}

func (o *unavailableAudioOutput) Err() error {
	return nil
}

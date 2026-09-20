//go:build !linux

package audio

import (
	"sync"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

type nativeAudioOutput struct {
	mutex       sync.Mutex
	initialized bool
}

func newAudioOutput() audioOutput {
	return &nativeAudioOutput{}
}

func (o *nativeAudioOutput) Init(sampleRate beep.SampleRate, bufferSize int) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.initialized {
		return nil
	}
	if err := speaker.Init(sampleRate, bufferSize); err != nil {
		return err
	}
	o.initialized = true
	return nil
}

func (o *nativeAudioOutput) Play(streamer beep.Streamer) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if !o.initialized {
		return errOutputNotInitialized
	}
	speaker.Play(streamer)
	return nil
}

func (o *nativeAudioOutput) Close() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	speaker.Close()
	o.initialized = false
}

func (o *nativeAudioOutput) Err() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return speaker.Err()
}

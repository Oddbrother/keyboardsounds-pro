package audio

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// Audio represents an audio file.
type Audio struct {
	buffer *beep.Buffer
}

// NewAudio creates a new audio file from a given format and file.
func NewAudio(formatType AudioFormat, file io.ReadCloser) (*Audio, error) {
	var (
		streamer beep.StreamSeekCloser
		format   beep.Format
		err      error
	)

	switch formatType {
	case MP3:
		streamer, format, err = mp3.Decode(file)
	case WAV:
		streamer, format, err = wav.Decode(file)
	default:
		return nil, fmt.Errorf("invalid audio format: %s", formatType)
	}

	if err != nil {
		return nil, err
	}

	defer streamer.Close()

	var finalStreamer beep.Streamer = streamer
	if format.SampleRate != sampleRate {
		finalStreamer = beep.Resample(4, format.SampleRate, sampleRate, streamer)
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(finalStreamer)

	return &Audio{buffer}, nil
}

// AudioPlayer is an interface for playing audio files.
type AudioPlayer interface {
	// Play plays the audio file for the given audio. It should be non-blocking, and can potentially return an error before
	// triggering the asynchronous playback.
	Play(audio *Audio, effects EffectsConfig) error
}

// Buffer duration for the speaker. Beep splits this between the audio driver
// and its player, so 20 ms provides roughly 10 ms of buffering at each layer.
const bufferDuration = 20 * time.Millisecond

// maxConcurrentVoices bounds the amount of work the mixer can perform when
// input events arrive faster than the audio device can consume them.
const maxConcurrentVoices = 32

var (
	audioPlayer     AudioPlayer
	audioPlayerOnce sync.Once
)

type audioPlayerImpl struct {
	initialized  bool
	initMutex    sync.Mutex
	playMutex    sync.Mutex
	activeVoices atomic.Int32
}

// GetAudioPlayer retrieves the audio player instance
func GetAudioPlayer() AudioPlayer {
	audioPlayerOnce.Do(func() {
		player := &audioPlayerImpl{}
		audioPlayer = player
		go player.monitorHealth()
	})

	return audioPlayer
}

func (a *audioPlayerImpl) monitorHealth() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.playMutex.Lock()
		a.initMutex.Lock()
		if a.initialized {
			if err := speaker.Err(); err != nil {
				slog.Error("audio backend stopped; restarting speaker", "error", err)
				speaker.Close()
				a.initialized = false
			}
		}
		a.initMutex.Unlock()
		a.playMutex.Unlock()
	}
}

// ensureInitialized initializes the speaker if it hasn't been initialized yet.
// This is thread-safe and will only initialize once.
func (a *audioPlayerImpl) ensureInitialized() error {
	a.initMutex.Lock()
	defer a.initMutex.Unlock()

	if a.initialized {
		if err := speaker.Err(); err == nil {
			return nil
		} else {
			slog.Error("audio backend stopped; restarting speaker", "error", err)
			speaker.Close()
			a.initialized = false
		}
	}

	err := speaker.Init(sampleRate, sampleRate.N(bufferDuration))
	if err != nil {
		return err
	}

	a.initialized = true
	return nil
}

func (a *audioPlayerImpl) Play(audio *Audio, effects EffectsConfig) error {
	a.playMutex.Lock()
	defer a.playMutex.Unlock()

	if !a.activeVoices.CompareAndSwap(0, 1) {
		for {
			active := a.activeVoices.Load()
			if active >= maxConcurrentVoices {
				slog.Warn("dropping audio playback because the voice limit was reached", "activeVoices", active)
				return nil
			}
			if a.activeVoices.CompareAndSwap(active, active+1) {
				break
			}
		}
	}

	// Ensure speaker is initialized (thread-safe, only initializes once)
	if err := a.ensureInitialized(); err != nil {
		a.activeVoices.Add(-1)
		return err
	}

	var streamer beep.Streamer = audio.buffer.Streamer(0, audio.buffer.Len())

	// Apply effects to streamer.
	for _, effect := range registeredEffects {
		streamer = effect.Apply(effects, streamer)
	}

	// Release the voice slot when the stream has been fully consumed.
	streamer = beep.Seq(streamer, beep.Callback(func() {
		a.activeVoices.Add(-1)
	}))

	// Play the audio - speaker.Play is non-blocking and supports simultaneous playback
	speaker.Play(streamer)

	return nil
}

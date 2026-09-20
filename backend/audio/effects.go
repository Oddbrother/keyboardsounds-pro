package audio

import (
	beep "github.com/gopxl/beep/v2"
)

// Common sample rate for all audio playback.
// Match the 48 kHz rate used by the system's PipeWire output to avoid resampling.
const sampleRate = beep.SampleRate(48000)

var registeredEffects = make([]Effect, 0)

func registerEffect(name string, effect Effect) {
	registeredEffects = append(registeredEffects, effect)
}

// Effect is an interface for an audio effect.
type Effect interface {
	// Apply applies the effect to the given streamer.
	Apply(config EffectsConfig, streamer beep.Streamer) beep.Streamer
}

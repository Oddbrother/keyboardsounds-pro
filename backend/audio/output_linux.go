//go:build linux

package audio

/*
#cgo LDFLAGS: -l:libpulse-simple.so.0 -l:libpulse.so.0
#include <stdint.h>
#include <stdlib.h>

typedef struct pa_simple pa_simple;
typedef struct {
	uint32_t format;
	uint32_t rate;
	uint8_t channels;
} pa_sample_spec;

extern pa_simple* pa_simple_new(const char*, const char*, int, const char*, const char*, const pa_sample_spec*, const void*, const void*, int*);
extern int pa_simple_write(pa_simple*, const void*, size_t, int*);
extern void pa_simple_free(pa_simple*);
extern const char* pa_strerror(int);

static pa_simple* pulse_simple_new(const pa_sample_spec* spec, int* error) {
	return pa_simple_new(NULL, "keyboardsounds-pro", 1, NULL, "keyboard sounds", spec, NULL, NULL, error);
}
*/
import "C"

import (
	"fmt"
	"math"
	"sync"
	"time"
	"unsafe"

	beep "github.com/gopxl/beep/v2"
)

const pulseSampleFloat32LE = 5

type pulseAudioOutput struct {
	mutex       sync.Mutex
	streamers   []beep.Streamer
	sampleRate  beep.SampleRate
	bufferSize  int
	client      *C.pa_simple
	done        chan struct{}
	finished    chan struct{}
	initialized bool
	closed      bool
	err         error
}

func newAudioOutput() audioOutput {
	return &pulseAudioOutput{}
}

func (o *pulseAudioOutput) Init(sampleRate beep.SampleRate, bufferSize int) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.initialized {
		return nil
	}
	if sampleRate != sampleRate48000 {
		return fmt.Errorf("PulseAudio output requires 48 kHz, got %d Hz", sampleRate)
	}
	spec := C.pa_sample_spec{
		format:   C.uint32_t(pulseSampleFloat32LE),
		rate:     C.uint32_t(sampleRate),
		channels: 2,
	}
	var pulseErr C.int
	client := C.pulse_simple_new(&spec, &pulseErr)
	if client == nil {
		return fmt.Errorf("initialize PulseAudio output: %s", C.GoString(C.pa_strerror(pulseErr)))
	}
	o.sampleRate = sampleRate
	o.bufferSize = bufferSize
	o.client = client
	o.done = make(chan struct{})
	o.finished = make(chan struct{})
	o.closed = false
	o.err = nil
	o.initialized = true
	go o.run()
	return nil
}

func (o *pulseAudioOutput) Play(streamer beep.Streamer) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if !o.initialized || o.closed {
		return errOutputNotInitialized
	}
	if o.err != nil {
		return o.err
	}
	o.streamers = append(o.streamers, streamer)
	return nil
}

func (o *pulseAudioOutput) Close() {
	o.mutex.Lock()
	if !o.initialized {
		o.mutex.Unlock()
		return
	}
	o.closed = true
	close(o.done)
	finished := o.finished
	o.mutex.Unlock()

	<-finished

	o.mutex.Lock()
	o.streamers = nil
	o.client = nil
	o.initialized = false
	o.mutex.Unlock()
}

func (o *pulseAudioOutput) Err() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.err
}

func (o *pulseAudioOutput) run() {
	defer func() {
		o.mutex.Lock()
		client := o.client
		o.mutex.Unlock()
		if client != nil {
			C.pa_simple_free(client)
		}
		close(o.finished)
	}()
	ticker := time.NewTicker(time.Duration(o.bufferSize) * time.Second / time.Duration(o.sampleRate))
	defer ticker.Stop()

	for {
		select {
		case <-o.done:
			return
		case <-ticker.C:
			if err := o.writeBuffer(); err != nil {
				o.mutex.Lock()
				o.err = err
				o.mutex.Unlock()
				return
			}
		}
	}
}

func (o *pulseAudioOutput) writeBuffer() error {
	o.mutex.Lock()
	if o.closed || o.err != nil || len(o.streamers) == 0 {
		o.mutex.Unlock()
		return nil
	}
	client := o.client
	bufferSize := o.bufferSize

	mixed := make([][2]float64, bufferSize)
	remaining := o.streamers[:0]
	for _, streamer := range o.streamers {
		samples := make([][2]float64, bufferSize)
		n, ok := streamer.Stream(samples)
		for i := 0; i < n; i++ {
			mixed[i][0] += samples[i][0]
			mixed[i][1] += samples[i][1]
		}
		if ok || n > 0 {
			remaining = append(remaining, streamer)
		}
	}
	if !o.closed {
		o.streamers = remaining
	}
	o.mutex.Unlock()
	if len(remaining) == 0 {
		return nil
	}

	pcm := make([]float32, bufferSize*2)
	for i := range mixed {
		pcm[i*2] = float32(math.Max(-1, math.Min(1, mixed[i][0])))
		pcm[i*2+1] = float32(math.Max(-1, math.Min(1, mixed[i][1])))
	}
	var pulseErr C.int
	if C.pa_simple_write(client, unsafe.Pointer(&pcm[0]), C.size_t(len(pcm)*4), &pulseErr) < 0 {
		return fmt.Errorf("write PulseAudio stream: %s", C.GoString(C.pa_strerror(pulseErr)))
	}
	return nil
}

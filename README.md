# Keyboard Sounds Pro — Linux fork

This fork is based on [keyboard-sounds/keyboardsounds-pro](https://github.com/keyboard-sounds/keyboardsounds-pro).

## Changes in this fork

- Persist keyboard and mouse default profiles on Linux in `~/keyboardsounds-pro/rules.json`.
- Load saved Linux default profiles when the application starts.
- Reduce the default audio buffer to 20 ms for responsive input playback.
- Bound concurrent audio voices to prevent unbounded mixer growth during heavy input.
- Remove redundant mouse-event goroutines.
- Add audio health monitoring and recovery for transient playback failures.
- Add ALSA/PipeWire write recovery so audio playback can continue after recoverable I/O errors.
- Include local Beep and Oto patches required by the Linux audio recovery behavior.
- Add Linux profile persistence coverage and documentation.

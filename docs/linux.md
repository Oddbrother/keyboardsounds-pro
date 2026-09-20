# Keyboard Sounds Pro - Linux Support

[Download for Linux](https://github.com/keyboard-sounds/keyboardsounds-pro/releases/latest)

On Linux, in order for the application to detect your input devices your user must be added to the `input` group through `usermod`.

```
sudo usermod -aG input $USER
```

After adding your user to the `input` group, you may need to reboot your system for the changes to take effect.

Keyboard and mouse default profiles are saved in `~/keyboardsounds-pro/rules.json` and restored when the application starts.

Linux audio uses the PulseAudio simple client API (including through PipeWire's
PulseAudio compatibility server). Install the PulseAudio runtime library before
running the application:

```
sudo apt install libpulse0
```

Building from source also requires the development package:

```
sudo apt install libpulse-dev
```

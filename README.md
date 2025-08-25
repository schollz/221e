<p align="center">
<a href="https://www.youtube.com/watch?v=zViMACW6VbQ">
<img width="600" alt="vlcsnap-2025-08-23-18h24m04s244" src="https://github.com/user-attachments/assets/7d4c36c0-bd28-4611-a41b-ddf864af045c" />
</a>
<br>
<a href="https://github.com/schollz/2n/releases/latest">
<img src="https://img.shields.io/github/v/release/schollz/2n" alt="Version">
</a>
<a href="https://github.com/schollz/2n/actions/workflows/build.yml">
<img src="https://github.com/schollz/2n/actions/workflows/build.yml/badge.svg" alt="Build Status">
</a>
<a href="https://github.com/sponsors/schollz">
<img alt="GitHub Sponsors" src="https://img.shields.io/github/sponsors/schollz">
</a>
</p>

A terminal-based music tracker powered by SuperCollider + JACK.

_IMPORTANT NOTE: this software is currently in development and is definetly unstable and chock full of bugs._

This program is heavily inspired by my norns tracker, [zxcvbn](https://zxcvbn.norns.online/) and the [dirtywave m8](https://dirtywave.com/) (which itself is inspired by countless trackers before it). While it may look similar, this is not **m8**! The **m8** is incredible, much better than this poc. This thing, **2n**, is based heavily on my own SuperCollider scripts I've written over the years, using an interface modeled after the **m8**, created in a cross-platform terminal user interface.

## Download

- 


## Prerequisites

- **SuperCollider** (required; extensions are checked at launch). Download [here](https://supercollider.github.io/downloads.html). 
- **JACK (jackd)** must be running with the output to your favorite speaker. Download [here](https://jackaudio.org/downloads/).
- **2n** binary. Grab the latest build from **[Releases](https://github.com/schollz/2n/releases/latest)**.
---

## Run

(After you have started Jack...)
```bash
./2n
```

Defaults: OSC **57120**, save file **tracker-save.json**.

## Keyboard — Quick Reference

### Shift navigation (views)
- **Shift+Right**
  - Song → Chain (selected track/row)
  - Chain → Phrase (selected row)
  - Phrase → Retrigger (if on **RT** and set) / Timestretch (if on **TS** and set) / File Browser (otherwise)
- **Shift+Up**
  - Song / Chain / Phrase → **Settings**
  - File Browser → **File Metadata** (for selected file)
- **Shift+Down**
  - Song / Chain / Phrase → **Mixer**
  - Mixer → back to previous view

### Editing / movement
- **Ctrl+Right / Ctrl+Left** – fine adjust values
- **Ctrl+Up / Ctrl+Down** – coarse adjust values

### Copy & paste
- **Ctrl+C** copy • **Ctrl+X** cut • **Ctrl+V** paste • **Ctrl+D** deep copy

### Playback
- **Space** play/stop from current row
- **Ctrl+@** play from top

### Misc
- **Esc** clear selection highlight
- **Ctrl+Q** quit


## Views

- **Song** – 8 tracks × 16 rows (chains per track)
- **Chain** – 16 rows mapping to phrases
- **Phrase** – tracker grid (notes & effects)
- **File Browser** – pick audio files per row
- **File Metadata** – per-file BPM / slices
- **Retrigger** – retrigger envelope
- **Timestretch** – stretch window settings
- **Mixer** – track levels/volumes
- **Settings** – BPM, PPQ, gains, etc.


## Phrase Columns

```
SL  P  NN  DT  GT  RT  TS  Я  PA  LP  HP  CO  VE  FI
```

- **SL** (slice) • **P** (play 0/1) • **NN** (note)
- **DT** (delta) • **GT** (gate) • **RT** (retrigger) • **TS** (timestretch)
- **Я** (reverse)
- **PA** (pan)
- **LP** / **HP** (filters)
- **CO** (comb) • **VE** (reverb) • **FI** (file index)


## License

MIT
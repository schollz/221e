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


## Prerequisites

- **SuperCollider** (required; extensions are checked at launch). Download [here](https://supercollider.github.io/downloads.html). 
- **JACK (jackd)** must be running with the output to your favorite speaker. Download [here](https://jackaudio.org/downloads/).
- **2n** binary. Grab the latest build from **[Releases](https://github.com/schollz/2n/releases/latest)**.

## Run

(After you have started Jack...)
```bash
./2n
```

Defaults: OSC **57120**, save file **tracker-save.json**.

## Keyboard — Quick Reference

### Navigation (views)
- **Shift+Right** – Navigate deeper into structure
  - Song → Chain (selected track/row)
  - Chain → Phrase (selected row)
  - Phrase → Retrigger/Timestretch/Arpeggio (if set) or File Browser
- **Shift+Left** – Navigate back to parent view
- **Shift+Up** – Go to Settings (from Song/Chain/Phrase) or File Metadata (from File Browser)
- **Shift+Down** – Go to Mixer (from Song/Chain/Phrase) or back from Mixer
- **Arrow keys** – Move cursor/navigate within views
- **Left/Right** – Navigate tracks (Song), chains (Chain), or columns (Phrase)

### Editing
- **Ctrl+Up/Down** – Coarse adjust values (+/-16, coarse increments)
- **Ctrl+Left/Right** – Fine adjust values (+/-1, fine increments)
- **Backspace** – Clear cell/value
- **Ctrl+H** – Delete entire row
- **S** – Paste last edited row
- **C** – Smart trigger/fill function:
  - **Non-empty values**: Triggers `EmitRowDataFor` (plays row with full parameters)
  - **Empty values**: Fills with next available content or copies last row
  - Works in Song, Chain, and Phrase views

### Copy & Paste
- **Ctrl+C** – Copy cell
- **Ctrl+X** – Cut row  
- **Ctrl+V** – Paste
- **Ctrl+D** – Deep copy

### Playback & Recording
- **Space** – Play/stop from current position
- **Ctrl+@** – Play/stop from top (global)
- **Ctrl+R** – Toggle recording mode

### Advanced Functions
- **Ctrl+F** – Smart fill/clear for DT column (Delta Time)
- **Ctrl+S** – Manual save

### Misc
- **Esc** – Clear selection highlight
- **Ctrl+Q** – Quit


## Views

- **Song** – 8 tracks × 16 rows (chains per track), supports Instrument/Sampler track types
- **Chain** – 16 rows mapping to phrases  
- **Phrase** – Main tracker grid with two modes:
  - **Sampler** – Full sample manipulation (pitch, effects, files)
  - **Instrument** – Note-based with chords, ADSR, arpeggio
- **File Browser** – Select audio files for sampler tracks
- **File Metadata** – Configure BPM and slice count per file
- **Retrigger** – Envelope settings for retrigger effects
- **Timestretch** – Time-stretching parameters
- **Arpeggio** – Arpeggio pattern editor (Instrument tracks only)
- **Mixer** – Per-track volume levels
- **Settings** – Global settings (BPM, PPQ, audio gains, etc.)

## Smart 'C' Key Functionality

The **C** key provides context-aware trigger and fill functionality across all views:

### Phrase View
- **Non-empty row**: Triggers `EmitRowDataFor` with complete parameter set:
  - **Instrument tracks**: Note, Chord (C/A/T), ADSR (A/D/S/R), Arpeggio (AR), MIDI (MI), SoundMaker (SO)
  - **Sampler tracks**: All traditional sampler parameters
- **Empty row**: Copies last row with increment

### Chain View  
- **Non-empty slot**: Triggers first row of the referenced phrase
- **Empty slot**: Fills with next unused phrase

### Song View
- **Non-empty slot**: Finds first phrase in referenced chain and triggers its first row
- **Empty slot**: Fills with next unused chain

This unified approach allows instant playback testing of any musical element while maintaining the original fill functionality for composition workflow.


## Phrase Columns

### Sampler View
```
SL  DT  NN  PI  GT  RT  TS  Я  PA  LP  HP  CO  VE  FI
```

### Instrument View  
```
SL  DT  NOT  C  A  T  A D S R  AR  MI  SO
```

### Column Descriptions
- **SL** (slice) – Row number display
- **DT** (delta time) – **Unified playback control**: `--`/`00` = skip, `>00` = play for N ticks
- **NN/NOT** (note) – MIDI note (hex) or note name
- **PI** (pitch) – Pitch bend (sampler only)
- **GT** (gate) – Note length/gate time
- **RT** (retrigger) – Retrigger effect index
- **TS** (timestretch) – Time-stretch effect index  
- **Я** (reverse) – Reverse playback flag
- **PA** (pan) – Stereo panning
- **LP/HP** (filters) – Low-pass/High-pass filters
- **CO** (comb) – Comb filter effect
- **VE** (reverb) – Reverb effect
- **FI** (file index) – Sample file selection (sampler only)
- **C** (chord) – Chord type: None(-), Major(M), minor(m), Dominant(d) (instrument only)
- **A** (chord addition) – Chord addition: None(-), 7th(7), 9th(9), 4th(4) (instrument only)  
- **T** (transposition) – Chord transposition: 0-F semitones (instrument only)
- **A D S R** (ADSR) – Attack/Decay/Sustain/Release envelope (instrument only)
- **AR** (arpeggio) – Arpeggio pattern index (instrument only)
- **MI** (MIDI) – MIDI settings index for external MIDI output (instrument only)
- **SO** (SoundMaker) – SoundMaker settings index for built-in synthesis (instrument only)

### Key Feature: Unified DT Column
Both Sampler and Instrument views now use the same **DT** (Delta Time) column for playback control, replacing the previous separate P/DT system. This provides consistent behavior across both track types.


## License

MIT
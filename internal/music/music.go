package music

import (
	"fmt"
	"strings"
)

// MidiToNoteName converts MIDI note number (0-127) to note name like "c-1", "c#4", etc.
// For negative octaves: natural notes show minus (e.g., "c-1"), sharp notes drop minus (e.g., "f#1") - all stay 3 chars
// MIDI note 60 = C4, note 21 = A0, etc.
func MidiToNoteName(midiNote int) string {
	if midiNote < 0 || midiNote > 127 {
		return "---"
	}

	noteNames := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}

	// Calculate octave (MIDI note 12 = C0)
	octave := (midiNote / 12) - 1

	// Get note name
	noteName := noteNames[midiNote%12]

	// Always maintain exactly 3 characters for all notes
	if strings.Contains(noteName, "#") {
		// Sharp notes: "c#4", "f#1" (already 3 chars for most cases)
		if octave < 0 {
			return fmt.Sprintf("%s%d", noteName, -octave) // "c#1" for negative
		} else {
			return fmt.Sprintf("%s%d", noteName, octave) // "c#4" for positive
		}
	} else {
		// Natural notes: always use minus separator to reach 3 chars
		if octave < 0 {
			return fmt.Sprintf("%s-%d", noteName, -octave) // "c-1" for negative
		} else {
			return fmt.Sprintf("%s-%d", noteName, octave) // "c-4" for positive
		}
	}
}

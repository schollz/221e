package music

import (
	"testing"
)

func TestMidiToNoteName(t *testing.T) {
	tests := []struct {
		name     string
		midiNote int
		expected string
	}{
		// Test valid notes
		{"MIDI 60 should be C4", 60, "c-4"},
		{"MIDI 61 should be C#4", 61, "c#4"},
		{"MIDI 21 should be A0", 21, "a-0"},
		{"MIDI 0 should be C-1", 0, "c-1"},
		{"MIDI 12 should be C0", 12, "c-0"},
		{"MIDI 127 should be G9", 127, "g-9"},

		// Test sharp notes in different octaves
		{"MIDI 1 should be C#1", 1, "c#1"},
		{"MIDI 13 should be C#0", 13, "c#0"},
		{"MIDI 25 should be C#1", 25, "c#1"},

		// Test natural notes across octaves
		{"MIDI 24 should be C1", 24, "c-1"},
		{"MIDI 36 should be C2", 36, "c-2"},
		{"MIDI 48 should be C3", 48, "c-3"},
		{"MIDI 72 should be C5", 72, "c-5"},

		// Test all notes in octave 4 (60-71)
		{"MIDI 60 C4", 60, "c-4"},
		{"MIDI 61 C#4", 61, "c#4"},
		{"MIDI 62 D4", 62, "d-4"},
		{"MIDI 63 D#4", 63, "d#4"},
		{"MIDI 64 E4", 64, "e-4"},
		{"MIDI 65 F4", 65, "f-4"},
		{"MIDI 66 F#4", 66, "f#4"},
		{"MIDI 67 G4", 67, "g-4"},
		{"MIDI 68 G#4", 68, "g#4"},
		{"MIDI 69 A4", 69, "a-4"},
		{"MIDI 70 A#4", 70, "a#4"},
		{"MIDI 71 B4", 71, "b-4"},

		// Test boundary cases
		{"MIDI -1 should be invalid", -1, "---"},
		{"MIDI 128 should be invalid", 128, "---"},
		{"MIDI -100 should be invalid", -100, "---"},
		{"MIDI 200 should be invalid", 200, "---"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MidiToNoteName(tt.midiNote)
			if result != tt.expected {
				t.Errorf("MidiToNoteName(%d) = %q, expected %q", tt.midiNote, result, tt.expected)
			}
		})
	}
}

func TestMidiToNoteNameLength(t *testing.T) {
	// Test that all valid MIDI notes return exactly 3 characters
	for i := 0; i <= 127; i++ {
		result := MidiToNoteName(i)
		if len(result) != 3 {
			t.Errorf("MidiToNoteName(%d) = %q, expected length 3 but got %d", i, result, len(result))
		}
	}
}

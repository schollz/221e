package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChordTypeToString(t *testing.T) {
	tests := []struct {
		name      string
		chordType ChordType
		expected  string
	}{
		{"None chord", ChordNone, "-"},
		{"Major chord", ChordMajor, "M"},
		{"Minor chord", ChordMinor, "m"},
		{"Dominant chord", ChordDominant, "d"},
		{"Invalid chord type", ChordType(999), "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChordTypeToString(tt.chordType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChordAdditionToString(t *testing.T) {
	tests := []struct {
		name     string
		chordAdd ChordAddition
		expected string
	}{
		{"No addition", ChordAddNone, "-"},
		{"7th addition", ChordAdd7, "7"},
		{"9th addition", ChordAdd9, "9"},
		{"4th addition", ChordAdd4, "4"},
		{"Invalid addition", ChordAddition(999), "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChordAdditionToString(tt.chordAdd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChordTranspositionToString(t *testing.T) {
	tests := []struct {
		name       string
		chordTrans ChordTransposition
		expected   string
	}{
		{"No transposition", ChordTransNone, "-"},
		{"Zero transposition", ChordTrans0, "0"},
		{"Middle values", ChordTrans5, "5"},
		{"Hex values", ChordTransA, "A"},
		{"Max hex value", ChordTransF, "F"},
		{"Invalid transposition", ChordTransposition(999), "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChordTranspositionToString(tt.chordTrans)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAttackToSeconds(t *testing.T) {
	tests := []struct {
		name      string
		hexValue  int
		expected  float32
		tolerance float64
	}{
		{"Minimum value (00)", 0, 0.02, 0.001},
		{"Maximum value (FE)", 254, 30.0, 0.001},
		{"Middle value", 127, 0.776, 0.01},   // Approximate middle of exponential curve
		{"Negative value", -1, 0.02, 0.001},  // Should return default
		{"Overflow value", 255, 0.02, 0.001}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AttackToSeconds(tt.hexValue)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

func TestDecayToSeconds(t *testing.T) {
	tests := []struct {
		name      string
		hexValue  int
		expected  float32
		tolerance float64
	}{
		{"Minimum value (00)", 0, 0.0, 0.001},
		{"Maximum value (FE)", 254, 30.0, 0.001},
		{"Middle value", 127, 15.0, 0.1},    // Linear mapping
		{"Negative value", -1, 0.0, 0.001},  // Should return default
		{"Overflow value", 255, 0.0, 0.001}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecayToSeconds(tt.hexValue)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

func TestSustainToLevel(t *testing.T) {
	tests := []struct {
		name      string
		hexValue  int
		expected  float32
		tolerance float64
	}{
		{"Minimum value (00)", 0, 0.0, 0.001},
		{"Maximum value (FE)", 254, 1.0, 0.001},
		{"Middle value", 127, 0.5, 0.01},    // Linear mapping
		{"Negative value", -1, 0.0, 0.001},  // Should return default
		{"Overflow value", 255, 0.0, 0.001}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SustainToLevel(tt.hexValue)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

func TestReleaseToSeconds(t *testing.T) {
	tests := []struct {
		name      string
		hexValue  int
		expected  float32
		tolerance float64
	}{
		{"Minimum value (00)", 0, 0.02, 0.001},
		{"Maximum value (FE)", 254, 30.0, 0.001},
		{"Middle value", 127, 0.776, 0.01},   // Same as Attack (exponential)
		{"Negative value", -1, 0.02, 0.001},  // Should return default
		{"Overflow value", 255, 0.02, 0.001}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReleaseToSeconds(tt.hexValue)
			assert.InDelta(t, tt.expected, result, tt.tolerance)
		})
	}
}

func TestADSRMappingConsistency(t *testing.T) {
	// Test that Attack and Release use the same exponential mapping
	for i := 0; i <= 254; i++ {
		attack := AttackToSeconds(i)
		release := ReleaseToSeconds(i)
		assert.Equal(t, attack, release, "Attack and Release should use identical mapping for value %d", i)
	}
}

func TestADSRBoundaryConditions(t *testing.T) {
	// Test edge cases and boundary conditions
	t.Run("All functions handle negative inputs", func(t *testing.T) {
		assert.Equal(t, float32(0.02), AttackToSeconds(-100))
		assert.Equal(t, float32(0.0), DecayToSeconds(-100))
		assert.Equal(t, float32(0.0), SustainToLevel(-100))
		assert.Equal(t, float32(0.02), ReleaseToSeconds(-100))
	})

	t.Run("All functions handle overflow inputs", func(t *testing.T) {
		assert.Equal(t, float32(0.02), AttackToSeconds(1000))
		assert.Equal(t, float32(0.0), DecayToSeconds(1000))
		assert.Equal(t, float32(0.0), SustainToLevel(1000))
		assert.Equal(t, float32(0.02), ReleaseToSeconds(1000))
	})
}

func TestFileMetadata(t *testing.T) {
	metadata := FileMetadata{
		BPM:    120.5,
		Slices: 8,
	}

	assert.Equal(t, float32(120.5), metadata.BPM)
	assert.Equal(t, 8, metadata.Slices)
}

func TestRetriggerSettings(t *testing.T) {
	settings := RetriggerSettings{
		Times:       4,
		Start:       1.0,
		End:         0.5,
		Beats:       2,
		VolumeDB:    -6.0,
		PitchChange: 2.0,
	}

	assert.Equal(t, 4, settings.Times)
	assert.Equal(t, float32(1.0), settings.Start)
	assert.Equal(t, float32(0.5), settings.End)
	assert.Equal(t, 2, settings.Beats)
	assert.Equal(t, float32(-6.0), settings.VolumeDB)
	assert.Equal(t, float32(2.0), settings.PitchChange)
}

func TestTimestrechSettings(t *testing.T) {
	settings := TimestrechSettings{
		Start: 1.0,
		End:   2.0,
		Beats: 4,
	}

	assert.Equal(t, float32(1.0), settings.Start)
	assert.Equal(t, float32(2.0), settings.End)
	assert.Equal(t, 4, settings.Beats)
}

func TestArpeggioRow(t *testing.T) {
	row := ArpeggioRow{
		Direction: 1, // "u-"
		Count:     16,
		Divisor:   4,
	}

	assert.Equal(t, 1, row.Direction)
	assert.Equal(t, 16, row.Count)
	assert.Equal(t, 4, row.Divisor)
}

func TestMidiSettings(t *testing.T) {
	settings := MidiSettings{
		Device:  "Test MIDI Device",
		Channel: "10",
	}

	assert.Equal(t, "Test MIDI Device", settings.Device)
	assert.Equal(t, "10", settings.Channel)
}

func TestSoundMakerSettings(t *testing.T) {
	settings := SoundMakerSettings{
		Name: "Polyperc",
		A:    50,
		B:    100,
		C:    -1, // "--" value
		D:    200,
	}

	assert.Equal(t, "Polyperc", settings.Name)
	assert.Equal(t, 50, settings.A)
	assert.Equal(t, 100, settings.B)
	assert.Equal(t, -1, settings.C)
	assert.Equal(t, 200, settings.D)
}

func TestClipboardData(t *testing.T) {
	clipboard := ClipboardData{
		Value:           0x80,
		CellType:        HexCell,
		Mode:            CellMode,
		HasData:         true,
		HighlightRow:    5,
		HighlightCol:    3,
		HighlightPhrase: 10,
		HighlightView:   PhraseView,
	}

	assert.Equal(t, 0x80, clipboard.Value)
	assert.Equal(t, HexCell, clipboard.CellType)
	assert.Equal(t, CellMode, clipboard.Mode)
	assert.True(t, clipboard.HasData)
	assert.Equal(t, 5, clipboard.HighlightRow)
	assert.Equal(t, 3, clipboard.HighlightCol)
	assert.Equal(t, 10, clipboard.HighlightPhrase)
	assert.Equal(t, PhraseView, clipboard.HighlightView)
}

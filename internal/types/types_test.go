package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChordNotes(t *testing.T) {
	tests := []struct {
		name      string
		root      int
		ctype     ChordType
		add       ChordAddition
		transpose ChordTransposition
		want      []int
	}{
		{
			name:      "ChordNone returns root only",
			root:      60,
			ctype:     ChordNone,
			add:       ChordAddNone,
			transpose: ChordTransNone,
			want:      []int{60},
		},
		{
			name:      "Major triad",
			root:      60,
			ctype:     ChordMajor,
			add:       ChordAddNone,
			transpose: ChordTransNone,
			want:      []int{60, 64, 67},
		},
		{
			name:      "Minor triad",
			root:      60,
			ctype:     ChordMinor,
			add:       ChordAddNone,
			transpose: ChordTransNone,
			want:      []int{60, 63, 67},
		},
		{
			name:      "Dominant triad (same as major without 7th)",
			root:      60,
			ctype:     ChordDominant,
			add:       ChordAddNone,
			transpose: ChordTransNone,
			want:      []int{60, 64, 67},
		},
		{
			name:      "Minor + add7 (minor 7th)",
			root:      60,
			ctype:     ChordMinor,
			add:       ChordAdd7,
			transpose: ChordTransNone,
			want:      []int{60, 63, 67, 70}, // 60 + 10
		},
		{
			name:      "Dominant + add7 (major 7th per current logic)",
			root:      60,
			ctype:     ChordDominant,
			add:       ChordAdd7,
			transpose: ChordTransNone,
			want:      []int{60, 64, 67, 71}, // 60 + 11
		},
		{
			name:      "Major + add9",
			root:      60,
			ctype:     ChordMajor,
			add:       ChordAdd9,
			transpose: ChordTransNone,
			want:      []int{60, 64, 67, 74}, // 60 + 14
		},
		{
			name:      "Minor + add4 (order keeps addition last)",
			root:      60,
			ctype:     ChordMinor,
			add:       ChordAdd4,
			transpose: ChordTransNone,
			want:      []int{60, 63, 67, 65}, // 60 + 5 appended after triad
		},
		{
			name:      "Minor + 2 transpose (chord rotation)",
			root:      60, // C minor = C, Eb, G = [60, 63, 67]
			ctype:     ChordMinor,
			add:       ChordAddNone,
			transpose: ChordTrans2, // t=2: G, C+oct, Eb+oct = [67, 72, 75]
			want:      []int{67, 72, 75},
		},
		{
			name:      "Major + 2 transpose (chord rotation)",
			root:      60, // C major = C, E, G = [60, 64, 67]
			ctype:     ChordMajor,
			add:       ChordAddNone,
			transpose: ChordTrans2, // t=2: G, C+oct, E+oct = [67, 72, 76]
			want:      []int{67, 72, 76},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := GetChordNotes(tt.root, tt.ctype, tt.add, tt.transpose)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetChordNotes(%d, %v, %v) = %v, want %v",
					tt.root, tt.ctype, tt.add, got, tt.want)
			}
		})
	}
}

func TestChordTypeToString(t *testing.T) {
	tests := []struct {
		chordType ChordType
		expected  string
	}{
		{ChordNone, "-"},
		{ChordMajor, "M"},
		{ChordMinor, "m"},
		{ChordDominant, "d"},
		{ChordType(999), "-"}, // Invalid value should default to "-"
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ChordTypeToString(tt.chordType)
			if result != tt.expected {
				t.Errorf("ChordTypeToString(%v) = %q, expected %q", tt.chordType, result, tt.expected)
			}
		})
	}
}

func TestChordAdditionToString(t *testing.T) {
	tests := []struct {
		chordAdd ChordAddition
		expected string
	}{
		{ChordAddNone, "-"},
		{ChordAdd7, "7"},
		{ChordAdd9, "9"},
		{ChordAdd4, "4"},
		{ChordAddition(999), "-"}, // Invalid value should default to "-"
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ChordAdditionToString(tt.chordAdd)
			if result != tt.expected {
				t.Errorf("ChordAdditionToString(%v) = %q, expected %q", tt.chordAdd, result, tt.expected)
			}
		})
	}
}

func TestChordTranspositionToString(t *testing.T) {
	tests := []struct {
		chordTrans ChordTransposition
		expected   string
	}{
		{ChordTransNone, "-"},
		{ChordTrans0, "0"},
		{ChordTrans1, "1"},
		{ChordTrans2, "2"},
		{ChordTrans3, "3"},
		{ChordTrans4, "4"},
		{ChordTrans5, "5"},
		{ChordTrans6, "6"},
		{ChordTrans7, "7"},
		{ChordTrans8, "8"},
		{ChordTrans9, "9"},
		{ChordTransA, "A"},
		{ChordTransB, "B"},
		{ChordTransC, "C"},
		{ChordTransD, "D"},
		{ChordTransE, "E"},
		{ChordTransF, "F"},
		{ChordTransposition(999), "-"}, // Invalid value should default to "-"
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ChordTranspositionToString(tt.chordTrans)
			if result != tt.expected {
				t.Errorf("ChordTranspositionToString(%v) = %q, expected %q", tt.chordTrans, result, tt.expected)
			}
		})
	}
}

func TestAttackToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		hexValue int
		expected float32
	}{
		{"minimum value", 0, 0.02},
		{"maximum value", 254, 30.0},
		{"mid value", 127, 0.77459663}, // Actual exponential midpoint
		{"below range", -1, 0.02},
		{"above range", 255, 0.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AttackToSeconds(tt.hexValue)
			if tt.name == "mid value" {
				// For exponential functions, check within tolerance
				assert.InDelta(t, tt.expected, result, 0.1)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDecayToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		hexValue int
		expected float32
	}{
		{"minimum value", 0, 0.0},
		{"maximum value", 254, 30.0},
		{"mid value", 127, 15.0},
		{"below range", -1, 0.0},
		{"above range", 255, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecayToSeconds(tt.hexValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSustainToLevel(t *testing.T) {
	tests := []struct {
		name     string
		hexValue int
		expected float32
	}{
		{"minimum value", 0, 0.0},
		{"maximum value", 254, 1.0},
		{"mid value", 127, 0.5},
		{"below range", -1, 0.0},
		{"above range", 255, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SustainToLevel(tt.hexValue)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestReleaseToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		hexValue int
		expected float32
	}{
		{"minimum value", 0, 0.02},
		{"maximum value", 254, 30.0},
		{"mid value", 127, 0.77459663}, // Actual exponential midpoint
		{"below range", -1, 0.02},
		{"above range", 255, 0.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReleaseToSeconds(tt.hexValue)
			if tt.name == "mid value" {
				// For exponential functions, check within tolerance
				assert.InDelta(t, tt.expected, result, 0.1)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetVirtualDefault(t *testing.T) {
	tests := []struct {
		name     string
		col      PhraseColumn
		expected *VirtualDefaultConfig
	}{
		{"ColPitch", ColPitch, &VirtualDefaultConfig{DefaultValue: 0x80}},
		{"ColGate", ColGate, &VirtualDefaultConfig{DefaultValue: 0x80}},
		{"ColPan", ColPan, &VirtualDefaultConfig{DefaultValue: 0x80}},
		{"ColLowPassFilter", ColLowPassFilter, &VirtualDefaultConfig{DefaultValue: 0xFE}},
		{"ColNote (no default)", ColNote, nil},
		{"ColFilename (no default)", ColFilename, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVirtualDefault(tt.col)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.DefaultValue, result.DefaultValue)
			}
		})
	}
}

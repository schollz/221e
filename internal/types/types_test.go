package types

import (
	"reflect"
	"testing"
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

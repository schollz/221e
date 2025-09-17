package getbpm

import (
	"math"
	"testing"
)

func TestLength(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		numframes  int64
		samplerate int64
		wantErr    bool
	}{
		{
			name:       "amen wav file",
			filename:   "amen_beats8_bpm172.wav",
			samplerate: 44100,
			numframes:  123069,
			wantErr:    false,
		},
		{
			name:       "strega wav file",
			filename:   "strega.wav",
			samplerate: 48000,
			numframes:  1363244,
			wantErr:    false,
		},
		{
			name:       "invalid wav file",
			filename:   "testdata/invalid.wav",
			numframes:  0,
			samplerate: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, samplerate, numframes, err := Length(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Length() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if numframes != tt.numframes {
				t.Errorf("Length() got numframes = %v, want %v", numframes, tt.numframes)
			}
			if samplerate != tt.samplerate {
				t.Errorf("Length() got samplerate = %v, want %v", samplerate, tt.samplerate)
			}
		})
	}
}

func TestGetBPM(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		expectedBeats float64
		expectedBPM   float64
		wantErr       bool
	}{
		{
			name:          "amen beats8 bpm172",
			filename:      "amen_beats8_bpm172.wav",
			expectedBeats: 8.0,
			expectedBPM:   172.0,
			wantErr:       false,
		},
		{
			name:          "hands bpm176 beats32",
			filename:      "hands_bpm176_beats32.wav",
			expectedBeats: 32.0,
			expectedBPM:   176.0,
			wantErr:       false,
		},
		{
			name:          "strega no metadata",
			filename:      "strega.wav",
			expectedBeats: 0,
			expectedBPM:   0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beats, bpm, err := GetBPM(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBPM() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if tt.expectedBeats > 0 && math.Abs(beats-tt.expectedBeats) > 0.1 {
				t.Errorf("GetBPM() got beats = %v, want %v", beats, tt.expectedBeats)
			}
			if tt.expectedBPM > 0 && math.Abs(bpm-tt.expectedBPM) > 0.1 {
				t.Errorf("GetBPM() got bpm = %v, want %v", bpm, tt.expectedBPM)
			}

			if tt.expectedBeats == 0 && tt.expectedBPM == 0 {
				if beats == 0 || bpm == 0 {
					t.Errorf("GetBPM() for strega.wav should guess beats and bpm, got beats=%v, bpm=%v", beats, bpm)
				}
			}
		})
	}
}

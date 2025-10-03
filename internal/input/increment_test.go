package input

import (
	"testing"

	"github.com/schollz/collidertracker/internal/modulation"
)

func TestApplyIncrementIntegration(t *testing.T) {
	// Test that ApplyIncrement works with different counter values
	tests := []struct {
		name             string
		originalNote     int
		incrementCounter int
		incrementValue   int
		wrapValue        int
		expected         int
	}{
		{
			name:             "Counter -1 should not apply increment",
			originalNote:     60,
			incrementCounter: -1,
			incrementValue:   10,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "Counter 0 should apply increment (0 + note)",
			originalNote:     60,
			incrementCounter: 0,
			incrementValue:   10,
			wrapValue:        0,
			expected:         60, // 60 + 0 = 60
		},
		{
			name:             "Counter 5 should apply increment (5 + note)",
			originalNote:     40,
			incrementCounter: 5,
			incrementValue:   10,
			wrapValue:        0,
			expected:         45, // 40 + 5 = 45
		},
		{
			name:             "Increment value 0 should not apply",
			originalNote:     60,
			incrementCounter: 10,
			incrementValue:   0,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "Test wrap functionality",
			originalNote:     60,
			incrementCounter: 10,
			incrementValue:   5,
			wrapValue:        7,
			expected:         63, // 60 + (10 % 7) = 60 + 3 = 63
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modulation.ApplyIncrement(tt.originalNote, tt.incrementCounter, tt.incrementValue, tt.wrapValue)
			if result != tt.expected {
				t.Errorf("ApplyIncrement(%d, %d, %d, %d) = %d; want %d",
					tt.originalNote, tt.incrementCounter, tt.incrementValue, tt.wrapValue, result, tt.expected)
			}
		})
	}
}

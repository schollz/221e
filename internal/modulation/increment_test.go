package modulation

import (
	"testing"
)

func TestApplyIncrement(t *testing.T) {
	tests := []struct {
		name             string
		originalNote     int
		incrementCounter int
		incrementValue   int
		expected         int
	}{
		{
			name:             "No increment when counter is -1",
			originalNote:     60,
			incrementCounter: -1,
			incrementValue:   5,
			expected:         60,
		},
		{
			name:             "No increment when increment value is 0",
			originalNote:     60,
			incrementCounter: 3,
			incrementValue:   0,
			expected:         60,
		},
		{
			name:             "Apply increment when counter > -1 and increment > 0",
			originalNote:     60,
			incrementCounter: 3,
			incrementValue:   5,
			expected:         63,
		},
		{
			name:             "Apply increment with counter 0",
			originalNote:     60,
			incrementCounter: 0,
			incrementValue:   10,
			expected:         60,
		},
		{
			name:             "Apply increment with large values",
			originalNote:     40,
			incrementCounter: 20,
			incrementValue:   8,
			expected:         60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyIncrement(tt.originalNote, tt.incrementCounter, tt.incrementValue)
			if result != tt.expected {
				t.Errorf("ApplyIncrement(%d, %d, %d) = %d; want %d",
					tt.originalNote, tt.incrementCounter, tt.incrementValue, result, tt.expected)
			}
		})
	}
}

func TestNewModulateSettingsWithIncrement(t *testing.T) {
	settings := NewModulateSettings()
	if settings.Increment != 0 {
		t.Errorf("NewModulateSettings().Increment = %d; want 0", settings.Increment)
	}
}
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
		wrapValue        int
		expected         int
	}{
		{
			name:             "No increment when counter is -1",
			originalNote:     60,
			incrementCounter: -1,
			incrementValue:   5,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "No increment when increment value is 0",
			originalNote:     60,
			incrementCounter: 3,
			incrementValue:   0,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "Apply increment when counter > -1 and increment > 0",
			originalNote:     60,
			incrementCounter: 3,
			incrementValue:   5,
			wrapValue:        0,
			expected:         63,
		},
		{
			name:             "Apply increment with counter 0",
			originalNote:     60,
			incrementCounter: 0,
			incrementValue:   10,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "Apply increment with large values",
			originalNote:     40,
			incrementCounter: 20,
			incrementValue:   8,
			wrapValue:        0,
			expected:         60,
		},
		{
			name:             "Wrap counter when wrap value is set - simple case",
			originalNote:     60,
			incrementCounter: 10,
			incrementValue:   5,
			wrapValue:        8,
			expected:         62, // 60 + (10 % 8) = 60 + 2 = 62
		},
		{
			name:             "Wrap counter when counter equals wrap value",
			originalNote:     60,
			incrementCounter: 8,
			incrementValue:   5,
			wrapValue:        8,
			expected:         60, // 60 + (8 % 8) = 60 + 0 = 60
		},
		{
			name:             "No wrap when counter is less than wrap value",
			originalNote:     60,
			incrementCounter: 5,
			incrementValue:   5,
			wrapValue:        8,
			expected:         65, // 60 + 5 = 65 (no wrapping)
		},
		{
			name:             "Wrap with larger numbers",
			originalNote:     60,
			incrementCounter: 25,
			incrementValue:   5,
			wrapValue:        12,
			expected:         61, // 60 + (25 % 12) = 60 + 1 = 61
		},
		{
			name:             "No wrap when wrap value is 0",
			originalNote:     60,
			incrementCounter: 15,
			incrementValue:   5,
			wrapValue:        0,
			expected:         75, // 60 + 15 = 75 (no wrapping when wrap = 0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyIncrement(tt.originalNote, tt.incrementCounter, tt.incrementValue, tt.wrapValue)
			if result != tt.expected {
				t.Errorf("ApplyIncrement(%d, %d, %d, %d) = %d; want %d",
					tt.originalNote, tt.incrementCounter, tt.incrementValue, tt.wrapValue, result, tt.expected)
			}
		})
	}
}

func TestNewModulateSettingsWithIncrement(t *testing.T) {
	settings := NewModulateSettings()
	if settings.Increment != 0 {
		t.Errorf("NewModulateSettings().Increment = %d; want 0", settings.Increment)
	}
	if settings.Wrap != 0 {
		t.Errorf("NewModulateSettings().Wrap = %d; want 0", settings.Wrap)
	}
}

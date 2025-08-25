package views

import (
	"math"
	"testing"
)

func TestBrailleSineWave(t *testing.T) {
	// Configurable dimensions (in Braille cells)
	const width = 83
	const height = 3

	// Generate sine wave data in [-1, 1]
	const samples = 1000
	data := make([]float64, samples)
	for i := 0; i < samples; i++ {
		theta := 2 * math.Pi * float64(i) / float64(samples-1) // one full cycle
		data[i] = math.Sin(theta)
	}

	out := RenderWaveform(width, height, data)
	if out == "" {
		t.Fatalf("display returned empty output")
	}

	// Log the waveform so it shows up in test output.
	t.Log("\n" + out)
}

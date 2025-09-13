package views

import (
	"math"
	"strings"

	"github.com/schollz/collidertracker/internal/types"
)

// RenderWaveform renders waveform data (assumed in [-1,1]) into a Braille string.
// width and height are in Braille cells. Each Braille cell is 2x4 dots.
// RenderWaveform renders waveform data (assumed in [-1,1]) into a Braille string.
// width and height are in Braille cells. Each Braille cell is 2x4 dots.
//
// Optimized to minimize allocations/copies:
//   - Avoids building a [][]bool grid; uses a flat []byte mask per 2x4 cell.
//   - Single pass over fine columns sets bits directly into cell masks.
//   - Reuses small helpers and bounds checks to keep hot loop tight.
func RenderWaveform(width, height int, data []float64) string {
	if width <= 0 || height <= 0 || len(data) == 0 {
		return ""
	}

	fineW := width * 2  // dot-columns
	fineH := height * 4 // dot-rows

	// Helper: sample data at fractional index using linear interpolation.
	sampleAt := func(p float64) float64 {
		if p <= 0 {
			return data[0]
		}
		max := float64(len(data) - 1)
		if p >= max {
			return data[len(data)-1]
		}
		i := int(math.Floor(p))
		f := p - float64(i)
		return data[i]*(1-f) + data[i+1]*f
	}

	// Each cell is a single byte mask for its 2x4 Braille dots.
	masks := make([]byte, width*height)

	// Braille dot bit masks
	const (
		dot1 = 0x01
		dot2 = 0x02
		dot3 = 0x04
		dot4 = 0x08
		dot5 = 0x10
		dot6 = 0x20
		dot7 = 0x40
		dot8 = 0x80
	)
	const brailleBase = 0x2800

	// Precompute total data span
	span := float64(len(data) - 1)
	if span <= 0 {
		span = 1
	}

	// Single pass over fine X columns; set the one dot hit per column.
	for x := 0; x < fineW; x++ {
		p := (float64(x) / float64(fineW-1)) * span
		v := sampleAt(p) // in [-1,1]

		// Map v to vertical dot index (0 at top)
		y := int(math.Round((1.0 - (v+1.0)/2.0) * float64(fineH-1)))
		if y < 0 {
			y = 0
		} else if y >= fineH {
			y = fineH - 1
		}

		cellCol := x >> 1 // /2
		cellRow := y >> 2 // /4
		inCol := x & 1    // 0..1
		inRow := y & 3    // 0..3

		// Map (inRow,inCol) -> braille dot bit
		var bit byte
		switch types.BrailleDotRow(inRow) {
		case types.BrailleDotRow0:
			if inCol == 0 {
				bit = dot1
			} else {
				bit = dot4
			}
		case types.BrailleDotRow1:
			if inCol == 0 {
				bit = dot2
			} else {
				bit = dot5
			}
		case types.BrailleDotRow2:
			if inCol == 0 {
				bit = dot3
			} else {
				bit = dot6
			}
		default: // BrailleDotRow3
			if inCol == 0 {
				bit = dot7
			} else {
				bit = dot8
			}
		}

		idx := cellRow*width + cellCol
		masks[idx] |= bit
	}

	var b strings.Builder
	// Each cell becomes 1 rune; each row has width runes + 1 newline (except last).
	b.Grow(height*width + (height - 1))

	for row := 0; row < height; row++ {
		base := row * width
		for col := 0; col < width; col++ {
			mask := masks[base+col]
			r := rune(brailleBase + int(mask))
			b.WriteRune(r)
		}
		if row != height-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

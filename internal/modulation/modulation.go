package modulation

import (
	"log"
	"math/rand"
)

// ModulateSettings represents the settings for a single modulation entry
type ModulateSettings struct {
	Seed      int    `json:"seed"`      // Random seed: -1 for "none" (no randomization), 0 for "random" (time seeding), 1-128 for fixed seed
	IRandom   int    `json:"irandom"`   // Random range: 0-128 (0 means no randomization)
	Sub       int    `json:"sub"`       // Subtract value: 0-120
	Add       int    `json:"add"`       // Add value: 0-120
	ScaleRoot int    `json:"scaleRoot"` // Scale root note: 0-11 (C, C#, D, D#, E, F, F#, G, G#, A, A#, B)
	Scale     string `json:"scale"`     // Scale selection: "all", "major", "minor", etc.
}

// Scale represents a musical scale
type Scale struct {
	Name  string
	Notes []int // MIDI note offsets within an octave (0-11)
}

// Predefined scales
var Scales = map[string]Scale{
	"all": {
		Name:  "All Notes",
		Notes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
	"major": {
		Name:  "Major",
		Notes: []int{0, 2, 4, 5, 7, 9, 11},
	},
	"minor": {
		Name:  "Minor",
		Notes: []int{0, 2, 3, 5, 7, 8, 10},
	},
	"dorian": {
		Name:  "Dorian",
		Notes: []int{0, 2, 3, 5, 7, 9, 10},
	},
	"mixolydian": {
		Name:  "Mixolydian",
		Notes: []int{0, 2, 4, 5, 7, 9, 10},
	},
	"pentatonic": {
		Name:  "Pentatonic",
		Notes: []int{0, 2, 4, 7, 9},
	},
	"blues": {
		Name:  "Blues",
		Notes: []int{0, 3, 5, 6, 7, 10},
	},
	"chromatic": {
		Name:  "Chromatic",
		Notes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	},
}

// Note names for scale root selection
var NoteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// GetScaleNames returns a list of all available scale names
func GetScaleNames() []string {
	names := make([]string, 0, len(Scales))
	for name := range Scales {
		names = append(names, name)
	}
	return names
}

// GetNoteNames returns a list of all note names
func GetNoteNames() []string {
	return NoteNames
}

// ApplyModulation applies modulation to a MIDI note value using the provided RNG
func ApplyModulation(originalNote int, settings ModulateSettings, rng *rand.Rand) int {
	log.Printf("DEBUG: ApplyModulation - Seed=%d, IRandom=%d, Sub=%d, Add=%d",
		settings.Seed, settings.IRandom, settings.Sub, settings.Add)

	// Start with the original note
	result := originalNote
	log.Printf("DEBUG: Start with originalNote=%d", result)

	// Step 1: Apply random variation if IRandom > 0 and Seed != -1 (none)
	if settings.IRandom > 0 && settings.Seed != -1 {
		var randomValue int

		// Use fixed seed if specified (> 0), or time seeding for seed=0 ("random")
		if settings.Seed > 0 {
			// Create a new random source with the specified seed for reproducible results
			seedRng := rand.New(rand.NewSource(int64(settings.Seed)))
			randomValue = seedRng.Intn(settings.IRandom + 1)
			log.Printf("DEBUG: Using fixed seed %d, got random %d", settings.Seed, randomValue)
		} else if settings.Seed == 0 {
			// Seed 0 means "random" - use provided track-specific RNG for time seeding
			randomValue = rng.Intn(settings.IRandom + 1)
			log.Printf("DEBUG: Using random seeding (seed=0), got random %d", randomValue)
		}

		result = randomValue
		log.Printf("DEBUG: After IRandom, result=%d", result)
	} else if settings.Seed == -1 {
		log.Printf("DEBUG: Seed is 'none' (-1), skipping randomization")
	}

	// Step 2: Subtract the Sub value
	result -= settings.Sub
	log.Printf("DEBUG: After Sub-%d, result=%d", settings.Sub, result)

	// Step 3: Add the Add value
	result += settings.Add
	log.Printf("DEBUG: After Add+%d, result=%d", settings.Add, result)

	// Step 4: Apply scale quantization if a scale is selected
	if settings.Scale != "all" && settings.Scale != "" {
		result = quantizeToScale(result, settings.Scale, settings.ScaleRoot)
	}

	// Ensure result is within valid MIDI range (0-127)
	if result < 0 {
		result = 0
	} else if result > 127 {
		result = 127
	}

	return result
}

// quantizeToScale quantizes a MIDI note to the closest note in the specified scale
func quantizeToScale(note int, scaleName string, scaleRoot int) int {
	scale, exists := Scales[scaleName]
	if !exists {
		// If scale doesn't exist, return the original note
		return note
	}

	// Handle negative notes by wrapping to positive octave
	if note < 0 {
		octaves := (-note / 12) + 1
		note += octaves * 12
	}

	// Get the note within the octave (0-11)
	octave := note / 12
	noteInOctave := note % 12

	// Apply scale root transposition - transpose the input note down by scaleRoot
	// so we can work in the scale's natural form, then transpose back up
	transposedNote := (noteInOctave - scaleRoot + 12) % 12

	// Find the closest note in the scale
	minDistance := 12
	closestNote := transposedNote

	for _, scaleNote := range scale.Notes {
		distance := abs(transposedNote - scaleNote)
		if distance < minDistance {
			minDistance = distance
			closestNote = scaleNote
		}
	}

	// Transpose the result back to the correct root and reconstruct the full MIDI note
	finalNote := (closestNote + scaleRoot) % 12
	return octave*12 + finalNote
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// NewModulateSettings creates a new ModulateSettings with default values
func NewModulateSettings() ModulateSettings {
	return ModulateSettings{
		Seed:      -1, // Default to "none"
		IRandom:   0,
		Sub:       0,
		Add:       0,
		ScaleRoot: 0, // Default to C
		Scale:     "all",
	}
}

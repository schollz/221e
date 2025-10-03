package modulation

import (
	"log"
	"math/rand"
	"time"
)

// ModulateSettings represents the settings for a single modulation entry
type ModulateSettings struct {
	Seed        int    `json:"seed"`        // Random seed: -1 for "none" (no randomization), 0 for "random" (time seeding), 1-128 for fixed seed
	IRandom     int    `json:"irandom"`     // Random range: 0-128 (0 means no randomization)
	Sub         int    `json:"sub"`         // Subtract value: 0-120
	Add         int    `json:"add"`         // Add value: 0-120
	Increment   int    `json:"increment"`   // Increment value: 0-128 (added to note when increment counter > -1)
	Wrap        int    `json:"wrap"`        // Wrap value: 0-128 (0 = none, wraps increment counter when exceeded)
	ScaleRoot   int    `json:"scaleRoot"`   // Scale root note: 0-11 (C, C#, D, D#, E, F, F#, G, G#, A, A#, B)
	Scale       string `json:"scale"`       // Scale selection: "all", "major", "minor", etc.
	Probability int    `json:"probability"` // Probability percentage: 0-100 (100 = always apply modulation)
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

// ApplyIncrement applies increment to a note value based on increment counter
// This should be called before other modulation operations
func ApplyIncrement(originalNote int, incrementCounter int, incrementValue int, wrapValue int) int {
	if incrementCounter > -1 && incrementValue > 0 {
		log.Printf("DEBUG: ApplyIncrement - originalNote=%d, incrementCounter=%d, incrementValue=%d, wrapValue=%d",
			originalNote, incrementCounter, incrementValue, wrapValue)

		// Apply wrapping logic if wrap value is greater than 0
		wrappedCounter := incrementCounter
		if wrapValue > 0 && incrementCounter >= wrapValue {
			// Subtract wrap value until counter is less than wrap value
			wrappedCounter = incrementCounter % wrapValue
			log.Printf("DEBUG: ApplyIncrement - applied wrapping: %d -> %d (wrap=%d)",
				incrementCounter, wrappedCounter, wrapValue)
		}

		result := originalNote + wrappedCounter
		log.Printf("DEBUG: ApplyIncrement - result=%d", result)
		return result
	}
	return originalNote
}

// ApplyModulation applies modulation to a MIDI note value using the provided RNG
func ApplyModulation(originalNote int, settings ModulateSettings, rng *rand.Rand) int {
	log.Printf("DEBUG: ApplyModulation - Seed=%d, IRandom=%d, Sub=%d, Add=%d, Probability=%d",
		settings.Seed, settings.IRandom, settings.Sub, settings.Add, settings.Probability)

	// Check probability first - determine if modulation should occur at all
	if settings.Probability < 100 {
		// Generate a random number 1-100 to compare against probability
		probabilityRoll := rng.Intn(100) + 1
		if probabilityRoll > settings.Probability {
			log.Printf("DEBUG: Modulation skipped - probabilityRoll=%d > probability=%d", probabilityRoll, settings.Probability)
			return originalNote // Return original note without any modulation
		}
		log.Printf("DEBUG: Modulation proceeding - probabilityRoll=%d <= probability=%d", probabilityRoll, settings.Probability)
	}

	// Start with the original note
	result := originalNote
	log.Printf("DEBUG: Start with originalNote=%d", result)

	// Step 1: Apply random variation if IRandom > 0
	if settings.IRandom > 0 {

		// Use fixed seed if specified (> 0), or time seeding for seed=0 ("random")
		if settings.Seed > 0 {
			// Create a new random source with the specified seed for reproducible results
			rng.Seed(int64(settings.Seed))
		} else if settings.Seed == 0 {
			// time based seed
			rng.Seed(time.Now().UnixNano())
		}

		result += rng.Intn(settings.IRandom + 1)
		log.Printf("DEBUG: After IRandom, result=%d", result)
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
		Seed:        -1, // Default to "none"
		IRandom:     0,
		Sub:         0,
		Add:         0,
		Increment:   0, // Default increment value
		Wrap:        0, // Default wrap value (0 = none)
		ScaleRoot:   0, // Default to C
		Scale:       "all",
		Probability: 100, // Default to always apply modulation
	}
}

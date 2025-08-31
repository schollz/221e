package input

import (
	"fmt"
	"log"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/types"
)

// GetPhrasesDataForTrack returns the appropriate phrases data based on track type
func GetPhrasesDataForTrack(m *model.Model, track int) *[255][][]int {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentPhrasesData
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerPhrasesData
}

// GetChainsDataForTrack returns the appropriate chains data based on track type
func GetChainsDataForTrack(m *model.Model, track int) *[][]int {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentChainsData
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerChainsData
}

// ValueModifier represents a function that modifies a value with bounds checking
type ValueModifier struct {
	GetValue         func() interface{}
	SetValue         func(interface{})
	ApplyDelta       func(current interface{}, delta float32) interface{}
	ValidateAndClamp func(value interface{}) interface{}
	FormatValue      func(value interface{}) string
	LogPrefix        string
}

// modifyValueWithBounds provides common logic for value modification with bounds checking
func modifyValueWithBounds(modifier ValueModifier, delta float32) {
	current := modifier.GetValue()
	newValue := modifier.ApplyDelta(current, delta)
	clampedValue := modifier.ValidateAndClamp(newValue)

	// Log the change
	oldFormatted := modifier.FormatValue(current)
	newFormatted := modifier.FormatValue(clampedValue)
	log.Printf("Modified %s: %s -> %s (delta: %.2f)", modifier.LogPrefix, oldFormatted, newFormatted, delta)

	modifier.SetValue(clampedValue)
}

// createFloatModifier creates a ValueModifier for float32 values
func createFloatModifier(getValue func() float32, setValue func(float32), min, max float32, logPrefix string) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(float32)) },
		ApplyDelta: func(current interface{}, delta float32) interface{} {
			return current.(float32) + delta
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(float32)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%.2f", value.(float32))
		},
		LogPrefix: logPrefix,
	}
}

// createIntModifier creates a ValueModifier for int values
func createIntModifier(getValue func() int, setValue func(int), min, max int, logPrefix string) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(int)) },
		ApplyDelta: func(current interface{}, delta float32) interface{} {
			return current.(int) + int(delta)
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(int)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%d", value.(int))
		},
		LogPrefix: logPrefix,
	}
}

// deltaHandler handles different delta increments (coarse vs fine control)
func deltaHandler(baseDelta float32, coarseMultiplier float32) float32 {
	if baseDelta == 1.0 || baseDelta == -1.0 {
		return baseDelta * coarseMultiplier // Coarse control (Ctrl+Up/Down)
	} else if baseDelta == 0.05 || baseDelta == -0.05 {
		return baseDelta / 0.05 // Fine control (Ctrl+Left/Right): converts to +/-1
	}
	return baseDelta // Fallback
}

// createRetriggerIntModifier creates a ValueModifier for retrigger int values with special delta handling
func createRetriggerIntModifier(getValue func() int, setValue func(int), min, max int, logPrefix string, coarseMultiplier float32) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(int)) },
		ApplyDelta: func(current interface{}, baseDelta float32) interface{} {
			delta := deltaHandler(baseDelta, coarseMultiplier)
			return current.(int) + int(delta)
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(int)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%d", value.(int))
		},
		LogPrefix: logPrefix,
	}
}

// createRetriggerFloatModifier creates a ValueModifier for retrigger float values with special delta handling
func createRetriggerFloatModifier(getValue func() float32, setValue func(float32), min, max float32, logPrefix string) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(float32)) },
		ApplyDelta: func(current interface{}, baseDelta float32) interface{} {
			// For float values, use the delta directly for both coarse and fine control
			var delta float32
			if baseDelta == 1.0 || baseDelta == -1.0 {
				delta = baseDelta // Coarse control
			} else {
				delta = baseDelta // Fine control
			}
			return current.(float32) + delta
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(float32)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%.2f", value.(float32))
		},
		LogPrefix: logPrefix,
	}
}

func ModifySettingsValue(m *model.Model, delta float32) {
	switch m.CurrentRow {
	case 0: // BPM
		modifier := createFloatModifier(
			func() float32 { return m.BPM },
			func(v float32) { m.BPM = v },
			1, 999, "BPM",
		)
		modifyValueWithBounds(modifier, delta)

	case 1: // PPQ
		modifier := createIntModifier(
			func() int { return m.PPQ },
			func(v int) { m.PPQ = v },
			1, 32, "PPQ",
		)
		modifyValueWithBounds(modifier, delta)

	case 2: // PregainDB
		modifier := createFloatModifier(
			func() float32 { return m.PregainDB },
			func(v float32) {
				m.PregainDB = v
				m.SendOSCPregainMessage() // Send OSC message for pregain change
			},
			-96, 32, "PregainDB",
		)
		modifyValueWithBounds(modifier, delta)

	case 3: // PostgainDB
		modifier := createFloatModifier(
			func() float32 { return m.PostgainDB },
			func(v float32) {
				m.PostgainDB = v
				m.SendOSCPostgainMessage() // Send OSC message for postgain change
			},
			-96, 32, "PostgainDB",
		)
		modifyValueWithBounds(modifier, delta)

	case 4: // BiasDB
		modifier := createFloatModifier(
			func() float32 { return m.BiasDB },
			func(v float32) {
				m.BiasDB = v
				m.SendOSCBiasMessage() // Send OSC message for bias change
			},
			-96, 32, "BiasDB",
		)
		modifyValueWithBounds(modifier, delta)

	case 5: // SaturationDB
		modifier := createFloatModifier(
			func() float32 { return m.SaturationDB },
			func(v float32) {
				m.SaturationDB = v
				m.SendOSCSaturationMessage() // Send OSC message for saturation change
			},
			-96, 32, "SaturationDB",
		)
		modifyValueWithBounds(modifier, delta)

	case 6: // DriveDB
		modifier := createFloatModifier(
			func() float32 { return m.DriveDB },
			func(v float32) {
				m.DriveDB = v
				m.SendOSCDriveMessage() // Send OSC message for drive change
			},
			-96, 32, "DriveDB",
		)
		modifyValueWithBounds(modifier, delta)
	}
	storage.AutoSave(m)
}

func ModifyFileMetadataValue(m *model.Model, delta float32) {
	if m.MetadataEditingFile == "" {
		return
	}

	// Get current metadata or create default
	metadata, exists := m.FileMetadata[m.MetadataEditingFile]
	if !exists {
		metadata = types.FileMetadata{BPM: 120.0, Slices: 16} // Default values
	}

	switch m.CurrentRow {
	case 0: // BPM
		modifier := createFloatModifier(
			func() float32 { return metadata.BPM },
			func(v float32) {
				metadata.BPM = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
			},
			1, 999, fmt.Sprintf("file metadata BPM for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)

	case 1: // Slices
		modifier := createIntModifier(
			func() int { return metadata.Slices },
			func(v int) {
				metadata.Slices = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
			},
			1, 999, fmt.Sprintf("file metadata Slices for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)
	}

	storage.AutoSave(m)
}

func ModifyRetriggerValue(m *model.Model, baseDelta float32) {
	if m.RetriggerEditingIndex < 0 || m.RetriggerEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := m.RetriggerSettings[m.RetriggerEditingIndex]

	if m.CurrentRow == 0 { // Times
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newTimes := settings.Times + delta
		if newTimes < 0 {
			newTimes = 0
		} else if newTimes > 256 {
			newTimes = 256
		}
		settings.Times = newTimes
		log.Printf("Modified retrigger %02X Times: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Times-delta, settings.Times, delta)
	} else if m.CurrentRow == 1 { // Starting Rate
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newStart := settings.Start + delta
		if newStart < 0 {
			newStart = 0
		} else if newStart > 256 {
			newStart = 256
		}
		settings.Start = newStart
		log.Printf("Modified retrigger %02X Starting Rate: %.2f -> %.2f (delta: %.2f)", m.RetriggerEditingIndex, settings.Start-delta, settings.Start, delta)
	} else if m.CurrentRow == 2 { // Final Rate
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newEnd := settings.End + delta
		if newEnd < 0 {
			newEnd = 0
		} else if newEnd > 256 {
			newEnd = 256
		}
		settings.End = newEnd
		log.Printf("Modified retrigger %02X Final Rate: %.2f -> %.2f (delta: %.2f)", m.RetriggerEditingIndex, settings.End-delta, settings.End, delta)
	} else if m.CurrentRow == 3 { // Beats
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newBeats := settings.Beats + delta
		if newBeats < 0 {
			newBeats = 0
		} else if newBeats > 256 {
			newBeats = 256
		}
		settings.Beats = newBeats
		log.Printf("Modified retrigger %02X Beats: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Beats-delta, settings.Beats, delta)
	} else if m.CurrentRow == 4 { // Volume dB
		newVolumeDB := settings.VolumeDB + baseDelta
		if newVolumeDB < -16.0 {
			newVolumeDB = -16.0
		} else if newVolumeDB > 16.0 {
			newVolumeDB = 16.0
		}
		settings.VolumeDB = newVolumeDB
		log.Printf("Modified retrigger %02X VolumeDB: %.1f -> %.1f (delta: %.1f)", m.RetriggerEditingIndex, settings.VolumeDB-baseDelta, settings.VolumeDB, baseDelta)
	} else if m.CurrentRow == 5 { // Pitch change
		newPitchChange := settings.PitchChange + baseDelta
		if newPitchChange < -24.0 {
			newPitchChange = -24.0
		} else if newPitchChange > 24.0 {
			newPitchChange = 24.0
		}
		settings.PitchChange = newPitchChange
		log.Printf("Modified retrigger %02X PitchChange: %.1f -> %.1f (delta: %.1f)", m.RetriggerEditingIndex, settings.PitchChange-baseDelta, settings.PitchChange, baseDelta)
	}

	// Store back the modified settings
	m.RetriggerSettings[m.RetriggerEditingIndex] = settings
	storage.AutoSave(m)
}

func ModifyTimestrechValue(m *model.Model, baseDelta float32) {
	if m.TimestrechEditingIndex < 0 || m.TimestrechEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := m.TimestrechSettings[m.TimestrechEditingIndex]

	if m.CurrentRow == 0 { // Start
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newStart := settings.Start + delta
		if newStart < 0 {
			newStart = 0
		} else if newStart > 256 {
			newStart = 256
		}
		settings.Start = newStart
		log.Printf("Modified timestretch %02X Start: %.2f -> %.2f (delta: %.2f)", m.TimestrechEditingIndex, settings.Start-delta, settings.Start, delta)
	} else if m.CurrentRow == 1 { // End
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newEnd := settings.End + delta
		if newEnd < 0 {
			newEnd = 0
		} else if newEnd > 256 {
			newEnd = 256
		}
		settings.End = newEnd
		log.Printf("Modified timestretch %02X End: %.2f -> %.2f (delta: %.2f)", m.TimestrechEditingIndex, settings.End-delta, settings.End, delta)
	} else if m.CurrentRow == 2 { // Beats
		newBeats := settings.Beats + int(baseDelta)
		if newBeats < 0 {
			newBeats = 0
		} else if newBeats > 256 {
			newBeats = 256
		}
		settings.Beats = newBeats
		log.Printf("Modified timestretch %02X Beats: %d -> %d (delta: %.2f)", m.TimestrechEditingIndex, settings.Beats-int(baseDelta), settings.Beats, baseDelta)
	}

	// Store back the modified settings
	m.TimestrechSettings[m.TimestrechEditingIndex] = settings
	storage.AutoSave(m)
}

func ModifyArpeggioValue(m *model.Model, baseDelta float32) {
	if m.ArpeggioEditingIndex < 0 || m.ArpeggioEditingIndex >= 255 {
		return
	}
	if m.CurrentRow < 0 || m.CurrentRow >= 16 {
		return
	}

	// Get current settings
	settings := m.ArpeggioSettings[m.ArpeggioEditingIndex]
	currentRow := &settings.Rows[m.CurrentRow] // Get reference to specific row

	if m.CurrentCol == int(types.ArpeggioColDI) { // DI (Direction) column
		// Direction cycles through: 0="--", 1="u-", 2="d-"
		var delta int
		if baseDelta > 0 {
			delta = 1
		} else {
			delta = -1
		}

		newDirection := currentRow.Direction + delta
		if newDirection < 0 {
			newDirection = 2 // Wrap to "d-"
		} else if newDirection > 2 {
			newDirection = 0 // Wrap to "--"
		}
		currentRow.Direction = newDirection
		log.Printf("Modified arpeggio %02X row %02X Direction: %d -> %d", m.ArpeggioEditingIndex, m.CurrentRow, currentRow.Direction-delta, currentRow.Direction)
	} else if m.CurrentCol == int(types.ArpeggioColCO) { // CO (Count) column
		// Count: -1="--", 0-254 for hex values 00-FE
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 16 // Coarse control (Ctrl+Up/Down): +/-16
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newCount := currentRow.Count + delta
		if newCount < -1 {
			newCount = -1 // Can't go below "--"
		} else if newCount > 254 {
			newCount = 254 // Cap at FE
		}
		currentRow.Count = newCount
		log.Printf("Modified arpeggio %02X row %02X Count: %d -> %d (delta: %d)", m.ArpeggioEditingIndex, m.CurrentRow, currentRow.Count-delta, currentRow.Count, delta)
	} else if m.CurrentCol == int(types.ArpeggioColDIV) { // Divisor (/) column
		// Divisor: -1="--", 1-254 for hex values 01-FE (never allow 00)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 16 // Coarse control (Ctrl+Up/Down): +/-16
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newDivisor := currentRow.Divisor + delta
		if currentRow.Divisor == -1 && delta > 0 {
			// When going up from "--", start at 1 (not 0)
			newDivisor = 1
		} else if newDivisor <= 0 {
			// Going below 1, wrap to "--"
			newDivisor = -1
		} else if newDivisor > 254 {
			// Cap at FE (254)
			newDivisor = 254
		}
		currentRow.Divisor = newDivisor
		log.Printf("Modified arpeggio %02X row %02X Divisor: %d -> %d (delta: %d)", m.ArpeggioEditingIndex, m.CurrentRow, currentRow.Divisor-delta, currentRow.Divisor, delta)
	}

	// Store back the modified settings
	m.ArpeggioSettings[m.ArpeggioEditingIndex] = settings
	storage.AutoSave(m)
}

func ModifyMidiValue(m *model.Model, baseDelta float32) {
	if m.MidiEditingIndex < 0 || m.MidiEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := &m.MidiSettings[m.MidiEditingIndex]

	if m.CurrentRow == 0 { // Device row
		// Device cycles through available MIDI devices from AvailableMidiDevices
		var delta int
		if baseDelta > 0 {
			delta = 1
		} else {
			delta = -1
		}

		// Use the actual available MIDI devices list, with "None" as first option
		devices := []string{"None"}
		devices = append(devices, m.AvailableMidiDevices...)

		// Find current index
		currentIndex := -1
		for i, dev := range devices {
			if settings.Device == dev {
				currentIndex = i
				break
			}
		}

		// If current device not found, start from beginning
		if currentIndex == -1 {
			currentIndex = 0
		}

		// Apply delta with wrapping
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = len(devices) - 1 // Wrap to last device
		} else if newIndex >= len(devices) {
			newIndex = 0 // Wrap to first device
		}

		oldDevice := settings.Device
		settings.Device = devices[newIndex]
		log.Printf("Modified MIDI %02X Device: %s -> %s", m.MidiEditingIndex, oldDevice, settings.Device)
	} else if m.CurrentRow == 1 { // Channel row
		// Channel cycles through: "1"-"16" and "all"
		var delta int
		if baseDelta > 0 {
			delta = 1
		} else {
			delta = -1
		}

		channels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "all"}

		// Find current index
		currentIndex := -1
		for i, ch := range channels {
			if settings.Channel == ch {
				currentIndex = i
				break
			}
		}

		// If current channel not found, start from beginning
		if currentIndex == -1 {
			currentIndex = 0
		}

		// Apply delta with wrapping
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = len(channels) - 1 // Wrap to "all"
		} else if newIndex >= len(channels) {
			newIndex = 0 // Wrap to "1"
		}

		oldChannel := settings.Channel
		settings.Channel = channels[newIndex]
		log.Printf("Modified MIDI %02X Channel: %s -> %s", m.MidiEditingIndex, oldChannel, settings.Channel)
	}

	storage.AutoSave(m)
}

func ModifySoundMakerValue(m *model.Model, baseDelta float32) {
	if m.SoundMakerEditingIndex < 0 || m.SoundMakerEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := &m.SoundMakerSettings[m.SoundMakerEditingIndex]

	if m.CurrentRow == 0 { // Name row
		// Name cycles through available SoundMakers
		var delta int
		if baseDelta > 0 {
			delta = 1
		} else {
			delta = -1
		}

		// Use the available SoundMaker names
		soundMakers := []string{"None", "Polyperc", "Infinite Pad"}

		// Find current index
		currentIndex := -1
		for i, sm := range soundMakers {
			if settings.Name == sm {
				currentIndex = i
				break
			}
		}

		// If current SoundMaker not found, start from beginning
		if currentIndex == -1 {
			currentIndex = 0
		}

		// Apply delta with wrapping
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = len(soundMakers) - 1 // Wrap to last SoundMaker
		} else if newIndex >= len(soundMakers) {
			newIndex = 0 // Wrap to first SoundMaker
		}

		oldName := settings.Name
		settings.Name = soundMakers[newIndex]
		log.Printf("Modified SoundMaker %02X Name: %s -> %s", m.SoundMakerEditingIndex, oldName, settings.Name)
	} else if m.CurrentRow >= 1 && m.CurrentRow <= 4 { // Parameters A, B, C, D
		// Parameters cycle through 00-FE or -- (which is -1)
		var delta int
		if baseDelta >= 1.0 || baseDelta <= -1.0 {
			// Coarse control: Ctrl+Up/Down = ±16
			if baseDelta > 0 {
				delta = 16
			} else {
				delta = -16
			}
		} else {
			// Fine control: Ctrl+Left/Right = ±1
			if baseDelta > 0 {
				delta = 1
			} else {
				delta = -1
			}
		}

		// Get the current parameter value
		var currentValue *int
		switch m.CurrentRow {
		case 1: // Parameter A
			currentValue = &settings.A
		case 2: // Parameter B
			currentValue = &settings.B
		case 3: // Parameter C
			currentValue = &settings.C
		case 4: // Parameter D
			currentValue = &settings.D
		}

		if currentValue != nil {
			oldValue := *currentValue
			var newValue int

			if *currentValue == -1 {
				// If currently "--", start from 00
				if delta > 0 {
					newValue = 0
				} else {
					newValue = 254 // FE
				}
			} else {
				newValue = *currentValue + delta
				// Wrap around: 00-FE (0-254) or -- (-1)
				if newValue > 254 {
					newValue = -1 // Wrap to "--"
				} else if newValue < -1 {
					newValue = 254 // Wrap to FE
				}
			}

			*currentValue = newValue
			paramName := string(rune('A' + (m.CurrentRow - 1)))
			if newValue == -1 {
				log.Printf("Modified SoundMaker %02X Parameter %s: %02X -> -- (delta: %d)", m.SoundMakerEditingIndex, paramName, oldValue, delta)
			} else {
				if oldValue == -1 {
					log.Printf("Modified SoundMaker %02X Parameter %s: -- -> %02X (delta: %d)", m.SoundMakerEditingIndex, paramName, newValue, delta)
				} else {
					log.Printf("Modified SoundMaker %02X Parameter %s: %02X -> %02X (delta: %d)", m.SoundMakerEditingIndex, paramName, oldValue, newValue, delta)
				}
			}
		}
	}

	storage.AutoSave(m)
}

func ModifyValue(m *model.Model, delta int) {
	if m.ViewMode == types.ChainView {
		// Chain view now only has phrase editing
		chainsData := m.GetCurrentChainsData()
		currentValue := (*chainsData)[m.CurrentChain][m.CurrentRow]

		var newValue int
		if currentValue == -1 {
			// First edit on an empty cell: initialize to 00 and DO NOT apply delta
			newValue = 0
		} else {
			newValue = currentValue + delta
		}

		if newValue < 0 {
			newValue = 0
		} else if newValue > 254 {
			newValue = 254
		}
		(*chainsData)[m.CurrentChain][m.CurrentRow] = newValue

		log.Printf("Modified chain %02X row %02X phrase: %d -> %d (delta: %d)", m.CurrentChain, m.CurrentRow, currentValue, newValue, delta)
		storage.AutoSave(m)
		return
	}

	// Phrase view: modify cell under cursor
	// Use centralized column mapping system
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	if columnMapping == nil || !columnMapping.IsEditable {
		return // Invalid or non-editable column
	}

	colIndex := columnMapping.DataColumnIndex
	phrasesData := m.GetCurrentPhrasesData()
	currentValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex]

	if colIndex == int(types.ColDeltaTime) {
		// DT column: clamp 0..254 (hex range)
		if currentValue == -1 {
			currentValue = 0
		}
		newValue := currentValue + delta
		if newValue < 0 {
			newValue = 0
		} else if newValue > 254 {
			newValue = 254
		}
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

	} else if colIndex == int(types.ColEffectReverse) {
		// Я column: single bit 0..1 (keep existing behavior)
		if currentValue == -1 {
			currentValue = 0
		}
		newValue := currentValue + delta
		if newValue < 0 {
			newValue = 0
		} else if newValue > 1 {
			newValue = 1
		}
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

	} else {
		// Handle different behavior for Instrument vs Sampler views
		phraseViewType := m.GetPhraseViewType()

		if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColNote) {
			// Instrument view note column: MIDI notes (0-127) with special increment behavior
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to middle C (60)
				newValue = 60
			} else {
				// Apply special increment logic for instrument notes
				// Coarse (Ctrl+Up/Down) should increment by 12 (octaves)
				// Fine (Ctrl+Left/Right) should increment by 1 (semitones)
				if delta == 16 || delta == -16 {
					// This is coarse increment - convert to octave increment (12 semitones)
					octaveDelta := (delta / 16) * 12
					newValue = currentValue + octaveDelta
				} else {
					// This is fine increment (+/-1)
					newValue = currentValue + delta
				}
			}

			// Clamp to MIDI range (0-127)
			if newValue < 0 {
				newValue = 0
			} else if newValue > 127 {
				newValue = 127
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

			// Auto-set DT=1 only when changing from no note (-1) to a note
			if currentValue == -1 && newValue != -1 {
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = 1
			}
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChord) {
			// Instrument view chord column: Cycle through chord types, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordNone if unset
				currentValue = int(types.ChordNone)
			}

			// Determine direction based on delta (Ctrl+Up/Right = forward, Ctrl+Down/Left = backward)
			if delta > 0 {
				// Forward: "-" -> "M" -> "m" -> "d" (stop at "d")
				newValue = currentValue + 1
				if newValue >= int(types.ChordTypeCount) {
					newValue = int(types.ChordTypeCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "d" -> "m" -> "M" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordAddition) {
			// Instrument view chord addition column: Cycle through addition types, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordAddNone if unset
				currentValue = int(types.ChordAddNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward: "-" -> "7" -> "9" -> "4" (stop at "4")
				newValue = currentValue + 1
				if newValue >= int(types.ChordAdditionCount) {
					newValue = int(types.ChordAdditionCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "4" -> "9" -> "7" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordTransposition) {
			// Instrument view chord transposition column: Cycle through transposition values, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordTransNone if unset
				currentValue = int(types.ChordTransNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward: "-" -> "0" -> "1" -> ... -> "F" (stop at "F")
				newValue = currentValue + 1
				if newValue >= int(types.ChordTranspositionCount) {
					newValue = int(types.ChordTranspositionCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "F" -> "E" -> ... -> "0" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColMidi) {
			// Instrument view MIDI column: hex values 00-FE
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to 00 and DO NOT apply delta
				newValue = 0
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 254 {
				newValue = 254
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else {
			// All other hex-ish columns (NN, DT, GT, RT, TS, CO, VE, FI index) - original behavior
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to 00 and DO NOT apply delta
				newValue = 0
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 254 {
				newValue = 254
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		}

		// Auto-enable playback on first note entry - use DT for both views
		if colIndex == int(types.ColNote) {
			if (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] == -1 {
				// Auto-set DT to 01 when a note is added (only if DT is currently -1/"--")
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] = 1
				log.Printf("Auto-set DT=01 for phrase %d row %d due to note change", m.CurrentPhrase, m.CurrentRow)
			}
		}
	}

	m.LastEditRow = m.CurrentRow
	log.Printf("Modified phrase %d row %d, col %d: %d -> %d (delta: %d)",
		m.CurrentPhrase, m.CurrentRow, colIndex, currentValue,
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex], delta)
	storage.AutoSave(m)
}

func TogglePlayback(m *model.Model) tea.Cmd {
	var config PlaybackConfig

	if m.ViewMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: true,         // Start from current selected row/track
			Chain:         -1,           // Not used for song mode
			Phrase:        -1,           // Not used for song mode
			Row:           m.CurrentRow, // Song row
		}
	} else if m.ViewMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: true,           // Start from current chain row position
			Chain:         m.CurrentChain, // Use the chain we're currently viewing
			Phrase:        0,              // Will be determined from chain
			Row:           m.CurrentRow,   // Use current chain row
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: true, // Start from current selected row
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           m.CurrentRow,
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromTop(m *model.Model) tea.Cmd {
	var config PlaybackConfig

	if m.ViewMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: false, // Always start from song row 0
			Chain:         -1,    // Not used for song mode
			Phrase:        -1,    // Not used for song mode
			Row:           0,     // Start from song row 0
		}
	} else if m.ViewMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: false,          // Always start from top/first non-empty
			Chain:         m.CurrentChain, // Use current chain being viewed
			Phrase:        0,              // Will be determined from chain
			Row:           -1,             // Will be determined
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: false, // Start from first non-empty row in phrase
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           -1, // Will be determined
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromTopGlobal(m *model.Model) tea.Cmd {
	// Determine playback mode based on the current view
	var playbackMode types.ViewMode
	if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		playbackMode = m.ViewMode
	} else {
		// Use PreviousView if it's Song, Chain or Phrase, otherwise default to Phrase
		if m.PreviousView == types.SongView || m.PreviousView == types.ChainView || m.PreviousView == types.PhraseView {
			playbackMode = m.PreviousView
		} else {
			// Default to phrase view if no clear editing history
			playbackMode = types.PhraseView
		}
	}

	var config PlaybackConfig

	if playbackMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: false, // Always start from song row 0
			Chain:         -1,    // Not used for song mode
			Phrase:        -1,    // Not used for song mode
			Row:           0,     // Start from song row 0
		}
	} else if playbackMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: false, // Always start from top/first non-empty
			Chain:         m.CurrentChain,
			Phrase:        0,  // Will be determined from chain
			Row:           -1, // Will be determined
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: false, // Start from first non-empty row in phrase
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           -1, // Will be determined
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromLastSongRow(m *model.Model) tea.Cmd {
	// Always play ALL tracks from the last Song view row, regardless of current view
	config := PlaybackConfig{
		Mode:          types.SongView,
		UseCurrentRow: false,
		Chain:         -1,            // Not used for song mode
		Phrase:        -1,            // Not used for song mode
		Row:           m.LastSongRow, // Start from last selected song row
	}

	return togglePlaybackWithConfig(m, config)
}

func Tick(m *model.Model) tea.Cmd {
	ms := rowDurationMS(m)
	return tea.Tick(time.Duration(ms)*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AdvancePlayback(m *model.Model) {
	oldRow := m.PlaybackRow

	if m.PlaybackMode == types.SongView {
		// Song playback mode with per-track tick counting
		log.Printf("Song playback advancing - checking %d tracks", 8)
		activeTrackCount := 0

		for track := 0; track < 8; track++ {
			if !m.SongPlaybackActive[track] {
				continue
			}
			activeTrackCount++
			log.Printf("DEBUG_SONG: Processing active track %d, ticksLeft=%d", track, m.SongPlaybackTicksLeft[track])

			// Decrement ticks for this track
			if m.SongPlaybackTicksLeft[track] > 0 {
				m.SongPlaybackTicksLeft[track]--
				log.Printf("Song track %d: %d ticks remaining", track, m.SongPlaybackTicksLeft[track])
			}

			// Only advance this track when its ticks reach 0
			if m.SongPlaybackTicksLeft[track] > 0 {
				continue
			}

			// Now advance to next playable row for this track
			if !advanceToNextPlayableRowForTrack(m, track) {
				// Track finished, deactivate
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d deactivated (end of sequence)", track)
				continue
			}

			// Load new ticks for the advanced row
			m.LoadTicksLeftForTrack(track)

			// Emit the newly advanced row immediately (at start of its DT period)
			phraseNum := m.SongPlaybackPhrase[track]
			currentRow := m.SongPlaybackRowInPhrase[track]
			if phraseNum >= 0 && phraseNum < 255 && currentRow >= 0 && currentRow < 255 {
				EmitRowDataFor(m, phraseNum, currentRow, track)
				log.Printf("Song track %d emitted phrase %02X row %d with %d ticks", track, phraseNum, currentRow, m.SongPlaybackTicksLeft[track])
			}
		}
		log.Printf("Song playback: processed %d active tracks", activeTrackCount)
	} else if m.PlaybackMode == types.ChainView {
		// Chain playback mode - advance through phrases in sequence
		// Find next row with playback enabled (unified DT-based playback)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)

		// Validate PlaybackPhrase is within bounds before accessing array
		if m.PlaybackPhrase >= 0 && m.PlaybackPhrase < 255 {
			for i := m.PlaybackRow + 1; i < 255; i++ {
				// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
				dtValue := (*phrasesData)[m.PlaybackPhrase][i][types.ColDeltaTime]
				if IsRowPlayable(dtValue) {
					m.PlaybackRow = i
					DebugLogRowEmission(m)
					log.Printf("Chain playback advanced from row %d to %d", oldRow, m.PlaybackRow)
					return
				}
			}
		}

		// End of phrase reached, move to next phrase slot in the same chain
		chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
		for i := m.PlaybackChainRow + 1; i < 16; i++ {
			phraseID := (*chainsData)[m.PlaybackChain][i]
			if phraseID != -1 && phraseID >= 0 && phraseID < 255 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = phraseID
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback moved to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}

		// End of chain reached, loop back to first phrase slot in the same chain
		for i := 0; i < 16; i++ {
			phraseID := (*chainsData)[m.PlaybackChain][i]
			if phraseID != -1 && phraseID >= 0 && phraseID < 255 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = phraseID
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback looped back to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}

		// No valid phrases found in this chain - stop playback
		log.Printf("Chain playback stopped - no valid phrases found in chain %d", m.PlaybackChain)
		return
	} else {
		// Phrase-only playback mode
		// Find next row with playback enabled (unified DT-based playback)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
		for i := m.PlaybackRow + 1; i < 255; i++ {
			// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
			dtValue := (*phrasesData)[m.PlaybackPhrase][i][types.ColDeltaTime]
			if IsRowPlayable(dtValue) {
				m.PlaybackRow = i
				DebugLogRowEmission(m)
				log.Printf("Phrase playback advanced from row %d to %d", oldRow, m.PlaybackRow)
				return
			}
		}

		// Loop back to beginning of phrase
		m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
		DebugLogRowEmission(m)
		log.Printf("Phrase playback looped from row %d back to %d", oldRow, m.PlaybackRow)
	}
}

// advanceToNextPlayableRowForTrack advances a track to its next playable row
// Returns true if successful, false if track should be stopped
func advanceToNextPlayableRowForTrack(m *model.Model, track int) bool {
	if track < 0 || track >= 8 {
		return false
	}

	// Try to advance within current phrase first
	phraseNum := m.SongPlaybackPhrase[track]
	if phraseNum >= 0 && phraseNum < 255 {
		phrasesData := GetPhrasesDataForTrack(m, track)
		for i := m.SongPlaybackRowInPhrase[track] + 1; i < 255; i++ {
			dtValue := (*phrasesData)[phraseNum][i][types.ColDeltaTime]
			if dtValue >= 1 {
				m.SongPlaybackRowInPhrase[track] = i
				log.Printf("Song track %d advanced within phrase to row %d", track, i)
				return true
			}
		}
	}

	// End of phrase reached, try to advance within current chain
	currentChain := m.SongPlaybackChain[track]
	chainsData := m.GetChainsDataForTrack(track)
	for chainRow := m.SongPlaybackChainRow[track] + 1; chainRow < 16; chainRow++ {
		phraseID := (*chainsData)[currentChain][chainRow]
		if phraseID != -1 {
			// Found next phrase in chain, find its first playable row
			m.SongPlaybackChainRow[track] = chainRow
			m.SongPlaybackPhrase[track] = phraseID
			if findFirstPlayableRowInPhraseForTrack(m, phraseID, track) {
				log.Printf("Song track %d advanced to chain row %d, phrase %02X", track, chainRow, phraseID)
				return true
			}
		}
	}

	// End of chain reached, find next valid song row
	startSearchRow := m.SongPlaybackRow[track] + 1
	for searchOffset := 0; searchOffset < 16; searchOffset++ {
		searchRow := (startSearchRow + searchOffset) % 16
		chainID := m.SongData[track][searchRow]

		if chainID != -1 {
			// Check if this chain has any phrases with playable rows
			for chainRow := 0; chainRow < 16; chainRow++ {
				phraseID := (*chainsData)[chainID][chainRow]
				if phraseID != -1 {
					// Found a phrase, check if it has playable rows
					if findFirstPlayableRowInPhraseForTrack(m, phraseID, track) {
						// Valid chain found
						m.SongPlaybackRow[track] = searchRow
						m.SongPlaybackChain[track] = chainID
						m.SongPlaybackChainRow[track] = chainRow
						m.SongPlaybackPhrase[track] = phraseID
						log.Printf("Song track %d advanced to song row %02X, chain %02X", track, searchRow, chainID)
						return true
					}
				}
			}
		}
	}

	// No valid sequences found, track should stop
	return false
}

// findFirstPlayableRowInPhraseForTrack finds the first playable row in a phrase for a track
// Sets the track's SongPlaybackRowInPhrase and returns true if found
func findFirstPlayableRowInPhraseForTrack(m *model.Model, phraseNum, track int) bool {
	if phraseNum < 0 || phraseNum >= 255 || track < 0 || track >= 8 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, track)
	for row := 0; row < 255; row++ {
		dtValue := (*phrasesData)[phraseNum][row][types.ColDeltaTime]
		if dtValue >= 1 {
			m.SongPlaybackRowInPhrase[track] = row
			return true
		}
	}
	return false
}

func DebugLogRowEmission(m *model.Model) {
	// Delegate to the single canonical emitter so "space" playback and "c" manual emit behave identically.
	if m.PlaybackPhrase < 0 || m.PlaybackPhrase >= 255 || m.PlaybackRow < 0 || m.PlaybackRow >= 255 {
		log.Printf("ROW_EMIT: Invalid playback position - Phrase: %d, Row: %d", m.PlaybackPhrase, m.PlaybackRow)
		return
	}
	// Use current track context for chain and phrase playback modes
	EmitRowDataFor(m, m.PlaybackPhrase, m.PlaybackRow, m.CurrentTrack)
}

func FindFirstNonEmptyRowInPhrase(m *model.Model, phraseNum int) int {
	// Use current track context to determine which data pool to search
	return FindFirstNonEmptyRowInPhraseForTrack(m, phraseNum, m.CurrentTrack)
}

func FindFirstNonEmptyRowInPhraseForTrack(m *model.Model, phraseNum int, track int) int {
	if phraseNum >= 0 && phraseNum < 255 {
		phrasesData := GetPhrasesDataForTrack(m, track)
		log.Printf("DEBUG: FindFirstNonEmptyRowInPhraseForTrack - phrase=%d, track=%d", phraseNum, track)
		for i := 0; i < 255; i++ {
			// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
			dtValue := (*phrasesData)[phraseNum][i][types.ColDeltaTime]
			if IsRowPlayable(dtValue) {
				log.Printf("DEBUG: Found playback row %d with DT=%d", i, dtValue)
				return i
			}
		}
		log.Printf("DEBUG: No playback rows found in phrase %d track %d", phraseNum, track)
	}
	return 0 // Fallback to row 0 if no playback rows found
}

func FindFirstNonEmptyChain(m *model.Model) int {
	chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
	for i := 0; i < 255; i++ {
		// Check if any phrase is assigned in this chain
		for row := 0; row < 16; row++ {
			if (*chainsData)[i][row] != -1 {
				return i
			}
		}
	}
	return 0 // Fallback to chain 0 if no data found
}

func CopyLastRowWithIncrement(m *model.Model) {
	if m.ViewMode != types.PhraseView {
		return // Only works in phrase view
	}

	log.Printf("copyLastRowWithIncrement called - currentRow: %d", m.CurrentRow)

	// Get the appropriate phrases data based on view type
	phrasesData := m.GetCurrentPhrasesData()
	phraseViewType := m.GetPhraseViewType()

	// Check if current row is empty (note, deltatime, filename are -1, playback is 0)
	if (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColNote] != -1 ||
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] != -1 ||
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename] != -1 {
		log.Printf("Current row %d is not empty, skipping copy", m.CurrentRow)
		return
	}

	// Find the first non-null note above the current row
	var sourceNote int = -1
	for r := m.CurrentRow - 1; r >= 0; r-- {
		if (*phrasesData)[m.CurrentPhrase][r][types.ColNote] != -1 {
			sourceNote = (*phrasesData)[m.CurrentPhrase][r][types.ColNote]
			log.Printf("Found non-null note %d at row %d", sourceNote, r)
			break
		}
	}

	// If no non-null note found above, use default starting note
	if sourceNote == -1 {
		if phraseViewType == types.InstrumentPhraseView {
			// For Instrument view, start with middle C (60)
			sourceNote = 59 // Will be incremented to 60
		} else {
			// For Sampler view, start with 0
			sourceNote = -1 // Will be incremented to 0
		}
		log.Printf("No non-null note found above current row, using default start")
	}

	// Set DT to 1 for both view types
	(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = 1

	// Increment the note and set it
	var newNote int
	if phraseViewType == types.InstrumentPhraseView {
		// For Instrument view: increment MIDI notes (0-127), with chromatic progression
		newNote = sourceNote + 1
		if newNote > 127 { // Wrap around MIDI range
			newNote = 0
		}
	} else {
		// For Sampler view: increment sample/note numbers (0-254)
		newNote = sourceNote + 1
		if newNote > 254 { // Wrap around if needed
			newNote = 0
		}
	}
	(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = newNote

	log.Printf("Set note on row %d: %d->%d, P=1", m.CurrentRow, sourceNote, newNote)

	// Update the last edit row to current row
	m.LastEditRow = m.CurrentRow
}

func CopyCellToClipboard(m *model.Model) {
	if m.ViewMode == types.SongView {
		// Copy chain ID from song view
		value := m.SongData[m.CurrentCol][m.CurrentRow]
		clipboard := types.ClipboardData{
			Value:           value,
			CellType:        types.HexCell,
			Mode:            types.CellMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    m.CurrentCol,
			HighlightPhrase: -1, // Not applicable for song view
			HighlightView:   types.SongView,
		}
		m.Clipboard = clipboard
		log.Printf("Copied song chain value: %d", value)
	} else if m.ViewMode == types.ChainView {
		// Copy phrase number from chain view
		chainsData := m.GetCurrentChainsData()
		value := (*chainsData)[m.CurrentChain][m.CurrentRow]
		clipboard := types.ClipboardData{
			Value:           value,
			CellType:        types.HexCell,
			Mode:            types.CellMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    1, // still highlight the PH column
			HighlightPhrase: -1,
			HighlightView:   types.ChainView,
		}
		m.Clipboard = clipboard
		log.Printf("Copied chain phrase value: %d", value)
	} else if m.ViewMode == types.PhraseView {
		// Copy from phrase view
		phrasesData := m.GetCurrentPhrasesData()

		// Use centralized column mapping system
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping == nil || !columnMapping.IsCopyable {
			return // Invalid or non-copyable column
		}

		colIndex := columnMapping.DataColumnIndex
		if colIndex >= 0 && colIndex < int(types.ColCount) {
			value := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex]
			var cellType types.CellType
			if colIndex == int(types.ColFilename) { // Filename column
				cellType = types.FilenameCell
			} else { // Playback, note, or deltatime column
				cellType = types.HexCell
			}
			clipboard := types.ClipboardData{
				Value:           value,
				CellType:        cellType,
				Mode:            types.CellMode,
				HasData:         true,
				HighlightRow:    m.CurrentRow,
				HighlightCol:    m.CurrentCol,
				HighlightPhrase: m.CurrentPhrase,
				HighlightView:   types.PhraseView,
			}
			m.Clipboard = clipboard
			log.Printf("Copied phrase cell value: %d, type: %v", value, cellType)
		}
	}
}

func CutRowToClipboard(m *model.Model) {
	if m.ViewMode == types.ChainView {
		// Cut row from chain view
		rowData := make([]int, 2)
		copy(rowData, m.ChainsData[m.CurrentRow])
		clipboard := types.ClipboardData{
			RowData:         rowData,
			SourceView:      types.ChainView,
			Mode:            types.RowMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    -1, // Highlight entire row
			HighlightPhrase: -1, // Not applicable for chain view
			HighlightView:   types.ChainView,
		}
		m.Clipboard = clipboard
		// Clear the row (but keep chain number)
		m.ChainsData[m.CurrentChain][m.CurrentRow] = -1
		log.Printf("Cut chain row %d", m.CurrentRow)
	} else if m.ViewMode == types.PhraseView {
		// Cut row from phrase view
		phrasesData := m.GetCurrentPhrasesData()
		rowData := make([]int, int(types.ColCount))
		copy(rowData, (*phrasesData)[m.CurrentPhrase][m.CurrentRow])

		// Get filename if exists
		var filename string
		fileIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if fileIndex >= 0 && fileIndex < len(*phrasesFiles) {
			filename = (*phrasesFiles)[fileIndex]
		}

		clipboard := types.ClipboardData{
			RowData:         rowData,
			RowFilename:     filename,
			SourceView:      types.PhraseView,
			Mode:            types.RowMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    -1, // Highlight entire row
			HighlightPhrase: m.CurrentPhrase,
			HighlightView:   types.PhraseView,
		}
		m.Clipboard = clipboard
		// Clear the row
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1      // Clear note
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1 // Clear deltatime (clears playback for both views)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = -1  // Clear filename
		log.Printf("Cut phrase row %d", m.CurrentRow)
	}
}

func PasteFromClipboard(m *model.Model) {
	if !m.Clipboard.HasData {
		log.Printf("No data in clipboard")
		return
	}

	if m.Clipboard.Mode == types.CellMode {
		PasteCellFromClipboard(m)
	} else if m.Clipboard.Mode == types.RowMode {
		PasteRowFromClipboard(m)
	}
}

func PasteCellFromClipboard(m *model.Model) {
	if m.ViewMode == types.SongView {
		// Paste to song view (chain ID)
		if m.Clipboard.CellType == types.HexCell {
			m.SongData[m.CurrentCol][m.CurrentRow] = m.Clipboard.Value
			log.Printf("Pasted to song track %d row %02X chain: %d", m.CurrentCol, m.CurrentRow, m.Clipboard.Value)
		} else {
			log.Printf("Cannot paste: wrong cell type for song view")
		}
	} else if m.ViewMode == types.ChainView {
		// Paste to chain view (phrase column only)
		if m.Clipboard.CellType == types.HexCell {
			chainsData := m.GetCurrentChainsData()
			(*chainsData)[m.CurrentChain][m.CurrentRow] = m.Clipboard.Value
			log.Printf("Pasted to chain %02X row %02X phrase: %d", m.CurrentChain, m.CurrentRow, m.Clipboard.Value)
		} else {
			log.Printf("Cannot paste: wrong cell type or position")
		}
	} else if m.ViewMode == types.PhraseView {
		// Paste to phrase view
		phrasesData := m.GetCurrentPhrasesData()

		// Use centralized column mapping system
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping == nil || !columnMapping.IsPasteable {
			log.Printf("Cannot paste: invalid or non-pasteable column at position %d", m.CurrentCol)
			return
		}

		colIndex := columnMapping.DataColumnIndex

		if colIndex >= 0 && colIndex < int(types.ColCount) {
			var canPaste bool
			if colIndex == int(types.ColFilename) { // Filename column
				canPaste = m.Clipboard.CellType == types.FilenameCell
			} else { // Playback, note, or deltatime column
				canPaste = m.Clipboard.CellType == types.HexCell
			}

			if canPaste {
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
				log.Printf("Pasted to phrase cell: %d", m.Clipboard.Value)
				// Track this row as the last edited row
				m.LastEditRow = m.CurrentRow
			} else {
				log.Printf("Cannot paste: incompatible cell type")
			}
		}
	}
}

func PasteRowFromClipboard(m *model.Model) {
	if m.ViewMode == types.ChainView && m.Clipboard.SourceView == types.ChainView {
		// Paste chain row to chain row
		// Don't overwrite chain number (column 0)
		m.ChainsData[m.CurrentChain][m.CurrentRow] = m.Clipboard.RowData[m.CurrentRow]
		log.Printf("Pasted chain row to row %d", m.CurrentRow)
	} else if m.ViewMode == types.PhraseView && m.Clipboard.SourceView == types.PhraseView {
		// Paste phrase row to phrase row
		phrasesData := m.GetCurrentPhrasesData()
		for i, value := range m.Clipboard.RowData {
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][i] = value
		}

		// Handle filename if it exists
		if m.Clipboard.RowFilename != "" {
			// Add filename to files array and update index
			fileIndex := m.AppendPhrasesFile(m.Clipboard.RowFilename)
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = fileIndex
		}

		log.Printf("Pasted phrase row to row %d", m.CurrentRow)
		// Track this row as the last edited row
		m.LastEditRow = m.CurrentRow
	} else {
		log.Printf("Cannot paste: incompatible row types")
	}
}

func PasteLastEditedRow(m *model.Model) {
	// Only works if we have a valid last edited row
	if m.LastEditRow == -1 || m.LastEditRow >= 255 {
		log.Printf("No valid last edited row to paste from")
		return
	}

	if m.ViewMode == types.ChainView {
		// Check if current row is empty
		if m.ChainsData[m.CurrentRow][1] != -1 {
			log.Printf("Chain row %d is not empty, skipping paste", m.CurrentRow)
			return
		}
		// Paste chain data
		m.ChainsData[m.CurrentRow][1] = m.ChainsData[m.LastEditRow][1]
		log.Printf("Pasted chain row from %d to %d", m.LastEditRow, m.CurrentRow)
	} else if m.ViewMode == types.PhraseView {
		// Check if current row is empty (note, deltatime, filename are -1, playback is 0)
		if m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColNote] != -1 ||
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] != -1 ||
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColFilename] != -1 {
			log.Printf("Phrase row %d is not empty, skipping paste", m.CurrentRow)
			return
		}

		// Copy all fields from last edited row (including filename index)
		for i := 0; i < int(types.ColCount); i++ {
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][i] = m.PhrasesData[m.CurrentPhrase][m.LastEditRow][i]
		}

		log.Printf("Pasted entire phrase row from %d to %d", m.LastEditRow, m.CurrentRow)
	}

	// Update the last edit row to current row
	m.LastEditRow = m.CurrentRow
}

func ClearClipboardHighlight(m *model.Model) {
	clipboard := m.Clipboard
	clipboard.HasData = false
	m.Clipboard = clipboard
	log.Printf("Cleared clipboard highlight")
}

// IsRowEmpty checks if the current row in phrase view is empty
func IsRowEmpty(m *model.Model) bool {
	if m.ViewMode != types.PhraseView {
		return true
	}

	// Use track-aware data access
	phrasesData := m.GetCurrentPhrasesData()
	rowData := (*phrasesData)[m.CurrentPhrase][m.CurrentRow]

	// A row is considered empty if all key data fields are at their default values
	// For both instrument and sampler tracks, check Note and DeltaTime
	isEmpty := rowData[types.ColNote] == -1 && rowData[types.ColDeltaTime] == -1

	// For sampler tracks, also check filename
	if m.GetPhraseViewType() == types.SamplerPhraseView {
		isEmpty = isEmpty && rowData[types.ColFilename] == -1
	}

	return isEmpty
}

// GetEffectiveValue searches backwards from the current row to find the first non-null value for a given column
func GetEffectiveValue(m *model.Model, phrase, row, colIndex int) int {
	return GetEffectiveValueForTrack(m, phrase, row, colIndex, m.CurrentTrack)
}

// GetEffectiveValueForTrack searches backwards from the current row to find the first non-null value for a given column
func GetEffectiveValueForTrack(m *model.Model, phrase, row, colIndex, trackId int) int {
	// Use track-specific data pool
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	// Search backwards from the given row to find first non-null value
	for r := row; r >= 0; r-- {
		value := (*phrasesData)[phrase][r][colIndex]
		if value != -1 {
			return value
		}
	}
	return -1 // No non-null value found
}

// GetEffectiveMidiValue gets the effective MIDI value for a row (sticky behavior)
func GetEffectiveMidiValue(m *model.Model, phrase, row int) int {
	return GetEffectiveMidiValueForTrack(m, phrase, row, m.CurrentTrack)
}

// GetEffectiveMidiValueForTrack gets the effective MIDI value for a row for a specific track
func GetEffectiveMidiValueForTrack(m *model.Model, phrase, row, trackId int) int {
	return GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
}

// GetEffectiveFilename gets the effective filename for a row
func GetEffectiveFilename(m *model.Model, phrase, row int) string {
	return GetEffectiveFilenameForTrack(m, phrase, row, m.CurrentTrack)
}

// GetEffectiveFilenameForTrack gets the effective filename for a row for a specific track
func GetEffectiveFilenameForTrack(m *model.Model, phrase, row, trackId int) string {
	effectiveFileIndex := GetEffectiveValueForTrack(m, phrase, row, int(types.ColFilename), trackId)

	// Use track-specific files array based on track type
	var phrasesFiles *[]string
	if trackId >= 0 && trackId < 8 && !m.TrackTypes[trackId] {
		// TrackTypes[trackId] = false means Instrument - don't use files
		return "none"
	} else {
		// TrackTypes[trackId] = true means Sampler - use SamplerPhrasesFiles
		phrasesFiles = &m.SamplerPhrasesFiles
	}

	if effectiveFileIndex >= 0 && effectiveFileIndex < len(*phrasesFiles) && (*phrasesFiles)[effectiveFileIndex] != "" {
		return (*phrasesFiles)[effectiveFileIndex]
	}
	return "none"
}

func EmitRowData(m *model.Model) {
	if m.ViewMode != types.PhraseView {
		return
	}
	EmitRowDataFor(m, m.CurrentPhrase, m.CurrentRow, m.CurrentTrack)
}

func EmitLastSelectedPhraseRowData(m *model.Model) {
	EmitRowDataFor(m, m.CurrentPhrase, m.LastPhraseRow, m.CurrentTrack)
}

// EmitRowDataFor logs row data (rich debug) and emits OSC if applicable.
// This is the single canonical emitter used by both manual "c" triggers and playback ("space").
func EmitRowDataFor(m *model.Model, phrase, row, trackId int) {
	log.Printf("DEBUG_EMIT: EmitRowDataFor called with phrase=%d, row=%d, trackId=%d", phrase, row, trackId)

	// Use track-aware data access for correct playback
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	rowData := (*phrasesData)[phrase][row]

	// Raw values - DT used for playback control in both views
	rawNote := rowData[types.ColNote]
	rawPitch := rowData[types.ColPitch]
	rawDeltaTime := rowData[types.ColDeltaTime]
	rawGate := rowData[types.ColGate]
	rawRetrigger := rowData[types.ColRetrigger]
	rawTimestretch := rowData[types.ColTimestretch]
	rawFilenameIndex := rowData[types.ColFilename]

	// Effect columns (may be -1) - get both raw and effective values
	rawEffectReverse := -1
	rawPan := -1
	rawLowPassFilter := -1
	rawHighPassFilter := -1
	rawEffectComb := -1
	rawEffectReverb := -1
	if int(types.ColCount) > int(types.ColEffectReverb) { // guard in case of older saves
		rawEffectReverse = rowData[types.ColEffectReverse]
		rawPan = rowData[types.ColPan]
		rawLowPassFilter = rowData[types.ColLowPassFilter]
		rawHighPassFilter = rowData[types.ColHighPassFilter]
		rawEffectComb = rowData[types.ColEffectComb]
		rawEffectReverb = rowData[types.ColEffectReverb]
	}

	// Get effective values for sticky columns (PA, LP, HP, CO, VE)
	effectivePan := GetEffectiveValueForTrack(m, phrase, row, int(types.ColPan), trackId)
	effectiveLowPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColLowPassFilter), trackId)
	effectiveHighPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColHighPassFilter), trackId)
	effectiveComb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectComb), trackId)
	effectiveReverb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectReverb), trackId)

	// Effective/inherited values
	effectiveNote := GetEffectiveValueForTrack(m, phrase, row, int(types.ColNote), trackId)
	effectiveDeltaTime := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDeltaTime), trackId)
	effectiveFilenameIndex := GetEffectiveValueForTrack(m, phrase, row, int(types.ColFilename), trackId)
	effectiveFilename := GetEffectiveFilenameForTrack(m, phrase, row, trackId)

	// Helper to format hex-ish cells
	formatHex := func(value int) string {
		if value == -1 {
			return "--"
		}
		return fmt.Sprintf("%02X", value)
	}

	// RICH DEBUG LOG (adapted from previous DebugLogRowEmission)
	log.Printf("=== ROW EMISSION ===")
	if m.IsPlaying {
		log.Printf("Context: PlaybackMode=%v Chain:%02X Phrase:%02X Row:%02X IsPlaying=%v BPM=%.2f PPQ=%d",
			m.PlaybackMode, m.PlaybackChain, m.PlaybackPhrase, m.PlaybackRow, m.IsPlaying, m.BPM, m.PPQ)
	} else {
		log.Printf("Context: Manual emit  Phrase:%02X Row:%02X BPM=%.2f PPQ=%d", phrase, row, m.BPM, m.PPQ)
	}
	log.Printf("DeltaTime (playback control): %d", rawDeltaTime)
	// Show different debug info based on track type
	if trackId >= 0 && trackId < 8 && !m.TrackTypes[trackId] {
		// Instrument track - show all instrument parameters
		rawChord := rowData[types.ColChord]
		rawChordAdd := rowData[types.ColChordAddition]
		rawChordTrans := rowData[types.ColChordTransposition]
		rawAttack := rowData[types.ColAttack]
		rawDecay := rowData[types.ColDecay]
		rawSustain := rowData[types.ColSustain]
		rawRelease := rowData[types.ColRelease]
		rawArpeggio := rowData[types.ColArpeggio]
		rawMidi := rowData[types.ColMidi]
		rawSoundMaker := rowData[types.ColSoundMaker]

		log.Printf("Raw:  NN=%s GT=%s C=%d A=%d T=%d A=%s D=%s S=%s R=%s AR=%s MI=%s SO=%s",
			formatHex(rawNote),
			formatHex(rawGate),
			rawChord,
			rawChordAdd,
			rawChordTrans,
			formatHex(rawAttack),
			formatHex(rawDecay),
			formatHex(rawSustain),
			formatHex(rawRelease),
			formatHex(rawArpeggio),
			formatHex(rawMidi),
			formatHex(rawSoundMaker),
		)

		// Show effective values for sticky parameters
		effAttack := GetEffectiveValueForTrack(m, phrase, row, int(types.ColAttack), trackId)
		effDecay := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDecay), trackId)
		effSustain := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSustain), trackId)
		effRelease := GetEffectiveValueForTrack(m, phrase, row, int(types.ColRelease), trackId)
		effArpeggio := GetEffectiveValueForTrack(m, phrase, row, int(types.ColArpeggio), trackId)
		effMidi := GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
		effSoundMaker := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSoundMaker), trackId)

		log.Printf("Eff:  NN=%s A=%s D=%s S=%s R=%s AR=%s MI=%s SO=%s",
			formatHex(effectiveNote),
			formatHex(effAttack),
			formatHex(effDecay),
			formatHex(effSustain),
			formatHex(effRelease),
			formatHex(effArpeggio),
			formatHex(effMidi),
			formatHex(effSoundMaker),
		)
	} else {
		// Sampler track - show traditional sampler information
		log.Printf("Raw:  NN=%s DT=%s GT=%s RT=%s TS=%s FI=%d Я=%s PA=%s LP=%s HP=%s CO=%s VE=%s",
			formatHex(rawNote),
			formatHex(rawDeltaTime),
			formatHex(rawGate),
			formatHex(rawRetrigger),
			formatHex(rawTimestretch),
			rawFilenameIndex,
			func() string {
				if rawEffectReverse == -1 {
					return "-"
				}
				if rawEffectReverse != 0 {
					return "1"
				}
				return "0"
			}(),
			formatHex(rawPan),
			formatHex(rawLowPassFilter),
			formatHex(rawHighPassFilter),
			formatHex(rawEffectComb),
			formatHex(rawEffectReverb),
		)
		log.Printf("Eff:  NN=%s DT=%s FI=%d FN=%s PA=%s LP=%s HP=%s CO=%s VE=%s",
			formatHex(effectiveNote),
			formatHex(effectiveDeltaTime),
			effectiveFilenameIndex,
			effectiveFilename,
			formatHex(effectivePan),
			formatHex(effectiveLowPassFilter),
			formatHex(effectiveHighPassFilter),
			formatHex(effectiveComb),
			formatHex(effectiveReverb),
		)
	}

	// Only emit if we have playback enabled and a concrete note
	// For samplers, also check that we have a filename
	needsFile := trackId >= 0 && trackId < 8 && m.TrackTypes[trackId] // Sampler tracks need files

	// Unified DT-based playback condition: DT > 0 means play for both instruments and samplers
	if !IsRowPlayable(rawDeltaTime) || rawNote == -1 || (needsFile && effectiveFilename == "none") {
		if !IsRowPlayable(rawDeltaTime) {
			log.Printf("ROW_EMIT: skipped (DT<=0, value=%d)", rawDeltaTime)
		} else if rawNote == -1 {
			log.Printf("ROW_EMIT: skipped (NN==null)")
		} else if needsFile && effectiveFilename == "none" {
			log.Printf("ROW_EMIT: skipped (sampler track needs filename)")
		}
		return
	}

	// shouldEmitRow() hook (kept from playback path; e.g., to skip DT==00, etc.)
	if shouldEmitRowForTrackAtPosition(m, phrase, row, trackId) == false {
		log.Printf("ROW_EMIT: skipped by shouldEmitRow()")
		return
	}

	// Build slice/OSC params
	fileMetadata, exists := m.FileMetadata[effectiveFilename]
	sliceCount := 16
	bpmSource := float32(120.0)
	if exists {
		sliceCount = fileMetadata.Slices
		bpmSource = fileMetadata.BPM
	}
	sliceNumber := rawNote % sliceCount

	baseDuration := 1.0 / float32(m.PPQ)
	gateMultiplier := float32(rawGate) / 96.0
	sliceDuration := baseDuration * gateMultiplier

	// Calculate delta time in seconds (time per row * DT)
	deltaTimeSeconds := calculateDeltaTimeSeconds(m, phrase, row, trackId)

	var oscParams model.SamplerOSCParams
	if rawRetrigger != -1 && rawRetrigger >= 0 && rawRetrigger < 255 {
		retriggerSettings := m.RetriggerSettings[rawRetrigger]
		oscParams = model.NewSamplerOSCParamsWithRetrigger(
			effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration,
			retriggerSettings.Times,
			float32(retriggerSettings.Beats),
			retriggerSettings.Start,
			retriggerSettings.End,
			retriggerSettings.PitchChange,
			retriggerSettings.VolumeDB,
			deltaTimeSeconds,
		)
	} else {
		oscParams = model.NewSamplerOSCParams(effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration, deltaTimeSeconds)
	}

	// Pitch conversion from hex to float: 128 (0x80) = 0.0, range 0-254 maps to -24 to +24
	if rawPitch != -1 {
		// Map 0-254 to -24 to +24, with 128 as center (0.0)
		oscParams.Pitch = ((float32(rawPitch) - 128.0) / 128.0) * 24.0
	} else {
		// Default pitch is 0.0 when cleared (-1)
		oscParams.Pitch = 0.0
	}

	// Timestretch
	if rawTimestretch != -1 {
		ts := m.TimestrechSettings[rawTimestretch]
		oscParams.TimestretchStart = float32(ts.Start)
		oscParams.TimestretchEnd = float32(ts.End)
		oscParams.TimestretchBeats = float32(ts.Beats)
	}

	// NEW effect params
	if rawEffectReverse != -1 {
		if rawEffectReverse != 0 {
			oscParams.EffectReverse = 1
		} else {
			oscParams.EffectReverse = 0
		}
	}
	// Pan: Use effective value, map 0-254 to -1.0 to 1.0, with 128 as center (0.0)
	if effectivePan != -1 {
		if effectivePan == 128 {
			oscParams.Pan = 0.0 // Explicit center
		} else {
			oscParams.Pan = (float32(effectivePan) - 127.0) / 127.0
		}
	} else {
		oscParams.Pan = 0.0 // Default to center when no effective value found
	}

	// Low Pass Filter: Use effective value, send 20000 Hz when -1, otherwise exponential mapping
	if effectiveLowPassFilter != -1 {
		// Exponential mapping: 00 -> 20kHz, FE -> 20Hz
		logMin := float32(1.301) // log10(20)
		logMax := float32(4.301) // log10(20000)
		logFreq := logMax - (float32(effectiveLowPassFilter)/254.0)*(logMax-logMin)
		oscParams.LowPassFilter = float32(math.Pow(10, float64(logFreq)))
	} else {
		oscParams.LowPassFilter = 20000.0 // Send 20kHz when no effective value found
	}

	// High Pass Filter: Use effective value, send 20 Hz when -1, otherwise exponential mapping
	if effectiveHighPassFilter != -1 {
		// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
		logMin := float32(1.301) // log10(20)
		logMax := float32(4.301) // log10(20000)
		logFreq := logMin + (float32(effectiveHighPassFilter)/254.0)*(logMax-logMin)
		oscParams.HighPassFilter = float32(math.Pow(10, float64(logFreq)))
	} else {
		oscParams.HighPassFilter = 20.0 // Send 20Hz when no effective value found
	}

	// Comb: Use effective value
	if effectiveComb != -1 {
		oscParams.EffectComb = float32(effectiveComb) / 254.0
	}

	// Reverb: Use effective value
	if effectiveReverb != -1 {
		oscParams.EffectReverb = float32(effectiveReverb) / 254.0
	}

	log.Printf("[EmitRowDataFor] oscParams: %+v", oscParams)

	// Determine track type and emit appropriate message
	if isInstrumentTrack(m, trackId) {
		// For instrument tracks, extract all instrument-specific parameters
		velocity := float32(rawGate) / 128.0 // Convert gate to velocity (0.0-1.0)
		if velocity > 1.0 {
			velocity = 1.0
		}

		// Extract chord parameters
		rawChord := rowData[types.ColChord]
		rawChordAdd := rowData[types.ColChordAddition]
		rawChordTrans := rowData[types.ColChordTransposition]

		// Extract ADSR parameters with effective values (sticky)
		rawAttack := GetEffectiveValueForTrack(m, phrase, row, int(types.ColAttack), trackId)
		rawDecay := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDecay), trackId)
		rawSustain := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSustain), trackId)
		rawRelease := GetEffectiveValueForTrack(m, phrase, row, int(types.ColRelease), trackId)

		// Extract other parameters with effective values (sticky)
		rawArpeggio := GetEffectiveValueForTrack(m, phrase, row, int(types.ColArpeggio), trackId)
		rawMidi := GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
		rawSoundMaker := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSoundMaker), trackId)

		// Extract Gate parameter with effective value (sticky)
		effectiveGate := GetEffectiveValueForTrack(m, phrase, row, int(types.ColGate), trackId)
		if effectiveGate == -1 {
			effectiveGate = 80 // Default Gate value
		}

		// Calculate delta time in seconds (time per row * DT)
		deltaTimeSeconds := calculateDeltaTimeSeconds(m, phrase, row, trackId)

		// Convert ADSR values using the conversion functions from types package
		attack := float32(0.02) // Default
		if rawAttack != -1 {
			attack = types.AttackToSeconds(rawAttack)
		}

		decay := float32(0.0) // Default
		if rawDecay != -1 {
			decay = types.DecayToSeconds(rawDecay)
		}

		sustain := float32(1.0) // Default
		if rawSustain != -1 {
			sustain = types.SustainToLevel(rawSustain)
		}

		release := float32(0.02) // Default
		if rawRelease != -1 {
			release = types.ReleaseToSeconds(rawRelease)
		}

		instrumentParams := model.NewInstrumentOSCParams(
			int32(trackId),
			velocity,
			rawChord,
			rawChordAdd,
			rawChordTrans,
			effectiveGate,
			deltaTimeSeconds,
			attack,
			decay,
			sustain,
			release,
			rawArpeggio,
			rawMidi,
			rawSoundMaker,
		)
		// add notes
		instrumentParams.Notes = make([]float32, 1)
		instrumentParams.Notes[0] = float32(midiNote)
		m.SendOSCInstrumentMessage(instrumentParams)
	} else {
		// For sampler tracks, emit full sampler message
		m.SendOSCSamplerMessage(oscParams)
	}
}

// isInstrumentTrack determines if the given track should use instrument OSC messages
// Uses the track type from mixer settings (false = Instrument, true = Sampler)
func isInstrumentTrack(m *model.Model, trackId int) bool {
	if trackId >= 0 && trackId < 8 {
		return !m.TrackTypes[trackId] // false = Instrument, true = Sampler
	}
	return false // Invalid track defaults to Sampler
}

// rowDurationMS returns the per-row duration in milliseconds.
// DT is read *row-locally* and never inherited.
// New behavior:
// - DT == -1 (--) -> behaves like DT == 1 (1 tick duration)
// - DT == 0       -> skip row (duration irrelevant, handled by shouldEmitRow)
// - DT > 0        -> hold for DT number of ticks (baseMs * DT)
func rowDurationMS(m *model.Model) float64 {
	// Guard against invalid BPM/PPQ
	if m.BPM <= 0 || m.PPQ <= 0 {
		// Fallback to a sane default: 120 BPM, PPQ=2  => 250ms per row
		return 250.0
	}

	beatsPerSecond := float64(m.BPM) / 60.0
	ticksPerSecond := beatsPerSecond * float64(m.PPQ)
	baseMs := 1000.0 / ticksPerSecond

	p := m.PlaybackPhrase
	r := m.PlaybackRow
	// Bounds checks (255 x 255 grid)
	if p < 0 || p >= 255 || r < 0 || r >= 255 {
		return baseMs
	}

	// Get the correct phrases data based on current track type
	phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
	dtRaw := (*phrasesData)[p][r][types.ColDeltaTime] // row-local DT
	if dtRaw == -1 {
		// -- behaves like 01 (1 tick)
		return baseMs
	} else if dtRaw == 0 {
		// 00 means skip row, but we still return base duration for timing consistency
		return baseMs
	} else {
		// DT > 0: hold for DT number of ticks
		return baseMs * float64(dtRaw)
	}
}

// calculateDeltaTimeSeconds calculates the DT value in seconds for a specific phrase/row
// This is the time per row (based on BPM/PPQ) multiplied by the DT value
func calculateDeltaTimeSeconds(m *model.Model, phrase, row, trackId int) float32 {
	// Guard against invalid BPM/PPQ
	if m.BPM <= 0 || m.PPQ <= 0 {
		// Fallback to a sane default: 120 BPM, PPQ=2  => 0.25s per row
		return 0.25
	}

	// Calculate base time per row (tick) in seconds
	beatsPerSecond := float64(m.BPM) / 60.0
	ticksPerSecond := beatsPerSecond * float64(m.PPQ)
	baseSecondsPerTick := 1.0 / ticksPerSecond

	// Bounds checks (255 x 255 grid)
	if phrase < 0 || phrase >= 255 || row < 0 || row >= 255 {
		return float32(baseSecondsPerTick)
	}

	// Get the correct phrases data based on specified track type
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	dtRaw := (*phrasesData)[phrase][row][types.ColDeltaTime] // row-local DT

	if dtRaw == -1 {
		// -- behaves like 01 (1 tick)
		return float32(baseSecondsPerTick)
	} else if dtRaw == 0 {
		// 00 means skip row (0 ticks)
		return 0.0
	} else {
		// DT > 0: hold for DT number of ticks
		return float32(baseSecondsPerTick * float64(dtRaw))
	}
}

// shouldEmitRow enforces:
// - NN must be present (not -1)
// - P (playback flag) must be 1
// - DT == 00 => do NOT emit (for all track types)
// DT is read row-locally and never inherited.
func shouldEmitRow(m *model.Model) bool {
	return shouldEmitRowForTrack(m, m.CurrentTrack)
}

func shouldEmitRowForTrack(m *model.Model, trackId int) bool {
	p := m.PlaybackPhrase
	r := m.PlaybackRow
	if p < 0 || p >= 255 || r < 0 || r >= 255 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, trackId)
	nn := (*phrasesData)[p][r][types.ColNote]         // may still be inherited elsewhere in your code
	dtRaw := (*phrasesData)[p][r][types.ColDeltaTime] // unified playback control

	// Unified DT-based playback: DT > 0 means play, DT <= 0 means don't play
	if !IsRowPlayable(dtRaw) {
		return false
	}

	if nn == -1 {
		return false
	}

	return true
}

func shouldEmitRowForTrackAtPosition(m *model.Model, phrase, row, trackId int) bool {
	if phrase < 0 || phrase >= 255 || row < 0 || row >= 255 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, trackId)
	nn := (*phrasesData)[phrase][row][types.ColNote]
	dtRaw := (*phrasesData)[phrase][row][types.ColDeltaTime]

	// Unified DT-based playback: DT > 0 means play, DT <= 0 means don't play
	if !IsRowPlayable(dtRaw) {
		return false
	}
	if nn == -1 {
		return false
	}
	return true
}

func goToPhrase(m *model.Model, phrase int) {
	if phrase < 0 || phrase >= 255 {
		return
	}
	m.CurrentPhrase = phrase
	// Reset to beginning of phrase if it hasn't been visited/edited
	if m.LastEditRow == -1 || len(m.PhrasesData[phrase]) == 0 {
		m.CurrentRow = 0
		m.CurrentCol = 1
		m.ScrollOffset = 0
	} else {
		// keep last edit row if desired, but usually we just reset
		m.CurrentRow = 0
		m.CurrentCol = 1
		m.ScrollOffset = 0
	}
	storage.AutoSave(m)
}

// ModifySongValue modifies chain ID values in song view
func ModifySongValue(m *model.Model, delta int) {
	track := m.CurrentCol
	row := m.CurrentRow

	// Bounds check
	if track < 0 || track >= 8 || row < 0 || row >= 16 {
		return
	}

	currentValue := m.SongData[track][row]
	var newValue int

	if currentValue == -1 {
		// First edit on an empty cell: initialize to 00 and DO NOT apply delta
		newValue = 0
	} else {
		newValue = currentValue + delta
	}

	// Clamp to valid chain range (0-254, which is 00-FE in hex)
	if newValue < 0 {
		newValue = 0
	} else if newValue > 254 {
		newValue = 254
	}

	m.SongData[track][row] = newValue
	log.Printf("Modified song track %d row %d: %d -> %d (delta: %d)", track, row, currentValue, newValue, delta)

	storage.AutoSave(m)
}

// PlaybackConfig represents the configuration for starting playback
type PlaybackConfig struct {
	Mode          types.ViewMode
	UseCurrentRow bool // false means start from top/first non-empty
	Chain         int  // for chain playback
	Phrase        int  // for phrase playback
	Row           int  // starting row
}

// stopPlayback provides common logic for stopping playback
func stopPlayback(m *model.Model) {
	m.IsPlaying = false

	// Stop recording if active
	if m.RecordingActive {
		stopRecording(m)
	}

	// Clear file browser playback state when stopping tracker playback
	if m.CurrentlyPlayingFile != "" {
		m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
		m.CurrentlyPlayingFile = ""
		log.Printf("Stopped file browser playback when stopping tracker playback")
	}

	m.SendStopOSC()
	log.Printf("Playback stopped")
}

// startPlaybackWithConfig provides common logic for starting playback
func startPlaybackWithConfig(m *model.Model, config PlaybackConfig) tea.Cmd {
	m.IsPlaying = true
	m.PlaybackMode = config.Mode

	if config.Mode == types.SongView {
		// Song playback mode - reset single-track playback variables and initialize all tracks with data
		m.PlaybackPhrase = -1
		m.PlaybackRow = -1
		m.PlaybackChain = -1
		m.PlaybackChainRow = -1

		startRow := 0
		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			startRow = config.Row
		}
		log.Printf("Song playback starting from row %02X", startRow)
		// Debug: show song data for first few rows
		for r := 0; r < 4 && r < 16; r++ {
			log.Printf("Song row %02X data: %v", r, [8]int{
				m.SongData[0][r], m.SongData[1][r], m.SongData[2][r], m.SongData[3][r],
				m.SongData[4][r], m.SongData[5][r], m.SongData[6][r], m.SongData[7][r],
			})
		}

		for track := 0; track < 8; track++ {
			chainID := m.SongData[track][startRow]
			log.Printf("Song track %d at row %02X: chainID = %d", track, startRow, chainID)
			if chainID == -1 {
				// No chain at this position
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d: no chain data, skipping", track)
				continue
			}

			// Check if chain has valid phrase data (find first phrase in chain)
			firstPhraseID := -1
			firstChainRow := -1
			chainsData := m.GetChainsDataForTrack(track)
			for chainRow := 0; chainRow < 16; chainRow++ {
				if (*chainsData)[chainID][chainRow] != -1 {
					firstPhraseID = (*chainsData)[chainID][chainRow]
					firstChainRow = chainRow
					log.Printf("Song track %d: found phrase %d in chain %d at row %d", track, firstPhraseID, chainID, chainRow)
					break
				}
			}

			if firstPhraseID != -1 {
				m.SongPlaybackActive[track] = true
				m.SongPlaybackRow[track] = startRow
				m.SongPlaybackChain[track] = chainID
				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)

				// Initialize ticks for this track
				m.LoadTicksLeftForTrack(track)

				// Emit initial row for this track
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d started at row %02X, chain %02X (chain row %d), phrase %02X with %d ticks", track, startRow, chainID, firstChainRow, firstPhraseID, m.SongPlaybackTicksLeft[track])
			} else {
				// Chain exists but has no phrases
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d skipped at row %02X (chain %02X has no phrases)", track, startRow, chainID)
			}
		}

		// Count how many tracks will be active
		activeTracks := 0
		for t := 0; t < 8; t++ {
			if m.SongPlaybackActive[t] {
				activeTracks++
			}
		}
		log.Printf("Song playback started from row %02X with %d active tracks", startRow, activeTracks)

		// If no tracks are active, make track 0 active with chain 0 as fallback
		if activeTracks == 0 {
			log.Printf("No active tracks found, activating track 0 with chain 0 as fallback")
			m.SongPlaybackActive[0] = true
			m.SongPlaybackRow[0] = startRow
			m.SongPlaybackChain[0] = 0
			m.SongPlaybackChainRow[0] = 0
			m.SongPlaybackPhrase[0] = 0 // Use phrase 0 as fallback
			m.SongPlaybackRowInPhrase[0] = FindFirstNonEmptyRowInPhraseForTrack(m, 0, 0)
			// Initialize ticks for fallback track 0
			m.LoadTicksLeftForTrack(0)
			EmitRowDataFor(m, 0, m.SongPlaybackRowInPhrase[0], 0)
			log.Printf("Song track 0 fallback started at phrase 0, row %d with %d ticks", m.SongPlaybackRowInPhrase[0], m.SongPlaybackTicksLeft[0])
		}
	} else if config.Mode == types.ChainView {
		// Chain playback mode - find appropriate starting phrase
		m.PlaybackChain = config.Chain
		m.PlaybackPhrase = -1

		chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			// Start from specified chain row
			if (*chainsData)[config.Chain][config.Row] != -1 {
				m.PlaybackChainRow = config.Row
				m.PlaybackPhrase = (*chainsData)[config.Chain][config.Row]
			}
		}

		// If no phrase found yet, find first non-empty phrase slot in this chain
		if m.PlaybackPhrase == -1 {
			for row := 0; row < 16; row++ {
				if (*chainsData)[config.Chain][row] != -1 {
					m.PlaybackPhrase = (*chainsData)[config.Chain][row]
					m.PlaybackChainRow = row
					break
				}
			}
		}

		if m.PlaybackPhrase == -1 {
			// This chain has no phrases, find a different chain
			m.PlaybackChain = FindFirstNonEmptyChain(m)
			log.Printf("Chain playback fallback: switching to chain %d", m.PlaybackChain)
			m.PlaybackChainRow = 0
			for row := 0; row < 16; row++ {
				if (*chainsData)[m.PlaybackChain][row] != -1 {
					m.PlaybackPhrase = (*chainsData)[m.PlaybackChain][row]
					m.PlaybackChainRow = row
					log.Printf("Chain playback fallback: found phrase %d at chain row %d", m.PlaybackPhrase, row)
					break
				}
			}
		}

		// If still no valid phrase found, just start with phrase 0 as fallback
		if m.PlaybackPhrase == -1 {
			log.Printf("Chain playback warning - no valid phrases found, using phrase 0 as fallback (Chain: %d, ChainRow: %d)", m.PlaybackChain, m.PlaybackChainRow)
			// Let's log the chain data for debugging
			log.Printf("Chain %d contents: %v", m.PlaybackChain, (*chainsData)[m.PlaybackChain])
			m.PlaybackPhrase = 0
			m.PlaybackChainRow = 0
		}

		if config.UseCurrentRow && config.Row >= 0 {
			m.PlaybackRow = config.Row
		} else {
			m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
		}

		DebugLogRowEmission(m)
		log.Printf("Chain playback started at chain %d, chain row %d, phrase %d, phrase row %d", m.PlaybackChain, m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
	} else {
		// Phrase playback mode
		m.PlaybackPhrase = config.Phrase
		m.PlaybackChain = -1

		trackType := "Sampler"
		if m.CurrentTrack >= 0 && m.CurrentTrack < 8 && !m.TrackTypes[m.CurrentTrack] {
			trackType = "Instrument"
		}
		log.Printf("DEBUG: Phrase playback starting - CurrentTrack=%d (%s), Phrase=%d", m.CurrentTrack, trackType, m.PlaybackPhrase)

		if config.UseCurrentRow && config.Row >= 0 {
			m.PlaybackRow = config.Row
		} else {
			m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
			log.Printf("DEBUG: FindFirstNonEmptyRowInPhrase returned row %d for track %d", m.PlaybackRow, m.CurrentTrack)
		}

		DebugLogRowEmission(m)
		log.Printf("Phrase playback started at phrase %d, row %d", m.PlaybackPhrase, m.PlaybackRow)
	}

	// Start recording if enabled
	if m.RecordingEnabled && !m.RecordingActive {
		startRecording(m)
	}

	return Tick(m)
}

// togglePlaybackWithConfig provides common toggle logic
func togglePlaybackWithConfig(m *model.Model, config PlaybackConfig) tea.Cmd {
	if m.IsPlaying {
		stopPlayback(m)
		return nil
	}
	return startPlaybackWithConfig(m, config)
}

// Deep copy functionality for Ctrl+B
func DeepCopyToClipboard(m *model.Model) {
	if m.ViewMode == types.SongView {
		DeepCopyChainToClipboard(m)
	} else if m.ViewMode == types.ChainView {
		DeepCopyPhraseToClipboard(m)
	} else {
		log.Printf("Deep copy not supported in this view")
	}
}

func DeepCopyChainToClipboard(m *model.Model) {
	sourceChainID := m.SongData[m.CurrentCol][m.CurrentRow]
	if sourceChainID == -1 {
		log.Printf("Cannot deep copy: no chain at track %d row %02X (cell is empty)", m.CurrentCol, m.CurrentRow)
		return
	}

	// Find next unused chain
	destChainID := FindNextUnusedChain(m, sourceChainID)
	if destChainID == -1 {
		log.Printf("Cannot deep copy: no unused chains available")
		return
	}

	// Copy all 16 phrase slots from source chain to destination chain
	for row := 0; row < 16; row++ {
		m.ChainsData[destChainID][row] = m.ChainsData[sourceChainID][row]
	}

	// Put the new chain ID in clipboard
	clipboard := types.ClipboardData{
		Value:           destChainID,
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: -1,
		HighlightView:   types.SongView,
	}
	m.Clipboard = clipboard

	log.Printf("Deep copied chain %02X to chain %02X", sourceChainID, destChainID)
}

func DeepCopyPhraseToClipboard(m *model.Model) {
	// Bounds check for current row
	if m.CurrentRow < 0 || m.CurrentRow >= 16 {
		log.Printf("Cannot deep copy: invalid row %d", m.CurrentRow)
		return
	}

	// In Chain view, always use the phrase from the current row, regardless of column
	sourcePhraseID := m.ChainsData[m.CurrentRow][1]
	if sourcePhraseID == -1 {
		log.Printf("Cannot deep copy: no phrase at chain %d row %d (cell is empty)", m.CurrentChain, m.CurrentRow)
		return
	}

	// Additional bounds check for source phrase ID
	if sourcePhraseID < 0 || sourcePhraseID >= 255 {
		log.Printf("Cannot deep copy: invalid source phrase ID %d", sourcePhraseID)
		return
	}

	// Find next unused phrase
	destPhraseID := FindNextUnusedPhrase(m, sourcePhraseID)
	if destPhraseID == -1 {
		log.Printf("Cannot deep copy: no unused phrases available")
		return
	}

	// Additional bounds check for destination phrase ID
	if destPhraseID < 0 || destPhraseID >= 255 {
		log.Printf("Cannot deep copy: invalid destination phrase ID %d", destPhraseID)
		return
	}

	// Copy all 255 rows of phrase data from source to destination
	for row := 0; row < 255; row++ {
		for col := 0; col < int(types.ColCount); col++ {
			m.PhrasesData[destPhraseID][row][col] = m.PhrasesData[sourcePhraseID][row][col]
		}
	}

	// Put the new phrase ID in clipboard
	clipboard := types.ClipboardData{
		Value:           destPhraseID,
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: -1,
		HighlightView:   types.ChainView,
	}
	m.Clipboard = clipboard

	log.Printf("Deep copied phrase %02X to phrase %02X", sourcePhraseID, destPhraseID)
}

func FindNextUnusedChain(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		chainID := (startingFrom + offset) % 255
		if chainID >= 0 && chainID < 255 && IsChainUnused(m, chainID) {
			return chainID
		}
	}
	return -1 // No unused chain found
}

func FindNextUnusedPhrase(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		phraseID := (startingFrom + offset) % 255

		if phraseID >= 0 && phraseID < 255 && IsPhraseUnused(m, phraseID) {
			return phraseID
		}
	}
	return -1 // No unused phrase found
}

func IsChainUnused(m *model.Model, chainID int) bool {
	// Bounds check first
	if chainID < 0 || chainID >= 255 {
		return false
	}

	// Check if chain is referenced in song data
	for track := 0; track < 8; track++ {
		for row := 0; row < 16; row++ {
			if m.SongData[track][row] == chainID {
				return false
			}
		}
	}

	// Check if chain has any phrase data
	for row := 0; row < 16; row++ {
		if m.ChainsData[chainID][row] != -1 {
			return false
		}
	}

	return true
}

func IsPhraseUnused(m *model.Model, phraseID int) bool {
	// Bounds check first
	if phraseID < 0 || phraseID >= 255 {
		return false
	}

	// Check if phrase is referenced in any chain
	for chain := 0; chain < 255; chain++ { // ChainsData has 255 elements (0-254)
		for row := 0; row < 16; row++ {
			if m.ChainsData[chain][row] == phraseID {
				return false
			}
		}
	}

	// Check if phrase has any playback-enabled rows
	phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
	for row := 0; row < 255; row++ {
		// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
		dtValue := (*phrasesData)[phraseID][row][types.ColDeltaTime]
		if IsRowPlayable(dtValue) {
			return false
		}
	}

	return true
}

// ModifyMixerSetLevel adjusts the set level for the currently selected track in mixer view
func ModifyMixerSetLevel(m *model.Model, delta float32) {
	// Bounds check
	if m.CurrentMixerTrack < 0 || m.CurrentMixerTrack >= 8 {
		return
	}

	oldValue := m.TrackSetLevels[m.CurrentMixerTrack]
	newValue := oldValue + delta

	// Clamp to valid range (-96.0 to +32.0 dB)
	if newValue < -96.0 {
		newValue = -96.0
	} else if newValue > 32.0 {
		newValue = 32.0
	}

	m.TrackSetLevels[m.CurrentMixerTrack] = newValue
	log.Printf("Modified mixer track %d set level: %.2f -> %.2f (delta: %.2f)", m.CurrentMixerTrack+1, oldValue, newValue, delta)

	// Send OSC message for track set level
	m.SendOSCTrackSetLevelMessage(m.CurrentMixerTrack)

	storage.AutoSave(m)
}

// ToggleTrackType toggles the track type for the specified track (used in Song view)
func ToggleTrackType(m *model.Model, track int) {
	// Bounds check
	if track < 0 || track >= 8 {
		return
	}

	// Toggle the track type
	oldType := m.TrackTypes[track]
	m.TrackTypes[track] = !oldType

	var oldTypeStr, newTypeStr string
	if oldType {
		oldTypeStr = "SA"
	} else {
		oldTypeStr = "IN"
	}
	if m.TrackTypes[track] {
		newTypeStr = "SA"
	} else {
		newTypeStr = "IN"
	}

	log.Printf("Toggled track %d type: %s -> %s", track, oldTypeStr, newTypeStr)
	storage.AutoSave(m)
}

// FillSequential fills from the last null cell to the current cell in increments of 1
func FillSequential(m *model.Model) {
	if m.ViewMode == types.SongView {
		FillSequentialSong(m)
	} else if m.ViewMode == types.ChainView {
		FillSequentialChain(m)
	} else if m.ViewMode == types.PhraseView {
		FillSequentialPhrase(m)
	}
}

// FillSequentialSong fills chain IDs in song view
func FillSequentialSong(m *model.Model) {
	track := m.CurrentCol
	currentRow := m.CurrentRow

	// Find the last non-null value going upward
	var startValue int = 0
	var startRow int = 0

	for row := currentRow - 1; row >= 0; row-- {
		if m.SongData[track][row] != -1 {
			startValue = m.SongData[track][row] + 1
			startRow = row + 1
			break
		}
	}

	// Fill from startRow to currentRow
	for row := startRow; row <= currentRow; row++ {
		value := startValue + (row - startRow)
		if value > 254 {
			value = value % 255 // Wrap around (0-254)
		}
		m.SongData[track][row] = value
	}

	log.Printf("Filled song track %d from row %d to %d, starting with %d", track, startRow, currentRow, startValue)
}

// FillSequentialChain fills phrase IDs in chain view
func FillSequentialChain(m *model.Model) {
	currentRow := m.CurrentRow

	// Find the last non-null value going upward
	var startValue int = 0
	var startRow int = 0

	for row := currentRow - 1; row >= 0; row-- {
		if m.ChainsData[m.CurrentChain][row] != -1 {
			startValue = m.ChainsData[m.CurrentChain][row] + 1
			startRow = row + 1
			break
		}
	}

	// Fill from startRow to currentRow
	for row := startRow; row <= currentRow; row++ {
		value := startValue + (row - startRow)
		if value > 254 {
			value = value % 255 // Wrap around (0-254)
		}
		m.ChainsData[m.CurrentChain][row] = value
	}

	log.Printf("Filled chain %02X from row %d to %d, starting with %d", m.CurrentChain, startRow, currentRow, startValue)
}

// FillSequentialPhrase fills values in phrase view for the current column
func FillSequentialPhrase(m *model.Model) {
	currentRow := m.CurrentRow

	// Use centralized column mapping system
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	if columnMapping == nil || !columnMapping.IsEditable {
		return // Invalid or non-editable column
	}

	colIndex := columnMapping.DataColumnIndex

	// Find the last non-null value going upward
	var startValue int = 0
	var startRow int = 0

	// Use current phrases data based on view type
	phrasesData := m.GetCurrentPhrasesData()

	for row := currentRow - 1; row >= 0; row-- {
		cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
		if cellValue != -1 {
			startValue = cellValue + 1
			startRow = row + 1
			break
		}
	}

	// Handle different column types
	if colIndex == int(types.ColDeltaTime) {
		// Special Ctrl+F logic for DT column
		currentValue := (*phrasesData)[m.CurrentPhrase][currentRow][colIndex]

		if currentValue <= 0 {
			// Current cell is "--" (value <= 0): Find last non-"--" value and copy it down
			lastNonEmptyValue := 1 // Default to 1 if no previous value found
			fillStartRow := 0

			// Find the last non-"--" value going upward
			for row := currentRow - 1; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue > 0 {
					lastNonEmptyValue = cellValue
					fillStartRow = row + 1
					break
				}
			}

			// Fill from fillStartRow to currentRow with lastNonEmptyValue
			for row := fillStartRow; row <= currentRow; row++ {
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = lastNonEmptyValue
			}
		} else {
			// Current cell is "XX" (value > 0): Switch XX to "--" on current and all above until first non-XX
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != currentValue {
					break // Stop at first cell that doesn't match current value
				}
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = -1 // Set to "--"
			}
		}
	} else if colIndex == int(types.ColEffectReverse) {
		// Binary columns (0-1) - keep original logic for other binary columns
		maxValue := 1
		for row := startRow; row <= currentRow; row++ {
			value := startValue + (row - startRow)
			if value > maxValue {
				value = value % (maxValue + 1) // Wrap: 0, 1, 0, 1, ...
			}
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
		}
	} else if colIndex == int(types.ColPitch) {
		// Pitch column has default of 128, not -1
		if startValue == 0 && startRow == 0 {
			// Special case: if all above are empty, start with default pitch (128)
			startValue = 128
		}
		// Hex columns (0-254)
		maxValue := 254
		for row := startRow; row <= currentRow; row++ {
			value := startValue + (row - startRow)
			if value > maxValue {
				value = value % (maxValue + 1) // Wrap: 0-254, 0-254, ...
			}
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
		}
	} else {
		// Hex columns (0-254)
		maxValue := 254
		for row := startRow; row <= currentRow; row++ {
			value := startValue + (row - startRow)
			if value > maxValue {
				value = value % (maxValue + 1) // Wrap: 0-254, 0-254, ...
			}
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
		}
	}

	// Auto-enable playback when filling note column (NN or NOT)
	if colIndex == int(types.ColNote) {
		for row := startRow; row <= currentRow; row++ {
			// Use DT column for both views (only if not already set)
			if (*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime] == -1 {
				(*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime] = 1
			}
		}
	}

	// Update last edit row
	m.LastEditRow = currentRow

	log.Printf("Filled phrase %02X column %d from row %d to %d, starting with %d", m.CurrentPhrase, colIndex, startRow, currentRow, startValue)
}

// Shared DT (Delta Time) utility functions for both Sampler and Instrument views
// DT controls playback: >0 = play for N ticks, 0 = skip, -1 = skip

// IsRowPlayable checks if a row should be played based on its DT value
func IsRowPlayable(dtValue int) bool {
	return dtValue > 0
}

// GetEffectiveDTValue gets the effective DT value for a row (for display/status)
func GetEffectiveDTValue(dtValue int) string {
	if dtValue == -1 {
		return "--"
	}
	return fmt.Sprintf("%02X", dtValue)
}

// SetDTForPlayback sets DT to 01 (default playback value) for a row
func SetDTForPlayback(phrasesData *[255][][]int, phrase, row int) {
	(*phrasesData)[phrase][row][types.ColDeltaTime] = 1
}

// GetDTStatusMessage returns a status message for DT column
func GetDTStatusMessage(dtValue int) string {
	if dtValue == -1 {
		return "Delta Time: -- (row not played)"
	} else if dtValue == 0 {
		return "Delta Time: 00 (row not played)"
	} else {
		return fmt.Sprintf("Delta Time: %02X (%d ticks, row played)", dtValue, dtValue)
	}
}

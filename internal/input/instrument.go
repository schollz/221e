package input

import (
	"log"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/types"
)

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

// ClearArpeggioCell clears the current cell in Arpeggio Settings view
func ClearArpeggioCell(m *model.Model) {
	if m.ViewMode != types.ArpeggioView {
		return
	}

	if m.ArpeggioEditingIndex < 0 || m.ArpeggioEditingIndex >= 255 {
		return
	}
	if m.CurrentRow < 0 || m.CurrentRow >= 16 {
		return
	}

	// Get current row
	currentRow := &m.ArpeggioSettings[m.ArpeggioEditingIndex].Rows[m.CurrentRow]

	switch m.CurrentCol {
	case 0: // DI (Direction) column
		currentRow.Direction = 0 // Clear to "--"
		log.Printf("Cleared arpeggio %02X row %02X Direction", m.ArpeggioEditingIndex, m.CurrentRow)
	case 1: // CO (Count) column
		currentRow.Count = -1 // Clear to "--"
		log.Printf("Cleared arpeggio %02X row %02X Count", m.ArpeggioEditingIndex, m.CurrentRow)
	case 2: // Divisor (/) column
		currentRow.Divisor = -1 // Clear to "--"
		log.Printf("Cleared arpeggio %02X row %02X Divisor", m.ArpeggioEditingIndex, m.CurrentRow)
	}
	storage.AutoSave(m)
}

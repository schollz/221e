package input

import (
	"log"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/supercollider"
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
		// Direction cycles through: ArpeggioDirectionNone, ArpeggioDirectionUp, ArpeggioDirectionDown
		var delta int
		if baseDelta > 0 {
			delta = 1
		} else {
			delta = -1
		}

		newDirection := currentRow.Direction + delta
		if newDirection < int(types.ArpeggioDirectionNone) {
			newDirection = int(types.ArpeggioDirectionNone) // Stay at "--"
		} else if newDirection > int(types.ArpeggioDirectionDown) {
			newDirection = int(types.ArpeggioDirectionDown) // Stay at "d-"
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

		// Apply delta with clamping (no wrapping)
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = 0 // Stay at first device
		} else if newIndex >= len(devices) {
			newIndex = len(devices) - 1 // Stay at last device
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

		// Apply delta with clamping (no wrapping)
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = 0 // Stay at "1"
		} else if newIndex >= len(channels) {
			newIndex = len(channels) - 1 // Stay at "all"
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
		soundMakers := []string{"None", "PolyPerc", "DX7", "MiBraids", "MiPlaits"}

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

		// Apply delta with clamping (no wrapping)
		newIndex := currentIndex + delta
		if newIndex < 0 {
			newIndex = 0 // Stay at first SoundMaker
		} else if newIndex >= len(soundMakers) {
			newIndex = len(soundMakers) - 1 // Stay at last SoundMaker
		}

		oldName := settings.Name
		settings.Name = soundMakers[newIndex]

		// Initialize parameters for the new instrument
		settings.InitializeParameters()

		log.Printf("Modified SoundMaker %02X Name: %s -> %s", m.SoundMakerEditingIndex, oldName, settings.Name)
	} else {
		// Handle parameter modification using the instrument framework
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			// Parameter rows start at row 1
			paramIndex := m.CurrentRow - 1
			if paramIndex >= 0 && paramIndex < len(def.Parameters) {
				param := def.Parameters[paramIndex]
				oldValue := settings.GetParameterValue(param.Key)

				// Calculate delta based on parameter type and input
				var delta int
				if param.Type == types.ParameterTypeInt {
					// For DX7 preset, use larger steps
					if baseDelta >= 1.0 || baseDelta <= -1.0 {
						// Coarse control
						if baseDelta > 0 {
							delta = 50
						} else {
							delta = -50
						}
					} else {
						// Fine control
						if baseDelta > 0 {
							delta = 1
						} else {
							delta = -1
						}
					}
				} else {
					// For hex values, use standard steps
					if baseDelta >= 1.0 || baseDelta <= -1.0 {
						// Coarse control
						if baseDelta > 0 {
							delta = 16
						} else {
							delta = -16
						}
					} else {
						// Fine control
						if baseDelta > 0 {
							delta = 1
						} else {
							delta = -1
						}
					}
				}

				var newValue int
				if oldValue == -1 {
					// If currently "--", start from min or max
					if delta > 0 {
						newValue = param.MinValue
					} else {
						newValue = param.MaxValue
					}
				} else {
					newValue = oldValue + delta
					// Handle clamping (no wrapping)
					if newValue > param.MaxValue {
						newValue = param.MaxValue // Clamp to max
					} else if newValue < param.MinValue {
						newValue = param.MinValue // Clamp to min
					}
				}

				// Set the new value
				settings.SetParameterValue(param.Key, newValue)

				// Special handling for DX7 patch name updates
				if param.Key == "preset" && settings.Name == "DX7" {
					if newValue >= 0 {
						if patchName, err := supercollider.GetDX7PatchName(newValue); err == nil {
							settings.PatchName = patchName
						}
					} else {
						settings.PatchName = ""
					}
				}

				// Log the change
				if newValue == -1 {
					if oldValue == -1 {
						log.Printf("Modified SoundMaker %02X %s: -- -> -- (delta: %d)", m.SoundMakerEditingIndex, param.DisplayName, delta)
					} else {
						log.Printf("Modified SoundMaker %02X %s: %d -> -- (delta: %d)", m.SoundMakerEditingIndex, param.DisplayName, oldValue, delta)
					}
				} else {
					if oldValue == -1 {
						log.Printf("Modified SoundMaker %02X %s: -- -> %d (delta: %d)", m.SoundMakerEditingIndex, param.DisplayName, newValue, delta)
					} else {
						log.Printf("Modified SoundMaker %02X %s: %d -> %d (delta: %d)", m.SoundMakerEditingIndex, param.DisplayName, oldValue, newValue, delta)
					}
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
		currentRow.Direction = int(types.ArpeggioDirectionNone) // Clear to "--"
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

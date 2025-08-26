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

// GetPhrasesDataForTrack returns the appropriate phrases data based on track number
func GetPhrasesDataForTrack(m *model.Model, track int) *[255][][]int {
	if track >= 0 && track <= 3 {
		return &m.InstrumentPhrasesData
	}
	return &m.SamplerPhrasesData
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

	if colIndex == int(types.ColPlayback) {
		// P column: clamp 0..1 (keep existing behavior)
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
				if delta == 4 || delta == -4 {
					// This is coarse increment - convert to octave increment (12 semitones)
					octaveDelta := (delta / 4) * 12
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
			
			// Auto-set P=1 when any note is added in Instrument view
			if newValue != -1 {
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 1
			}
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChord) {
			// Instrument view chord column: Cycle through chord types
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordNone if unset
				currentValue = int(types.ChordNone)
			}

			// Determine direction based on delta (Ctrl+Up/Right = forward, Ctrl+Down/Left = backward)
			if delta > 0 {
				// Forward cycling: "-" -> "M" -> "m" -> "d" -> back to "-"
				newValue = (currentValue + 1) % int(types.ChordTypeCount)
			} else {
				// Backward cycling: "d" -> "m" -> "M" -> "-" -> back to "d"
				newValue = (currentValue - 1 + int(types.ChordTypeCount)) % int(types.ChordTypeCount)
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordAddition) {
			// Instrument view chord addition column: Cycle through addition types
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordAddNone if unset
				currentValue = int(types.ChordAddNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward cycling: "-" -> "7" -> "9" -> "4" -> back to "-"
				newValue = (currentValue + 1) % int(types.ChordAdditionCount)
			} else {
				// Backward cycling: "4" -> "9" -> "7" -> "-" -> back to "4"
				newValue = (currentValue - 1 + int(types.ChordAdditionCount)) % int(types.ChordAdditionCount)
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordTransposition) {
			// Instrument view chord transposition column: Cycle through transposition values
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordTransNone if unset
				currentValue = int(types.ChordTransNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward cycling: "-" -> "0" -> "1" -> ... -> "F" -> back to "-"
				newValue = (currentValue + 1) % int(types.ChordTranspositionCount)
			} else {
				// Backward cycling: "F" -> "E" -> ... -> "0" -> "-" -> back to "F"
				newValue = (currentValue - 1 + int(types.ChordTranspositionCount)) % int(types.ChordTranspositionCount)
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

		// Auto-enable playback on first note entry
		if colIndex == int(types.ColNote) && (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColPlayback] == 0 {
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColPlayback] = 1
			log.Printf("Auto-enabled playback for phrase %d row %d due to note change", m.CurrentPhrase, m.CurrentRow)
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

func Tick(m *model.Model) tea.Cmd {
	ms := rowDurationMS(m)
	return tea.Tick(time.Duration(ms)*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AdvancePlayback(m *model.Model) {
	oldRow := m.PlaybackRow

	if m.PlaybackMode == types.SongView {
		// Song playback mode - advance each active track independently
		log.Printf("Song playback advancing - checking %d tracks", 8)
		activeTrackCount := 0
		for track := 0; track < 8; track++ {
			if !m.SongPlaybackActive[track] {
				continue
			}
			activeTrackCount++

			oldPhraseRow := m.SongPlaybackRowInPhrase[track]

			// Declare variables early to avoid goto issues
			var chainID int
			var nextPhraseFound bool
			var currentChain int
			var chainsData *[][]int

			// Advance within current phrase
			phraseNum := m.SongPlaybackPhrase[track]
			if phraseNum >= 0 && phraseNum < 255 {
				// Find next row with playback enabled in current phrase
				phrasesData := GetPhrasesDataForTrack(m, track)
				for i := m.SongPlaybackRowInPhrase[track] + 1; i < 255; i++ {
					if (*phrasesData)[phraseNum][i][types.ColPlayback] == 1 {
						m.SongPlaybackRowInPhrase[track] = i
						EmitRowDataFor(m, phraseNum, i, track)
						log.Printf("Song track %d advanced within phrase from row %d to %d", track, oldPhraseRow, i)
						goto nextTrack
					}
				}
			}

			// End of phrase reached, advance within current chain first
			currentChain = m.SongPlaybackChain[track]

			// Try to find next phrase slot in current chain
			nextPhraseFound = false
			chainsData = m.GetChainsDataForTrack(track)
			for chainRow := m.SongPlaybackChainRow[track] + 1; chainRow < 16; chainRow++ {
				if (*chainsData)[currentChain][chainRow] != -1 {
					// Found next phrase in chain
					m.SongPlaybackChainRow[track] = chainRow
					m.SongPlaybackPhrase[track] = (*chainsData)[currentChain][chainRow]
					m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, m.SongPlaybackPhrase[track], track)
					EmitRowDataFor(m, m.SongPlaybackPhrase[track], m.SongPlaybackRowInPhrase[track], track)
					log.Printf("Song track %d advanced within chain to chain row %d, phrase %02X", track, chainRow, m.SongPlaybackPhrase[track])
					nextPhraseFound = true
					break
				}
			}

			if !nextPhraseFound {
				// End of chain reached, find next valid song row
				startSearchRow := m.SongPlaybackRow[track] + 1
				foundValidChain := false

				// Search from next row onwards, wrapping around to beginning if needed
				for searchOffset := 0; searchOffset < 16; searchOffset++ {
					searchRow := (startSearchRow + searchOffset) % 16
					chainID = m.SongData[track][searchRow]

					if chainID != -1 {
						// Found valid chain, check if it has phrases
						firstPhraseID := -1
						for chainRow := 0; chainRow < 16; chainRow++ {
							if (*chainsData)[chainID][chainRow] != -1 {
								firstPhraseID = (*chainsData)[chainID][chainRow]
								break
							}
						}

						if firstPhraseID != -1 {
							// Found valid chain with phrases
							m.SongPlaybackRow[track] = searchRow
							foundValidChain = true

							if searchRow < startSearchRow {
								log.Printf("Song track %d looped back to row %02X", track, searchRow)
							} else {
								log.Printf("Song track %d advanced to row %02X", track, searchRow)
							}
							break
						}
					}
				}

				if !foundValidChain {
					// No valid chains found in entire song, deactivate track
					m.SongPlaybackActive[track] = false
					log.Printf("Song track %d deactivated (no valid chains found in song)", track)
					continue
				}

				// Use the chainID and phrase info found in the search above
				chainID = m.SongData[track][m.SongPlaybackRow[track]]
				m.SongPlaybackChain[track] = chainID

				// Find first phrase in this chain (we know it exists from the search)
				firstPhraseID := -1
				firstChainRow := -1
				for chainRow := 0; chainRow < 16; chainRow++ {
					if (*chainsData)[chainID][chainRow] != -1 {
						firstPhraseID = (*chainsData)[chainID][chainRow]
						firstChainRow = chainRow
						break
					}
				}

				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d now at row %02X, chain %02X (chain row %d), phrase %02X", track, m.SongPlaybackRow[track], chainID, firstChainRow, firstPhraseID)
			}

		nextTrack:
			continue
		}
		log.Printf("Song playback: processed %d active tracks", activeTrackCount)
	} else if m.PlaybackMode == types.ChainView {
		// Chain playback mode - advance through phrases in sequence
		// Find next row with playback enabled (P=1)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
		
		// Validate PlaybackPhrase is within bounds before accessing array
		if m.PlaybackPhrase >= 0 && m.PlaybackPhrase < 255 {
			for i := m.PlaybackRow + 1; i < 255; i++ {
				if (*phrasesData)[m.PlaybackPhrase][i][types.ColPlayback] == 1 {
					m.PlaybackRow = i
					DebugLogRowEmission(m)
					log.Printf("Chain playback advanced from row %d to %d", oldRow, m.PlaybackRow)
					return
				}
			}
		}

		// End of phrase reached, move to next phrase slot in the same chain
		for i := m.PlaybackChainRow + 1; i < 16; i++ {
			phraseID := m.ChainsData[m.PlaybackChain][i]
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
			phraseID := m.ChainsData[m.PlaybackChain][i]
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
		// Find next row with playback enabled (P=1)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
		for i := m.PlaybackRow + 1; i < 255; i++ {
			if (*phrasesData)[m.PlaybackPhrase][i][types.ColPlayback] == 1 {
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
		for i := 0; i < 255; i++ {
			if (*phrasesData)[phraseNum][i][types.ColPlayback] == 1 {
				return i
			}
		}
	}
	return 0 // Fallback to row 0 if no playback rows found
}

func FindFirstNonEmptyChain(m *model.Model) int {
	for i := 0; i < 255; i++ {
		// Check if any phrase is assigned in this chain
		for row := 0; row < 16; row++ {
			if m.ChainsData[i][row] != -1 {
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

	// Set playback flag to 1
	(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 1

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
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 0   // Reset playback to 0
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1      // Clear note
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1 // Clear deltatime
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

	rowData := m.PhrasesData[m.CurrentPhrase][m.CurrentRow]

	// A row is considered empty if all data fields are at their default values
	return rowData[types.ColNote] == -1 &&
		rowData[types.ColDeltaTime] == -1 &&
		rowData[types.ColFilename] == -1
}

// GetEffectiveValue searches backwards from the current row to find the first non-null value for a given column
func GetEffectiveValue(m *model.Model, phrase, row, colIndex int) int {
	// Search backwards from the given row to find first non-null value
	for r := row; r >= 0; r-- {
		value := m.PhrasesData[phrase][r][colIndex]
		if value != -1 {
			return value
		}
	}
	return -1 // No non-null value found
}

// GetEffectiveFilename gets the effective filename for a row
func GetEffectiveFilename(m *model.Model, phrase, row int) string {
	effectiveFileIndex := GetEffectiveValue(m, phrase, row, int(types.ColFilename))
	if effectiveFileIndex >= 0 && effectiveFileIndex < len(m.PhrasesFiles) && m.PhrasesFiles[effectiveFileIndex] != "" {
		return m.PhrasesFiles[effectiveFileIndex]
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
	// Use track-aware data access for correct playback
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	rowData := (*phrasesData)[phrase][row]

	// Raw values
	rawPlayback := rowData[types.ColPlayback]
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
	effectivePan := GetEffectiveValue(m, phrase, row, int(types.ColPan))
	effectiveLowPassFilter := GetEffectiveValue(m, phrase, row, int(types.ColLowPassFilter))
	effectiveHighPassFilter := GetEffectiveValue(m, phrase, row, int(types.ColHighPassFilter))
	effectiveComb := GetEffectiveValue(m, phrase, row, int(types.ColEffectComb))
	effectiveReverb := GetEffectiveValue(m, phrase, row, int(types.ColEffectReverb))

	// Effective/inherited values
	effectiveNote := GetEffectiveValue(m, phrase, row, int(types.ColNote))
	effectiveDeltaTime := GetEffectiveValue(m, phrase, row, int(types.ColDeltaTime))
	effectiveFilenameIndex := GetEffectiveValue(m, phrase, row, int(types.ColFilename))
	effectiveFilename := GetEffectiveFilename(m, phrase, row)

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
	log.Printf("Playback: %d", rawPlayback)
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

	// Only emit if we have playback enabled, a concrete note, and a file
	if rawPlayback != 1 || rawNote == -1 || effectiveFilename == "none" {
		log.Printf("ROW_EMIT: skipped (P!=1 or NN==null or FN==none)")
		return
	}

	// shouldEmitRow() hook (kept from playback path; e.g., to skip DT==00, etc.)
	if shouldEmitRow(m) == false {
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
		)
	} else {
		oscParams = model.NewSamplerOSCParams(effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration)
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
	if isInstrumentTrack(trackId) {
		// For instrument tracks, emit simplified instrument message
		velocity := float32(1.0) // Default velocity - could be derived from Gate in future
		instrumentParams := model.NewInstrumentOSCParams(trackId, effectiveNote, velocity)
		m.SendOSCInstrumentMessage(instrumentParams)
	} else {
		// For sampler tracks, emit full sampler message
		m.SendOSCSamplerMessage(oscParams)
	}
}

// isInstrumentTrack determines if the given track should use instrument OSC messages
// Tracks 0-3 (displayed as 1-4) are Instrument tracks, tracks 4-7 (displayed as 5-8) are Sampler tracks
func isInstrumentTrack(trackId int) bool {
	return trackId >= 0 && trackId <= 3
}

// rowDurationMS returns the per-row duration in milliseconds.
// DT is read *row-locally* and never inherited.
// - DT == -1 (null) -> base duration
// - DT == 0         -> base duration (no emission rule handled elsewhere)
// - DT > 0          -> base * (1 + DT/16)
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

	dtRaw := m.PhrasesData[p][r][types.ColDeltaTime] // row-local DT
	if dtRaw > 0 {
		return baseMs * (1.0 + float64(dtRaw)/16.0)
	}
	return baseMs
}

// shouldEmitRow enforces:
// - NN must be present (not -1)
// - P (playback flag) must be 1
// - DT == 00 => do NOT emit
// DT is read row-locally and never inherited.
func shouldEmitRow(m *model.Model) bool {
	p := m.PlaybackPhrase
	r := m.PlaybackRow
	if p < 0 || p >= 255 || r < 0 || r >= 255 {
		return false
	}

	nn := m.PhrasesData[p][r][types.ColNote]         // may still be inherited elsewhere in your code
	dtRaw := m.PhrasesData[p][r][types.ColDeltaTime] // strictly local for this rule
	pFlag := m.PhrasesData[p][r][types.ColPlayback]

	if pFlag != 1 {
		return false
	}
	if nn == -1 {
		return false
	}
	if dtRaw == 0 {
		// "DT == 00 => no emission" (time still passes)
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
		// Song playback mode - initialize all tracks with data
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

				// Emit initial row for this track
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d started at row %02X, chain %02X (chain row %d), phrase %02X", track, startRow, chainID, firstChainRow, firstPhraseID)
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
			m.SongPlaybackPhrase[0] = 0  // Use phrase 0 as fallback
			m.SongPlaybackRowInPhrase[0] = FindFirstNonEmptyRowInPhraseForTrack(m, 0, 0)
			EmitRowDataFor(m, 0, m.SongPlaybackRowInPhrase[0], 0)
			log.Printf("Song track 0 fallback started at phrase 0, row %d", m.SongPlaybackRowInPhrase[0])
		}
	} else if config.Mode == types.ChainView {
		// Chain playback mode - find appropriate starting phrase
		m.PlaybackChain = config.Chain
		m.PlaybackPhrase = -1

		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			// Start from specified chain row
			if m.ChainsData[config.Chain][config.Row] != -1 {
				m.PlaybackChainRow = config.Row
				m.PlaybackPhrase = m.ChainsData[config.Chain][config.Row]
			}
		}

		// If no phrase found yet, find first non-empty phrase slot in this chain
		if m.PlaybackPhrase == -1 {
			for row := 0; row < 16; row++ {
				if m.ChainsData[config.Chain][row] != -1 {
					m.PlaybackPhrase = m.ChainsData[config.Chain][row]
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
				if m.ChainsData[m.PlaybackChain][row] != -1 {
					m.PlaybackPhrase = m.ChainsData[m.PlaybackChain][row]
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
			log.Printf("Chain %d contents: %v", m.PlaybackChain, m.ChainsData[m.PlaybackChain])
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

		if config.UseCurrentRow && config.Row >= 0 {
			m.PlaybackRow = config.Row
		} else {
			m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
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
		if (*phrasesData)[phraseID][row][types.ColPlayback] == 1 {
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

// ToggleMixerTrackType toggles the track type between Instrument (IN) and Sampler (SA)
func ToggleMixerTrackType(m *model.Model) {
	// Bounds check
	if m.CurrentMixerTrack < 0 || m.CurrentMixerTrack >= 8 {
		return
	}

	// Toggle the track type
	oldType := m.TrackTypes[m.CurrentMixerTrack]
	m.TrackTypes[m.CurrentMixerTrack] = !oldType
	
	var oldTypeStr, newTypeStr string
	if oldType {
		oldTypeStr = "SA"
	} else {
		oldTypeStr = "IN"
	}
	if m.TrackTypes[m.CurrentMixerTrack] {
		newTypeStr = "SA"
	} else {
		newTypeStr = "IN"
	}

	log.Printf("Toggled mixer track %d type: %s -> %s", m.CurrentMixerTrack+1, oldTypeStr, newTypeStr)
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
	if colIndex == int(types.ColPlayback) {
		// Special bit-flipping logic for P column
		currentValue := (*phrasesData)[m.CurrentPhrase][currentRow][colIndex]
		
		if currentValue == 0 {
			// Current cell is 0: flip all 0's above (and current) until most recent "1" to 1
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue == 1 {
					break // Stop at the most recent "1"
				}
				if cellValue == 0 {
					(*phrasesData)[m.CurrentPhrase][row][colIndex] = 1
				}
			}
		} else {
			// Current cell is 1: flip all 1's above (and current) until most recent "0" to 0
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue == 0 {
					break // Stop at the most recent "0"
				}
				if cellValue == 1 {
					(*phrasesData)[m.CurrentPhrase][row][colIndex] = 0
				}
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
			if (*phrasesData)[m.CurrentPhrase][row][types.ColPlayback] == 0 {
				(*phrasesData)[m.CurrentPhrase][row][types.ColPlayback] = 1
			}
		}
	}
	
	// Update last edit row
	m.LastEditRow = currentRow
	
	log.Printf("Filled phrase %02X column %d from row %d to %d, starting with %d", m.CurrentPhrase, colIndex, startRow, currentRow, startValue)
}

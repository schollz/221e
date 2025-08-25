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
		currentValue := m.ChainsData[m.CurrentChain][m.CurrentRow]

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
		m.ChainsData[m.CurrentChain][m.CurrentRow] = newValue

		log.Printf("Modified chain %02X row %02X phrase: %d -> %d (delta: %d)", m.CurrentChain, m.CurrentRow, currentValue, newValue, delta)
		storage.AutoSave(m)
		return
	}

	// Phrase view: modify cell under cursor
	colIndex := m.CurrentCol - 1 // data index
	currentValue := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex]

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
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

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
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

	} else {
		// All other hex-ish columns (NN, DT, GT, RT, TS, CO, VE, FI index)
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
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

		// Auto-enable playback on first note entry
		if colIndex == int(types.ColNote) && m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColPlayback] == 0 {
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColPlayback] = 1
			log.Printf("Auto-enabled playback for phrase %d row %d due to note change", m.CurrentPhrase, m.CurrentRow)
		}
	}

	m.LastEditRow = m.CurrentRow
	log.Printf("Modified phrase %d row %d, col %d: %d -> %d (delta: %d)",
		m.CurrentPhrase, m.CurrentRow, colIndex, currentValue,
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex], delta)
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
		for track := 0; track < 8; track++ {
			if !m.SongPlaybackActive[track] {
				continue
			}

			oldPhraseRow := m.SongPlaybackRowInPhrase[track]

			// Declare variables early to avoid goto issues
			var chainID int
			var nextPhraseFound bool
			var currentChain int

			// Advance within current phrase
			phraseNum := m.SongPlaybackPhrase[track]
			if phraseNum >= 0 && phraseNum < 255 {
				// Find next row with playback enabled in current phrase
				for i := m.SongPlaybackRowInPhrase[track] + 1; i < 255; i++ {
					if m.PhrasesData[phraseNum][i][types.ColPlayback] == 1 {
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
			for chainRow := m.SongPlaybackChainRow[track] + 1; chainRow < 16; chainRow++ {
				if m.ChainsData[currentChain][chainRow] != -1 {
					// Found next phrase in chain
					m.SongPlaybackChainRow[track] = chainRow
					m.SongPlaybackPhrase[track] = m.ChainsData[currentChain][chainRow]
					m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhrase(m, m.SongPlaybackPhrase[track])
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
							if m.ChainsData[chainID][chainRow] != -1 {
								firstPhraseID = m.ChainsData[chainID][chainRow]
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
					if m.ChainsData[chainID][chainRow] != -1 {
						firstPhraseID = m.ChainsData[chainID][chainRow]
						firstChainRow = chainRow
						break
					}
				}

				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhrase(m, firstPhraseID)
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d now at row %02X, chain %02X (chain row %d), phrase %02X", track, m.SongPlaybackRow[track], chainID, firstChainRow, firstPhraseID)
			}

		nextTrack:
			continue
		}
	} else if m.PlaybackMode == types.ChainView {
		// Chain playback mode - advance through phrases in sequence
		// Find next row with playback enabled (P=1)
		for i := m.PlaybackRow + 1; i < 255; i++ {
			if m.PhrasesData[m.PlaybackPhrase][i][types.ColPlayback] == 1 {
				m.PlaybackRow = i
				DebugLogRowEmission(m)
				log.Printf("Chain playback advanced from row %d to %d", oldRow, m.PlaybackRow)
				return
			}
		}

		// End of phrase reached, move to next phrase slot in the same chain
		for i := m.PlaybackChainRow + 1; i < 16; i++ {
			if m.ChainsData[m.PlaybackChain][i] != -1 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = m.ChainsData[m.PlaybackChain][i]
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback moved to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}

		// End of chain reached, loop back to first phrase slot in the same chain
		for i := 0; i < 16; i++ {
			if m.ChainsData[m.PlaybackChain][i] != -1 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = m.ChainsData[m.PlaybackChain][i]
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback looped back to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}
	} else {
		// Phrase-only playback mode
		// Find next row with playback enabled (P=1)
		for i := m.PlaybackRow + 1; i < 255; i++ {
			if m.PhrasesData[m.PlaybackPhrase][i][types.ColPlayback] == 1 {
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
	if phraseNum >= 0 && phraseNum < 255 {
		for i := 0; i < 255; i++ {
			if m.PhrasesData[phraseNum][i][types.ColPlayback] == 1 {
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

	// Check if current row is empty (note, deltatime, filename are -1, playback is 0)
	if m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColNote] != -1 ||
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] != -1 ||
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColFilename] != -1 {
		log.Printf("Current row %d is not empty, skipping copy", m.CurrentRow)
		return
	}

	// Find the first non-null note above the current row
	var sourceNote int = -1
	for r := m.CurrentRow - 1; r >= 0; r-- {
		if m.PhrasesData[m.CurrentPhrase][r][types.ColNote] != -1 {
			sourceNote = m.PhrasesData[m.CurrentPhrase][r][types.ColNote]
			log.Printf("Found non-null note %d at row %d", sourceNote, r)
			break
		}
	}

	// If no non-null note found above, do nothing
	if sourceNote == -1 {
		log.Printf("No non-null note found above current row, skipping")
		return
	}

	// Set playback flag to 1
	m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 1

	// Increment the note and set it
	newNote := sourceNote + 1
	if newNote > 254 { // Wrap around if needed
		newNote = 0
	}
	m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = newNote

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
		value := m.ChainsData[m.CurrentChain][m.CurrentRow]
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
		colIndex := m.CurrentCol - 1
		if colIndex >= 0 && colIndex < int(types.ColCount) {
			value := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex]
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
		rowData := make([]int, int(types.ColCount))
		copy(rowData, m.PhrasesData[m.CurrentPhrase][m.CurrentRow])

		// Get filename if exists
		var filename string
		fileIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		if fileIndex >= 0 && fileIndex < len(m.PhrasesFiles) {
			filename = m.PhrasesFiles[fileIndex]
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
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 0   // Reset playback to 0
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1      // Clear note
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1 // Clear deltatime
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = -1  // Clear filename
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
			m.ChainsData[m.CurrentChain][m.CurrentRow] = m.Clipboard.Value
			log.Printf("Pasted to chain %02X row %02X phrase: %d", m.CurrentChain, m.CurrentRow, m.Clipboard.Value)
		} else {
			log.Printf("Cannot paste: wrong cell type or position")
		}
	} else if m.ViewMode == types.PhraseView {
		// Paste to phrase view
		colIndex := m.CurrentCol - 1
		if colIndex >= 0 && colIndex < int(types.ColCount) {
			var canPaste bool
			if colIndex == int(types.ColFilename) { // Filename column
				canPaste = m.Clipboard.CellType == types.FilenameCell
			} else { // Playback, note, or deltatime column
				canPaste = m.Clipboard.CellType == types.HexCell
			}

			if canPaste {
				m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
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
		for i, value := range m.Clipboard.RowData {
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][i] = value
		}

		// Handle filename if it exists
		if m.Clipboard.RowFilename != "" {
			// Add filename to files array and update index
			fileIndex := m.AppendPhrasesFile(m.Clipboard.RowFilename)
			m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = fileIndex
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
	rowData := m.PhrasesData[phrase][row]

	// Raw values
	rawPlayback := rowData[types.ColPlayback]
	rawNote := rowData[types.ColNote]
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

	// Emit
	m.SendOSCSamplerMessage(oscParams)
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

		for track := 0; track < 8; track++ {
			chainID := m.SongData[track][startRow]
			if chainID == -1 {
				// No chain at this position
				m.SongPlaybackActive[track] = false
				continue
			}

			// Check if chain has valid phrase data (find first phrase in chain)
			firstPhraseID := -1
			firstChainRow := -1
			for chainRow := 0; chainRow < 16; chainRow++ {
				if m.ChainsData[chainID][chainRow] != -1 {
					firstPhraseID = m.ChainsData[chainID][chainRow]
					firstChainRow = chainRow
					break
				}
			}

			if firstPhraseID != -1 {
				m.SongPlaybackActive[track] = true
				m.SongPlaybackRow[track] = startRow
				m.SongPlaybackChain[track] = chainID
				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhrase(m, firstPhraseID)

				// Emit initial row for this track
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d started at row %02X, chain %02X (chain row %d), phrase %02X", track, startRow, chainID, firstChainRow, firstPhraseID)
			} else {
				// Chain exists but has no phrases
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d skipped at row %02X (chain %02X has no phrases)", track, startRow, chainID)
			}
		}

		log.Printf("Song playback started from row %02X", startRow)
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
			m.PlaybackChainRow = 0
			for row := 0; row < 16; row++ {
				if m.ChainsData[m.PlaybackChain][row] != -1 {
					m.PlaybackPhrase = m.ChainsData[m.PlaybackChain][row]
					m.PlaybackChainRow = row
					break
				}
			}
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
	for row := 0; row < 255; row++ {
		if m.PhrasesData[phraseID][row][types.ColPlayback] == 1 {
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

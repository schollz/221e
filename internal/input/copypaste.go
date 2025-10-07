package input

import (
	"log"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

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

		// Check if we're on a header row (invalid row index)
		if m.CurrentRow < 0 || m.CurrentRow >= 256 {
			log.Printf("Cannot copy from header row (row %d)", m.CurrentRow)
			return
		}

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
	} else if m.ViewMode == types.ArpeggioView {
		// Copy from arpeggio view
		settings := m.ArpeggioSettings[m.ArpeggioEditingIndex]
		currentRow := &settings.Rows[m.CurrentRow]

		var value int
		switch m.CurrentCol {
		case int(types.ArpeggioColDI): // Direction column
			value = currentRow.Direction
		case int(types.ArpeggioColCO): // Count column
			value = currentRow.Count
		case int(types.ArpeggioColDIV): // Divisor column
			value = currentRow.Divisor
		default:
			return // Invalid column
		}

		clipboard := types.ClipboardData{
			Value:           value,
			CellType:        types.HexCell,
			Mode:            types.CellMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    m.CurrentCol,
			HighlightPhrase: -1, // Not applicable for arpeggio view
			HighlightView:   types.ArpeggioView,
		}
		m.Clipboard = clipboard
		log.Printf("Copied arpeggio cell value: %d from row %02X col %d", value, m.CurrentRow, m.CurrentCol)
	} else if m.ViewMode == types.RetriggerView {
		// Copy retrigger index from retrigger view
		value := m.RetriggerEditingIndex
		clipboard := types.ClipboardData{
			Value:           value,
			CellType:        types.HexCell,
			Mode:            types.CellMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    m.CurrentCol,
			HighlightPhrase: -1, // Not applicable for retrigger view
			HighlightView:   types.RetriggerView,
		}
		m.Clipboard = clipboard
		log.Printf("Copied retrigger index: %02X", value)
	} else if m.ViewMode == types.TimestrechView {
		// Copy timestrech index from timestrech view
		value := m.TimestrechEditingIndex
		clipboard := types.ClipboardData{
			Value:           value,
			CellType:        types.HexCell,
			Mode:            types.CellMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    m.CurrentCol,
			HighlightPhrase: -1, // Not applicable for timestrech view
			HighlightView:   types.TimestrechView,
		}
		m.Clipboard = clipboard
		log.Printf("Copied timestrech index: %02X", value)
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

		// Check if we're on a header row (invalid row index)
		if m.CurrentRow < 0 || m.CurrentRow >= 256 {
			log.Printf("Cannot cut from header row (row %d)", m.CurrentRow)
			return
		}

		phrasesData := m.GetCurrentPhrasesData()
		rowData := make([]int, int(types.ColCount))
		copy(rowData, (*phrasesData)[m.CurrentPhrase][m.CurrentRow])

		// Get filename if exists
		var filename string
		fileIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if phrasesFiles != nil && fileIndex >= 0 && fileIndex < len(*phrasesFiles) {
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
		// Clear the row - reset all columns to their default values
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1                               // Clear note
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPitch)] = -1                              // Clear pitch (displays "--", behaves as 80)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1                          // Clear deltatime (clears playback for both views)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColVelocity)] = -1                           // Clear velocity
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColGate)] = -1                               // Clear gate (displays "--", behaves as 80)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColRetrigger)] = -1                          // Clear retrigger
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColTimestretch)] = -1                        // Clear timestretch
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColEffectDucking)] = -1                      // Clear ducking
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColModulate)] = -1                           // Clear modulation
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColEffectReverse)] = -1                      // Clear effect reverse
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPan)] = -1                                // Clear pan
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = -1                           // Clear filename
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColChord)] = int(types.ChordNone)            // Clear chord type
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColChordAddition)] = int(types.ChordAddNone) // Clear chord addition
		// Clear Instrument-specific columns (A, D, S, R, RE, CO, LP, HP, AR, MI, SO)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColAttack)] = -1                                    // Clear attack
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDecay)] = -1                                     // Clear decay
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColSustain)] = -1                                   // Clear sustain
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColRelease)] = -1                                   // Clear release
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColEffectReverb)] = -1                              // Clear reverb
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColEffectComb)] = -1                                // Clear comb
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColLowPassFilter)] = -1                             // Clear low pass
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColHighPassFilter)] = -1                            // Clear high pass
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColArpeggio)] = -1                                  // Clear arpeggio
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColMidi)] = -1                                      // Clear MIDI
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColSoundMaker)] = -1                                // Clear SoundMaker
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColChordTransposition)] = int(types.ChordTransNone) // Clear chord transposition
		log.Printf("Cut phrase row %d", m.CurrentRow)
	} else if m.ViewMode == types.ArpeggioView {
		// Cut row from arpeggio view
		if m.ArpeggioEditingIndex < 0 || m.ArpeggioEditingIndex >= 255 {
			return
		}
		if m.CurrentRow < 0 || m.CurrentRow >= 16 {
			return
		}

		// Copy current arpeggio row data (3 columns: Direction, Count, Divisor)
		currentRow := m.ArpeggioSettings[m.ArpeggioEditingIndex].Rows[m.CurrentRow]
		arpeggioRowData := make([]int, 3)
		arpeggioRowData[0] = currentRow.Direction
		arpeggioRowData[1] = currentRow.Count
		arpeggioRowData[2] = currentRow.Divisor

		clipboard := types.ClipboardData{
			RowData:         arpeggioRowData,
			SourceView:      types.ArpeggioView,
			Mode:            types.RowMode,
			HasData:         true,
			HighlightRow:    m.CurrentRow,
			HighlightCol:    -1, // Highlight entire row
			HighlightPhrase: -1, // Not applicable for arpeggio view
			HighlightView:   types.ArpeggioView,
		}
		m.Clipboard = clipboard

		// Clear the row - reset to defaults
		currentRowRef := &m.ArpeggioSettings[m.ArpeggioEditingIndex].Rows[m.CurrentRow]
		currentRowRef.Direction = 0 // Clear to "--"
		currentRowRef.Count = -1    // Clear to "--"
		currentRowRef.Divisor = -1  // Clear to "--"

		log.Printf("Cut arpeggio %02X row %02X", m.ArpeggioEditingIndex, m.CurrentRow)
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

		// Check if we're on a header row (invalid row index)
		if m.CurrentRow < 0 || m.CurrentRow >= 256 {
			log.Printf("Cannot paste to header row (row %d)", m.CurrentRow)
			return
		}

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
				// Special handling for retrigger column - implement deep copying
				if colIndex == int(types.ColRetrigger) && m.Clipboard.Value >= 0 && m.Clipboard.Value < 255 {
					// Check if this is marked for deep copy on paste (Ctrl+D)
					if m.Clipboard.IsFreshDeepCopy {
						// Create the deep copy now (on paste)
						newRetriggerIndex := FindNextUnusedRetrigger(m, m.Clipboard.Value)
						if newRetriggerIndex != -1 {
							// Deep copy the retrigger settings
							m.RetriggerSettings[newRetriggerIndex] = m.RetriggerSettings[m.Clipboard.Value]
							// Update the phrase data with the new retrigger index
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newRetriggerIndex
							log.Printf("Deep copied retrigger settings %02X to %02X and pasted to phrase cell", m.Clipboard.Value, newRetriggerIndex)
						} else {
							// No unused retrigger slots available, just copy the reference
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
							log.Printf("Warning: No unused retrigger slots available, pasted reference to retrigger %02X", m.Clipboard.Value)
						}
					} else {
						// Regular copy (Ctrl+C) - just paste the same value (reference)
						(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
						log.Printf("Pasted retrigger reference %02X to phrase cell", m.Clipboard.Value)
					}
				} else if colIndex == int(types.ColTimestretch) && m.Clipboard.Value >= 0 && m.Clipboard.Value < 255 {
					// Special handling for timestretch column - implement deep copying
					// Check if this is marked for deep copy on paste (Ctrl+D)
					if m.Clipboard.IsFreshDeepCopy {
						// Create the deep copy now (on paste)
						newTimestrechIndex := FindNextUnusedTimestrech(m, m.Clipboard.Value)
						if newTimestrechIndex != -1 {
							// Deep copy the timestrech settings
							m.TimestrechSettings[newTimestrechIndex] = m.TimestrechSettings[m.Clipboard.Value]
							// Update the phrase data with the new timestrech index
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newTimestrechIndex
							log.Printf("Deep copied timestrech settings %02X to %02X and pasted to phrase cell", m.Clipboard.Value, newTimestrechIndex)
						} else {
							// No unused timestrech slots available, just copy the reference
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
							log.Printf("Warning: No unused timestrech slots available, pasted reference to timestrech %02X", m.Clipboard.Value)
						}
					} else {
						// Regular copy (Ctrl+C) - just paste the reference
						(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
						log.Printf("Pasted timestrech reference %02X to phrase cell", m.Clipboard.Value)
					}
				} else if colIndex == int(types.ColArpeggio) && m.Clipboard.Value >= 0 && m.Clipboard.Value < 255 {
					// Special handling for arpeggio column - implement deep copying
					// Check if this is marked for deep copy on paste (Ctrl+D)
					if m.Clipboard.IsFreshDeepCopy {
						// Create the deep copy now (on paste)
						newArpeggioIndex := FindNextUnusedArpeggio(m, m.Clipboard.Value)
						if newArpeggioIndex != -1 {
							// Deep copy the arpeggio settings
							m.ArpeggioSettings[newArpeggioIndex] = m.ArpeggioSettings[m.Clipboard.Value]
							// Update the phrase data with the new arpeggio index
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newArpeggioIndex
							log.Printf("Deep copied arpeggio settings %02X to %02X and pasted to phrase cell", m.Clipboard.Value, newArpeggioIndex)
						} else {
							// No unused arpeggio slots available, just copy the reference
							(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
							log.Printf("Warning: No unused arpeggio slots available, pasted reference to arpeggio %02X", m.Clipboard.Value)
						}
					} else {
						// Regular copy (Ctrl+C) - just paste the same value (reference)
						(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
						log.Printf("Pasted arpeggio reference %02X to phrase cell", m.Clipboard.Value)
					}
				} else {
					// Normal paste for all other columns
					(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = m.Clipboard.Value
					log.Printf("Pasted to phrase cell: %d", m.Clipboard.Value)
				}
				// Track this row as the last edited row
				m.LastEditRow = m.CurrentRow
			} else {
				log.Printf("Cannot paste: incompatible cell type")
			}
		}
	} else if m.ViewMode == types.ArpeggioView {
		// Paste to arpeggio view - only paste within same column
		if m.Clipboard.CellType == types.HexCell && m.Clipboard.HighlightView == types.ArpeggioView && m.Clipboard.HighlightCol == m.CurrentCol {
			settings := &m.ArpeggioSettings[m.ArpeggioEditingIndex]
			currentRow := &settings.Rows[m.CurrentRow]

			switch m.CurrentCol {
			case int(types.ArpeggioColDI): // Direction column
				currentRow.Direction = m.Clipboard.Value
				log.Printf("Pasted to arpeggio %02X row %02X Direction: %d", m.ArpeggioEditingIndex, m.CurrentRow, m.Clipboard.Value)
			case int(types.ArpeggioColCO): // Count column
				currentRow.Count = m.Clipboard.Value
				log.Printf("Pasted to arpeggio %02X row %02X Count: %d", m.ArpeggioEditingIndex, m.CurrentRow, m.Clipboard.Value)
			case int(types.ArpeggioColDIV): // Divisor column
				currentRow.Divisor = m.Clipboard.Value
				log.Printf("Pasted to arpeggio %02X row %02X Divisor: %d", m.ArpeggioEditingIndex, m.CurrentRow, m.Clipboard.Value)
			default:
				log.Printf("Cannot paste: invalid arpeggio column %d", m.CurrentCol)
			}
		} else {
			log.Printf("Cannot paste: incompatible cell type or different column (source col: %d, target col: %d)", m.Clipboard.HighlightCol, m.CurrentCol)
		}
	} else if m.ViewMode == types.RetriggerView {
		// Paste to retrigger view - find next empty slot in retrigger pool
		if m.Clipboard.CellType == types.HexCell && m.Clipboard.Value >= 0 && m.Clipboard.Value < 255 {
			// Find next unused retrigger slot
			nextSlot := FindNextUnusedRetrigger(m, m.Clipboard.Value)
			if nextSlot != -1 {
				// Deep copy the retrigger settings to the next empty slot
				sourceSettings := m.RetriggerSettings[m.Clipboard.Value]
				m.RetriggerSettings[nextSlot] = sourceSettings
				log.Printf("Deep copied retrigger settings %02X to next empty slot %02X (Times: %d, Start: %.2f, End: %.2f, Beats: %d, Every: %d, Probability: %d)",
					m.Clipboard.Value, nextSlot, sourceSettings.Times, sourceSettings.Start, sourceSettings.End, sourceSettings.Beats, sourceSettings.Every, sourceSettings.Probability)
			} else {
				log.Printf("Cannot paste: no empty retrigger slots available")
			}
		} else {
			log.Printf("Cannot paste: incompatible cell type for retrigger view")
		}
	} else if m.ViewMode == types.TimestrechView {
		// Paste to timestrech view - find next empty slot in timestrech pool
		if m.Clipboard.CellType == types.HexCell && m.Clipboard.Value >= 0 && m.Clipboard.Value < 255 {
			// Find next unused timestrech slot
			nextSlot := FindNextUnusedTimestrech(m, m.Clipboard.Value)
			if nextSlot != -1 {
				// Deep copy the timestrech settings to the next empty slot
				sourceSettings := m.TimestrechSettings[m.Clipboard.Value]
				m.TimestrechSettings[nextSlot] = sourceSettings
				log.Printf("Deep copied timestrech settings %02X to next empty slot %02X (Start: %.2f, End: %.2f, Beats: %d, Every: %d, Probability: %d)",
					m.Clipboard.Value, nextSlot, sourceSettings.Start, sourceSettings.End, sourceSettings.Beats, sourceSettings.Every, sourceSettings.Probability)
			} else {
				log.Printf("Cannot paste: no empty timestrech slots available")
			}
		} else {
			log.Printf("Cannot paste: incompatible cell type for timestrech view")
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

		// Check if we're on a header row (invalid row index)
		if m.CurrentRow < 0 || m.CurrentRow >= 256 {
			log.Printf("Cannot paste to header row (row %d)", m.CurrentRow)
			return
		}

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
	} else if m.ViewMode == types.ArpeggioView && m.Clipboard.SourceView == types.ArpeggioView {
		// Paste arpeggio row to arpeggio row
		settings := &m.ArpeggioSettings[m.ArpeggioEditingIndex]
		if len(m.Clipboard.RowData) == 3 {
			settings.Rows[m.CurrentRow].Direction = m.Clipboard.RowData[0]
			settings.Rows[m.CurrentRow].Count = m.Clipboard.RowData[1]
			settings.Rows[m.CurrentRow].Divisor = m.Clipboard.RowData[2]
			log.Printf("Pasted arpeggio row to row %d", m.CurrentRow)
		} else {
			log.Printf("Cannot paste: invalid arpeggio clipboard data")
		}
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

package input

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/schollz/2n/internal/audio"
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/types"
)

// ViewSwitchConfig represents the configuration for switching to a new view
type ViewSwitchConfig struct {
	ViewMode     types.ViewMode
	Row          int
	Col          int
	ScrollOffset int
}

// switchToView provides common logic for view transitions
func switchToView(m *model.Model, config ViewSwitchConfig) {
	m.ViewMode = config.ViewMode
	m.CurrentRow = config.Row
	m.CurrentCol = config.Col
	m.ScrollOffset = config.ScrollOffset
	storage.AutoSave(m)
}

// switchToViewWithVisibilityCheck ensures the cursor row is visible after switching
func switchToViewWithVisibilityCheck(m *model.Model, config ViewSwitchConfig) {
	switchToView(m, config)

	// Ensure the cursor row is visible
	visibleRows := m.GetVisibleRows()

	// Handle negative rows (like TYPE row in Song view)
	if m.CurrentRow < 0 {
		m.ScrollOffset = 0
	} else if m.CurrentRow >= visibleRows {
		m.ScrollOffset = m.CurrentRow - visibleRows + 1
	} else if m.CurrentRow < m.ScrollOffset {
		m.ScrollOffset = m.CurrentRow
	}
}

// Common view switch configurations
func songViewConfig(row, col int) ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.SongView,
		Row:          row,
		Col:          col,
		ScrollOffset: 0,
	}
}

func settingsViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.SettingsView,
		Row:          0,
		Col:          0, // Start in Global column
		ScrollOffset: 0,
	}
}

func chainViewConfig(row int) ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.ChainView,
		Row:          row,
		Col:          1,
		ScrollOffset: 0,
	}
}

func phraseViewConfig(row, col int) ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.PhraseView,
		Row:          row,
		Col:          col,
		ScrollOffset: 0,
	}
}

func fileViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.FileView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func fileMetadataViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.FileMetadataView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func retriggerViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.RetriggerView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func timestrechViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.TimestrechView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func arpeggioViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.ArpeggioView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func mixerViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.MixerView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func midiViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.MidiView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

func soundMakerViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.SoundMakerView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	}
}

type TickMsg time.Time

func HandleKeyInput(m *model.Model, msg tea.KeyMsg) tea.Cmd {
	log.Printf("key: %s", msg.String())
	switch msg.String() {
	case "ctrl+q":
		return tea.Quit

	case "esc":
		ClearClipboardHighlight(m)
		// Clear current cell in Arpeggio Settings
		if m.ViewMode == types.ArpeggioView {
			ClearArpeggioCell(m)
		}

	case "shift+right":
		return handleShiftRight(m)

	case "shift+up":
		return handleShiftUp(m)

	case "shift+down":
		return handleShiftDown(m)

	case "shift+left":
		return handleShiftLeft(m)

	case "up":
		return handleUp(m)

	case "down":
		return handleDown(m)

	case "left":
		return handleLeft(m)

	case "right":
		return handleRight(m)

	case "ctrl+up":
		return handleCtrlUp(m)

	case "ctrl+s":
		return handleCtrlS(m)

	case "ctrl+down":
		return handleCtrlDown(m)

	case "ctrl+left":
		return handleCtrlLeft(m)

	case "ctrl+right":
		return handleCtrlRight(m)

	case "s":
		return handleS(m)

	case "c":
		return handleC(m)

	case "ctrl+c":
		return handleCtrlC(m)

	case "ctrl+x":
		return handleCtrlX(m)

	case "ctrl+v":
		return handleCtrlV(m)

	case "ctrl+d":
		return handleCtrlD(m)

	case " ":
		return handleSpace(m)

	case "ctrl+@":
		return handleCtrlSpace(m)

	case "backspace":
		return handleBackspace(m)

	case "x":
		return handleBackspace(m)

	case "ctrl+h":
		return handleCtrlH(m)

	case "ctrl+r":
		return handleCtrlR(m)

	case "ctrl+f":
		return handleCtrlF(m)

	case "p":
		return handleP(m)

	case "m":
		return handleM(m)
	}

	return nil
}

func handleShiftRight(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		// Don't navigate when on track type row (row -1)
		if m.CurrentRow == -1 {
			log.Printf("Cannot navigate from track type row (Sampler/Instrument toggle)")
			return nil
		}

		// Navigate to chain view for the selected song cell's chain
		track := m.CurrentCol
		row := m.CurrentRow
		chainID := m.SongData[track][row]

		if chainID != -1 {
			// Remember current song position
			m.LastSongRow = m.CurrentRow
			m.LastSongTrack = m.CurrentCol

			// Switch to chain view for the selected chain
			m.ViewMode = types.ChainView
			m.CurrentChain = chainID // Set which chain we're viewing
			m.CurrentTrack = track   // Set track context for playback markers
			m.CurrentRow = 0         // Start at first row of the chain
			m.CurrentCol = 0         // Only one column in chain view (phrase column)
			m.ScrollOffset = 0

			log.Printf("Navigated from Song (T%d R%02X) to Chain %02X (Track context: %d)", track, row, chainID, track)
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ChainView {
		// Navigate to phrase view for the selected chain row's phrase
		chainsData := m.GetCurrentChainsData()
		phraseNum := (*chainsData)[m.CurrentChain][m.CurrentRow]
		if phraseNum != -1 {
			// Remember current chain and row within chain
			m.LastChainRow = m.CurrentRow
			log.Printf("Saved LastChainRow = %d (Chain %02X Row %02X)", m.LastChainRow, m.CurrentChain, m.CurrentRow)
			prevPhrase := m.CurrentPhrase

			// Switch to phrase view for the selected phrase
			m.ViewMode = types.PhraseView
			goToPhrase(m, phraseNum)

			// If we came from a *different* phrase, always start at the top (row 0).
			// If it's the same phrase, restore the last row.
			if prevPhrase != phraseNum {
				m.CurrentRow = 0
			} else {
				m.CurrentRow = m.LastPhraseRow
			}

			// Set appropriate starting column based on view type
			phraseViewType := m.GetPhraseViewType()
			if phraseViewType == types.InstrumentPhraseView {
				m.CurrentCol = int(types.InstrumentColDT) // Instrument: Start on DT column
			} else {
				m.CurrentCol = int(types.SamplerColNN) // Sampler: Start on Note column
			}
			m.ScrollOffset = 0

			// Ensure the cursor row is visible
			visibleRows := m.GetVisibleRows()
			if m.CurrentRow >= visibleRows {
				m.ScrollOffset = m.CurrentRow - visibleRows + 1
			}

			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.PhraseView {
		// Use centralized column mapping to check if we're on RT or TS columns
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColRetrigger) {
			// Navigate to retrigger view only if a retrigger is selected (not -1)
			phrasesData := m.GetCurrentPhrasesData()
			retriggerIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColRetrigger]
			if retriggerIndex == -1 {
				return nil // Don't navigate if no retrigger is selected
			}
			// Save current phrase view position
			m.LastPhraseRow = m.CurrentRow
			m.RetriggerEditingIndex = retriggerIndex
			m.ViewMode = types.RetriggerView
			m.CurrentRow = 0 // Start at first setting
			m.CurrentCol = 0
			m.ScrollOffset = 0
			storage.AutoSave(m)
			return nil
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColTimestretch) {
			// Check if we're on the TS column
			// Navigate to timestretch view only if a timestretch is selected (not -1)
			phrasesData := m.GetCurrentPhrasesData()
			timestrechIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColTimestretch]
			if timestrechIndex == -1 {
				return nil // Don't navigate if no timestretch is selected
			}
			// Save current phrase view position
			m.LastPhraseRow = m.CurrentRow
			m.TimestrechEditingIndex = timestrechIndex
			m.ViewMode = types.TimestrechView
			m.CurrentRow = 0 // Start at first setting
			m.CurrentCol = 0
			m.ScrollOffset = 0
			storage.AutoSave(m)
			return nil
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColArpeggio) {
			// Navigate to arpeggio view only if an arpeggio is selected (not -1)
			phrasesData := m.GetCurrentPhrasesData()
			arpeggioIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColArpeggio]
			if arpeggioIndex == -1 {
				return nil // Don't navigate if no arpeggio is selected
			}
			// Save current phrase view position
			m.LastPhraseRow = m.CurrentRow
			m.ArpeggioEditingIndex = arpeggioIndex
			m.ViewMode = types.ArpeggioView
			m.CurrentRow = 0 // Start at first row
			m.CurrentCol = 0 // Start at DI column
			m.ScrollOffset = 0
			storage.AutoSave(m)
			return nil
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColMidi) {
			// Navigate to MIDI view - if no MIDI is selected, use 00
			phrasesData := m.GetCurrentPhrasesData()
			midiIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColMidi]
			if midiIndex == -1 {
				// If no MIDI is selected, default to index 00 for settings
				midiIndex = 0
				// Also set the value in the cell to 00 so it shows when we return
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColMidi] = 0
			}
			// Save current phrase view position
			m.LastPhraseRow = m.CurrentRow
			m.MidiEditingIndex = midiIndex
			m.ViewMode = types.MidiView
			m.CurrentRow = 0 // Start at first setting
			m.CurrentCol = 0 // Start at Device column
			m.ScrollOffset = 0
			storage.AutoSave(m)
			return nil
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColSoundMaker) {
			// Navigate to SoundMaker view - if no SoundMaker is selected, use 00
			phrasesData := m.GetCurrentPhrasesData()
			soundMakerIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSoundMaker]
			if soundMakerIndex == -1 {
				// If no SoundMaker is selected, default to index 00 for settings
				soundMakerIndex = 0
				// Also set the value in the cell to 00 so it shows when we return
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSoundMaker] = 0
			}
			// Save current phrase view position
			m.LastPhraseRow = m.CurrentRow
			m.SoundMakerEditingIndex = soundMakerIndex
			m.ViewMode = types.SoundMakerView
			m.CurrentRow = 0 // Start at first setting
			m.CurrentCol = 0 // Start at Name column
			m.ScrollOffset = 0
			storage.AutoSave(m)
			return nil
		}

		// Check if we're in Instrument view - file selection doesn't apply
		if m.GetPhraseViewType() == types.InstrumentPhraseView {
			return nil // Shift+Right does nothing in Instrument view
		}

		// Navigate to file view from any other column in phrase view
		m.FileSelectRow = m.CurrentRow // Remember which row we're selecting for
		m.FileSelectCol = m.CurrentCol // Remember which column we were on

		// Try to navigate to the folder containing the current row's file
		selectedFilename := ""
		phrasesData := m.GetCurrentPhrasesData()
		fileIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if phrasesFiles != nil && fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
			// Get the directory of the current file
			fullPath := (*phrasesFiles)[fileIndex]
			fileDir := filepath.Dir(fullPath)
			selectedFilename = filepath.Base(fullPath) // Remember just the filename
			if fileDir != "." && fileDir != "" {
				// Change to the file's directory
				m.CurrentDir = fileDir
			}
			log.Printf("Navigating to file browser for file: %s in directory: %s", selectedFilename, m.CurrentDir)
		} else {
			log.Printf("No file for current row, using current directory: %s", m.CurrentDir)
		}

		m.ViewMode = types.FileView
		m.CurrentRow = 0
		m.CurrentCol = 0
		m.ScrollOffset = 0
		storage.LoadFiles(m)

		// After loading files, try to position cursor on the selected file
		if selectedFilename != "" {
			for i, filename := range m.Files {
				if filename == selectedFilename {
					m.CurrentRow = i
					// Adjust scroll offset to ensure the file is visible
					visibleRows := m.GetVisibleRows()
					if m.CurrentRow >= m.ScrollOffset+visibleRows {
						m.ScrollOffset = m.CurrentRow - visibleRows + 1
					} else if m.CurrentRow < m.ScrollOffset {
						m.ScrollOffset = m.CurrentRow
					}
					log.Printf("Positioned cursor on selected file '%s' at row %d", selectedFilename, i)
					break
				}
			}
		}

		storage.AutoSave(m)
	}
	return nil
}

func handleShiftUp(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		// Remember where we came from
		m.PreviousView = m.ViewMode

		// NEW: snapshot row/track for the view we're leaving
		switch m.ViewMode {
		case types.SongView:
			m.LastSongRow = m.CurrentRow
			m.LastSongTrack = m.CurrentCol
		case types.ChainView:
			m.LastChainRow = m.CurrentRow
		case types.PhraseView:
			m.LastPhraseRow = m.CurrentRow
		}

		switchToView(m, settingsViewConfig())
	} else if m.ViewMode == types.MixerView {
		// Navigate back to previous view from mixer
		switch m.PreviousView {
		case types.SongView:
			switchToViewWithVisibilityCheck(m, songViewConfig(m.LastSongRow, m.LastSongTrack))
		case types.ChainView:
			switchToViewWithVisibilityCheck(m, chainViewConfig(m.LastChainRow))
		case types.PhraseView:
			switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, 2))
		default:
			// Fallback to song view if no previous view is set
			switchToView(m, songViewConfig(0, 0))
		}
		return nil
	} else if m.ViewMode == types.FileView {
		// Navigate to file metadata view (only if a file is selected)
		if len(m.Files) > 0 && m.CurrentRow < len(m.Files) {
			selectedFile := m.Files[m.CurrentRow]
			// Don't open metadata for directories
			if !strings.HasSuffix(selectedFile, "/") && selectedFile != ".." {
				fullPath := filepath.Join(m.CurrentDir, selectedFile)
				m.MetadataEditingFile = fullPath
				switchToView(m, fileMetadataViewConfig())
				log.Printf("Opening metadata editor for file: %s", fullPath)
			}
		}
	}
	return nil
}

func handleShiftDown(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		// Remember where we came from for return navigation
		m.PreviousView = m.ViewMode

		// Snapshot row/track/position for the view we're leaving
		switch m.ViewMode {
		case types.SongView:
			m.LastSongRow = m.CurrentRow
			m.LastSongTrack = m.CurrentCol
		case types.ChainView:
			m.LastChainRow = m.CurrentRow
		case types.PhraseView:
			m.LastPhraseRow = m.CurrentRow
		}

		// Navigate to mixer view
		switchToView(m, mixerViewConfig())
		return nil
	} else if m.ViewMode == types.SettingsView {
		// Navigate back to previous view
		switch m.PreviousView {
		case types.SongView:
			switchToViewWithVisibilityCheck(m, songViewConfig(m.LastSongRow, m.LastSongTrack))
		case types.ChainView:
			// Keep CurrentChain as-is, just restore row/col
			cfg := chainViewConfig(m.LastChainRow)
			// chainViewConfig sets Col=1; if you prefer 0, change there
			switchToViewWithVisibilityCheck(m, cfg)
		case types.PhraseView:
			// Go back to the last phrase row; keep whatever column policy you want
			switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, 2))
		}
		return nil
	} else if m.ViewMode == types.FileMetadataView {
		// Navigate back to file view
		var targetRow int = 0
		// Try to position on the file we were editing
		if m.MetadataEditingFile != "" {
			filename := filepath.Base(m.MetadataEditingFile)
			for i, file := range m.Files {
				if file == filename {
					targetRow = i
					break
				}
			}
		}
		switchToView(m, ViewSwitchConfig{
			ViewMode:     types.FileView,
			Row:          targetRow,
			Col:          0,
			ScrollOffset: 0,
		})
		m.MetadataEditingFile = "" // Clear the editing file
	}
	return nil
}

func handleShiftLeft(m *model.Model) tea.Cmd {
	if m.ViewMode == types.ChainView {
		// Check if we came from Song view
		if m.LastSongRow >= 0 && m.LastSongTrack >= 0 {
			// Navigate back to song view
			m.ViewMode = types.SongView
			m.CurrentRow = m.LastSongRow
			m.CurrentCol = m.LastSongTrack
			m.ScrollOffset = 0

			log.Printf("Navigated back from Chain to Song (T%d R%02X)", m.CurrentCol, m.CurrentRow)
			storage.AutoSave(m)
			return nil
		}
		// Otherwise, stay in chain view (no song context)
	} else if m.ViewMode == types.PhraseView {
		// Navigate back to chain view
		// Remember current phrase row
		m.LastPhraseRow = m.CurrentRow
		log.Printf("Returning to Chain view, Chain = %d, LastChainRow = %d", m.CurrentChain, m.LastChainRow)
		// Return to the same position within the chain
		m.ViewMode = types.ChainView
		m.CurrentRow = m.LastChainRow
		m.CurrentCol = 0
		m.ScrollOffset = 0
		storage.AutoSave(m)
	} else if m.ViewMode == types.FileView {
		// Navigate back to phrase view - return to the column we came from
		switchToView(m, phraseViewConfig(m.FileSelectRow, m.FileSelectCol)) // Go back to original column
	} else if m.ViewMode == types.RetriggerView {
		// Navigate back to phrase view - find the UI column for RT
		var rtColumn int
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
			rtColumn = 1 // RT is not accessible in instrument view, default to P column
		} else {
			rtColumn = int(types.ColRetrigger) + 1 // Sampler: data column + 1
		}
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, rtColumn))
	} else if m.ViewMode == types.TimestrechView {
		// Navigate back to phrase view - find the UI column for TS
		var tsColumn int
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
			tsColumn = 1 // TS is not accessible in instrument view, default to P column
		} else {
			tsColumn = int(types.ColTimestretch) + 1 // Sampler: data column + 1
		}
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, tsColumn))
	} else if m.ViewMode == types.ArpeggioView {
		// Navigate back to phrase view - find the UI column for AR
		var arColumn int
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
			arColumn = int(types.InstrumentColAR) // AR column in instrument view
		} else {
			arColumn = 1 // AR is not accessible in sampler view, default to P column
		}
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, arColumn))
	} else if m.ViewMode == types.MidiView {
		// Navigate back to phrase view - find the UI column for MI
		var miColumn int
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
			miColumn = int(types.InstrumentColMI) // MI column in instrument view
		} else {
			miColumn = 1 // MI is not accessible in sampler view, default to P column
		}
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, miColumn))
	} else if m.ViewMode == types.SoundMakerView {
		// Navigate back to phrase view - find the UI column for SO
		var soColumn int
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
			soColumn = int(types.InstrumentColSO) // SO column in instrument view
		} else {
			soColumn = 1 // SO is not accessible in sampler view, default to P column
		}
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, soColumn))
	}
	return nil
}

func handleUp(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow > -1 { // Allow going up to row -1 (type row)
			m.CurrentRow = m.CurrentRow - 1
			if m.CurrentRow >= 0 { // Only update LastSongRow for data rows, not type row
				m.LastSongRow = m.CurrentRow
			}
		}
	} else if m.ViewMode == types.ChainView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.SettingsView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.FileMetadataView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.RetriggerView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.TimestrechView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.ArpeggioView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
		}
	} else if m.ViewMode == types.MidiView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
			if m.CurrentRow < m.ScrollOffset {
				m.ScrollOffset = m.CurrentRow
			}
		}
	} else if m.ViewMode == types.SoundMakerView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
			if m.CurrentRow < m.ScrollOffset {
				m.ScrollOffset = m.CurrentRow
			}
		}
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerRow > 0 {
			m.CurrentMixerRow = m.CurrentMixerRow - 1
		}
	} else if m.ViewMode == types.FileView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
			if m.CurrentRow < m.ScrollOffset {
				m.ScrollOffset = m.CurrentRow
			}
		}
	} else if m.CurrentRow > 0 {
		m.CurrentRow = m.CurrentRow - 1
		if m.CurrentRow < m.ScrollOffset {
			m.ScrollOffset = m.CurrentRow
		}
		// Update position tracking for view navigation
		if m.ViewMode == types.ChainView {
			m.LastChainRow = m.CurrentRow
		} else if m.ViewMode == types.PhraseView {
			m.LastPhraseRow = m.CurrentRow
		}
	}
	storage.AutoSave(m)
	return nil
}

func handleDown(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow < 15 { // Song view has 16 rows (0-15), plus type row at -1
			m.CurrentRow = m.CurrentRow + 1
			if m.CurrentRow >= 0 { // Only update LastSongRow for data rows, not type row
				m.LastSongRow = m.CurrentRow
			}
		}
	} else if m.ViewMode == types.ChainView {
		if m.CurrentRow < 15 { // Chain view has 16 rows (0-15)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.SettingsView {
		// Column 0 (Global): BPM to Drive, Column 1 (Input): InputLevelDB to ReverbSendPercent
		var maxRow int
		if m.CurrentCol == 0 {
			maxRow = int(types.GlobalSettingsRowDriveDB) // Global column: BPM(0) to Drive(6)
		} else {
			maxRow = int(types.InputSettingsRowReverbSendPercent) // Input column: InputLevelDB(0) to ReverbSendPercent(1)
		}
		if m.CurrentRow < maxRow {
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.FileMetadataView {
		if m.CurrentRow < int(types.FileMetadataRowSlices) { // BPM(0) to Slices(1)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.RetriggerView {
		if m.CurrentRow < int(types.RetriggerSettingsRowFinalVolumeToStart) { // Times(0) to FinalVolumeToStart(7)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.TimestrechView {
		if m.CurrentRow < int(types.TimestrechSettingsRowBeats) { // Start(0) to Beats(2)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.ArpeggioView {
		if m.CurrentRow < 15 { // 0-15 (16 rows total)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.MidiView {
		// Calculate maximum row: 2 settings rows + available MIDI devices
		maxRow := int(types.MidiSettingsRowChannel) + len(m.AvailableMidiDevices) // Device(0), Channel(1), then devices starting at row 2
		if m.CurrentRow < maxRow {
			m.CurrentRow = m.CurrentRow + 1
			visibleRows := m.GetVisibleRows()
			if m.CurrentRow >= m.ScrollOffset+visibleRows {
				m.ScrollOffset = m.CurrentRow - visibleRows + 1
			}
		}
	} else if m.ViewMode == types.SoundMakerView {
		// Calculate maximum row: 5 settings rows + available SoundMakers
		availableSoundMakers := []string{"Polyperc", "Infinite Pad"}
		maxRow := 4 + len(availableSoundMakers) // Name(0), A(1), B(2), C(3), D(4), then SoundMakers starting at row 5
		if m.CurrentRow < maxRow {
			m.CurrentRow = m.CurrentRow + 1
			visibleRows := m.GetVisibleRows()
			if m.CurrentRow >= m.ScrollOffset+visibleRows {
				m.ScrollOffset = m.CurrentRow - visibleRows + 1
			}
		}
	} else if m.ViewMode == types.MixerView {
		// Only row 0 (set level) exists now, no navigation needed
		// Keep CurrentMixerRow at 0
	} else if m.ViewMode == types.FileView {
		// Ensure we don't go beyond the last file
		if len(m.Files) > 0 && m.CurrentRow < len(m.Files)-1 {
			m.CurrentRow = m.CurrentRow + 1
			visibleRows := m.GetVisibleRows()
			if m.CurrentRow >= m.ScrollOffset+visibleRows {
				m.ScrollOffset = m.CurrentRow - visibleRows + 1
			}
		}
	} else if m.CurrentRow < 254 { // 0-254 (FE in hex)
		m.CurrentRow = m.CurrentRow + 1
		visibleRows := m.GetVisibleRows()
		if m.CurrentRow >= m.ScrollOffset+visibleRows {
			m.ScrollOffset = m.CurrentRow - visibleRows + 1
		}
		// Update position tracking for view navigation
		if m.ViewMode == types.ChainView {
			m.LastChainRow = m.CurrentRow
		} else if m.ViewMode == types.PhraseView {
			m.LastPhraseRow = m.CurrentRow
		}
	}
	storage.AutoSave(m)
	return nil
}

func handleLeft(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentCol > 0 { // 8 tracks: 0-7
			m.CurrentCol = m.CurrentCol - 1
			m.LastSongTrack = m.CurrentCol
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ChainView {
		if m.CurrentChain > 0 { // Switch to previous chain
			m.CurrentChain = m.CurrentChain - 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.PhraseView {
		// Handle different column limits based on view type
		phraseViewType := m.GetPhraseViewType()
		minCol := 1 // Both views: Column 1 is the first editable column
		if phraseViewType == types.InstrumentPhraseView {
			minCol = 1 // Instrument: Column 1 is NOT (note)
		} else {
			minCol = 1 // Sampler: Column 1 is P (playback)
		}

		if m.CurrentCol > minCol {
			m.CurrentCol = m.CurrentCol - 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ArpeggioView {
		if m.CurrentCol > int(types.ArpeggioColDI) { // 3 columns: DI, CO, Divisor
			m.CurrentCol = m.CurrentCol - 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.MidiView {
		// No horizontal navigation in MIDI view - use up/down for settings
	} else if m.ViewMode == types.SoundMakerView {
		// No horizontal navigation in SoundMaker view - use up/down for settings
	} else if m.ViewMode == types.SettingsView {
		if m.CurrentCol > 0 { // Switch between Global (0) and Input (1) columns
			m.CurrentCol = m.CurrentCol - 1
			// Adjust row if it's beyond the bounds of the new column
			if m.CurrentCol == 0 && m.CurrentRow > 6 {
				m.CurrentRow = 6 // Global column max is 6
			}
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerTrack > 0 { // Select previous track (0-7)
			m.CurrentMixerTrack = m.CurrentMixerTrack - 1
			storage.AutoSave(m)
		}
	} else { // FileView
		// No horizontal navigation in file view
	}
	return nil
}

func handleRight(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentCol < 7 { // 8 tracks: 0-7
			m.CurrentCol = m.CurrentCol + 1
			m.LastSongTrack = m.CurrentCol
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ChainView {
		if m.CurrentChain < 254 { // Switch to next chain (0-254)
			m.CurrentChain = m.CurrentChain + 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.PhraseView {
		// Handle different column limits based on view type
		phraseViewType := m.GetPhraseViewType()
		var maxValidCol int
		if phraseViewType == types.InstrumentPhraseView {
			maxValidCol = int(types.InstrumentColSO) // Instrument: last valid column is SO (SoundMaker)
		} else {
			maxValidCol = 13 // Sampler: last valid column is 13 (FI)
		}

		if m.CurrentCol < maxValidCol {
			m.CurrentCol = m.CurrentCol + 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ArpeggioView {
		if m.CurrentCol < int(types.ArpeggioColDIV) { // 3 columns: DI, CO, Divisor
			m.CurrentCol = m.CurrentCol + 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.MidiView {
		// No horizontal navigation in MIDI view - use up/down for settings
	} else if m.ViewMode == types.SoundMakerView {
		// No horizontal navigation in SoundMaker view - use up/down for settings
	} else if m.ViewMode == types.SettingsView {
		if m.CurrentCol < 1 { // Switch between Global (0) and Input (1) columns
			m.CurrentCol = m.CurrentCol + 1
			// Adjust row if it's beyond the bounds of the new column
			if m.CurrentCol == 1 && m.CurrentRow > 1 {
				m.CurrentRow = 1 // Input column max is 1
			}
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerTrack < 8 { // Select next track (0-8, including Input track)
			m.CurrentMixerTrack = m.CurrentMixerTrack + 1
			storage.AutoSave(m)
		}
	} else { // FileView
		// No horizontal navigation in file view
	}
	return nil
}

func handleCtrlS(m *model.Model) tea.Cmd {
	storage.DoSave(m)
	return nil
}

func handleCtrlUp(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow == -1 {
			// Toggle track type when on type row
			ToggleTrackType(m, m.CurrentCol)
		} else {
			ModifySongValue(m, 4) // Coarse increment for song view
		}
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, 1.0)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, 1.0)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, 1.0)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, 1.0)
	} else if m.ViewMode == types.ArpeggioView {
		ModifyArpeggioValue(m, 1.0)
	} else if m.ViewMode == types.MidiView {
		ModifyMidiValue(m, 1.0)
	} else if m.ViewMode == types.SoundMakerView {
		ModifySoundMakerValue(m, 1.0)
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerRow == 0 {
			ModifyMixerSetLevel(m, 1.0) // Coarse increment for set level
		}
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, 16)
	}
	return nil
}

func handleCtrlDown(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow == -1 {
			// Toggle track type when on type row
			ToggleTrackType(m, m.CurrentCol)
		} else {
			ModifySongValue(m, -4) // Coarse decrement for song view
		}
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, -1.0)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, -1.0)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, -1.0)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, -1.0)
	} else if m.ViewMode == types.ArpeggioView {
		ModifyArpeggioValue(m, -1.0)
	} else if m.ViewMode == types.MidiView {
		ModifyMidiValue(m, -1.0)
	} else if m.ViewMode == types.SoundMakerView {
		ModifySoundMakerValue(m, -1.0)
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerRow == 0 {
			ModifyMixerSetLevel(m, -1.0) // Coarse decrement for set level
		}
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, -16)
	}
	return nil
}

func handleCtrlLeft(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow == -1 {
			// Toggle track type when on type row
			ToggleTrackType(m, m.CurrentCol)
		} else {
			ModifySongValue(m, -1) // Fine decrement for song view
		}
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, -0.05)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, -0.05)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, -0.05)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, -0.05)
	} else if m.ViewMode == types.ArpeggioView {
		ModifyArpeggioValue(m, -0.05)
	} else if m.ViewMode == types.MidiView {
		ModifyMidiValue(m, -0.05)
	} else if m.ViewMode == types.SoundMakerView {
		ModifySoundMakerValue(m, -0.05)
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerRow == 0 {
			ModifyMixerSetLevel(m, -0.05) // Fine decrement for set level
		}
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, -1)
	}
	return nil
}

func handleCtrlRight(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow == -1 {
			// Toggle track type when on type row
			ToggleTrackType(m, m.CurrentCol)
		} else {
			ModifySongValue(m, 1) // Fine increment for song view
		}
	} else if m.ViewMode == types.FileView {
		audio.PlayFile(m)
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, 0.05)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, 0.05)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, 0.05)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, 0.05)
	} else if m.ViewMode == types.ArpeggioView {
		ModifyArpeggioValue(m, 0.05)
	} else if m.ViewMode == types.MidiView {
		ModifyMidiValue(m, 0.05)
	} else if m.ViewMode == types.SoundMakerView {
		ModifySoundMakerValue(m, 0.05)
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerRow == 0 {
			ModifyMixerSetLevel(m, 0.05) // Fine increment for set level
		}
	} else {
		ModifyValue(m, 1)
	}
	return nil
}

func handleS(m *model.Model) tea.Cmd {
	if m.ViewMode != types.FileView && m.ViewMode != types.SettingsView && m.ViewMode != types.FileMetadataView {
		PasteLastEditedRow(m)
		storage.AutoSave(m)
	}
	return nil
}

// internal/input/input.go

func handleC(m *model.Model) tea.Cmd {
	// NEW: if playback is running, stop it (mirror TogglePlayback's stop path)
	if m.IsPlaying {
		m.IsPlaying = false

		// Stop recording if active (same behavior as TogglePlayback)
		if m.RecordingActive {
			stopRecording(m)
		}

		// Clear file browser playback state when stopping tracker playback
		if m.CurrentlyPlayingFile != "" {
			m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
			m.CurrentlyPlayingFile = ""
			log.Printf("Stopped file browser playback when stopping via 'C'")
		}

		// Send OSC "/stop" with no params (tiny helper on the model)
		m.SendStopOSC()

		log.Printf("Playback stopped via 'C'")
		return nil
	}

	// Existing behavior (unchanged)
	if m.ViewMode == types.PhraseView {
		if IsRowEmpty(m) {
			// If row is empty, do the copy operation
			CopyLastRowWithIncrement(m)
			storage.AutoSave(m)
		} else {
			// If row is not empty, emit the row data
			EmitRowData(m)
		}
	} else if m.ViewMode == types.ChainView {
		chainsData := m.GetCurrentChainsData()
		if (*chainsData)[m.CurrentChain][m.CurrentRow] == -1 {
			// If chain slot is empty, fill it with next unused phrase
			seed := 254 // 254 => first check will be 0 (wrap-around)
			if m.CurrentRow > 0 && (*chainsData)[m.CurrentChain][m.CurrentRow-1] != -1 {
				seed = (*chainsData)[m.CurrentChain][m.CurrentRow-1]
			}

			next := FindNextUnusedPhrase(m, seed)
			if next == -1 {
				log.Printf("No unused phrases available")
				return nil
			}

			(*chainsData)[m.CurrentChain][m.CurrentRow] = next
			log.Printf("Filled Chain %02X Row %02X with next empty phrase %02X",
				m.CurrentChain, m.CurrentRow, next)
			storage.AutoSave(m)
		} else {
			// If chain slot is not empty, emit the phrase data for that slot
			phraseNumber := (*chainsData)[m.CurrentChain][m.CurrentRow]
			EmitRowDataFor(m, phraseNumber, 0, m.CurrentTrack) // Emit first row of the phrase
			log.Printf("Emitting data for Chain %02X Row %02X -> Phrase %02X",
				m.CurrentChain, m.CurrentRow, phraseNumber)
		}
		return nil
	} else if m.ViewMode == types.SongView {
		track := m.CurrentCol
		row := m.CurrentRow

		if m.SongData[track][row] == -1 {
			// If song slot is empty, fill it with next unused chain
			seed := 254 // 254 => first check will be 0 (wrap-around)
			if row > 0 && m.SongData[track][row-1] != -1 {
				seed = m.SongData[track][row-1]
			}

			next := FindNextUnusedChain(m, seed)
			if next == -1 {
				log.Printf("No unused chains available")
				return nil
			}

			m.SongData[track][row] = next
			log.Printf("Filled Song T%d R%02X with next empty chain %02X", track, row, next)
			storage.AutoSave(m)
		} else {
			// If song slot is not empty, emit the chain data for that slot
			chainNumber := m.SongData[track][row]
			// Find the first non-empty phrase in this chain
			chainsData := m.GetChainsDataForTrack(track)
			firstPhraseNumber := -1
			for chainRow := 0; chainRow < 16; chainRow++ {
				if (*chainsData)[chainNumber][chainRow] != -1 {
					firstPhraseNumber = (*chainsData)[chainNumber][chainRow]
					break
				}
			}

			if firstPhraseNumber != -1 {
				EmitRowDataFor(m, firstPhraseNumber, 0, track) // Emit first row of the first phrase
				log.Printf("Emitting data for Song T%d R%02X -> Chain %02X -> Phrase %02X",
					track, row, chainNumber, firstPhraseNumber)
			} else {
				log.Printf("Chain %02X is empty, cannot emit data", chainNumber)
			}
		}
		return nil
	} else if m.ViewMode == types.RetriggerView {
		EmitLastSelectedPhraseRowData(m)
	} else if m.ViewMode == types.ArpeggioView {
		// Play the last edited phrase row from Instrument view
		EmitLastSelectedPhraseRowData(m)
	}
	return nil
}

func handleCtrlC(m *model.Model) tea.Cmd {
	CopyCellToClipboard(m)
	return nil
}

func handleCtrlX(m *model.Model) tea.Cmd {
	CutRowToClipboard(m)
	storage.AutoSave(m)
	return nil
}

func handleCtrlV(m *model.Model) tea.Cmd {
	PasteFromClipboard(m)
	storage.AutoSave(m)
	return nil
}

func handleCtrlD(m *model.Model) tea.Cmd {
	DeepCopyToClipboard(m)
	storage.AutoSave(m)
	return nil
}

func handleSpace(m *model.Model) tea.Cmd {
	if m.ViewMode == types.FileView {
		audio.SelectFile(m)
		return nil
	} else if m.ViewMode == types.MidiView {
		// Handle device selection in MIDI view
		if m.CurrentRow >= 2 && m.CurrentRow-2+m.ScrollOffset < len(m.AvailableMidiDevices) {
			deviceIndex := m.CurrentRow - 2 + m.ScrollOffset
			selectedDevice := m.AvailableMidiDevices[deviceIndex]
			m.MidiSettings[m.MidiEditingIndex].Device = selectedDevice
			log.Printf("Selected MIDI device: %s for MIDI %02X", selectedDevice, m.MidiEditingIndex)
			storage.AutoSave(m)
		}
		return nil
	} else if m.ViewMode == types.SoundMakerView {
		// Handle SoundMaker selection in SoundMaker view
		availableSoundMakers := []string{"Polyperc", "Infinite Pad"}
		if m.CurrentRow >= 5 && m.CurrentRow-5+m.ScrollOffset < len(availableSoundMakers) {
			soundMakerIndex := m.CurrentRow - 5 + m.ScrollOffset
			selectedSoundMaker := availableSoundMakers[soundMakerIndex]
			m.SoundMakerSettings[m.SoundMakerEditingIndex].Name = selectedSoundMaker
			log.Printf("Selected SoundMaker: %s for SoundMaker %02X", selectedSoundMaker, m.SoundMakerEditingIndex)
			storage.AutoSave(m)
		}
		return nil
	} else if m.ViewMode != types.SettingsView && m.ViewMode != types.FileMetadataView {
		return TogglePlayback(m)
	}
	return nil
}

func handleCtrlSpace(m *model.Model) tea.Cmd {
	// Ctrl+Space always plays ALL tracks from last selected Song view row, regardless of current view
	return TogglePlaybackFromLastSongRow(m)
}

func handleBackspace(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		// Clear chain ID in song view
		m.SongData[m.CurrentCol][m.CurrentRow] = -1
		log.Printf("Cleared song track %d row %02X chain", m.CurrentCol, m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.ChainView {
		// Clear phrase number in chain view (works from any column)
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[m.CurrentChain][m.CurrentRow] = -1
		log.Printf("Cleared chain %d phrase", m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.PhraseView {
		// Clear the current cell in phrase view
		phrasesData := m.GetCurrentPhrasesData()

		// Use centralized column mapping system
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping == nil || !columnMapping.IsDeletable {
			return nil // Invalid or non-deletable column
		}

		colIndex := columnMapping.DataColumnIndex

		if colIndex >= 0 && colIndex < int(types.ColCount) {
			if colIndex == int(types.ColDeltaTime) {
				// Reset DT to -1 (means skip/not played)
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = -1
			} else {
				// Clear all other columns to -1 (including GT, PI, RT, etc.)
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = -1
			}
			log.Printf("Cleared phrase %d row %d col %d", m.CurrentPhrase, m.CurrentRow, colIndex)
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.ArpeggioView {
		// Clear the current cell in arpeggio view
		settings := &m.ArpeggioSettings[m.ArpeggioEditingIndex]
		currentRow := &settings.Rows[m.CurrentRow]

		switch m.CurrentCol {
		case 0: // Direction column
			currentRow.Direction = int(types.ArpeggioDirectionNone) // Clear to "--"
			log.Printf("Cleared arpeggio %02X row %02X Direction", m.ArpeggioEditingIndex, m.CurrentRow)
		case 1: // Count column
			currentRow.Count = -1 // Clear to "--"
			log.Printf("Cleared arpeggio %02X row %02X Count", m.ArpeggioEditingIndex, m.CurrentRow)
		case 2: // Divisor column
			currentRow.Divisor = -1 // Clear to "--"
			log.Printf("Cleared arpeggio %02X row %02X Divisor", m.ArpeggioEditingIndex, m.CurrentRow)
		}
		storage.AutoSave(m)
	}
	return nil
}

func handleCtrlH(m *model.Model) tea.Cmd {
	if m.ViewMode == types.ChainView {
		// Delete entire chain row (clear phrase, keep chain number)
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[m.CurrentChain][m.CurrentRow] = -1
		log.Printf("Deleted chain %d row (cleared phrase)", m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.PhraseView {
		// Delete entire phrase row (clear all columns)
		phrasesData := m.GetCurrentPhrasesData()
		phraseViewType := m.GetPhraseViewType()
		if phraseViewType == types.InstrumentPhraseView {
		}
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1      // Clear note
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColPitch)] = -1     // Clear pitch (displays "--", behaves as 80)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1 // Clear deltatime (for samplers this also clears playback)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColGate)] = -1      // Clear gate (displays "--", behaves as 80)
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColRetrigger)] = -1 // Clear retrigger
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = -1  // Clear filename
		log.Printf("Deleted phrase %d row %d (cleared all columns)", m.CurrentPhrase, m.CurrentRow)
		storage.AutoSave(m)
	}
	return nil
}

func handleCtrlR(m *model.Model) tea.Cmd {
	// Toggle recording enabled state
	m.RecordingEnabled = !m.RecordingEnabled

	if m.RecordingEnabled {
		log.Printf("Recording enabled (queued)")
		// If playback is already active, start recording immediately
		if m.IsPlaying {
			startRecording(m)
		}
	} else {
		log.Printf("Recording disabled")
		// If recording is currently active, stop it
		if m.RecordingActive {
			stopRecording(m)
		}
	}

	storage.AutoSave(m)
	return nil
}

func startRecording(m *model.Model) {
	if !m.RecordingEnabled || m.RecordingActive {
		return
	}

	// Generate timestamped filename
	filename := m.GenerateRecordingFilename()
	m.CurrentRecordingFile = filename
	m.RecordingActive = true

	// Send OSC message to start recording
	m.SendOSCRecordMessage(filename, true)
	log.Printf("Recording started: %s", filename)
}

func stopRecording(m *model.Model) {
	if !m.RecordingActive || m.CurrentRecordingFile == "" {
		return
	}

	// Send OSC message to stop recording with the same filename
	m.SendOSCRecordMessage(m.CurrentRecordingFile, false)
	log.Printf("Recording stopped: %s", m.CurrentRecordingFile)

	// Reset recording state but keep enabled flag
	m.RecordingActive = false
	m.CurrentRecordingFile = ""
}

func handleCtrlF(m *model.Model) tea.Cmd {
	FillSequential(m)
	storage.AutoSave(m)
	return nil
}

func handleP(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SettingsView {
		// If we're in settings, act like Shift+Down (go back to previous view)
		return handleShiftDown(m)
	} else if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		// If we're in Song, Chain, or Phrase view, act like Shift+Up (go to settings)
		return handleShiftUp(m)
	}
	// For other views (FileView, MixerView, etc.), do nothing
	return nil
}

func handleM(m *model.Model) tea.Cmd {
	if m.ViewMode == types.MixerView {
		// If we're in mixer view, act like Shift+Up (go back to previous view)
		return handleShiftUp(m)
	} else if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		// If we're in Song, Chain, or Phrase view, act like Shift+Down (go to mixer)
		return handleShiftDown(m)
	}
	// For other views (FileView, SettingsView, etc.), do nothing
	return nil
}

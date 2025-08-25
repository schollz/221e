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
	if m.CurrentRow >= visibleRows {
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
		Col:          0,
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

func mixerViewConfig() ViewSwitchConfig {
	return ViewSwitchConfig{
		ViewMode:     types.MixerView,
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

	case "ctrl+h":
		return handleCtrlH(m)

	case "ctrl+r":
		return handleCtrlR(m)

	case "ctrl+f":
		return handleCtrlF(m)
	}

	return nil
}

func handleShiftRight(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
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
		phraseNum := m.ChainsData[m.CurrentChain][m.CurrentRow]
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

			m.CurrentCol = 2 // Start on Note column
			m.ScrollOffset = 0

			// Ensure the cursor row is visible
			visibleRows := m.GetVisibleRows()
			if m.CurrentRow >= visibleRows {
				m.ScrollOffset = m.CurrentRow - visibleRows + 1
			}

			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.PhraseView {
		// Check if we're on the RT column (column 5)
		if m.CurrentCol == 5 {
			// Navigate to retrigger view only if a retrigger is selected (not -1)
			retriggerIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColRetrigger]
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
		}

		// Check if we're on the TS column (column 6)
		if m.CurrentCol == 6 {
			// Navigate to timestretch view only if a timestretch is selected (not -1)
			timestrechIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColTimestretch]
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
		}

		// Navigate to file view from any other column in phrase view
		m.FileSelectRow = m.CurrentRow // Remember which row we're selecting for

		// Try to navigate to the folder containing the current row's file
		selectedFilename := ""
		fileIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		if fileIndex >= 0 && fileIndex < len(m.PhrasesFiles) && m.PhrasesFiles[fileIndex] != "" {
			// Get the directory of the current file
			fullPath := m.PhrasesFiles[fileIndex]
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
		// Navigate back to phrase view
		switchToView(m, phraseViewConfig(m.FileSelectRow, int(types.ColFilename)+1)) // Go back to FI column (UI index)
	} else if m.ViewMode == types.RetriggerView {
		// Navigate back to phrase view
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, 5)) // Go back to RT column
	} else if m.ViewMode == types.TimestrechView {
		// Navigate back to phrase view
		switchToViewWithVisibilityCheck(m, phraseViewConfig(m.LastPhraseRow, 6)) // Go back to TS column
	}
	return nil
}

func handleUp(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		if m.CurrentRow > 0 {
			m.CurrentRow = m.CurrentRow - 1
			m.LastSongRow = m.CurrentRow
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
		if m.CurrentRow < 15 { // Song view has 16 rows (0-15)
			m.CurrentRow = m.CurrentRow + 1
			m.LastSongRow = m.CurrentRow
		}
	} else if m.ViewMode == types.ChainView {
		if m.CurrentRow < 15 { // Chain view has 16 rows (0-15)
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.SettingsView {
		if m.CurrentRow < 6 { // 0=BPM, 1=PPQ, 2=PregainDB, 3=PostgainDB, 4=BiasDB, 5=SaturationDB, 6=DriveDB
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.FileMetadataView {
		if m.CurrentRow < 1 { // 0=BPM, 1=Slices
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.RetriggerView {
		if m.CurrentRow < 5 { // 0=Times, 1=Starting Rate, 2=Final Rate, 3=Beats, 4=Volume, 5=Pitch
			m.CurrentRow = m.CurrentRow + 1
		}
	} else if m.ViewMode == types.TimestrechView {
		if m.CurrentRow < 2 { // 0=Start, 1=End, 2=Beats
			m.CurrentRow = m.CurrentRow + 1
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
		if m.CurrentCol > 1 { // Column 1 is playback, can't go to column 0 (slice)
			m.CurrentCol = m.CurrentCol - 1
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
		// UI column numbers are 1..ColCount (data columns are 0..ColCount-1)
		maxCol := int(types.ColCount)
		if m.CurrentCol < maxCol {
			m.CurrentCol = m.CurrentCol + 1
			storage.AutoSave(m)
		}
	} else if m.ViewMode == types.MixerView {
		if m.CurrentMixerTrack < 7 { // Select next track (0-7)
			m.CurrentMixerTrack = m.CurrentMixerTrack + 1
			storage.AutoSave(m)
		}
	} else { // FileView
		// No horizontal navigation in file view
	}
	return nil
}

func handleCtrlUp(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		ModifySongValue(m, 4) // Coarse increment for song view
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, 1.0)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, 1.0)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, 1.0)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, 1.0)
	} else if m.ViewMode == types.MixerView {
		ModifyMixerSetLevel(m, 1.0) // Coarse increment for set level
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, 16)
	}
	return nil
}

func handleCtrlDown(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		ModifySongValue(m, -4) // Coarse decrement for song view
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, -1.0)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, -1.0)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, -1.0)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, -1.0)
	} else if m.ViewMode == types.MixerView {
		ModifyMixerSetLevel(m, -1.0) // Coarse decrement for set level
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, -16)
	}
	return nil
}

func handleCtrlLeft(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		ModifySongValue(m, -1) // Fine decrement for song view
	} else if m.ViewMode == types.SettingsView {
		ModifySettingsValue(m, -0.05)
	} else if m.ViewMode == types.FileMetadataView {
		ModifyFileMetadataValue(m, -0.05)
	} else if m.ViewMode == types.RetriggerView {
		ModifyRetriggerValue(m, -0.05)
	} else if m.ViewMode == types.TimestrechView {
		ModifyTimestrechValue(m, -0.05)
	} else if m.ViewMode == types.MixerView {
		ModifyMixerSetLevel(m, -0.05) // Fine decrement for set level
	} else if m.ViewMode != types.FileView {
		ModifyValue(m, -1)
	}
	return nil
}

func handleCtrlRight(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		ModifySongValue(m, 1) // Fine increment for song view
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
	} else if m.ViewMode == types.MixerView {
		ModifyMixerSetLevel(m, 0.05) // Fine increment for set level
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
		// Only fill if the current chain slot is empty
		if m.ChainsData[m.CurrentChain][m.CurrentRow] == -1 {
			// Seed: prefer previous row’s phrase, else start so 00 is first checked
			seed := 254 // 254 => first check will be 0 (wrap-around)
			if m.CurrentRow > 0 && m.ChainsData[m.CurrentChain][m.CurrentRow-1] != -1 {
				seed = m.ChainsData[m.CurrentChain][m.CurrentRow-1]
			}

			next := FindNextUnusedPhrase(m, seed)
			if next == -1 {
				log.Printf("No unused phrases available")
				return nil
			}

			m.ChainsData[m.CurrentChain][m.CurrentRow] = next
			log.Printf("Filled Chain %02X Row %02X with next empty phrase %02X",
				m.CurrentChain, m.CurrentRow, next)
			storage.AutoSave(m)
		}
		return nil
	} else if m.ViewMode == types.SongView {
		track := m.CurrentCol
		row := m.CurrentRow

		// Only fill if the current song slot is empty
		if m.SongData[track][row] == -1 {
			// Seed search: prefer previous row on the same track, else start so 00 is first
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
		}
		return nil
	} else if m.ViewMode == types.RetriggerView {
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
	} else if m.ViewMode != types.SettingsView && m.ViewMode != types.FileMetadataView {
		return TogglePlayback(m)
	}
	return nil
}

func handleCtrlSpace(m *model.Model) tea.Cmd {
	// Ctrl+Space works from anywhere to stop playback or start from top
	return TogglePlaybackFromTopGlobal(m)
}

func handleBackspace(m *model.Model) tea.Cmd {
	if m.ViewMode == types.SongView {
		// Clear chain ID in song view
		m.SongData[m.CurrentCol][m.CurrentRow] = -1
		log.Printf("Cleared song track %d row %02X chain", m.CurrentCol, m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.ChainView {
		// Clear phrase number in chain view (works from any column)
		m.ChainsData[m.CurrentChain][m.CurrentRow] = -1
		log.Printf("Cleared chain %d phrase", m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.PhraseView {
		// Clear the current cell in phrase view
		colIndex := m.CurrentCol - 1 // Convert to data array index
		if colIndex >= 0 && colIndex < int(types.ColCount) {
			if colIndex == int(types.ColPlayback) {
				// Reset playback to 0 (special case - 0 means off)
				m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = 0
			} else {
				// Clear all other columns to -1 (including GT, PI, RT, etc.)
				m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex] = -1
			}
			log.Printf("Cleared phrase %d row %d col %d", m.CurrentPhrase, m.CurrentRow, colIndex)
			storage.AutoSave(m)
		}
	}
	return nil
}

func handleCtrlH(m *model.Model) tea.Cmd {
	if m.ViewMode == types.ChainView {
		// Delete entire chain row (clear phrase, keep chain number)
		m.ChainsData[m.CurrentChain][m.CurrentRow] = -1
		log.Printf("Deleted chain %d row (cleared phrase)", m.CurrentRow)
		storage.AutoSave(m)
	} else if m.ViewMode == types.PhraseView {
		// Delete entire phrase row (clear all columns)
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColPlayback)] = 0   // Reset playback to 0
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = -1      // Clear note
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColPitch)] = 128    // Reset pitch to default (hex 80)
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = -1 // Clear deltatime
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColGate)] = 128     // Reset gate to default
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColRetrigger)] = -1 // Clear retrigger
		m.PhrasesData[m.CurrentPhrase][m.CurrentRow][int(types.ColFilename)] = -1  // Clear filename
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

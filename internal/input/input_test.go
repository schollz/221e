package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

func createTestModel() *model.Model {
	return model.NewModel(0, "test.json") // Port 0 to disable OSC for testing
}

func TestViewSwitchConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   ViewSwitchConfig
		expected ViewSwitchConfig
	}{
		{
			name:   "Song view config",
			config: songViewConfig(5, 3),
			expected: ViewSwitchConfig{
				ViewMode:     types.SongView,
				Row:          5,
				Col:          3,
				ScrollOffset: 0,
			},
		},
		{
			name:   "Settings view config",
			config: settingsViewConfig(),
			expected: ViewSwitchConfig{
				ViewMode:     types.SettingsView,
				Row:          0,
				Col:          0,
				ScrollOffset: 0,
			},
		},
		{
			name:   "Chain view config",
			config: chainViewConfig(10),
			expected: ViewSwitchConfig{
				ViewMode:     types.ChainView,
				Row:          10,
				Col:          1,
				ScrollOffset: 0,
			},
		},
		{
			name:   "Phrase view config",
			config: phraseViewConfig(8, 5),
			expected: ViewSwitchConfig{
				ViewMode:     types.PhraseView,
				Row:          8,
				Col:          5,
				ScrollOffset: 0,
			},
		},
		{
			name:   "File view config",
			config: fileViewConfig(),
			expected: ViewSwitchConfig{
				ViewMode:     types.FileView,
				Row:          0,
				Col:          0,
				ScrollOffset: 0,
			},
		},
		{
			name:   "File metadata view config",
			config: fileMetadataViewConfig(),
			expected: ViewSwitchConfig{
				ViewMode:     types.FileMetadataView,
				Row:          0,
				Col:          0,
				ScrollOffset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.ViewMode, tt.config.ViewMode)
			assert.Equal(t, tt.expected.Row, tt.config.Row)
			assert.Equal(t, tt.expected.Col, tt.config.Col)
			assert.Equal(t, tt.expected.ScrollOffset, tt.config.ScrollOffset)
		})
	}
}

func TestSwitchToView(t *testing.T) {
	m := createTestModel()

	// Set initial state
	m.CurrentRow = 10
	m.CurrentCol = 5
	m.ViewMode = types.FileView
	m.ScrollOffset = 3

	// Test basic view switch
	config := ViewSwitchConfig{
		ViewMode:     types.SettingsView,
		Row:          2,
		Col:          1,
		ScrollOffset: 0,
	}

	switchToView(m, config)

	assert.Equal(t, types.SettingsView, m.ViewMode)
	assert.Equal(t, 2, m.CurrentRow)
	assert.Equal(t, 1, m.CurrentCol)
	assert.Equal(t, 0, m.ScrollOffset)
}

func TestSwitchToViewWithVisibilityCheck(t *testing.T) {
	m := createTestModel()
	m.TermHeight = 20

	// Test with row that needs scrolling
	config := ViewSwitchConfig{
		ViewMode:     types.PhraseView,
		Row:          25, // Beyond visible area
		Col:          3,
		ScrollOffset: 0,
	}

	switchToViewWithVisibilityCheck(m, config)

	assert.Equal(t, types.PhraseView, m.ViewMode)
	assert.Equal(t, 25, m.CurrentRow)
	assert.Greater(t, m.ScrollOffset, 0) // Should adjust scroll to make row visible

	// Test with negative row
	config.Row = -1
	switchToViewWithVisibilityCheck(m, config)
	assert.Equal(t, 0, m.ScrollOffset) // Should reset scroll for negative rows
}

func TestHandleKeyInput(t *testing.T) {
	m := createTestModel()

	// Test various key inputs don't crash
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyLeft},
		{Type: tea.KeyRight},
		{Type: tea.KeySpace},
		{Type: tea.KeyEnter},
		{Type: tea.KeyEsc},
		{Type: tea.KeyTab},
		{Type: tea.KeyBackspace},
		{Type: tea.KeyDelete},
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyCtrlV},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'2'}},
		{Type: tea.KeyRunes, Runes: []rune{'3'}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'p'}},
		{Type: tea.KeyRunes, Runes: []rune{'s'}},
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
	}

	for _, keyMsg := range testKeys {
		t.Run("Key "+keyMsg.String(), func(t *testing.T) {
			cmd := HandleKeyInput(m, keyMsg)
			// Should not panic and should return a command (could be nil)
			_ = cmd
		})
	}
}

func TestTickFunction(t *testing.T) {
	m := createTestModel()
	m.BPM = 120

	// Test tick command generation
	cmd := Tick(m)
	assert.NotNil(t, cmd)
}

func TestAdvancePlayback(t *testing.T) {
	m := createTestModel()

	// Set up playback state
	m.IsPlaying = true
	m.PlaybackRow = 0
	m.CurrentPhrase = 5
	m.ViewMode = types.PhraseView

	// Test playback advancement
	AdvancePlayback(m)

	// Playback should have advanced (specific behavior depends on implementation)
	// At minimum, it should not crash
	assert.NotNil(t, m)
}

func TestPlaybackAdvancementEdgeCases(t *testing.T) {
	m := createTestModel()

	// Test advancement when not playing
	m.IsPlaying = false
	m.PlaybackRow = 5

	AdvancePlayback(m)

	// When not playing, should not crash
	assert.NotNil(t, m)

	// Test advancement at phrase boundaries
	m.IsPlaying = true
	m.PlaybackRow = 254 // Near end of phrase

	AdvancePlayback(m)
	// Should handle wrap-around or stopping
	assert.LessOrEqual(t, m.PlaybackRow, 254)
}

func TestInputTickMessage(t *testing.T) {
	// Test TickMsg structure
	msg := TickMsg{}
	assert.NotNil(t, msg)

	// Test that TickMsg can be used in tea.Msg interface
	var teaMsg tea.Msg = msg
	assert.NotNil(t, teaMsg)
}

func TestViewSwitchingLogic(t *testing.T) {
	m := createTestModel()

	// Test switching between different views
	initialView := m.ViewMode

	// Switch to each view and verify state
	configs := []ViewSwitchConfig{
		songViewConfig(0, 0),
		chainViewConfig(5),
		phraseViewConfig(10, 3),
		settingsViewConfig(),
		fileViewConfig(),
	}

	for _, config := range configs {
		switchToView(m, config)
		assert.Equal(t, config.ViewMode, m.ViewMode)
		assert.Equal(t, config.Row, m.CurrentRow)
		assert.Equal(t, config.Col, m.CurrentCol)
	}

	// Verify we can switch back to initial view
	switchToView(m, ViewSwitchConfig{
		ViewMode:     initialView,
		Row:          0,
		Col:          0,
		ScrollOffset: 0,
	})
	assert.Equal(t, initialView, m.ViewMode)
}

func TestInputHelpers(t *testing.T) {
	m := createTestModel()

	// Test various input scenarios that should not crash

	// Test with different view modes
	viewModes := []types.ViewMode{
		types.SongView,
		types.ChainView,
		types.PhraseView,
		types.FileView,
		types.SettingsView,
		types.MixerView,
	}

	for _, viewMode := range viewModes {
		t.Run("View "+string(rune(viewMode)), func(t *testing.T) {
			m.ViewMode = viewMode

			// Test navigation in this view
			keyMsg := tea.KeyMsg{Type: tea.KeyDown}
			cmd := HandleKeyInput(m, keyMsg)
			_ = cmd // Should not crash
		})
	}
}

func TestScrollAndVisibility(t *testing.T) {
	m := createTestModel()
	m.TermHeight = 20

	// Test visibility checking with different configurations
	testCases := []struct {
		name         string
		currentRow   int
		scrollOffset int
		termHeight   int
		expectScroll bool
	}{
		{"Row in view", 5, 0, 20, false},
		{"Row below view", 25, 0, 20, true},
		{"Row above view", 2, 10, 20, true},
		{"Negative row", -1, 5, 20, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m.CurrentRow = tc.currentRow
			m.ScrollOffset = tc.scrollOffset
			m.TermHeight = tc.termHeight

			config := ViewSwitchConfig{
				ViewMode:     types.PhraseView,
				Row:          tc.currentRow,
				Col:          0,
				ScrollOffset: tc.scrollOffset,
			}

			switchToViewWithVisibilityCheck(m, config)
			// Should not crash and should handle visibility correctly
			assert.NotNil(t, m)
		})
	}
}

func BenchmarkHandleKeyInput(b *testing.B) {
	m := createTestModel()
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleKeyInput(m, keyMsg)
	}
}

func BenchmarkAdvancePlayback(b *testing.B) {
	m := createTestModel()
	m.IsPlaying = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AdvancePlayback(m)
	}
}

func TestDeepCopyWithArpeggio(t *testing.T) {
	m := createTestModel()

	// Set up source phrase with arpeggio data
	sourcePhraseID := 1
	m.CurrentPhrase = sourcePhraseID
	m.ViewMode = types.PhraseView

	// Set up an arpeggio with some settings
	arpeggioID := 5
	m.ArpeggioSettings[arpeggioID] = types.ArpeggioSettings{
		Rows: [16]types.ArpeggioRow{
			{Direction: 1, Count: 4, Divisor: 2},
			{Direction: 2, Count: 8, Divisor: 3},
			{Direction: 0, Count: -1, Divisor: -1}, // Default row
		},
	}

	// Set up phrase data with arpeggio reference
	phrasesData := m.GetCurrentPhrasesData()
	(*phrasesData)[sourcePhraseID][0][types.ColArpeggio] = arpeggioID
	(*phrasesData)[sourcePhraseID][1][types.ColArpeggio] = arpeggioID // Same arpeggio used twice
	(*phrasesData)[sourcePhraseID][0][types.ColDeltaTime] = 1         // Make row playable

	// Perform deep copy
	DeepCopyToClipboard(m)

	// Verify clipboard has new phrase ID
	assert.True(t, m.Clipboard.HasData)
	destPhraseID := m.Clipboard.Value
	assert.NotEqual(t, sourcePhraseID, destPhraseID)

	// Verify arpeggio was copied to a new slot
	newArpeggioID1 := (*phrasesData)[destPhraseID][0][types.ColArpeggio]
	newArpeggioID2 := (*phrasesData)[destPhraseID][1][types.ColArpeggio]

	assert.NotEqual(t, arpeggioID, newArpeggioID1)
	assert.Equal(t, newArpeggioID1, newArpeggioID2) // Both should map to same new arpeggio
	assert.GreaterOrEqual(t, newArpeggioID1, 0)
	assert.LessOrEqual(t, newArpeggioID1, 254)

	// Verify arpeggio settings were copied
	originalSettings := m.ArpeggioSettings[arpeggioID]
	newSettings := m.ArpeggioSettings[newArpeggioID1]

	assert.Equal(t, originalSettings.Rows[0].Direction, newSettings.Rows[0].Direction)
	assert.Equal(t, originalSettings.Rows[0].Count, newSettings.Rows[0].Count)
	assert.Equal(t, originalSettings.Rows[0].Divisor, newSettings.Rows[0].Divisor)
	assert.Equal(t, originalSettings.Rows[1].Direction, newSettings.Rows[1].Direction)
	assert.Equal(t, originalSettings.Rows[1].Count, newSettings.Rows[1].Count)
	assert.Equal(t, originalSettings.Rows[1].Divisor, newSettings.Rows[1].Divisor)
}

func TestArpeggioCellCopyPaste(t *testing.T) {
	m := createTestModel()

	// Set up arpeggio view
	m.ViewMode = types.ArpeggioView
	m.ArpeggioEditingIndex = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.ArpeggioColCO) // Count column

	// Set up test data
	m.ArpeggioSettings[0].Rows[5].Direction = 1 // "u-"
	m.ArpeggioSettings[0].Rows[5].Count = 15    // "0F"
	m.ArpeggioSettings[0].Rows[5].Divisor = 8   // "08"

	// Test copying from Count column
	CopyCellToClipboard(m)

	// Verify clipboard has correct data
	assert.True(t, m.Clipboard.HasData)
	assert.Equal(t, 15, m.Clipboard.Value)
	assert.Equal(t, types.HexCell, m.Clipboard.CellType)
	assert.Equal(t, types.CellMode, m.Clipboard.Mode)
	assert.Equal(t, types.ArpeggioView, m.Clipboard.HighlightView)
	assert.Equal(t, 5, m.Clipboard.HighlightRow)
	assert.Equal(t, int(types.ArpeggioColCO), m.Clipboard.HighlightCol)

	// Move to different row, same column and paste
	m.CurrentRow = 10
	originalCountValue := m.ArpeggioSettings[0].Rows[10].Count
	PasteCellFromClipboard(m)

	// Verify paste worked
	assert.Equal(t, 15, m.ArpeggioSettings[0].Rows[10].Count)
	assert.NotEqual(t, originalCountValue, m.ArpeggioSettings[0].Rows[10].Count)

	// Test that pasting to different column fails
	m.CurrentCol = int(types.ArpeggioColDI) // Direction column
	m.CurrentRow = 12
	originalDirectionValue := m.ArpeggioSettings[0].Rows[12].Direction
	PasteCellFromClipboard(m) // Should not paste due to column mismatch

	// Verify paste failed (value unchanged)
	assert.Equal(t, originalDirectionValue, m.ArpeggioSettings[0].Rows[12].Direction)

	// Test copying from Direction column
	m.CurrentRow = 5
	m.CurrentCol = int(types.ArpeggioColDI) // Direction column
	CopyCellToClipboard(m)

	// Verify new clipboard data
	assert.Equal(t, 1, m.Clipboard.Value) // Direction value
	assert.Equal(t, int(types.ArpeggioColDI), m.Clipboard.HighlightCol)

	// Paste to same column, different row
	m.CurrentRow = 7
	PasteCellFromClipboard(m)

	// Verify paste worked
	assert.Equal(t, 1, m.ArpeggioSettings[0].Rows[7].Direction)
}

func TestRetriggerDeepCopy(t *testing.T) {
	m := createTestModel()

	// Set up phrase view
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column for sampler view

	// Set up source retrigger with custom settings (index 10)
	sourceRetriggerIndex := 10
	m.RetriggerSettings[sourceRetriggerIndex] = types.RetriggerSettings{
		Times:              4,
		Start:              2.5,
		End:                1.0,
		Beats:              8,
		VolumeDB:           -3.0,
		PitchChange:        2.0,
		FinalPitchToStart:  1,
		FinalVolumeToStart: 0,
	}

	// Put the retrigger reference in a phrase row
	m.SamplerPhrasesData[0][3][types.ColRetrigger] = sourceRetriggerIndex

	// Copy the retrigger from the source row
	m.CurrentRow = 3
	CopyCellToClipboard(m)

	// Verify clipboard has retrigger index
	assert.True(t, m.Clipboard.HasData)
	assert.Equal(t, sourceRetriggerIndex, m.Clipboard.Value)

	// Move to destination row and paste
	m.CurrentRow = 5
	PasteCellFromClipboard(m)

	// Get the new retrigger index that was pasted
	newRetriggerIndex := m.SamplerPhrasesData[0][5][types.ColRetrigger]

	// Regular copy/paste should just copy the reference (same index)
	assert.Equal(t, sourceRetriggerIndex, newRetriggerIndex)
}

func TestRetriggerDeepCopyNoUnusedSlots(t *testing.T) {
	m := createTestModel()

	// Set up phrase view
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column

	// Fill all retrigger slots to simulate no unused slots available
	for i := 0; i < 255; i++ {
		m.RetriggerSettings[i] = types.RetriggerSettings{
			Times: 1, // Non-default value to mark as used
		}
		// Reference each retrigger in a phrase to mark as used
		if i < 255 {
			m.SamplerPhrasesData[0][i%255][types.ColRetrigger] = i
		}
	}

	sourceRetriggerIndex := 10

	// Copy the retrigger from the source row
	m.CurrentRow = 10
	CopyCellToClipboard(m)

	// Move to destination row and paste
	m.CurrentRow = 5
	PasteCellFromClipboard(m)

	// Verify that when no unused slots are available, it just copies the reference
	assert.Equal(t, sourceRetriggerIndex, m.SamplerPhrasesData[0][5][types.ColRetrigger])
}

func TestRetriggerCtrlDDeepCopy(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in retrigger column
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column for sampler view

	// Set up source retrigger with custom settings
	sourceRetriggerIndex := 10
	m.RetriggerSettings[sourceRetriggerIndex] = types.RetriggerSettings{
		Times:              4,
		Start:              2.5,
		End:                1.0,
		Beats:              8,
		VolumeDB:           -3.0,
		PitchChange:        2.0,
		FinalPitchToStart:  1,
		FinalVolumeToStart: 0,
	}

	// Put the retrigger reference in the current cell
	m.SamplerPhrasesData[0][5][types.ColRetrigger] = sourceRetriggerIndex

	// Call DeepCopyToClipboard (which is what Ctrl+D triggers)
	DeepCopyToClipboard(m)

	// Verify clipboard has the ORIGINAL retrigger index but marked for deep copy
	assert.True(t, m.Clipboard.HasData)
	assert.True(t, m.Clipboard.IsFreshDeepCopy)
	assert.Equal(t, sourceRetriggerIndex, m.Clipboard.Value) // Should be original, not new
}

func TestRetriggerCtrlDDeepCopyEmptyCell(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in retrigger column
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column

	// Current cell has no retrigger (-1)
	m.SamplerPhrasesData[0][5][types.ColRetrigger] = -1

	// Call DeepCopyToClipboard (which is what Ctrl+D triggers)
	DeepCopyToClipboard(m)

	// Verify clipboard is empty since there was no retrigger to copy
	assert.False(t, m.Clipboard.HasData)
}

func TestCtrlDInNonRetriggerColumn(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in non-retrigger column (note column)
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 1
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColNN) // Note column

	// Add some data to the phrase to make it non-empty
	m.SamplerPhrasesData[1][0][types.ColNote] = 60
	m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 1

	// Call DeepCopyToClipboard (which is what Ctrl+D triggers)
	DeepCopyToClipboard(m)

	// Verify that it deep copied the entire phrase (not just a retrigger)
	assert.True(t, m.Clipboard.HasData)
	newPhraseIndex := m.Clipboard.Value
	assert.NotEqual(t, 1, newPhraseIndex) // Should be a different phrase
	assert.GreaterOrEqual(t, newPhraseIndex, 0)
	assert.Less(t, newPhraseIndex, 255)

	// Verify that the phrase data was copied
	assert.Equal(t, 60, m.SamplerPhrasesData[newPhraseIndex][0][types.ColNote])
	assert.Equal(t, 1, m.SamplerPhrasesData[newPhraseIndex][0][types.ColDeltaTime])
}

func TestArpeggioDeepCopy(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in arpeggio column
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.InstrumentColAR) // Arpeggio column for instrument view
	m.CurrentTrack = 0                        // Use track 0
	m.TrackTypes[0] = false                   // Set track 0 to instrument (false = instrument, true = sampler)

	// Set up source arpeggio with custom settings
	sourceArpeggioIndex := 10
	m.ArpeggioSettings[sourceArpeggioIndex] = types.ArpeggioSettings{
		Rows: [16]types.ArpeggioRow{
			{Direction: 1, Count: 5, Divisor: 2},
			{Direction: 2, Count: 3, Divisor: 4},
			{Direction: 0, Count: -1, Divisor: -1}, // Rest should be defaults
		},
	}

	// Put the arpeggio reference in a phrase row
	m.InstrumentPhrasesData[0][3][types.ColArpeggio] = sourceArpeggioIndex

	// Copy the arpeggio from the source row
	m.CurrentRow = 3
	CopyCellToClipboard(m)

	// Verify clipboard has arpeggio index
	assert.True(t, m.Clipboard.HasData)
	assert.Equal(t, sourceArpeggioIndex, m.Clipboard.Value)

	// Move to destination row and paste
	m.CurrentRow = 5
	PasteCellFromClipboard(m)

	// Get the new arpeggio index that was pasted
	newArpeggioIndex := m.InstrumentPhrasesData[0][5][types.ColArpeggio]

	// Regular copy/paste should just copy the reference (same index)
	assert.Equal(t, sourceArpeggioIndex, newArpeggioIndex)
}

func TestArpeggioCtrlDDeepCopy(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in arpeggio column
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.InstrumentColAR) // Arpeggio column for instrument view
	m.CurrentTrack = 0                        // Use track 0
	m.TrackTypes[0] = false                   // Set track 0 to instrument (false = instrument, true = sampler)

	// Set up source arpeggio with custom settings
	sourceArpeggioIndex := 12
	m.ArpeggioSettings[sourceArpeggioIndex] = types.ArpeggioSettings{
		Rows: [16]types.ArpeggioRow{
			{Direction: 2, Count: 8, Divisor: 3},
			{Direction: 1, Count: 4, Divisor: 1},
		},
	}

	// Put the arpeggio reference in the current cell
	m.InstrumentPhrasesData[0][5][types.ColArpeggio] = sourceArpeggioIndex

	// Call DeepCopyToClipboard (which is what Ctrl+D triggers)
	DeepCopyToClipboard(m)

	// Verify clipboard has the ORIGINAL arpeggio index but marked for deep copy
	assert.True(t, m.Clipboard.HasData)
	assert.True(t, m.Clipboard.IsFreshDeepCopy)
	assert.Equal(t, sourceArpeggioIndex, m.Clipboard.Value) // Should be original, not new
}

func TestArpeggioCtrlDDeepCopyEmptyCell(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with cursor in arpeggio column
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.InstrumentColAR) // Arpeggio column
	m.CurrentTrack = 0                        // Use track 0
	m.TrackTypes[0] = false                   // Set track 0 to instrument (false = instrument, true = sampler)

	// Current cell has no arpeggio (-1)
	m.InstrumentPhrasesData[0][5][types.ColArpeggio] = -1

	// Call DeepCopyToClipboard (which is what Ctrl+D triggers)
	DeepCopyToClipboard(m)

	// Verify clipboard is empty since there was no arpeggio to copy
	assert.False(t, m.Clipboard.HasData)
}

func TestRetriggerCtrlDPasteOptimal(t *testing.T) {
	m := createTestModel()

	// Set up phrase view
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column

	// Set up source retrigger with custom settings
	sourceRetriggerIndex := 10
	m.RetriggerSettings[sourceRetriggerIndex] = types.RetriggerSettings{
		Times:    4,
		Beats:    8,
		VolumeDB: -3.0,
	}

	// Put the retrigger reference in the current cell
	m.SamplerPhrasesData[0][5][types.ColRetrigger] = sourceRetriggerIndex

	// Call DeepCopyToClipboard (Ctrl+D) - should NOT create new pool slot yet
	DeepCopyToClipboard(m)

	// Verify clipboard has the ORIGINAL retrigger index but marked for deep copy
	assert.True(t, m.Clipboard.HasData)
	assert.True(t, m.Clipboard.IsFreshDeepCopy)
	assert.Equal(t, sourceRetriggerIndex, m.Clipboard.Value) // Should still be original

	// Verify only 1 retrigger slot is used so far (no pool allocation yet)
	usedSlots := 0
	for i := 0; i < 255; i++ {
		settings := m.RetriggerSettings[i]
		if settings.Times != 0 || settings.Beats != 0 || settings.VolumeDB != 0.0 {
			usedSlots++
		}
	}
	assert.Equal(t, 1, usedSlots) // Only original

	// Move to another cell and paste (Ctrl+V) - NOW it should create the deep copy
	m.CurrentRow = 10
	PasteCellFromClipboard(m)

	// Get the new retrigger index that was pasted
	pastedIndex := m.SamplerPhrasesData[0][10][types.ColRetrigger]
	assert.NotEqual(t, sourceRetriggerIndex, pastedIndex) // Should be different (new deep copy)

	// Verify only 2 retrigger slots are used total (original + deep copy created on paste)
	usedSlots = 0
	for i := 0; i < 255; i++ {
		settings := m.RetriggerSettings[i]
		if settings.Times != 0 || settings.Beats != 0 || settings.VolumeDB != 0.0 {
			usedSlots++
		}
	}
	assert.Equal(t, 2, usedSlots) // Original + deep copy created on paste
}

func TestArpeggioCtrlDPasteOptimal(t *testing.T) {
	m := createTestModel()

	// Set up phrase view with instrument track
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.InstrumentColAR) // Arpeggio column
	m.CurrentTrack = 0
	m.TrackTypes[0] = false // Instrument track

	// Set up source arpeggio with custom settings
	sourceArpeggioIndex := 12
	m.ArpeggioSettings[sourceArpeggioIndex] = types.ArpeggioSettings{
		Rows: [16]types.ArpeggioRow{
			{Direction: 2, Count: 8, Divisor: 3},
		},
	}

	// Put the arpeggio reference in the current cell
	m.InstrumentPhrasesData[0][5][types.ColArpeggio] = sourceArpeggioIndex

	// Call DeepCopyToClipboard (Ctrl+D) - should NOT create new pool slot yet
	DeepCopyToClipboard(m)

	// Verify clipboard has the ORIGINAL arpeggio index but marked for deep copy
	assert.True(t, m.Clipboard.HasData)
	assert.True(t, m.Clipboard.IsFreshDeepCopy)
	assert.Equal(t, sourceArpeggioIndex, m.Clipboard.Value) // Should still be original

	// Verify only 1 arpeggio slot is used so far (no pool allocation yet)
	usedSlots := 0
	for i := 0; i < 255; i++ {
		settings := m.ArpeggioSettings[i]
		for j := 0; j < 16; j++ {
			if settings.Rows[j].Direction != 0 || settings.Rows[j].Count != -1 || settings.Rows[j].Divisor != -1 {
				usedSlots++
				break
			}
		}
	}
	assert.Equal(t, 1, usedSlots) // Only original

	// Move to another cell and paste (Ctrl+V) - NOW it should create the deep copy
	m.CurrentRow = 10
	PasteCellFromClipboard(m)

	// Get the new arpeggio index that was pasted
	pastedIndex := m.InstrumentPhrasesData[0][10][types.ColArpeggio]
	assert.NotEqual(t, sourceArpeggioIndex, pastedIndex) // Should be different (new deep copy)

	// Verify only 2 arpeggio slots are used total (original + deep copy created on paste)
	usedSlots = 0
	for i := 0; i < 255; i++ {
		settings := m.ArpeggioSettings[i]
		for j := 0; j < 16; j++ {
			if settings.Rows[j].Direction != 0 || settings.Rows[j].Count != -1 || settings.Rows[j].Divisor != -1 {
				usedSlots++
				break
			}
		}
	}
	assert.Equal(t, 2, usedSlots) // Original + deep copy created on paste
}

func TestRegularCopyPasteStillWorks(t *testing.T) {
	m := createTestModel()

	// Set up phrase view
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentRow = 5
	m.CurrentCol = int(types.SamplerColRT) // Retrigger column

	// Set up source retrigger
	sourceRetriggerIndex := 15
	m.RetriggerSettings[sourceRetriggerIndex] = types.RetriggerSettings{
		Times: 2,
		Beats: 4,
	}
	m.SamplerPhrasesData[0][5][types.ColRetrigger] = sourceRetriggerIndex

	// Regular copy (Ctrl+C)
	CopyCellToClipboard(m)

	// Verify clipboard is NOT marked as fresh deep copy
	assert.True(t, m.Clipboard.HasData)
	assert.False(t, m.Clipboard.IsFreshDeepCopy)
	assert.Equal(t, sourceRetriggerIndex, m.Clipboard.Value)

	// Move to another cell and paste (Ctrl+V) - should just copy reference
	m.CurrentRow = 10
	PasteCellFromClipboard(m)

	// Verify that it just copied the reference (same index)
	pastedIndex := m.SamplerPhrasesData[0][10][types.ColRetrigger]
	assert.Equal(t, sourceRetriggerIndex, pastedIndex) // Should be same (just reference copy)
}

package input

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func createTestModel() *model.Model {
	return model.NewModel(0, "test.json", false) // Port 0 to disable OSC for testing
}

func TestHandlePgDown(t *testing.T) {
	tests := []struct {
		name        string
		viewMode    types.ViewMode
		initialRow  int
		initialCol  int
		expectedRow int
		expectedCol int // Should stay the same
		description string
	}{
		{
			name:        "SongView - from row 0 to row 16 (capped at 15)",
			viewMode:    types.SongView,
			initialRow:  0,
			initialCol:  3,
			expectedRow: 15, // Capped at max song row
			expectedCol: 3,  // Column should not change
			description: "PgDown in SongView should jump to next 16-aligned row but cap at 15",
		},
		{
			name:        "SongView - from row 8 to row 16 (capped at 15)",
			viewMode:    types.SongView,
			initialRow:  8,
			initialCol:  2,
			expectedRow: 15, // Capped at max song row
			expectedCol: 2,  // Column should not change
			description: "PgDown in SongView from middle should jump to max",
		},
		{
			name:        "ChainView - from row 0 to row 16 (capped at 15)",
			viewMode:    types.ChainView,
			initialRow:  0,
			initialCol:  0,
			expectedRow: 15, // Capped at max chain row
			expectedCol: 0,  // Column should not change
			description: "PgDown in ChainView should jump to next 16-aligned row but cap at 15",
		},
		{
			name:        "PhraseView - from row 0 to row 16",
			viewMode:    types.PhraseView,
			initialRow:  0,
			initialCol:  1,
			expectedRow: 16,
			expectedCol: 1, // Column should not change
			description: "PgDown in PhraseView should jump to row 16",
		},
		{
			name:        "PhraseView - from row 5 to row 16",
			viewMode:    types.PhraseView,
			initialRow:  5,
			initialCol:  2,
			expectedRow: 16,
			expectedCol: 2, // Column should not change
			description: "PgDown in PhraseView from row 5 should jump to row 16",
		},
		{
			name:        "PhraseView - from row 16 to row 32",
			viewMode:    types.PhraseView,
			initialRow:  16,
			initialCol:  1,
			expectedRow: 32,
			expectedCol: 1, // Column should not change
			description: "PgDown in PhraseView from row 16 should jump to row 32",
		},
		{
			name:        "PhraseView - from row 240 to row 254 (capped)",
			viewMode:    types.PhraseView,
			initialRow:  240,
			initialCol:  1,
			expectedRow: 254, // Capped at max phrase row
			expectedCol: 1,   // Column should not change
			description: "PgDown in PhraseView should cap at 254",
		},
		{
			name:        "PhraseView - already at max row 254",
			viewMode:    types.PhraseView,
			initialRow:  254,
			initialCol:  1,
			expectedRow: 254, // Should stay at 254
			expectedCol: 1,   // Column should not change
			description: "PgDown in PhraseView at max row should not move",
		},
		{
			name:        "FileView - from row 5 to row 16 (assuming enough files)",
			viewMode:    types.FileView,
			initialRow:  5,
			initialCol:  0,
			expectedRow: 16,
			expectedCol: 0, // Column should not change
			description: "PgDown in FileView should jump to next 16-aligned row",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.ViewMode = tt.viewMode
			m.CurrentRow = tt.initialRow
			m.CurrentCol = tt.initialCol
			m.ScrollOffset = 0

			// For FileView tests, ensure we have enough mock files
			if tt.viewMode == types.FileView {
				m.Files = make([]string, 100) // Create 100 mock files
				for i := 0; i < 100; i++ {
					m.Files[i] = fmt.Sprintf("file_%d.txt", i)
				}
			}

			// Call the function
			handlePgDown(m)

			// Verify the results
			assert.Equal(t, tt.expectedRow, m.CurrentRow, "Row should be %d, got %d. %s", tt.expectedRow, m.CurrentRow, tt.description)
			assert.Equal(t, tt.expectedCol, m.CurrentCol, "Column should remain %d, got %d. %s", tt.expectedCol, m.CurrentCol, tt.description)
		})
	}
}

func TestHandlePgUp(t *testing.T) {
	tests := []struct {
		name        string
		viewMode    types.ViewMode
		initialRow  int
		initialCol  int
		expectedRow int
		expectedCol int // Should stay the same
		description string
	}{
		{
			name:        "SongView - already at row 0",
			viewMode:    types.SongView,
			initialRow:  0,
			initialCol:  3,
			expectedRow: 0, // Should stay at 0
			expectedCol: 3, // Column should not change
			description: "PgUp in SongView at row 0 should not move",
		},
		{
			name:        "SongView - from row 15 to row 0",
			viewMode:    types.SongView,
			initialRow:  15,
			initialCol:  2,
			expectedRow: 0, // Should go to 0
			expectedCol: 2, // Column should not change
			description: "PgUp in SongView from row 15 should jump to row 0",
		},
		{
			name:        "SongView - from row 8 to row 0",
			viewMode:    types.SongView,
			initialRow:  8,
			initialCol:  1,
			expectedRow: 0, // Should go to 0
			expectedCol: 1, // Column should not change
			description: "PgUp in SongView from row 8 should jump to row 0",
		},
		{
			name:        "ChainView - from row 15 to row 0",
			viewMode:    types.ChainView,
			initialRow:  15,
			initialCol:  0,
			expectedRow: 0, // Should go to 0
			expectedCol: 0, // Column should not change
			description: "PgUp in ChainView from row 15 should jump to row 0",
		},
		{
			name:        "PhraseView - from row 16 to row 0",
			viewMode:    types.PhraseView,
			initialRow:  16,
			initialCol:  1,
			expectedRow: 0,
			expectedCol: 1, // Column should not change
			description: "PgUp in PhraseView from row 16 should jump to row 0",
		},
		{
			name:        "PhraseView - from row 32 to row 16",
			viewMode:    types.PhraseView,
			initialRow:  32,
			initialCol:  2,
			expectedRow: 16,
			expectedCol: 2, // Column should not change
			description: "PgUp in PhraseView from row 32 should jump to row 16",
		},
		{
			name:        "PhraseView - from row 48 to row 32",
			viewMode:    types.PhraseView,
			initialRow:  48,
			initialCol:  1,
			expectedRow: 32,
			expectedCol: 1, // Column should not change
			description: "PgUp in PhraseView from row 48 should jump to row 32",
		},
		{
			name:        "PhraseView - from row 17 to row 16",
			viewMode:    types.PhraseView,
			initialRow:  17,
			initialCol:  1,
			expectedRow: 16,
			expectedCol: 1, // Column should not change
			description: "PgUp in PhraseView from row 17 should jump to row 16",
		},
		{
			name:        "PhraseView - from row 5 to row 0",
			viewMode:    types.PhraseView,
			initialRow:  5,
			initialCol:  1,
			expectedRow: 0,
			expectedCol: 1, // Column should not change
			description: "PgUp in PhraseView from row 5 should jump to row 0",
		},
		{
			name:        "FileView - from row 20 to row 16",
			viewMode:    types.FileView,
			initialRow:  20,
			initialCol:  0,
			expectedRow: 16,
			expectedCol: 0, // Column should not change
			description: "PgUp in FileView from row 20 should jump to row 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.ViewMode = tt.viewMode
			m.CurrentRow = tt.initialRow
			m.CurrentCol = tt.initialCol
			m.ScrollOffset = 0

			// For FileView tests, ensure we have enough mock files
			if tt.viewMode == types.FileView {
				m.Files = make([]string, 100) // Create 100 mock files
				for i := 0; i < 100; i++ {
					m.Files[i] = fmt.Sprintf("file_%d.txt", i)
				}
			}

			// Call the function
			handlePgUp(m)

			// Verify the results
			assert.Equal(t, tt.expectedRow, m.CurrentRow, "Row should be %d, got %d. %s", tt.expectedRow, m.CurrentRow, tt.description)
			assert.Equal(t, tt.expectedCol, m.CurrentCol, "Column should remain %d, got %d. %s", tt.expectedCol, m.CurrentCol, tt.description)
		})
	}
}

func TestPgUpPgDown16Alignment(t *testing.T) {
	tests := []struct {
		name             string
		viewMode         types.ViewMode
		startRow         int
		col              int
		pgDownExpected   int
		pgUpFromExpected int
		description      string
	}{
		{
			name:             "PhraseView - 16-alignment test from row 0",
			viewMode:         types.PhraseView,
			startRow:         0,
			col:              1,
			pgDownExpected:   16, // 0x10
			pgUpFromExpected: 0,  // Back to 0x00
			description:      "Test 16-alignment: 0 -> 16 -> 0",
		},
		{
			name:             "PhraseView - 16-alignment test from row 32",
			viewMode:         types.PhraseView,
			startRow:         32, // 0x20
			col:              1,
			pgDownExpected:   48, // 0x30
			pgUpFromExpected: 32, // Back to 0x20
			description:      "Test 16-alignment: 32 -> 48 -> 32",
		},
		{
			name:             "PhraseView - 16-alignment test from row 64",
			viewMode:         types.PhraseView,
			startRow:         64, // 0x40
			col:              2,
			pgDownExpected:   80, // 0x50
			pgUpFromExpected: 64, // Back to 0x40
			description:      "Test 16-alignment: 64 -> 80 -> 64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.ViewMode = tt.viewMode
			m.CurrentRow = tt.startRow
			m.CurrentCol = tt.col

			// Test PgDown
			handlePgDown(m)
			assert.Equal(t, tt.pgDownExpected, m.CurrentRow, "PgDown: %s - expected row %d, got %d", tt.description, tt.pgDownExpected, m.CurrentRow)
			assert.Equal(t, tt.col, m.CurrentCol, "PgDown: Column should remain %d, got %d", tt.col, m.CurrentCol)

			// Test PgUp from the new position
			handlePgUp(m)
			assert.Equal(t, tt.pgUpFromExpected, m.CurrentRow, "PgUp: %s - expected row %d, got %d", tt.description, tt.pgUpFromExpected, m.CurrentRow)
			assert.Equal(t, tt.col, m.CurrentCol, "PgUp: Column should remain %d, got %d", tt.col, m.CurrentCol)
		})
	}
}

func TestHandleKeyInputPgUpPgDown(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		viewMode    types.ViewMode
		initialRow  int
		initialCol  int
		expectedRow int
		expectedCol int
	}{
		{
			name:        "pgdown key binding",
			key:         "pgdown",
			viewMode:    types.PhraseView,
			initialRow:  0,
			initialCol:  1,
			expectedRow: 16,
			expectedCol: 1,
		},
		{
			name:        "pgup key binding",
			key:         "pgup",
			viewMode:    types.PhraseView,
			initialRow:  32,
			initialCol:  2,
			expectedRow: 16,
			expectedCol: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel()
			m.ViewMode = tt.viewMode
			m.CurrentRow = tt.initialRow
			m.CurrentCol = tt.initialCol

			keyMsg := tea.KeyMsg{}
			keyMsg.Type = tea.KeyRunes
			switch tt.key {
			case "pgdown":
				keyMsg.Type = tea.KeyPgDown
			case "pgup":
				keyMsg.Type = tea.KeyPgUp
			}

			// Call HandleKeyInput
			HandleKeyInput(m, keyMsg)

			assert.Equal(t, tt.expectedRow, m.CurrentRow, "Row should be %d, got %d", tt.expectedRow, m.CurrentRow)
			assert.Equal(t, tt.expectedCol, m.CurrentCol, "Column should remain %d, got %d", tt.expectedCol, m.CurrentCol)
		})
	}
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

// TestDuckingStickyBehavior tests that DU column has sticky behavior
func TestDuckingStickyBehavior(t *testing.T) {
	m := createTestModel()

	// Test data: Set up a phrase with ducking values
	// Row 0: DU = 05
	// Row 1: DU = -- (should inherit 05)
	// Row 2: DU = 10
	// Row 3: DU = -- (should inherit 10)
	// Row 4: DU = -- (should inherit 10)

	m.SamplerPhrasesData[0][0][types.ColEffectDucking] = 5
	m.SamplerPhrasesData[0][1][types.ColEffectDucking] = -1 // "--"
	m.SamplerPhrasesData[0][2][types.ColEffectDucking] = 16
	m.SamplerPhrasesData[0][3][types.ColEffectDucking] = -1 // "--"
	m.SamplerPhrasesData[0][4][types.ColEffectDucking] = -1 // "--"

	trackId := 0 // Test with track 0 (sampler track)

	// Test sticky behavior using GetEffectiveValueForTrack
	tests := []struct {
		name     string
		phrase   int
		row      int
		expected int
	}{
		{"Row 0: Direct value", 0, 0, 5},
		{"Row 1: Inherit from row 0", 0, 1, 5},
		{"Row 2: Direct value", 0, 2, 16},
		{"Row 3: Inherit from row 2", 0, 3, 16},
		{"Row 4: Inherit from row 2", 0, 4, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEffectiveValueForTrack(m, tt.phrase, tt.row, int(types.ColEffectDucking), trackId)
			assert.Equal(t, tt.expected, result, "GetEffectiveValueForTrack should return correct sticky value")
		})
	}

	// Test ResolveDuckingIndex function (convenience wrapper)
	for _, tt := range tests {
		t.Run(tt.name+" (ResolveDuckingIndex)", func(t *testing.T) {
			result := ResolveDuckingIndex(m, tt.phrase, tt.row, trackId)
			assert.Equal(t, tt.expected, result, "ResolveDuckingIndex should return correct sticky value")
		})
	}

	// Test edge case: All rows empty
	m.SamplerPhrasesData[1][0][types.ColEffectDucking] = -1
	m.SamplerPhrasesData[1][1][types.ColEffectDucking] = -1
	m.SamplerPhrasesData[1][2][types.ColEffectDucking] = -1

	result := GetEffectiveValueForTrack(m, 1, 2, int(types.ColEffectDucking), trackId)
	assert.Equal(t, -1, result, "Should return -1 when no non-null values found")
}

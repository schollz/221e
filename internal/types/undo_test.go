package types

import (
	"testing"
)

func TestUndoState(t *testing.T) {
	// Test that UndoState can be created and has the expected fields
	undoState := UndoState{
		ViewMode:        SongView,
		CurrentRow:      5,
		CurrentCol:      2,
		CurrentPhrase:   10,
		CurrentChain:    3,
		CurrentTrack:    1,
		ScrollOffset:    0,
		ChangedDataType: "phrase",
		Description:     "Test undo state",
	}

	if undoState.ViewMode != SongView {
		t.Errorf("Expected ViewMode to be SongView, got %v", undoState.ViewMode)
	}

	if undoState.CurrentRow != 5 {
		t.Errorf("Expected CurrentRow to be 5, got %d", undoState.CurrentRow)
	}

	if undoState.ChangedDataType != "phrase" {
		t.Errorf("Expected ChangedDataType to be 'phrase', got %s", undoState.ChangedDataType)
	}

	if undoState.Description != "Test undo state" {
		t.Errorf("Expected Description to be 'Test undo state', got %s", undoState.Description)
	}
}

func TestUndoStateDataSnapshots(t *testing.T) {
	// Test that UndoState can hold data snapshots
	undoState := UndoState{}

	// Test song data
	songData := [8][16]int{}
	songData[0][0] = 5
	undoState.SongData = &songData

	if undoState.SongData == nil {
		t.Error("Expected SongData to be set")
	}

	if (*undoState.SongData)[0][0] != 5 {
		t.Errorf("Expected SongData[0][0] to be 5, got %d", (*undoState.SongData)[0][0])
	}

	// Test phrase data
	phraseData := [255][][]int{}
	phraseData[0] = make([][]int, 255)
	for i := range phraseData[0] {
		phraseData[0][i] = make([]int, int(ColCount))
		phraseData[0][i][ColNote] = 60 // Set middle C
	}
	undoState.InstrumentPhrasesData = &phraseData

	if undoState.InstrumentPhrasesData == nil {
		t.Error("Expected InstrumentPhrasesData to be set")
	}

	if (*undoState.InstrumentPhrasesData)[0][0][ColNote] != 60 {
		t.Errorf("Expected phrase note to be 60, got %d", (*undoState.InstrumentPhrasesData)[0][0][ColNote])
	}
}
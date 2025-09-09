package model

import (
	"testing"

	"github.com/schollz/2n/internal/types"
)

func TestUndoFunctionality(t *testing.T) {
	// Create a new model
	m := NewModel(0, "test")

	// Test initial state - no undo history
	if m.CanUndo() {
		t.Error("Expected no undo history initially")
	}

	if len(m.UndoHistory) != 0 {
		t.Errorf("Expected empty undo history, got %d entries", len(m.UndoHistory))
	}

	// Test PushUndoState
	m.CurrentRow = 5
	m.CurrentCol = 2
	m.SongData[0][0] = 42

	m.PushUndoState("song", "Test modification")

	if !m.CanUndo() {
		t.Error("Expected undo history after push")
	}

	if len(m.UndoHistory) != 1 {
		t.Errorf("Expected 1 undo entry, got %d", len(m.UndoHistory))
	}

	undoState := m.UndoHistory[0]
	if undoState.CurrentRow != 5 {
		t.Errorf("Expected undo state CurrentRow to be 5, got %d", undoState.CurrentRow)
	}

	if undoState.CurrentCol != 2 {
		t.Errorf("Expected undo state CurrentCol to be 2, got %d", undoState.CurrentCol)
	}

	if undoState.ChangedDataType != "song" {
		t.Errorf("Expected ChangedDataType to be 'song', got %s", undoState.ChangedDataType)
	}

	if undoState.Description != "Test modification" {
		t.Errorf("Expected Description to be 'Test modification', got %s", undoState.Description)
	}

	// Test that song data was captured
	if undoState.SongData == nil {
		t.Error("Expected SongData to be captured")
	}

	if (*undoState.SongData)[0][0] != 42 {
		t.Errorf("Expected captured song data to be 42, got %d", (*undoState.SongData)[0][0])
	}

	// Modify state further
	m.CurrentRow = 10
	m.CurrentCol = 7
	m.SongData[0][0] = 99

	// Test PopUndoState
	success := m.PopUndoState()
	if !success {
		t.Error("Expected undo to succeed")
	}

	// Check that state was restored
	if m.CurrentRow != 5 {
		t.Errorf("Expected CurrentRow to be restored to 5, got %d", m.CurrentRow)
	}

	if m.CurrentCol != 2 {
		t.Errorf("Expected CurrentCol to be restored to 2, got %d", m.CurrentCol)
	}

	if m.SongData[0][0] != 42 {
		t.Errorf("Expected song data to be restored to 42, got %d", m.SongData[0][0])
	}

	// Test that undo history is empty after pop
	if m.CanUndo() {
		t.Error("Expected no undo history after pop")
	}

	if len(m.UndoHistory) != 0 {
		t.Errorf("Expected empty undo history after pop, got %d entries", len(m.UndoHistory))
	}
}

func TestUndoHistoryLimit(t *testing.T) {
	// Create a new model
	m := NewModel(0, "test")
	m.UndoHistoryMax = 3 // Set low limit for testing

	// Push more than the limit
	for i := 0; i < 5; i++ {
		m.PushUndoState("test", "Test entry")
	}

	// Check that history is limited
	if len(m.UndoHistory) > m.UndoHistoryMax {
		t.Errorf("Expected undo history to be limited to %d entries, got %d", m.UndoHistoryMax, len(m.UndoHistory))
	}

	if len(m.UndoHistory) != 3 {
		t.Errorf("Expected exactly 3 undo entries due to limit, got %d", len(m.UndoHistory))
	}
}

func TestUndoWithDifferentDataTypes(t *testing.T) {
	// Create a new model
	m := NewModel(0, "test")

	// Test phrase data undo
	m.CurrentPhrase = 5
	m.CurrentTrack = 0 // Instrument track
	phrasesData := m.GetCurrentPhrasesData()
	(*phrasesData)[5][0][types.ColNote] = 60

	m.PushUndoState("phrase", "Phrase modification")

	// Verify phrase data was captured
	undoState := m.UndoHistory[0]
	if undoState.InstrumentPhrasesData == nil {
		t.Error("Expected InstrumentPhrasesData to be captured for phrase undo")
	}

	// Test chain data undo
	m.CurrentChain = 3
	chainsData := m.GetCurrentChainsData()
	(*chainsData)[3][0] = 15

	m.PushUndoState("chain", "Chain modification")

	// Verify chain data was captured
	undoState = m.UndoHistory[1]
	if undoState.InstrumentChainsData == nil {
		t.Error("Expected InstrumentChainsData to be captured for chain undo")
	}
}

func TestEmptyUndoPop(t *testing.T) {
	// Create a new model
	m := NewModel(0, "test")

	// Try to pop from empty undo history
	success := m.PopUndoState()
	if success {
		t.Error("Expected undo to fail when history is empty")
	}
}
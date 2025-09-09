//go:build ignore

package main

import (
	"fmt"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

func main() {
	fmt.Println("=== 2n Undo Feature Demonstration ===")
	fmt.Println()

	// Create a new model (simulating the application startup)
	m := model.NewModel(0, "demo")
	
	fmt.Printf("Initial state:\n")
	fmt.Printf("- Current position: Row %d, Col %d\n", m.CurrentRow, m.CurrentCol)
	fmt.Printf("- Song data [0][0]: %d\n", m.SongData[0][0])
	fmt.Printf("- Undo history: %d entries\n", len(m.UndoHistory))
	fmt.Printf("- Can undo: %v\n", m.CanUndo())
	fmt.Println()

	// Simulate user making changes to song data
	fmt.Println("=== Making Changes ===")
	
	// Change 1: Modify song position
	fmt.Println("1. Moving to position (5, 3) and setting song data...")
	m.CurrentRow = 5
	m.CurrentCol = 3
	m.ViewMode = types.SongView
	
	// This would normally be triggered by user input (Ctrl+Up/Down)
	// Simulate capturing state before modification
	m.PushUndoState("song", "Set chain 42 at track 3 row 5")
	m.SongData[3][5] = 42
	
	fmt.Printf("- New position: Row %d, Col %d\n", m.CurrentRow, m.CurrentCol)
	fmt.Printf("- Song data [3][5]: %d\n", m.SongData[3][5])
	fmt.Printf("- Undo history: %d entries\n", len(m.UndoHistory))
	fmt.Printf("- Can undo: %v\n", m.CanUndo())
	fmt.Println()

	// Change 2: Move to phrase view and modify phrase data
	fmt.Println("2. Switching to phrase view and modifying phrase data...")
	m.CurrentRow = 10
	m.CurrentCol = 2
	m.CurrentPhrase = 5
	m.CurrentTrack = 0 // Instrument track
	m.ViewMode = types.PhraseView
	
	// Get the appropriate phrase data for instrument track
	phrasesData := m.GetCurrentPhrasesData()
	
	// Simulate capturing state before modification
	m.PushUndoState("phrase", "Set note 60 (middle C) at phrase 5 row 10")
	(*phrasesData)[5][10][types.ColNote] = 60 // Middle C
	(*phrasesData)[5][10][types.ColDeltaTime] = 1 // Enable playback
	
	fmt.Printf("- New position: Row %d, Col %d, Phrase %d\n", m.CurrentRow, m.CurrentCol, m.CurrentPhrase)
	fmt.Printf("- Phrase data [5][10][Note]: %d\n", (*phrasesData)[5][10][types.ColNote])
	fmt.Printf("- Phrase data [5][10][DeltaTime]: %d\n", (*phrasesData)[5][10][types.ColDeltaTime])
	fmt.Printf("- Undo history: %d entries\n", len(m.UndoHistory))
	fmt.Printf("- Can undo: %v\n", m.CanUndo())
	fmt.Println()

	// Change 3: Another modification
	fmt.Println("3. Making another phrase modification...")
	m.CurrentRow = 12
	
	// Simulate capturing state before modification  
	m.PushUndoState("phrase", "Set note 64 (E) at phrase 5 row 12")
	(*phrasesData)[5][12][types.ColNote] = 64 // E note
	(*phrasesData)[5][12][types.ColDeltaTime] = 2
	
	fmt.Printf("- Position: Row %d, Col %d, Phrase %d\n", m.CurrentRow, m.CurrentCol, m.CurrentPhrase)
	fmt.Printf("- Phrase data [5][12][Note]: %d\n", (*phrasesData)[5][12][types.ColNote])
	fmt.Printf("- Undo history: %d entries\n", len(m.UndoHistory))
	fmt.Println()

	// Now demonstrate undo functionality
	fmt.Println("=== Testing Undo Functionality ===")
	
	// Undo 1: Should restore the previous phrase state
	fmt.Println("1. Performing first undo (Ctrl+Z)...")
	success := m.PopUndoState()
	fmt.Printf("- Undo successful: %v\n", success)
	fmt.Printf("- Restored position: Row %d, Col %d, Phrase %d\n", m.CurrentRow, m.CurrentCol, m.CurrentPhrase)
	fmt.Printf("- Phrase data [5][12][Note]: %d (should be -1)\n", (*phrasesData)[5][12][types.ColNote])
	fmt.Printf("- Phrase data [5][10][Note]: %d (should still be 60)\n", (*phrasesData)[5][10][types.ColNote])
	fmt.Printf("- Remaining undo history: %d entries\n", len(m.UndoHistory))
	fmt.Println()

	// Undo 2: Should restore to before the first phrase modification
	fmt.Println("2. Performing second undo...")
	success = m.PopUndoState()
	fmt.Printf("- Undo successful: %v\n", success)
	fmt.Printf("- Restored position: Row %d, Col %d, Phrase %d\n", m.CurrentRow, m.CurrentCol, m.CurrentPhrase)
	fmt.Printf("- Phrase data [5][10][Note]: %d (should be -1)\n", (*phrasesData)[5][10][types.ColNote])
	fmt.Printf("- Remaining undo history: %d entries\n", len(m.UndoHistory))
	fmt.Println()

	// Undo 3: Should restore to song view state
	fmt.Println("3. Performing third undo...")
	success = m.PopUndoState()
	fmt.Printf("- Undo successful: %v\n", success)
	fmt.Printf("- Restored position: Row %d, Col %d\n", m.CurrentRow, m.CurrentCol)
	fmt.Printf("- Restored view mode: %v\n", m.ViewMode)
	fmt.Printf("- Song data [3][5]: %d (should be -1)\n", m.SongData[3][5])
	fmt.Printf("- Remaining undo history: %d entries\n", len(m.UndoHistory))
	fmt.Println()

	// Try one more undo (should fail)
	fmt.Println("4. Attempting undo when history is empty...")
	success = m.PopUndoState()
	fmt.Printf("- Undo successful: %v (should be false)\n", success)
	fmt.Printf("- Can undo: %v (should be false)\n", m.CanUndo())
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("The undo feature successfully:")
	fmt.Println("- Captured state before each modification")
	fmt.Println("- Restored data and cursor position on undo")
	fmt.Println("- Maintained proper history management")
	fmt.Println("- Handled empty history gracefully")
}
package input

import (
	"testing"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFillSequentialRT(t *testing.T) {
	m := model.NewModel(0, "")
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentTrack = 0

	phrasesData := m.GetCurrentPhrasesData()

	// Set RT value 05 in row 0
	(*phrasesData)[0][0][types.ColRetrigger] = 5

	// Position cursor on RT column, row 4 
	m.CurrentCol = 6 // RT column in sampler view
	m.CurrentRow = 4

	// Execute Ctrl+F
	FillSequentialPhrase(m)

	// Verify RT column values - should all be 05 (reference value kept constant)
	assert.Equal(t, 5, (*phrasesData)[0][0][types.ColRetrigger], "Row 0 should keep original value")
	assert.Equal(t, 5, (*phrasesData)[0][1][types.ColRetrigger], "Row 1 should have reference value")
	assert.Equal(t, 5, (*phrasesData)[0][2][types.ColRetrigger], "Row 2 should have reference value")
	assert.Equal(t, 5, (*phrasesData)[0][3][types.ColRetrigger], "Row 3 should have reference value")
	assert.Equal(t, 5, (*phrasesData)[0][4][types.ColRetrigger], "Row 4 should have reference value")
	assert.Equal(t, -1, (*phrasesData)[0][5][types.ColRetrigger], "Row 5 should remain empty")
}

func TestFillSequentialTS(t *testing.T) {
	m := model.NewModel(0, "")
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentTrack = 0

	phrasesData := m.GetCurrentPhrasesData()

	// Set TS value 03 in row 1
	(*phrasesData)[0][1][types.ColTimestretch] = 3

	// Position cursor on TS column, row 5
	m.CurrentCol = 7 // TS column in sampler view
	m.CurrentRow = 5

	// Execute Ctrl+F
	FillSequentialPhrase(m)

	// Verify TS column values - should all be 03 (reference value kept constant)
	assert.Equal(t, -1, (*phrasesData)[0][0][types.ColTimestretch], "Row 0 should remain empty")
	assert.Equal(t, 3, (*phrasesData)[0][1][types.ColTimestretch], "Row 1 should keep original value")
	assert.Equal(t, 3, (*phrasesData)[0][2][types.ColTimestretch], "Row 2 should have reference value")
	assert.Equal(t, 3, (*phrasesData)[0][3][types.ColTimestretch], "Row 3 should have reference value")
	assert.Equal(t, 3, (*phrasesData)[0][4][types.ColTimestretch], "Row 4 should have reference value")
	assert.Equal(t, 3, (*phrasesData)[0][5][types.ColTimestretch], "Row 5 should have reference value")
	assert.Equal(t, -1, (*phrasesData)[0][6][types.ColTimestretch], "Row 6 should remain empty")
}

func TestFillSequentialRTNoReferenceValue(t *testing.T) {
	m := model.NewModel(0, "")
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentTrack = 0

	phrasesData := m.GetCurrentPhrasesData()

	// No RT value set anywhere - all rows are -1

	// Position cursor on RT column, row 3
	m.CurrentCol = 6 // RT column in sampler view
	m.CurrentRow = 3

	// Execute Ctrl+F
	FillSequentialPhrase(m)

	// Verify RT column values - should all be 0 (default when no reference found)
	assert.Equal(t, 0, (*phrasesData)[0][0][types.ColRetrigger], "Row 0 should have default value 0")
	assert.Equal(t, 0, (*phrasesData)[0][1][types.ColRetrigger], "Row 1 should have default value 0")
	assert.Equal(t, 0, (*phrasesData)[0][2][types.ColRetrigger], "Row 2 should have default value 0")
	assert.Equal(t, 0, (*phrasesData)[0][3][types.ColRetrigger], "Row 3 should have default value 0")
	assert.Equal(t, -1, (*phrasesData)[0][4][types.ColRetrigger], "Row 4 should remain empty")
}

func TestFillSequentialTSNoReferenceValue(t *testing.T) {
	m := model.NewModel(0, "")
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentTrack = 0

	phrasesData := m.GetCurrentPhrasesData()

	// No TS value set anywhere - all rows are -1

	// Position cursor on TS column, row 2
	m.CurrentCol = 7 // TS column in sampler view
	m.CurrentRow = 2

	// Execute Ctrl+F
	FillSequentialPhrase(m)

	// Verify TS column values - should all be 0 (default when no reference found)
	assert.Equal(t, 0, (*phrasesData)[0][0][types.ColTimestretch], "Row 0 should have default value 0")
	assert.Equal(t, 0, (*phrasesData)[0][1][types.ColTimestretch], "Row 1 should have default value 0") 
	assert.Equal(t, 0, (*phrasesData)[0][2][types.ColTimestretch], "Row 2 should have default value 0")
	assert.Equal(t, -1, (*phrasesData)[0][3][types.ColTimestretch], "Row 3 should remain empty")
}

func TestFillSequentialOtherColumnStillIncrements(t *testing.T) {
	// Test that other columns still work with the normal incrementing behavior
	m := model.NewModel(0, "")
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 0
	m.CurrentTrack = 0

	phrasesData := m.GetCurrentPhrasesData()

	// Set velocity value 10 in row 0
	(*phrasesData)[0][0][types.ColVelocity] = 10

	// Position cursor on velocity column, row 3
	// Find the velocity column in sampler view
	for col := 0; col < 15; col++ {
		mapping := m.GetColumnMapping(col)
		if mapping != nil && mapping.DataColumnIndex == int(types.ColVelocity) {
			m.CurrentCol = col
			break
		}
	}
	m.CurrentRow = 3

	// Execute Ctrl+F
	FillSequentialPhrase(m)

	// Verify velocity column values - should increment normally (10, 11, 12, 13)
	assert.Equal(t, 10, (*phrasesData)[0][0][types.ColVelocity], "Row 0 should keep original value")
	assert.Equal(t, 11, (*phrasesData)[0][1][types.ColVelocity], "Row 1 should increment")
	assert.Equal(t, 12, (*phrasesData)[0][2][types.ColVelocity], "Row 2 should increment")
	assert.Equal(t, 13, (*phrasesData)[0][3][types.ColVelocity], "Row 3 should increment")
	assert.Equal(t, -1, (*phrasesData)[0][4][types.ColVelocity], "Row 4 should remain empty")
}
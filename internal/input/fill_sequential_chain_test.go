package input

import (
	"testing"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFillSequentialChain(t *testing.T) {
	t.Run("Chain with existing phrases at rows 0,1 - cursor at row 5", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.ViewMode = types.ChainView
		m.CurrentChain = 0
		m.CurrentRow = 5

		// Set existing phrases
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[0][0] = 0 // phrase 00 at row 0
		(*chainsData)[0][1] = 1 // phrase 01 at row 1
		// rows 2-5 should be -1 (empty)

		// Execute Ctrl+F
		FillSequentialChain(m)

		// Verify the results
		assert.Equal(t, 0, (*chainsData)[0][0], "Row 0 should keep existing phrase 00")
		assert.Equal(t, 1, (*chainsData)[0][1], "Row 1 should keep existing phrase 01")
		assert.Equal(t, 2, (*chainsData)[0][2], "Row 2 should be filled with phrase 02")
		assert.Equal(t, 3, (*chainsData)[0][3], "Row 3 should be filled with phrase 03")
		assert.Equal(t, 4, (*chainsData)[0][4], "Row 4 should be filled with phrase 04")
		assert.Equal(t, 5, (*chainsData)[0][5], "Row 5 should be filled with phrase 05")
		assert.Equal(t, -1, (*chainsData)[0][6], "Row 6 should remain empty")
	})

	t.Run("Empty chain - cursor at row 3", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.ViewMode = types.ChainView
		m.CurrentChain = 0
		m.CurrentRow = 3

		// All rows should be -1 (empty) by default
		chainsData := m.GetCurrentChainsData()

		// Execute Ctrl+F
		FillSequentialChain(m)

		// Verify the results
		assert.Equal(t, 0, (*chainsData)[0][0], "Row 0 should be filled with phrase 00")
		assert.Equal(t, 1, (*chainsData)[0][1], "Row 1 should be filled with phrase 01")
		assert.Equal(t, 2, (*chainsData)[0][2], "Row 2 should be filled with phrase 02")
		assert.Equal(t, 3, (*chainsData)[0][3], "Row 3 should be filled with phrase 03")
		assert.Equal(t, -1, (*chainsData)[0][4], "Row 4 should remain empty")
	})

	t.Run("Chain with phrase at row 0 - cursor at row 2", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.ViewMode = types.ChainView
		m.CurrentChain = 0
		m.CurrentRow = 2

		// Set existing phrase
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[0][0] = 5 // phrase 05 at row 0

		// Execute Ctrl+F
		FillSequentialChain(m)

		// Verify the results
		assert.Equal(t, 5, (*chainsData)[0][0], "Row 0 should keep existing phrase 05")
		assert.Equal(t, 6, (*chainsData)[0][1], "Row 1 should be filled with phrase 06")
		assert.Equal(t, 7, (*chainsData)[0][2], "Row 2 should be filled with phrase 07")
		assert.Equal(t, -1, (*chainsData)[0][3], "Row 3 should remain empty")
	})

	t.Run("Chain with existing phrase at row 2 - cursor at row 4", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.ViewMode = types.ChainView
		m.CurrentChain = 0
		m.CurrentRow = 4

		// Set existing phrase
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[0][2] = 10 // phrase 10 at row 2

		// Execute Ctrl+F
		FillSequentialChain(m)

		// Verify the results - should fill from last non-empty value + 1
		assert.Equal(t, -1, (*chainsData)[0][0], "Row 0 should remain empty")
		assert.Equal(t, -1, (*chainsData)[0][1], "Row 1 should remain empty")
		assert.Equal(t, 10, (*chainsData)[0][2], "Row 2 should keep existing phrase 10")
		assert.Equal(t, 11, (*chainsData)[0][3], "Row 3 should be filled with phrase 11")
		assert.Equal(t, 12, (*chainsData)[0][4], "Row 4 should be filled with phrase 12")
		assert.Equal(t, -1, (*chainsData)[0][5], "Row 5 should remain empty")
	})

	t.Run("User scenario: phrases 00,01 at rows 0,1 - cursor at row 5", func(t *testing.T) {
		// This matches the exact scenario from the user's debug output
		m := model.NewModel(0, "")
		m.ViewMode = types.ChainView
		m.CurrentChain = 0
		m.CurrentRow = 5

		// Set existing phrases as shown in user's chain view
		chainsData := m.GetCurrentChainsData()
		(*chainsData)[0][0] = 0 // phrase 00 at row 0
		(*chainsData)[0][1] = 1 // phrase 01 at row 1
		// rows 2-5 should be -1 (shown as "--" in UI)

		// Execute Ctrl+F (what user pressed)
		FillSequentialChain(m)

		// Verify expected behavior: fill rows 2-5 with phrases 02,03,04,05
		assert.Equal(t, 0, (*chainsData)[0][0], "Row 0 should keep existing phrase 00")
		assert.Equal(t, 1, (*chainsData)[0][1], "Row 1 should keep existing phrase 01")
		assert.Equal(t, 2, (*chainsData)[0][2], "Row 2 should be filled with phrase 02")
		assert.Equal(t, 3, (*chainsData)[0][3], "Row 3 should be filled with phrase 03")
		assert.Equal(t, 4, (*chainsData)[0][4], "Row 4 should be filled with phrase 04")
		assert.Equal(t, 5, (*chainsData)[0][5], "Row 5 should be filled with phrase 05")
		assert.Equal(t, -1, (*chainsData)[0][6], "Row 6 should remain empty")
		assert.Equal(t, -1, (*chainsData)[0][7], "Row 7 should remain empty")
	})
}

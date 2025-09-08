package ticks

import (
	"testing"

	"github.com/schollz/2n/internal/types"
)

func TestCalculatePhraseTicks(t *testing.T) {
	// Create test phrases data
	var phrasesData [255][][]int
	
	// Initialize phrase 0 with some test data (255 rows, ColCount columns)
	phrasesData[0] = make([][]int, 255)
	for i := 0; i < 255; i++ {
		phrasesData[0][i] = make([]int, types.ColCount)
		for j := 0; j < int(types.ColCount); j++ {
			phrasesData[0][i][j] = -1 // Initialize all to -1 (no data)
		}
	}
	
	// Add some test ticks to phrase 0
	phrasesData[0][0][types.ColDeltaTime] = 8  // 8 ticks
	phrasesData[0][1][types.ColDeltaTime] = 4  // 4 ticks
	phrasesData[0][2][types.ColDeltaTime] = 2  // 2 ticks
	phrasesData[0][3][types.ColDeltaTime] = -1 // no ticks (should be ignored)
	phrasesData[0][4][types.ColDeltaTime] = 0  // 0 ticks (should be ignored)
	
	totalTicks := CalculatePhraseTicks(&phrasesData, 0)
	expected := 8 + 4 + 2 // = 14
	
	if totalTicks != expected {
		t.Errorf("CalculatePhraseTicks() = %d, expected %d", totalTicks, expected)
	}
	
	// Test empty phrase
	emptyTicks := CalculatePhraseTicks(&phrasesData, 1) // phrase 1 is empty
	if emptyTicks != 0 {
		t.Errorf("CalculatePhraseTicks() for empty phrase = %d, expected 0", emptyTicks)
	}
	
	// Test invalid phrase ID
	invalidTicks := CalculatePhraseTicks(&phrasesData, -1)
	if invalidTicks != 0 {
		t.Errorf("CalculatePhraseTicks() for invalid phrase = %d, expected 0", invalidTicks)
	}
}

func TestCalculateChainTicks(t *testing.T) {
	// Create test phrases data
	var phrasesData [255][][]int
	
	// Initialize phrase 0 with test data
	phrasesData[0] = make([][]int, 255)
	for i := 0; i < 255; i++ {
		phrasesData[0][i] = make([]int, types.ColCount)
		for j := 0; j < int(types.ColCount); j++ {
			phrasesData[0][i][j] = -1
		}
	}
	phrasesData[0][0][types.ColDeltaTime] = 8
	phrasesData[0][1][types.ColDeltaTime] = 4
	// phrase 0 total: 12 ticks
	
	// Initialize phrase 1 with test data  
	phrasesData[1] = make([][]int, 255)
	for i := 0; i < 255; i++ {
		phrasesData[1][i] = make([]int, types.ColCount)
		for j := 0; j < int(types.ColCount); j++ {
			phrasesData[1][i][j] = -1
		}
	}
	phrasesData[1][0][types.ColDeltaTime] = 2
	phrasesData[1][1][types.ColDeltaTime] = 2
	phrasesData[1][2][types.ColDeltaTime] = 2
	// phrase 1 total: 6 ticks
	
	// Create test chain data
	var chainsData [][]int
	chainsData = make([][]int, 1) // 1 chain
	chainsData[0] = make([]int, 16) // 16 rows
	for i := 0; i < 16; i++ {
		chainsData[0][i] = -1 // Initialize to empty
	}
	
	// Chain 0 contains: phrase 0, phrase 1, phrase 0 again, empty
	chainsData[0][0] = 0  // phrase 0 (12 ticks)
	chainsData[0][1] = 1  // phrase 1 (6 ticks)
	chainsData[0][2] = 0  // phrase 0 again (12 ticks)
	chainsData[0][3] = -1 // empty (0 ticks)
	
	totalTicks := CalculateChainTicks(&chainsData, &phrasesData, 0)
	expected := 12 + 6 + 12 // = 30
	
	if totalTicks != expected {
		t.Errorf("CalculateChainTicks() = %d, expected %d", totalTicks, expected)
	}
	
	// Test invalid chain ID
	invalidTicks := CalculateChainTicks(&chainsData, &phrasesData, -1)
	if invalidTicks != 0 {
		t.Errorf("CalculateChainTicks() for invalid chain = %d, expected 0", invalidTicks)
	}
}

func TestCalculateTrackTicks(t *testing.T) {
	// Create test phrases data
	var phrasesData [255][][]int
	
	// Initialize phrase 0
	phrasesData[0] = make([][]int, 255)
	for i := 0; i < 255; i++ {
		phrasesData[0][i] = make([]int, types.ColCount)
		for j := 0; j < int(types.ColCount); j++ {
			phrasesData[0][i][j] = -1
		}
	}
	phrasesData[0][0][types.ColDeltaTime] = 8
	// phrase 0 total: 8 ticks
	
	// Create test chain data
	var chainsData [][]int
	chainsData = make([][]int, 2) // 2 chains
	
	// Chain 0: contains phrase 0 twice
	chainsData[0] = make([]int, 16)
	for i := 0; i < 16; i++ {
		chainsData[0][i] = -1
	}
	chainsData[0][0] = 0 // phrase 0 (8 ticks)
	chainsData[0][1] = 0 // phrase 0 again (8 ticks)
	// chain 0 total: 16 ticks
	
	// Chain 1: contains phrase 0 once
	chainsData[1] = make([]int, 16)
	for i := 0; i < 16; i++ {
		chainsData[1][i] = -1
	}
	chainsData[1][0] = 0 // phrase 0 (8 ticks)
	// chain 1 total: 8 ticks
	
	// Create test song data
	var songData [8][16]int
	for i := 0; i < 8; i++ {
		for j := 0; j < 16; j++ {
			songData[i][j] = -1
		}
	}
	
	// Track 0: contains chain 0 and chain 1
	songData[0][0] = 0  // chain 0 (16 ticks)
	songData[0][1] = 1  // chain 1 (8 ticks)
	songData[0][2] = -1 // empty
	
	totalTicks := CalculateTrackTicks(&songData, &chainsData, &phrasesData, 0)
	expected := 16 + 8 // = 24
	
	if totalTicks != expected {
		t.Errorf("CalculateTrackTicks() = %d, expected %d", totalTicks, expected)
	}
	
	// Test invalid track ID
	invalidTicks := CalculateTrackTicks(&songData, &chainsData, &phrasesData, -1)
	if invalidTicks != 0 {
		t.Errorf("CalculateTrackTicks() for invalid track = %d, expected 0", invalidTicks)
	}
}
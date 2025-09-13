package ticks

import (
	"github.com/schollz/collidertracker/internal/types"
)

// CalculatePhraseTicks calculates the total ticks in a phrase by summing all DT values
func CalculatePhraseTicks(phrasesData *[255][][]int, phraseID int) int {
	if phraseID < 0 || phraseID >= 255 || phrasesData == nil {
		return 0
	}

	totalTicks := 0
	for row := 0; row < len((*phrasesData)[phraseID]); row++ {
		if row >= len((*phrasesData)[phraseID]) {
			break
		}
		dtValue := (*phrasesData)[phraseID][row][types.ColDeltaTime]
		if dtValue > 0 {
			totalTicks += dtValue
		}
	}
	return totalTicks
}

// CalculateChainTicks calculates the total ticks in a chain by summing all phrase ticks
func CalculateChainTicks(chainsData *[][]int, phrasesData *[255][][]int, chainID int) int {
	if chainID < 0 || chainsData == nil || chainID >= len(*chainsData) || phrasesData == nil {
		return 0
	}

	totalTicks := 0
	for row := 0; row < len((*chainsData)[chainID]); row++ {
		phraseID := (*chainsData)[chainID][row]
		if phraseID != -1 {
			totalTicks += CalculatePhraseTicks(phrasesData, phraseID)
		}
	}
	return totalTicks
}

// CalculateTrackTicks calculates the total ticks in a track by summing all chain ticks
func CalculateTrackTicks(songData *[8][16]int, chainsData *[][]int, phrasesData *[255][][]int, trackID int) int {
	if trackID < 0 || trackID >= 8 || songData == nil || chainsData == nil || phrasesData == nil {
		return 0
	}

	totalTicks := 0
	for row := 0; row < 16; row++ {
		chainID := songData[trackID][row]
		if chainID != -1 {
			totalTicks += CalculateChainTicks(chainsData, phrasesData, chainID)
		}
	}
	return totalTicks
}

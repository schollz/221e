package input

import (
	"fmt"
	"log"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/storage"
)

func ModifyRetriggerValue(m *model.Model, baseDelta float32) {
	if m.RetriggerEditingIndex < 0 || m.RetriggerEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := m.RetriggerSettings[m.RetriggerEditingIndex]

	if m.CurrentRow == 0 { // Times
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newTimes := settings.Times + delta
		if newTimes < 0 {
			newTimes = 0
		} else if newTimes > 256 {
			newTimes = 256
		}
		settings.Times = newTimes
		log.Printf("Modified retrigger %02X Times: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Times-delta, settings.Times, delta)
	} else if m.CurrentRow == 1 { // Starting Rate
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newStart := settings.Start + delta
		if newStart < 0 {
			newStart = 0
		} else if newStart > 256 {
			newStart = 256
		}
		settings.Start = newStart
		log.Printf("Modified retrigger %02X Starting Rate: %.2f -> %.2f (delta: %.2f)", m.RetriggerEditingIndex, settings.Start-delta, settings.Start, delta)
	} else if m.CurrentRow == 2 { // Final Rate
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newEnd := settings.End + delta
		if newEnd < 0 {
			newEnd = 0
		} else if newEnd > 256 {
			newEnd = 256
		}
		settings.End = newEnd
		log.Printf("Modified retrigger %02X Final Rate: %.2f -> %.2f (delta: %.2f)", m.RetriggerEditingIndex, settings.End-delta, settings.End, delta)
	} else if m.CurrentRow == 3 { // Beats
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newBeats := settings.Beats + delta
		if newBeats < 0 {
			newBeats = 0
		} else if newBeats > 256 {
			newBeats = 256
		}
		settings.Beats = newBeats
		log.Printf("Modified retrigger %02X Beats: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Beats-delta, settings.Beats, delta)
	} else if m.CurrentRow == 4 { // Volume dB
		newVolumeDB := settings.VolumeDB + baseDelta
		if newVolumeDB < -16.0 {
			newVolumeDB = -16.0
		} else if newVolumeDB > 16.0 {
			newVolumeDB = 16.0
		}
		settings.VolumeDB = newVolumeDB
		log.Printf("Modified retrigger %02X VolumeDB: %.1f -> %.1f (delta: %.1f)", m.RetriggerEditingIndex, settings.VolumeDB-baseDelta, settings.VolumeDB, baseDelta)
	} else if m.CurrentRow == 5 { // Pitch change
		newPitchChange := settings.PitchChange + baseDelta
		if newPitchChange < -24.0 {
			newPitchChange = -24.0
		} else if newPitchChange > 24.0 {
			newPitchChange = 24.0
		}
		settings.PitchChange = newPitchChange
		log.Printf("Modified retrigger %02X PitchChange: %.1f -> %.1f (delta: %.1f)", m.RetriggerEditingIndex, settings.PitchChange-baseDelta, settings.PitchChange, baseDelta)
	} else if m.CurrentRow == 6 { // Final pitch to start
		// Toggle between 0 (No) and 1 (Yes)
		if baseDelta > 0 {
			settings.FinalPitchToStart = 1
		} else {
			settings.FinalPitchToStart = 0
		}
		finalPitchValue := "No"
		if settings.FinalPitchToStart == 1 {
			finalPitchValue = "Yes"
		}
		log.Printf("Modified retrigger %02X FinalPitchToStart: %s", m.RetriggerEditingIndex, finalPitchValue)
	} else if m.CurrentRow == 7 { // Final volume to start
		// Toggle between 0 (No) and 1 (Yes)
		if baseDelta > 0 {
			settings.FinalVolumeToStart = 1
		} else {
			settings.FinalVolumeToStart = 0
		}
		finalVolumeValue := "No"
		if settings.FinalVolumeToStart == 1 {
			finalVolumeValue = "Yes"
		}
		log.Printf("Modified retrigger %02X FinalVolumeToStart: %s", m.RetriggerEditingIndex, finalVolumeValue)
	} else if m.CurrentRow == 8 { // Every
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newEvery := settings.Every + delta
		if newEvery < 1 {
			newEvery = 1
		} else if newEvery > 64 {
			newEvery = 64
		}
		settings.Every = newEvery
		log.Printf("Modified retrigger %02X Every: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Every-delta, settings.Every, delta)
	} else if m.CurrentRow == 9 { // Probability
		// Use different increments: 10 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 10 // Coarse control (Ctrl+Up/Down): +/-10%
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1%
		} else {
			delta = int(baseDelta) // Fallback
		}

		newProbability := settings.Probability + delta
		if newProbability < 0 {
			newProbability = 0
		} else if newProbability > 100 {
			newProbability = 100
		}
		settings.Probability = newProbability
		log.Printf("Modified retrigger %02X Probability: %d -> %d (delta: %d)", m.RetriggerEditingIndex, settings.Probability-delta, settings.Probability, delta)
	}

	// Store back the modified settings
	m.RetriggerSettings[m.RetriggerEditingIndex] = settings
	storage.AutoSave(m)
}

func ModifyTimestrechValue(m *model.Model, baseDelta float32) {
	if m.TimestrechEditingIndex < 0 || m.TimestrechEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := m.TimestrechSettings[m.TimestrechEditingIndex]

	if m.CurrentRow == 0 { // Start
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newStart := settings.Start + delta
		if newStart < 0 {
			newStart = 0
		} else if newStart > 256 {
			newStart = 256
		}
		settings.Start = newStart
		log.Printf("Modified timestretch %02X Start: %.2f -> %.2f (delta: %.2f)", m.TimestrechEditingIndex, settings.Start-delta, settings.Start, delta)
	} else if m.CurrentRow == 1 { // End
		// Use different increments: 0.05 for fine, 1.0 for coarse (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta float32
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = baseDelta // Coarse control (Ctrl+Up/Down)
		} else {
			delta = baseDelta // Fine control (Ctrl+Left/Right)
		}

		newEnd := settings.End + delta
		if newEnd < 0 {
			newEnd = 0
		} else if newEnd > 256 {
			newEnd = 256
		}
		settings.End = newEnd
		log.Printf("Modified timestretch %02X End: %.2f -> %.2f (delta: %.2f)", m.TimestrechEditingIndex, settings.End-delta, settings.End, delta)
	} else if m.CurrentRow == 2 { // Beats
		newBeats := settings.Beats + int(baseDelta)
		if newBeats < 0 {
			newBeats = 0
		} else if newBeats > 256 {
			newBeats = 256
		}
		settings.Beats = newBeats
		log.Printf("Modified timestretch %02X Beats: %d -> %d (delta: %.2f)", m.TimestrechEditingIndex, settings.Beats-int(baseDelta), settings.Beats, baseDelta)
	} else if m.CurrentRow == 3 { // Every
		// Use different increments: 4 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 4 // Coarse control (Ctrl+Up/Down): +/-4
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newEvery := settings.Every + delta
		if newEvery < 1 {
			newEvery = 1
		} else if newEvery > 64 {
			newEvery = 64
		}
		settings.Every = newEvery
		log.Printf("Modified timestretch %02X Every: %d -> %d (delta: %d)", m.TimestrechEditingIndex, settings.Every-delta, settings.Every, delta)
	} else if m.CurrentRow == 4 { // Probability
		// Use different increments: 10 for coarse, 1 for fine (based on Ctrl+Up/Down vs Ctrl+Left/Right)
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 10 // Coarse control (Ctrl+Up/Down): +/-10%
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control (Ctrl+Left/Right): +/-1%
		} else {
			delta = int(baseDelta) // Fallback
		}

		newProbability := settings.Probability + delta
		if newProbability < 0 {
			newProbability = 0
		} else if newProbability > 100 {
			newProbability = 100
		}
		settings.Probability = newProbability
		log.Printf("Modified timestretch %02X Probability: %d -> %d (delta: %d)", m.TimestrechEditingIndex, settings.Probability-delta, settings.Probability, delta)
	}

	// Store back the modified settings
	m.TimestrechSettings[m.TimestrechEditingIndex] = settings
	storage.AutoSave(m)
}

func ModifyModulateValue(m *model.Model, baseDelta float32) {
	if m.ModulateEditingIndex < 0 || m.ModulateEditingIndex >= 255 {
		return
	}

	// Get current settings
	settings := m.ModulateSettings[m.ModulateEditingIndex]

	if m.CurrentRow == 0 { // Seed
		// Use different increments: 5 for coarse, 1 for fine
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 5 // Coarse control: +/-5
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control: +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newSeed := settings.Seed + delta
		if newSeed < -1 {
			newSeed = -1 // "none"
		} else if newSeed > 128 {
			newSeed = 128
		}
		oldSeedValue := "none"
		if settings.Seed >= 0 {
			oldSeedValue = fmt.Sprintf("%d", settings.Seed)
		}
		settings.Seed = newSeed
		newSeedValue := "none"
		if settings.Seed >= 0 {
			newSeedValue = fmt.Sprintf("%d", settings.Seed)
		}
		log.Printf("Modified modulate %02X Seed: %s -> %s", m.ModulateEditingIndex, oldSeedValue, newSeedValue)
	} else if m.CurrentRow == 1 { // IRandom
		// Use different increments: 5 for coarse, 1 for fine
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 5 // Coarse control: +/-5
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control: +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newIRandom := settings.IRandom + delta
		if newIRandom < 0 {
			newIRandom = 0
		} else if newIRandom > 128 {
			newIRandom = 128
		}
		settings.IRandom = newIRandom
		log.Printf("Modified modulate %02X IRandom: %d -> %d (delta: %d)", m.ModulateEditingIndex, settings.IRandom-delta, settings.IRandom, delta)
	} else if m.CurrentRow == 2 { // Sub
		// Use different increments: 5 for coarse, 1 for fine
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 5 // Coarse control: +/-5
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control: +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newSub := settings.Sub + delta
		if newSub < 0 {
			newSub = 0
		} else if newSub > 120 {
			newSub = 120
		}
		settings.Sub = newSub
		log.Printf("Modified modulate %02X Sub: %d -> %d (delta: %d)", m.ModulateEditingIndex, settings.Sub-delta, settings.Sub, delta)
	} else if m.CurrentRow == 3 { // Add
		// Use different increments: 5 for coarse, 1 for fine
		var delta int
		if baseDelta == 1.0 || baseDelta == -1.0 {
			delta = int(baseDelta) * 5 // Coarse control: +/-5
		} else if baseDelta == 0.05 || baseDelta == -0.05 {
			delta = int(baseDelta / 0.05) // Fine control: +/-1
		} else {
			delta = int(baseDelta) // Fallback
		}

		newAdd := settings.Add + delta
		if newAdd < 0 {
			newAdd = 0
		} else if newAdd > 120 {
			newAdd = 120
		}
		settings.Add = newAdd
		log.Printf("Modified modulate %02X Add: %d -> %d (delta: %d)", m.ModulateEditingIndex, settings.Add-delta, settings.Add, delta)
	} else if m.CurrentRow == 4 { // ScaleRoot
		// Cycle through note names (C, C#, D, D#, E, F, F#, G, G#, A, A#, B)
		noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
		currentIndex := settings.ScaleRoot
		
		// Ensure current index is within valid range
		if currentIndex < 0 || currentIndex >= len(noteNames) {
			currentIndex = 0
		}
		
		// Move to next/previous note
		var newIndex int
		if baseDelta > 0 {
			newIndex = (currentIndex + 1) % len(noteNames)
		} else {
			newIndex = (currentIndex - 1 + len(noteNames)) % len(noteNames)
		}
		
		oldNote := noteNames[currentIndex]
		settings.ScaleRoot = newIndex
		log.Printf("Modified modulate %02X ScaleRoot: %s -> %s", m.ModulateEditingIndex, oldNote, noteNames[newIndex])
	} else if m.CurrentRow == 5 { // Scale
		// Cycle through available scales
		availableScales := []string{"all", "major", "minor", "dorian", "mixolydian", "pentatonic", "blues", "chromatic"}
		currentIndex := -1
		
		// Find current scale index
		for i, scale := range availableScales {
			if settings.Scale == scale {
				currentIndex = i
				break
			}
		}
		
		// If current scale not found, default to "all"
		if currentIndex == -1 {
			currentIndex = 0
		}
		
		// Move to next/previous scale
		var newIndex int
		if baseDelta > 0 {
			newIndex = (currentIndex + 1) % len(availableScales)
		} else {
			newIndex = (currentIndex - 1 + len(availableScales)) % len(availableScales)
		}
		
		oldScale := settings.Scale
		settings.Scale = availableScales[newIndex]
		log.Printf("Modified modulate %02X Scale: %s -> %s", m.ModulateEditingIndex, oldScale, settings.Scale)
	}

	// Save the modified settings back to the model
	m.ModulateSettings[m.ModulateEditingIndex] = settings
	storage.AutoSave(m)
}

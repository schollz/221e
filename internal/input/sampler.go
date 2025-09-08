package input

import (
	"log"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
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
	}

	// Store back the modified settings
	m.TimestrechSettings[m.TimestrechEditingIndex] = settings
	storage.AutoSave(m)
}

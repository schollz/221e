package input

import (
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
)

func ModifySettingsValue(m *model.Model, delta float32) {
	if m.CurrentCol == 0 {
		// Global column settings
		switch m.CurrentRow {
		case 0: // BPM
			modifier := createFloatModifier(
				func() float32 { return m.BPM },
				func(v float32) { m.BPM = v },
				1, 999, "BPM",
			)
			modifyValueWithBounds(modifier, delta)

		case 1: // PPQ
			modifier := createIntModifier(
				func() int { return m.PPQ },
				func(v int) { m.PPQ = v },
				1, 32, "PPQ",
			)
			modifyValueWithBounds(modifier, delta)

		case 2: // PregainDB
			modifier := createFloatModifier(
				func() float32 { return m.PregainDB },
				func(v float32) {
					m.PregainDB = v
					m.SendOSCPregainMessage() // Send OSC message for pregain change
				},
				-96, 32, "PregainDB",
			)
			modifyValueWithBounds(modifier, delta)

		case 3: // PostgainDB
			modifier := createFloatModifier(
				func() float32 { return m.PostgainDB },
				func(v float32) {
					m.PostgainDB = v
					m.SendOSCPostgainMessage() // Send OSC message for postgain change
				},
				-96, 32, "PostgainDB",
			)
			modifyValueWithBounds(modifier, delta)

		case 4: // BiasDB
			modifier := createFloatModifier(
				func() float32 { return m.BiasDB },
				func(v float32) {
					m.BiasDB = v
					m.SendOSCBiasMessage() // Send OSC message for bias change
				},
				-96, 32, "BiasDB",
			)
			modifyValueWithBounds(modifier, delta)

		case 5: // SaturationDB
			modifier := createFloatModifier(
				func() float32 { return m.SaturationDB },
				func(v float32) {
					m.SaturationDB = v
					m.SendOSCSaturationMessage() // Send OSC message for saturation change
				},
				-96, 32, "SaturationDB",
			)
			modifyValueWithBounds(modifier, delta)

		case 6: // DriveDB
			modifier := createFloatModifier(
				func() float32 { return m.DriveDB },
				func(v float32) {
					m.DriveDB = v
					m.SendOSCDriveMessage() // Send OSC message for drive change
				},
				-96, 32, "DriveDB",
			)
			modifyValueWithBounds(modifier, delta)
		}
	} else if m.CurrentCol == 1 {
		// Input column settings
		switch m.CurrentRow {
		case 0: // InputLevelDB
			modifier := createFloatModifier(
				func() float32 { return m.InputLevelDB },
				func(v float32) {
					m.InputLevelDB = v
					m.SendOSCInputLevelMessage() // Send OSC message for input level change
				},
				-48, 24, "InputLevelDB",
			)
			modifyValueWithBounds(modifier, delta)

		case 1: // ReverbSendPercent
			modifier := createFloatModifier(
				func() float32 { return m.ReverbSendPercent },
				func(v float32) {
					m.ReverbSendPercent = v
					m.SendOSCReverbSendMessage() // Send OSC message for reverb send change
				},
				0, 100, "ReverbSendPercent",
			)
			modifyValueWithBounds(modifier, delta)
		}
	}
	storage.AutoSave(m)
}

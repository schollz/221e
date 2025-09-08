package input

import (
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/types"
)

func ModifySettingsValue(m *model.Model, delta float32) {
	if m.CurrentCol == 0 {
		// Global column settings
		switch types.GlobalSettingsRow(m.CurrentRow) {
		case types.GlobalSettingsRowBPM: // BPM
			modifier := createFloatModifier(
				func() float32 { return m.BPM },
				func(v float32) { m.BPM = v },
				1, 999, "BPM",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowPPQ: // PPQ
			modifier := createIntModifier(
				func() int { return m.PPQ },
				func(v int) { m.PPQ = v },
				1, 32, "PPQ",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowPregainDB: // PregainDB
			modifier := createFloatModifier(
				func() float32 { return m.PregainDB },
				func(v float32) {
					m.PregainDB = v
					m.SendOSCPregainMessage() // Send OSC message for pregain change
				},
				-96, 32, "PregainDB",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowPostgainDB: // PostgainDB
			modifier := createFloatModifier(
				func() float32 { return m.PostgainDB },
				func(v float32) {
					m.PostgainDB = v
					m.SendOSCPostgainMessage() // Send OSC message for postgain change
				},
				-96, 32, "PostgainDB",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowBiasDB: // BiasDB
			modifier := createFloatModifier(
				func() float32 { return m.BiasDB },
				func(v float32) {
					m.BiasDB = v
					m.SendOSCBiasMessage() // Send OSC message for bias change
				},
				-96, 32, "BiasDB",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowSaturationDB: // SaturationDB
			modifier := createFloatModifier(
				func() float32 { return m.SaturationDB },
				func(v float32) {
					m.SaturationDB = v
					m.SendOSCSaturationMessage() // Send OSC message for saturation change
				},
				-96, 32, "SaturationDB",
			)
			modifyValueWithBounds(modifier, delta)

		case types.GlobalSettingsRowDriveDB: // DriveDB
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
		switch types.InputSettingsRow(m.CurrentRow) {
		case types.InputSettingsRowInputLevelDB: // InputLevelDB
			modifier := createFloatModifier(
				func() float32 { return m.InputLevelDB },
				func(v float32) {
					m.InputLevelDB = v
					m.SendOSCInputLevelMessage() // Send OSC message for input level change
				},
				-48, 24, "InputLevelDB",
			)
			modifyValueWithBounds(modifier, delta)

		case types.InputSettingsRowReverbSendPercent: // ReverbSendPercent
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

package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
)

func GetMidiStatusMessage(m *model.Model) string {
	settings := m.MidiSettings[m.MidiEditingIndex]

	var columnStatus string
	switch m.CurrentRow {
	case 0: // MIDI Device row
		columnStatus = fmt.Sprintf("MIDI Device: %s", settings.Device)
	case 1: // MIDI Channel row
		columnStatus = fmt.Sprintf("MIDI Channel: %s", settings.Channel)
	}

	baseMsg := "Up/Down: Navigate settings | Left/Right: Navigate columns | Ctrl+Arrow: Adjust values | Shift+Left: Back to Phrase view"
	return fmt.Sprintf("%s | %s", columnStatus, baseMsg)
}

func RenderMidiView(m *model.Model) string {
	statusMsg := GetMidiStatusMessage(m)
	return renderViewWithCommonPattern(m, "MIDI Settings", fmt.Sprintf("MIDI %02X", m.MidiEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current MIDI settings
		settings := m.MidiSettings[m.MidiEditingIndex]

		// Settings rows with common rendering pattern
		settingsRows := []struct {
			label string
			value string
			row   int
		}{
			{"Device:", settings.Device, 0},
			{"Channel:", settings.Channel, 1},
		}

		for _, setting := range settingsRows {
			var valueCell string
			if m.CurrentRow == setting.row {
				valueCell = styles.Selected.Render(setting.value)
			} else {
				valueCell = styles.Normal.Render(setting.value)
			}
			row := fmt.Sprintf("  %-8s %s", styles.Label.Render(setting.label), valueCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		return content.String()
	}, statusMsg, 4) // 2 settings rows + spacing + header
}
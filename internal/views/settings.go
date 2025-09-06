package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
)

func RenderSettingsView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "Preferences ", "", func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Settings rows with common rendering pattern
		settings := []struct {
			label string
			value string
			row   int
		}{
			{"BPM:", fmt.Sprintf("%.2f", m.BPM), 0},
			{"PPQ:", fmt.Sprintf("%d", m.PPQ), 1},
			{"Pre:", fmt.Sprintf("%.1f dB", m.PregainDB), 2},
			{"Post:", fmt.Sprintf("%.1f dB", m.PostgainDB), 3},
			{"Bias:", fmt.Sprintf("%.1f dB", m.BiasDB), 4},
			{"Sat:", fmt.Sprintf("%.1f dB", m.SaturationDB), 5},
			{"Drive:", fmt.Sprintf("%.1f dB", m.DriveDB), 6},
			{"Input:", fmt.Sprintf("%.1f dB", m.InputLevelDB), 7},
			{"Reverb:", fmt.Sprintf("%.1f%%", m.ReverbSendPercent), 8},
		}

		for _, setting := range settings {
			var valueCell string
			if m.CurrentRow == setting.row {
				valueCell = styles.Selected.Render(setting.value)
			} else {
				valueCell = styles.Normal.Render(setting.value)
			}
			row := fmt.Sprintf("  %-4s %s", styles.Label.Render(setting.label), valueCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		content.WriteString("\n")

		// Timing info
		beatsPerSecond := float64(m.BPM) / 60.0
		ticksPerSecond := beatsPerSecond * float64(m.PPQ)
		secondsPerTick := 1.0 / ticksPerSecond
		timingInfo := fmt.Sprintf("Timing: %.3f seconds per row", secondsPerTick)
		content.WriteString(styles.Normal.Render(timingInfo))
		content.WriteString("\n\n")

		return content.String()
	}, "Up/Down: Navigate | Ctrl+Arrow: Adjust values | Shift+Down: Back to Chain view", 12)
}

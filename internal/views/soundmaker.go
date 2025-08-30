package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
)

func GetSoundMakerStatusMessage(m *model.Model) string {
	settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

	var columnStatus string
	switch m.CurrentRow {
	case 0: // SoundMaker Name row
		columnStatus = fmt.Sprintf("SoundMaker: %s", settings.Name)
	case 1: // Parameter A row
		if settings.A == -1 {
			columnStatus = "Parameter A: --"
		} else {
			columnStatus = fmt.Sprintf("Parameter A: %02X", settings.A)
		}
	case 2: // Parameter B row
		if settings.B == -1 {
			columnStatus = "Parameter B: --"
		} else {
			columnStatus = fmt.Sprintf("Parameter B: %02X", settings.B)
		}
	case 3: // Parameter C row
		if settings.C == -1 {
			columnStatus = "Parameter C: --"
		} else {
			columnStatus = fmt.Sprintf("Parameter C: %02X", settings.C)
		}
	case 4: // Parameter D row
		if settings.D == -1 {
			columnStatus = "Parameter D: --"
		} else {
			columnStatus = fmt.Sprintf("Parameter D: %02X", settings.D)
		}
	default:
		// SoundMaker selection rows
		availableSoundMakers := []string{"Polyperc", "Infinite Pad"}
		if m.CurrentRow >= 5 && m.CurrentRow-5+m.ScrollOffset < len(availableSoundMakers) {
			soundMakerIndex := m.CurrentRow - 5 + m.ScrollOffset
			columnStatus = fmt.Sprintf("Select SoundMaker: %s", availableSoundMakers[soundMakerIndex])
		} else {
			columnStatus = "Available SoundMakers"
		}
	}

	baseMsg := "Up/Down: Navigate | SPACE: Select SoundMaker | Ctrl+Arrow: Adjust values | Shift+Left: Back to Phrase view"
	return fmt.Sprintf("%s | %s", columnStatus, baseMsg)
}

func RenderSoundMakerView(m *model.Model) string {
	statusMsg := GetSoundMakerStatusMessage(m)
	return renderViewWithCommonPattern(m, "SoundMaker Settings", fmt.Sprintf("SoundMaker %02X", m.SoundMakerEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current SoundMaker settings
		settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

		// Settings rows with common rendering pattern
		settingsRows := []struct {
			label string
			value string
			row   int
		}{
			{"Name:", settings.Name, 0},
			{"A:", func() string {
				if settings.A == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.A)
			}(), 1},
			{"B:", func() string {
				if settings.B == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.B)
			}(), 2},
			{"C:", func() string {
				if settings.C == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.C)
			}(), 3},
			{"D:", func() string {
				if settings.D == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.D)
			}(), 4},
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

		// Separator line
		content.WriteString("\n")
		content.WriteString(styles.Label.Render("Available SoundMakers:"))
		content.WriteString("\n\n")

		// Available SoundMakers list (scrollable)
		availableSoundMakers := []string{"Polyperc", "Infinite Pad"}
		visibleRows := m.GetVisibleRows() - 9 // Reserve space for header, settings, and labels
		soundMakerStartRow := 5               // SoundMakers start at row 5 (after Name and A,B,C,D settings)

		for i := 0; i < visibleRows && i+m.ScrollOffset < len(availableSoundMakers); i++ {
			dataIndex := i + m.ScrollOffset
			soundMakerRow := soundMakerStartRow + i

			// Arrow for current selection
			arrow := " "
			if m.CurrentRow == soundMakerRow {
				arrow = "▶"
			}

			// SoundMaker name with appropriate styling
			soundMakerName := availableSoundMakers[dataIndex]
			var soundMakerCell string
			if m.CurrentRow == soundMakerRow {
				soundMakerCell = styles.Selected.Render(soundMakerName)
			} else {
				soundMakerCell = styles.Normal.Render(soundMakerName)
			}

			row := fmt.Sprintf("%s %s", arrow, soundMakerCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		return content.String()
	}, statusMsg, 12) // Use dynamic visible rows
}

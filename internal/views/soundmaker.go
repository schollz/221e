package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/supercollider"
)

func GetSoundMakerStatusMessage(m *model.Model) string {
	settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

	var columnStatus string
	switch m.CurrentRow {
	case 0: // SoundMaker Name row
		columnStatus = fmt.Sprintf("SoundMaker: %s", settings.Name)
	case 1: // Parameter row (A or Preset depending on SoundMaker)
		if settings.Name == "DX7" {
			if settings.Preset == -1 {
				columnStatus = "Preset: --"
			} else {
				patchName, err := supercollider.GetDX7PatchName(settings.Preset)
				if err == nil {
					columnStatus = fmt.Sprintf("Preset: %s (%d)", patchName, settings.Preset)
				} else {
					columnStatus = fmt.Sprintf("Preset: %d", settings.Preset)
				}
			}
		} else {
			if settings.A == -1 {
				columnStatus = "Parameter A: --"
			} else {
				columnStatus = fmt.Sprintf("Parameter A: %02X", settings.A)
			}
		}
	case 2: // Parameter B row (only for non-DX7)
		if settings.Name != "DX7" {
			if settings.B == -1 {
				columnStatus = "Parameter B: --"
			} else {
				columnStatus = fmt.Sprintf("Parameter B: %02X", settings.B)
			}
		}
	case 3: // Parameter C row (only for non-DX7)
		if settings.Name != "DX7" {
			if settings.C == -1 {
				columnStatus = "Parameter C: --"
			} else {
				columnStatus = fmt.Sprintf("Parameter C: %02X", settings.C)
			}
		}
	case 4: // Parameter D row (only for non-DX7)
		if settings.Name != "DX7" {
			if settings.D == -1 {
				columnStatus = "Parameter D: --"
			} else {
				columnStatus = fmt.Sprintf("Parameter D: %02X", settings.D)
			}
		}
	default:
		// SoundMaker selection rows
		availableSoundMakers := []string{"Polyperc", "Infinite Pad", "DX7"}
		var soundMakerStartRow int
		if settings.Name == "DX7" {
			soundMakerStartRow = 2 // SoundMakers start at row 2 for DX7
		} else {
			soundMakerStartRow = 5 // SoundMakers start at row 5 for others
		}
		
		if m.CurrentRow >= soundMakerStartRow && m.CurrentRow-soundMakerStartRow+m.ScrollOffset < len(availableSoundMakers) {
			soundMakerIndex := m.CurrentRow - soundMakerStartRow + m.ScrollOffset
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
		var settingsRows []struct {
			label string
			value string
			row   int
		}

		// Always show Name row
		settingsRows = append(settingsRows, struct {
			label string
			value string
			row   int
		}{"Name:", settings.Name, 0})

		// Show different parameters based on SoundMaker type
		if settings.Name == "DX7" {
			// For DX7, show Preset instead of A, B, C, D
			settingsRows = append(settingsRows, struct {
				label string
				value string
				row   int
			}{"Preset:", func() string {
				if settings.Preset == -1 {
					return "--"
				}
				patchName, err := supercollider.GetDX7PatchName(settings.Preset)
				if err == nil {
					return fmt.Sprintf("%s (%d)", patchName, settings.Preset)
				}
				return fmt.Sprintf("%d", settings.Preset)
			}(), 1})
		} else {
			// For other SoundMakers, show A, B, C, D
			settingsRows = append(settingsRows, struct {
				label string
				value string
				row   int
			}{"A:", func() string {
				if settings.A == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.A)
			}(), 1})
			settingsRows = append(settingsRows, struct {
				label string
				value string
				row   int
			}{"B:", func() string {
				if settings.B == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.B)
			}(), 2})
			settingsRows = append(settingsRows, struct {
				label string
				value string
				row   int
			}{"C:", func() string {
				if settings.C == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.C)
			}(), 3})
			settingsRows = append(settingsRows, struct {
				label string
				value string
				row   int
			}{"D:", func() string {
				if settings.D == -1 {
					return "--"
				}
				return fmt.Sprintf("%02X", settings.D)
			}(), 4})
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
		availableSoundMakers := []string{"Polyperc", "Infinite Pad", "DX7"}
		visibleRows := m.GetVisibleRows() - 9 // Reserve space for header, settings, and labels
		var soundMakerStartRow int
		if settings.Name == "DX7" {
			soundMakerStartRow = 2 // SoundMakers start at row 2 for DX7 (after Name and Preset)
		} else {
			soundMakerStartRow = 5 // SoundMakers start at row 5 (after Name and A,B,C,D settings)
		}

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

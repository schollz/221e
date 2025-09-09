package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/supercollider"
	"github.com/schollz/2n/internal/types"
)

func GetSoundMakerStatusMessage(m *model.Model) string {
	settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

	var columnStatus string

	// Check if we're on the name row
	if m.CurrentRow == 0 {
		columnStatus = fmt.Sprintf("SoundMaker: %s", settings.Name)
	} else {
		// Check if we're on a parameter row
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			col0, col1 := def.GetParametersSortedByColumn()
			allParams := append(col0, col1...)

			// Parameter rows start at row 1
			paramIndex := m.CurrentRow - 1
			if paramIndex >= 0 && paramIndex < len(allParams) {
				param := allParams[paramIndex]
				value := settings.GetParameterValue(param.Key)

				// Special handling for DX7 preset display
				if param.Key == "preset" && settings.Name == "DX7" {
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						patchName, err := supercollider.GetDX7PatchName(value)
						if err == nil {
							columnStatus = fmt.Sprintf("%s: %s (%d)", param.DisplayName, patchName, value)
						} else {
							columnStatus = fmt.Sprintf("%s: %d", param.DisplayName, value)
						}
					}
				} else {
					// Standard parameter display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						if param.Type == types.ParameterTypeHex {
							columnStatus = fmt.Sprintf("%s: %02X", param.DisplayName, value)
						} else {
							columnStatus = fmt.Sprintf("%s: %d", param.DisplayName, value)
						}
					}
				}
			} else {
				// SoundMaker selection rows
				availableSoundMakers := []string{"Polyperc", "Infinite Pad", "DX7"}
				soundMakerStartRow := 1 + len(allParams)

				if m.CurrentRow >= soundMakerStartRow && m.CurrentRow-soundMakerStartRow+m.ScrollOffset < len(availableSoundMakers) {
					soundMakerIndex := m.CurrentRow - soundMakerStartRow + m.ScrollOffset
					columnStatus = fmt.Sprintf("Select SoundMaker: %s", availableSoundMakers[soundMakerIndex])
				} else {
					columnStatus = "Available SoundMakers"
				}
			}
		} else {
			columnStatus = "Unknown SoundMaker"
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

		// Initialize parameters if needed
		settings.InitializeParameters()

		// Always show Name row first
		var nameCell string
		if m.CurrentRow == 0 {
			nameCell = styles.Selected.Render(settings.Name)
		} else {
			nameCell = styles.Normal.Render(settings.Name)
		}
		content.WriteString(fmt.Sprintf("  %-12s %s\n", styles.Label.Render("Name:"), nameCell))
		content.WriteString("\n")

		// Get instrument definition and render parameters in two columns
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			col0, col1 := def.GetParametersSortedByColumn()

			// Build list of all parameters for row indexing
			allParams := append(col0, col1...)

			// Two-column layout
			content.WriteString(styles.Label.Render("Parameters:"))
			content.WriteString("\n\n")

			maxRows := len(col0)
			if len(col1) > maxRows {
				maxRows = len(col1)
			}

			for i := 0; i < maxRows; i++ {
				var leftParam, rightParam *types.InstrumentParameterDef
				var leftRowIndex, rightRowIndex int = -1, -1

				// Get parameters for this row
				if i < len(col0) {
					leftParam = &col0[i]
					// Find the row index in allParams for left param
					for idx, p := range allParams {
						if p.Key == leftParam.Key {
							leftRowIndex = idx + 1 // +1 because name is row 0
							break
						}
					}
				}

				if i < len(col1) {
					rightParam = &col1[i]
					// Find the row index in allParams for right param
					for idx, p := range allParams {
						if p.Key == rightParam.Key {
							rightRowIndex = idx + 1 // +1 because name is row 0
							break
						}
					}
				}

				// Render left column
				leftContent := ""
				if leftParam != nil {
					value := settings.GetParameterValue(leftParam.Key)
					var valueStr string

					// Special formatting for DX7 preset
					if leftParam.Key == "preset" && settings.Name == "DX7" {
						if value == -1 {
							valueStr = "--"
						} else {
							patchName, err := supercollider.GetDX7PatchName(value)
							if err == nil {
								valueStr = fmt.Sprintf("%s", patchName) // Show just the name for space
							} else {
								valueStr = fmt.Sprintf("%d", value)
							}
						}
					} else {
						if value == -1 {
							valueStr = "--"
						} else {
							if leftParam.Type == types.ParameterTypeHex {
								valueStr = fmt.Sprintf("%02X", value)
							} else {
								valueStr = fmt.Sprintf("%d", value)
							}
						}
					}

					var valueCell string
					if m.CurrentRow == leftRowIndex {
						valueCell = styles.Selected.Render(valueStr)
					} else {
						valueCell = styles.Normal.Render(valueStr)
					}
					leftContent = fmt.Sprintf("%-6s %s", styles.Label.Render(leftParam.DisplayName+":"), valueCell)
				}

				// Render right column
				rightContent := ""
				if rightParam != nil {
					value := settings.GetParameterValue(rightParam.Key)
					var valueStr string

					if value == -1 {
						valueStr = "--"
					} else {
						if rightParam.Type == types.ParameterTypeHex {
							valueStr = fmt.Sprintf("%02X", value)
						} else {
							valueStr = fmt.Sprintf("%d", value)
						}
					}

					var valueCell string
					if m.CurrentRow == rightRowIndex {
						valueCell = styles.Selected.Render(valueStr)
					} else {
						valueCell = styles.Normal.Render(valueStr)
					}
					rightContent = fmt.Sprintf("%-6s %s", styles.Label.Render(rightParam.DisplayName+":"), valueCell)
				}

				// Combine columns with proper spacing
				if leftContent != "" && rightContent != "" {
					content.WriteString(fmt.Sprintf("  %-20s  %s\n", leftContent, rightContent))
				} else if leftContent != "" {
					content.WriteString(fmt.Sprintf("  %s\n", leftContent))
				} else if rightContent != "" {
					content.WriteString(fmt.Sprintf("  %-20s  %s\n", "", rightContent))
				}
			}
		}

		// Separator line
		content.WriteString("\n")
		content.WriteString(styles.Label.Render("Available SoundMakers:"))
		content.WriteString("\n\n")

		// Available SoundMakers list (scrollable)
		availableSoundMakers := []string{"Polyperc", "Infinite Pad", "DX7"}
		visibleRows := m.GetVisibleRows() - 10 // Reserve space for header, settings, and labels

		// Calculate start row for SoundMaker selection
		var paramCount int
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			paramCount = len(def.Parameters)
		}
		soundMakerStartRow := 1 + paramCount // 1 for name row + parameters

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

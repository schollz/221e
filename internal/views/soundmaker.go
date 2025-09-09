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
			// Parameter rows start at row 1
			paramIndex := m.CurrentRow - 1
			if paramIndex >= 0 && paramIndex < len(def.Parameters) {
				param := def.Parameters[paramIndex]
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
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special handling for MiBraids model display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						modelName := types.GetMiBraidsModelName(value)
						columnStatus = fmt.Sprintf("%s: %s (%d)", param.DisplayName, modelName, value)
					}
				} else {
					// Standard parameter display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						if param.Type == types.ParameterTypeHex {
							columnStatus = fmt.Sprintf("%s: %02X", param.DisplayName, value)
						} else if param.Type == types.ParameterTypeFloat {
							floatValue := float32(value) / 1000.0
							columnStatus = fmt.Sprintf("%s: %.2f", param.DisplayName, floatValue)
						} else {
							columnStatus = fmt.Sprintf("%s: %d", param.DisplayName, value)
						}
					}
				}
			} else {
				columnStatus = "Use Up/Down to navigate parameters"
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

		// Show description if available
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists && def.Description != "" {
			content.WriteString(fmt.Sprintf("  %-12s %s\n", styles.Label.Render("Description:"), styles.Normal.Render(def.Description)))
		}
		content.WriteString("\n")

		// Get instrument definition and render parameters in single column
		// Always reserve space for maximum parameters (7) to keep stable height
		content.WriteString("\n")

		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			// Render all parameters in a single column, sorted by their original order
			for i, param := range def.Parameters {
				value := settings.GetParameterValue(param.Key)
				var valueStr string

				// Special formatting for DX7 preset
				if param.Key == "preset" && settings.Name == "DX7" {
					if value == -1 {
						valueStr = "--"
					} else {
						patchName, err := supercollider.GetDX7PatchName(value)
						if err == nil {
							valueStr = fmt.Sprintf("%s", patchName)
						} else {
							valueStr = fmt.Sprintf("%d", value)
						}
					}
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special formatting for MiBraids model
					if value == -1 {
						valueStr = "--"
					} else {
						modelName := types.GetMiBraidsModelName(value)
						valueStr = fmt.Sprintf("%s", modelName)
					}
				} else {
					if value == -1 {
						valueStr = "--"
					} else {
						if param.Type == types.ParameterTypeHex {
							valueStr = fmt.Sprintf("%02X", value)
						} else if param.Type == types.ParameterTypeFloat {
							// Display float parameters as normalized values
							floatValue := float32(value) / 1000.0
							valueStr = fmt.Sprintf("%.2f", floatValue)
						} else {
							valueStr = fmt.Sprintf("%d", value)
						}
					}
				}

				// Row index is i+1 because name is row 0
				paramRowIndex := i + 1
				var valueCell string
				if m.CurrentRow == paramRowIndex {
					valueCell = styles.Selected.Render(valueStr)
				} else {
					valueCell = styles.Normal.Render(valueStr)
				}

				row := fmt.Sprintf("  %-10s %s", styles.Label.Render(param.DisplayName+":"), valueCell)
				content.WriteString(row)
				content.WriteString("\n")
			}

			// Add empty rows to maintain consistent height (max parameters is 7)
			const maxParameters = 7
			for i := len(def.Parameters); i < maxParameters; i++ {
				content.WriteString("\n") // Empty row for consistent spacing
			}
		} else {
			// If no instrument definition found, add empty rows to maintain height
			const maxParameters = 7
			for i := 0; i < maxParameters; i++ {
				content.WriteString("\n")
			}
		}

		return content.String()
	}, statusMsg, 15) // Fixed height for stable view
}

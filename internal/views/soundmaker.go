package views

import (
	"fmt"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/supercollider"
	"github.com/schollz/collidertracker/internal/types"
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
						patchName, err := supercollider.GetDX7PatchName(int(value))
						if err == nil {
							columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, patchName, value)
						} else {
							columnStatus = fmt.Sprintf("%s: %.0f", param.DisplayName, value)
						}
					}
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special handling for MiBraids model display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						modelName := types.GetMiBraidsModelName(int(value))
						columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, modelName, value)
					}
				} else if param.Key == "engine" && settings.Name == "MiPlaits" {
					// Special handling for MiPlaits engine display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						engineName := types.GetMiPlaitsEngineName(int(value))
						columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, engineName, value)
					}
				} else {
					// Standard parameter display
					if value == -1 {
						columnStatus = fmt.Sprintf("%s: --", param.DisplayName)
					} else {
						// Use DisplayFormat if available, otherwise use default formatting
						if param.DisplayFormat != "" {
							formattedValue := fmt.Sprintf(param.DisplayFormat, value)
							columnStatus = fmt.Sprintf("%s: %s", param.DisplayName, formattedValue)
						} else if param.Type == types.ParameterTypeHex {
							columnStatus = fmt.Sprintf("%s: %02X", param.DisplayName, int(value))
						} else if param.Type == types.ParameterTypeFloat {
							columnStatus = fmt.Sprintf("%s: %.2f", param.DisplayName, value)
						} else {
							columnStatus = fmt.Sprintf("%s: %.0f", param.DisplayName, value)
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

	baseMsg := fmt.Sprintf("Up/Down: Navigate | SPACE: Select SoundMaker | %s+Arrow: Adjust values | Shift+Left: Back to Phrase view", input.GetModifierKey())
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
						patchName, err := supercollider.GetDX7PatchName(int(value))
						if err == nil {
							valueStr = fmt.Sprintf("%s", patchName)
						} else {
							valueStr = fmt.Sprintf("%.0f", value)
						}
					}
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special formatting for MiBraids model
					if value == -1 {
						valueStr = "--"
					} else {
						modelName := types.GetMiBraidsModelName(int(value))
						valueStr = fmt.Sprintf("%s", modelName)
					}
				} else if param.Key == "engine" && settings.Name == "MiPlaits" {
					// Special formatting for MiPlaits engine
					if value == -1 {
						valueStr = "--"
					} else {
						engineName := types.GetMiPlaitsEngineName(int(value))
						valueStr = fmt.Sprintf("%s", engineName)
					}
				} else {
					if value == -1 {
						valueStr = "--"
					} else {
						// Use DisplayFormat if available, otherwise use default formatting
						if param.DisplayFormat != "" {
							valueStr = fmt.Sprintf(param.DisplayFormat, value)
						} else if param.Type == types.ParameterTypeHex {
							valueStr = fmt.Sprintf("%02X", int(value))
						} else if param.Type == types.ParameterTypeFloat {
							// Display float parameters directly
							valueStr = fmt.Sprintf("%.2f", value)
						} else {
							valueStr = fmt.Sprintf("%.0f", value)
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

			// Add empty rows to maintain consistent height (max parameters is 9)
			const maxParameters = 9
			for i := len(def.Parameters); i < maxParameters; i++ {
				content.WriteString("\n") // Empty row for consistent spacing
			}
		} else {
			// If no instrument definition found, add empty rows to maintain height
			const maxParameters = 9
			for i := 0; i < maxParameters; i++ {
				content.WriteString("\n")
			}
		}

		return content.String()
	}, statusMsg, 15) // Fixed height for stable view
}

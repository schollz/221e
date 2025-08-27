package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
)

func RenderArpeggioView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "Arpeggio Settings", fmt.Sprintf("Arpeggio %02X", m.ArpeggioEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Render header for the arpeggio table
		headerRow := fmt.Sprintf("  %-4s %-4s %-4s", styles.Label.Render("Row"), styles.Label.Render("DI"), styles.Label.Render("CO"))
		content.WriteString(headerRow)
		content.WriteString("\n")

		// Get current arpeggio settings
		settings := m.ArpeggioSettings[m.ArpeggioEditingIndex]

		// Render 16 rows (00 to 0F), each with its own DI and CO values
		for row := 0; row < 16; row++ {
			// Row label
			rowLabel := fmt.Sprintf("%02X", row)

			// Get DI and CO values for this specific row
			arpeggioRow := settings.Rows[row]

			// Direction (DI) text for this row
			var diText string
			switch arpeggioRow.Direction {
			case 0:
				diText = "--"
			case 1:
				diText = "u-"
			case 2:
				diText = "d-"
			default:
				diText = "--"
			}

			// Count (CO) text for this row
			coText := "--"
			if arpeggioRow.Count != -1 {
				coText = fmt.Sprintf("%02X", arpeggioRow.Count)
			}

			// Direction (DI) column - selectable if this row and column are selected
			var diCell string
			if m.CurrentRow == row && m.CurrentCol == 0 { // Column 0 for DI
				diCell = styles.Selected.Render(diText)
			} else {
				diCell = styles.Normal.Render(diText)
			}

			// Count (CO) column - selectable if this row and column are selected
			var coCell string
			if m.CurrentRow == row && m.CurrentCol == 1 { // Column 1 for CO
				coCell = styles.Selected.Render(coText)
			} else {
				coCell = styles.Normal.Render(coText)
			}

			rowData := fmt.Sprintf("  %-4s %-4s %-4s", styles.Label.Render(rowLabel), diCell, coCell)
			content.WriteString(rowData)
			content.WriteString("\n")
		}

		return content.String()
	}, "Up/Down: Navigate rows | Left/Right: Navigate columns | Ctrl+Arrow: Adjust values | Shift+Left: Back to Phrase view", 18) // 16 rows + 1 header + 1 spacing
}

package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

// RenderSongView renders the new song view with 8 tracks × 16 rows
func RenderSongView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "", "", func(styles ViewStyles) string {
		var content strings.Builder

		// Render header with song name on the right (like Phrase View)
		columnHeader := "    "
		for track := 0; track < 8; track++ {
			columnHeader += fmt.Sprintf("  T%d", track)
		}
		songHeader := "Song"
		content.WriteString(RenderHeader(m, columnHeader, songHeader))

		// Render 16 rows of data
		visibleRows := 16 // Song view always shows all 16 rows
		for row := 0; row < visibleRows; row++ {
			// Row indicator (no playback arrow here - arrows go per track)
			rowIndicator := fmt.Sprintf(" %02X ", row)
			content.WriteString(rowIndicator)

			// Render each track column
			for track := 0; track < 8; track++ {
				// Check if this specific track is playing and on current song row
				trackPlaying := false
				if m.IsPlaying && m.PlaybackMode == types.SongView {
					if m.SongPlaybackActive[track] && m.SongPlaybackRow[track] == row {
						trackPlaying = true
					}
				}
				chainID := m.SongData[track][row]
				var chainCell string

				// Format the chain ID with playback marker
				var baseCell string
				if chainID == -1 {
					baseCell = "--"
				} else {
					baseCell = fmt.Sprintf("%02X", chainID)
				}

				if trackPlaying {
					chainCell = styles.Playback.Render("▶") + baseCell
				} else {
					chainCell = " " + baseCell
				}

				// Determine cell styling
				isSelected := (m.CurrentRow == row && m.CurrentCol == track)

				if isSelected {
					// Selected cell
					chainCell = styles.Selected.Render(chainCell)
				} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.SongView &&
					m.Clipboard.HighlightRow == row && m.Clipboard.HighlightCol == track {
					// Copied cell
					chainCell = styles.Copied.Render(chainCell)
				} else if chainID == -1 {
					// Empty chain - dimmed
					chainCell = styles.Label.Render(chainCell)
				} else {
					// Check if this chain has actual data (any phrase assigned)
					hasChainData := false
					if chainID >= 0 && chainID < len(m.ChainsData) {
						for row := 0; row < 16; row++ {
							if m.ChainsData[chainID][row] != -1 {
								hasChainData = true
								break
							}
						}
					}

					if hasChainData {
						// Chain has data - normal style
						chainCell = styles.Normal.Render(chainCell)
					} else {
						// Chain exists but is empty - dimmed
						chainCell = styles.Label.Render(chainCell)
					}
				}

				content.WriteString(" " + chainCell)
			}

			content.WriteString("\n")
		}

		return content.String()
	}, GetSongStatusMessage(m), 17) // 16 rows + 1 for header
}

// GetSongStatusMessage returns the status message for song view
func GetSongStatusMessage(m *model.Model) string {
	trackCol := m.CurrentCol
	songRow := m.CurrentRow
	chainID := m.SongData[trackCol][songRow]

	var statusMsg string

	if chainID == -1 {
		statusMsg = "No chain selected"
	} else {
		// Check if chain has data and get first phrase for display
		hasChainData := false
		firstPhraseID := -1
		if chainID >= 0 && chainID < len(m.ChainsData) {
			for row := 0; row < 16; row++ {
				if m.ChainsData[chainID][row] != -1 {
					hasChainData = true
					if firstPhraseID == -1 {
						firstPhraseID = m.ChainsData[chainID][row]
					}
				}
			}
		}

		if hasChainData {
			statusMsg = fmt.Sprintf("Track %d Row %02X: Chain %02X (First: %02X)", trackCol, songRow, chainID, firstPhraseID)
		} else {
			statusMsg = fmt.Sprintf("Track %d Row %02X: Chain %02X (Empty)", trackCol, songRow, chainID)
		}
	}

	// Add playback info
	if m.IsPlaying {
		if m.PlaybackMode == types.SongView {
			activeTracksCount := 0
			for i := 0; i < 8; i++ {
				if m.SongPlaybackActive[i] {
					activeTracksCount++
				}
			}
			statusMsg += fmt.Sprintf(" | Song playing (%d tracks) (SPACE to stop)", activeTracksCount)
		} else {
			statusMsg += " | Chain/Phrase playing (SPACE to stop)"
		}
	} else {
		statusMsg += " | Stopped (SPACE to play)"
	}

	statusMsg += " | Shift+Right: Enter Chain | Ctrl+Arrows: Edit chain"
	return statusMsg
}

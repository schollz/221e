package views

import (
	"fmt"
	"strings"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

// RenderSongView renders the new song view with 8 tracks × 16 rows
func RenderSongView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "", "", func(styles *ViewStyles) string {
		var content strings.Builder

		// Render header with song name on the right (like Phrase View)
		columnHeader := "    "
		for track := 0; track < 8; track++ {
			columnHeader += fmt.Sprintf("  T%d", track)
		}
		songHeader := "Song"
		content.WriteString(RenderHeader(m, columnHeader, songHeader))

		// Render track type toggle row (IN/SA)
		typeRowIndicator := "    "
		content.WriteString(typeRowIndicator)
		for track := 0; track < 8; track++ {
			var trackTypeText string
			if m.TrackTypes[track] {
				trackTypeText = " SA" // Sampler
			} else {
				trackTypeText = " IN" // Instrument
			}

			// Check if this track type cell is selected
			// We'll use row -1 to represent the type row
			var typeCell string
			if m.CurrentRow == -1 && m.CurrentCol == track {
				typeCell = " " + styles.Selected.Render(trackTypeText)
			} else {
				typeCell = " " + styles.Label.Render(trackTypeText)
			}
			content.WriteString(typeCell)
		}
		content.WriteString("\n")

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
					content.WriteString(" " + styles.Selected.Render(chainCell))
				} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.SongView &&
					m.Clipboard.HighlightRow == row && m.Clipboard.HighlightCol == track {
					// Copied cell
					content.WriteString(" " + styles.Copied.Render(chainCell))
				} else if chainID == -1 {
					// Empty chain - dimmed
					content.WriteString(" " + styles.Label.Render(chainCell))
				} else {
					// Check if this chain has actual data (any phrase assigned)
					hasChainData := false
					chainsData := m.GetChainsDataForTrack(track)
					if chainID >= 0 && chainID < len(*chainsData) {
						for row := 0; row < 16; row++ {
							if (*chainsData)[chainID][row] != -1 {
								hasChainData = true
								break
							}
						}
					}

					if hasChainData {
						// Chain has data - normal style
						content.WriteString(" " + styles.Normal.Render(chainCell))
					} else {
						// Chain exists but is empty - dimmed
						content.WriteString(" " + styles.Label.Render(chainCell))
					}
				}
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

	var statusMsg string

	// Handle TYPE row (row -1)
	if songRow == -1 {
		var trackTypeText string
		if m.TrackTypes[trackCol] {
			trackTypeText = "Sampler"
		} else {
			trackTypeText = "Instrument"
		}
		statusMsg = fmt.Sprintf("Track %d Type: %s", trackCol, trackTypeText)
	} else {
		// Handle normal data rows
		chainID := m.SongData[trackCol][songRow]

		// Determine track type
		var trackType string
		if m.TrackTypes[trackCol] {
			trackType = "Sampler"
		} else {
			trackType = "Instrument"
		}

		if chainID == -1 {
			statusMsg = fmt.Sprintf("Track %d: %s", trackCol, trackType)
		} else {
			// Check if chain has data and get first phrase for display
			hasChainData := false
			firstPhraseID := -1
			chainsData := m.GetChainsDataForTrack(trackCol)
			if chainID >= 0 && chainID < len(*chainsData) {
				for row := 0; row < 16; row++ {
					if (*chainsData)[chainID][row] != -1 {
						hasChainData = true
						if firstPhraseID == -1 {
							firstPhraseID = (*chainsData)[chainID][row]
						}
					}
				}
			}

			// Determine track type
			var trackType string
			if m.TrackTypes[trackCol] {
				trackType = "Sampler"
			} else {
				trackType = "Instrument"
			}

			if hasChainData {
				statusMsg = fmt.Sprintf("Track %d: %s", trackCol, trackType)
			} else {
				statusMsg = fmt.Sprintf("Track %d: %s (Empty)", trackCol, trackType)
			}
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

package views

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

// Common styles used across all views
type ViewStyles struct {
	Selected      lipgloss.Style
	Normal        lipgloss.Style
	Label         lipgloss.Style
	Container     lipgloss.Style
	Playback      lipgloss.Style
	Copied        lipgloss.Style
	Chain         lipgloss.Style
	Slice         lipgloss.Style
	SliceDownbeat lipgloss.Style
	Dir           lipgloss.Style
	AssignedFile  lipgloss.Style
}

// getCommonStyles returns the standard style definitions used across views
func getCommonStyles() ViewStyles {
	return ViewStyles{
		Selected:      lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")),
		Normal:        lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Label:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Container:     lipgloss.NewStyle().Padding(1, 2),
		Playback:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Copied:        lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")),
		Chain:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Slice:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		SliceDownbeat: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Dir:           lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		AssignedFile:  lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")),
	}
}

// renderViewWithCommonPattern provides a common structure for rendering views
func renderViewWithCommonPattern(m *model.Model, leftHeader, rightHeader string, renderContent func(styles ViewStyles) string, statusMsg string, contentLines int) string {
	styles := getCommonStyles()

	// Content builder - same pattern as working views
	var content strings.Builder

	// Render header (includes waveform) - same as working views
	content.WriteString(RenderHeader(m, leftHeader, rightHeader))

	// Render view-specific content
	content.WriteString(renderContent(styles))

	// Render footer
	content.WriteString(RenderFooter(m, contentLines, statusMsg))

	// Apply container padding to entire content - same as working views
	return styles.Container.Render(content.String())
}

// cellHighlighting handles common cell highlighting logic for clipboard operations
func cellHighlighting(m *model.Model, viewType types.ViewMode, row, col, phrase int, text string, styles ViewStyles, isSelected bool) string {
	if isSelected {
		return styles.Selected.Render(text)
	} else if m.Clipboard.HasData && m.Clipboard.HighlightView == viewType &&
		m.Clipboard.HighlightRow == row &&
		(phrase == -1 || m.Clipboard.HighlightPhrase == phrase) {
		// Highlight copied cell or row
		if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == col) {
			return styles.Copied.Render(text)
		}
	}
	return styles.Normal.Render(text)
}

func getRecordingIndicator(m *model.Model) string {
	if m.RecordingActive {
		// Closed red circle for active recording
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("●")
	} else if m.RecordingEnabled {
		// Open red circle for queued recording
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("○")
	}
	// No indicator when recording is disabled
	return ""
}

// RenderHeader renders the common waveform + header pattern used by all views
func RenderHeader(m *model.Model, leftContent, rightContent string) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	var content strings.Builder

	// Render waveform
	cellsHigh := (types.WaveformHeight+1)/2 - 1 // round up consistently
	waveWidth := m.TermWidth - 4                // account for container padding
	if waveWidth < 1 {
		waveWidth = 1
	}

	// If no waveform data available, create a simple test pattern to show the waveform area
	waveformData := m.WaveformBuf
	if len(waveformData) == 0 {
		// Generate a simple sine wave for display when no OSC data is available
		testLength := waveWidth * 2 / 3
		if testLength < 10 {
			testLength = 10
		}
		waveformData = make([]float64, testLength)
		for i := range waveformData {
			waveformData[i] = 0.5 * math.Sin(2*math.Pi*float64(i)/float64(testLength)*3)
		}
	}

	content.WriteString(RenderWaveform(waveWidth, cellsHigh, waveformData))
	content.WriteString("\n")

	// Build header with recording indicator
	recordingIndicator := getRecordingIndicator(m)

	// Calculate available space for padding (account for container padding)
	availableWidth := m.TermWidth - 4 // Container padding (2 on each side)
	leftLen := len(leftContent)
	rightLen := len(rightContent)
	indicatorLen := 0
	if recordingIndicator != "" {
		indicatorLen = 2 // Space + circle
	}

	// Ensure we have enough space
	paddingSize := availableWidth - leftLen - rightLen - indicatorLen
	if paddingSize < 1 {
		paddingSize = 1
	}

	// Build full header
	fullHeader := leftContent
	if rightContent != "" {
		fullHeader += strings.Repeat(" ", paddingSize) + rightContent
	}
	if recordingIndicator != "" {
		fullHeader += " " + recordingIndicator
	}

	content.WriteString(headerStyle.Render(fullHeader))
	content.WriteString("\n")

	return content.String()
}

// RenderFooter handles the common pattern of filling remaining space and adding status
func RenderFooter(m *model.Model, contentLines int, statusMsg string) string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	var content strings.Builder

	// Fill remaining space if terminal is larger
	if m.TermHeight > 0 && contentLines < m.TermHeight-4 { // -4 for container padding
		for i := contentLines; i < m.TermHeight-4; i++ {
			content.WriteString("\n")
		}
	}

	// Status message
	content.WriteString(statusStyle.Render(statusMsg))

	return content.String()
}

func RenderChainView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "", "", func(styles ViewStyles) string {
		var content strings.Builder

		// Render header with chain name on the right (like Phrase View)
		columnHeader := "      PH"
		chainHeader := fmt.Sprintf("Chain %02X", m.CurrentChain)
		content.WriteString(RenderHeader(m, columnHeader, chainHeader))

		// Render 16 rows of the current chain
		visibleRows := 16            // Always show all 16 rows of a chain
		chainIndex := m.CurrentChain // We need to track which chain we're viewing

		for row := 0; row < visibleRows; row++ {
			// Row indicator with playback arrow
			rowIndicator := fmt.Sprintf(" %02X ", row)

			// Check if this chain row is currently playing (track-specific)
			playingOnThisRow := false
			if m.IsPlaying {
				if m.PlaybackMode == types.SongView {
					// In song playback, check if this track's chain row is playing
					if m.SongPlaybackActive[m.CurrentTrack] &&
						m.SongPlaybackChain[m.CurrentTrack] == chainIndex {
						// Find which chain row contains the current phrase
						currentPhrase := m.SongPlaybackPhrase[m.CurrentTrack]
						if m.ChainsData[chainIndex][row] == currentPhrase {
							playingOnThisRow = true
						}
					}
				} else if m.PlaybackMode == types.ChainView {
					// In chain playback, check if this is the current chain row
					if m.PlaybackChain == chainIndex && m.PlaybackChainRow == row {
						playingOnThisRow = true
					}
				}
			}

			if playingOnThisRow {
				rowIndicator = styles.Playback.Render(fmt.Sprintf("▶%02X ", row))
			}

			content.WriteString(rowIndicator)

			// Get phrase ID for this chain row
			phraseID := m.ChainsData[chainIndex][row]
			var phraseCell string

			// Format the phrase ID
			if phraseID == -1 {
				phraseCell = "--"
			} else {
				phraseCell = fmt.Sprintf("%02X", phraseID)
			}

			// Determine cell styling
			isSelected := (m.CurrentRow == row)

			if isSelected {
				// Selected cell
				phraseCell = styles.Selected.Render(phraseCell)
			} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.ChainView &&
				m.Clipboard.HighlightRow == row {
				// Copied cell
				phraseCell = styles.Copied.Render(phraseCell)
			} else if phraseID == -1 {
				// Empty phrase - dimmed
				phraseCell = styles.Label.Render(phraseCell)
			} else {
				// Normal style
				phraseCell = styles.Normal.Render(phraseCell)
			}

			content.WriteString("  " + phraseCell)
			content.WriteString("\n")
		}

		return content.String()
	}, GetChainStatusMessage(m), 17) // 16 rows + 1 for header
}

func RenderPhraseView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")) // Lighter background, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	sliceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sliceDownbeatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))                          // Lighter gray for downbeats
	playbackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))                              // Green
	copiedStyle := lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")) // Yellow background

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header (Я is a single-character column)
	columnHeader := "  SL  P  NN  DT  GT  RT  TS  Я  PA  LP  HP  CO  VE  FI"
	phraseHeader := fmt.Sprintf("Phrase %02X", m.CurrentPhrase)
	content.WriteString(RenderHeader(m, columnHeader, phraseHeader))

	// Data rows
	visibleRows := m.GetVisibleRows()
	for i := 0; i < visibleRows && i+m.ScrollOffset < 255; i++ {
		dataIndex := i + m.ScrollOffset

		// Arrow for current row or playback
		arrow := " "
		if m.IsPlaying {
			if m.PlaybackMode == types.SongView {
				// Song playback mode - check all tracks to see which ones are playing this phrase at this row
				playingTracksCount := 0
				for track := 0; track < 8; track++ {
					if m.SongPlaybackActive[track] &&
						m.SongPlaybackPhrase[track] == m.CurrentPhrase &&
						m.SongPlaybackRowInPhrase[track] == dataIndex {
						playingTracksCount++
					}
				}
				if playingTracksCount == 1 {
					arrow = playbackStyle.Render("▶")
				} else if playingTracksCount > 1 {
					// Multiple tracks playing at this position - show double arrow
					arrow = playbackStyle.Render("▶▶")
				}
			} else {
				// Chain/Phrase playback mode - use existing logic
				if m.PlaybackPhrase == m.CurrentPhrase && m.PlaybackRow == dataIndex {
					arrow = playbackStyle.Render("▶")
				}
			}
		} else if m.CurrentRow == dataIndex {
			// Not playing - show cursor arrow
			arrow = "▶"
		}

		// Slice number (hex)
		sliceHex := fmt.Sprintf("%02X", dataIndex)
		var sliceCell string
		if dataIndex%4 == 0 {
			sliceCell = sliceDownbeatStyle.Render(sliceHex) // Lighter for downbeats
		} else {
			sliceCell = sliceStyle.Render(sliceHex)
		}

		// Playback (P)
		playbackText := "0"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColPlayback] == 1 {
			playbackText = "1"
		}
		var playbackCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 1 {
			playbackCell = selectedStyle.Render(playbackText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 1) {
				playbackCell = copiedStyle.Render(playbackText)
			} else {
				playbackCell = normalStyle.Render(playbackText)
			}
		} else {
			playbackCell = normalStyle.Render(playbackText)
		}

		// Note (NN)
		noteText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColNote] != -1 {
			noteText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColNote])
		}
		var noteCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 2 {
			noteCell = selectedStyle.Render(noteText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 2) {
				noteCell = copiedStyle.Render(noteText)
			} else {
				noteCell = normalStyle.Render(noteText)
			}
		} else {
			noteCell = normalStyle.Render(noteText)
		}

		// Delta Time (DT)
		dtText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColDeltaTime] != -1 {
			dtText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColDeltaTime])
		}
		var dtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 3 {
			dtCell = selectedStyle.Render(dtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 3) {
				dtCell = copiedStyle.Render(dtText)
			} else {
				dtCell = normalStyle.Render(dtText)
			}
		} else {
			dtCell = normalStyle.Render(dtText)
		}

		// Gate (GT)
		gtText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColGate] != -1 {
			gtText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColGate])
		}
		var gtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 4 {
			gtCell = selectedStyle.Render(gtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 4) {
				gtCell = copiedStyle.Render(gtText)
			} else {
				gtCell = normalStyle.Render(gtText)
			}
		} else {
			gtCell = normalStyle.Render(gtText)
		}

		// Retrigger (RT)
		rtText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColRetrigger] != -1 {
			rtText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColRetrigger])
		}
		var rtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 5 {
			rtCell = selectedStyle.Render(rtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 5) {
				rtCell = copiedStyle.Render(rtText)
			} else {
				rtCell = normalStyle.Render(rtText)
			}
		} else {
			rtCell = normalStyle.Render(rtText)
		}

		// Timestretch (TS)
		tsText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColTimestretch] != -1 {
			tsText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColTimestretch])
		}
		var tsCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 6 {
			tsCell = selectedStyle.Render(tsText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 6) {
				tsCell = copiedStyle.Render(tsText)
			} else {
				tsCell = normalStyle.Render(tsText)
			}
		} else {
			tsCell = normalStyle.Render(tsText)
		}

		// Я (EffectReverse) — single char: "-", "0", or "1"
		revText := "-"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != -1 {
			if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != 0 {
				revText = "1"
			} else {
				revText = "0"
			}
		}
		var revCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 7 {
			revCell = selectedStyle.Render(revText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 7) {
				revCell = copiedStyle.Render(revText)
			} else {
				revCell = normalStyle.Render(revText)
			}
		} else {
			revCell = normalStyle.Render(revText)
		}

		// PA (Pan)
		paText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColPan] != -1 {
			paText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColPan])
		}
		var paCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 8 {
			paCell = selectedStyle.Render(paText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 8) {
				paCell = copiedStyle.Render(paText)
			} else {
				paCell = normalStyle.Render(paText)
			}
		} else {
			paCell = normalStyle.Render(paText)
		}

		// LP (LowPassFilter)
		lpText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColLowPassFilter] != -1 {
			lpText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColLowPassFilter])
		}
		var lpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 9 {
			lpCell = selectedStyle.Render(lpText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 9) {
				lpCell = copiedStyle.Render(lpText)
			} else {
				lpCell = normalStyle.Render(lpText)
			}
		} else {
			lpCell = normalStyle.Render(lpText)
		}

		// HP (HighPassFilter)
		hpText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColHighPassFilter] != -1 {
			hpText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColHighPassFilter])
		}
		var hpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 10 {
			hpCell = selectedStyle.Render(hpText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 10) {
				hpCell = copiedStyle.Render(hpText)
			} else {
				hpCell = normalStyle.Render(hpText)
			}
		} else {
			hpCell = normalStyle.Render(hpText)
		}

		// CO (EffectComb)
		combText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectComb] != -1 {
			combText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectComb])
		}
		var combCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 11 {
			combCell = selectedStyle.Render(combText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 11) {
				combCell = copiedStyle.Render(combText)
			} else {
				combCell = normalStyle.Render(combText)
			}
		} else {
			combCell = normalStyle.Render(combText)
		}

		// VE (EffectReverb)
		reverbText := "--"
		if m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectReverb] != -1 {
			reverbText = fmt.Sprintf("%02X", m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColEffectReverb])
		}
		var reverbCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 12 {
			reverbCell = selectedStyle.Render(reverbText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 12) {
				reverbCell = copiedStyle.Render(reverbText)
			} else {
				reverbCell = normalStyle.Render(reverbText)
			}
		} else {
			reverbCell = normalStyle.Render(reverbText)
		}

		// Filename (FI) - first 8 characters (now column 13 in UI)
		fiText := "--------"
		fileIndex := m.PhrasesData[m.CurrentPhrase][dataIndex][types.ColFilename]
		if fileIndex >= 0 && fileIndex < len(m.PhrasesFiles) && m.PhrasesFiles[fileIndex] != "" {
			fullPath := m.PhrasesFiles[fileIndex]
			filename := filepath.Base(fullPath)
			if len(filename) > 8 {
				fiText = filename[:8]
			} else {
				fiText = fmt.Sprintf("%-8s", filename)
			}
		}
		var fiCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 13 {
			fiCell = selectedStyle.Render(fiText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 13) {
				fiCell = copiedStyle.Render(fiText)
			} else {
				fiCell = normalStyle.Render(fiText)
			}
		} else {
			fiCell = normalStyle.Render(fiText)
		}

		// NOTE the %-1s for Я to keep it one character wide
		row := fmt.Sprintf("%s %-3s  %-1s  %-3s  %-3s  %-3s  %-3s  %-3s  %-1s  %-3s  %-3s  %-3s  %-3s  %-3s  %-8s",
			arrow, sliceCell, playbackCell, noteCell, dtCell, gtCell, rtCell, tsCell, revCell, paCell, lpCell, hpCell, combCell, reverbCell, fiCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	statusMsg := GetPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
}

func RenderSettingsView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "Settings", "", func(styles ViewStyles) string {
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

func RenderFileMetadataView(m *model.Model) string {
	filename := filepath.Base(m.MetadataEditingFile)
	header := fmt.Sprintf("File Metadata: %s", filename)

	return renderViewWithCommonPattern(m, header, "", func(styles ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current metadata or defaults
		metadata, exists := m.FileMetadata[m.MetadataEditingFile]
		if !exists {
			metadata = types.FileMetadata{BPM: 120.0, Slices: 16} // Default values
		}

		// Metadata settings with common rendering pattern
		settings := []struct {
			label string
			value string
			row   int
		}{
			{"BPM:", fmt.Sprintf("%.2f", metadata.BPM), 0},
			{"Slices:", fmt.Sprintf("%d", metadata.Slices), 1},
		}

		for _, setting := range settings {
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

		content.WriteString("\n")

		// File info
		fileInfo := fmt.Sprintf("File: %s", m.MetadataEditingFile)
		content.WriteString(styles.Normal.Render(fileInfo))
		content.WriteString("\n\n")

		return content.String()
	}, "Up/Down: Navigate | Ctrl+Arrow: Adjust values | Shift+Down: Back to File Browser", 7)
}

func RenderFileView(m *model.Model) string {
	header := fmt.Sprintf("File Browser: %s", m.CurrentDir)

	return renderViewWithCommonPattern(m, header, "", func(styles ViewStyles) string {
		var content strings.Builder

		// File list
		visibleRows := m.GetVisibleRows()
		for i := 0; i < visibleRows && i+m.ScrollOffset < len(m.Files); i++ {
			dataIndex := i + m.ScrollOffset

			// Arrow for current selection
			arrow := " "
			if m.CurrentRow == dataIndex {
				arrow = "▶"
			}

			// File/directory name with appropriate styling
			filename := m.Files[dataIndex]
			var fileCell string
			if m.CurrentRow == dataIndex {
				fileCell = styles.Selected.Render(filename)
			} else if strings.HasSuffix(filename, "/") || filename == ".." {
				fileCell = styles.Dir.Render(filename)
			} else if IsCurrentRowFile(m, filename) {
				fileCell = styles.AssignedFile.Render(filename)
			} else {
				fileCell = styles.Normal.Render(filename)
			}

			row := fmt.Sprintf("%s %s", arrow, fileCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		return content.String()
	}, "SPACE: Select file | Ctrl+Right: Play/Stop | Shift+Left: Back to phrase", m.GetVisibleRows()+1)
}

func GetChainStatusMessage(m *model.Model) string {
	phraseID := m.ChainsData[m.CurrentChain][m.CurrentRow]

	var statusMsg string
	if phraseID == -1 {
		statusMsg = fmt.Sprintf("Chain %02X Row %02X: --", m.CurrentChain, m.CurrentRow)
	} else {
		statusMsg = fmt.Sprintf("Chain %02X Row %02X: Phrase %02X", m.CurrentChain, m.CurrentRow, phraseID)
	}

	statusMsg += " | Shift+Right: Enter phrase | Left/Right: Switch chains | Ctrl+Arrow: Edit phrase"
	return statusMsg
}
func GetPhraseStatusMessage(m *model.Model) string {
	var statusMsg string

	// Use enum-driven UI indices to avoid hardcoding
	rtUI := int(types.ColRetrigger) + 1
	fiUI := int(types.ColFilename) + 1

	if m.CurrentCol == rtUI {
		// On retrigger column - show retrigger info
		retriggerIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColRetrigger]
		if retriggerIndex >= 0 && retriggerIndex < 255 {
			statusMsg = fmt.Sprintf("Retrigger: %02X", retriggerIndex)
		} else {
			statusMsg = "No retrigger selected"
		}
	} else if m.CurrentCol == fiUI {
		// On filename column - show file info
		fileIndex := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		if fileIndex >= 0 && fileIndex < len(m.PhrasesFiles) && m.PhrasesFiles[fileIndex] != "" {
			statusMsg = fmt.Sprintf("File: %s", m.PhrasesFiles[fileIndex])
		} else {
			statusMsg = "No file selected"
		}
	} else {
		// On other columns - show value
		colIndex := m.CurrentCol - 1
		value := m.PhrasesData[m.CurrentPhrase][m.CurrentRow][colIndex]
		if colIndex == int(types.ColPlayback) {
			statusMsg = fmt.Sprintf("Playback enabled: %d", value)
		} else if colIndex == int(types.ColGate) {
			gateFloat := float32(value) / 96.0
			statusMsg = fmt.Sprintf("Gate: %02X (%.2f)", value, gateFloat)
		} else if colIndex == int(types.ColPan) {
			// PA (Pan) column - show -1.0 to 1.0 mapping, -1 means center (0.0)
			if value == -1 {
				statusMsg = "Pan: -- (0.0, sticky)"
			} else {
				panFloat := (float32(value) - 127.0) / 127.0 // Map 0-254 to -1.0 to 1.0, with 128 as center (0.0)
				statusMsg = fmt.Sprintf("Pan: %02X (%.2f, sticky)", value, panFloat)
			}
		} else if colIndex == int(types.ColLowPassFilter) {
			// LP (Low Pass Filter) column - show exponential frequency mapping
			if value == -1 {
				statusMsg = "Low Pass: -- (20kHz, sticky)"
			} else {
				// Exponential mapping: 00 -> 20kHz, FE -> 20Hz
				logMin := float32(1.301) // log10(20)
				logMax := float32(4.301) // log10(20000)
				logFreq := logMax - (float32(value)/254.0)*(logMax-logMin)
				freq := float32(math.Pow(10, float64(logFreq)))
				if freq >= 1000 {
					statusMsg = fmt.Sprintf("Low Pass: %02X (%.1fkHz, sticky)", value, freq/1000)
				} else {
					statusMsg = fmt.Sprintf("Low Pass: %02X (%.0fHz, sticky)", value, freq)
				}
			}
		} else if colIndex == int(types.ColHighPassFilter) {
			// HP (High Pass Filter) column - show exponential frequency mapping
			if value == -1 {
				statusMsg = "High Pass: -- (20Hz, sticky)"
			} else {
				// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
				logMin := float32(1.301) // log10(20)
				logMax := float32(4.301) // log10(20000)
				logFreq := logMin + (float32(value)/254.0)*(logMax-logMin)
				freq := float32(math.Pow(10, float64(logFreq)))
				if freq >= 1000 {
					statusMsg = fmt.Sprintf("High Pass: %02X (%.1fkHz, sticky)", value, freq/1000)
				} else {
					statusMsg = fmt.Sprintf("High Pass: %02X (%.0fHz, sticky)", value, freq)
				}
			}
		} else if colIndex == int(types.ColEffectReverse) {
			// Reverse (Я) column - show Off/On
			if value == -1 {
				statusMsg = "Reverse: -- (Off)"
			} else if value == 0 {
				statusMsg = "Reverse: 0 (Off)"
			} else {
				statusMsg = "Reverse: 1 (On)"
			}
		} else if colIndex == int(types.ColEffectComb) {
			// CO (Comb) column - show 0.0 to 1.0 mapping
			if value == -1 {
				statusMsg = "Comb: -- (sticky)"
			} else {
				combFloat := float32(value) / 254.0
				statusMsg = fmt.Sprintf("Comb: %02X (%.2f, sticky)", value, combFloat)
			}
		} else if colIndex == int(types.ColEffectReverb) {
			// VE (Reverb) column - show 0.0 to 1.0 mapping
			if value == -1 {
				statusMsg = "Reverb: -- (sticky)"
			} else {
				reverbFloat := float32(value) / 254.0
				statusMsg = fmt.Sprintf("Reverb: %02X (%.2f, sticky)", value, reverbFloat)
			}
		} else if value == -1 {
			statusMsg = "Current value: --"
		} else {
			statusMsg = fmt.Sprintf("Current value: %d", value)
		}
	}

	if m.IsPlaying {
		if m.PlaybackMode == types.ChainView {
			statusMsg += fmt.Sprintf(" | Chain playing (C:%02X P:%02X) (SPACE to stop)", m.PlaybackChain, m.PlaybackPhrase)
		} else {
			statusMsg += " | Phrase playing (SPACE to stop)"
		}
	} else {
		statusMsg += " | Stopped (SPACE to play)"
	}

	statusMsg += " | Shift+Right: File browser | Shift+Left: Back to chain view"
	return statusMsg
}

func IsCurrentRowFile(m *model.Model, filename string) bool {
	// Check if this file is assigned to the current fileSelectRow
	fileIndex := m.PhrasesData[m.CurrentPhrase][m.FileSelectRow][types.ColFilename]
	if fileIndex >= 0 && fileIndex < len(m.PhrasesFiles) && m.PhrasesFiles[fileIndex] != "" {
		assignedFile := m.PhrasesFiles[fileIndex]
		fullPath := filepath.Join(m.CurrentDir, filename)
		return assignedFile == filename || assignedFile == fullPath
	}
	return false
}

func RenderTimestrechView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")) // Lighter background, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header
	header := "Timestretch Settings"
	timestrechHeader := fmt.Sprintf("Timestretch %02X", m.TimestrechEditingIndex)
	content.WriteString(RenderHeader(m, header, timestrechHeader))
	content.WriteString("\n")

	// Get current timestretch settings
	settings := m.TimestrechSettings[m.TimestrechEditingIndex]

	// Start setting
	startLabel := "Start:"
	startValue := fmt.Sprintf("%.2fx", settings.Start)
	var startCell string
	if m.CurrentRow == 0 {
		startCell = selectedStyle.Render(startValue)
	} else {
		startCell = normalStyle.Render(startValue)
	}
	startRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(startLabel), startCell)
	content.WriteString(startRow)
	content.WriteString("\n")

	// End setting
	endLabel := "End:"
	endValue := fmt.Sprintf("%.2fx", settings.End)
	var endCell string
	if m.CurrentRow == 1 {
		endCell = selectedStyle.Render(endValue)
	} else {
		endCell = normalStyle.Render(endValue)
	}
	endRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(endLabel), endCell)
	content.WriteString(endRow)
	content.WriteString("\n")

	// Beats setting
	beatsLabel := "Beats:"
	beatsValue := fmt.Sprintf("%d", settings.Beats)
	var beatsCell string
	if m.CurrentRow == 2 {
		beatsCell = selectedStyle.Render(beatsValue)
	} else {
		beatsCell = normalStyle.Render(beatsValue)
	}
	beatsRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(beatsLabel), beatsCell)
	content.WriteString(beatsRow)
	content.WriteString("\n\n")

	// Footer with status
	statusMsg := "Up/Down: Navigate | Ctrl+Arrow: Adjust values | Shift+Left: Back to Phrase view"
	content.WriteString(RenderFooter(m, 6, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}

func RenderRetriggerView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")) // Lighter background, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header
	header := "Retrigger Settings"
	retriggerHeader := fmt.Sprintf("Retrigger %02X", m.RetriggerEditingIndex)
	content.WriteString(RenderHeader(m, header, retriggerHeader))
	content.WriteString("\n")

	// Get current retrigger settings
	settings := m.RetriggerSettings[m.RetriggerEditingIndex]

	// Times setting
	timesLabel := "Times:"
	timesValue := fmt.Sprintf("%d", settings.Times)
	var timesCell string
	if m.CurrentRow == 0 {
		timesCell = selectedStyle.Render(timesValue)
	} else {
		timesCell = normalStyle.Render(timesValue)
	}
	timesRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(timesLabel), timesCell)
	content.WriteString(timesRow)
	content.WriteString("\n")

	// Starting Rate setting
	startLabel := "Starting Rate:"
	startValue := fmt.Sprintf("%.2f/beat", settings.Start)
	var startCell string
	if m.CurrentRow == 1 {
		startCell = selectedStyle.Render(startValue)
	} else {
		startCell = normalStyle.Render(startValue)
	}
	startRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(startLabel), startCell)
	content.WriteString(startRow)
	content.WriteString("\n")

	// Final Rate setting
	endLabel := "Final Rate:"
	endValue := fmt.Sprintf("%.2f/beat", settings.End)
	var endCell string
	if m.CurrentRow == 2 {
		endCell = selectedStyle.Render(endValue)
	} else {
		endCell = normalStyle.Render(endValue)
	}
	endRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(endLabel), endCell)
	content.WriteString(endRow)
	content.WriteString("\n")

	// Beats setting
	beatsLabel := "Beats:"
	beatsValue := fmt.Sprintf("%d", settings.Beats)
	var beatsCell string
	if m.CurrentRow == 3 {
		beatsCell = selectedStyle.Render(beatsValue)
	} else {
		beatsCell = normalStyle.Render(beatsValue)
	}
	beatsRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(beatsLabel), beatsCell)
	content.WriteString(beatsRow)
	content.WriteString("\n")

	// Volume change setting
	volumeLabel := "Volume dB:"
	var volumeValue string
	if settings.VolumeDB >= 0 {
		volumeValue = fmt.Sprintf("+%.1f", settings.VolumeDB)
	} else {
		volumeValue = fmt.Sprintf("%.1f", settings.VolumeDB)
	}
	var volumeCell string
	if m.CurrentRow == 4 {
		volumeCell = selectedStyle.Render(volumeValue)
	} else {
		volumeCell = normalStyle.Render(volumeValue)
	}
	volumeRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(volumeLabel), volumeCell)
	content.WriteString(volumeRow)
	content.WriteString("\n")

	// Pitch change setting
	pitchLabel := "Pitch:"
	var pitchValue string
	if settings.PitchChange >= 0 {
		pitchValue = fmt.Sprintf("+%.1f", settings.PitchChange)
	} else {
		pitchValue = fmt.Sprintf("%.1f", settings.PitchChange)
	}
	var pitchCell string
	if m.CurrentRow == 5 {
		pitchCell = selectedStyle.Render(pitchValue)
	} else {
		pitchCell = normalStyle.Render(pitchValue)
	}
	pitchRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(pitchLabel), pitchCell)
	content.WriteString(pitchRow)
	content.WriteString("\n\n")

	// Footer with status
	statusMsg := "Up/Down: Navigate | Ctrl+Arrow: Adjust values | Shift+Left: Back to Phrase view"
	content.WriteString(RenderFooter(m, 9, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}

// RenderWaveform renders waveform data (assumed in [-1,1]) into a Braille string.
// width and height are in Braille cells. Each Braille cell is 2x4 dots.
func RenderWaveform(width, height int, data []float64) string {
	if width <= 0 || height <= 0 || len(data) == 0 {
		return ""
	}

	fineW := width * 2  // dot-columns
	fineH := height * 4 // dot-rows

	// Helper: sample data at fractional index using linear interpolation.
	sampleAt := func(p float64) float64 {
		if p <= 0 {
			return data[0]
		}
		max := float64(len(data) - 1)
		if p >= max {
			return data[len(data)-1]
		}
		i := int(math.Floor(p))
		f := p - float64(i)
		return data[i]*(1-f) + data[i+1]*f
	}

	// Dot grid: fineH rows x fineW cols
	grid := make([][]bool, fineH)
	for r := range grid {
		grid[r] = make([]bool, fineW)
	}

	// Plot the waveform onto the dot grid.
	for x := 0; x < fineW; x++ {
		// map x-> position in data
		p := (float64(x) / float64(fineW-1)) * float64(len(data)-1)
		v := sampleAt(p) // in [-1,1]

		// Map v to vertical dot index (0 at top)
		y := int(math.Round((1.0 - (v+1.0)/2.0) * float64(fineH-1)))
		if y < 0 {
			y = 0
		} else if y >= fineH {
			y = fineH - 1
		}
		grid[y][x] = true
	}

	var b strings.Builder
	b.Grow((width + 1) * height) // small pre-alloc (each row + newline)

	// Braille dot bit masks
	const (
		dot1 = 0x01
		dot2 = 0x02
		dot3 = 0x04
		dot4 = 0x08
		dot5 = 0x10
		dot6 = 0x20
		dot7 = 0x40
		dot8 = 0x80
	)
	const brailleBase = 0x2800

	// Pack each 2x4 block of dots into a single Braille rune.
	for cellRow := 0; cellRow < height; cellRow++ {
		fr := cellRow * 4 // fine-row base
		for cellCol := 0; cellCol < width; cellCol++ {
			fc := cellCol * 2 // fine-col base

			var mask int
			// Rows 0..3, Cols 0..1
			// Row 0: dots 1(left), 4(right)
			if grid[fr+0][fc+0] {
				mask |= dot1
			}
			if grid[fr+0][fc+1] {
				mask |= dot4
			}
			// Row 1: dots 2,5
			if grid[fr+1][fc+0] {
				mask |= dot2
			}
			if grid[fr+1][fc+1] {
				mask |= dot5
			}
			// Row 2: dots 3,6
			if grid[fr+2][fc+0] {
				mask |= dot3
			}
			if grid[fr+2][fc+1] {
				mask |= dot6
			}
			// Row 3: dots 7,8
			if grid[fr+3][fc+0] {
				mask |= dot7
			}
			if grid[fr+3][fc+1] {
				mask |= dot8
			}

			r := rune(brailleBase + mask)
			b.WriteRune(r)
		}
		if cellRow != height-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// RenderMixerView renders the mixer view with 8 track volume bars
func RenderMixerView(m *model.Model) string {
	styles := getCommonStyles()

	// Content builder
	var content strings.Builder

	// Render header
	leftHeader := fmt.Sprintf("MIXER | Track %d", m.CurrentMixerTrack+1)
	rightHeader := fmt.Sprintf("%.1f BPM", m.BPM)
	content.WriteString(RenderHeader(m, leftHeader, rightHeader))
	content.WriteString("\n")

	// Track headers
	content.WriteString("  ")
	for track := 0; track < 8; track++ {
		trackHeader := fmt.Sprintf(" T%d ", track+1)
		if track == m.CurrentMixerTrack {
			content.WriteString(styles.Selected.Render(trackHeader))
		} else {
			content.WriteString(styles.Normal.Render(trackHeader))
		}
		content.WriteString(" ")
	}
	content.WriteString("\n\n")

	// Volume bars (scale -96 to +12 dB)
	barHeight := 10
	for row := 0; row < barHeight; row++ {
		content.WriteString("  ")
		for track := 0; track < 8; track++ {
			currentVol := m.TrackVolumes[track] // Current volume from SuperCollider
			setLevel := m.TrackSetLevels[track] // User-controllable set level

			// Map volume to bar position (reverse row order, 0 is top)
			// Volume range: -96 to +12 dB = 108 dB range
			currentPos := int((currentVol + 96.0) / 108.0 * float32(barHeight))
			setPos := int((setLevel + 96.0) / 108.0 * float32(barHeight))

			// Display character for this position
			barRow := barHeight - 1 - row // Invert so 0dB is at top
			var char string

			if barRow == setPos {
				// Show set level marker
				char = "━"
			} else if barRow < currentPos {
				// Show current volume level
				char = "█"
			} else {
				// Empty space
				char = " "
			}

			// Apply styling
			barText := fmt.Sprintf(" %s ", char)
			if track == m.CurrentMixerTrack {
				content.WriteString(styles.Selected.Render(barText))
			} else if barRow < currentPos {
				content.WriteString(styles.Playback.Render(barText))
			} else if barRow == setPos {
				content.WriteString(styles.Label.Render(barText))
			} else {
				content.WriteString(styles.Normal.Render(barText))
			}
			content.WriteString(" ")
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Volume values
	content.WriteString("  ")
	for track := 0; track < 8; track++ {
		volText := fmt.Sprintf("%4.1f", m.TrackVolumes[track])
		if track == m.CurrentMixerTrack {
			content.WriteString(styles.Selected.Render(volText))
		} else {
			content.WriteString(styles.Normal.Render(volText))
		}
		content.WriteString("  ")
	}
	content.WriteString("\n")

	// Set level values
	content.WriteString("  ")
	for track := 0; track < 8; track++ {
		setText := fmt.Sprintf("%4.1f", m.TrackSetLevels[track])
		if track == m.CurrentMixerTrack {
			content.WriteString(styles.Selected.Render(setText))
		} else {
			content.WriteString(styles.Label.Render(setText))
		}
		content.WriteString("  ")
	}
	content.WriteString("\n\n")

	// Instructions
	statusMsg := "Left/Right: Select track | Ctrl+Arrow: Adjust set level | Shift+Down: Back to previous view"
	content.WriteString(RenderFooter(m, barHeight+6, statusMsg))

	// Apply container padding
	return styles.Container.Render(content.String())
}

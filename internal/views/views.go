package views

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"

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
func getCommonStyles() *ViewStyles {
	return &ViewStyles{
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
func renderViewWithCommonPattern(m *model.Model, leftHeader, rightHeader string, renderContent func(styles *ViewStyles) string, statusMsg string, contentLines int) string {
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
func cellHighlighting(m *model.Model, viewType types.ViewMode, row, col, phrase int, text string, styles *ViewStyles, isSelected bool) string {
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
	return renderViewWithCommonPattern(m, "", "", func(styles *ViewStyles) string {
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
						// Use track-specific chain data
						chainsData := m.GetChainsDataForTrack(m.CurrentTrack)
						if (*chainsData)[chainIndex][row] == currentPhrase {
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
			chainsData := m.GetCurrentChainsData()
			phraseID := (*chainsData)[chainIndex][row]
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
	// Route to appropriate sub-view based on track context
	phraseViewType := m.GetPhraseViewType()
	if phraseViewType == types.InstrumentPhraseView {
		return RenderInstrumentPhraseView(m)
	}
	return RenderSamplerPhraseView(m)
}

func RenderSamplerPhraseView(m *model.Model) string {
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
	columnHeader := "  SL  P  NN  PI  DT  GT  RT  TS  Я  PA  LP  HP  CO  VE  FI"
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
				// Song playback mode - check only sampler tracks (4-7) for sampler phrase view
				playingTracksCount := 0
				for track := 4; track < 8; track++ {
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

		// Playback (P) - use current phrases data based on view type
		phrasesData := m.GetCurrentPhrasesData()
		playbackText := "0"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPlayback] == 1 {
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
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColNote] != -1 {
			noteText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColNote])
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

		// Pitch (PI)
		pitchText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPitch] != -1 {
			pitchText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPitch])
		}
		var pitchCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 3 {
			pitchCell = selectedStyle.Render(pitchText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 3) {
				pitchCell = copiedStyle.Render(pitchText)
			} else {
				pitchCell = normalStyle.Render(pitchText)
			}
		} else {
			pitchCell = normalStyle.Render(pitchText)
		}

		// Delta Time (DT)
		dtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime] != -1 {
			dtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime])
		}
		var dtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 4 {
			dtCell = selectedStyle.Render(dtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 4) {
				dtCell = copiedStyle.Render(dtText)
			} else {
				dtCell = normalStyle.Render(dtText)
			}
		} else {
			dtCell = normalStyle.Render(dtText)
		}

		// Gate (GT)
		gtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColGate] != -1 {
			gtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColGate])
		}
		var gtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 5 {
			gtCell = selectedStyle.Render(gtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 5) {
				gtCell = copiedStyle.Render(gtText)
			} else {
				gtCell = normalStyle.Render(gtText)
			}
		} else {
			gtCell = normalStyle.Render(gtText)
		}

		// Retrigger (RT)
		rtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRetrigger] != -1 {
			rtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRetrigger])
		}
		var rtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 6 {
			rtCell = selectedStyle.Render(rtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 6) {
				rtCell = copiedStyle.Render(rtText)
			} else {
				rtCell = normalStyle.Render(rtText)
			}
		} else {
			rtCell = normalStyle.Render(rtText)
		}

		// Timestretch (TS)
		tsText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColTimestretch] != -1 {
			tsText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColTimestretch])
		}
		var tsCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 7 {
			tsCell = selectedStyle.Render(tsText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 7) {
				tsCell = copiedStyle.Render(tsText)
			} else {
				tsCell = normalStyle.Render(tsText)
			}
		} else {
			tsCell = normalStyle.Render(tsText)
		}

		// Я (EffectReverse) — single char: "-", "0", or "1"
		revText := "-"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != -1 {
			if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != 0 {
				revText = "1"
			} else {
				revText = "0"
			}
		}
		var revCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 8 {
			revCell = selectedStyle.Render(revText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 8) {
				revCell = copiedStyle.Render(revText)
			} else {
				revCell = normalStyle.Render(revText)
			}
		} else {
			revCell = normalStyle.Render(revText)
		}

		// PA (Pan)
		paText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPan] != -1 {
			paText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPan])
		}
		var paCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 9 {
			paCell = selectedStyle.Render(paText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 9) {
				paCell = copiedStyle.Render(paText)
			} else {
				paCell = normalStyle.Render(paText)
			}
		} else {
			paCell = normalStyle.Render(paText)
		}

		// LP (LowPassFilter)
		lpText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColLowPassFilter] != -1 {
			lpText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColLowPassFilter])
		}
		var lpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 10 {
			lpCell = selectedStyle.Render(lpText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 10) {
				lpCell = copiedStyle.Render(lpText)
			} else {
				lpCell = normalStyle.Render(lpText)
			}
		} else {
			lpCell = normalStyle.Render(lpText)
		}

		// HP (HighPassFilter)
		hpText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColHighPassFilter] != -1 {
			hpText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColHighPassFilter])
		}
		var hpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 11 {
			hpCell = selectedStyle.Render(hpText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 11) {
				hpCell = copiedStyle.Render(hpText)
			} else {
				hpCell = normalStyle.Render(hpText)
			}
		} else {
			hpCell = normalStyle.Render(hpText)
		}

		// CO (EffectComb)
		combText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectComb] != -1 {
			combText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectComb])
		}
		var combCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 12 {
			combCell = selectedStyle.Render(combText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 12) {
				combCell = copiedStyle.Render(combText)
			} else {
				combCell = normalStyle.Render(combText)
			}
		} else {
			combCell = normalStyle.Render(combText)
		}

		// VE (EffectReverb)
		reverbText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverb] != -1 {
			reverbText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverb])
		}
		var reverbCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 13 {
			reverbCell = selectedStyle.Render(reverbText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 13) {
				reverbCell = copiedStyle.Render(reverbText)
			} else {
				reverbCell = normalStyle.Render(reverbText)
			}
		} else {
			reverbCell = normalStyle.Render(reverbText)
		}

		// Filename (FI) - first 8 characters (now column 13 in UI)
		fiText := "--------"
		fileIndex := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColFilename]
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
			fullPath := (*phrasesFiles)[fileIndex]
			filename := filepath.Base(fullPath)
			if len(filename) > 8 {
				fiText = filename[:8]
			} else {
				fiText = fmt.Sprintf("%-8s", filename)
			}
		}
		var fiCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 14 {
			fiCell = selectedStyle.Render(fiText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 14) {
				fiCell = copiedStyle.Render(fiText)
			} else {
				fiCell = normalStyle.Render(fiText)
			}
		} else {
			fiCell = normalStyle.Render(fiText)
		}

		// NOTE the %-1s for Я to keep it one character wide
		row := fmt.Sprintf("%s %-3s  %-1s  %-3s  %-3s  %-3s  %-3s  %-3s  %-3s  %-1s  %-3s  %-3s  %-3s  %-3s  %-3s  %-8s",
			arrow, sliceCell, playbackCell, noteCell, pitchCell, dtCell, gtCell, rtCell, tsCell, revCell, paCell, lpCell, hpCell, combCell, reverbCell, fiCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	statusMsg := GetPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
}

func RenderInstrumentPhraseView(m *model.Model) string {
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

	// Render header for Instrument view (row, playback, note, and chord columns)
	columnHeader := "  SL  P  NOT  CAT  AR"
	phraseHeader := fmt.Sprintf("Instrument %02X", m.CurrentPhrase)
	content.WriteString(RenderHeader(m, columnHeader, phraseHeader))

	// Data rows
	visibleRows := m.GetVisibleRows()
	for i := 0; i < visibleRows && i+m.ScrollOffset < 255; i++ {
		dataIndex := i + m.ScrollOffset

		// Arrow for current row or playback
		arrow := " "
		if m.IsPlaying {
			if m.PlaybackMode == types.SongView {
				// Song playback mode - check only instrument tracks (0-3) for instrument phrase view
				playingTracksCount := 0
				for track := 0; track < 4; track++ {
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

		// Playback (P) column
		phrasesData := m.GetCurrentPhrasesData()
		playbackValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPlayback]
		playbackText := "0"
		if playbackValue == 1 {
			playbackText = "1"
		}

		var playbackCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 1 { // Column 1 is the P column
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

		// Note (NOT) - use ColNote but display as note name
		// For Instrument view, we're using the Note column to store MIDI note values (0-127)
		noteValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColNote]
		noteText := "---"
		if noteValue != -1 {
			noteText = midiToNoteName(noteValue)
		}

		var noteCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 2 { // Column 2 is the NOT column (0=slice, 1=P, 2=NOT)
			noteCell = selectedStyle.Render(fmt.Sprintf("%3s", noteText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 2) {
				noteCell = copiedStyle.Render(fmt.Sprintf("%3s", noteText))
			} else {
				noteCell = normalStyle.Render(fmt.Sprintf("%3s", noteText))
			}
		} else {
			noteCell = normalStyle.Render(fmt.Sprintf("%3s", noteText))
		}

		// Chord (C) - display chord type
		chordValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChord]
		chordText := types.ChordTypeToString(types.ChordType(chordValue))

		var chordCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 3 { // Column 3 is the C column
			chordCell = selectedStyle.Render(fmt.Sprintf("%1s", chordText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 3) {
				chordCell = copiedStyle.Render(fmt.Sprintf("%1s", chordText))
			} else {
				chordCell = normalStyle.Render(fmt.Sprintf("%1s", chordText))
			}
		} else {
			chordCell = normalStyle.Render(fmt.Sprintf("%1s", chordText))
		}

		// Chord Addition (A) - display chord addition
		chordAddValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChordAddition]
		chordAddText := types.ChordAdditionToString(types.ChordAddition(chordAddValue))

		var chordAddCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 4 { // Column 4 is the A column
			chordAddCell = selectedStyle.Render(fmt.Sprintf("%1s", chordAddText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 4) {
				chordAddCell = copiedStyle.Render(fmt.Sprintf("%1s", chordAddText))
			} else {
				chordAddCell = normalStyle.Render(fmt.Sprintf("%1s", chordAddText))
			}
		} else {
			chordAddCell = normalStyle.Render(fmt.Sprintf("%1s", chordAddText))
		}

		// Chord Transposition (T) - display transposition value
		chordTransValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChordTransposition]
		chordTransText := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))

		var chordTransCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 5 { // Column 5 is the T column
			chordTransCell = selectedStyle.Render(fmt.Sprintf("%1s", chordTransText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 5) {
				chordTransCell = copiedStyle.Render(fmt.Sprintf("%1s", chordTransText))
			} else {
				chordTransCell = normalStyle.Render(fmt.Sprintf("%1s", chordTransText))
			}
		} else {
			chordTransCell = normalStyle.Render(fmt.Sprintf("%1s", chordTransText))
		}

		// Arpeggio (AR) - display arpeggio index
		arpeggioValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColArpeggio]
		arpeggioText := "--"
		if arpeggioValue != -1 {
			arpeggioText = fmt.Sprintf("%02X", arpeggioValue)
		}

		var arpeggioCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 6 { // Column 6 is the AR column
			arpeggioCell = selectedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 6) {
				arpeggioCell = copiedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			} else {
				arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			}
		} else {
			arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		}

		row := fmt.Sprintf("%s %-3s  %s  %s  %s%s%s  %s", arrow, sliceCell, playbackCell, noteCell, chordCell, chordAddCell, chordTransCell, arpeggioCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	statusMsg := GetInstrumentPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
}

func RenderSettingsView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "Settings", "", func(styles *ViewStyles) string {
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

	return renderViewWithCommonPattern(m, header, "", func(styles *ViewStyles) string {
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

	return renderViewWithCommonPattern(m, header, "", func(styles *ViewStyles) string {
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
	chainsData := m.GetCurrentChainsData()
	phraseID := (*chainsData)[m.CurrentChain][m.CurrentRow]

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
		phrasesData := m.GetCurrentPhrasesData()
		retriggerIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColRetrigger]
		if retriggerIndex >= 0 && retriggerIndex < 255 {
			statusMsg = fmt.Sprintf("Retrigger: %02X", retriggerIndex)
		} else {
			statusMsg = "No retrigger selected"
		}
	} else if m.CurrentCol == fiUI {
		// On filename column - show file info
		phrasesData := m.GetCurrentPhrasesData()
		fileIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename]
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
			statusMsg = fmt.Sprintf("File: %s", (*phrasesFiles)[fileIndex])
		} else {
			statusMsg = "No file selected"
		}
	} else {
		// On other columns - show value using centralized column mapping
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping == nil || columnMapping.DataColumnIndex == -1 {
			statusMsg = "Column info not available"
		} else {
			phrasesData := m.GetCurrentPhrasesData()
			colIndex := columnMapping.DataColumnIndex
			value := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex]
			if colIndex == int(types.ColPlayback) {
				statusMsg = fmt.Sprintf("Playback enabled: %d", value)
			} else if colIndex == int(types.ColGate) {
				gateFloat := float32(value) / 96.0
				statusMsg = fmt.Sprintf("Gate: %02X (%.2f)", value, gateFloat)
			} else if colIndex == int(types.ColPitch) {
				// PI (Pitch) column - show -24 to +24 mapping, 128 (0x80) means 0.0 pitch
				if value == -1 {
					statusMsg = "Pitch: -- (cleared)"
				} else {
					pitchFloat := ((float32(value) - 128.0) / 128.0) * 24.0 // Map 0-254 to -24 to +24, with 128 as center (0.0)
					statusMsg = fmt.Sprintf("Pitch: %02X (%.1f)", value, pitchFloat)
				}
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

func GetInstrumentPhraseStatusMessage(m *model.Model) string {
	var statusMsg string

	// Use centralized column mapping to determine current column
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	phrasesData := m.GetCurrentPhrasesData()

	if columnMapping != nil && (columnMapping.DataColumnIndex == int(types.ColNote) ||
		columnMapping.DataColumnIndex == int(types.ColChord) ||
		columnMapping.DataColumnIndex == int(types.ColChordAddition) ||
		columnMapping.DataColumnIndex == int(types.ColChordTransposition)) { // NOT, C, A, or T columns
		// Get current row data
		noteValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColNote]
		chordValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChord]
		chordAddValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChordAddition]
		chordTransValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChordTransposition]

		if noteValue >= 0 && noteValue <= 127 {
			noteName := midiToNoteName(noteValue)

			// Check if chord is defined (not null/"-")
			if chordValue > int(types.ChordNone) {
				// Extract note name and octave from MIDI note
				noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
				rootNote := noteNames[noteValue%12]
				octave := (noteValue / 12) - 1

				// Build chord name
				var chordName string
				switch types.ChordType(chordValue) {
				case types.ChordMajor:
					chordName = rootNote + "maj"
				case types.ChordMinor:
					chordName = rootNote + "min"
				case types.ChordDominant:
					chordName = rootNote // Dominant chords have no suffix
				default:
					chordName = rootNote
				}

				// Add chord addition if defined
				if chordAddValue > int(types.ChordAddNone) {
					switch types.ChordAddition(chordAddValue) {
					case types.ChordAdd7:
						chordName += "7"
					case types.ChordAdd9:
						chordName += "9"
					case types.ChordAdd4:
						chordName += "4"
					}
				}

				// Add transposition if defined and not 0
				if chordTransValue > int(types.ChordTrans0) {
					transpositionStr := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))
					statusMsg = fmt.Sprintf("Chord: %s (octave %d, transpose %s)", chordName, octave, transpositionStr)
				} else {
					statusMsg = fmt.Sprintf("Chord: %s (octave %d)", chordName, octave)
				}
			} else {
				// Chord is null, show simple note info with transposition if defined and not 0
				if chordTransValue > int(types.ChordTrans0) {
					transpositionStr := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))
					statusMsg = fmt.Sprintf("Note: %s (transpose %s)", noteName, transpositionStr)
				} else {
					statusMsg = fmt.Sprintf("Note: %s", noteName)
				}
			}
		} else {
			statusMsg = "No note selected"
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColPlayback) { // P column
		// Show playback info
		playbackValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColPlayback]
		if playbackValue == 1 {
			statusMsg = "Playback: ON"
		} else {
			statusMsg = "Playback: OFF"
		}
	} else {
		// On other columns - show basic info
		statusMsg = fmt.Sprintf("Instrument Phrase %02X Row %02X", m.CurrentPhrase, m.CurrentRow)
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

	statusMsg += " | Shift+Left: Back to chain view"
	return statusMsg
}

func IsCurrentRowFile(m *model.Model, filename string) bool {
	// Check if this file is assigned to the current fileSelectRow
	phrasesData := m.GetCurrentPhrasesData()
	fileIndex := (*phrasesData)[m.CurrentPhrase][m.FileSelectRow][types.ColFilename]
	phrasesFiles := m.GetCurrentPhrasesFiles()
	if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
		assignedFile := (*phrasesFiles)[fileIndex]
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
// RenderWaveform renders waveform data (assumed in [-1,1]) into a Braille string.
// width and height are in Braille cells. Each Braille cell is 2x4 dots.
//
// Optimized to minimize allocations/copies:
//   - Avoids building a [][]bool grid; uses a flat []byte mask per 2x4 cell.
//   - Single pass over fine columns sets bits directly into cell masks.
//   - Reuses small helpers and bounds checks to keep hot loop tight.
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

	// Each cell is a single byte mask for its 2x4 Braille dots.
	masks := make([]byte, width*height)

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

	// Precompute total data span
	span := float64(len(data) - 1)
	if span <= 0 {
		span = 1
	}

	// Single pass over fine X columns; set the one dot hit per column.
	for x := 0; x < fineW; x++ {
		p := (float64(x) / float64(fineW-1)) * span
		v := sampleAt(p) // in [-1,1]

		// Map v to vertical dot index (0 at top)
		y := int(math.Round((1.0 - (v+1.0)/2.0) * float64(fineH-1)))
		if y < 0 {
			y = 0
		} else if y >= fineH {
			y = fineH - 1
		}

		cellCol := x >> 1 // /2
		cellRow := y >> 2 // /4
		inCol := x & 1    // 0..1
		inRow := y & 3    // 0..3

		// Map (inRow,inCol) -> braille dot bit
		var bit byte
		switch inRow {
		case 0:
			if inCol == 0 {
				bit = dot1
			} else {
				bit = dot4
			}
		case 1:
			if inCol == 0 {
				bit = dot2
			} else {
				bit = dot5
			}
		case 2:
			if inCol == 0 {
				bit = dot3
			} else {
				bit = dot6
			}
		default: // 3
			if inCol == 0 {
				bit = dot7
			} else {
				bit = dot8
			}
		}

		idx := cellRow*width + cellCol
		masks[idx] |= bit
	}

	var b strings.Builder
	// Each cell becomes 1 rune; each row has width runes + 1 newline (except last).
	b.Grow(height*width + (height - 1))

	for row := 0; row < height; row++ {
		base := row * width
		for col := 0; col < width; col++ {
			mask := masks[base+col]
			r := rune(brailleBase + int(mask))
			b.WriteRune(r)
		}
		if row != height-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// midiToNoteName converts MIDI note number (0-127) to note name like "c-1", "c#4", etc.
// For negative octaves: natural notes show minus (e.g., "c-1"), sharp notes drop minus (e.g., "f#1") - all stay 3 chars
// MIDI note 60 = C4, note 21 = A0, etc.
func midiToNoteName(midiNote int) string {
	if midiNote < 0 || midiNote > 127 {
		return "---"
	}

	noteNames := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}

	// Calculate octave (MIDI note 12 = C0)
	octave := (midiNote / 12) - 1

	// Get note name
	noteName := noteNames[midiNote%12]

	// Always maintain exactly 3 characters for all notes
	if strings.Contains(noteName, "#") {
		// Sharp notes: "c#4", "f#1" (already 3 chars for most cases)
		if octave < 0 {
			return fmt.Sprintf("%s%d", noteName, -octave) // "c#1" for negative
		} else {
			return fmt.Sprintf("%s%d", noteName, octave) // "c#4" for positive
		}
	} else {
		// Natural notes: always use minus separator to reach 3 chars
		if octave < 0 {
			return fmt.Sprintf("%s-%d", noteName, -octave) // "c-1" for negative
		} else {
			return fmt.Sprintf("%s-%d", noteName, octave) // "c-4" for positive
		}
	}
}

// noteNameToMidi converts note name back to MIDI note number for validation
func noteNameToMidi(noteName string) int {
	if noteName == "---" || len(noteName) < 2 {
		return -1
	}

	noteMap := map[string]int{
		"c": 0, "c#": 1, "d": 2, "d#": 3, "e": 4, "f": 5,
		"f#": 6, "g": 7, "g#": 8, "a": 9, "a#": 10, "b": 11,
	}

	// Extract note and octave parts
	var noteStr string
	var octaveStr string

	if len(noteName) >= 3 && noteName[1] == '#' {
		// Sharp note like "c#4"
		noteStr = noteName[:2]
		octaveStr = noteName[2:]
	} else {
		// Natural note like "c4"
		noteStr = noteName[:1]
		octaveStr = noteName[1:]
	}

	noteVal, noteExists := noteMap[noteStr]
	if !noteExists {
		return -1
	}

	// Parse octave
	var octave int
	if _, err := fmt.Sscanf(octaveStr, "%d", &octave); err != nil {
		return -1
	}

	midiNote := (octave+1)*12 + noteVal
	if midiNote < 0 || midiNote > 127 {
		return -1
	}

	return midiNote
}

// getLevelColorSmooth returns a smoothly interpolated color for a dB level
func getLevelColorSmooth(dbLevel float32) colorful.Color {
	db := float64(dbLevel)

	// Define color stops for smooth gradient (updated for -48 to +12 dB range)
	veryLowColor, _ := colorful.Hex("#404040") // Dark gray
	lowColor, _ := colorful.Hex("#808080")     // Gray
	normalColor, _ := colorful.Hex("#FFFFFF")  // White
	warmColor, _ := colorful.Hex("#FFE135")    // Subtle yellow
	hotColor, _ := colorful.Hex("#FF6B35")     // Orange-red
	clipColor, _ := colorful.Hex("#FF0000")    // Red

	// Create smooth transitions between color zones (adjusted for -48 to +12 dB)
	if db <= -36 {
		// Very low: dark gray to gray (-48 to -36 dB)
		t := (db + 48.0) / 12.0
		t = math.Max(0, math.Min(1, t))
		t = ease.OutQuad(t)
		return veryLowColor.BlendHcl(lowColor, t)
	} else if db <= -24 {
		// Low: gray to white (-36 to -24 dB)
		t := (db + 36.0) / 12.0
		t = ease.OutQuad(t)
		return lowColor.BlendHcl(normalColor, t)
	} else if db <= -12 {
		// Normal: white to subtle yellow (-24 to -12 dB)
		t := (db + 24.0) / 12.0
		t = ease.InQuad(t)
		return normalColor.BlendHcl(warmColor, t)
	} else if db <= -3 {
		// Approaching hot: subtle yellow to orange (-12 to -3 dB)
		t := (db + 12.0) / 9.0
		t = ease.InCubic(t)
		return warmColor.BlendHcl(hotColor, t)
	} else {
		// Hot to clipping: orange to red (-3 to +12 dB)
		t := (db + 3.0) / 15.0
		t = math.Max(0, math.Min(1, t))
		t = ease.InExpo(t)
		return hotColor.BlendHcl(clipColor, t)
	}
}

// getBackgroundColorSmooth returns a smooth background/empty color
func getBackgroundColorSmooth(position float64) colorful.Color {
	// Subtle gradient for empty sections
	topColor, _ := colorful.Hex("#2A2A2A")    // Darker at top
	bottomColor, _ := colorful.Hex("#3A3A3A") // Slightly lighter at bottom

	t := ease.InOutQuad(position)
	return topColor.BlendHcl(bottomColor, t)
}

// getUnicodeBlock returns the appropriate Unicode block character for a fill ratio (0-1)
func getUnicodeBlock(fillRatio float64) string {
	if fillRatio <= 0 {
		return "  " // Empty
	} else if fillRatio <= 0.125 {
		return "▁▁" // 1/8 block
	} else if fillRatio <= 0.25 {
		return "▂▂" // 2/8 block
	} else if fillRatio <= 0.375 {
		return "▃▃" // 3/8 block
	} else if fillRatio <= 0.5 {
		return "▄▄" // 4/8 block
	} else if fillRatio <= 0.625 {
		return "▅▅" // 5/8 block
	} else if fillRatio <= 0.75 {
		return "▆▆" // 6/8 block
	} else if fillRatio <= 0.875 {
		return "▇▇" // 7/8 block
	} else {
		return "██" // Full block
	}
}

// createVerticalBar creates a simple vertical level meter bar with Unicode blocks
func createVerticalBar(currentLevel, setLevel float32, height int, isSelected bool) []string {
	// Convert dB range (-48 to +12) to bar height scale (60 dB range)
	currentPos := (float64(currentLevel) + 48.0) / 60.0 * float64(height)
	setPos := (float64(setLevel) + 48.0) / 60.0 * float64(height)

	// Clamp positions to valid range
	currentPos = math.Max(0, math.Min(float64(height), currentPos))
	setPos = math.Max(0, math.Min(float64(height), setPos))

	lines := make([]string, height)
	profile := termenv.ColorProfile()

	// Choose colors based on selection
	var fillColor, emptyColor colorful.Color
	if isSelected {
		fillColor, _ = colorful.Hex("#FFFFFF")  // White for selected
		emptyColor, _ = colorful.Hex("#808080") // Medium gray for empty parts of selected
	} else {
		fillColor, _ = colorful.Hex("#C0C0C0")  // Light gray for unselected
		emptyColor, _ = colorful.Hex("#404040") // Dark gray for empty parts of unselected
	}

	// Fill from bottom to top (invert the display)
	for row := 0; row < height; row++ {
		displayRow := float64(height - 1 - row) // Invert so 0dB is at top

		var barContent string
		var color colorful.Color

		// Determine what to show at this row
		if math.Abs(displayRow-setPos) < 0.5 && math.Abs(displayRow-currentPos) > 0.5 {
			// Set level marker (horizontal line)
			barContent = "━━"
			color = fillColor // Use same color as fill for consistency
		} else if displayRow < currentPos {
			// Full fill
			color = fillColor
			barContent = "██"
		} else if displayRow >= currentPos && displayRow < currentPos+1 {
			// Partial fill - use Unicode blocks for smooth edges
			partialFill := currentPos - math.Floor(currentPos) // Fractional part

			if partialFill > 0 {
				color = fillColor
				// Use appropriate Unicode block
				barContent = getUnicodeBlock(partialFill)
			} else {
				// Empty
				barContent = "▒▒"
				color = emptyColor
			}
		} else {
			// Empty space
			barContent = "▒▒"
			color = emptyColor
		}

		// Apply color using termenv
		colorHex := color.Hex()
		termColor := profile.Color(colorHex)
		styledContent := termenv.String(barContent).Foreground(termColor).String()

		lines[row] = styledContent
	}

	return lines
}

// dbToHex converts dB value (-48 to +12) to hex (00 to FE)
func dbToHex(db float32) int {
	// Clamp to valid range
	if db < -48.0 {
		db = -48.0
	}
	if db > 12.0 {
		db = 12.0
	}

	// Map -48 to +12 dB (60 dB range) to 0 to 254 (255 values)
	hex := int(((db + 48.0) * 254.0) / 60.0)
	if hex < 0 {
		hex = 0
	}
	if hex > 254 {
		hex = 254
	}
	return hex
}

// hexToDb converts hex value (0 to 254) back to dB (-48 to +12)
func hexToDb(hex int) float32 {
	if hex < 0 {
		hex = 0
	}
	if hex > 254 {
		hex = 254
	}

	// Map 0 to 254 back to -48 to +12 dB
	return ((float32(hex) * 60.0) / 254.0) - 48.0
}

// getMixerStatusMessage returns the status message for mixer view
func getMixerStatusMessage(m *model.Model) string {
	track := m.CurrentMixerTrack
	setLevel := m.TrackSetLevels[track]

	statusMsg := fmt.Sprintf("Track %d: Set %.1fdB (Hex %02X)",
		track+1, setLevel, dbToHex(setLevel))
	statusMsg += " | Left/Right: Select track │ Ctrl+Arrow: Adjust set level │ Shift+Up: Back to previous view"

	return statusMsg
}

// RenderMixerView renders a modern, sleek mixer view with vertical level meters
func RenderMixerView(m *model.Model) string {
	// Column headers (matching song view format)
	columnHeader := "    " // 4 spaces for left padding like song view row numbers
	for track := 0; track < 8; track++ {
		columnHeader += fmt.Sprintf("  T%d", track+1)
	}
	mixerHeader := fmt.Sprintf("Track %d", m.CurrentMixerTrack+1)

	return renderViewWithCommonPattern(m, columnHeader, mixerHeader, func(styles *ViewStyles) string {
		var content strings.Builder

		// Calculate bar height - smaller to fit with waveform
		barHeight := 8
		if m.TermHeight > 25 {
			barHeight = 10
		}

		// Create vertical bars for all tracks
		trackBars := make([][]string, 8)
		for track := 0; track < 8; track++ {
			isSelected := track == m.CurrentMixerTrack
			trackBars[track] = createVerticalBar(m.TrackVolumes[track], m.TrackSetLevels[track], barHeight, isSelected)
		}

		// Render the vertical bars row by row
		for row := 0; row < barHeight; row++ {
			content.WriteString("    ") // Left padding like song view
			for track := 0; track < 8; track++ {
				content.WriteString("  ") // 2 spaces before each track (like song view)
				content.WriteString(trackBars[track][row])
			}
			content.WriteString("\n")
		}

		// Current level values row (hex codes)
		content.WriteString("    ")
		for track := 0; track < 8; track++ {
			content.WriteString("  ")
			currentLevel := m.TrackVolumes[track]
			levelHex := fmt.Sprintf("%02X", dbToHex(currentLevel))

			if track == m.CurrentMixerTrack {
				content.WriteString(styles.Selected.Render(levelHex))
			} else {
				content.WriteString(styles.Normal.Render(levelHex))
			}
		}
		content.WriteString("\n")

		// Set level values row (hex codes)
		content.WriteString("    ")
		for track := 0; track < 8; track++ {
			content.WriteString("  ")
			setLevel := m.TrackSetLevels[track]
			setHex := fmt.Sprintf("%02X", dbToHex(setLevel))

			if track == m.CurrentMixerTrack && m.CurrentMixerRow == 0 {
				content.WriteString(styles.Selected.Render(setHex))
			} else {
				content.WriteString(styles.Label.Render(setHex))
			}
		}
		content.WriteString("\n\n")

		return content.String()
	}, getMixerStatusMessage(m), 14)
}

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

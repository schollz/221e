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
	columnHeader := "  SL  DT  NN  PI  GT  RT  TS  Я  PA  LP  HP  CO  VE  FI"
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
				// Song playback mode - check the current track context
				if m.SongPlaybackActive[m.CurrentTrack] &&
					m.SongPlaybackPhrase[m.CurrentTrack] == m.CurrentPhrase &&
					m.SongPlaybackRowInPhrase[m.CurrentTrack] == dataIndex {
					arrow = playbackStyle.Render("▶")
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

		// Delta Time (DT) - now moved to position 1 (replacing P)
		phrasesData := m.GetCurrentPhrasesData()
		dtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime] != -1 {
			dtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime])
		}
		var dtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 1 {
			dtCell = selectedStyle.Render(dtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 1) {
				dtCell = copiedStyle.Render(dtText)
			} else {
				dtCell = normalStyle.Render(dtText)
			}
		} else {
			dtCell = normalStyle.Render(dtText)
		}

		// Note (NN) - now at position 2
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

		// Pitch (PI) - now at position 3
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

		// Gate (GT) - now at position 4
		gtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColGate] != -1 {
			gtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColGate])
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

		// Retrigger (RT) - now at position 5
		rtText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRetrigger] != -1 {
			rtText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRetrigger])
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

		// Timestretch (TS) - now at position 6
		tsText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColTimestretch] != -1 {
			tsText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColTimestretch])
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

		// Я (EffectReverse) — single char: "-", "0", or "1" - now at position 7
		revText := "-"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != -1 {
			if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverse] != 0 {
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

		// PA (Pan) - now at position 8
		paText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPan] != -1 {
			paText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPan])
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

		// LP (LowPassFilter) - now at position 9
		lpText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColLowPassFilter] != -1 {
			lpText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColLowPassFilter])
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

		// HP (HighPassFilter) - now at position 10
		hpText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColHighPassFilter] != -1 {
			hpText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColHighPassFilter])
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

		// CO (EffectComb) - now at position 11
		combText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectComb] != -1 {
			combText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectComb])
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

		// VE (EffectReverb) - now at position 12
		reverbText := "--"
		if (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverb] != -1 {
			reverbText = fmt.Sprintf("%02X", (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverb])
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

		// Filename (FI) - first 8 characters - now at position 13
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
		row := fmt.Sprintf("%s %-3s  %-3s  %-3s  %-3s  %-3s  %-3s  %-3s  %-1s  %-3s  %-3s  %-3s  %-3s  %-3s  %-8s",
			arrow, sliceCell, dtCell, noteCell, pitchCell, gtCell, rtCell, tsCell, revCell, paCell, lpCell, hpCell, combCell, reverbCell, fiCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	statusMsg := GetPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
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
			if colIndex == int(types.ColDeltaTime) {
				// DT (Delta Time) column - show ticks and playback status
				if value == -1 {
					statusMsg = "Delta Time: -- (row not played)"
				} else if value == 0 {
					statusMsg = "Delta Time: 00 (row not played)"
				} else {
					statusMsg = fmt.Sprintf("Delta Time: %02X (%d ticks, row played)", value, value)
				}
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
					// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
					logMin := float32(1.301) // log10(20)
					logMax := float32(4.301) // log10(20000)
					logFreq := logMin + (float32(value)/254.0)*(logMax-logMin)
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
			} else if colIndex == int(types.ColTimestretch) {
				// TS (Timestretch) column - show timestretch info
				if value == -1 {
					statusMsg = "No timestretch selected"
				} else {
					statusMsg = fmt.Sprintf("Timestretch: %02X", value)
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

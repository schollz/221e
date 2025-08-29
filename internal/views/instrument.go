package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/2n/internal/input"
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/music"
	"github.com/schollz/2n/internal/types"
)

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
	columnHeader := "  SL  DT  NOT  CAT  A D S R   AR  MI  SO"
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

		// Delta Time (DT) column - unified playback control for both Sampler and Instrument views
		phrasesData := m.GetCurrentPhrasesData()
		dtValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime]
		dtText := input.GetEffectiveDTValue(dtValue)

		var dtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 1 { // Column 1 is the DT column
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

		// Note (NOT) - use ColNote but display as note name
		// For Instrument view, we're using the Note column to store MIDI note values (0-127)
		noteValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColNote]
		noteText := "---"
		if noteValue != -1 {
			noteText = music.MidiToNoteName(noteValue)
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

		// Attack (A) - display attack value
		attackValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColAttack]
		attackText := "--"
		if attackValue != -1 {
			attackText = fmt.Sprintf("%02X", attackValue)
		}

		var attackCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 6 { // Column 6 is the A column
			attackCell = selectedStyle.Render(fmt.Sprintf("%2s", attackText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 6) {
				attackCell = copiedStyle.Render(fmt.Sprintf("%2s", attackText))
			} else {
				attackCell = normalStyle.Render(fmt.Sprintf("%2s", attackText))
			}
		} else {
			attackCell = normalStyle.Render(fmt.Sprintf("%2s", attackText))
		}

		// Decay (D) - display decay value
		decayValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDecay]
		decayText := "--"
		if decayValue != -1 {
			decayText = fmt.Sprintf("%02X", decayValue)
		}

		var decayCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 7 { // Column 7 is the D column
			decayCell = selectedStyle.Render(fmt.Sprintf("%2s", decayText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 7) {
				decayCell = copiedStyle.Render(fmt.Sprintf("%2s", decayText))
			} else {
				decayCell = normalStyle.Render(fmt.Sprintf("%2s", decayText))
			}
		} else {
			decayCell = normalStyle.Render(fmt.Sprintf("%2s", decayText))
		}

		// Sustain (S) - display sustain value
		sustainValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColSustain]
		sustainText := "--"
		if sustainValue != -1 {
			sustainText = fmt.Sprintf("%02X", sustainValue)
		}

		var sustainCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 8 { // Column 8 is the S column
			sustainCell = selectedStyle.Render(fmt.Sprintf("%2s", sustainText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 8) {
				sustainCell = copiedStyle.Render(fmt.Sprintf("%2s", sustainText))
			} else {
				sustainCell = normalStyle.Render(fmt.Sprintf("%2s", sustainText))
			}
		} else {
			sustainCell = normalStyle.Render(fmt.Sprintf("%2s", sustainText))
		}

		// Release (R) - display release value
		releaseValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRelease]
		releaseText := "--"
		if releaseValue != -1 {
			releaseText = fmt.Sprintf("%02X", releaseValue)
		}

		var releaseCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 9 { // Column 9 is the R column
			releaseCell = selectedStyle.Render(fmt.Sprintf("%2s", releaseText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 9) {
				releaseCell = copiedStyle.Render(fmt.Sprintf("%2s", releaseText))
			} else {
				releaseCell = normalStyle.Render(fmt.Sprintf("%2s", releaseText))
			}
		} else {
			releaseCell = normalStyle.Render(fmt.Sprintf("%2s", releaseText))
		}

		// Arpeggio (AR) - display arpeggio index
		arpeggioValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColArpeggio]
		arpeggioText := "--"
		if arpeggioValue != -1 {
			arpeggioText = fmt.Sprintf("%02X", arpeggioValue)
		}

		var arpeggioCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 10 { // Column 10 is the AR column (was 6, now 10 due to ADSR)
			arpeggioCell = selectedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 10) {
				arpeggioCell = copiedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			} else {
				arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			}
		} else {
			arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		}

		// MIDI (MI) - display MIDI index
		midiValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidi]
		midiText := "--"
		if midiValue != -1 {
			midiText = fmt.Sprintf("%02X", midiValue)
		}

		var midiCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 11 { // Column 11 is the MI column (after AR)
			midiCell = selectedStyle.Render(fmt.Sprintf("%2s", midiText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 11) {
				midiCell = copiedStyle.Render(fmt.Sprintf("%2s", midiText))
			} else {
				midiCell = normalStyle.Render(fmt.Sprintf("%2s", midiText))
			}
		} else {
			midiCell = normalStyle.Render(fmt.Sprintf("%2s", midiText))
		}

		// SoundMaker (SO) - display SoundMaker index
		soundMakerValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColSoundMaker]
		soundMakerText := "--"
		if soundMakerValue != -1 {
			soundMakerText = fmt.Sprintf("%02X", soundMakerValue)
		}

		var soundMakerCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == 12 { // Column 12 is the SO column (after MI)
			soundMakerCell = selectedStyle.Render(fmt.Sprintf("%2s", soundMakerText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == 12) {
				soundMakerCell = copiedStyle.Render(fmt.Sprintf("%2s", soundMakerText))
			} else {
				soundMakerCell = normalStyle.Render(fmt.Sprintf("%2s", soundMakerText))
			}
		} else {
			soundMakerCell = normalStyle.Render(fmt.Sprintf("%2s", soundMakerText))
		}

		row := fmt.Sprintf("%s %-3s  %s  %s  %s%s%s  %s%s%s%s  %s  %s  %s", arrow, sliceCell, dtCell, noteCell, chordCell, chordAddCell, chordTransCell, attackCell, decayCell, sustainCell, releaseCell, arpeggioCell, midiCell, soundMakerCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	statusMsg := GetInstrumentPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
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
			noteName := music.MidiToNoteName(noteValue)

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
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColDeltaTime) { // DT column
		// Show DT playback info
		playbackValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime]
		if playbackValue > 0 {
			statusMsg = fmt.Sprintf("DT: %02X (play %d ticks)", playbackValue, playbackValue)
		} else {
			statusMsg = "DT: not played"
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColAttack) { // A column
		// Show Attack info
		attackValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColAttack]
		if attackValue == -1 {
			// Check for effective (sticky) Attack value
			effectiveAttackValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColAttack), m.CurrentTrack)
			if effectiveAttackValue == -1 {
				statusMsg = "Attack: -- (sticky)"
			} else {
				attackSeconds := types.AttackToSeconds(effectiveAttackValue)
				statusMsg = fmt.Sprintf("Attack: -- (%.3fs, sticky)", attackSeconds)
			}
		} else {
			attackSeconds := types.AttackToSeconds(attackValue)
			statusMsg = fmt.Sprintf("Attack: %02X (%.3fs, sticky)", attackValue, attackSeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColDecay) { // D column
		// Show Decay info
		decayValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDecay]
		if decayValue == -1 {
			// Check for effective (sticky) Decay value
			effectiveDecayValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColDecay), m.CurrentTrack)
			if effectiveDecayValue == -1 {
				statusMsg = "Decay: -- (sticky)"
			} else {
				decaySeconds := types.DecayToSeconds(effectiveDecayValue)
				statusMsg = fmt.Sprintf("Decay: -- (%.3fs, sticky)", decaySeconds)
			}
		} else {
			decaySeconds := types.DecayToSeconds(decayValue)
			statusMsg = fmt.Sprintf("Decay: %02X (%.3fs, sticky)", decayValue, decaySeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColSustain) { // S column
		// Show Sustain info
		sustainValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSustain]
		if sustainValue == -1 {
			// Check for effective (sticky) Sustain value
			effectiveSustainValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColSustain), m.CurrentTrack)
			if effectiveSustainValue == -1 {
				statusMsg = "Sustain: -- (sticky)"
			} else {
				sustainLevel := types.SustainToLevel(effectiveSustainValue)
				statusMsg = fmt.Sprintf("Sustain: -- (%.3f, sticky)", sustainLevel)
			}
		} else {
			sustainLevel := types.SustainToLevel(sustainValue)
			statusMsg = fmt.Sprintf("Sustain: %02X (%.3f, sticky)", sustainValue, sustainLevel)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColRelease) { // R column
		// Show Release info
		releaseValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColRelease]
		if releaseValue == -1 {
			// Check for effective (sticky) Release value
			effectiveReleaseValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColRelease), m.CurrentTrack)
			if effectiveReleaseValue == -1 {
				statusMsg = "Release: -- (sticky)"
			} else {
				releaseSeconds := types.ReleaseToSeconds(effectiveReleaseValue)
				statusMsg = fmt.Sprintf("Release: -- (%.3fs, sticky)", releaseSeconds)
			}
		} else {
			releaseSeconds := types.ReleaseToSeconds(releaseValue)
			statusMsg = fmt.Sprintf("Release: %02X (%.3fs, sticky)", releaseValue, releaseSeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColArpeggio) { // AR column
		// Show Arpeggio info
		arpeggioValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColArpeggio]
		if arpeggioValue == -1 {
			statusMsg = "Arpeggio: -- (not assigned)"
		} else {
			statusMsg = fmt.Sprintf("Arpeggio: %02X", arpeggioValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColMidi) { // MI column
		// Show MIDI info with sticky behavior
		midiValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColMidi]
		if midiValue == -1 {
			// Check for effective (sticky) MIDI value
			effectiveMidiValue := input.GetEffectiveMidiValueForTrack(m, m.CurrentPhrase, m.CurrentRow, m.CurrentTrack)
			if effectiveMidiValue == -1 {
				statusMsg = "MIDI: -- (sticky)"
			} else {
				statusMsg = fmt.Sprintf("MIDI: -- (%02X sticky)", effectiveMidiValue)
			}
		} else {
			statusMsg = fmt.Sprintf("MIDI: %02X (sticky)", midiValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColSoundMaker) { // SO column
		// Show SoundMaker info with sticky behavior
		soundMakerValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSoundMaker]
		if soundMakerValue == -1 {
			// Check for effective (sticky) SoundMaker value
			effectiveSoundMakerValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColSoundMaker), m.CurrentTrack)
			if effectiveSoundMakerValue == -1 {
				statusMsg = "SoundMaker: -- (sticky)"
			} else {
				statusMsg = fmt.Sprintf("SoundMaker: -- (%02X sticky)", effectiveSoundMakerValue)
			}
		} else {
			statusMsg = fmt.Sprintf("SoundMaker: %02X (sticky)", soundMakerValue)
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

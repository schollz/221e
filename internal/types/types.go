package types

import "math"

type ViewMode int

const (
	SongView ViewMode = iota
	ChainView
	PhraseView
	FileView
	SettingsView
	FileMetadataView
	RetriggerView
	TimestrechView
	MixerView
	ArpeggioView
	MidiView
	SoundMakerView
)

type PhraseViewType int

const (
	SamplerPhraseView PhraseViewType = iota
	InstrumentPhraseView
)

type CellType int

const (
	HexCell CellType = iota
	FilenameCell
)

type ClipboardMode int

const (
	CellMode ClipboardMode = iota
	RowMode
)

type PhraseColumn int

const (
	ColNote               PhraseColumn = iota // Column 0: Note value (hex)
	ColPitch                                  // Column 1: Pitch value (hex, default 80/0x80 = 0.0 pitch)
	ColDeltaTime                              // Column 2: Delta time (hex) - controls playback: >0=play, <=0=skip
	ColGate                                   // Column 3: Gate value (hex, default 80/0x50, sticky)
	ColRetrigger                              // Column 4: Retrigger setting index (hex, 00-FE)
	ColTimestretch                            // Column 5: Timestretch setting index (hex, 00-FE)
	ColEffectReverse                          // Column 6: Я (reverse) 0/1
	ColPan                                    // Column 7: PA (pan) (hex, default 128/0x80, 00-FE maps -1.0 to 1.0)
	ColLowPassFilter                          // Column 8: LP (low pass filter) (hex, default FE/20kHz, 00-FE maps 20Hz to 20kHz exponentially)
	ColHighPassFilter                         // Column 9: HP (high pass filter) (hex, default -1/null, 00-FE maps 20Hz to 20kHz exponentially)
	ColEffectComb                             // Column 10: CO (00-FE)
	ColEffectReverb                           // Column 11: VE (00-FE)
	ColFilename                               // Column 12: Filename index
	ColChord                                  // Column 13: Chord (Instrument view only: "-", "M", "m", "d")
	ColChordAddition                          // Column 14: Chord Addition (Instrument view only: "-", "7", "9", "4")
	ColChordTransposition                     // Column 15: Chord Transposition (Instrument view only: "-", "0"-"F")
	ColArpeggio                               // Column 16: Arpeggio (Instrument view only: 00-FE)
	ColMidi                                   // Column 17: MIDI (Instrument view only: 00-FE, sticky)
	ColSoundMaker                             // Column 18: SoundMaker (Instrument view only: 00-FE, sticky)
	ColAttack                                 // Column 19: Attack (Instrument view only: 00-FE, 0.02-30s exponential, default -1, sticky)
	ColDecay                                  // Column 20: Decay (Instrument view only: 00-FE, 0.0-30.0s linear, default -1, sticky)
	ColSustain                                // Column 21: Sustain (Instrument view only: 00-FE, 0.0-1.0 linear, default -1, sticky)
	ColRelease                                // Column 22: Release (Instrument view only: 00-FE, 0.02-30s exponential, default -1, sticky)
	ColCount                                  // Total number of columns
)

// ChordType represents different chord types for instrument tracks
type ChordType int

const (
	ChordNone      ChordType = iota // "-" (default)
	ChordMajor                      // "M"
	ChordMinor                      // "m"
	ChordDominant                   // "d"
	ChordTypeCount                  // Total number of chord types
)

// ChordAddition represents different chord additions for instrument tracks
type ChordAddition int

const (
	ChordAddNone       ChordAddition = iota // "-" (default)
	ChordAdd7                               // "7"
	ChordAdd9                               // "9"
	ChordAdd4                               // "4"
	ChordAdditionCount                      // Total number of chord additions
)

// ChordTransposition represents different chord transpositions for instrument tracks (hex values)
type ChordTransposition int

const (
	ChordTransNone          ChordTransposition = iota // "-" (default, same as 0)
	ChordTrans0                                       // "0"
	ChordTrans1                                       // "1"
	ChordTrans2                                       // "2"
	ChordTrans3                                       // "3"
	ChordTrans4                                       // "4"
	ChordTrans5                                       // "5"
	ChordTrans6                                       // "6"
	ChordTrans7                                       // "7"
	ChordTrans8                                       // "8"
	ChordTrans9                                       // "9"
	ChordTransA                                       // "A"
	ChordTransB                                       // "B"
	ChordTransC                                       // "C"
	ChordTransD                                       // "D"
	ChordTransE                                       // "E"
	ChordTransF                                       // "F"
	ChordTranspositionCount                           // Total number of chord transpositions
)

// ChordTypeToString converts a ChordType enum to its display string
func ChordTypeToString(chordType ChordType) string {
	switch chordType {
	case ChordNone:
		return "-"
	case ChordMajor:
		return "M"
	case ChordMinor:
		return "m"
	case ChordDominant:
		return "d"
	default:
		return "-"
	}
}

// ChordAdditionToString converts a ChordAddition enum to its display string
func ChordAdditionToString(chordAdd ChordAddition) string {
	switch chordAdd {
	case ChordAddNone:
		return "-"
	case ChordAdd7:
		return "7"
	case ChordAdd9:
		return "9"
	case ChordAdd4:
		return "4"
	default:
		return "-"
	}
}

// UI Column positions for Instrument Phrase View - to prevent hardcoding issues
type InstrumentUIColumn int

const (
	InstrumentColSL    InstrumentUIColumn = 0  // SL - Slice (display only)
	InstrumentColDT    InstrumentUIColumn = 1  // DT - Delta Time
	InstrumentColNOT   InstrumentUIColumn = 2  // NOT - Note
	InstrumentColC     InstrumentUIColumn = 3  // C - Chord
	InstrumentColA     InstrumentUIColumn = 4  // A - Chord Addition
	InstrumentColT     InstrumentUIColumn = 5  // T - Chord Transposition
	InstrumentColGT    InstrumentUIColumn = 6  // GT - Gate
	InstrumentColATK   InstrumentUIColumn = 7  // A - Attack
	InstrumentColDECAY InstrumentUIColumn = 8  // D - Decay
	InstrumentColSUS   InstrumentUIColumn = 9  // S - Sustain
	InstrumentColREL   InstrumentUIColumn = 10 // R - Release
	InstrumentColRE    InstrumentUIColumn = 11 // RE - Reverb
	InstrumentColCO    InstrumentUIColumn = 12 // CO - Comb
	InstrumentColPA    InstrumentUIColumn = 13 // PA - Pan
	InstrumentColLP    InstrumentUIColumn = 14 // LP - LowPass
	InstrumentColHP    InstrumentUIColumn = 15 // HP - HighPass
	InstrumentColAR    InstrumentUIColumn = 16 // AR - Arpeggio
	InstrumentColMI    InstrumentUIColumn = 17 // MI - MIDI
	InstrumentColSO    InstrumentUIColumn = 18 // SO - SoundMaker
)

// UI Column positions for Sampler Phrase View - to prevent hardcoding issues
type SamplerUIColumn int

const (
	SamplerColSL  SamplerUIColumn = 0  // SL - Slice (display only)
	SamplerColDT  SamplerUIColumn = 1  // DT - Delta Time
	SamplerColNN  SamplerUIColumn = 2  // NN - Note
	SamplerColPI  SamplerUIColumn = 3  // PI - Pitch
	SamplerColGT  SamplerUIColumn = 4  // GT - Gate
	SamplerColRT  SamplerUIColumn = 5  // RT - Retrigger
	SamplerColTS  SamplerUIColumn = 6  // TS - Timestretch
	SamplerColREV SamplerUIColumn = 7  // Я - Reverse
	SamplerColPA  SamplerUIColumn = 8  // PA - Pan
	SamplerColLP  SamplerUIColumn = 9  // LP - Low Pass Filter
	SamplerColHP  SamplerUIColumn = 10 // HP - High Pass Filter
	SamplerColCO  SamplerUIColumn = 11 // CO - Comb
	SamplerColVE  SamplerUIColumn = 12 // VE - Reverb
	SamplerColFI  SamplerUIColumn = 13 // FI - Filename
)

// UI Column positions for Arpeggio View - to prevent hardcoding issues
type ArpeggioUIColumn int

const (
	ArpeggioColDI  ArpeggioUIColumn = 0 // DI - Direction
	ArpeggioColCO  ArpeggioUIColumn = 1 // CO - Count
	ArpeggioColDIV ArpeggioUIColumn = 2 // Divisor
)

// ChordTranspositionToString converts a ChordTransposition enum to its display string
func ChordTranspositionToString(chordTrans ChordTransposition) string {
	switch chordTrans {
	case ChordTransNone:
		return "-"
	case ChordTrans0:
		return "0"
	case ChordTrans1:
		return "1"
	case ChordTrans2:
		return "2"
	case ChordTrans3:
		return "3"
	case ChordTrans4:
		return "4"
	case ChordTrans5:
		return "5"
	case ChordTrans6:
		return "6"
	case ChordTrans7:
		return "7"
	case ChordTrans8:
		return "8"
	case ChordTrans9:
		return "9"
	case ChordTransA:
		return "A"
	case ChordTransB:
		return "B"
	case ChordTransC:
		return "C"
	case ChordTransD:
		return "D"
	case ChordTransE:
		return "E"
	case ChordTransF:
		return "F"
	default:
		return "-"
	}
}

func GetChordNotes(root int, ctype ChordType, add ChordAddition, transpose ChordTransposition) []int {
	notes := []int{root}

	if ctype == ChordNone {
		return notes
	}

	switch ctype {
	case ChordMajor:
		notes = append(notes, root+4, root+7)
	case ChordMinor:
		notes = append(notes, root+3, root+7)
	case ChordDominant:
		notes = append(notes, root+4, root+7) // dominant is like major triad, add7 handled below
	}

	switch add {
	case ChordAdd7:
		if ctype == ChordMinor {
			notes = append(notes, root+10) // minor 7th
		} else {
			notes = append(notes, root+11) // major 7th
		}
	case ChordAdd9:
		notes = append(notes, root+14) // 9th = 2nd + octave
	case ChordAdd4:
		notes = append(notes, root+5) // 4th
	}

	if transpose > ChordTrans0 {
		for i := ChordTrans0; i < transpose; i++ {
			first := notes[0]
			notes = notes[1:]
			notes = append(notes, first+12)
		}
	}

	return notes
}

type FileMetadata struct {
	BPM    float32 `json:"bpm"`    // Source BPM for the file
	Slices int     `json:"slices"` // Number of slices in the file
}

type RetriggerSettings struct {
	Times       int     `json:"times"`       // Number of retriggers (0-256)
	Start       float32 `json:"start"`       // Starting rate (0-256, 0.05 increments) /beat
	End         float32 `json:"end"`         // Final rate (0-256, 0.05 increments, default 0) /beat
	Beats       int     `json:"beats"`       // Beats value (0-256)
	VolumeDB    float32 `json:"volumeDB"`    // Volume change (-16 to +16 dB, default 0) - NOTE: This is for retrigger-specific volume, not global
	PitchChange float32 `json:"pitchChange"` // Pitch change (-24 to +24, default 0)
}

type TimestrechSettings struct {
	Start float32 `json:"start"` // Start value (0-256, 0.05 increments)
	End   float32 `json:"end"`   // End value (0-256, 0.05 increments, default 0)
	Beats int     `json:"beats"` // Beats value (0-256)
}

type ArpeggioRow struct {
	Direction int `json:"direction"` // Direction: 0="--", 1="u-", 2="d-"
	Count     int `json:"count"`     // Count: -1="--", 0-254 for hex values 00-FE
	Divisor   int `json:"divisor"`   // Divisor: -1="--", 1-254 for hex values 01-FE
}

type ArpeggioSettings struct {
	Rows [16]ArpeggioRow `json:"rows"` // 16 rows (00-0F), each with its own DI and CO
}

type MidiSettings struct {
	Device  string `json:"device"`  // MIDI Device name
	Channel string `json:"channel"` // MIDI Channel (1-16 or "all")
}

type SoundMakerSettings struct {
	Name   string `json:"name"`   // SoundMaker name ("Polyperc", "Infinite Pad", "DX7", etc.)
	A      int    `json:"a"`      // Parameter A (00-FE, -1 for "--") - used for Polyperc, Infinite Pad
	B      int    `json:"b"`      // Parameter B (00-FE, -1 for "--") - used for Polyperc, Infinite Pad
	C      int    `json:"c"`      // Parameter C (00-FE, -1 for "--") - used for Polyperc, Infinite Pad
	D      int    `json:"d"`      // Parameter D (00-FE, -1 for "--") - used for Polyperc, Infinite Pad
	Preset int    `json:"preset"` // Preset number (0-1000, -1 for "--") - used for DX7
}

type ClipboardData struct {
	// Cell data
	Value    int
	CellType CellType
	// Row data
	RowData     []int
	RowFilename string
	SourceView  ViewMode
	// Arpeggio row data
	ArpeggioRowData struct {
		Direction []int
		Count     []int
		Divisor   []int
	}
	// Highlighting
	HighlightRow    int
	HighlightCol    int
	HighlightPhrase int
	HighlightView   ViewMode
	// Common
	Mode    ClipboardMode
	HasData bool
}

type SaveData struct {
	ViewMode      ViewMode     `json:"viewMode"`
	CurrentRow    int          `json:"currentRow"`
	CurrentCol    int          `json:"currentCol"`
	ScrollOffset  int          `json:"scrollOffset"`
	CurrentPhrase int          `json:"currentPhrase"`
	FileSelectRow int          `json:"fileSelectRow"`
	FileSelectCol int          `json:"fileSelectCol"`
	ChainsData    [][]int      `json:"chainsData"`
	PhrasesData   [255][][]int `json:"phrasesData"`
	// New separate data pools for Instruments and Samplers
	InstrumentChainsData  [][]int                 `json:"instrumentChainsData"`
	InstrumentPhrasesData [255][][]int            `json:"instrumentPhrasesData"`
	SamplerChainsData     [][]int                 `json:"samplerChainsData"`
	SamplerPhrasesData    [255][][]int            `json:"samplerPhrasesData"`
	SamplerPhrasesFiles   []string                `json:"samplerPhrasesFiles"`
	LastEditRow           int                     `json:"lastEditRow"`
	PhrasesFiles          []string                `json:"phrasesFiles"`
	CurrentDir            string                  `json:"currentDir"`
	BPM                   float32                 `json:"bpm"`
	PPQ                   int                     `json:"ppq"`
	PregainDB             float32                 `json:"pregainDB"`
	PostgainDB            float32                 `json:"postgainDB"`
	BiasDB                float32                 `json:"biasDB"`
	SaturationDB          float32                 `json:"saturationDB"`
	DriveDB               float32                 `json:"driveDB"`
	FileMetadata          map[string]FileMetadata `json:"fileMetadata"`
	LastChainRow          int                     `json:"lastChainRow"`
	LastPhraseRow         int                     `json:"lastPhraseRow"`
	RecordingEnabled      bool                    `json:"recordingEnabled"`
	RetriggerSettings     [255]RetriggerSettings  `json:"retriggerSettings"`
	TimestrechSettings    [255]TimestrechSettings `json:"timestrechSettings"`
	ArpeggioSettings      [255]ArpeggioSettings   `json:"arpeggioSettings"`
	MidiSettings          [255]MidiSettings       `json:"midiSettings"`
	SoundMakerSettings    [255]SoundMakerSettings `json:"soundMakerSettings"`
	SongData              [8][16]int              `json:"songData"`
	LastSongRow           int                     `json:"lastSongRow"`
	LastSongTrack         int                     `json:"lastSongTrack"`
	CurrentChain          int                     `json:"currentChain"`
	CurrentTrack          int                     `json:"currentTrack"`
	TrackSetLevels        [8]float32              `json:"trackSetLevels"`
	TrackTypes            [8]bool                 `json:"trackTypes"`
	CurrentMixerTrack     int                     `json:"currentMixerTrack"`
}

const SaveFile = "tracker-save.json"
const WaveformHeight = 5

// ADSR mapping functions for Instrument view

// AttackToSeconds converts Attack hex value (00-FE) to seconds using exponential mapping
// 00 maps to 0.02s, FE maps to 30s
func AttackToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.02 // Default minimum value
	}
	// Exponential mapping: 00 -> 0.02s, FE -> 30s
	minSeconds := float32(0.02)
	maxSeconds := float32(30.0)
	ratio := float32(hexValue) / 254.0
	return minSeconds * float32(math.Pow(float64(maxSeconds/minSeconds), float64(ratio)))
}

// DecayToSeconds converts Decay hex value (00-FE) to seconds using linear mapping
// 00 maps to 0.0s, FE maps to 30.0s
func DecayToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.0 // Default minimum value
	}
	// Linear mapping: 00 -> 0.0s, FE -> 30.0s
	return (float32(hexValue) / 254.0) * 30.0
}

// SustainToLevel converts Sustain hex value (00-FE) to level using linear mapping
// 00 maps to 0.0, FE maps to 1.0
func SustainToLevel(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.0 // Default minimum value
	}
	// Linear mapping: 00 -> 0.0, FE -> 1.0
	return float32(hexValue) / 254.0
}

// ReleaseToSeconds converts Release hex value (00-FE) to seconds using exponential mapping
// 00 maps to 0.02s, FE maps to 30s (same as Attack)
func ReleaseToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.02 // Default minimum value
	}
	// Exponential mapping: 00 -> 0.02s, FE -> 30s
	minSeconds := float32(0.02)
	maxSeconds := float32(30.0)
	ratio := float32(hexValue) / 254.0
	return minSeconds * float32(math.Pow(float64(maxSeconds/minSeconds), float64(ratio)))
}

// VirtualDefaultConfig holds virtual default value for columns that display "--" but behave as a specific value
type VirtualDefaultConfig struct {
	DefaultValue int
}

// GetVirtualDefault returns the virtual default config for a column, or nil if no virtual default
func GetVirtualDefault(col PhraseColumn) *VirtualDefaultConfig {
	switch col {
	case ColPitch:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = center pitch
	case ColGate:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = default gate
	case ColPan:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = center pan
	case ColLowPassFilter:
		return &VirtualDefaultConfig{DefaultValue: 0xFE} // 0xFE = 20kHz (no filtering)
	default:
		return nil
	}
}

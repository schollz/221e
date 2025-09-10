package types

import (
	"math"
)

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
	ColVelocity                               // Column 23: Velocity (VE) (00-7F, 0-127)
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
	InstrumentColVE    InstrumentUIColumn = 6  // VE - Velocity
	InstrumentColGT    InstrumentUIColumn = 7  // GT - Gate
	InstrumentColATK   InstrumentUIColumn = 8  // A - Attack
	InstrumentColDECAY InstrumentUIColumn = 9  // D - Decay
	InstrumentColSUS   InstrumentUIColumn = 10 // S - Sustain
	InstrumentColREL   InstrumentUIColumn = 11 // R - Release
	InstrumentColRE    InstrumentUIColumn = 12 // RE - Reverb
	InstrumentColCO    InstrumentUIColumn = 13 // CO - Comb
	InstrumentColPA    InstrumentUIColumn = 14 // PA - Pan
	InstrumentColLP    InstrumentUIColumn = 15 // LP - LowPass
	InstrumentColHP    InstrumentUIColumn = 16 // HP - HighPass
	InstrumentColAR    InstrumentUIColumn = 17 // AR - Arpeggio
	InstrumentColMI    InstrumentUIColumn = 18 // MI - MIDI
	InstrumentColSO    InstrumentUIColumn = 19 // SO - SoundMaker
)

// UI Column positions for Sampler Phrase View - to prevent hardcoding issues
type SamplerUIColumn int

const (
	SamplerColSL  SamplerUIColumn = 0  // SL - Slice (display only)
	SamplerColDT  SamplerUIColumn = 1  // DT - Delta Time
	SamplerColNN  SamplerUIColumn = 2  // NN - Note
	SamplerColVE  SamplerUIColumn = 3  // VE - Velocity
	SamplerColPI  SamplerUIColumn = 4  // PI - Pitch
	SamplerColGT  SamplerUIColumn = 5  // GT - Gate
	SamplerColRT  SamplerUIColumn = 6  // RT - Retrigger
	SamplerColTS  SamplerUIColumn = 7  // TS - Timestretch
	SamplerColREV SamplerUIColumn = 8  // Я - Reverse
	SamplerColPA  SamplerUIColumn = 9  // PA - Pan
	SamplerColLP  SamplerUIColumn = 10 // LP - Low Pass Filter
	SamplerColHP  SamplerUIColumn = 11 // HP - High Pass Filter
	SamplerColCO  SamplerUIColumn = 12 // CO - Comb
	SamplerColRE  SamplerUIColumn = 13 // RE - Reverb
	SamplerColFI  SamplerUIColumn = 14 // FI - Filename
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
	Times              int     `json:"times"`              // Number of retriggers (0-256)
	Start              float32 `json:"start"`              // Starting rate (0-256, 0.05 increments) /beat
	End                float32 `json:"end"`                // Final rate (0-256, 0.05 increments, default 0) /beat
	Beats              int     `json:"beats"`              // Beats value (0-256)
	VolumeDB           float32 `json:"volumeDB"`           // Volume change (-16 to +16 dB, default 0) - NOTE: This is for retrigger-specific volume, not global
	PitchChange        float32 `json:"pitchChange"`        // Pitch change (-24 to +24, default 0)
	FinalPitchToStart  int     `json:"finalPitchToStart"`  // Final pitch to start: 0=No, 1=Yes (default 0)
	FinalVolumeToStart int     `json:"finalVolumeToStart"` // Final volume to start: 0=No, 1=Yes (default 0)
}

type TimestrechSettings struct {
	Start float32 `json:"start"` // Start value (0-256, 0.05 increments)
	End   float32 `json:"end"`   // End value (0-256, 0.05 increments, default 0)
	Beats int     `json:"beats"` // Beats value (0-256)
}

// ArpeggioDirection represents different arpeggio directions
type ArpeggioDirection int

const (
	ArpeggioDirectionNone ArpeggioDirection = iota // 0: "--"
	ArpeggioDirectionUp                            // 1: "u-"
	ArpeggioDirectionDown                          // 2: "d-"
)

// SoundMakerRow represents different rows in the SoundMaker settings view
type SoundMakerRow int

const (
	SoundMakerRowName   SoundMakerRow = iota // 0: Name row
	SoundMakerRowParamA                      // 1: Parameter A / Preset row
	SoundMakerRowParamB                      // 2: Parameter B row
	SoundMakerRowParamC                      // 3: Parameter C row
	SoundMakerRowParamD                      // 4: Parameter D row
)

// GlobalSettingsRow represents different rows in the Global settings column
type GlobalSettingsRow int

const (
	GlobalSettingsRowBPM          GlobalSettingsRow = iota // 0: BPM
	GlobalSettingsRowPPQ                                   // 1: PPQ
	GlobalSettingsRowPregainDB                             // 2: PregainDB
	GlobalSettingsRowPostgainDB                            // 3: PostgainDB
	GlobalSettingsRowBiasDB                                // 4: BiasDB
	GlobalSettingsRowSaturationDB                          // 5: SaturationDB
	GlobalSettingsRowDriveDB                               // 6: DriveDB
)

// InputSettingsRow represents different rows in the Input settings column
type InputSettingsRow int

const (
	InputSettingsRowInputLevelDB      InputSettingsRow = iota // 0: InputLevelDB
	InputSettingsRowReverbSendPercent                         // 1: ReverbSendPercent
)

// BrailleDotRow represents different rows in a 2x4 Braille cell
type BrailleDotRow int

const (
	BrailleDotRow0 BrailleDotRow = iota // 0: Top row
	BrailleDotRow1                      // 1: Second row
	BrailleDotRow2                      // 2: Third row
	BrailleDotRow3                      // 3: Bottom row
)

// FileMetadataRow represents different rows in the file metadata view
type FileMetadataRow int

const (
	FileMetadataRowBPM    FileMetadataRow = iota // 0: BPM
	FileMetadataRowSlices                        // 1: Slices
)

// MidiSettingsRow represents different rows in the MIDI settings view
type MidiSettingsRow int

const (
	MidiSettingsRowDevice  MidiSettingsRow = iota // 0: MIDI Device
	MidiSettingsRowChannel                        // 1: MIDI Channel
)

// RetriggerSettingsRow represents different rows in the retrigger settings view
type RetriggerSettingsRow int

const (
	RetriggerSettingsRowTimes              RetriggerSettingsRow = iota // 0: Times
	RetriggerSettingsRowStartingRate                                   // 1: Starting Rate
	RetriggerSettingsRowFinalRate                                      // 2: Final Rate
	RetriggerSettingsRowBeats                                          // 3: Beats
	RetriggerSettingsRowVolume                                         // 4: Volume
	RetriggerSettingsRowPitch                                          // 5: Pitch
	RetriggerSettingsRowFinalPitchToStart                              // 6: FinalPitchToStart
	RetriggerSettingsRowFinalVolumeToStart                             // 7: FinalVolumeToStart
)

// TimestrechSettingsRow represents different rows in the timestrech settings view
type TimestrechSettingsRow int

const (
	TimestrechSettingsRowStart TimestrechSettingsRow = iota // 0: Start
	TimestrechSettingsRowEnd                                // 1: End
	TimestrechSettingsRowBeats                              // 2: Beats
)

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
	Name       string         `json:"name"`       // SoundMaker name ("PolyPerc", "Infinite Pad", "DX7", etc.)
	Parameters map[string]int `json:"parameters"` // Key-value pairs for parameters (e.g. "preset": 5, "A": 128)
	PatchName  string         `json:"patchName"`  // Patch name (used for DX7 when setting by name)
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
	// Flag to indicate this value is a fresh deep copy that shouldn't be deep copied again
	IsFreshDeepCopy bool
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
	LastPhraseCol         int                     `json:"lastPhraseCol"`
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
	TrackSetLevels        [9]float32              `json:"trackSetLevels"`
	TrackTypes            [9]bool                 `json:"trackTypes"`
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
	case ColVelocity:
		return &VirtualDefaultConfig{DefaultValue: 0x40} // 0x40 = 64 = default velocity
	default:
		return nil
	}
}

// Instrument Parameter Framework
type InstrumentParameterType int

const (
	ParameterTypeHex   InstrumentParameterType = iota // 0x00-0xFE hex values (default)
	ParameterTypeInt                                  // Integer values with custom range
	ParameterTypeFloat                                // Float values with custom range
)

type InstrumentParameterDef struct {
	Key          string                  `json:"key"`          // Key for OSC sending (e.g. "preset", "cutoff")
	DisplayName  string                  `json:"displayName"`  // Name shown in UI (e.g. "Preset", "Cutoff")
	Type         InstrumentParameterType `json:"type"`         // Data type
	MinValue     int                     `json:"minValue"`     // Minimum value
	MaxValue     int                     `json:"maxValue"`     // Maximum value
	DefaultValue int                     `json:"defaultValue"` // Default value (-1 for "--")
	Column       int                     `json:"column"`       // Which column to display in (0 or 1)
	Order        int                     `json:"order"`        // Order within the column
}

type InstrumentDefinition struct {
	Name        string                   `json:"name"`        // Instrument name (e.g. "DX7", "PolyPerc")
	Description string                   `json:"description"` // Short description of the instrument
	Parameters  []InstrumentParameterDef `json:"parameters"`  // Parameter definitions
}

// Global registry of all instrument definitions
var InstrumentRegistry = map[string]InstrumentDefinition{
	"DX7": {
		Name:        "DX7",
		Description: "Classic FM synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "preset", DisplayName: "Preset", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 0,
			},
		},
	},
	"PolyPerc": {
		Name:        "PolyPerc",
		Description: "Polyphonic percussion synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "A", DisplayName: "A", Type: ParameterTypeHex,
				MinValue: 0, MaxValue: 254, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "B", DisplayName: "B", Type: ParameterTypeHex,
				MinValue: 0, MaxValue: 254, DefaultValue: -1, Column: 0, Order: 1,
			},
			{
				Key: "C", DisplayName: "C", Type: ParameterTypeHex,
				MinValue: 0, MaxValue: 254, DefaultValue: -1, Column: 1, Order: 0,
			},
			{
				Key: "D", DisplayName: "D", Type: ParameterTypeHex,
				MinValue: 0, MaxValue: 254, DefaultValue: -1, Column: 1, Order: 1,
			},
		},
	},
	"MiBraids": {
		Name:        "MiBraids",
		Description: "MiBraids is a digital macro oscillator that offers an atlas of waveform generation techniques.",
		Parameters: []InstrumentParameterDef{
			{
				Key: "timbre", DisplayName: "Timbre", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "color", DisplayName: "Color", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 1,
			},
			{
				Key: "model", DisplayName: "Model", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 47, DefaultValue: -1, Column: 0, Order: 2,
			},
			{
				Key: "resamp", DisplayName: "Resamp", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 2, DefaultValue: -1, Column: 1, Order: 0,
			},
			{
				Key: "decim", DisplayName: "Decim", Type: ParameterTypeInt,
				MinValue: 1, MaxValue: 32, DefaultValue: -1, Column: 1, Order: 1,
			},
			{
				Key: "bits", DisplayName: "Bits", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 6, DefaultValue: -1, Column: 1, Order: 2,
			},
			{
				Key: "ws", DisplayName: "WS", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 3,
			},
		},
	},
	"MiPlaits": {
		Name:        "MiPlaits",
		Description: "A macro oscillator offering a multitude of synthesis methods.",
		Parameters: []InstrumentParameterDef{
			{
				Key: "engine", DisplayName: "Engine", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 15, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "harm", DisplayName: "Harm", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 1,
			},
			{
				Key: "timbre", DisplayName: "Timbre", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 2,
			},
			{
				Key: "morph", DisplayName: "Morph", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 3,
			},
			{
				Key: "fm_mod", DisplayName: "FM Mod", Type: ParameterTypeFloat,
				MinValue: -1000, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 0,
			},
			{
				Key: "timb_mod", DisplayName: "Timb Mod", Type: ParameterTypeFloat,
				MinValue: -1000, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 1,
			},
			{
				Key: "morph_mod", DisplayName: "Morph Mod", Type: ParameterTypeFloat,
				MinValue: -1000, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 2,
			},
			{
				Key: "decay", DisplayName: "Decay", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 3,
			},
			{
				Key: "lpg_colour", DisplayName: "LPG Color", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 4,
			},
		},
	},
}

// Helper functions for the instrument framework

// GetInstrumentDefinition returns the definition for a given instrument name
func GetInstrumentDefinition(name string) (InstrumentDefinition, bool) {
	def, exists := InstrumentRegistry[name]
	return def, exists
}

// GetInstrumentParameterByKey returns a specific parameter definition by key
func (def InstrumentDefinition) GetParameterByKey(key string) (InstrumentParameterDef, bool) {
	for _, param := range def.Parameters {
		if param.Key == key {
			return param, true
		}
	}
	return InstrumentParameterDef{}, false
}

// GetParametersSortedByColumn returns parameters sorted by column and order
func (def InstrumentDefinition) GetParametersSortedByColumn() (col0 []InstrumentParameterDef, col1 []InstrumentParameterDef) {
	for _, param := range def.Parameters {
		if param.Column == 0 {
			col0 = append(col0, param)
		} else {
			col1 = append(col1, param)
		}
	}

	// Sort by order within each column
	for i := 0; i < len(col0)-1; i++ {
		for j := i + 1; j < len(col0); j++ {
			if col0[i].Order > col0[j].Order {
				col0[i], col0[j] = col0[j], col0[i]
			}
		}
	}

	for i := 0; i < len(col1)-1; i++ {
		for j := i + 1; j < len(col1); j++ {
			if col1[i].Order > col1[j].Order {
				col1[i], col1[j] = col1[j], col1[i]
			}
		}
	}

	return col0, col1
}

// MiBraids model names for display
var MiBraidsModelNames = []string{
	"CSAW", "MORPH", "SAW_SQUARE", "SINE_TRIANGLE", "BUZZ", "SQUARE_SUB", "SAW_SUB", "SQUARE_SYNC",
	"SAW_SYNC", "TRIPLE_SAW", "TRIPLE_SQUARE", "TRIPLE_TRIANGLE", "TRIPLE_SINE", "TRIPLE_RING_MOD",
	"SAW_SWARM", "SAW_COMB", "TOY", "DIGITAL_FILTER_LP", "DIGITAL_FILTER_PK", "DIGITAL_FILTER_BP",
	"DIGITAL_FILTER_HP", "VOSIM", "VOWEL", "VOWEL_FOF", "HARMONICS", "FM", "FEEDBACK_FM",
	"CHAOTIC_FEEDBACK_FM", "PLUCKED", "BOWED", "BLOWN", "FLUTED", "STRUCK_BELL", "STRUCK_DRUM",
	"KICK", "CYMBAL", "SNARE", "WAVETABLES", "WAVE_MAP", "WAVE_LINE", "WAVE_PARAPHONIC",
	"FILTERED_NOISE", "TWIN_PEAKS_NOISE", "CLOCKED_NOISE", "GRANULAR_CLOUD", "PARTICLE_NOISE",
	"DIGITAL_MODULATION", "QUESTION_MARK",
}

// GetMiBraidsModelName returns the name for a given model index
func GetMiBraidsModelName(index int) string {
	if index >= 0 && index < len(MiBraidsModelNames) {
		return MiBraidsModelNames[index]
	}
	return "UNKNOWN"
}

// MiPlaits engine names for display
var MiPlaitsEngineNames = []string{
	"virtual_analog_engine", "waveshaping_engine", "fm_engine", "grain_engine",
	"additive_engine", "wavetable_engine", "chord_engine", "speech_engine",
	"swarm_engine", "noise_engine", "particle_engine", "string_engine",
	"modal_engine", "bass_drum_engine", "snare_drum_engine", "hi_hat_engine",
}

// GetMiPlaitsEngineName returns the name for a given engine index
func GetMiPlaitsEngineName(index int) string {
	if index >= 0 && index < len(MiPlaitsEngineNames) {
		return MiPlaitsEngineNames[index]
	}
	return "UNKNOWN"
}

// Helper functions for SoundMakerSettings with the new parameter framework

// GetParameterValue gets a parameter value with fallback to default
func (settings *SoundMakerSettings) GetParameterValue(key string) int {
	if settings.Parameters == nil {
		return -1
	}

	if value, exists := settings.Parameters[key]; exists {
		return value
	}

	// Return default value from instrument definition if available
	if def, exists := GetInstrumentDefinition(settings.Name); exists {
		if param, found := def.GetParameterByKey(key); found {
			return param.DefaultValue
		}
	}

	return -1
}

// SetParameterValue sets a parameter value
func (settings *SoundMakerSettings) SetParameterValue(key string, value int) {
	if settings.Parameters == nil {
		settings.Parameters = make(map[string]int)
	}
	settings.Parameters[key] = value
}

// InitializeParameters ensures all parameters exist with default values
func (settings *SoundMakerSettings) InitializeParameters() {
	if def, exists := GetInstrumentDefinition(settings.Name); exists {
		if settings.Parameters == nil {
			settings.Parameters = make(map[string]int)
		}

		for _, param := range def.Parameters {
			if _, exists := settings.Parameters[param.Key]; !exists {
				settings.Parameters[param.Key] = param.DefaultValue
			}
		}
	}
}

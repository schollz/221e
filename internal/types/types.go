package types

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
	ColPlayback       PhraseColumn = iota // Column 0: Playback flag (0 or 1)
	ColNote                               // Column 1: Note value (hex)
	ColPitch                              // Column 2: Pitch value (hex, default 80/0x80 = 0.0 pitch)
	ColDeltaTime                          // Column 3: Delta time (hex)
	ColGate                               // Column 4: Gate value (hex, default 128/0x80)
	ColRetrigger                          // Column 5: Retrigger setting index (hex, 00-FE)
	ColTimestretch                        // Column 6: Timestretch setting index (hex, 00-FE)
	ColEffectReverse                      // Column 7: Я (reverse) 0/1
	ColPan                                // Column 8: PA (pan) (hex, default 128/0x80, 00-FE maps -1.0 to 1.0)
	ColLowPassFilter                      // Column 9: LP (low pass filter) (hex, default -1/null, 00-FE maps 20kHz to 20Hz exponentially)
	ColHighPassFilter                     // Column 10: HP (high pass filter) (hex, default -1/null, 00-FE maps 20Hz to 20kHz exponentially)
	ColEffectComb                         // Column 11: CO (00-FE)
	ColEffectReverb                       // Column 12: VE (00-FE)
	ColFilename                           // Column 13: Filename index
	ColChord                              // Column 14: Chord (Instrument view only: "-", "M", "m", "d")
	ColChordAddition                      // Column 15: Chord Addition (Instrument view only: "-", "7", "9", "4")
	ColChordTransposition                 // Column 16: Chord Transposition (Instrument view only: "-", "0"-"F")
	ColArpeggio                           // Column 17: Arpeggio (Instrument view only: 00-FE)
	ColCount                              // Total number of columns
)

// ChordType represents different chord types for instrument tracks
type ChordType int

const (
	ChordNone ChordType = iota // "-" (default)
	ChordMajor                 // "M"
	ChordMinor                 // "m"  
	ChordDominant              // "d"
	ChordTypeCount             // Total number of chord types
)

// ChordAddition represents different chord additions for instrument tracks
type ChordAddition int

const (
	ChordAddNone ChordAddition = iota // "-" (default)
	ChordAdd7                         // "7"
	ChordAdd9                         // "9"
	ChordAdd4                         // "4"
	ChordAdditionCount                // Total number of chord additions
)

// ChordTransposition represents different chord transpositions for instrument tracks (hex values)
type ChordTransposition int

const (
	ChordTransNone ChordTransposition = iota // "-" (default, same as 0)
	ChordTrans0                              // "0"
	ChordTrans1                              // "1"
	ChordTrans2                              // "2"
	ChordTrans3                              // "3"
	ChordTrans4                              // "4"
	ChordTrans5                              // "5"
	ChordTrans6                              // "6"
	ChordTrans7                              // "7"
	ChordTrans8                              // "8"
	ChordTrans9                              // "9"
	ChordTransA                              // "A"
	ChordTransB                              // "B"
	ChordTransC                              // "C"
	ChordTransD                              // "D"
	ChordTransE                              // "E"
	ChordTransF                              // "F"
	ChordTranspositionCount                  // Total number of chord transpositions
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
}

type ArpeggioSettings struct {
	Rows [16]ArpeggioRow `json:"rows"` // 16 rows (00-0F), each with its own DI and CO
}

type ClipboardData struct {
	// Cell data
	Value    int
	CellType CellType
	// Row data
	RowData     []int
	RowFilename string
	SourceView  ViewMode
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
	ViewMode           ViewMode                `json:"viewMode"`
	CurrentRow         int                     `json:"currentRow"`
	CurrentCol         int                     `json:"currentCol"`
	ScrollOffset       int                     `json:"scrollOffset"`
	CurrentPhrase      int                     `json:"currentPhrase"`
	FileSelectRow      int                     `json:"fileSelectRow"`
	FileSelectCol      int                     `json:"fileSelectCol"`
	ChainsData         [][]int                 `json:"chainsData"`
	PhrasesData        [255][][]int            `json:"phrasesData"`
	// New separate data pools for Instruments and Samplers
	InstrumentChainsData  [][]int            `json:"instrumentChainsData"`
	InstrumentPhrasesData [255][][]int       `json:"instrumentPhrasesData"`
	SamplerChainsData     [][]int            `json:"samplerChainsData"`
	SamplerPhrasesData    [255][][]int       `json:"samplerPhrasesData"`
	SamplerPhrasesFiles   []string           `json:"samplerPhrasesFiles"`
	LastEditRow        int                     `json:"lastEditRow"`
	PhrasesFiles       []string                `json:"phrasesFiles"`
	CurrentDir         string                  `json:"currentDir"`
	BPM                float32                 `json:"bpm"`
	PPQ                int                     `json:"ppq"`
	PregainDB          float32                 `json:"pregainDB"`
	PostgainDB         float32                 `json:"postgainDB"`
	BiasDB             float32                 `json:"biasDB"`
	SaturationDB       float32                 `json:"saturationDB"`
	DriveDB            float32                 `json:"driveDB"`
	FileMetadata       map[string]FileMetadata `json:"fileMetadata"`
	LastChainRow       int                     `json:"lastChainRow"`
	LastPhraseRow      int                     `json:"lastPhraseRow"`
	RecordingEnabled   bool                    `json:"recordingEnabled"`
	RetriggerSettings  [255]RetriggerSettings  `json:"retriggerSettings"`
	TimestrechSettings [255]TimestrechSettings `json:"timestrechSettings"`
	ArpeggioSettings   [255]ArpeggioSettings   `json:"arpeggioSettings"`
	SongData           [8][16]int              `json:"songData"`
	LastSongRow        int                     `json:"lastSongRow"`
	LastSongTrack      int                     `json:"lastSongTrack"`
	CurrentChain       int                     `json:"currentChain"`
	CurrentTrack       int                     `json:"currentTrack"`
	TrackSetLevels     [8]float32              `json:"trackSetLevels"`
	TrackTypes         [8]bool                 `json:"trackTypes"`
	CurrentMixerTrack  int                     `json:"currentMixerTrack"`
}

const SaveFile = "tracker-save.json"
const WaveformHeight = 5

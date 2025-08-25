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
	ColCount                              // Total number of columns
)

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
	ChainsData         [][]int                 `json:"chainsData"`
	PhrasesData        [255][][]int            `json:"phrasesData"`
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
	SongData           [8][16]int              `json:"songData"`
	LastSongRow        int                     `json:"lastSongRow"`
	LastSongTrack      int                     `json:"lastSongTrack"`
	CurrentChain       int                     `json:"currentChain"`
	CurrentTrack       int                     `json:"currentTrack"`
	TrackSetLevels     [8]float32              `json:"trackSetLevels"`
	CurrentMixerTrack  int                     `json:"currentMixerTrack"`
}

const SaveFile = "tracker-save.json"
const WaveformHeight = 5

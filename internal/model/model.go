package model

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hypebeast/go-osc/osc"

	"github.com/schollz/2n/internal/types"
)

// OSCMessageConfig represents configuration for sending OSC messages
type OSCMessageConfig struct {
	Address    string
	Parameters []interface{}
	LogFormat  string
	LogArgs    []interface{}
}

type Model struct {
	CurrentRow   int
	CurrentCol   int
	ScrollOffset int
	ViewMode     types.ViewMode
	// Legacy shared data structures (will be phased out)
	PhrasesData  [255][][]int // [phrase][row][col] where col uses PhraseColumn enum
	ChainsData   [][]int      // [chain][row] where each chain has 16 rows, each row contains a phrase_number
	PhrasesFiles []string     // [phrase] filename for each phrase row
	// Separate data pools for Instruments (tracks 0-3) and Samplers (tracks 4-7)
	InstrumentPhrasesData [255][][]int        // [phrase][row][col] for instrument tracks - simplified data
	InstrumentChainsData  [][]int             // [chain][row] for instrument tracks
	SamplerPhrasesData    [255][][]int        // [phrase][row][col] for sampler tracks - full complexity
	SamplerChainsData     [][]int             // [chain][row] for sampler tracks
	SamplerPhrasesFiles   []string            // [phrase] filename for sampler phrases only
	CurrentPhrase         int                 // Which phrase we're viewing/editing
	CurrentChain          int                 // Which chain we're viewing/editing
	CurrentTrack          int                 // Which track context we're viewing (0-7)
	FileSelectRow         int                 // Which phrase row we're selecting a file for
	FileSelectCol         int                 // Which phrase column we were on when navigating to file browser
	Clipboard             types.ClipboardData // Cell clipboard
	CurrentDir            string              // Current directory for file browser
	Files                 []string            // Files in current directory
	TermHeight            int
	TermWidth             int
	IsPlaying             bool
	PlaybackRow           int            // Current row within phrase
	PlaybackChain         int            // Current chain being played
	PlaybackChainRow      int            // Current row within chain during playback
	PlaybackPhrase        int            // Current phrase being played
	PlaybackMode          types.ViewMode // Whether playback started from Chain or Phrase view
	ticker                *time.Ticker
	LastEditRow           int            // Track the last row that was edited
	BPM                   float32        // Beats per minute
	PPQ                   int            // Pulses per quarter note
	PregainDB             float32        // Pre-gain in decibels (-96.0 to +32.0, default 0.0)
	PostgainDB            float32        // Post-gain in decibels (-96.0 to +32.0, default 0.0)
	BiasDB                float32        // Bias in decibels (-96.0 to +32.0, default -6.0)
	SaturationDB          float32        // Saturation in decibels (-96.0 to +32.0, default -6.0)
	DriveDB               float32        // Drive in decibels (-96.0 to +32.0, default -6.0)
	PreviousView          types.ViewMode // Track the view we came from when entering Settings
	// Playback state for inheriting values from previous rows
	lastPlaybackNote     int    // Last non-null note value during playback
	lastPlaybackDT       int    // Last non-null deltatime value during playback
	lastPlaybackFileIdx  int    // Last non-null filename index during playback
	lastPlaybackFilename string // Last non-null filename during playback
	// OSC client configuration
	oscClient    *osc.Client
	oscPort      int
	LastWaveform float64   // Last waveform value received from OSC
	WaveformBuf  []float64 // Buffer for waveform data
	// File browser playback state
	CurrentlyPlayingFile string // Track which file is currently playing in file browser
	// File metadata management
	FileMetadata        map[string]types.FileMetadata // Map of filepath -> metadata
	MetadataEditingFile string                        // Currently editing metadata for this file
	// Retrigger settings management
	RetriggerSettings     [255]types.RetriggerSettings // Array of retrigger settings (00-FE)
	RetriggerEditingIndex int                          // Currently editing retrigger index
	// Timestretch settings management
	TimestrechSettings     [255]types.TimestrechSettings // Array of timestretch settings (00-FE)
	TimestrechEditingIndex int                           // Currently editing timestretch index
	// Arpeggio settings management
	ArpeggioSettings     [255]types.ArpeggioSettings // Array of arpeggio settings (00-FE)
	ArpeggioEditingIndex int                         // Currently editing arpeggio index
	// View navigation state
	LastChainRow  int // Last selected row in chain view
	LastPhraseRow int // Last selected row in phrase view
	LastSongRow   int // Last selected row in song view
	LastSongTrack int // Last selected track in song view

	// Song data structure (8 tracks × 16 rows)
	SongData [8][16]int // [track][row] = chain ID (00-FE, -1 for empty)

	// Song playback state
	SongPlaybackRow         [8]int  // Current row for each track during playback
	SongPlaybackActive      [8]bool // Whether each track is actively playing
	SongPlaybackChain       [8]int  // Current chain being played for each track
	SongPlaybackChainRow    [8]int  // Current row within chain for each track
	SongPlaybackPhrase      [8]int  // Current phrase being played for each track
	SongPlaybackRowInPhrase [8]int  // Current row within phrase for each track
	// Save file configuration
	SaveFile string // Path to the save file
	// Recording state
	RecordingEnabled     bool   // Whether recording is queued/enabled
	RecordingActive      bool   // Whether recording is currently active
	CurrentRecordingFile string // Current recording filename
	// Mixer state
	TrackVolumes      [8]float32 // Current volume levels received from SuperCollider (-96 to +12 dB)
	TrackSetLevels    [8]float32 // User-controllable set levels for each track (-96 to +32 dB, default -6.0)
	TrackTypes        [8]bool    // Track type: false = Instrument (IN), true = Sampler (SA), default SA
	CurrentMixerTrack int        // Currently selected track in mixer view (0-7)
	CurrentMixerRow   int        // Current row in mixer: 0 = level (track type now in Song view)
}

// Methods for modifying data structures
func (m *Model) SetChainsData(row, col, value int) {
	if row >= 0 && row < len(m.ChainsData) && col >= 0 && col < len(m.ChainsData[row]) {
		m.ChainsData[row][col] = value
	}
}

func (m *Model) SetPhrasesData(phrase, row, col, value int) {
	if phrase >= 0 && phrase < 255 && row >= 0 && row < len(m.PhrasesData[phrase]) && col >= 0 && col < len(m.PhrasesData[phrase][row]) {
		m.PhrasesData[phrase][row][col] = value
	}
}

func (m *Model) AppendPhrasesFile(filename string) int {
	if m.GetPhraseViewType() == types.InstrumentPhraseView {
		// Instruments don't use files - should not happen
		return -1
	}
	m.SamplerPhrasesFiles = append(m.SamplerPhrasesFiles, filename)
	return len(m.SamplerPhrasesFiles) - 1
}

// GetCurrentPhrasesData returns the appropriate phrases data based on current track
func (m *Model) GetCurrentPhrasesData() *[255][][]int {
	if m.GetPhraseViewType() == types.InstrumentPhraseView {
		return &m.InstrumentPhrasesData
	}
	return &m.SamplerPhrasesData
}

// GetCurrentChainsData returns the appropriate chains data based on current track
func (m *Model) GetCurrentChainsData() *[][]int {
	if m.GetPhraseViewType() == types.InstrumentPhraseView {
		return &m.InstrumentChainsData
	}
	return &m.SamplerChainsData
}

// GetCurrentPhrasesFiles returns the appropriate phrases files based on current track
func (m *Model) GetCurrentPhrasesFiles() *[]string {
	if m.GetPhraseViewType() == types.InstrumentPhraseView {
		// Instruments don't use files - return empty slice
		return nil
	}
	return &m.SamplerPhrasesFiles
}

// GetChainsDataForTrack returns the appropriate chains data based on track type
// Used by Song view to check chain contents across different tracks
func (m *Model) GetChainsDataForTrack(track int) *[][]int {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentChainsData
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerChainsData
}

// ColumnMapping represents the mapping from UI column to data column
type ColumnMapping struct {
	DataColumnIndex int    // Which data column this maps to (types.ColPlayback, types.ColNote, etc.)
	IsEditable      bool   // Whether this column can be edited
	IsCopyable      bool   // Whether this column can be copied
	IsPasteable     bool   // Whether this column can be pasted to
	IsDeletable     bool   // Whether this column can be deleted (backspace)
	DisplayName     string // Name shown in UI
}

// GetColumnMapping returns the column mapping for the current phrase view type
// This centralizes all column mapping logic to prevent inconsistencies
func (m *Model) GetColumnMapping(uiColumn int) *ColumnMapping {
	phraseViewType := m.GetPhraseViewType()

	if phraseViewType == types.InstrumentPhraseView {
		// Instrument view: SL (0), DT (1), NOT (2)
		switch uiColumn {
		case 0: // SL - display only
			return &ColumnMapping{
				DataColumnIndex: -1, // No data column mapping
				IsEditable:      false,
				IsCopyable:      false,
				IsPasteable:     false,
				IsDeletable:     false,
				DisplayName:     "SL",
			}
		case 1: // DT - delta time column (unified playback control)
			return &ColumnMapping{
				DataColumnIndex: int(types.ColDeltaTime),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "DT",
			}
		case 2: // NOT - note column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColNote),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "NOT",
			}
		case 3: // C - chord column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColChord),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "C",
			}
		case 4: // A - chord addition column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColChordAddition),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "A",
			}
		case 5: // T - chord transposition column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColChordTransposition),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "T",
			}
		case 6: // A - attack column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColAttack),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "A",
			}
		case 7: // D - decay column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColDecay),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "D",
			}
		case 8: // S - sustain column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColSustain),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "S",
			}
		case 9: // R - release column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColRelease),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "R",
			}
		case 10: // AR - arpeggio column
			return &ColumnMapping{
				DataColumnIndex: int(types.ColArpeggio),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "AR",
			}
		default:
			return nil // Invalid column
		}
	} else {
		// Sampler view: Custom mapping after removing P column and moving DT to front
		// New order: SL (0), DT (1), NN (2), PI (3), GT (4), RT (5), TS (6), Я (7), PA (8), LP (9), HP (10), CO (11), VE (12), FI (13)
		switch uiColumn {
		case 0: // SL - display only
			return &ColumnMapping{
				DataColumnIndex: -1,
				IsEditable:      false,
				IsCopyable:      false,
				IsPasteable:     false,
				IsDeletable:     false,
				DisplayName:     "SL",
			}
		case 1: // DT - Delta Time (moved from position 4)
			return &ColumnMapping{
				DataColumnIndex: int(types.ColDeltaTime),
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "DT",
			}
		case 2: // NN - Note
			return &ColumnMapping{
				DataColumnIndex: int(types.ColNote), // Now index 0
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "NN",
			}
		case 3: // PI - Pitch
			return &ColumnMapping{
				DataColumnIndex: int(types.ColPitch), // Now index 1
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "PI",
			}
		case 4: // GT - Gate
			return &ColumnMapping{
				DataColumnIndex: int(types.ColGate), // Now index 3
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "GT",
			}
		case 5: // RT - Retrigger
			return &ColumnMapping{
				DataColumnIndex: int(types.ColRetrigger), // Now index 4
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "RT",
			}
		case 6: // TS - Timestretch
			return &ColumnMapping{
				DataColumnIndex: int(types.ColTimestretch), // Now index 5
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "TS",
			}
		case 7: // Я - Reverse
			return &ColumnMapping{
				DataColumnIndex: int(types.ColEffectReverse), // Now index 6
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "Я",
			}
		case 8: // PA - Pan
			return &ColumnMapping{
				DataColumnIndex: int(types.ColPan), // Now index 7
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "PA",
			}
		case 9: // LP - Low Pass Filter
			return &ColumnMapping{
				DataColumnIndex: int(types.ColLowPassFilter), // Now index 8
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "LP",
			}
		case 10: // HP - High Pass Filter
			return &ColumnMapping{
				DataColumnIndex: int(types.ColHighPassFilter), // Now index 9
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "HP",
			}
		case 11: // CO - Comb
			return &ColumnMapping{
				DataColumnIndex: int(types.ColEffectComb), // Now index 10
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "CO",
			}
		case 12: // VE - Reverb
			return &ColumnMapping{
				DataColumnIndex: int(types.ColEffectReverb), // Now index 11
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "VE",
			}
		case 13: // FI - Filename
			return &ColumnMapping{
				DataColumnIndex: int(types.ColFilename), // Now index 12
				IsEditable:      true,
				IsCopyable:      true,
				IsPasteable:     true,
				IsDeletable:     true,
				DisplayName:     "FI",
			}
		default:
			return nil // Invalid column
		}
	}
}

func NewModel(oscPort int, saveFile string) *Model {
	m := &Model{
		CurrentRow:    0,
		CurrentCol:    1,              // Start at phrase column in chain view
		ViewMode:      types.SongView, // Start with song view
		CurrentPhrase: 0,
		LastEditRow:   -1,    // No row edited yet
		BPM:           120.0, // Default BPM
		PPQ:           2,     // Default PPQ
		PregainDB:     0.0,   // Default pregain (0 dB)
		PostgainDB:    0.0,   // Default postgain (0 dB)
		BiasDB:        -6.0,  // Default bias (-6 dB)
		SaturationDB:  -6.0,  // Default saturation (-6 dB)
		DriveDB:       -6.0,  // Default drive (-6 dB)
		// Initialize playback inheritance values
		lastPlaybackNote:     -1,
		lastPlaybackDT:       -1,
		lastPlaybackFileIdx:  -1,
		lastPlaybackFilename: "",
		// Initialize OSC client
		oscPort: oscPort,
		// Initialize file browser playback state
		CurrentlyPlayingFile: "",
		// Initialize file metadata
		FileMetadata:        make(map[string]types.FileMetadata),
		MetadataEditingFile: "",
		// Initialize retrigger settings
		RetriggerEditingIndex: 0,
		// Initialize timestretch settings
		TimestrechEditingIndex: 0,
		// Initialize arpeggio settings
		ArpeggioEditingIndex: 0,
		// Initialize view navigation state
		CurrentChain:  0,
		CurrentTrack:  0,
		LastChainRow:  0,
		LastPhraseRow: 0,
		LastSongRow:   0,
		LastSongTrack: 0,
		// Set save file
		SaveFile: saveFile,
		// Initialize recording state
		RecordingEnabled:     false,
		RecordingActive:      false,
		CurrentRecordingFile: "",
	}

	// Initialize mixer state with defaults
	for i := 0; i < 8; i++ {
		m.TrackVolumes[i] = -96.0  // Start with silence (-96 dB)
		m.TrackSetLevels[i] = -6.0 // Default set level (-6 dB)
		m.TrackTypes[i] = true     // Default to Sampler (SA)
	}
	m.CurrentMixerRow = 0   // Start on level row
	m.CurrentMixerTrack = 0 // Default to track 0

	// Initialize OSC client if port is provided
	if oscPort > 0 {
		m.oscClient = osc.NewClient("localhost", oscPort)
		log.Printf("OSC client initialized for localhost:%d", oscPort)
	}

	// Initialize default data
	m.initializeDefaultData()
	return m
}

func (m *Model) initializeDefaultData() {
	// Initialize chains data (255 chains, each with chain_number and phrase_number)
	m.ChainsData = make([][]int, 255)
	for i := range m.ChainsData {
		m.ChainsData[i] = make([]int, 16) // Each chain has 16 rows for phrases
		for row := range m.ChainsData[i] {
			m.ChainsData[i][row] = -1 // -1 means no phrase assigned
		}
	}

	// Initialize phrases data (255 phrases, each with 255 rows using PhraseColumn enum structure)
	for p := 0; p < 255; p++ {
		m.PhrasesData[p] = make([][]int, 255)
		for i := range m.PhrasesData[p] {
			m.PhrasesData[p][i] = make([]int, int(types.ColCount)) // Use ColCount for array size
			m.PhrasesData[p][i][types.ColNote] = -1                // Note value (-1 means no data "--")
			m.PhrasesData[p][i][types.ColPitch] = 128              // Pitch value (128 = 0x80 = 0.0 pitch, default)
			m.PhrasesData[p][i][types.ColDeltaTime] = -1           // Delta time (-1 means no data "--", controls playback)
			m.PhrasesData[p][i][types.ColGate] = 128               // Gate value (128 = 1.0, default)
			m.PhrasesData[p][i][types.ColRetrigger] = -1           // Retrigger index (-1 means no retrigger)
			m.PhrasesData[p][i][types.ColTimestretch] = -1         // Timestretch index (-1 means no timestretch)
			m.PhrasesData[p][i][types.ColEffectReverse] = -1       // Reverse effect (-1 means no effect)
			m.PhrasesData[p][i][types.ColPan] = -1                 // Pan (-1 = null, will use effective value or default to center)
			m.PhrasesData[p][i][types.ColLowPassFilter] = -1       // Low pass filter (-1 means no filter/20kHz)
			m.PhrasesData[p][i][types.ColHighPassFilter] = -1      // High pass filter (-1 means no filter/20Hz)
			m.PhrasesData[p][i][types.ColEffectComb] = -1          // Comb effect (-1 means no effect)
			m.PhrasesData[p][i][types.ColEffectReverb] = -1        // Reverb effect (-1 means no effect)
			m.PhrasesData[p][i][types.ColFilename] = -1            // Filename index (-1 means no file selected)
		}
	}

	// Initialize phrases files array (start empty, grows as files are added)
	m.PhrasesFiles = make([]string, 0) // phrase*row filename storage

	// Initialize separate data pools for Instruments and Samplers
	// Initialize instrument phrases data (simplified structure - only Note column matters)
	for p := 0; p < 255; p++ {
		m.InstrumentPhrasesData[p] = make([][]int, 255)
		for i := range m.InstrumentPhrasesData[p] {
			m.InstrumentPhrasesData[p][i] = make([]int, int(types.ColCount))
			// For instruments, initialize with minimal defaults
			m.InstrumentPhrasesData[p][i][types.ColNote] = -1        // No note by default
			m.InstrumentPhrasesData[p][i][types.ColDeltaTime] = -1   // DT controls playback for instruments too
			// Initialize chord columns (use int values corresponding to enum defaults)
			m.InstrumentPhrasesData[p][i][types.ColChord] = int(types.ChordNone)                   // Default: "-"
			m.InstrumentPhrasesData[p][i][types.ColChordAddition] = int(types.ChordAddNone)        // Default: "-"
			m.InstrumentPhrasesData[p][i][types.ColChordTransposition] = int(types.ChordTransNone) // Default: "-"
			m.InstrumentPhrasesData[p][i][types.ColArpeggio] = -1                                  // Default: "--" (no arpeggio)
			// Initialize ADSR columns (all sticky, default to undefined)
			m.InstrumentPhrasesData[p][i][types.ColAttack] = -1                                    // Default: "--" (sticky)
			m.InstrumentPhrasesData[p][i][types.ColDecay] = -1                                     // Default: "--" (sticky)
			m.InstrumentPhrasesData[p][i][types.ColSustain] = -1                                   // Default: "--" (sticky)
			m.InstrumentPhrasesData[p][i][types.ColRelease] = -1                                   // Default: "--" (sticky)
			// Other columns can stay -1 (unused for instruments)
		}
	}

	// Initialize sampler phrases data (full complexity - copy from legacy initialization)
	for p := 0; p < 255; p++ {
		m.SamplerPhrasesData[p] = make([][]int, 255)
		for i := range m.SamplerPhrasesData[p] {
			m.SamplerPhrasesData[p][i] = make([]int, int(types.ColCount))
			m.SamplerPhrasesData[p][i][types.ColNote] = -1           // Note value (-1 means no data "--")
			m.SamplerPhrasesData[p][i][types.ColPitch] = 128         // Pitch value (128 = 0x80 = 0.0 pitch, default)
			m.SamplerPhrasesData[p][i][types.ColDeltaTime] = -1      // Delta time (-1 means no data "--", controls playback)
			m.SamplerPhrasesData[p][i][types.ColGate] = 128          // Gate value (128 = 1.0, default)
			m.SamplerPhrasesData[p][i][types.ColRetrigger] = -1      // Retrigger index (-1 means no retrigger)
			m.SamplerPhrasesData[p][i][types.ColTimestretch] = -1    // Timestretch index (-1 means no timestretch)
			m.SamplerPhrasesData[p][i][types.ColEffectReverse] = -1  // Reverse effect (-1 means no effect)
			m.SamplerPhrasesData[p][i][types.ColPan] = -1            // Pan (-1 = null, will use effective value or default to center)
			m.SamplerPhrasesData[p][i][types.ColLowPassFilter] = -1  // Low pass filter (-1 means no filter/20kHz)
			m.SamplerPhrasesData[p][i][types.ColHighPassFilter] = -1 // High pass filter (-1 means no filter/20Hz)
			m.SamplerPhrasesData[p][i][types.ColEffectComb] = -1     // Comb effect (-1 means no effect)
			m.SamplerPhrasesData[p][i][types.ColEffectReverb] = -1   // Reverb effect (-1 means no effect)
			m.SamplerPhrasesData[p][i][types.ColFilename] = -1       // Filename index (-1 means no file selected)
		}
	}

	// Initialize separate chains data
	m.InstrumentChainsData = make([][]int, 255)
	for i := range m.InstrumentChainsData {
		m.InstrumentChainsData[i] = make([]int, 16)
		for j := range m.InstrumentChainsData[i] {
			m.InstrumentChainsData[i][j] = -1
		}
	}

	m.SamplerChainsData = make([][]int, 255)
	for i := range m.SamplerChainsData {
		m.SamplerChainsData[i] = make([]int, 16)
		for j := range m.SamplerChainsData[i] {
			m.SamplerChainsData[i][j] = -1
		}
	}

	// Initialize sampler phrases files array
	m.SamplerPhrasesFiles = make([]string, 0)

	// Initialize retrigger settings with defaults
	for i := 0; i < 255; i++ {
		m.RetriggerSettings[i] = types.RetriggerSettings{
			Times:       0,   // Default times (0)
			Start:       0.0, // Default starting rate
			End:         0.0, // Default final rate
			Beats:       0,   // Default beats
			VolumeDB:    0.0, // Default volume change (0 dB)
			PitchChange: 0.0, // Default pitch change (0 semitones)
		}
	}

	// Initialize timestretch settings with defaults
	for i := 0; i < 255; i++ {
		m.TimestrechSettings[i] = types.TimestrechSettings{
			Start: 0.0, // Default start
			End:   0.0, // Default end
			Beats: 0,   // Default beats
		}
	}

	// Initialize arpeggio settings with defaults
	for i := 0; i < 255; i++ {
		var arpeggioSettings types.ArpeggioSettings
		for row := 0; row < 16; row++ {
			arpeggioSettings.Rows[row] = types.ArpeggioRow{
				Direction: 0,  // Default direction (0 = "--")
				Count:     -1, // Default count (-1 = "--")
				Divisor:   -1, // Default divisor (-1 = "--")
			}
		}
		m.ArpeggioSettings[i] = arpeggioSettings
	}

	// Initialize song data (8 tracks × 16 rows, all empty initially)
	for track := 0; track < 8; track++ {
		for row := 0; row < 16; row++ {
			m.SongData[track][row] = -1 // -1 means no chain assigned
		}
		// Initialize song playback state
		m.SongPlaybackRow[track] = 0
		m.SongPlaybackActive[track] = false
		m.SongPlaybackChain[track] = -1
		m.SongPlaybackChainRow[track] = 0
		m.SongPlaybackPhrase[track] = -1
		m.SongPlaybackRowInPhrase[track] = 0
	}

	// Initialize current directory
	var err error
	m.CurrentDir, err = os.Getwd()
	if err != nil {
		m.CurrentDir = "."
	}
}

func (m *Model) GetVisibleRows() int {
	cellsHigh := (types.WaveformHeight + 1) / 2
	if m.TermHeight == 0 {
		return 20 - cellsHigh
	}
	return m.TermHeight - 5 - cellsHigh
}

// SamplerOSCParams holds parameters for OSC sampler messages
type SamplerOSCParams struct {
	Filename              string  // Path to the audio file
	TrackId               int     // Track ID
	SliceCount            int     // Total number of slices in the file
	SliceNumber           int     // Which slice to trigger (0-based)
	SliceDuration         float32 // Duration multiplier (default 1.0)
	BPMSource             float32 // Source BPM from file metadata
	BPMTarget             float32 // Target BPM from global settings
	Pitch                 float32 // Pitch value (-24 to +24, default 0.0 when hex 80)
	RetriggerNumTotal     int     // Retrigger Settings "Times"
	RetriggerBeats        float32 // Retrigger Settings "Beats"
	RetriggerRateStart    float32 // Retrigger Settings "Starting Rate"
	RetriggerRateEnd      float32 // Retrigger Settings "Final Rate"
	RetriggerPitchChange  float32 // Retrigger Settings "Pitch"
	RetriggerVolumeChange float32 // Retrigger Settings "Volume dB"
	TimestretchStart      float32 // Timestretch Settings "Start"
	TimestretchEnd        float32 // Timestretch Settings "End"
	TimestretchBeats      float32 // Timestretch Settings "Beats"
	EffectReverse         int     // 0 or 1
	Pan                   float32 // -1.0 to 1.0 (pan position)
	LowPassFilter         float32 // Frequency in Hz (20Hz to 20kHz) or -1 for no filter
	HighPassFilter        float32 // Frequency in Hz (20Hz to 20kHz) or -1 for no filter
	EffectComb            float32 // 0.0 .. 1.0
	EffectReverb          float32 // 0.0 .. 1.0
}

type InstrumentOSCParams struct {
	TrackId  int     // Track ID
	MidiNote int     // MIDI note number (0-127)
	Velocity float32 // Note velocity (0.0-1.0)
}

// NewSamplerOSCParams creates sampler parameters with custom slice duration
func NewSamplerOSCParams(filename string, trackId int, sliceCount, sliceNumber int, bpmSource, bpmTarget, sliceDuration float32) SamplerOSCParams {
	return SamplerOSCParams{
		Filename:              filename,
		TrackId:               trackId,
		SliceCount:            sliceCount,
		SliceNumber:           sliceNumber,
		SliceDuration:         sliceDuration,
		BPMSource:             bpmSource,
		BPMTarget:             bpmTarget,
		Pitch:                 0.0, // Default pitch (hex 80 = 0.0 pitch)
		RetriggerNumTotal:     0,
		RetriggerBeats:        0,
		RetriggerRateStart:    0,
		RetriggerRateEnd:      0,
		RetriggerPitchChange:  0,
		RetriggerVolumeChange: 0,
		TimestretchStart:      0,
		TimestretchEnd:        0,
		TimestretchBeats:      0,
		EffectReverse:         0,
		Pan:                   0,     // Default center pan
		LowPassFilter:         20000, // Default no filter (20kHz)
		HighPassFilter:        20,    // Default no filter (20Hz)
		EffectComb:            0,
		EffectReverb:          0,
	}
}

// NewSamplerOSCParamsWithRetrigger creates sampler parameters with retrigger settings
func NewSamplerOSCParamsWithRetrigger(filename string, trackId, sliceCount, sliceNumber int, bpmSource, bpmTarget, sliceDuration float32,
	retrigTimes int, retrigBeats float32, retrigRateStart, retrigRateEnd, retrigPitch, retrigVolume float32) SamplerOSCParams {
	return SamplerOSCParams{
		Filename:              filename,
		TrackId:               trackId,
		SliceCount:            sliceCount,
		SliceNumber:           sliceNumber,
		SliceDuration:         sliceDuration,
		BPMSource:             bpmSource,
		BPMTarget:             bpmTarget,
		Pitch:                 0.0, // Default pitch (hex 80 = 0.0 pitch)
		RetriggerNumTotal:     retrigTimes,
		RetriggerBeats:        retrigBeats,
		RetriggerRateStart:    retrigRateStart,
		RetriggerRateEnd:      retrigRateEnd,
		RetriggerPitchChange:  retrigPitch,
		RetriggerVolumeChange: retrigVolume,
		TimestretchStart:      0,
		TimestretchEnd:        0,
		TimestretchBeats:      0,
		EffectReverse:         0,
		Pan:                   0,     // Default center pan
		LowPassFilter:         20000, // Default no filter (20kHz)
		HighPassFilter:        20,    // Default no filter (20Hz)
		EffectComb:            0,
		EffectReverb:          0,
	}
}

// NewInstrumentOSCParams creates instrument parameters
func NewInstrumentOSCParams(trackId, midiNote int, velocity float32) InstrumentOSCParams {
	return InstrumentOSCParams{
		TrackId:  trackId,
		MidiNote: midiNote,
		Velocity: velocity,
	}
}

func (m *Model) SendOSCInstrumentMessage(params InstrumentOSCParams) {
	if m.oscClient == nil {
		return // OSC not configured
	}

	msg := osc.NewMessage("/instrument")
	msg.Append(int32(params.TrackId)) // Track ID
	msg.Append("trackVolume")
	msg.Append(float32(m.TrackSetLevels[params.TrackId]))
	msg.Append("midiNote")
	msg.Append(int32(params.MidiNote))
	msg.Append("velocity")
	msg.Append(float32(params.Velocity))

	err := m.oscClient.Send(msg)
	if err != nil {
		log.Printf("Error sending OSC instrument message: %v", err)
	} else {
		log.Printf("OSC instrument message sent: /instrument track=%d note=%d velocity=%.2f",
			params.TrackId, params.MidiNote, params.Velocity)
	}
}

func (m *Model) SendOSCSamplerMessage(params SamplerOSCParams) {
	if m.oscClient == nil {
		return // OSC not configured
	}

	msg := osc.NewMessage("/sampler")
	msg.Append(params.Filename)
	msg.Append(int32(params.TrackId)) // Track ID
	msg.Append("trackVolume")
	msg.Append(float32(m.TrackSetLevels[params.TrackId]))
	msg.Append("sliceCount")
	msg.Append(int32(params.SliceCount))
	msg.Append("sliceNum")
	msg.Append(int32(params.SliceNumber))
	msg.Append("sliceDurationBeats")
	msg.Append(float32(params.SliceDuration))
	msg.Append("bpmSource")
	msg.Append(float32(params.BPMSource))
	msg.Append("bpmTarget")
	msg.Append(float32(params.BPMTarget))
	msg.Append("pitch")
	msg.Append(float32(params.Pitch))
	msg.Append("retrigNumTotal")
	msg.Append(int32(params.RetriggerNumTotal))
	msg.Append("retrigRateChangeBeats")
	msg.Append(float32(params.RetriggerBeats))
	msg.Append("retrigRateStart")
	msg.Append(float32(params.RetriggerRateStart))
	msg.Append("retrigRateEnd")
	msg.Append(float32(params.RetriggerRateEnd))
	msg.Append("retrigPitchChange")
	msg.Append(float32(params.RetriggerPitchChange))
	msg.Append("retrigVolumeChange")
	msg.Append(float32(params.RetriggerVolumeChange))
	msg.Append("effectTimestretchStart")
	msg.Append(float32(params.TimestretchStart))
	msg.Append("effectTimestretchEnd")
	msg.Append(float32(params.TimestretchEnd))
	msg.Append("effectTimestretchBeats")
	msg.Append(float32(params.TimestretchBeats))
	msg.Append("effectReverse")
	msg.Append(int32(params.EffectReverse))
	msg.Append("pan")
	msg.Append(float32(params.Pan))
	msg.Append("lowPassFilter")
	msg.Append(float32(params.LowPassFilter))
	msg.Append("highPassFilter")
	msg.Append(float32(params.HighPassFilter))
	msg.Append("effectComb")
	msg.Append(float32(params.EffectComb))
	msg.Append("effectReverb")
	msg.Append(float32(params.EffectReverb))

	err := m.oscClient.Send(msg)
	if err != nil {
		log.Printf("Error sending OSC sampler message: %v", err)
	} else {
		log.Printf("OSC sampler message sent: /sampler '%s' slices=%d slice=%d duration=%.2f bpmSource=%.2f bpmTarget=%.2f",
			params.Filename, params.SliceCount, params.SliceNumber, params.SliceDuration, params.BPMSource, params.BPMTarget)
	}
}

func (m *Model) SendOSCPlaybackMessage(filepath string, playing bool) {
	playingInt := int32(0)
	if playing {
		playingInt = 1
	}

	config := OSCMessageConfig{
		Address:    "/playback",
		Parameters: []interface{}{filepath, playingInt},
		LogFormat:  "OSC message sent: /playback '%s' %d",
		LogArgs:    []interface{}{filepath, int(playingInt)},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCPregainMessage() {
	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{"pregain", m.PregainDB},
		LogFormat:  "OSC pregain message sent: /set 'pregain' %.1f",
		LogArgs:    []interface{}{m.PregainDB},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCPostgainMessage() {
	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{"postgain", m.PostgainDB},
		LogFormat:  "OSC postgain message sent: /set 'postgain' %.1f",
		LogArgs:    []interface{}{m.PostgainDB},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCBiasMessage() {
	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{"bias", m.BiasDB},
		LogFormat:  "OSC bias message sent: /set 'bias' %.1f",
		LogArgs:    []interface{}{m.BiasDB},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCSaturationMessage() {
	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{"saturation", m.SaturationDB},
		LogFormat:  "OSC saturation message sent: /set 'saturation' %.1f",
		LogArgs:    []interface{}{m.SaturationDB},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCDriveMessage() {
	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{"drive", m.DriveDB},
		LogFormat:  "OSC drive message sent: /set 'drive' %.1f",
		LogArgs:    []interface{}{m.DriveDB},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCTrackSetLevelMessage(trackNum int) {
	if trackNum < 0 || trackNum >= 8 {
		return
	}

	trackParam := fmt.Sprintf("track%d", trackNum)
	setLevel := m.TrackSetLevels[trackNum]

	config := OSCMessageConfig{
		Address:    "/set",
		Parameters: []interface{}{trackParam, setLevel},
		LogFormat:  "OSC track set level message sent: /set '%s' %.1f",
		LogArgs:    []interface{}{trackParam, setLevel},
	}

	m.sendOSCMessage(config)
}

func (m *Model) SendOSCRecordMessage(filename string, recording bool) {
	recordingInt := int32(0)
	if recording {
		recordingInt = 1
	}

	config := OSCMessageConfig{
		Address:    "/record",
		Parameters: []interface{}{filename, recordingInt},
		LogFormat:  "OSC recording message sent: /record '%s' %d",
		LogArgs:    []interface{}{filename, int(recordingInt)},
	}

	m.sendOSCMessage(config)
}

func (m *Model) GenerateRecordingFilename() string {
	now := time.Now()
	return fmt.Sprintf("%04d-%02d-%02d-%02d-%02d-%02d.wav",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())
}

func (m *Model) PushWaveformSample(v float64, maxCols int) {
	// keep just enough points to draw across the current width
	// we draw one "dot column" per half braille cell, so keep 2*maxCols
	target := maxCols
	m.WaveformBuf = append(m.WaveformBuf, v)
	if len(m.WaveformBuf) > target {
		m.WaveformBuf = m.WaveformBuf[len(m.WaveformBuf)-target:]
	}
}

func (m *Model) SendStopOSC() {
	if m.oscClient == nil {
		return
	}
	msg := osc.NewMessage("/stop")
	_ = m.oscClient.Send(msg) // ignore error or log if you prefer
}

// GetPhraseViewType determines if the current track context should use Sampler or Instrument phrase view
// Uses the TrackTypes array set in the mixer view (false = Instrument, true = Sampler)
func (m *Model) GetPhraseViewType() types.PhraseViewType {
	if m.CurrentTrack >= 0 && m.CurrentTrack < 8 {
		if m.TrackTypes[m.CurrentTrack] {
			return types.SamplerPhraseView // true = Sampler
		} else {
			return types.InstrumentPhraseView // false = Instrument
		}
	}
	// Default to Sampler for invalid track numbers
	return types.SamplerPhraseView
}

// sendOSCMessage provides common logic for sending OSC messages
func (m *Model) sendOSCMessage(config OSCMessageConfig) {
	if m.oscClient == nil {
		return // OSC not configured
	}

	msg := osc.NewMessage(config.Address)
	for _, param := range config.Parameters {
		msg.Append(param)
	}

	err := m.oscClient.Send(msg)
	if err != nil {
		log.Printf("Error sending OSC message to %s: %v", config.Address, err)
	} else {
		if config.LogFormat != "" {
			log.Printf(config.LogFormat, config.LogArgs...)
		}
	}
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/schollz/2n/internal/types"
)

func TestNewModel(t *testing.T) {
	m := NewModel(57120, "test-save.json")

	// Test default initialization
	assert.Equal(t, 0, m.CurrentRow)
	assert.Equal(t, 1, m.CurrentCol) // Starts at phrase column in chain view
	assert.Equal(t, 0, m.ScrollOffset)
	assert.Equal(t, types.SongView, m.ViewMode) // Starts with song view
	assert.False(t, m.IsPlaying)
	assert.Equal(t, float32(120), m.BPM)
	assert.Equal(t, 2, m.PPQ) // Default PPQ is 2
	assert.Equal(t, float32(0.0), m.PregainDB)
	assert.Equal(t, float32(0.0), m.PostgainDB)
	assert.Equal(t, float32(-6.0), m.BiasDB)
	assert.Equal(t, float32(-6.0), m.SaturationDB)
	assert.Equal(t, float32(-6.0), m.DriveDB)

	// Test data structure initialization
	assert.Len(t, m.PhrasesData, 255)
	assert.Len(t, m.InstrumentPhrasesData, 255)
	assert.Len(t, m.SamplerPhrasesData, 255)
	assert.Len(t, m.RetriggerSettings, 255)
	assert.Len(t, m.TimestrechSettings, 255)
	assert.Len(t, m.ArpeggioSettings, 255)
	assert.Len(t, m.MidiSettings, 255)
	assert.Len(t, m.SoundMakerSettings, 255)

	// Test song data structure (8 tracks × 16 rows)
	assert.Len(t, m.SongData, 8)
	for i := 0; i < 8; i++ {
		assert.Len(t, m.SongData[i], 16)
		// All should be initialized to -1 (empty)
		for j := 0; j < 16; j++ {
			assert.Equal(t, -1, m.SongData[i][j])
		}
	}
}

func TestModelDataStructures(t *testing.T) {
	m := NewModel(0, "")

	// Test accessing existing methods
	data := m.GetCurrentPhrasesData()
	assert.NotNil(t, data)

	chainsData := m.GetCurrentChainsData()
	assert.NotNil(t, chainsData)

	// Test phrase view type determination
	m.CurrentTrack = 2 // Instrument track
	phraseType := m.GetPhraseViewType()
	assert.Equal(t, types.SamplerPhraseView, phraseType) // Track type defaults to sampler

	m.CurrentTrack = 5 // Sampler track
	phraseType = m.GetPhraseViewType()
	assert.Equal(t, types.SamplerPhraseView, phraseType)
}

func TestModelDataManipulation(t *testing.T) {
	m := NewModel(0, "")

	// Test SetChainsData method
	m.SetChainsData(0, 0, 10)
	assert.Equal(t, 10, m.ChainsData[0][0])

	// Test SetPhrasesData method
	m.SetPhrasesData(5, 10, int(types.ColNote), 0x60)
	assert.Equal(t, 0x60, m.PhrasesData[5][10][types.ColNote])

	// Test bounds checking - should not panic
	m.SetChainsData(-1, 0, 10)
	m.SetChainsData(300, 0, 10)
	m.SetPhrasesData(-1, 0, 0, 10)
	m.SetPhrasesData(300, 0, 0, 10)
}

func TestModelWaveformOperations(t *testing.T) {
	m := NewModel(0, "")

	// Test PushWaveformSample method
	maxCols := 50
	m.PushWaveformSample(0.5, maxCols)
	m.PushWaveformSample(-0.3, maxCols)
	m.PushWaveformSample(0.8, maxCols)

	assert.Len(t, m.WaveformBuf, 3)
	assert.Equal(t, 0.5, m.WaveformBuf[0])
	assert.Equal(t, -0.3, m.WaveformBuf[1])
	assert.Equal(t, 0.8, m.WaveformBuf[2])

	// Test LastWaveform tracking
	m.LastWaveform = 0.25
	assert.Equal(t, 0.25, m.LastWaveform)
}

func TestModelOSCIntegration(t *testing.T) {
	m := NewModel(57120, "")

	// Test OSC message methods (they should not panic)
	m.SendOSCPregainMessage()
	m.SendOSCPostgainMessage()
	m.SendOSCBiasMessage()
	m.SendOSCSaturationMessage()
	m.SendOSCDriveMessage()

	// Test track level OSC messages
	for track := 0; track < 8; track++ {
		m.SendOSCTrackSetLevelMessage(track)
	}

	// Test playback OSC messages
	m.SendOSCPlaybackMessage("test.wav", true)
	m.SendOSCPlaybackMessage("test.wav", false)

	// Test stop OSC
	m.SendStopOSC()
}

func TestModelFileOperations(t *testing.T) {
	m := NewModel(0, "")

	// Test file appending for sampler tracks
	m.CurrentTrack = 5 // Sampler track
	fileIndex := m.AppendPhrasesFile("test.wav")
	assert.Equal(t, 0, fileIndex) // First file
	assert.Len(t, m.SamplerPhrasesFiles, 1)
	assert.Equal(t, "test.wav", m.SamplerPhrasesFiles[0])

	// Test adding another file
	fileIndex = m.AppendPhrasesFile("test2.wav")
	assert.Equal(t, 1, fileIndex) // Second file
	assert.Len(t, m.SamplerPhrasesFiles, 2)
}

func TestModelColumnMapping(t *testing.T) {
	m := NewModel(0, "")

	// Test column mapping for instrument view
	m.CurrentTrack = 1 // Instrument track
	mapping := m.GetColumnMapping(0)
	assert.NotNil(t, mapping)

	// Test column mapping for sampler view
	m.CurrentTrack = 6 // Sampler track
	mapping = m.GetColumnMapping(0)
	assert.NotNil(t, mapping)
}

func TestModelRecordingFilename(t *testing.T) {
	m := NewModel(0, "")

	filename := m.GenerateRecordingFilename()
	assert.NotEmpty(t, filename)
	assert.Contains(t, filename, ".wav")
}

func TestModelPlaybackState(t *testing.T) {
	m := NewModel(0, "")

	// Test initial playback state
	assert.False(t, m.IsPlaying)
	assert.Equal(t, 0, m.PlaybackRow)
	assert.Equal(t, 0, m.PlaybackChain)
	assert.Equal(t, 0, m.PlaybackChainRow)
	assert.Equal(t, 0, m.PlaybackPhrase)

	// Test basic playback state changes
	m.IsPlaying = true
	assert.True(t, m.IsPlaying)

	m.PlaybackRow = 5
	assert.Equal(t, 5, m.PlaybackRow)
}

func TestModelDataInitialization(t *testing.T) {
	m := NewModel(0, "")

	// Test that data is properly initialized with defaults
	// Check phrase data defaults for sampler
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColNote])
	assert.Equal(t, 128, m.SamplerPhrasesData[0][0][types.ColPitch])
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColDeltaTime])
	assert.Equal(t, 80, m.SamplerPhrasesData[0][0][types.ColGate])

	// Check phrase data defaults for instrument
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColNote])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColDeltaTime])
	assert.Equal(t, int(types.ChordNone), m.InstrumentPhrasesData[0][0][types.ColChord])
	assert.Equal(t, int(types.ChordAddNone), m.InstrumentPhrasesData[0][0][types.ColChordAddition])

	// Check chain data initialization
	assert.Equal(t, -1, m.InstrumentChainsData[0][0])
	assert.Equal(t, -1, m.SamplerChainsData[0][0])
}

func BenchmarkModelInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModel(57120, "bench-test.json")
	}
}

func BenchmarkWaveformBufferPush(b *testing.B) {
	m := NewModel(0, "")
	maxCols := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.PushWaveformSample(0.5, maxCols)
	}
}

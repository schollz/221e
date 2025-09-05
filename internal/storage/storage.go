package storage

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	mu           sync.Mutex
	timer        *time.Timer
	debounceTime = 1 * time.Second
)

func AutoSave(m *model.Model) {
	mu.Lock()
	defer mu.Unlock()

	if timer != nil {
		// Stop the previous timer if still running
		timer.Stop()
	}

	// Start a new timer
	timer = time.AfterFunc(debounceTime, func() {
		// Place your actual save logic here
		go func() {
			startTime := time.Now()
			DoSave(m)
			elapsed := time.Since(startTime).Milliseconds()
			log.Printf("autosaved in %d ms", elapsed)
		}()
	})
}

func DoSave(m *model.Model) {
	log.Printf("doing save")
	saveData := types.SaveData{
		ViewMode:      m.ViewMode,
		CurrentRow:    m.CurrentRow,
		CurrentCol:    m.CurrentCol,
		ScrollOffset:  m.ScrollOffset,
		CurrentPhrase: m.CurrentPhrase,
		FileSelectRow: m.FileSelectRow,
		FileSelectCol: m.FileSelectCol,
		ChainsData:    m.ChainsData,
		PhrasesData:   m.PhrasesData,
		// New separate data pools
		InstrumentChainsData:  m.InstrumentChainsData,
		InstrumentPhrasesData: m.InstrumentPhrasesData,
		SamplerChainsData:     m.SamplerChainsData,
		SamplerPhrasesData:    m.SamplerPhrasesData,
		SamplerPhrasesFiles:   m.SamplerPhrasesFiles,
		LastEditRow:           m.LastEditRow,
		PhrasesFiles:          m.PhrasesFiles,
		CurrentDir:            m.CurrentDir,
		BPM:                   m.BPM,
		PPQ:                   m.PPQ,
		PregainDB:             m.PregainDB,
		PostgainDB:            m.PostgainDB,
		BiasDB:                m.BiasDB,
		SaturationDB:          m.SaturationDB,
		DriveDB:               m.DriveDB,
		FileMetadata:          m.FileMetadata,
		LastChainRow:          m.LastChainRow,
		LastPhraseRow:         m.LastPhraseRow,
		RecordingEnabled:      m.RecordingEnabled,
		RetriggerSettings:     m.RetriggerSettings,
		TimestrechSettings:    m.TimestrechSettings,
		ArpeggioSettings:      m.ArpeggioSettings,
		MidiSettings:          m.MidiSettings,
		SoundMakerSettings:    m.SoundMakerSettings,
		SongData:              m.SongData,
		LastSongRow:           m.LastSongRow,
		LastSongTrack:         m.LastSongTrack,
		CurrentChain:          m.CurrentChain,
		CurrentTrack:          m.CurrentTrack,
		TrackSetLevels:        m.TrackSetLevels,
		TrackTypes:            m.TrackTypes,
		CurrentMixerTrack:     m.CurrentMixerTrack,
	}

	data, err := json.Marshal(saveData)
	if err != nil {
		log.Printf("Error marshaling save data: %v", err)
		return
	}

	err = os.WriteFile(m.SaveFile, data, 0644)
	if err != nil {
		log.Printf("Error writing save file: %v", err)
		return
	}
}

func LoadState(m *model.Model, oscPort int, saveFile string) error {
	data, err := os.ReadFile(saveFile)
	if err != nil {
		return err
	}

	var saveData types.SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return err
	}

	// Force-return to PhraseView from non-main views (keep SongView, ChainView, and MixerView)
	if saveData.ViewMode == types.FileView ||
		saveData.ViewMode == types.SettingsView ||
		saveData.ViewMode == types.FileMetadataView ||
		saveData.ViewMode == types.RetriggerView ||
		saveData.ViewMode == types.TimestrechView {
		saveData.ViewMode = types.PhraseView
		saveData.CurrentCol = int(types.ColFilename)
	}

	// Restore state (assumes save file is already in the current format)
	m.ViewMode = saveData.ViewMode
	m.CurrentRow = saveData.CurrentRow
	m.CurrentCol = saveData.CurrentCol
	m.ScrollOffset = saveData.ScrollOffset
	m.CurrentPhrase = saveData.CurrentPhrase
	m.FileSelectRow = saveData.FileSelectRow
	m.FileSelectCol = saveData.FileSelectCol
	m.LastEditRow = saveData.LastEditRow
	m.CurrentDir = saveData.CurrentDir
	m.BPM = saveData.BPM
	m.PPQ = saveData.PPQ
	m.PregainDB = saveData.PregainDB
	m.PostgainDB = saveData.PostgainDB
	m.BiasDB = saveData.BiasDB
	m.SaturationDB = saveData.SaturationDB
	m.DriveDB = saveData.DriveDB
	m.FileMetadata = saveData.FileMetadata
	m.LastChainRow = saveData.LastChainRow
	m.LastPhraseRow = saveData.LastPhraseRow
	m.RecordingEnabled = saveData.RecordingEnabled
	m.RetriggerSettings = saveData.RetriggerSettings
	m.TimestrechSettings = saveData.TimestrechSettings
	m.ArpeggioSettings = saveData.ArpeggioSettings
	m.MidiSettings = saveData.MidiSettings
	m.SoundMakerSettings = saveData.SoundMakerSettings
	m.SongData = saveData.SongData
	m.LastSongRow = saveData.LastSongRow
	m.LastSongTrack = saveData.LastSongTrack
	m.CurrentChain = saveData.CurrentChain
	m.CurrentTrack = saveData.CurrentTrack
	m.TrackSetLevels = saveData.TrackSetLevels
	m.TrackTypes = saveData.TrackTypes
	m.CurrentMixerTrack = saveData.CurrentMixerTrack

	// Bulk-assign arrays
	m.ChainsData = saveData.ChainsData
	m.PhrasesData = saveData.PhrasesData

	// Load new separate data pools (with backwards compatibility)
	if saveData.InstrumentChainsData != nil {
		m.InstrumentChainsData = saveData.InstrumentChainsData
	}
	if len(saveData.InstrumentPhrasesData) > 0 && saveData.InstrumentPhrasesData[0] != nil {
		m.InstrumentPhrasesData = saveData.InstrumentPhrasesData
	}
	if saveData.SamplerChainsData != nil {
		m.SamplerChainsData = saveData.SamplerChainsData
	}
	if len(saveData.SamplerPhrasesData) > 0 && saveData.SamplerPhrasesData[0] != nil {
		m.SamplerPhrasesData = saveData.SamplerPhrasesData
	}
	if saveData.SamplerPhrasesFiles != nil {
		m.SamplerPhrasesFiles = append([]string(nil), saveData.SamplerPhrasesFiles...)
	}

	// Restore phrase file list
	m.PhrasesFiles = append([]string(nil), saveData.PhrasesFiles...)

	// Refresh file browser and push current volume to OSC
	LoadFiles(m)
	m.SendOSCPregainMessage()
	m.SendOSCPostgainMessage()
	m.SendOSCBiasMessage()
	m.SendOSCSaturationMessage()
	m.SendOSCDriveMessage()

	// Send track set levels to OSC on load
	for track := 0; track < 8; track++ {
		m.SendOSCTrackSetLevelMessage(track)
	}

	return nil
}

func LoadFiles(m *model.Model) {
	entries, err := os.ReadDir(m.CurrentDir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", m.CurrentDir, err)
		m.Files = []string{}
		return
	}

	var files []string

	// Add parent directory if not at root
	if m.CurrentDir != "/" && m.CurrentDir != "." {
		files = append(files, "..")
	}

	// Add directories first (including symlinked directories)
	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name()+"/")
		} else {
			// Check if this is a symlink to a directory
			fullPath := filepath.Join(m.CurrentDir, entry.Name())
			if stat, err := os.Stat(fullPath); err == nil && stat.IsDir() {
				files = append(files, entry.Name()+"/")
			}
		}
	}

	// Add audio files (including symlinked audio files)
	for _, entry := range entries {
		if !entry.IsDir() {
			fullPath := filepath.Join(m.CurrentDir, entry.Name())
			
			// Check if it's a regular file or a symlink to a file
			if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".wav" || ext == ".flac" {
					files = append(files, entry.Name())
				}
			}
		}
	}

	sort.Strings(files[1:]) // Sort everything except ".."
	m.Files = files
	log.Printf("Loaded %d files in %s", len(files), m.CurrentDir)
}

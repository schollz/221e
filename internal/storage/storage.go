package storage

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

func AutoSave(m *model.Model) {
	saveData := types.SaveData{
		ViewMode:           m.ViewMode,
		CurrentRow:         m.CurrentRow,
		CurrentCol:         m.CurrentCol,
		ScrollOffset:       m.ScrollOffset,
		CurrentPhrase:      m.CurrentPhrase,
		FileSelectRow:      m.FileSelectRow,
		ChainsData:         m.ChainsData,
		PhrasesData:        m.PhrasesData,
		LastEditRow:        m.LastEditRow,
		PhrasesFiles:       m.PhrasesFiles,
		CurrentDir:         m.CurrentDir,
		BPM:                m.BPM,
		PPQ:                m.PPQ,
		PregainDB:          m.PregainDB,
		PostgainDB:         m.PostgainDB,
		BiasDB:             m.BiasDB,
		SaturationDB:       m.SaturationDB,
		DriveDB:            m.DriveDB,
		FileMetadata:       m.FileMetadata,
		LastChainRow:       m.LastChainRow,
		LastPhraseRow:      m.LastPhraseRow,
		RecordingEnabled:   m.RecordingEnabled,
		RetriggerSettings:  m.RetriggerSettings,
		TimestrechSettings: m.TimestrechSettings,
		SongData:           m.SongData,
		LastSongRow:        m.LastSongRow,
		LastSongTrack:      m.LastSongTrack,
		CurrentChain:       m.CurrentChain,
		CurrentTrack:       m.CurrentTrack,
		TrackSetLevels:     m.TrackSetLevels,
		CurrentMixerTrack:  m.CurrentMixerTrack,
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
	m.SongData = saveData.SongData
	m.LastSongRow = saveData.LastSongRow
	m.LastSongTrack = saveData.LastSongTrack
	m.CurrentChain = saveData.CurrentChain
	m.CurrentTrack = saveData.CurrentTrack
	m.TrackSetLevels = saveData.TrackSetLevels
	m.CurrentMixerTrack = saveData.CurrentMixerTrack

	// Bulk-assign arrays
	m.ChainsData = saveData.ChainsData
	m.PhrasesData = saveData.PhrasesData

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

	// Add directories first
	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name()+"/")
		}
	}

	// Add audio files
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".wav" || ext == ".flac" {
				files = append(files, entry.Name())
			}
		}
	}

	sort.Strings(files[1:]) // Sort everything except ".."
	m.Files = files
	log.Printf("Loaded %d files in %s", len(files), m.CurrentDir)
}

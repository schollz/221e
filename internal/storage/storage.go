package storage

import (
	"compress/gzip"
	"fmt"
	"io"
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

	// Create save folder and copy sampler files, then get relative paths
	log.Printf("Saving SamplerPhrasesFiles: %v", m.SamplerPhrasesFiles)
	relativePaths, err := createSaveFolder(m.SaveFolder, m.SamplerPhrasesFiles, m.FileMetadata)
	if err != nil {
		log.Printf("Error creating save folder: %v", err)
		// Continue with normal save without bundling
		relativePaths = m.SamplerPhrasesFiles
	} else {
		log.Printf("Created save folder: %s", m.SaveFolder)
	}
	log.Printf("Relative paths for save: %v", relativePaths)

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
		SamplerPhrasesFiles:   relativePaths, // Use relative paths in save data
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
		LastPhraseCol:         m.LastPhraseCol,
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

	// Create the data.json.gz file inside the save folder
	dataFilePath := filepath.Join(m.SaveFolder, "data.json.gz")
	file, err := os.Create(dataFilePath)
	if err != nil {
		log.Printf("Error creating save file: %v", err)
		return
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	_, err = gzWriter.Write(data)
	if err != nil {
		log.Printf("Error writing gzipped save data: %v", err)
		return
	}
}

func LoadState(m *model.Model, oscPort int, saveFolder string) error {
	// Construct path to data.json.gz inside save folder
	dataFilePath := filepath.Join(saveFolder, "data.json.gz")

	// Open the gzipped save file
	file, err := os.Open(dataFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Read the decompressed data
	data, err := io.ReadAll(gzReader)
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
	m.LastPhraseCol = saveData.LastPhraseCol
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
		// Convert relative paths to absolute paths for portable bundles
		log.Printf("Loading SamplerPhrasesFiles: %v", saveData.SamplerPhrasesFiles)
		resolvedPaths := resolvePortablePaths(saveFolder, saveData.SamplerPhrasesFiles)
		log.Printf("Resolved SamplerPhrasesFiles: %v", resolvedPaths)
		m.SamplerPhrasesFiles = append([]string(nil), resolvedPaths...)
	}

	// Restore phrase file list
	m.PhrasesFiles = append([]string(nil), saveData.PhrasesFiles...)

	// Load metadata for files in save folder
	err = LoadMetadataFromSaveFolder(saveFolder, m.FileMetadata)
	if err != nil {
		log.Printf("Warning: Failed to load metadata from save folder: %v", err)
	}

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

// createSaveFolder creates the save folder and copies sampler files into it
func createSaveFolder(saveFolder string, samplerFiles []string, fileMetadata map[string]types.FileMetadata) ([]string, error) {
	// Create save folder
	err := os.MkdirAll(saveFolder, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create save folder %s: %w", saveFolder, err)
	}

	if len(samplerFiles) == 0 {
		// No files to copy, return empty slice
		return []string{}, nil
	}

	// Process each sampler file
	relativePaths := make([]string, len(samplerFiles))
	for i, originalPath := range samplerFiles {
		if originalPath == "" {
			relativePaths[i] = ""
			continue
		}

		// Get original file name
		fileName := filepath.Base(originalPath)
		destPath := filepath.Join(saveFolder, fileName)

		// Check if the file is already in the save folder
		// This handles both exact path matches and files that are already in the target directory
		originalDir := filepath.Dir(originalPath)
		if originalPath == destPath || originalDir == saveFolder {
			// File is already in the save folder, just use the filename as relative path
			relativePaths[i] = fileName
			log.Printf("File already in save folder: %s (relative: %s)", originalPath, fileName)
			continue
		}

		// Copy file to save folder
		err := copyFile(originalPath, destPath)
		if err != nil {
			log.Printf("Warning: Failed to copy file %s to %s: %v", originalPath, destPath, err)
			// Use original path if copy fails
			relativePaths[i] = originalPath
			continue
		}

		// Save metadata alongside the file if it exists
		if metadata, exists := fileMetadata[originalPath]; exists {
			err = saveFileMetadata(saveFolder, originalPath, metadata)
			if err != nil {
				log.Printf("Warning: Failed to save metadata for %s: %v", originalPath, err)
			}
		}

		// Store just the filename as relative path (since files are in the same folder as data.json.gz)
		relativePaths[i] = fileName

		log.Printf("Copied file to save folder: %s -> %s (relative: %s)", originalPath, destPath, fileName)
	}

	return relativePaths, nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close() // Close immediately on error
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Sync and close destination file explicitly
	err = dstFile.Sync()
	if err != nil {
		dstFile.Close()
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	err = dstFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// saveFileMetadata saves metadata as JSON for a wav file in the save folder
func saveFileMetadata(saveFolder, originalPath string, metadata types.FileMetadata) error {
	fileName := filepath.Base(originalPath)
	metadataFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".metadata.json"
	metadataPath := filepath.Join(saveFolder, metadataFileName)

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = os.WriteFile(metadataPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	log.Printf("Saved metadata for %s to %s", fileName, metadataPath)
	return nil
}

// loadFileMetadata loads metadata from JSON for a wav file in the save folder
func loadFileMetadata(saveFolder, fileName string) (types.FileMetadata, error) {
	metadataFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".metadata.json"
	metadataPath := filepath.Join(saveFolder, metadataFileName)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		// Return zero-value metadata if file doesn't exist
		if os.IsNotExist(err) {
			return types.FileMetadata{}, nil
		}
		return types.FileMetadata{}, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata types.FileMetadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return types.FileMetadata{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	log.Printf("Loaded metadata for %s from %s", fileName, metadataPath)
	return metadata, nil
}

// resolvePortablePaths converts relative paths from save folder back to absolute paths
func resolvePortablePaths(saveFolder string, paths []string) []string {
	if len(paths) == 0 {
		return paths
	}

	resolvedPaths := make([]string, len(paths))

	for i, path := range paths {
		if path == "" {
			resolvedPaths[i] = ""
			continue
		}

		// If path is already absolute, use it as-is
		if filepath.IsAbs(path) {
			resolvedPaths[i] = path
			continue
		}

		// Convert relative path to absolute by joining with save folder
		absolutePath := filepath.Join(saveFolder, path)

		// Check if the file exists in save folder
		if _, err := os.Stat(absolutePath); err == nil {
			resolvedPaths[i] = absolutePath
			log.Printf("Resolved file from save folder: %s -> %s", path, absolutePath)
		} else {
			// File doesn't exist in save folder, keep original relative path
			// This handles cases where files were saved before bundling feature
			log.Printf("Warning: File not found in save folder: %s", absolutePath)
			resolvedPaths[i] = path
		}
	}

	return resolvedPaths
}

// LoadMetadataFromSaveFolder loads metadata for all wav files in the save folder
func LoadMetadataFromSaveFolder(saveFolder string, fileMetadata map[string]types.FileMetadata) error {
	entries, err := os.ReadDir(saveFolder)
	if err != nil {
		return fmt.Errorf("failed to read save folder: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		// Only process wav files
		if ext == ".wav" {
			filePath := filepath.Join(saveFolder, fileName)
			metadata, err := loadFileMetadata(saveFolder, fileName)
			if err != nil {
				log.Printf("Warning: Failed to load metadata for %s: %v", fileName, err)
				continue
			}

			// Only add metadata if it has meaningful data (non-zero BPM or slices)
			if metadata.BPM > 0 || metadata.Slices > 0 {
				fileMetadata[filePath] = metadata
				log.Printf("Loaded metadata for %s: BPM=%.1f, Slices=%d", fileName, metadata.BPM, metadata.Slices)
			}
		}
	}

	return nil
}

// SaveMetadataForFile saves metadata for a specific file if it exists in the FileMetadata map
// This can be called whenever a wav file is created to save its associated metadata
func SaveMetadataForFile(filePath string, fileMetadata map[string]types.FileMetadata) error {
	if metadata, exists := fileMetadata[filePath]; exists {
		dir := filepath.Dir(filePath)
		return saveFileMetadata(dir, filePath, metadata)
	}
	return nil
}

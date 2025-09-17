package audio

import (
	"log"
	"math"
	"path/filepath"
	"strings"

	"github.com/schollz/collidertracker/internal/getbpm"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/types"
)

func PlayFile(m *model.Model) {
	if len(m.Files) == 0 || m.CurrentRow >= len(m.Files) {
		return
	}

	filename := m.Files[m.CurrentRow]

	// Don't play directories
	if strings.HasSuffix(filename, "/") || filename == ".." {
		return
	}

	// Get full path for the file
	fullPath := filepath.Join(m.CurrentDir, filename)

	// Check if this specific file is currently playing
	if m.CurrentlyPlayingFile == fullPath {
		// Same file is playing, so stop it
		m.CurrentlyPlayingFile = ""
		m.IsPlaying = false
		m.SendOSCPlaybackMessage(fullPath, false)
		log.Printf("File playback stopped: %s", filename)
	} else {
		// Different file or no file playing, so start playing this one
		// First stop any currently playing file
		if m.CurrentlyPlayingFile != "" {
			m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
			log.Printf("Stopping previously playing file: %s", m.CurrentlyPlayingFile)
		}
		// Now start the new file
		m.CurrentlyPlayingFile = fullPath
		m.IsPlaying = true
		m.SendOSCPlaybackMessage(fullPath, true)
		log.Printf("File playback started: %s", filename)
	}
}

func SelectFile(m *model.Model) {
	if len(m.Files) == 0 || m.CurrentRow >= len(m.Files) {
		return
	}

	selected := m.Files[m.CurrentRow]

	// Handle directory navigation
	if selected == ".." {
		m.CurrentDir = filepath.Dir(m.CurrentDir)
		storage.LoadFiles(m)
		m.CurrentRow = 0
		m.ScrollOffset = 0
		return
	}

	if strings.HasSuffix(selected, "/") {
		m.CurrentDir = filepath.Join(m.CurrentDir, strings.TrimSuffix(selected, "/"))
		storage.LoadFiles(m)
		m.CurrentRow = 0
		m.ScrollOffset = 0
		return
	}

	// Select audio file - store the full path
	fullPath := filepath.Join(m.CurrentDir, selected)
	fileIndex := m.AppendPhrasesFile(fullPath)
	phrasesData := m.GetCurrentPhrasesData()
	(*phrasesData)[m.CurrentPhrase][m.FileSelectRow][int(types.ColFilename)] = fileIndex

	// Set initial metadata using getbpm.GetBPM
	// BPM should be float, slices should be 2x beats (rounded to int)
	var bpm float64
	var beats float64
	var err error
	beats, bpm, err = getbpm.GetBPM(fullPath)
	if err == nil {
		slices := int(2 * math.Round(beats))
		m.FileMetadata[fullPath] = types.FileMetadata{
			BPM:    float32(bpm),
			Slices: slices,
		}
	} else {
		log.Printf("Could not get BPM for %s: %v", fullPath, err)
	}

	// Track this as the last edited row so "S" key will work
	m.LastEditRow = m.FileSelectRow

	log.Printf("Selected file %s (full path: %s) for phrase %d row %d", selected, fullPath, m.CurrentPhrase, m.FileSelectRow)
	storage.AutoSave(m)
}

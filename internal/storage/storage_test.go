package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDoSave(t *testing.T) {
	t.Run("successful save", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFile := filepath.Join(tmpDir, "test_save.json")

		m := model.NewModel(0, saveFile)
		m.BPM = 140
		m.CurrentRow = 5

		DoSave(m)

		// Check that file was created
		_, err := os.Stat(saveFile)
		assert.NoError(t, err)

		// Check that file has content
		data, err := os.ReadFile(saveFile)
		assert.NoError(t, err)
		assert.True(t, len(data) > 0)
	})

	t.Run("save to invalid path", func(t *testing.T) {
		m := model.NewModel(0, "/invalid/path/that/does/not/exist/save.json")

		// Should not panic, just log error
		DoSave(m)
	})
}

func TestLoadState(t *testing.T) {
	t.Run("load existing save file", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFile := filepath.Join(tmpDir, "test_load.json")

		// Create and save a model with specific state
		m1 := model.NewModel(0, saveFile)
		m1.BPM = 140
		m1.CurrentRow = 10
		m1.ViewMode = types.ChainView
		DoSave(m1)

		// Create new model and load state
		m2 := model.NewModel(0, saveFile)
		err := LoadState(m2, 0, saveFile)

		assert.NoError(t, err)
		assert.Equal(t, float32(140), m2.BPM)
		assert.Equal(t, 10, m2.CurrentRow)
		assert.Equal(t, types.ChainView, m2.ViewMode)
	})

	t.Run("load nonexistent file", func(t *testing.T) {
		m := model.NewModel(0, "")
		err := LoadState(m, 0, "/path/that/does/not/exist.json")

		assert.Error(t, err)
	})

	t.Run("force return to phrase view from non-main views", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFile := filepath.Join(tmpDir, "test_force_view.json")

		// Create and save a model in a non-main view
		m1 := model.NewModel(0, saveFile)
		m1.ViewMode = types.FileView // This should be forced to PhraseView
		DoSave(m1)

		// Load state
		m2 := model.NewModel(0, saveFile)
		err := LoadState(m2, 0, saveFile)

		assert.NoError(t, err)
		assert.Equal(t, types.PhraseView, m2.ViewMode)
		assert.Equal(t, int(types.ColFilename), m2.CurrentCol)
	})
}

func TestLoadFiles(t *testing.T) {
	t.Run("load files from existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		os.WriteFile(filepath.Join(tmpDir, "test1.wav"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "test2.flac"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "test3.txt"), []byte("test"), 0644) // Should be ignored
		os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

		m := model.NewModel(0, "")
		m.CurrentDir = tmpDir

		LoadFiles(m)

		// Should have parent dir, subdir, and audio files
		assert.Contains(t, m.Files, "..")
		assert.Contains(t, m.Files, "subdir/")
		assert.Contains(t, m.Files, "test1.wav")
		assert.Contains(t, m.Files, "test2.flac")
		assert.NotContains(t, m.Files, "test3.txt") // Non-audio files should be excluded
	})

	t.Run("load files from root directory", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.CurrentDir = "/"

		LoadFiles(m)

		// Should not have ".." when at root
		assert.NotContains(t, m.Files, "..")
	})

	t.Run("load files from nonexistent directory", func(t *testing.T) {
		m := model.NewModel(0, "")
		m.CurrentDir = "/path/that/does/not/exist"

		LoadFiles(m)

		// Should have empty files list
		assert.Equal(t, []string{}, m.Files)
	})
}

func TestAutoSave(t *testing.T) {
	t.Run("autosave debouncing", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFile := filepath.Join(tmpDir, "autosave_test.json")

		m := model.NewModel(0, saveFile)
		m.BPM = 150

		// Call AutoSave multiple times quickly
		AutoSave(m)
		AutoSave(m)
		AutoSave(m)

		// Should not save immediately
		_, err := os.Stat(saveFile)
		assert.True(t, os.IsNotExist(err))

		// Wait for debounce timeout plus a bit more
		time.Sleep(1200 * time.Millisecond)

		// Should have saved by now
		_, err = os.Stat(saveFile)
		assert.NoError(t, err)
	})
}

func BenchmarkDoSave(b *testing.B) {
	// Create a temporary file for testing
	tmpDir := b.TempDir()
	saveFile := filepath.Join(tmpDir, "test_save.json")

	// Create a model with default data
	m := model.NewModel(0, saveFile) // OSC port 0 to disable OSC

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DoSave(m)
	}
}

func BenchmarkLoadState(b *testing.B) {
	// Create a temporary file for testing
	tmpDir := b.TempDir()
	saveFile := filepath.Join(tmpDir, "test_load.json")

	// Create a model with default data and save it once
	m := model.NewModel(0, saveFile) // OSC port 0 to disable OSC
	DoSave(m)

	// Verify the save file exists
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		b.Fatal("Save file was not created")
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh model for each load operation
		testModel := model.NewModel(0, saveFile)
		err := LoadState(testModel, 0, saveFile)
		if err != nil {
			b.Fatalf("LoadState failed: %v", err)
		}
	}
}

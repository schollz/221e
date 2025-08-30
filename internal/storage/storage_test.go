package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/schollz/2n/internal/model"
)

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

package audio

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/types"
)

func TestPlayFile(t *testing.T) {
	t.Run("empty files list", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json") // OSC port 0 to disable OSC
		m.Files = []string{}

		PlayFile(m)

		// Should not change any state
		assert.Equal(t, "", m.CurrentlyPlayingFile)
		assert.False(t, m.IsPlaying)
	})

	t.Run("current row out of bounds", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test.wav"}
		m.CurrentRow = 1 // Out of bounds

		PlayFile(m)

		// Should not change any state
		assert.Equal(t, "", m.CurrentlyPlayingFile)
		assert.False(t, m.IsPlaying)
	})

	t.Run("skip directory", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"subdir/"}
		m.CurrentRow = 0
		m.CurrentDir = "/test"

		PlayFile(m)

		// Should not change any state for directories
		assert.Equal(t, "", m.CurrentlyPlayingFile)
		assert.False(t, m.IsPlaying)
	})

	t.Run("skip parent directory", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{".."}
		m.CurrentRow = 0
		m.CurrentDir = "/test"

		PlayFile(m)

		// Should not change any state for parent directory
		assert.Equal(t, "", m.CurrentlyPlayingFile)
		assert.False(t, m.IsPlaying)
	})

	t.Run("start playing new file", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test.wav"}
		m.CurrentRow = 0
		m.CurrentDir = "/test"

		PlayFile(m)

		expectedPath := filepath.Join("/test", "test.wav")
		assert.Equal(t, expectedPath, m.CurrentlyPlayingFile)
		assert.True(t, m.IsPlaying)
	})

	t.Run("stop currently playing file", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test.wav"}
		m.CurrentRow = 0
		m.CurrentDir = "/test"
		expectedPath := filepath.Join("/test", "test.wav")
		m.CurrentlyPlayingFile = expectedPath
		m.IsPlaying = true

		PlayFile(m)

		// Should stop the file
		assert.Equal(t, "", m.CurrentlyPlayingFile)
		assert.False(t, m.IsPlaying)
	})

	t.Run("switch to different file", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test1.wav", "test2.wav"}
		m.CurrentRow = 1
		m.CurrentDir = "/test"
		m.CurrentlyPlayingFile = filepath.Join("/test", "other.wav")
		m.IsPlaying = true

		PlayFile(m)

		expectedPath := filepath.Join("/test", "test2.wav")
		assert.Equal(t, expectedPath, m.CurrentlyPlayingFile)
		assert.True(t, m.IsPlaying)
	})
}

func TestSelectFile(t *testing.T) {
	t.Run("empty files list", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{}

		SelectFile(m)

		// Should not crash or change state
	})

	t.Run("current row out of bounds", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test.wav"}
		m.CurrentRow = 1 // Out of bounds

		SelectFile(m)

		// Should not crash or change state
	})

	t.Run("navigate to parent directory", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{".."}
		m.CurrentRow = 0
		m.CurrentDir = "/test/subdir"

		SelectFile(m)

		// Should navigate to parent directory
		assert.Equal(t, "/test", m.CurrentDir)
		assert.Equal(t, 0, m.CurrentRow)
		assert.Equal(t, 0, m.ScrollOffset)
	})

	t.Run("navigate to subdirectory", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"subdir/"}
		m.CurrentRow = 0
		m.CurrentDir = "/test"

		SelectFile(m)

		// Should navigate to subdirectory
		assert.Equal(t, "/test/subdir", m.CurrentDir)
		assert.Equal(t, 0, m.CurrentRow)
		assert.Equal(t, 0, m.ScrollOffset)
	})

	t.Run("select audio file", func(t *testing.T) {
		m := model.NewModel(0, "test-save.json")
		m.Files = []string{"test.wav"}
		m.CurrentRow = 0
		m.CurrentDir = "/test"
		m.FileSelectRow = 2
		m.CurrentPhrase = 1
		m.CurrentTrack = 5 // Set to sampler track so files can be appended
		initialFileCount := len(m.SamplerPhrasesFiles)

		SelectFile(m)

		// Should append file and set phrase data
		assert.Equal(t, initialFileCount+1, len(m.SamplerPhrasesFiles))
		assert.Equal(t, 2, m.LastEditRow)

		// Check that the file was added to phrases data
		phrasesData := m.GetCurrentPhrasesData()
		fileIndex := (*phrasesData)[1][2][int(types.ColFilename)]
		assert.NotEqual(t, -1, fileIndex)

		// The full path should be stored in SamplerPhrasesFiles
		fullPath := filepath.Join("/test", "test.wav")
		assert.Contains(t, m.SamplerPhrasesFiles, fullPath)
	})
}

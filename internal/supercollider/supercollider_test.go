package supercollider

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		assert.True(t, fileExists(testFile))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		assert.False(t, fileExists("/path/that/does/not/exist.txt"))
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		assert.True(t, fileExists(tmpDir))
	})
}

func TestWasStartedBySelf(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		// Should start as false
		result := WasStartedBySelf()
		assert.False(t, result)
	})
}

func TestGetPortedPluginsURL(t *testing.T) {
	t.Run("returns correct URL for platform", func(t *testing.T) {
		url := getPortedPluginsURL()

		switch runtime.GOOS {
		case "linux":
			if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
				assert.Contains(t, url, "RaspberryPi.zip")
			} else {
				assert.Contains(t, url, "Linux.zip")
			}

		case "darwin":
			if runtime.GOARCH == "arm64" {
				assert.Contains(t, url, "macOS-ARM.zip")
			} else {
				assert.Contains(t, url, "macOS.zip")
			}

		case "windows":
			assert.Contains(t, url, "Windows.zip")

		default:
			assert.Equal(t, "", url) // Unsupported platform
		}

		// If URL is not empty, should be a valid GitHub releases URL
		if url != "" {
			assert.Contains(t, url, "github.com/schollz/portedplugins/releases")
		}
	})
}

func TestHasExtension(t *testing.T) {
	t.Run("extension not found in empty dirs", func(t *testing.T) {
		// This will check real system dirs, but since we're looking for specific
		// extension files that probably don't exist, should return false
		result := hasExtension("NonexistentExtension.sc")
		assert.False(t, result)
	})
}

func TestHasRequiredExtensions(t *testing.T) {
	t.Run("check required extensions", func(t *testing.T) {
		// This will likely return false unless SuperCollider extensions are installed
		result := HasRequiredExtensions()
		// We can't assert true/false since it depends on system state,
		// but we can verify it doesn't panic
		assert.IsType(t, true, result)
	})
}

func TestGetMiUGensURL(t *testing.T) {
	t.Run("returns correct URL for platform", func(t *testing.T) {
		url := getMiUGensURL()

		switch runtime.GOOS {
		case "linux":
			assert.Contains(t, url, "mi-UGens-Linux.zip")
		case "darwin":
			assert.Contains(t, url, "mi-UGens-macOS.zip")
		case "windows":
			assert.Contains(t, url, "mi-UGens-Windows.zip")
		default:
			assert.Equal(t, "", url) // Unsupported platform
		}

		// If URL is not empty, should be a valid GitHub releases URL
		if url != "" {
			assert.Contains(t, url, "github.com/v7b1/mi-UGens/releases")
			assert.Contains(t, url, "v0.0.8")
		}
	})
}

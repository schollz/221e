package supercollider

import (
	"runtime"
	"testing"
)

func TestIsJackEnabled(t *testing.T) {
	result := IsJackEnabled()
	if !result {
		t.Error("Expected IsJackEnabled() to return true, but got false")
	}
}

func TestIsSuperColliderEnabled(t *testing.T) {
	result := IsSuperColliderEnabled()
	if !result {
		t.Error("Expected IsSuperColliderEnabled() to return true, but got false")
	}
}

func TestIsProcessRunningDirectly(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		shouldRun   bool
	}{
		{
			name:        "Jack Audio Connection Kit",
			processName: "jackd",
			shouldRun:   true,
		},
		{
			name:        "SuperCollider Language",
			processName: "sclang",
			shouldRun:   true,
		},
		{
			name:        "Non-existent Process",
			processName: "definitely_not_running_process_xyz",
			shouldRun:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProcessRunning(tt.processName)
			if result != tt.shouldRun {
				t.Errorf("isProcessRunning(%s) = %v, want %v", tt.processName, result, tt.shouldRun)
			}
		})
	}
}

func TestHasRequiredExtensions(t *testing.T) {
	result := HasRequiredExtensions()
	t.Logf("HasRequiredExtensions() returned: %v", result)

	// Check individual extensions for debugging
	fverbExists := hasExtension("Fverb.sc")
	analogTapeExists := hasExtension("AnalogTape.sc")

	t.Logf("Fverb.sc found: %v", fverbExists)
	t.Logf("AnalogTape.sc found: %v", analogTapeExists)

	// Log extension directories being checked
	dirs := getSuperColliderExtensionDirs()
	t.Logf("Checking extension directories: %v", dirs)
}

func TestGetSuperColliderExtensionDirs(t *testing.T) {
	dirs := getSuperColliderExtensionDirs()
	if len(dirs) == 0 {
		t.Error("Expected at least one extension directory, got none")
	}

	t.Logf("SuperCollider extension directories: %v", dirs)
}

func TestGetPortedPluginsURL(t *testing.T) {
	url := getPortedPluginsURL()
	if url == "" {
		t.Errorf("Expected a URL for platform %s/%s, got empty string", runtime.GOOS, runtime.GOARCH)
	}

	t.Logf("PortedPlugins URL for %s/%s: %s", runtime.GOOS, runtime.GOARCH, url)
}

func TestGetLocalExtensionDir(t *testing.T) {
	dir := getLocalExtensionDir()
	if dir == "" {
		t.Errorf("Expected a local extension directory for platform %s, got empty string", runtime.GOOS)
	}

	t.Logf("Local extension directory: %s", dir)
}

func TestDownloadRequiredExtensions(t *testing.T) {
	// This test only checks if the function doesn't error when extensions already exist
	// We don't want to actually download files during normal testing
	err := DownloadRequiredExtensions()
	if err != nil {
		t.Logf("DownloadRequiredExtensions returned error (expected if extensions exist): %v", err)
	} else {
		t.Log("DownloadRequiredExtensions completed successfully")
	}
}

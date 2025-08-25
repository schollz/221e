package supercollider

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed sampler.scd
var embeddedSamplerSCD []byte

var (
	startedBySelf   = false
	tempSamplerFile = ""
	sclangProcess   *exec.Cmd
	cleanupCalled   = false
)

func IsJackEnabled() bool {
	return isProcessRunning("jackd")
}

func IsSuperColliderEnabled() bool {
	return isProcessRunning("sclang")
}

func StartSuperCollider() error {
	if IsSuperColliderEnabled() {
		return nil // Already running (started externally)
	}

	// Find sclang executable
	sclangPath, err := findSclangPath()
	if err != nil {
		return fmt.Errorf("sclang not found: %v", err)
	}

	// Create temporary file from embedded sampler.scd
	tempFile, err := os.CreateTemp("", "sampler-*.scd")
	if err != nil {
		return fmt.Errorf("failed to create temporary sampler file: %v", err)
	}

	// Write embedded content to temp file
	_, err = tempFile.Write(embeddedSamplerSCD)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return fmt.Errorf("failed to write sampler content: %v", err)
	}
	tempFile.Close()

	// Store the temp file path for cleanup
	tempSamplerFile = tempFile.Name()

	// Start sclang with the temporary scd file
	sclangProcess = exec.Command(sclangPath, tempSamplerFile)

	// Start the process but don't wait for it to complete
	err = sclangProcess.Start()
	if err != nil {
		os.Remove(tempSamplerFile)
		tempSamplerFile = ""
		return fmt.Errorf("failed to start SuperCollider: %v", err)
	}

	// Mark that we started it
	startedBySelf = true

	// Wait a moment and check if it's actually running
	time.Sleep(2 * time.Second)
	if !IsSuperColliderEnabled() {
		// Clean up if it failed to start
		if sclangProcess.Process != nil {
			sclangProcess.Process.Kill()
		}
		os.Remove(tempSamplerFile)
		tempSamplerFile = ""
		startedBySelf = false
		return fmt.Errorf("SuperCollider failed to start properly")
	}

	return nil
}

func StartSuperColliderWithProgress(readyChannel <-chan struct{}) error {
	if IsSuperColliderEnabled() {
		return nil // Already running (started externally)
	}

	// Create and run the progress bar TUI
	model := NewStartupProgressModel(readyChannel)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if there was an error during startup
	if progressModel, ok := finalModel.(StartupProgressModel); ok {
		return progressModel.Error()
	}

	return nil
}

func Cleanup() {
	// Prevent multiple cleanup calls
	if cleanupCalled {
		return
	}
	cleanupCalled = true

	if startedBySelf {
		// Stop SuperCollider process if we started it
		if sclangProcess != nil && sclangProcess.Process != nil {
			sclangProcess.Process.Kill()
			sclangProcess.Wait() // Wait for it to actually stop
		}
		startedBySelf = false
		sclangProcess = nil
	}

	// Remove temporary sampler file if we created it
	if tempSamplerFile != "" {
		os.Remove(tempSamplerFile)
		tempSamplerFile = ""
	}
}

func WasStartedBySelf() bool {
	return startedBySelf
}

func findSclangPath() (string, error) {
	// First try to find sclang in PATH
	if path, err := exec.LookPath("sclang"); err == nil {
		return path, nil
	}

	// Platform-specific fallback locations
	var possiblePaths []string

	switch runtime.GOOS {
	case "windows":
		// Common installation paths on Windows
		possiblePaths = []string{
			"C:\\Program Files\\SuperCollider\\sclang.exe",
			"C:\\Program Files (x86)\\SuperCollider\\sclang.exe",
		}
		// Also check user's local app data
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			possiblePaths = append(possiblePaths, filepath.Join(localAppData, "SuperCollider", "sclang.exe"))
		}
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			possiblePaths = append(possiblePaths, filepath.Join(programData, "SuperCollider", "sclang.exe"))
		}

	case "darwin":
		// Common installation paths on macOS
		possiblePaths = []string{
			"/Applications/SuperCollider.app/Contents/MacOS/sclang",
			"/Applications/SuperCollider/SuperCollider.app/Contents/MacOS/sclang",
		}
		// Also check user's Applications
		if homeDir, err := os.UserHomeDir(); err == nil {
			possiblePaths = append(possiblePaths,
				filepath.Join(homeDir, "Applications", "SuperCollider.app", "Contents", "MacOS", "sclang"),
				filepath.Join(homeDir, "Applications", "SuperCollider", "SuperCollider.app", "Contents", "MacOS", "sclang"),
			)
		}

	case "linux":
		// Common installation paths on Linux
		possiblePaths = []string{
			"/usr/bin/sclang",
			"/usr/local/bin/sclang",
			"/opt/supercollider/bin/sclang",
		}
		// Also check user's local bin
		if homeDir, err := os.UserHomeDir(); err == nil {
			possiblePaths = append(possiblePaths,
				filepath.Join(homeDir, ".local", "bin", "sclang"),
				filepath.Join(homeDir, "bin", "sclang"),
			)
		}
	}

	// Check each possible path
	for _, path := range possiblePaths {
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("sclang executable not found in PATH or common installation locations")
}

func isProcessRunning(processName string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Exact image match on Windows is already supported by tasklist
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq "+processName+".exe")
	default: // darwin, linux, etc.
		// Use -x for exact match of the process name (no substring matches like "jackdbus")
		cmd = exec.Command("pgrep", "-x", processName)
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		out := strings.ToLower(string(output))
		return strings.Contains(out, strings.ToLower(processName+".exe"))
	}

	// pgrep returns PIDs if found; empty output means not running
	return len(strings.TrimSpace(string(output))) > 0
}

func HasRequiredExtensions() bool {
	extensions := []string{"Fverb.sc", "AnalogTape.sc"}

	for _, ext := range extensions {
		if !hasExtension(ext) {
			return false
		}
	}
	return true
}

func hasExtension(filename string) bool {
	extensionDirs := getSuperColliderExtensionDirs()

	for _, dir := range extensionDirs {
		// Check direct file path
		if fileExists(filepath.Join(dir, filename)) {
			return true
		}

		// Check in subdirectories recursively
		found := false
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == filename {
				found = true
				return filepath.SkipDir
			}
			return nil
		})

		if found {
			return true
		}
	}
	return false
}

func getSuperColliderExtensionDirs() []string {
	var dirs []string

	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions"))
		}
		dirs = append(dirs, "/Library/Application Support/SuperCollider/Extensions")
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(homeDir, ".local/share/SuperCollider/Extensions"))
		}
		dirs = append(dirs, "/usr/share/SuperCollider/Extensions")
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "SuperCollider/Extensions"))
		}
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			dirs = append(dirs, filepath.Join(programData, "SuperCollider/Extensions"))
		}
	}

	return dirs
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func DownloadRequiredExtensions() error {
	if HasRequiredExtensions() {
		return nil // Extensions already exist
	}

	downloadURL := getPortedPluginsURL()
	if downloadURL == "" {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	extensionDir := getLocalExtensionDir()
	if extensionDir == "" {
		return fmt.Errorf("could not determine local extension directory")
	}

	// Create extension directory if it doesn't exist
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension directory: %v", err)
	}

	// Download and extract the zip file
	return downloadAndExtract(downloadURL, extensionDir)
}

func getPortedPluginsURL() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
			return "https://github.com/schollz/portedplugins/releases/download/v0.4.6/PortedPlugins-RaspberryPi.zip"
		}
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-Linux.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-macOS-ARM.zip"
		}
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-macOS.zip"
	case "windows":
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-Windows.zip"
	}
	return ""
}

func getLocalExtensionDir() string {
	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions")
		}
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, ".local/share/SuperCollider/Extensions")
		}
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "SuperCollider/Extensions")
		}
	}
	return ""
}

func downloadAndExtract(url, destDir string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "portedplugins-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy response body to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}

	// Close temp file before reading
	tmpFile.Close()

	// Extract zip file
	return extractZip(tmpFile.Name(), destDir)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// Create destination directory
	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		// Create the directories for this file
		destPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.FileInfo().Mode())
			continue
		}

		// Create the directories for this file
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		// Open file in zip
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		// Create destination file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file: %v", err)
		}

		// Copy file contents
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}
	}

	return nil
}

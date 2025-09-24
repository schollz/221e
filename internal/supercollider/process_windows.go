//go:build windows

package supercollider

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// setupProcessGroup sets up platform-specific process attributes for Windows
func setupProcessGroup(cmd *exec.Cmd) {
	// On Windows, we don't need to set up process groups like Unix
	// The working directory is already set in the main code
}

// killProcessGroup kills the process and its children on Windows
func killProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	// On Windows, use taskkill for more reliable termination
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", cmd.Process.Pid))
	killCmd.Run() // Run taskkill but don't wait for output
	// Give it a moment to terminate
	time.Sleep(250 * time.Millisecond)
}

// isProcessStillRunning checks if a process is still running by PID on Windows
func isProcessStillRunning(pid int) bool {
	// Use tasklist to check if the process is still running
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If the process exists, tasklist will include it in the output
	return strings.Contains(string(output), strconv.Itoa(pid))
}

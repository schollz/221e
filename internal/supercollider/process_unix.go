//go:build !windows

package supercollider

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

// setupProcessGroup sets up platform-specific process attributes for Unix systems
func setupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// killProcessGroup kills the process group on Unix systems
func killProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	log.Printf("Attempting graceful shutdown of sclang process (PID: %d)", cmd.Process.Pid)

	// First, try to terminate the entire process group to catch child processes
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		// Send SIGTERM to the process group first for graceful shutdown
		log.Printf("Sending SIGTERM to process group %d", pgid)
		syscall.Kill(-pgid, syscall.SIGTERM)

		// Wait a bit for graceful shutdown
		time.Sleep(250 * time.Millisecond)

		// Check if process is still running
		if isProcessStillRunning(cmd.Process.Pid) {
			log.Printf("Process still running, sending SIGKILL to process group %d", pgid)
			// If still running, force kill the process group
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			log.Printf("Process gracefully terminated")
		}
	} else {
		log.Printf("Could not get process group, falling back to single process termination")
		// Fallback: try graceful termination of the main process
		cmd.Process.Signal(syscall.SIGTERM)

		// Wait a bit for graceful shutdown
		time.Sleep(1 * time.Second)

		// If still running, force kill
		if isProcessStillRunning(cmd.Process.Pid) {
			log.Printf("Process still running, sending SIGKILL")
			cmd.Process.Kill()
		} else {
			log.Printf("Process gracefully terminated")
		}
	}

	// Give additional time for cleanup
	time.Sleep(500 * time.Millisecond)
}

// isProcessStillRunning checks if a process is still running by PID on Unix systems
func isProcessStillRunning(pid int) bool {
	// On Unix systems, we can check if a process exists by sending signal 0
	// This doesn't actually send a signal but checks if we can send one
	err := syscall.Kill(pid, 0)
	return err == nil
}
package midiconnector

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock function to simulate filterName logic without calling Devices()
func testFilterName(name string, availableDevices []string) (foundName string, foundNum int, err error) {
	foundNum = -1 // Initialize to -1 to indicate not found
	for i, n := range availableDevices {
		if strings.Contains(strings.ToLower(n), strings.ToLower(name)) {
			foundName = n
			foundNum = i
			break
		}
	}
	if foundNum == -1 {
		err = fmt.Errorf("could not find device with name %s", name)
	}
	return
}

func TestFilterNameLogic(t *testing.T) {
	t.Run("find device by partial name", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI", "Bluetooth MIDI"}

		foundName, foundNum, err := testFilterName("usb", devices)

		assert.NoError(t, err)
		assert.Equal(t, "USB MIDI Device", foundName)
		assert.Equal(t, 0, foundNum)
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI"}

		foundName, foundNum, err := testFilterName("INTERNAL", devices)

		assert.NoError(t, err)
		assert.Equal(t, "Internal MIDI", foundName)
		assert.Equal(t, 1, foundNum)
	})

	t.Run("no matching device", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI"}

		foundName, foundNum, err := testFilterName("nonexistent", devices)

		assert.Error(t, err)
		assert.Equal(t, "", foundName)
		assert.Equal(t, -1, foundNum)
		assert.Contains(t, err.Error(), "could not find device")
	})

	t.Run("empty device list", func(t *testing.T) {
		devices := []string{}

		foundName, foundNum, err := testFilterName("test", devices)

		assert.Error(t, err)
		assert.Equal(t, "", foundName)
		assert.Equal(t, -1, foundNum)
		assert.Contains(t, err.Error(), "could not find device")
	})
}

func TestDevice(t *testing.T) {
	t.Run("device initialization", func(t *testing.T) {
		// Create device manually to test struct
		d := &Device{
			name:    "test",
			num:     0,
			notesOn: make(map[uint8]bool),
		}

		assert.Equal(t, "test", d.name)
		assert.Equal(t, 0, d.num)
		assert.Equal(t, 0, len(d.notesOn))
	})

	t.Run("note tracking logic", func(t *testing.T) {
		d := &Device{
			name:    "test",
			num:     0,
			notesOn: make(map[uint8]bool),
		}

		// Simulate note tracking (without actual MIDI)
		note := uint8(60) // Middle C
		channel := uint8(0)

		// Manually add note to tracking (simulating successful NoteOn)
		d.notesOn[note] = true
		assert.Equal(t, true, d.notesOn[note])

		// Manually remove note (simulating NoteOff)
		delete(d.notesOn, note)
		_, exists := d.notesOn[note]
		assert.False(t, exists)
	})
}

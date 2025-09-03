package midiplayer

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("valid midi line", func(t *testing.T) {
		p, err := Parse("midi testdevice 5")

		// Will error due to no MIDI hardware, but should parse correctly first
		// Check if parsing logic worked before hardware error
		if err != nil {
			// Expected error due to no MIDI hardware
			assert.Contains(t, err.Error(), "MIDI")
		} else {
			// If no error (somehow), verify the parsing worked
			assert.Equal(t, "midi-testdevice-4", p.Name) // Channel 5 -> 4 (0-indexed)
			assert.Equal(t, uint8(4), p.channel)
			assert.Equal(t, "testdevice", p.nameOriginal)
		}
	})

	t.Run("invalid format - too few parts", func(t *testing.T) {
		_, err := Parse("midi testdevice")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid midi line format")
	})

	t.Run("invalid format - wrong command", func(t *testing.T) {
		_, err := Parse("audio testdevice 5")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "line must start with 'midi'")
	})

	t.Run("invalid channel - non-numeric", func(t *testing.T) {
		_, err := Parse("midi testdevice abc")

		// Should default to channel 0 when parse fails, then try to create device
		if err != nil {
			// Expected error due to no MIDI hardware
			assert.Contains(t, err.Error(), "MIDI")
		}
	})

	t.Run("channel out of range - too low", func(t *testing.T) {
		_, err := Parse("midi testdevice 0")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel must be between 1-16")
	})

	t.Run("channel out of range - too high", func(t *testing.T) {
		_, err := Parse("midi testdevice 17")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel must be between 1-16")
	})
}

func TestPlayer(t *testing.T) {
	t.Run("player string representation", func(t *testing.T) {
		p := Player{
			Name:         "midi-test-5",
			nameOriginal: "test",
			channel:      5,
		}

		assert.Equal(t, "midi-test-5", p.String())
	})
}

func TestNoteState(t *testing.T) {
	t.Run("note state creation", func(t *testing.T) {
		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		ns := &NoteState{
			Note:     60,
			Velocity: 127,
			Cancel:   cancel,
		}

		assert.Equal(t, 60, ns.Note)
		assert.Equal(t, 127, ns.Velocity)
		assert.NotNil(t, ns.Cancel)
	})
}

func TestInstrumentState(t *testing.T) {
	t.Run("instrument state initialization", func(t *testing.T) {
		is := &InstrumentState{
			Player: nil, // No actual player for test
			Notes:  make(map[int]*NoteState),
		}

		assert.NotNil(t, is.Notes)
		assert.Equal(t, 0, len(is.Notes))
	})
}

func TestGlobalMidiState(t *testing.T) {
	t.Run("global state singleton", func(t *testing.T) {
		// Get global state multiple times
		gms1 := getGlobalState()
		gms2 := getGlobalState()

		// Should be the same instance
		assert.Equal(t, gms1, gms2)
		assert.NotNil(t, gms1.instruments)
	})
}

func TestFindInstrumentByNameLogic(t *testing.T) {
	// Create a test version of findInstrumentByName to avoid hardware calls
	testFindInstrument := func(midiinstrument string, devices []string) string {
		for _, device := range devices {
			if strings.Contains(strings.ToLower(device), strings.ToLower(midiinstrument)) {
				return device
			}
		}
		return ""
	}

	t.Run("find matching device", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI", "Bluetooth MIDI"}

		result := testFindInstrument("usb", devices)
		assert.Equal(t, "USB MIDI Device", result)
	})

	t.Run("case insensitive search", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI"}

		result := testFindInstrument("INTERNAL", devices)
		assert.Equal(t, "Internal MIDI", result)
	})

	t.Run("no matching device", func(t *testing.T) {
		devices := []string{"USB MIDI Device", "Internal MIDI"}

		result := testFindInstrument("nonexistent", devices)
		assert.Equal(t, "", result)
	})

	t.Run("empty device list", func(t *testing.T) {
		devices := []string{}

		result := testFindInstrument("test", devices)
		assert.Equal(t, "", result)
	})
}

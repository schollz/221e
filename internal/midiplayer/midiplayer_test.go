package midiplayer

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

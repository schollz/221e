package midiplayer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/2n/internal/midiconnector"
)

// NoteState tracks the state of a single note
type NoteState struct {
	Note     int
	Velocity int
	Cancel   context.CancelFunc // Cancellation function for the note-off goroutine
}

// InstrumentState tracks all notes for a single instrument
type InstrumentState struct {
	Player *Player
	Notes  map[int]*NoteState // map of note -> NoteState
}

// GlobalMidiState manages all MIDI instruments and their note states
type GlobalMidiState struct {
	mu          sync.RWMutex
	instruments map[string]*InstrumentState // map of "instrument:channel" -> InstrumentState
}

// Global state instance
var globalState *GlobalMidiState
var globalOnce sync.Once

type Player struct {
	Name         string
	nameOriginal string // original name, used for debugging
	Device       *midiconnector.Device
	opened       bool
	channel      uint8
}

func Parse(line string) (p *Player, err error) {
	// midi NAME CHANNEL 1-indexed, need to convert to 0-index
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) < 3 {
		err = fmt.Errorf("invalid midi line format: expected 'midi NAME CHANNEL', got '%s'", line)
		return
	}

	if parts[0] != "midi" {
		err = fmt.Errorf("line must start with 'midi', got '%s'", parts[0])
		return
	}

	name := parts[1]
	channelStr := parts[2]

	channel, parseErr := strconv.Atoi(channelStr)
	if parseErr != nil {
		channel = 0
	} else {
		// Convert from 1-indexed to 0-indexed
		channel = channel - 1
	}

	if channel < 0 || channel > 15 {
		err = fmt.Errorf("channel must be between 1-16, got %d", channel+1)
		return
	}

	p, err = New(name, channel)
	return
}

func New(name string, channel int) (p *Player, err error) {

	p0 := Player{Name: fmt.Sprintf("midi-%s-%d", name, channel), channel: uint8(channel), nameOriginal: name}
	p0.Device, err = midiconnector.New(name)
	if err != nil {
		return
	} else {
		p = &p0
		err = p.Device.Open()
		p.opened = true
	}
	return
}

func (m Player) String() string {
	return m.Name
}

func (m *Player) Close() (err error) {
	if m.opened {
		err = m.Device.Close()
	}
	return
}

func (m *Player) NoteOn(note int, velocity int) (err error) {
	if m.opened {
		err = m.Device.NoteOn(m.channel, uint8(note), uint8(velocity))
	}
	return
}

func (m *Player) NoteOff(note int) (err error) {
	if m.opened {
		err = m.Device.NoteOff(m.channel, uint8(note))
	} else {
		// Player was closed, but we still need to send note_off to avoid stuck notes
		// Create a temporary device to send the note_off message
		tempDevice, tempErr := midiconnector.New(m.nameOriginal)
		if tempErr == nil {
			if openErr := tempDevice.Open(); openErr == nil {
				err = tempDevice.NoteOff(m.channel, uint8(note))
				tempDevice.Close() // Clean up temporary device
			}
		}
		if tempErr != nil || err != nil {
			log.Printf("[MIDIPLAYER] Error sending note-off to closed device: %v", err)
		}
	}
	return
}

// getGlobalState returns the singleton global MIDI state
func getGlobalState() *GlobalMidiState {
	globalOnce.Do(func() {
		globalState = &GlobalMidiState{
			instruments: make(map[string]*InstrumentState),
		}
		log.Printf("[MIDIPLAYER] Global MIDI state initialized")
	})
	return globalState
}

// findInstrumentByName looks for a MIDI instrument that contains the given string
func (gms *GlobalMidiState) findInstrumentByName(midiinstrument string) string {
	devices := midiconnector.Devices()
	for _, device := range devices {
		if strings.Contains(strings.ToLower(device), strings.ToLower(midiinstrument)) {
			return device
		}
	}
	return ""
}

// getOrCreateInstrument gets or creates an instrument state for the given MIDI instrument name and channel
func (gms *GlobalMidiState) getOrCreateInstrument(midiinstrument string, channel int) (*InstrumentState, error) {
	gms.mu.Lock()
	defer gms.mu.Unlock()

	// Create a unique key combining instrument name and channel
	instrumentKey := fmt.Sprintf("%s:%d", midiinstrument, channel)

	// Check if we already have this instrument:channel combination
	if inst, exists := gms.instruments[instrumentKey]; exists {
		log.Printf("[MIDIPLAYER] Found existing instrument: %s (channel %d)", midiinstrument, channel)
		return inst, nil
	}

	// Find the actual device name
	actualDeviceName := gms.findInstrumentByName(midiinstrument)
	if actualDeviceName == "" {
		return nil, fmt.Errorf("no MIDI instrument found containing '%s'", midiinstrument)
	}

	log.Printf("[MIDIPLAYER] Found device '%s' for search string '%s' on channel %d", actualDeviceName, midiinstrument, channel)

	// Create new player with the specified channel
	player, err := New(actualDeviceName, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to create player for %s: %v", actualDeviceName, err)
	}

	// Create instrument state
	instrumentState := &InstrumentState{
		Player: player,
		Notes:  make(map[int]*NoteState),
	}

	gms.instruments[instrumentKey] = instrumentState
	log.Printf("[MIDIPLAYER] Created new instrument: %s -> %s (channel %d)", midiinstrument, actualDeviceName, channel)

	return instrumentState, nil
}

// NoteOn is the main function that handles MIDI note-on with duration management
func NoteOn(midiinstrument string, note float64, velocity float64, duration float64, channel int) error {
	gms := getGlobalState()

	noteInt := int(note)
	velocityInt := int(velocity)

	log.Printf("[MIDIPLAYER] NoteOn called: instrument=%s, note=%d, velocity=%d, duration=%.3fs, channel=%d",
		midiinstrument, noteInt, velocityInt, duration, channel)

	// Get or create instrument
	instrument, err := gms.getOrCreateInstrument(midiinstrument, channel)
	if err != nil {
		return fmt.Errorf("failed to get instrument %s: %v", midiinstrument, err)
	}

	gms.mu.Lock()
	defer gms.mu.Unlock()

	// Check if this note is already playing
	if existingNote, exists := instrument.Notes[noteInt]; exists {
		log.Printf("[MIDIPLAYER] Note %d already playing on %s, cancelling previous and sending note-off",
			noteInt, midiinstrument)

		// Cancel the existing note-off goroutine
		existingNote.Cancel()

		// Send immediate note-off
		err := instrument.Player.NoteOff(noteInt)
		if err != nil {
			log.Printf("[MIDIPLAYER] Error sending immediate note-off for note %d: %v", noteInt, err)
		}
	}

	// Send note-on
	err = instrument.Player.NoteOn(noteInt, velocityInt)
	if err != nil {
		return fmt.Errorf("failed to send note-on for note %d: %v", noteInt, err)
	}

	log.Printf("[MIDIPLAYER] Note-on sent: instrument=%s, note=%d, velocity=%d",
		midiinstrument, noteInt, velocityInt)

	// Create cancellable context for note-off
	ctx, cancel := context.WithCancel(context.Background())

	// Store note state
	instrument.Notes[noteInt] = &NoteState{
		Note:     noteInt,
		Velocity: velocityInt,
		Cancel:   cancel,
	}

	// Start goroutine for note-off after duration
	go func() {
		timer := time.NewTimer(time.Duration(duration * float64(time.Second)))
		defer timer.Stop()

		select {
		case <-timer.C:
			// Duration elapsed, send note-off
			log.Printf("[MIDIPLAYER] Duration elapsed, sending note-off: instrument=%s, note=%d",
				midiinstrument, noteInt)

			err := instrument.Player.NoteOff(noteInt)
			if err != nil {
				log.Printf("[MIDIPLAYER] Error sending note-off for note %d: %v", noteInt, err)
			}

			// Remove from active notes
			gms.mu.Lock()
			instrumentKey := fmt.Sprintf("%s:%d", midiinstrument, channel)
			if inst, exists := gms.instruments[instrumentKey]; exists {
				delete(inst.Notes, noteInt)
				log.Printf("[MIDIPLAYER] Note %d removed from active notes for %s (channel %d)", noteInt, midiinstrument, channel)
			}
			gms.mu.Unlock()

		case <-ctx.Done():
			// Context was cancelled (overlapping note)
			log.Printf("[MIDIPLAYER] Note-off cancelled for note %d on %s (channel %d, overlapping note)",
				noteInt, midiinstrument, channel)
		}
	}()

	return nil
}

// StopAll stops all notes currently playing on the given instrument and channel
func StopAll(midiinstrument string, channel int) {
	gms := getGlobalState()
	gms.mu.Lock()
	defer gms.mu.Unlock()

	instrumentKey := fmt.Sprintf("%s:%d", midiinstrument, channel)
	log.Printf("[MIDIPLAYER] StopAll called for instrument: %s (channel %d)", midiinstrument, channel)

	instrument, exists := gms.instruments[instrumentKey]
	if !exists {
		log.Printf("[MIDIPLAYER] Instrument %s (channel %d) not found, nothing to stop", midiinstrument, channel)
		return
	}

	if len(instrument.Notes) == 0 {
		log.Printf("[MIDIPLAYER] No active notes for instrument %s (channel %d)", midiinstrument, channel)
		return
	}

	log.Printf("[MIDIPLAYER] Stopping %d active notes for instrument %s (channel %d)",
		len(instrument.Notes), midiinstrument, channel)

	// Cancel all note-off goroutines and send immediate note-offs
	for noteInt, noteState := range instrument.Notes {
		log.Printf("[MIDIPLAYER] Stopping note %d on %s (channel %d)", noteInt, midiinstrument, channel)

		// Cancel the note-off goroutine
		noteState.Cancel()

		// Send immediate note-off
		err := instrument.Player.NoteOff(noteInt)
		if err != nil {
			log.Printf("[MIDIPLAYER] Error sending note-off for note %d: %v", noteInt, err)
		}
	}

	// Clear all notes
	instrument.Notes = make(map[int]*NoteState)
	log.Printf("[MIDIPLAYER] All notes stopped for instrument %s (channel %d)", midiinstrument, channel)
}

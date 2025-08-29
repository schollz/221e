package midiplayer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/schollz/2n/internal/midiconnector"
)

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
		}
	}
	return
}

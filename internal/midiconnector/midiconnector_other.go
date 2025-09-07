//go:build !windows

package midiconnector

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var mutex sync.Mutex

var devicesOpen map[string]drivers.Out

func init() {
	devicesOpen = make(map[string]drivers.Out)
}

type Device struct {
	name    string
	num     int
	notesOn map[uint8]uint8
}

func filterName(name string) (foundName string, foundNum int, err error) {
	names := Devices()
	
	// Truncate name to first 3 words
	words := strings.Fields(name)
	if len(words) > 3 {
		words = words[:3]
	}
	truncatedName := strings.Join(words, " ")
	
	// First try exact match with truncated name
	for i, n := range names {
		if strings.EqualFold(n, truncatedName) {
			foundName = n
			foundNum = i
			return
		}
	}
	
	// Then try prefix match with truncated name
	for i, n := range names {
		if strings.HasPrefix(strings.ToLower(n), strings.ToLower(truncatedName)) {
			foundName = n
			foundNum = i
			return
		}
	}
	
	// Finally try contains match for backward compatibility
	for i, n := range names {
		if strings.Contains(strings.ToLower(n), strings.ToLower(truncatedName)) {
			foundName = n
			foundNum = i
			return
		}
	}
	
	if foundNum == -1 {
		err = fmt.Errorf("could not find device with name %s", truncatedName)
	}
	return
}

func New(name string) (*Device, error) {
	var d Device
	var err error
	d.name, d.num, err = filterName(name)
	d.notesOn = make(map[uint8]uint8)
	return &d, err
}

func Close() {
	mutex.Lock()
	defer mutex.Unlock()
	for _, out := range devicesOpen {
		out.Close()
	}
}

func (d *Device) Open() (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := devicesOpen[d.name]; ok {
		return
	}
	out, err := midi.FindOutPort(d.name)
	if err == nil {
		devicesOpen[d.name] = out
		err = out.Open()
	}
	if err == nil {
	} else {
	}
	return
}

func (d *Device) Close() (err error) {
	// send note off to every note
	for note, ch := range d.notesOn {
		d.NoteOff(ch, note)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if out, ok := devicesOpen[d.name]; ok {
		err = out.Close()
		delete(devicesOpen, d.name)
	}
	return
}

func (d *Device) NoteOn(channel, note, velocity uint8) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if out, ok := devicesOpen[d.name]; ok {
		err = out.Send([]byte{0x90 | channel, note, velocity})
		if err != nil {
			// Log MIDI errors instead of letting them print to stderr
			log.Printf("MIDI NoteOn error for device %s: %v", d.name, err)
		} else {
			d.notesOn[note] = channel
		}
	}
	return
}

func (d *Device) NoteOff(channel, note uint8) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if out, ok := devicesOpen[d.name]; ok {
		err = out.Send([]byte{0x80 | channel, note, 0})
		if err != nil {
			// Log MIDI errors instead of letting them print to stderr
			log.Printf("MIDI NoteOff error for device %s: %v", d.name, err)
		} else {
			delete(d.notesOn, note)
		}
	}
	return
}

func Devices() (devices []string) {
	outs := midi.GetOutPorts()
	for _, out := range outs {
		devices = append(devices, out.String())
	}
	return
}

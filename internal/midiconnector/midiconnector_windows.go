//go:build windows

package midiconnector

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var mutex sync.Mutex

type Device struct {
	name    string
	num     int
	notesOn map[uint8]bool
}

func filterName(name string) (foundName string, foundNum int, err error) {
	names := Devices()
	foundNum = -1
	for i, n := range names {
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

func New(name string) (*Device, error) {
	var d Device
	var err error
	d.name, d.num, err = filterName(name)
	d.notesOn = make(map[uint8]bool)
	return &d, err
}

func Close() {

}

func (d *Device) Open() (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := devicesOpen[d.name]; ok {
		return
	}
	// Open the first MIDI output device
	var hmo HMIDIOUT
	if midiOutOpen(&hmo, uint32(d.num), 0, 0, 0) != 0 { // Change the second argument to select a different device
		err = fmt.Errorf("failed to open MIDI output device: %s", d.name)
		return
	}
	devicesOpen[d.name] = hmo
	return
}

func (d *Device) Close() (err error) {
	for note := range d.notesOn {
		d.NoteOff(0, note)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if hmo, ok := devicesOpen[d.name]; ok {
		midiOutClose(hmo)
		delete(devicesOpen, d.name)
	}
	return
}

func (d *Device) NoteOn(channel, note, velocity uint8) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if hmo, ok := devicesOpen[d.name]; ok {
		if sendNoteOn(hmo, channel, note, velocity) != 0 {
			err = fmt.Errorf("failed to send Note On message")
		} else {
			d.notesOn[note] = true
		}
	}
	return
}

func (d *Device) NoteOff(channel, note uint8) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if hmo, ok := devicesOpen[d.name]; ok {
		noteOff := uint32(MIDI_NOTE_ON | channel) // MIDI_NOTE_OFF can be used instead of MIDI_NOTE_ON with velocity zero
		noteOff |= uint32(note) << 8
		noteOff |= uint32(0) << 16 // Set velocity to 0 for Note Off
		if midiOutShortMsg(hmo, noteOff) != 0 {
			err = fmt.Errorf("failed to send Note Off message")
		} else {
			delete(d.notesOn, note)
		}
	}
	return
}

// Constants
const (
	MAXPNAMELEN  = 32
	MIDI_NOTE_ON = 0x90
)

// Structures
type MIDIOUTCAPS struct {
	WMid           uint16
	WPid           uint16
	VDriverVersion uint32
	SPname         [MAXPNAMELEN]byte
	XSzGuid        [16]byte
}

// Load winmm.dll and get procedures
var (
	winmm                  = syscall.NewLazyDLL("winmm.dll")
	procMidiOutGetNumDevs  = winmm.NewProc("midiOutGetNumDevs")
	procMidiOutGetDevCapsA = winmm.NewProc("midiOutGetDevCapsA")
	procMidiOutOpen        = winmm.NewProc("midiOutOpen")
	procMidiOutClose       = winmm.NewProc("midiOutClose")
	procMidiOutShortMsg    = winmm.NewProc("midiOutShortMsg")
)

var devicesOpen map[string]HMIDIOUT

func init() {
	devicesOpen = make(map[string]HMIDIOUT)
}

type HMIDIOUT uintptr

// Wrapper functions
func midiOutGetNumDevs() uint32 {
	ret, _, _ := procMidiOutGetNumDevs.Call()
	return uint32(ret)
}

func midiOutGetDevCapsA(uDeviceID uint32, pmoc *MIDIOUTCAPS, cbmoc uint32) uint32 {
	ret, _, _ := procMidiOutGetDevCapsA.Call(
		uintptr(uDeviceID),
		uintptr(unsafe.Pointer(pmoc)),
		uintptr(cbmoc),
	)
	return uint32(ret)
}

func midiOutOpen(phmo *HMIDIOUT, uDeviceID uint32, dwCallback, dwInstance uintptr, dwFlags uint32) uint32 {
	ret, _, _ := procMidiOutOpen.Call(
		uintptr(unsafe.Pointer(phmo)),
		uintptr(uDeviceID),
		dwCallback,
		dwInstance,
		uintptr(dwFlags),
	)
	return uint32(ret)
}

func midiOutClose(hmo HMIDIOUT) uint32 {
	ret, _, _ := procMidiOutClose.Call(uintptr(hmo))
	return uint32(ret)
}

func midiOutShortMsg(hmo HMIDIOUT, dwMsg uint32) uint32 {
	ret, _, _ := procMidiOutShortMsg.Call(uintptr(hmo), uintptr(dwMsg))
	return uint32(ret)
}

// Function to send a Note On message
func sendNoteOn(hmo HMIDIOUT, channel, note, velocity uint8) uint32 {
	message := uint32(MIDI_NOTE_ON | channel)
	message |= uint32(note) << 8
	message |= uint32(velocity) << 16
	return midiOutShortMsg(hmo, message)
}

func Devices() []string {
	numOutDevs := midiOutGetNumDevs()
	names := make([]string, numOutDevs)
	for i := uint32(0); i < numOutDevs; i++ {
		var outCaps MIDIOUTCAPS
		if midiOutGetDevCapsA(i, &outCaps, uint32(unsafe.Sizeof(outCaps))) == 0 {
			names[i] = string(outCaps.SPname[:])
		} else {
			names[i] = "Failed to get capabilities"
		}
	}
	return names
}

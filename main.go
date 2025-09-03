package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"

	"github.com/schollz/2n/internal/input"
	"github.com/schollz/2n/internal/midiconnector"
	"github.com/schollz/2n/internal/model"
	"github.com/schollz/2n/internal/storage"
	"github.com/schollz/2n/internal/supercollider"
	"github.com/schollz/2n/internal/types"
	"github.com/schollz/2n/internal/views"
)

type scReadyMsg struct{}

func main() {
	// Redirect stderr to discard to suppress C library error messages like "MidiOutAlsa::sendMessage: incomplete message!"
	// This prevents MIDI library errors from interrupting the TUI display
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		// Save original stderr for potential restoration
		originalStderr := os.Stderr
		os.Stderr = devNull
		defer func() {
			devNull.Close()
			os.Stderr = originalStderr
		}()
	}

	log.SetOutput(io.Discard)

	// Start CPU profiling for the first 30 seconds
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Printf("Could not create CPU profile: %v", err)
	} else {
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			log.Printf("Could not start CPU profile: %v", err)
		} else {
			go func() {
				time.Sleep(30 * time.Second)
				pprof.StopCPUProfile()
				cpuFile.Close()
				log.Println("CPU profiling stopped after 30 seconds")
			}()
		}
	}

	// Set up cleanup on exit
	setupCleanupOnExit()

	// Parse command line arguments (no no-splash anymore)
	var oscPort int
	var skipJackCheck bool
	var saveFile string
	var debugLog string
	flag.IntVar(&oscPort, "osc-port", 57120, "OSC port for sending playback messages")
	flag.StringVar(&saveFile, "save-file", "tracker-save.json", "Save file to load from or create")
	flag.BoolVar(&skipJackCheck, "skip-jack-check", false, "Skip checking for JACK server (for testing only)")
	flag.StringVar(&debugLog, "debug", "", "If set, write debug logs to this file; empty disables logging")
	flag.Parse()

	if !supercollider.IsJackEnabled() && !skipJackCheck {
		dialog := supercollider.NewJackDialogModel()
		p := tea.NewProgram(dialog, tea.WithAltScreen())
		_, _ = p.Run()
		os.Exit(1)
	}

	// Check for required SuperCollider extensions before starting
	if !supercollider.HasRequiredExtensions() {
		dialog := supercollider.NewInstallDialogModel()
		p := tea.NewProgram(dialog, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			log.Printf("Error running install dialog: %v", err)
			os.Exit(1)
		}

		if result, ok := finalModel.(supercollider.InstallDialogModel); ok {
			if !result.ShouldInstall() {
				os.Exit(1)
			}
			if result.Error() != nil {
				log.Printf("Failed to install SuperCollider extensions: %v", result.Error())
				os.Exit(1)
			}
		} else {
			log.Printf("Unexpected model type returned from install dialog")
			os.Exit(1)
		}
	}

	// Set up debug logging early
	if debugLog != "" {
		f, err := tea.LogToFile(debugLog, "debug")
		if err != nil {
			log.Printf("Fatal: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		// Set log flags to include file and line number for VS Code clickable links
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Println("Debug logging enabled")
	log.Printf("OSC port configured: %d", oscPort)

	// Create readiness channel for SuperCollider startup detection
	readyChannel := make(chan struct{}, 1)

	// Set up OSC dispatcher early to detect SuperCollider readiness
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/cpuusage", func(msg *osc.Message) {
		// log.Printf("SuperCollider CPU Usage: %v", msg.Arguments[0])
		// Signal that SuperCollider is ready (non-blocking)
		select {
		case readyChannel <- struct{}{}:
		default:
		}
	})
	var tm *TrackerModel // Will be set after model creation

	d.AddMsgHandler("/track_volume", func(msg *osc.Message) {
		// argument 0: track number (0 to 7)
		// argument 1: volume in dB (floating point)
		trackNum := int(msg.Arguments[0].(float32))
		volume := msg.Arguments[1].(float32)

		// Store the track volume in the model if available
		if tm != nil && trackNum >= 0 && trackNum < 8 {
			tm.model.TrackVolumes[trackNum] = volume
		}
	})
	// Start OSC server early
	server := &osc.Server{Addr: fmt.Sprintf(":%d", oscPort+1), Dispatcher: d}
	go func() {
		log.Printf("Starting OSC server on port %d", oscPort+1)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting OSC server: %v", err)
		}
	}()

	// Build program
	tm = initialModel(oscPort, saveFile, d)

	p := tea.NewProgram(tm, tea.WithAltScreen())

	// Start SuperCollider in the background so it doesn't block the splash
	if supercollider.IsJackEnabled() {
		go func() {
			if !supercollider.IsSuperColliderEnabled() {
				if err := supercollider.StartSuperCollider(); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			}
		}()
	} else {
		if !skipJackCheck {
			log.Printf("JACK server not enabled; cannot start SuperCollider")
			os.Exit(1)
		}
	}

	// When SC signals readiness via /cpuusage, hide the splash
	go func() {
		if skipJackCheck {
			p.Send(scReadyMsg{}) // skip splash if skipping JACK check
		} else {
			<-readyChannel
			log.Printf("Received SuperCollider ready; hiding splash")
			p.Send(scReadyMsg{})
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Printf("Error: %v", err)
	}

	// Always call cleanup when the program exits normally (e.g., Ctrl+Q)
	supercollider.Cleanup()
}

func initialModel(oscPort int, saveFile string, dispatcher *osc.StandardDispatcher) *TrackerModel {
	m := model.NewModel(oscPort, saveFile)

	// Try to load saved state
	if err := storage.LoadState(m, oscPort, saveFile); err == nil {
		log.Printf("Loaded saved state successfully from %s", saveFile)
	} else {
		log.Printf("No saved state found or error loading from %s: %v", saveFile, err)
		// Load files for new model
		storage.LoadFiles(m)
	}

	// Send current dB settings to OSC on startup
	m.SendOSCPregainMessage()
	m.SendOSCPostgainMessage()
	m.SendOSCBiasMessage()
	m.SendOSCSaturationMessage()
	m.SendOSCDriveMessage()

	// Send track set levels to OSC on startup
	for track := 0; track < 8; track++ {
		m.SendOSCTrackSetLevelMessage(track)
	}

	// Add waveform handler to the existing OSC dispatcher
	dispatcher.AddMsgHandler("/waveform", func(msg *osc.Message) {
		sample := float64(msg.Arguments[0].(float32)) // expected in [-1,+1]
		m.LastWaveform = sample
		// available content width inside the padded container (2 spaces each side)
		maxCols := m.TermWidth - 4
		if maxCols < 1 {
			maxCols = 1
		}
		m.PushWaveformSample(sample, maxCols*2/3)
	})

	m.AvailableMidiDevices = midiconnector.Devices()
	for _, device := range m.AvailableMidiDevices {
		log.Printf("MIDI device found: %+v", device)
	}

	// Set default MIDI device to first available device
	if len(m.AvailableMidiDevices) > 0 {
		firstDevice := m.AvailableMidiDevices[0]
		// Update all MIDI settings to use the first device by default
		for i := 0; i < 255; i++ {
			m.MidiSettings[i].Device = firstDevice
			// Channel is already set to "1" by default in initializeDefaultData()
		}
		log.Printf("Default MIDI device set to: %s", firstDevice)
	}

	return &TrackerModel{
		model:         m,
		splashState:   views.NewSplashState(3 * time.Second),
		showingSplash: true, // splash is ALWAYS shown until SC ready
	}
}

// TrackerModel wraps the model and implements the tea.Model interface
type TrackerModel struct {
	model         *model.Model
	splashState   *views.SplashState
	showingSplash bool
}

// WaveformTickMsg is a special message that fires at a steady UI rate (30fps)
// to refresh/redraw waveform and UI without advancing playback.
type WaveformTickMsg struct{}

// SplashTickMsg drives the splash screen animation
type SplashTickMsg struct{}

// tickWaveform schedules the next WaveformTickMsg at the requested fps.
func tickWaveform(fps int) tea.Cmd {
	if fps <= 0 {
		fps = 30
	}
	interval := time.Second / time.Duration(fps)
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return WaveformTickMsg{}
	})
}

// tickSplash schedules the next SplashTickMsg for smooth animation
func tickSplash() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return SplashTickMsg{}
	})
}

func (tm *TrackerModel) Init() tea.Cmd {
	if tm.showingSplash {
		// Start splash screen animation at 60fps
		return tickSplash()
	}
	// Start a 30fps UI loop so the waveform redraws smoothly.
	// Playback advancement stays on its own schedule (input.TickMsg).
	return tickWaveform(30)
}

func (tm *TrackerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tm.model.TermHeight = msg.Height
		tm.model.TermWidth = msg.Width
		// keep the appropriate loop going
		if tm.showingSplash {
			return tm, nil
		}
		return tm, nil

	case SplashTickMsg:
		// Keep animating the splash; do NOT auto-dismiss on duration.
		// We'll exit the splash only on scReadyMsg or a keypress.
		return tm, tickSplash()

	case WaveformTickMsg:
		// Redraw UI/waveform at 30fps. Do NOT advance playback here.
		// Reschedule the next UI tick.
		if tm.showingSplash {
			return tm, nil
		}
		return tm, tickWaveform(30)

	case input.TickMsg:
		// Tempo/engine ticks: only advance playback here, at your musical rate.
		if tm.model.IsPlaying {
			input.AdvancePlayback(tm.model)
			// Reschedule the next tempo tick according to your input package.
			return tm, input.Tick(tm.model)
		}
		return tm, nil

	case scReadyMsg:
		// SC is ready — leave the splash screen
		tm.showingSplash = false
		return tm, nil

	case tea.KeyMsg:
		// Skip splash screen on any key press
		if tm.showingSplash {
			tm.showingSplash = false
			return tm, tickWaveform(30)
		}
		// Keys may toggle playback, change views, etc.
		return tm, input.HandleKeyInput(tm.model, msg)
	}

	return tm, nil
}

func (tm TrackerModel) View() string {
	if tm.showingSplash {
		return views.RenderSplashScreen(tm.model.TermWidth, tm.model.TermHeight, tm.splashState)
	}

	switch tm.model.ViewMode {
	case types.SongView:
		return views.RenderSongView(tm.model)
	case types.ChainView:
		return views.RenderChainView(tm.model)
	case types.PhraseView:
		return views.RenderPhraseView(tm.model)
	case types.SettingsView:
		return views.RenderSettingsView(tm.model)
	case types.FileMetadataView:
		return views.RenderFileMetadataView(tm.model)
	case types.RetriggerView:
		return views.RenderRetriggerView(tm.model)
	case types.TimestrechView:
		return views.RenderTimestrechView(tm.model)
	case types.ArpeggioView:
		return views.RenderArpeggioView(tm.model)
	case types.MidiView:
		return views.RenderMidiView(tm.model)
	case types.SoundMakerView:
		return views.RenderSoundMakerView(tm.model)
	case types.MixerView:
		return views.RenderMixerView(tm.model)
	default: // FileView
		return views.RenderFileView(tm.model)
	}
}

func setupCleanupOnExit() {
	// Handle cleanup on various exit signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-c
		supercollider.Cleanup()
		os.Exit(0)
	}()
}

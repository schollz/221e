package main

import (
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
	"github.com/spf13/cobra"

	"github.com/schollz/collidertracker/internal/hacks"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/midiconnector"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/project"
	"github.com/schollz/collidertracker/internal/sox"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/supercollider"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/schollz/collidertracker/internal/views"
)

var (
	Version = "dev"

	// Command-line configuration
	config struct {
		port            int
		project         string
		projectProvided bool // Track if --project flag was explicitly provided
		record          bool
		debug           string
		skipJack        bool
	}
)

type scReadyMsg struct{}

var rootCmd = &cobra.Command{
	Use:   "collidertracker",
	Short: "A modern music tracker for SuperCollider",
	Long: `ColliderTracker is a modern, terminal-based music tracker that integrates 
with SuperCollider for real-time audio synthesis and sampling.

Features:
• Real-time audio synthesis with SuperCollider
• Sample-based music composition
• MIDI integration
• Live audio recording and playback
• Retrigger and time-stretch effects`,
	Version: Version,
	Run:     runColliderTracker,
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&config.port, "port", "p", 57120,
		"OSC port for SuperCollider communication")
	rootCmd.PersistentFlags().StringVar(&config.project, "project", "save",
		"Project directory for songs and audio files")
	rootCmd.PersistentFlags().BoolVar(&config.record, "record", false,
		"Enable automatic session recording")
	rootCmd.PersistentFlags().StringVar(&config.debug, "log", "",
		"Write debug logs to specified file (empty disables)")
	rootCmd.PersistentFlags().BoolVar(&config.skipJack, "skip-jack", false,
		"Skip JACK server verification (for testing)")
	
	// Set up a callback to track when --project is explicitly provided
	rootCmd.PersistentFlags().Lookup("project").Changed = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func restartWithProject() {
	// This function restarts the ColliderTracker with the new project
	// without going through cobra command parsing again
	
	// Check JACK and SuperCollider requirements (same as in runColliderTracker)
	if !supercollider.IsJackEnabled() && !config.skipJack {
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
	if config.debug != "" {
		f, err := tea.LogToFile(config.debug, "debug")
		if err != nil {
			log.Printf("Fatal: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		// Set log flags to include file and line number for VS Code clickable links
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		// send log to io.Discard
		log.SetOutput(io.Discard)
	}

	log.Println("Debug logging enabled")
	log.Printf("OSC port configured: %d", config.port)

	// Create readiness channel for SuperCollider startup detection
	readyChannel := make(chan struct{}, 1)

	// Set up OSC dispatcher early to detect SuperCollider readiness
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/cpuusage", func(msg *osc.Message) {
		log.Printf("SuperCollider CPU Usage: %v", msg.Arguments[0])
		// Signal that SuperCollider is ready (non-blocking)
		select {
		case readyChannel <- struct{}{}:
		default:
		}
	})
	var tm *TrackerModel // Will be set after model creation

	d.AddMsgHandler("/track_volume", func(msg *osc.Message) {
		for i := 0; i < len(tm.model.TrackVolumes); i++ {
			tm.model.TrackVolumes[i] = msg.Arguments[i].(float32)
		}
	})
	// Build program
	tm = initialModel(config.port, config.project, d)

	p := tea.NewProgram(tm, tea.WithAltScreen())

	// Start OSC server after p is created but before p.Run()
	server := &osc.Server{Addr: fmt.Sprintf(":%d", config.port+1), Dispatcher: d}
	go func() {
		log.Printf("Starting OSC server on port %d", config.port+1)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting OSC server: %v", err)
		}
	}()

	// Start SuperCollider in the background so it doesn't block the splash
	// Always check JACK status, but only exit if --skip-jack is not set
	if supercollider.IsJackEnabled() {
		log.Printf("JACK server enabled; starting SuperCollider if not already running")
		go func() {
			if !supercollider.IsSuperColliderEnabled() {
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			}
		}()
	} else {
		// JACK is not running - log this but don't start SuperCollider
		log.Printf("JACK server not enabled; skipping SuperCollider startup")
		if !config.skipJack {
			// Only exit if --skip-jack flag was not provided
			os.Exit(1)
		}
	}

	// When SC signals readiness via /cpuusage, hide the splash
	go func() {
		if config.skipJack {
			p.Send(scReadyMsg{}) // skip splash if skipping JACK check
		} else {
			<-readyChannel
			log.Printf("Received SuperCollider ready; hiding splash")
			p.Send(scReadyMsg{})
		}
	}()

	// Initialize sox
	sox.Init()
	// hack to make sure Ctrl+V works on Windows
	hacks.StoreWinClipboard()

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Check if we should return to project selection again (recursive)
	if finalModel != nil {
		if trackerModel, ok := finalModel.(*TrackerModel); ok && trackerModel.model.ReturnToProjectSelector {
			log.Printf("Returning to project selection...")
			// Clean up current session
			supercollider.Cleanup()
			sox.Clean()
			
			// Run project selector again
			selectedPath, cancelled := project.RunProjectSelector()
			if !cancelled && selectedPath != "" {
				// Update project path and restart
				config.project = selectedPath
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			} else if !cancelled && selectedPath == "" {
				// User chose to create new project, prompt for name
				fmt.Print("Enter project name (or press Enter for 'save'): ")
				var projectName string
				fmt.Scanln(&projectName)
				
				if projectName == "" {
					projectName = "save"
				}
				
				config.project = projectName
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			}
		}
	}

	// Always call cleanup when the program exits normally (e.g., Ctrl+Q)
	supercollider.Cleanup()
	sox.Clean()
}

func runColliderTracker(cmd *cobra.Command, args []string) {
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

	// Check if --project flag was explicitly provided
	config.projectProvided = cmd.PersistentFlags().Changed("project")
	
	// If no project was specified, show project selector
	if !config.projectProvided {
		selectedPath, cancelled := project.RunProjectSelector()
		if cancelled {
			os.Exit(0)
		}
		
		if selectedPath != "" {
			// User selected an existing project
			config.project = selectedPath
		} else {
			// User chose to create new project, prompt for name
			fmt.Print("Enter project name (or press Enter for 'save'): ")
			var projectName string
			fmt.Scanln(&projectName)
			
			if projectName == "" {
				projectName = "save"
			}
			
			config.project = projectName
		}
	}

	if !supercollider.IsJackEnabled() && !config.skipJack {
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
	if config.debug != "" {
		f, err := tea.LogToFile(config.debug, "debug")
		if err != nil {
			log.Printf("Fatal: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		// Set log flags to include file and line number for VS Code clickable links
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		// send log to io.Discard
		log.SetOutput(io.Discard)
	}

	log.Println("Debug logging enabled")
	log.Printf("OSC port configured: %d", config.port)

	// Create readiness channel for SuperCollider startup detection
	readyChannel := make(chan struct{}, 1)

	// Set up OSC dispatcher early to detect SuperCollider readiness
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/cpuusage", func(msg *osc.Message) {
		log.Printf("SuperCollider CPU Usage: %v", msg.Arguments[0])
		// Signal that SuperCollider is ready (non-blocking)
		select {
		case readyChannel <- struct{}{}:
		default:
		}
	})
	var tm *TrackerModel // Will be set after model creation

	d.AddMsgHandler("/track_volume", func(msg *osc.Message) {
		for i := 0; i < len(tm.model.TrackVolumes); i++ {
			tm.model.TrackVolumes[i] = msg.Arguments[i].(float32)
		}
	})
	// Build program
	tm = initialModel(config.port, config.project, d)

	p := tea.NewProgram(tm, tea.WithAltScreen())

	// Start OSC server after p is created but before p.Run()
	server := &osc.Server{Addr: fmt.Sprintf(":%d", config.port+1), Dispatcher: d}
	go func() {
		log.Printf("Starting OSC server on port %d", config.port+1)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting OSC server: %v", err)
		}
	}()

	// Start SuperCollider in the background so it doesn't block the splash
	// Always check JACK status, but only exit if --skip-jack is not set
	if supercollider.IsJackEnabled() {
		log.Printf("JACK server enabled; starting SuperCollider if not already running")
		go func() {
			if !supercollider.IsSuperColliderEnabled() {
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			}
		}()
	} else {
		// JACK is not running - log this but don't start SuperCollider
		log.Printf("JACK server not enabled; skipping SuperCollider startup")
		if !config.skipJack {
			// Only exit if --skip-jack flag was not provided
			os.Exit(1)
		}
	}

	// When SC signals readiness via /cpuusage, hide the splash
	go func() {
		if config.skipJack {
			p.Send(scReadyMsg{}) // skip splash if skipping JACK check
		} else {
			<-readyChannel
			log.Printf("Received SuperCollider ready; hiding splash")
			p.Send(scReadyMsg{})
		}
	}()

	// Initialize sox
	sox.Init()
	// hack to make sure Ctrl+V works on Windows
	hacks.StoreWinClipboard()

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Check if we should return to project selection
	if finalModel != nil {
		if trackerModel, ok := finalModel.(*TrackerModel); ok && trackerModel.model.ReturnToProjectSelector {
			log.Printf("Returning to project selection...")
			// Clean up current session
			supercollider.Cleanup()
			sox.Clean()
			
			// Run project selector again
			selectedPath, cancelled := project.RunProjectSelector()
			if !cancelled && selectedPath != "" {
				// Update project path and restart
				config.project = selectedPath
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			} else if !cancelled && selectedPath == "" {
				// User chose to create new project, prompt for name
				fmt.Print("Enter project name (or press Enter for 'save'): ")
				var projectName string
				fmt.Scanln(&projectName)
				
				if projectName == "" {
					projectName = "save"
				}
				
				config.project = projectName
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			}
		}
	}

	// Always call cleanup when the program exits normally (e.g., Ctrl+Q)
	supercollider.Cleanup()
	sox.Clean()
}

func initialModel(oscPort int, saveFolder string, dispatcher *osc.StandardDispatcher) *TrackerModel {
	m := model.NewModel(oscPort, saveFolder)

	// Try to load saved state
	if err := storage.LoadState(m, oscPort, saveFolder); err == nil {
		log.Printf("Loaded saved state successfully from %s", saveFolder)
	} else {
		log.Printf("No saved state found or error loading from %s: %v", saveFolder, err)
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

	// Set default MIDI device to first available device (only for unset devices)
	if len(m.AvailableMidiDevices) > 0 {
		firstDevice := m.AvailableMidiDevices[0]
		// Only update MIDI settings that are still set to "None" (preserve user selections)
		for i := 0; i < 255; i++ {
			if m.MidiSettings[i].Device == "None" {
				m.MidiSettings[i].Device = firstDevice
				// Channel is already set to "1" by default in initializeDefaultData()
			}
		}
		log.Printf("Default MIDI device set to: %s (for unset devices only)", firstDevice)
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
		return views.RenderSplashScreen(tm.model.TermWidth, tm.model.TermHeight, tm.splashState, Version)
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

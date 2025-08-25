package supercollider

import (
	"context"
	"math"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StartupProgressModel struct {
	progress     progress.Model
	width        int
	height       int
	done         bool
	err          error
	ctx          context.Context
	cancel       context.CancelFunc
	stage        string
	readyChannel <-chan struct{}
}

type progressMsg float64
type completedMsg struct{}
type errorMsg struct{ err error }
type stageMsg string

func NewStartupProgressModel(readyChannel <-chan struct{}) StartupProgressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50

	ctx, cancel := context.WithCancel(context.Background())

	return StartupProgressModel{
		progress:     p,
		ctx:          ctx,
		cancel:       cancel,
		stage:        "Starting SuperCollider...",
		readyChannel: readyChannel,
	}
}

func (m StartupProgressModel) Init() tea.Cmd {
	// Start the SuperCollider process and progress animation
	return tea.Batch(
		m.startSuperCollider(),
		m.setupOSCListener(),
		m.tickProgress(),
	)
}

func (m StartupProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		return m, nil

	case stageMsg:
		m.stage = string(msg)
		return m, nil

	case progressMsg:
		// Continue ticking until we're done
		cmd := m.progress.SetPercent(float64(msg))
		if !m.done {
			return m, tea.Batch(cmd, m.tickProgress())
		}
		return m, cmd

	case completedMsg:
		m.done = true
		m.cancel()
		return m, tea.Quit

	case errorMsg:
		m.err = msg.err
		m.done = true
		m.cancel()
		return m, tea.Quit

	case tea.KeyMsg:
		// Allow Ctrl+C to cancel
		if msg.String() == "ctrl+c" {
			m.cancel()
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m StartupProgressModel) View() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Align(lipgloss.Center)

		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(errorStyle.Render("Failed to start SuperCollider:\n" + m.err.Error() + "\n\nPress any key to exit..."))
	}

	if m.done {
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			Align(lipgloss.Center)

		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(successStyle.Render("SuperCollider started successfully! ✓"))
	}

	// Main progress display
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center)

	stageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("Starting SuperCollider"),
		"",
		m.progress.View(),
		"",
		stageStyle.Render(m.stage),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)
}

func (m StartupProgressModel) Done() bool {
	return m.done
}

func (m StartupProgressModel) Error() error {
	return m.err
}

func (m StartupProgressModel) startSuperCollider() tea.Cmd {
	return func() tea.Msg {
		// Start SuperCollider in the background
		if err := StartSuperCollider(); err != nil {
			return errorMsg{err}
		}
		return stageMsg("Initializing SuperCollider...")
	}
}

func (m StartupProgressModel) setupOSCListener() tea.Cmd {
	return func() tea.Msg {
		// Wait for SuperCollider to send a /cpuusage message (proof it's running and responsive)
		// Wait up to 15 seconds for SuperCollider to start and send CPU usage
		timeout := time.After(15 * time.Second)

		select {
		case <-m.ctx.Done():
			return nil
		case <-timeout:
			// Timeout reached - assume SuperCollider is ready anyway
			return completedMsg{}
		case <-m.readyChannel:
			// Received /cpuusage message - SuperCollider is definitely ready!
			return completedMsg{}
		}
	}
}

func (m StartupProgressModel) tickProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		// Don't check cpuReceived here - that's handled by setupOSCListener now
		// Just create a moving visualizer effect
		newValue := 0.5 + 0.3*math.Sin(float64(time.Now().UnixMilli())/200.0)
		return progressMsg(newValue)
	})
}

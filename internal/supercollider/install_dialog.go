package supercollider

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InstallDialogModel struct {
	width    int
	height   int
	selected int // 0 = Yes, 1 = No
	done     bool
	install  bool
	err      error
}

func NewInstallDialogModel() InstallDialogModel {
	return InstallDialogModel{
		selected: 0,
		done:     false,
	}
}

func (m InstallDialogModel) Init() tea.Cmd {
	return nil
}

func (m InstallDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case downloadCompleteMsg:
		m.err = msg.err
		m.done = true
		return m, nil

	case tea.KeyMsg:
		if m.done && (m.install || m.err != nil) {
			// After installation success/failure, any key exits
			return m, tea.Quit
		}

		switch msg.String() {
		case "left", "h":
			m.selected = 0
		case "right", "l":
			m.selected = 1
		case "enter":
			if m.selected == 0 {
				m.install = true
				return m, m.downloadExtensions
			} else {
				m.done = true
				m.install = false
				return m, tea.Quit
			}
		case "q", "ctrl+c":
			m.done = true
			m.install = false
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m InstallDialogModel) downloadExtensions() tea.Msg {
	err := DownloadRequiredExtensions()
	return downloadCompleteMsg{err: err}
}

type downloadCompleteMsg struct {
	err error
}

func (m InstallDialogModel) View() string {
	if m.done && m.install && m.err == nil {
		return m.renderSuccess()
	}

	if m.err != nil {
		return m.renderError()
	}

	// Calculate dimensions for centered dialog
	dialogWidth := 60

	if m.width > 0 && dialogWidth > m.width-4 {
		dialogWidth = m.width - 4
	}

	// Create the message
	message := "SuperCollider extensions are required but not found.\nWould you like to automatically install them?"

	// Create buttons
	yesButton := " Yes "
	noButton := " No "

	if m.selected == 0 {
		yesButton = "> Yes <"
	} else {
		noButton = "> No <"
	}

	// Style definitions
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(dialogWidth - 4)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Margin(0, 1)

	selectedButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("205")).
		Padding(0, 1).
		Margin(0, 1)

	// Apply button styles
	if m.selected == 0 {
		yesButton = selectedButtonStyle.Render(yesButton)
		noButton = buttonStyle.Render(noButton)
	} else {
		yesButton = buttonStyle.Render(yesButton)
		noButton = selectedButtonStyle.Render(noButton)
	}

	// Combine buttons
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesButton, noButton)

	// Create dialog content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		message,
		"",
		buttons,
	)

	dialog := dialogStyle.Render(content)

	// Center the dialog on screen
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(dialog)
}

func (m InstallDialogModel) renderSuccess() string {
	message := "SuperCollider extensions installed successfully!\nPress any key to continue..."

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("46")).
		Foreground(lipgloss.Color("46")).
		Padding(1, 2).
		Width(50)

	dialog := style.Render(message)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(dialog)
}

func (m InstallDialogModel) renderError() string {
	message := fmt.Sprintf("Failed to install SuperCollider extensions:\n%v\n\nPress any key to exit...", m.err)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Foreground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(60)

	dialog := style.Render(message)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(dialog)
}

func (m InstallDialogModel) Done() bool {
	return m.done
}

func (m InstallDialogModel) ShouldInstall() bool {
	return m.install
}

func (m InstallDialogModel) Error() error {
	return m.err
}

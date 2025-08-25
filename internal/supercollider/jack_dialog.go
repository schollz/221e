package supercollider

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type JackDialogModel struct {
	width  int
	height int
	done   bool
}

func NewJackDialogModel() JackDialogModel {
	return JackDialogModel{}
}

func (m JackDialogModel) Init() tea.Cmd { return nil }

func (m JackDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// Any key confirms and exits
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}
func (m JackDialogModel) View() string {
	message := "JACK server not detected.\n\nPlease start the JACK audio server and try again.\n\nPress any key to exit."

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Foreground(lipgloss.Color("15")).
		Padding(1, 2).
		Width(40).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	dialog := style.Render(message)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(dialog)
}

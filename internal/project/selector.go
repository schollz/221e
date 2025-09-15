package project

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Project represents a found project
type Project struct {
	Name     string    // Display name (usually the folder name)
	Path     string    // Full path to the project folder
	Modified time.Time // Last modification time of data.json.gz
}

// ProjectSelector is the bubbletea model for project selection
type ProjectSelector struct {
	projects       []Project
	selectedIndex  int
	searchComplete bool
	searching      bool
	width          int
	height         int
}

type searchCompleteMsg struct {
	projects []Project
	err      error
}

// ProjectNameInput is a dialog for entering a new project name
type ProjectNameInput struct {
	projectName string
	cursor      int
	width       int
	height      int
	done        bool
	cancelled   bool
}

// NewProjectNameInput creates a new project name input dialog
func NewProjectNameInput() *ProjectNameInput {
	return &ProjectNameInput{
		projectName: "",
		cursor:      0,
	}
}

// Init is required for tea.Model interface
func (pni *ProjectNameInput) Init() tea.Cmd {
	return nil
}

// Update handles messages for project name input
func (pni *ProjectNameInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pni.width = msg.Width
		pni.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "ctrl+q":
			pni.cancelled = true
			pni.done = true
			return pni, tea.Quit

		case "enter":
			// If empty, use default name "save"
			if strings.TrimSpace(pni.projectName) == "" {
				pni.projectName = "save"
			}
			pni.done = true
			return pni, tea.Quit

		case "backspace":
			if len(pni.projectName) > 0 && pni.cursor > 0 {
				pni.projectName = pni.projectName[:pni.cursor-1] + pni.projectName[pni.cursor:]
				pni.cursor--
			}

		case "left":
			if pni.cursor > 0 {
				pni.cursor--
			}

		case "right":
			if pni.cursor < len(pni.projectName) {
				pni.cursor++
			}

		case "home":
			pni.cursor = 0

		case "end":
			pni.cursor = len(pni.projectName)

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				char := msg.String()
				// Allow alphanumeric characters, spaces, dashes, and underscores
				if (char >= "a" && char <= "z") || (char >= "A" && char <= "Z") ||
					(char >= "0" && char <= "9") || char == " " || char == "-" || char == "_" {
					pni.projectName = pni.projectName[:pni.cursor] + char + pni.projectName[pni.cursor:]
					pni.cursor++
				}
			}
		}
	}

	return pni, nil
}

// View renders the project name input dialog
func (pni *ProjectNameInput) View() string {
	dialogWidth := 50
	if pni.width > 0 && dialogWidth > pni.width-4 {
		dialogWidth = pni.width - 4
	}

	// Create the message
	title := "Create New Project"
	message := "Enter project name:"

	// Show the input field with cursor
	input := pni.projectName
	if pni.cursor <= len(input) {
		// Insert cursor character
		if pni.cursor < len(input) {
			input = input[:pni.cursor] + "│" + input[pni.cursor:]
		} else {
			input = input + "│"
		}
	}

	// Add placeholder text if empty
	if pni.projectName == "" {
		input = "│"
	}

	instructions := "Enter: Confirm  •  Esc/Ctrl+Q: Cancel"
	hint := "(Leave empty for 'save')"

	// Create styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(dialogWidth - 6).
		Align(lipgloss.Left)

	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Align(lipgloss.Center)

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		"",
		messageStyle.Render(message),
		"",
		inputStyle.Render(input),
		"",
		hintStyle.Render(hint),
		"",
		instructionsStyle.Render(instructions),
	)

	// Create dialog container
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(dialogWidth)

	dialog := dialogStyle.Render(content)

	// Center the dialog on screen
	return lipgloss.NewStyle().
		Width(pni.width).
		Height(pni.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(dialog)
}

// NewProjectSelector creates a new project selector and starts the search
func NewProjectSelector() *ProjectSelector {
	ps := &ProjectSelector{
		projects:       []Project{},
		selectedIndex:  0,
		searchComplete: false,
		searching:      true,
	}
	return ps
}

// SearchProjects searches for ColliderTracker projects in common locations
func SearchProjects() ([]Project, error) {
	var projects []Project
	searchPaths := getSearchPaths()

	log.Printf("Searching for projects in: %v", searchPaths)

	for _, basePath := range searchPaths {
		found := searchInDirectory(basePath, 3) // Max depth of 3 levels
		projects = append(projects, found...)
	}

	// Sort projects by modification time (newest first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Modified.After(projects[j].Modified)
	})

	// Remove duplicates (same path)
	projects = removeDuplicates(projects)

	log.Printf("Found %d projects", len(projects))
	return projects, nil
}

// getSearchPaths returns common paths where projects might be located
func getSearchPaths() []string {
	paths := []string{}

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, cwd)
	}

	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, home)
		// Common project folders in home
		paths = append(paths, filepath.Join(home, "Music"))
		paths = append(paths, filepath.Join(home, "Documents"))
		paths = append(paths, filepath.Join(home, "Projects"))
		paths = append(paths, filepath.Join(home, "ColliderTracker"))
	}

	// Desktop (common on all platforms)
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "Desktop"))
	}

	return paths
}

// searchInDirectory recursively searches for projects in a directory
func searchInDirectory(dir string, maxDepth int) []Project {
	if maxDepth <= 0 {
		return []Project{}
	}

	var projects []Project

	// Check if this directory is a project
	dataFile := filepath.Join(dir, "data.json.gz")
	if stat, err := os.Stat(dataFile); err == nil && !stat.IsDir() {
		// This is a project directory
		project := Project{
			Name:     filepath.Base(dir),
			Path:     dir,
			Modified: stat.ModTime(),
		}
		projects = append(projects, project)
		log.Printf("Found project: %s at %s", project.Name, project.Path)
		return projects // Don't search subdirectories of a project
	}

	// Search subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return projects // Skip directories we can't read
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subDir := filepath.Join(dir, entry.Name())
			subProjects := searchInDirectory(subDir, maxDepth-1)
			projects = append(projects, subProjects...)
		}
	}

	return projects
}

// removeDuplicates removes duplicate projects based on path
func removeDuplicates(projects []Project) []Project {
	seen := make(map[string]bool)
	var result []Project

	for _, project := range projects {
		cleanPath := filepath.Clean(project.Path)
		if !seen[cleanPath] {
			seen[cleanPath] = true
			project.Path = cleanPath // Use cleaned path
			result = append(result, project)
		}
	}

	return result
}

// Init starts the project search
func (ps *ProjectSelector) Init() tea.Cmd {
	return ps.searchForProjects()
}

// searchForProjects is a tea.Cmd that searches for projects in the background
func (ps *ProjectSelector) searchForProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := SearchProjects()
		return searchCompleteMsg{
			projects: projects,
			err:      err,
		}
	}
}

// Update handles messages
func (ps *ProjectSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ps.width = msg.Width
		ps.height = msg.Height

	case searchCompleteMsg:
		ps.searching = false
		ps.searchComplete = true
		if msg.err != nil {
			log.Printf("Error searching for projects: %v", msg.err)
		} else {
			ps.projects = msg.projects
			// Select the first project by default
			if len(ps.projects) > 0 {
				ps.selectedIndex = 0
			} else {
				// No projects found, automatically transition to create new project
				nameInput := NewProjectNameInput()
				nameInput.width = ps.width
				nameInput.height = ps.height
				return nameInput, nameInput.Init()
			}
		}

	case tea.KeyMsg:
		if !ps.searchComplete {
			return ps, nil // Ignore keys while searching
		}

		switch msg.String() {
		case "q", "ctrl+c", "ctrl+q", "esc":
			return ps, tea.Quit

		case "up", "k":
			if ps.selectedIndex > 0 {
				ps.selectedIndex--
			}

		case "down", "j":
			if ps.selectedIndex < len(ps.projects)-1 {
				ps.selectedIndex++
			}

		case "enter":
			if len(ps.projects) > 0 {
				// Return the selected project
				selected := ps.projects[ps.selectedIndex]
				return &ProjectResult{
					SelectedProject: &selected,
					Cancelled:       false,
				}, tea.Quit
			}

		case "n":
			// Create new project option - transition to name input dialog
			nameInput := NewProjectNameInput()
			nameInput.width = ps.width
			nameInput.height = ps.height
			return nameInput, nameInput.Init()
		}
	}

	return ps, nil
}

// View renders the project selector
func (ps *ProjectSelector) View() string {
	if ps.searching {
		return ps.renderSearching()
	}

	if !ps.searchComplete {
		return "Loading..."
	}

	return ps.renderProjectList()
}

// renderSearching shows a searching message
func (ps *ProjectSelector) renderSearching() string {
	style := lipgloss.NewStyle().
		Padding(2).
		Foreground(lipgloss.Color("240"))

	return style.Render("🔍 Searching for ColliderTracker projects...")
}

// renderProjectList shows the list of found projects
func (ps *ProjectSelector) renderProjectList() string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Padding(0, 0, 1, 0)

	content.WriteString(titleStyle.Render("Select a ColliderTracker Project"))
	content.WriteString("\n")

	if len(ps.projects) == 0 {
		noProjectsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 0)

		content.WriteString(noProjectsStyle.Render("No projects found."))
		content.WriteString("\n")
	} else {
		// Render project list
		for i, project := range ps.projects {
			ps.renderProject(&content, project, i == ps.selectedIndex)
		}
	}

	// Add instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(1, 0, 0, 0)

	instructions := ""
	if len(ps.projects) > 0 {
		instructions += "↑/↓ or k/j: Navigate  •  Enter: Select  •  "
	}
	instructions += "n: New project  •  q/Esc: Quit"

	content.WriteString(instructionsStyle.Render(instructions))

	containerStyle := lipgloss.NewStyle().Padding(1, 2)
	return containerStyle.Render(content.String())
}

// renderProject renders a single project entry
func (ps *ProjectSelector) renderProject(content *strings.Builder, project Project, selected bool) {
	var style lipgloss.Style

	if selected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("7")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	}

	// Format the project info
	name := project.Name
	path := project.Path
	modified := project.Modified.Format("2006-01-02 15:04")

	// Truncate path if too long
	if len(path) > 60 {
		path = "..." + path[len(path)-57:]
	}

	projectInfo := fmt.Sprintf("%-20s %s", name, path)
	timeInfo := fmt.Sprintf("Modified: %s", modified)

	content.WriteString(style.Render(fmt.Sprintf("  %s", projectInfo)))
	content.WriteString("\n")

	if selected {
		timeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(0, 1)
		content.WriteString(timeStyle.Render(fmt.Sprintf("  %s", timeInfo)))
		content.WriteString("\n")
	}
}

// ProjectResult represents the result of project selection
type ProjectResult struct {
	SelectedProject *Project
	Cancelled       bool
}

// Init is required for tea.Model interface
func (pr *ProjectResult) Init() tea.Cmd {
	return nil
}

// Update is required for tea.Model interface
func (pr *ProjectResult) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return pr, nil
}

// View is required for tea.Model interface
func (pr *ProjectResult) View() string {
	return ""
}

// RunProjectSelector runs the project selector and returns the selected project path or name,
// whether it was cancelled, and whether it's a new project (true) or existing project (false)
func RunProjectSelector() (string, bool, bool) {
	selector := NewProjectSelector()
	p := tea.NewProgram(selector, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error running project selector: %v", err)
		return "", true, false // cancelled
	}

	// Handle project name input dialog result
	if nameInput, ok := finalModel.(*ProjectNameInput); ok {
		if nameInput.cancelled {
			return "", true, false
		}
		if nameInput.done {
			// Return the project name for new project creation
			return nameInput.projectName, false, true // not cancelled, is new project
		}
	}

	if result, ok := finalModel.(*ProjectResult); ok {
		if result.Cancelled {
			return "", true, false
		}
		if result.SelectedProject != nil {
			return result.SelectedProject.Path, false, false // not cancelled, is existing project
		}
		// User chose to create new project, return empty string (legacy fallback)
		return "", false, true // not cancelled, is new project (legacy)
	}

	return "", true, false // default to cancelled
}

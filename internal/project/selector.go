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
			}
		}

	case tea.KeyMsg:
		if !ps.searchComplete {
			return ps, nil // Ignore keys while searching
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
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
			// Create new project option
			return &ProjectResult{
				SelectedProject: nil,
				Cancelled:       false,
			}, tea.Quit
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

// RunProjectSelector runs the project selector and returns the selected project path
func RunProjectSelector() (string, bool) {
	selector := NewProjectSelector()
	p := tea.NewProgram(selector, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error running project selector: %v", err)
		return "", true // cancelled
	}

	if result, ok := finalModel.(*ProjectResult); ok {
		if result.Cancelled {
			return "", true
		}
		if result.SelectedProject != nil {
			return result.SelectedProject.Path, false
		}
		// User chose to create new project, return empty string
		return "", false
	}

	return "", true // default to cancelled
}
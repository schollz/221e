package views

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// SplashState represents the animation state of the splash screen
type SplashState struct {
	StartTime time.Time
	Duration  time.Duration
}

// NewSplashState creates a new splash state
func NewSplashState(duration time.Duration) *SplashState {
	return &SplashState{
		StartTime: time.Now(),
		Duration:  duration,
	}
}

// IsComplete returns true if the splash animation is done
func (s *SplashState) IsComplete() bool {
	return time.Since(s.StartTime) >= s.Duration
}

// GetProgress returns animation progress from 0.0 to 1.0
func (s *SplashState) GetProgress() float64 {
	elapsed := time.Since(s.StartTime)
	if elapsed >= s.Duration {
		return 1.0
	}
	return float64(elapsed) / float64(s.Duration)
}

// RenderSplashScreen renders the animated splash screen
func RenderSplashScreen(termWidth, termHeight int, state *SplashState, version string) string {
	var content strings.Builder

	// Handle edge cases for small or zero terminal dimensions
	if termWidth <= 0 || termHeight <= 0 {
		return "∞\nby infinite digits\ngithub.com/schollz/collidertracker\nversion " + version
	}

	// Ensure minimum dimensions for proper animation
	if termWidth < 40 {
		termWidth = 40
	}
	if termHeight < 10 {
		termHeight = 10
	}

	// Get animation progress
	progress := state.GetProgress()

	// Calculate center position
	centerY := termHeight / 2
	centerX := termWidth / 2

	// Fill screen with empty lines up to center
	for i := 0; i < centerY-6; i++ {
		content.WriteString("\n")
	}

	// Render animated blocks
	blocksLine := renderAnimatedBlocks(centerX, progress)
	content.WriteString(blocksLine)
	content.WriteString("\n\n")

	// Style for the text
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center)

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Align(lipgloss.Center)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Align(lipgloss.Center)

	// Render text with fade-in effect
	if progress > 0.05 {
		// Title
		titleLine := titleStyle.Width(termWidth).Render("∞")
		content.WriteString(titleLine)
		content.WriteString("\n")
	}

	if progress > 0.1 {
		// Subtitle
		subtitleLine := subtitleStyle.Width(termWidth).Render("by infinite digits")
		content.WriteString(subtitleLine)
		content.WriteString("\n")

	}

	if progress > 0.15 {
		// URL
		urlLine := urlStyle.Width(termWidth).Render("github.com/schollz/collidertracker")
		content.WriteString(urlLine)
		content.WriteString("\n")
	}

	if progress > 0.17 {
		// Version
		versionLine := subtitleStyle.Width(termWidth).Render("version " + version)
		content.WriteString(versionLine)
		content.WriteString("\n")
	}

	if progress > 0.2 {
		// Loading
		loadingLine := loadingStyle.Width(termWidth).Render("initializing SuperCollider...")
		content.WriteString(loadingLine)
		content.WriteString("\n")
	}

	// Fill remaining space
	for i := centerY + 4; i < termHeight-1; i++ {
		content.WriteString("\n")
	}

	return content.String()
}

// renderAnimatedBlocks creates the animated block pattern
func renderAnimatedBlocks(centerX int, progress float64) string {
	// Block characters for animation
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Create animated blocks coming together
	blockCount := 12
	halfBlocks := blockCount / 2

	// More dramatic movement - blocks start much further apart
	maxDistance := centerX / 2 // Start from edges of screen
	if maxDistance < 10 {
		maxDistance = 10
	}

	// Easing function for more dramatic movement (ease-out)
	easedProgress := 1.0 - ((1.0 - progress) * (1.0 - progress))

	// Calculate current distance from center (starts at maxDistance, ends at 0)
	currentDistance := int(float64(maxDistance) * (1.0 - easedProgress))

	// Calculate positions for left and right groups
	leftGroupCenter := centerX - currentDistance
	rightGroupCenter := centerX + currentDistance

	// Build the animation line
	var line strings.Builder

	// Calculate total width needed for proper centering
	blockSpacing := 1
	leftGroupWidth := halfBlocks + (halfBlocks-1)*blockSpacing
	rightGroupWidth := halfBlocks + (halfBlocks-1)*blockSpacing
	gapBetweenGroups := rightGroupCenter - leftGroupCenter - leftGroupWidth
	if gapBetweenGroups < 0 {
		gapBetweenGroups = 0
	}

	totalWidth := leftGroupWidth + gapBetweenGroups + rightGroupWidth
	startX := centerX - totalWidth/2

	// Ensure non-negative padding
	if startX < 0 {
		startX = 0
	}

	// Add initial padding to center the entire animation
	line.WriteString(strings.Repeat(" ", startX))

	// Left blocks (moving right)
	for i := 0; i < halfBlocks; i++ {
		// Block height animation - grows over time with staggered effect
		staggerDelay := float64(i) * 0.1
		adjustedProgress := progress - staggerDelay
		if adjustedProgress < 0 {
			adjustedProgress = 0
		}

		blockHeight := int(adjustedProgress * float64(len(blocks)-1))
		if blockHeight >= len(blocks) {
			blockHeight = len(blocks) - 1
		}
		if blockHeight < 0 {
			blockHeight = 0
		}

		// Color animation with more vibrant colors
		blockStyle := lipgloss.NewStyle()
		if progress < 0.3 {
			// Start with blue
			blockStyle = blockStyle.Foreground(lipgloss.Color("12"))
		} else if progress < 0.6 {
			// Transition to cyan
			blockStyle = blockStyle.Foreground(lipgloss.Color("14"))
		} else {
			// End with bright white
			blockStyle = blockStyle.Foreground(lipgloss.Color("15")).Bold(true)
		}

		block := blockStyle.Render(blocks[blockHeight])
		line.WriteString(block)

		// Add spacing between blocks
		if i < halfBlocks-1 {
			line.WriteString(strings.Repeat(" ", blockSpacing))
		}
	}

	// Gap between groups
	line.WriteString(strings.Repeat(" ", gapBetweenGroups))

	// Right blocks (moving left)
	for i := 0; i < halfBlocks; i++ {
		// Add spacing before block
		if i > 0 {
			line.WriteString(strings.Repeat(" ", blockSpacing))
		}

		// Block height animation with staggered effect (reverse order for right side)
		staggerDelay := float64(halfBlocks-1-i) * 0.1
		adjustedProgress := progress - staggerDelay
		if adjustedProgress < 0 {
			adjustedProgress = 0
		}

		blockHeight := int(adjustedProgress * float64(len(blocks)-1))
		if blockHeight >= len(blocks) {
			blockHeight = len(blocks) - 1
		}
		if blockHeight < 0 {
			blockHeight = 0
		}

		// Color animation matching left side
		blockStyle := lipgloss.NewStyle()
		if progress < 0.3 {
			blockStyle = blockStyle.Foreground(lipgloss.Color("12"))
		} else if progress < 0.6 {
			blockStyle = blockStyle.Foreground(lipgloss.Color("14"))
		} else {
			blockStyle = blockStyle.Foreground(lipgloss.Color("15")).Bold(true)
		}

		block := blockStyle.Render(blocks[blockHeight])
		line.WriteString(block)
	}

	return line.String()
}

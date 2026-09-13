package tui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#4493f8"}
	subtle = lipgloss.AdaptiveColor{Light: "#656d76", Dark: "#8b949e"}
	danger = lipgloss.AdaptiveColor{Light: "#cf222e", Dark: "#f85149"}

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle)

	focusedPane = paneStyle.BorderForeground(accent)

	titleStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	dimStyle   = lipgloss.NewStyle().Foreground(subtle)
	helpStyle  = lipgloss.NewStyle().Foreground(subtle)
	errStyle   = lipgloss.NewStyle().Foreground(danger)

	selectedStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(accent)

	scrollTrackStyle = lipgloss.NewStyle().Foreground(subtle)
	scrollThumbStyle = lipgloss.NewStyle().Foreground(accent)
)

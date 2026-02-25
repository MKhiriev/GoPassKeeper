package tui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle        = lipgloss.NewStyle().Padding(1, 2)
	titleStyle      = lipgloss.NewStyle().Bold(true)
	helpStyle       = lipgloss.NewStyle().Faint(true)
	errorStyle      = lipgloss.NewStyle().Bold(true)
	overlayBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

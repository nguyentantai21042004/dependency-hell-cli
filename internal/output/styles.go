package output

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Align(lipgloss.Center)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Align(lipgloss.Center)

	// Table styles
	TableBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4"))

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status colors
	StatusGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	StatusWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))

	StatusBadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	// Language name style
	LanguageStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7FF"))

	// Disk usage styles
	DiskUsageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	DiskUsageDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))
)

package output

import (
	"fmt"
	"strings"

	"dependency-hell-cli/internal/core"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// RenderInfo renders detailed information about a language installation
func RenderInfo(provider core.LanguageProvider, installation *core.Installation, diskUsage *core.DiskUsage) string {
	var output strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Render(fmt.Sprintf("%s Information", provider.Name()))

	border := strings.Repeat("─", 50)
	output.WriteString("╭" + border + "╮\n")
	output.WriteString("│" + lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(header) + "│\n")
	output.WriteString("╰" + border + "╯\n\n")

	// Version and Source
	output.WriteString(fmt.Sprintf("Version: %s\n", installation.Version))

	status := core.DetermineStatus(installation.Source)
	statusIcon := status.GetStatusIcon()
	output.WriteString(fmt.Sprintf("Source: %s %s\n\n", statusIcon, installation.Source))

	// Binary Paths
	output.WriteString(lipgloss.NewStyle().Bold(true).Render("Binary Paths:") + "\n")
	output.WriteString(fmt.Sprintf("  • Executable: %s\n", installation.BinaryPath))

	if installation.ManagerPath != "" {
		output.WriteString(fmt.Sprintf("  • Manager: %s\n", installation.ManagerPath))
	}
	output.WriteString("\n")

	// Environment Variables
	envVars := provider.GetEnvVars()
	if len(envVars) > 0 {
		output.WriteString(lipgloss.NewStyle().Bold(true).Render("Environment Variables:") + "\n")
		for key, value := range envVars {
			output.WriteString(fmt.Sprintf("  • %s: %s\n", key, value))
		}
		output.WriteString("\n")
	}

	// Cache Locations
	if diskUsage != nil && len(diskUsage.Items) > 0 {
		output.WriteString(lipgloss.NewStyle().Bold(true).Render("Cache Locations:") + "\n")
		for _, item := range diskUsage.Items {
			if item.Size > 0 {
				size := humanize.Bytes(uint64(item.Size))
				output.WriteString(fmt.Sprintf("  • %s: %s (%s)\n", item.Description, item.Path, size))
			}
		}
		output.WriteString("\n")
	}

	// Total Disk Usage
	if diskUsage != nil && diskUsage.Total > 0 {
		totalSize := humanize.Bytes(uint64(diskUsage.Total))
		total := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Render(fmt.Sprintf("Total Disk Usage: %s", totalSize))
		output.WriteString(total + "\n")
	}

	return output.String()
}

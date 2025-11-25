package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/nguyentantai21042004/dependency-hell-cli/internal/core"
)

// RenderCleanPreview shows what would be cleaned in dry-run mode
func RenderCleanPreview(language string, items []core.CleanableItem) string {
	var output strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Render(fmt.Sprintf("Clean Preview: %s", language))

	border := strings.Repeat("─", 60)
	output.WriteString("╭" + border + "╮\n")
	output.WriteString("│" + lipgloss.NewStyle().Width(60).Align(lipgloss.Center).Render(header) + "│\n")
	output.WriteString("╰" + border + "╯\n\n")

	// Items list
	output.WriteString("The following items will be cleaned:\n\n")

	var totalSize int64
	for _, item := range items {
		icon := "🗑️ "
		desc := item.Description

		if item.Command != "" {
			output.WriteString(fmt.Sprintf("  %s %s\n", icon, desc))
			output.WriteString(fmt.Sprintf("      Command: %s\n", item.Command))
		} else {
			output.WriteString(fmt.Sprintf("  %s %s\n", icon, desc))
			output.WriteString(fmt.Sprintf("      Path: %s\n", item.Path))
		}

		if item.Size > 0 {
			size := humanize.Bytes(uint64(item.Size))
			output.WriteString(fmt.Sprintf("      Size: %s\n", size))
			totalSize += item.Size
		}

		if !item.Safe {
			warning := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Render("      ⚠️  WARNING: This item requires careful consideration")
			output.WriteString(warning + "\n")
		}

		output.WriteString("\n")
	}

	// Total
	if totalSize > 0 {
		totalStr := humanize.Bytes(uint64(totalSize))
		total := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Render(fmt.Sprintf("Total space to reclaim: %s", totalStr))
		output.WriteString(total + "\n\n")
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("Run without --dry-run to execute cleaning.")
	output.WriteString(footer + "\n")

	return output.String()
}

// RenderCleanResult shows the result of cleaning operation
func RenderCleanResult(result *core.CleanResult, items []core.CleanableItem) string {
	var output strings.Builder

	if result.ItemsCleaned == 0 {
		output.WriteString("❌ No items were cleaned.\n")
		return output.String()
	}

	// Success header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Render("✅ Cleaning completed!")
	output.WriteString(header + "\n\n")

	// Cleaned items
	output.WriteString("Cleaned:\n")
	for _, item := range items {
		if item.Size > 0 {
			size := humanize.Bytes(uint64(item.Size))
			output.WriteString(fmt.Sprintf("  ✓ %s (%s)\n", item.Description, size))
		} else {
			output.WriteString(fmt.Sprintf("  ✓ %s\n", item.Description))
		}
	}

	// Total space reclaimed
	if result.SpaceReclaimed > 0 {
		output.WriteString("\n")
		totalStr := humanize.Bytes(uint64(result.SpaceReclaimed))
		total := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Render(fmt.Sprintf("Total space reclaimed: %s", totalStr))
		output.WriteString(total + "\n")
	}

	// Errors
	if len(result.Errors) > 0 {
		output.WriteString("\n")
		errorHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render("⚠️  Errors encountered:")
		output.WriteString(errorHeader + "\n")

		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  • %s\n", err.Error()))
		}
	}

	return output.String()
}

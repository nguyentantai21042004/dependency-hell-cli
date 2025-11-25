package output

import (
	"fmt"
	"runtime"
	"strings"

	"dependency-hell-cli/internal/core"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/host"
)

// ScanResult represents the result of scanning a language
type ScanResult struct {
	Provider     core.LanguageProvider
	Installation *core.Installation
	DiskUsage    *core.DiskUsage
	Error        error
}

// RenderScanResults renders the scan results as a beautiful table
func RenderScanResults(results []ScanResult) string {
	var output strings.Builder

	// Render header
	output.WriteString(renderHeader())
	output.WriteString("\n\n")

	// Render table
	output.WriteString(renderTable(results))

	return output.String()
}

// renderHeader renders the header with system info
func renderHeader() string {
	var output strings.Builder

	// Get system info
	osInfo := getOSInfo()

	// Title
	title := HeaderStyle.Render("Dependency Hell Analyzer (v0.1.0)")
	output.WriteString(title)
	output.WriteString("\n")

	// System info
	sysInfo := SubHeaderStyle.Render(osInfo)
	output.WriteString(sysInfo)

	return output.String()
}

// getOSInfo gets OS and architecture information
func getOSInfo() string {
	info, err := host.Info()
	if err != nil {
		return fmt.Sprintf("OS: %s (%s)", runtime.GOOS, runtime.GOARCH)
	}

	platform := info.Platform
	if info.PlatformVersion != "" {
		platform = fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion)
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "arm64" {
		arch = "ARM64"
	}

	return fmt.Sprintf("OS: %s (%s)", platform, arch)
}

// renderTable renders the results table
func renderTable(results []ScanResult) string {
	var rows []string

	// Table header
	header := renderTableHeader()
	rows = append(rows, header)

	// Separator
	separator := strings.Repeat("─", 100)
	rows = append(rows, TableBorderStyle.Render(separator))

	// Table rows
	for _, result := range results {
		if result.Error != nil {
			// Show error row
			errorRow := renderErrorRow(result)
			rows = append(rows, errorRow)
		} else {
			// Show normal rows
			resultRows := renderResultRows(result)
			rows = append(rows, resultRows...)
		}

		// Add separator between languages
		rows = append(rows, TableBorderStyle.Render(separator))
	}

	return strings.Join(rows, "\n")
}

// renderTableHeader renders the table header
func renderTableHeader() string {
	headers := []string{
		TableHeaderStyle.Width(8).Render("STATUS"),
		TableHeaderStyle.Width(12).Render("LANGUAGE"),
		TableHeaderStyle.Width(15).Render("VERSION"),
		TableHeaderStyle.Width(18).Render("SOURCE"),
		TableHeaderStyle.Width(45).Render("DISK USAGE"),
	}

	return strings.Join(headers, " ")
}

// renderErrorRow renders an error row
func renderErrorRow(result ScanResult) string {
	status := StatusBadStyle.Width(8).Render("🔴")
	language := LanguageStyle.Width(12).Render(result.Provider.Name())
	version := TableCellStyle.Width(15).Render("N/A")
	source := TableCellStyle.Width(18).Render("Not Found")
	diskUsage := TableCellStyle.Width(45).Render(result.Error.Error())

	return strings.Join([]string{status, language, version, source, diskUsage}, " ")
}

// renderResultRows renders result rows (can be multiple for disk usage breakdown)
func renderResultRows(result ScanResult) []string {
	var rows []string

	installation := result.Installation
	diskUsage := result.DiskUsage

	// Determine status
	status := core.DetermineStatus(installation.Source)
	statusIcon := status.GetStatusIcon()

	// First row with main info
	statusCell := TableCellStyle.Width(8).Render(statusIcon)
	languageCell := LanguageStyle.Width(12).Render(result.Provider.Name())
	versionCell := TableCellStyle.Width(15).Render(installation.Version)
	sourceCell := TableCellStyle.Width(18).Render(string(installation.Source))

	// Disk usage - show total first
	totalSize := humanize.Bytes(uint64(diskUsage.Total))
	diskUsageCell := DiskUsageStyle.Width(45).Render(fmt.Sprintf("Total: %s", totalSize))

	firstRow := strings.Join([]string{statusCell, languageCell, versionCell, sourceCell, diskUsageCell}, " ")
	rows = append(rows, firstRow)

	// Additional rows for disk usage breakdown
	for _, item := range diskUsage.Items {
		if item.Size > 0 {
			emptyCell := TableCellStyle.Width(8).Render("")
			emptyLang := TableCellStyle.Width(12).Render("")
			emptyVer := TableCellStyle.Width(15).Render("")
			emptySource := TableCellStyle.Width(18).Render("")

			size := humanize.Bytes(uint64(item.Size))
			desc := fmt.Sprintf("  ↳ %s: %s", item.Description, size)
			diskCell := DiskUsageDescStyle.Width(45).Render(desc)

			row := strings.Join([]string{emptyCell, emptyLang, emptyVer, emptySource, diskCell}, " ")
			rows = append(rows, row)
		}
	}

	return rows
}

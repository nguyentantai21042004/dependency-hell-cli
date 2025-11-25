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
	Provider      core.LanguageProvider
	Installations []core.Installation // Changed to array to support multiple versions
	DiskUsage     *core.DiskUsage
	Error         error
}

// RenderScanResults renders the scan results as a formatted table
func RenderScanResults(results []ScanResult) string {
	var output strings.Builder

	// Filter out results with errors (uninstalled languages)
	var validResults []ScanResult
	for _, result := range results {
		if result.Error == nil {
			validResults = append(validResults, result)
		}
	}

	// If no valid results, show message
	if len(validResults) == 0 {
		return "No languages detected in your environment.\n"
	}

	// Get system info
	osInfo, arch := getSystemInfo()

	// Header
	output.WriteString("                                     \n")
	output.WriteString("  Dependency Hell Analyzer (v0.1.0)  \n")
	output.WriteString("                                     \n")
	output.WriteString(fmt.Sprintf("OS: %s (%s)\n\n", osInfo, arch))

	// Table header
	output.WriteString(" STATUS   LANGUAGE     VERSION         SOURCE             DISK USAGE                                  \n")
	output.WriteString("────────────────────────────────────────────────────────────────────────────────────────────────────\n")

	// Table rows - only for valid results
	for _, result := range validResults {
		rows := renderResultRows(result)
		for _, row := range rows {
			output.WriteString(row + "\n")
		}
		output.WriteString("────────────────────────────────────────────────────────────────────────────────────────────────────\n")
	}

	return output.String()
}

// getSystemInfo gets OS and architecture information
func getSystemInfo() (string, string) {
	info, err := host.Info()
	if err != nil {
		return runtime.GOOS, runtime.GOARCH
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

	return platform, arch
}

// renderResultRows renders result rows (can be multiple for disk usage breakdown)
func renderResultRows(result ScanResult) []string {
	var rows []string

	installations := result.Installations
	diskUsage := result.DiskUsage

	// If multiple installations, show count
	versionInfo := ""
	if len(installations) == 1 {
		versionInfo = installations[0].Version
	} else {
		versionInfo = fmt.Sprintf("%d versions", len(installations))
	}

	// Determine status from first installation
	status := core.DetermineStatus(installations[0].Source)
	statusIcon := status.GetStatusIcon()

	// Determine source display
	sourceDisplay := string(installations[0].Source)
	if installations[0].ManagerName != "" {
		sourceDisplay = installations[0].ManagerName
	}

	// First row with main info
	statusStr := fmt.Sprintf(" %-7s", statusIcon)
	languageStr := fmt.Sprintf(" %-11s", result.Provider.Name())
	versionStr := fmt.Sprintf(" %-14s", versionInfo)
	sourceStr := fmt.Sprintf(" %-17s", sourceDisplay)

	// Disk usage - show total first
	totalSize := humanize.Bytes(uint64(diskUsage.Total))
	diskUsageStr := fmt.Sprintf(" Total: %-38s", totalSize)

	firstRow := statusStr + languageStr + versionStr + sourceStr + diskUsageStr
	rows = append(rows, firstRow)

	// If multiple versions, show each version
	if len(installations) > 1 {
		for i, inst := range installations {
			activeMarker := ""
			if i == 0 {
				activeMarker = " (active)"
			}
			versionLine := fmt.Sprintf("  • %s%s", inst.Version, activeMarker)

			emptyPrefix := strings.Repeat(" ", 8+12+15+18)
			versionCell := fmt.Sprintf(" %-43s", versionLine)

			row := emptyPrefix + versionCell
			rows = append(rows, row)
		}
	}

	// Additional rows for disk usage breakdown
	for _, item := range diskUsage.Items {
		if item.Size > 0 {
			size := humanize.Bytes(uint64(item.Size))
			desc := fmt.Sprintf("  ↳ %s: %s", item.Description, size)

			emptyPrefix := strings.Repeat(" ", 8+12+15+18)
			diskCell := fmt.Sprintf(" %-43s", desc)

			row := emptyPrefix + diskCell
			rows = append(rows, row)
		}
	}

	return rows
}

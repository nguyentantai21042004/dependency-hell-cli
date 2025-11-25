package cleaner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/scanner"
)

// ConfirmClean shows confirmation prompt and returns user's decision
func ConfirmClean(items []core.CleanableItem, totalSize int64) bool {
	fmt.Println()
	fmt.Println("⚠️  WARNING: This will delete cache files!")
	fmt.Println()
	fmt.Println("You are about to clean:")

	for _, item := range items {
		if item.Size > 0 {
			fmt.Printf("  • %s (%s)\n", item.Description, formatSize(item.Size))
		} else {
			fmt.Printf("  • %s\n", item.Description)
		}
	}

	fmt.Println()
	fmt.Printf("Total: %s will be reclaimed\n", formatSize(totalSize))
	fmt.Println()
	fmt.Println("These caches will be rebuilt on next use.")
	fmt.Println()
	fmt.Print("Do you want to continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// CleanItems executes cleaning for the given items
func CleanItems(items []core.CleanableItem, dryRun bool) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if dryRun {
			// In dry-run mode, just count what would be cleaned
			result.ItemsCleaned++
			result.SpaceReclaimed += item.Size
			continue
		}

		// Execute actual cleaning
		var err error
		if item.Command != "" {
			// Use command if specified
			err = RunCleanCommand(item.Command)
		} else if item.Path != "" {
			// Otherwise remove directory
			err = CleanDirectory(item.Path)
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to clean %s: %w", item.Description, err))
			continue
		}

		result.ItemsCleaned++
		result.SpaceReclaimed += item.Size
	}

	return result, nil
}

// CleanDirectory safely removes a directory
func CleanDirectory(path string) error {
	expandedPath := scanner.ExpandHome(path)

	if !scanner.PathExists(expandedPath) {
		return nil // Already clean
	}

	return os.RemoveAll(expandedPath)
}

// RunCleanCommand runs a clean command (e.g., go clean -modcache)
func RunCleanCommand(cmdStr string) error {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s (output: %s)", err, string(output))
	}

	return nil
}

// formatSize formats bytes to human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

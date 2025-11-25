package providers

import (
	"fmt"
	"os"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/scanner"
)

// PHPProvider implements the LanguageProvider interface for PHP
type PHPProvider struct{}

// NewPHPProvider creates a new PHP provider
func NewPHPProvider() *PHPProvider {
	return &PHPProvider{}
}

// Name returns the name of the language
func (p *PHPProvider) Name() string {
	return "PHP"
}

// DetectInstalled detects installed PHP versions
func (p *PHPProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if php is installed
	phpPath, err := scanner.FindExecutable("php")
	if err != nil {
		return nil, fmt.Errorf("php not found in PATH")
	}

	// Resolve symlinks
	realPath, err := scanner.ResolveSymlink(phpPath)
	if err != nil {
		realPath = phpPath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("php", "--version")
	if err != nil {
		return nil, fmt.Errorf("failed to get php version: %w", err)
	}

	// Parse version (e.g., "PHP 8.2.0 (cli) ...")
	versionStr := p.parseVersion(version)

	// Determine source
	source := p.determineSource(realPath)

	installation := core.Installation{
		Version:     versionStr,
		Source:      source,
		BinaryPath:  phpPath,
		ManagerPath: p.getManagerPath(realPath, source),
	}

	return []core.Installation{installation}, nil
}

// parseVersion extracts version from php --version output
func (p *PHPProvider) parseVersion(output string) string {
	// Example: "PHP 8.2.0 (cli) (built: Dec  6 2022 15:31:23) ( NTS )"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		if strings.HasPrefix(firstLine, "PHP ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

// determineSource determines the installation source based on path
func (p *PHPProvider) determineSource(path string) core.InstallSource {
	if strings.Contains(path, ".phpenv") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	if strings.Contains(path, "/usr/bin/php") {
		return core.SourceSystem
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *PHPProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager && strings.Contains(path, ".phpenv") {
		if idx := strings.Index(path, ".phpenv"); idx != -1 {
			return path[:idx+7]
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for PHP ecosystem
func (p *PHPProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// PHP installation (if via Homebrew)
	phpPath, err := scanner.FindExecutable("php")
	if err == nil {
		realPath, _ := scanner.ResolveSymlink(phpPath)
		if strings.Contains(realPath, "/opt/homebrew") || strings.Contains(realPath, "/usr/local/Cellar") {
			// Get Homebrew Cellar directory
			if idx := strings.Index(realPath, "/Cellar/php"); idx != -1 {
				phpDir := realPath[:strings.Index(realPath[idx:], "/bin")+idx]
				if scanner.PathExists(phpDir) {
					size, _ := scanner.CalculateDirSize(phpDir)
					items = append(items, core.DiskUsageItem{
						Path:        phpDir,
						Description: "PHP Installation",
						Size:        size,
					})
				}
			}
		}
	}

	// Composer cache
	composerCache := "~/.composer/cache"
	if scanner.PathExists(composerCache) {
		size, _ := scanner.CalculateDirSize(composerCache)
		items = append(items, core.DiskUsageItem{
			Path:        composerCache,
			Description: "Composer Cache",
			Size:        size,
		})
	}

	// Composer vendor (global packages)
	composerVendor := "~/.composer/vendor"
	if scanner.PathExists(composerVendor) {
		size, _ := scanner.CalculateDirSize(composerVendor)
		items = append(items, core.DiskUsageItem{
			Path:        composerVendor,
			Description: "Composer Global Packages",
			Size:        size,
		})
	}

	// Calculate total
	var total int64
	for _, item := range items {
		total += item.Size
	}

	return &core.DiskUsage{
		Items: items,
		Total: total,
	}, nil
}

// GetEnvVars returns relevant environment variables
func (p *PHPProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	envVars := []string{"COMPOSER_HOME", "PHP_INI_SCAN_DIR"}
	for _, name := range envVars {
		if value := scanner.GetEnvVar(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// GetCleanableItems returns items that can be cleaned for PHP
func (p *PHPProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// Composer cache (safe)
	composerCache := "~/.composer/cache"
	if scanner.PathExists(composerCache) {
		size, _ := scanner.CalculateDirSize(composerCache)
		items = append(items, core.CleanableItem{
			Description: "Composer Cache",
			Command:     "composer clear-cache",
			Size:        size,
			Safe:        true,
		})
	}

	return items, nil
}

// Clean executes cleaning for PHP
func (p *PHPProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if item.Command != "" {
			// For composer clear-cache, just remove the directory
			if item.Path != "" {
				if err := os.RemoveAll(scanner.ExpandHome(item.Path)); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to clean %s: %w", item.Description, err))
					continue
				}
			}
		}

		result.ItemsCleaned++
		result.SpaceReclaimed += item.Size
	}

	return result, nil
}

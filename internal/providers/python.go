package providers

import (
	"fmt"
	"os"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/scanner"
)

// PythonProvider implements the LanguageProvider interface for Python
type PythonProvider struct{}

// NewPythonProvider creates a new Python provider
func NewPythonProvider() *PythonProvider {
	return &PythonProvider{}
}

// Name returns the name of the language
func (p *PythonProvider) Name() string {
	return "Python"
}

// DetectInstalled detects installed Python versions
func (p *PythonProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if python3 is installed
	pythonPath, err := scanner.FindExecutable("python3")
	if err != nil {
		return nil, fmt.Errorf("python3 not found in PATH")
	}

	// Resolve symlinks
	realPath, err := scanner.ResolveSymlink(pythonPath)
	if err != nil {
		realPath = pythonPath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("python3", "--version")
	if err != nil {
		return nil, fmt.Errorf("failed to get python version: %w", err)
	}

	// Parse version (e.g., "Python 3.11.0")
	versionStr := "unknown"
	if strings.HasPrefix(version, "Python ") {
		versionStr = strings.TrimPrefix(version, "Python ")
	}

	// Determine source
	source := p.determineSource(realPath)

	installation := core.Installation{
		Version:     versionStr,
		Source:      source,
		BinaryPath:  pythonPath,
		ManagerPath: p.getManagerPath(realPath, source),
	}

	return []core.Installation{installation}, nil
}

// determineSource determines the installation source based on path
func (p *PythonProvider) determineSource(path string) core.InstallSource {
	if strings.Contains(path, ".pyenv") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "anaconda") || strings.Contains(path, "miniconda") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	if strings.Contains(path, "/usr/bin/python") {
		return core.SourceSystem
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *PythonProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager {
		if strings.Contains(path, ".pyenv") {
			if idx := strings.Index(path, ".pyenv"); idx != -1 {
				return path[:idx+6]
			}
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for Python ecosystem
func (p *PythonProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// Pyenv versions
	pyenvPath := "~/.pyenv/versions"
	if scanner.PathExists(pyenvPath) {
		size, _ := scanner.CalculateDirSize(pyenvPath)
		items = append(items, core.DiskUsageItem{
			Path:        pyenvPath,
			Description: "Pyenv Versions",
			Size:        size,
		})
	}

	// Pip cache
	pipCache := "~/Library/Caches/pip"
	if scanner.PathExists(pipCache) {
		size, _ := scanner.CalculateDirSize(pipCache)
		items = append(items, core.DiskUsageItem{
			Path:        pipCache,
			Description: "Pip Cache",
			Size:        size,
		})
	}

	// Virtualenvs (if using virtualenvwrapper)
	virtualenvs := "~/.virtualenvs"
	if scanner.PathExists(virtualenvs) {
		size, _ := scanner.CalculateDirSize(virtualenvs)
		items = append(items, core.DiskUsageItem{
			Path:        virtualenvs,
			Description: "Virtualenvs",
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
func (p *PythonProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	envVars := []string{"PYTHONPATH", "VIRTUAL_ENV", "PYENV_ROOT"}
	for _, name := range envVars {
		if value := scanner.GetEnvVar(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// GetCleanableItems returns items that can be cleaned for Python
func (p *PythonProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// Pip cache (safe)
	pipCache := "~/Library/Caches/pip"
	if scanner.PathExists(pipCache) {
		size, _ := scanner.CalculateDirSize(pipCache)
		items = append(items, core.CleanableItem{
			Description: "Pip Cache",
			Command:     "pip cache purge",
			Size:        size,
			Safe:        true,
		})
	}

	return items, nil
}

// Clean executes cleaning for Python
func (p *PythonProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if item.Command != "" {
			// Execute clean command
			parts := strings.Fields(item.Command)
			if len(parts) > 0 {
				// For pip cache purge, we need to handle it specially
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

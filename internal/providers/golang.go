package providers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/scanner"
)

// GoProvider implements the LanguageProvider interface for Go
type GoProvider struct{}

// NewGoProvider creates a new Go provider
func NewGoProvider() *GoProvider {
	return &GoProvider{}
}

// Name returns the name of the language
func (p *GoProvider) Name() string {
	return "Golang"
}

// DetectInstalled detects installed Go versions
func (p *GoProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if go is installed
	goPath, err := scanner.FindExecutable("go")
	if err != nil {
		return nil, fmt.Errorf("go not found in PATH")
	}

	// Resolve symlinks to get actual path
	realPath, err := scanner.ResolveSymlink(goPath)
	if err != nil {
		realPath = goPath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("go", "version")
	if err != nil {
		return nil, fmt.Errorf("failed to get go version: %w", err)
	}

	// Parse version (e.g., "go version go1.21.3 darwin/arm64")
	parts := strings.Fields(version)
	versionStr := "unknown"
	if len(parts) >= 3 {
		versionStr = strings.TrimPrefix(parts[2], "go")
	}

	// Determine source
	source := p.determineSource(realPath)
	managerName := p.getManagerName(realPath, source)

	installation := core.Installation{
		Version:     versionStr,
		Source:      source,
		BinaryPath:  goPath,
		ManagerPath: p.getManagerPath(realPath, source),
		ManagerName: managerName,
	}

	return []core.Installation{installation}, nil
}

// getManagerName returns the specific version manager name
func (p *GoProvider) getManagerName(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager {
		if strings.Contains(path, ".goenv") {
			return "goenv"
		}
	}
	return ""
}

// determineSource determines the installation source based on path
func (p *GoProvider) determineSource(path string) core.InstallSource {
	if strings.Contains(path, ".goenv") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	if strings.Contains(path, "/usr/local/go") {
		return core.SourceManual
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *GoProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager && strings.Contains(path, ".goenv") {
		// Extract .goenv path
		if idx := strings.Index(path, ".goenv"); idx != -1 {
			return path[:idx+6] // Include ".goenv"
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for Go caches
func (p *GoProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// Get GOROOT (SDK)
	goroot := p.getGoEnv("GOROOT")
	if goroot != "" && scanner.PathExists(goroot) {
		size, _ := scanner.CalculateDirSize(goroot)
		items = append(items, core.DiskUsageItem{
			Path:        goroot,
			Description: "SDK",
			Size:        size,
		})
	}

	// Get GOCACHE (Build cache)
	gocache := p.getGoEnv("GOCACHE")
	if gocache != "" && scanner.PathExists(gocache) {
		size, _ := scanner.CalculateDirSize(gocache)
		items = append(items, core.DiskUsageItem{
			Path:        gocache,
			Description: "Build Cache",
			Size:        size,
		})
	}

	// Get GOMODCACHE (Module cache - the big one!)
	gomodcache := p.getGoEnv("GOMODCACHE")
	if gomodcache == "" {
		// Fallback to GOPATH/pkg/mod
		gopath := p.getGoEnv("GOPATH")
		if gopath != "" {
			gomodcache = gopath + "/pkg/mod"
		}
	}
	if gomodcache != "" && scanner.PathExists(gomodcache) {
		size, _ := scanner.CalculateDirSize(gomodcache)
		items = append(items, core.DiskUsageItem{
			Path:        gomodcache,
			Description: "Module Cache",
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
func (p *GoProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	envVarNames := []string{"GOROOT", "GOPATH", "GOCACHE", "GOMODCACHE"}
	for _, name := range envVarNames {
		if value := p.getGoEnv(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// getGoEnv gets a Go environment variable
func (p *GoProvider) getGoEnv(name string) string {
	cmd := exec.Command("go", "env", name)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to OS environment variable
		return os.Getenv(name)
	}
	return strings.TrimSpace(string(output))
}

// GetCleanableItems returns items that can be cleaned for Go
func (p *GoProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// Module cache - use go clean -modcache (safe)
	gomodcache := p.getGoEnv("GOMODCACHE")
	if gomodcache == "" {
		gopath := p.getGoEnv("GOPATH")
		if gopath != "" {
			gomodcache = gopath + "/pkg/mod"
		}
	}
	if gomodcache != "" && scanner.PathExists(gomodcache) {
		size, _ := scanner.CalculateDirSize(gomodcache)
		items = append(items, core.CleanableItem{
			Description: "Go Module Cache",
			Command:     "go clean -modcache",
			Size:        size,
			Safe:        true,
		})
	}

	// Build cache - use go clean -cache (safe)
	gocache := p.getGoEnv("GOCACHE")
	if gocache != "" && scanner.PathExists(gocache) {
		size, _ := scanner.CalculateDirSize(gocache)
		items = append(items, core.CleanableItem{
			Description: "Go Build Cache",
			Command:     "go clean -cache",
			Size:        size,
			Safe:        true,
		})
	}

	return items, nil
}

// Clean executes cleaning for Go
func (p *GoProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if item.Command != "" {
			// Execute go clean command
			cmd := exec.Command("sh", "-c", item.Command)
			if err := cmd.Run(); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to clean %s: %w", item.Description, err))
				continue
			}
		}

		result.ItemsCleaned++
		result.SpaceReclaimed += item.Size
	}

	return result, nil
}

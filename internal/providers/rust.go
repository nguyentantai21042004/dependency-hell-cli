package providers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/scanner"
)

// RustProvider implements the LanguageProvider interface for Rust
type RustProvider struct{}

// NewRustProvider creates a new Rust provider
func NewRustProvider() *RustProvider {
	return &RustProvider{}
}

// Name returns the name of the language
func (p *RustProvider) Name() string {
	return "Rust"
}

// DetectInstalled detects installed Rust versions
func (p *RustProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if rustc is installed
	rustcPath, err := scanner.FindExecutable("rustc")
	if err != nil {
		return nil, fmt.Errorf("rustc not found in PATH")
	}

	// Resolve symlinks
	realPath, err := scanner.ResolveSymlink(rustcPath)
	if err != nil {
		realPath = rustcPath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("rustc", "--version")
	if err != nil {
		return nil, fmt.Errorf("failed to get rust version: %w", err)
	}

	// Parse version (e.g., "rustc 1.74.0 (79e9716c9 2023-11-13)")
	versionStr := p.parseVersion(version)

	// Determine source
	source := p.determineSource(realPath)

	installation := core.Installation{
		Version:     versionStr,
		Source:      source,
		BinaryPath:  rustcPath,
		ManagerPath: p.getManagerPath(realPath, source),
	}

	return []core.Installation{installation}, nil
}

// parseVersion extracts version from rustc --version output
func (p *RustProvider) parseVersion(output string) string {
	// Example: "rustc 1.74.0 (79e9716c9 2023-11-13)"
	parts := strings.Fields(output)
	if len(parts) >= 2 && parts[0] == "rustc" {
		return parts[1]
	}
	return "unknown"
}

// determineSource determines the installation source based on path
func (p *RustProvider) determineSource(path string) core.InstallSource {
	if strings.Contains(path, ".cargo/bin") {
		return core.SourceVersionManager // Rustup is the standard
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *RustProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager && strings.Contains(path, ".cargo") {
		if idx := strings.Index(path, ".cargo"); idx != -1 {
			return path[:idx+6]
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for Rust ecosystem
func (p *RustProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// Rustup toolchains
	rustupPath := "~/.rustup/toolchains"
	if scanner.PathExists(rustupPath) {
		size, _ := scanner.CalculateDirSize(rustupPath)
		items = append(items, core.DiskUsageItem{
			Path:        rustupPath,
			Description: "Rustup Toolchains",
			Size:        size,
		})
	}

	// Cargo registry (the big one!)
	cargoRegistry := "~/.cargo/registry"
	if scanner.PathExists(cargoRegistry) {
		size, _ := scanner.CalculateDirSize(cargoRegistry)
		items = append(items, core.DiskUsageItem{
			Path:        cargoRegistry,
			Description: "Cargo Registry",
			Size:        size,
		})
	}

	// Cargo git checkouts
	cargoGit := "~/.cargo/git"
	if scanner.PathExists(cargoGit) {
		size, _ := scanner.CalculateDirSize(cargoGit)
		items = append(items, core.DiskUsageItem{
			Path:        cargoGit,
			Description: "Cargo Git Checkouts",
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
func (p *RustProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	envVars := []string{"CARGO_HOME", "RUSTUP_HOME"}
	for _, name := range envVars {
		if value := scanner.GetEnvVar(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// GetCleanableItems returns items that can be cleaned for Rust
func (p *RustProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// Cargo registry (safe - can be re-downloaded)
	cargoRegistry := "~/.cargo/registry"
	if scanner.PathExists(cargoRegistry) {
		size, _ := scanner.CalculateDirSize(cargoRegistry)
		items = append(items, core.CleanableItem{
			Path:        cargoRegistry,
			Description: "Cargo Registry",
			Size:        size,
			Safe:        true,
		})
	}

	// Cargo git checkouts (safe)
	cargoGit := "~/.cargo/git"
	if scanner.PathExists(cargoGit) {
		size, _ := scanner.CalculateDirSize(cargoGit)
		items = append(items, core.CleanableItem{
			Path:        cargoGit,
			Description: "Cargo Git Checkouts",
			Size:        size,
			Safe:        true,
		})
	}

	return items, nil
}

// Clean executes cleaning for Rust
func (p *RustProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if item.Path != "" {
			// Remove directory
			expandedPath := scanner.ExpandHome(item.Path)
			if err := os.RemoveAll(expandedPath); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to clean %s: %w", item.Description, err))
				continue
			}
		} else if item.Command != "" {
			// Execute command
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

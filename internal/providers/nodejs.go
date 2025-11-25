package providers

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nguyentantai21042004/dependency-hell-cli/internal/core"
	"github.com/nguyentantai21042004/dependency-hell-cli/internal/scanner"
)

// NodeProvider implements the LanguageProvider interface for Node.js
type NodeProvider struct{}

// NewNodeProvider creates a new Node.js provider
func NewNodeProvider() *NodeProvider {
	return &NodeProvider{}
}

// Name returns the name of the language
func (p *NodeProvider) Name() string {
	return "Node.js"
}

// DetectInstalled detects installed Node.js versions
func (p *NodeProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if node is installed
	nodePath, err := scanner.FindExecutable("node")
	if err != nil {
		return nil, fmt.Errorf("node not found in PATH")
	}

	// Resolve symlinks
	realPath, err := scanner.ResolveSymlink(nodePath)
	if err != nil {
		realPath = nodePath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("node", "--version")
	if err != nil {
		return nil, fmt.Errorf("failed to get node version: %w", err)
	}

	version = strings.TrimSpace(version)

	// Determine source
	source := p.determineSource(realPath)

	installation := core.Installation{
		Version:     version,
		Source:      source,
		BinaryPath:  nodePath,
		ManagerPath: p.getManagerPath(realPath, source),
	}

	return []core.Installation{installation}, nil
}

// determineSource determines the installation source based on path
func (p *NodeProvider) determineSource(path string) core.InstallSource {
	if strings.Contains(path, ".nvm") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, ".volta") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *NodeProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager {
		if strings.Contains(path, ".nvm") {
			if idx := strings.Index(path, ".nvm"); idx != -1 {
				return path[:idx+4]
			}
		}
		if strings.Contains(path, ".volta") {
			if idx := strings.Index(path, ".volta"); idx != -1 {
				return path[:idx+6]
			}
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for Node.js ecosystem caches
func (p *NodeProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// NVM versions
	nvmPath := "~/.nvm/versions"
	if scanner.PathExists(nvmPath) {
		size, _ := scanner.CalculateDirSize(nvmPath)
		items = append(items, core.DiskUsageItem{
			Path:        nvmPath,
			Description: "NVM Versions",
			Size:        size,
		})
	}

	// NPM cache
	npmCache := "~/.npm/_cacache"
	if scanner.PathExists(npmCache) {
		size, _ := scanner.CalculateDirSize(npmCache)
		items = append(items, core.DiskUsageItem{
			Path:        npmCache,
			Description: "NPM Cache",
			Size:        size,
		})
	}

	// Yarn cache (macOS)
	yarnCache := "~/Library/Caches/Yarn"
	if scanner.PathExists(yarnCache) {
		size, _ := scanner.CalculateDirSize(yarnCache)
		items = append(items, core.DiskUsageItem{
			Path:        yarnCache,
			Description: "Yarn Cache",
			Size:        size,
		})
	}

	// Yarn v2+ cache
	yarnV2Cache := "~/.yarn"
	if scanner.PathExists(yarnV2Cache) {
		size, _ := scanner.CalculateDirSize(yarnV2Cache)
		items = append(items, core.DiskUsageItem{
			Path:        yarnV2Cache,
			Description: "Yarn v2+ Cache",
			Size:        size,
		})
	}

	// PNPM store (the big one!)
	pnpmStore := "~/.local/share/pnpm/store"
	if scanner.PathExists(pnpmStore) {
		size, _ := scanner.CalculateDirSize(pnpmStore)
		items = append(items, core.DiskUsageItem{
			Path:        pnpmStore,
			Description: "PNPM Store",
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
func (p *NodeProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	// Common Node.js environment variables
	envVars := []string{"NODE_PATH", "NPM_CONFIG_PREFIX", "NVM_DIR"}
	for _, name := range envVars {
		if value := scanner.GetEnvVar(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// GetCleanableItems returns items that can be cleaned for Node.js
func (p *NodeProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// NPM cache (safe)
	npmCache := "~/.npm/_cacache"
	if scanner.PathExists(npmCache) {
		size, _ := scanner.CalculateDirSize(npmCache)
		items = append(items, core.CleanableItem{
			Description: "NPM Cache",
			Command:     "npm cache clean --force",
			Size:        size,
			Safe:        true,
		})
	}

	// Yarn cache (safe)
	yarnCache := "~/Library/Caches/Yarn"
	if scanner.PathExists(yarnCache) {
		size, _ := scanner.CalculateDirSize(yarnCache)
		items = append(items, core.CleanableItem{
			Description: "Yarn Cache",
			Command:     "yarn cache clean",
			Size:        size,
			Safe:        true,
		})
	}

	// PNPM store (safe - pnpm store prune removes unreferenced packages)
	pnpmStore := "~/.local/share/pnpm/store"
	if scanner.PathExists(pnpmStore) {
		size, _ := scanner.CalculateDirSize(pnpmStore)
		items = append(items, core.CleanableItem{
			Description: "PNPM Store",
			Command:     "pnpm store prune",
			Size:        size,
			Safe:        true,
		})
	}

	return items, nil
}

// Clean executes cleaning for Node.js
func (p *NodeProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
	result := &core.CleanResult{
		ItemsCleaned:   0,
		SpaceReclaimed: 0,
		Errors:         []error{},
	}

	for _, item := range items {
		if item.Command != "" {
			// Execute clean command
			parts := strings.Fields(item.Command)
			cmd := exec.Command(parts[0], parts[1:]...)
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

package providers

import (
	"fmt"
	"os"
	"strings"

	"github.com/nguyentantai21042004/dependency-hell-cli/internal/core"
	"github.com/nguyentantai21042004/dependency-hell-cli/internal/scanner"
)

// JavaProvider implements the LanguageProvider interface for Java
type JavaProvider struct{}

// NewJavaProvider creates a new Java provider
func NewJavaProvider() *JavaProvider {
	return &JavaProvider{}
}

// Name returns the name of the language
func (p *JavaProvider) Name() string {
	return "Java"
}

// DetectInstalled detects installed Java versions
func (p *JavaProvider) DetectInstalled() ([]core.Installation, error) {
	// Check if java is installed
	javaPath, err := scanner.FindExecutable("java")
	if err != nil {
		return nil, fmt.Errorf("java not found in PATH")
	}

	// Resolve symlinks
	realPath, err := scanner.ResolveSymlink(javaPath)
	if err != nil {
		realPath = javaPath
	}

	// Get version
	version, err := scanner.GetExecutableVersion("java", "-version")
	if err != nil {
		return nil, fmt.Errorf("failed to get java version: %w", err)
	}

	// Parse version (java -version outputs to stderr and has complex format)
	versionStr := p.parseVersion(version)

	// Determine source
	source := p.determineSource(realPath)

	installation := core.Installation{
		Version:     versionStr,
		Source:      source,
		BinaryPath:  javaPath,
		ManagerPath: p.getManagerPath(realPath, source),
	}

	return []core.Installation{installation}, nil
}

// parseVersion extracts version from java -version output
func (p *JavaProvider) parseVersion(output string) string {
	// Example output:
	// openjdk version "17.0.9" 2023-10-17
	// or: java version "1.8.0_292"

	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		// Extract version between quotes
		if start := strings.Index(firstLine, "\""); start != -1 {
			if end := strings.Index(firstLine[start+1:], "\""); end != -1 {
				return firstLine[start+1 : start+1+end]
			}
		}
	}
	return "unknown"
}

// determineSource determines the installation source based on path
func (p *JavaProvider) determineSource(path string) core.InstallSource {
	// Check JAVA_HOME first
	javaHome := os.Getenv("JAVA_HOME")

	if strings.Contains(path, ".sdkman") || strings.Contains(javaHome, ".sdkman") {
		return core.SourceVersionManager
	}
	if strings.Contains(path, "/opt/homebrew") || strings.Contains(path, "/usr/local/Cellar") {
		return core.SourceHomebrew
	}
	if strings.Contains(path, "/Library/Java") {
		return core.SourceManual
	}
	return core.SourceUnknown
}

// getManagerPath extracts the manager path if applicable
func (p *JavaProvider) getManagerPath(path string, source core.InstallSource) string {
	if source == core.SourceVersionManager && strings.Contains(path, ".sdkman") {
		if idx := strings.Index(path, ".sdkman"); idx != -1 {
			return path[:idx+7]
		}
	}
	return ""
}

// GetGlobalCacheUsage calculates disk usage for Java ecosystem
func (p *JavaProvider) GetGlobalCacheUsage() (*core.DiskUsage, error) {
	var items []core.DiskUsageItem

	// SDKMAN Java versions
	sdkmanPath := "~/.sdkman/candidates/java"
	if scanner.PathExists(sdkmanPath) {
		size, _ := scanner.CalculateDirSize(sdkmanPath)
		items = append(items, core.DiskUsageItem{
			Path:        sdkmanPath,
			Description: "SDKMAN Java SDKs",
			Size:        size,
		})
	}

	// Maven repository (the big one!)
	mavenRepo := "~/.m2/repository"
	if scanner.PathExists(mavenRepo) {
		size, _ := scanner.CalculateDirSize(mavenRepo)
		items = append(items, core.DiskUsageItem{
			Path:        mavenRepo,
			Description: "Maven Repository",
			Size:        size,
		})
	}

	// Gradle cache
	gradleCache := "~/.gradle/caches"
	if scanner.PathExists(gradleCache) {
		size, _ := scanner.CalculateDirSize(gradleCache)
		items = append(items, core.DiskUsageItem{
			Path:        gradleCache,
			Description: "Gradle Cache",
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
func (p *JavaProvider) GetEnvVars() map[string]string {
	vars := make(map[string]string)

	envVars := []string{"JAVA_HOME", "M2_HOME", "GRADLE_HOME"}
	for _, name := range envVars {
		if value := scanner.GetEnvVar(name); value != "" {
			vars[name] = value
		}
	}

	return vars
}

// GetCleanableItems returns items that can be cleaned for Java
func (p *JavaProvider) GetCleanableItems() ([]core.CleanableItem, error) {
	var items []core.CleanableItem

	// Gradle cache (safe)
	gradleCache := "~/.gradle/caches"
	if scanner.PathExists(gradleCache) {
		size, _ := scanner.CalculateDirSize(gradleCache)
		items = append(items, core.CleanableItem{
			Path:        gradleCache,
			Description: "Gradle Cache",
			Size:        size,
			Safe:        true,
		})
	}

	// Maven repository (NOT safe - requires careful consideration)
	mavenRepo := "~/.m2/repository"
	if scanner.PathExists(mavenRepo) {
		size, _ := scanner.CalculateDirSize(mavenRepo)
		items = append(items, core.CleanableItem{
			Path:        mavenRepo,
			Description: "Maven Repository",
			Size:        size,
			Safe:        false, // Requires extra confirmation
		})
	}

	return items, nil
}

// Clean executes cleaning for Java
func (p *JavaProvider) Clean(items []core.CleanableItem) (*core.CleanResult, error) {
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
		}

		result.ItemsCleaned++
		result.SpaceReclaimed += item.Size
	}

	return result, nil
}

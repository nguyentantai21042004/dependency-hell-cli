package cmd

import (
	"fmt"
	"strings"
	"sync"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/output"
	"dependency-hell-cli/internal/providers"

	"github.com/spf13/cobra"
)

var (
	langFilter string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan installed languages and their disk usage",
	Long: `Scan your development environment to discover installed languages,
their sources (Homebrew, Version Managers, System), and disk usage.

Examples:
  dhell scan                    # Scan all languages
  dhell scan --lang go          # Scan only Go
  dhell scan --lang go,node     # Scan Go and Node.js`,
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVarP(&langFilter, "lang", "l", "", "Filter languages to scan (comma-separated: go,node,java)")
}

func runScan(cmd *cobra.Command, args []string) {
	// Initialize all providers
	allProviders := []core.LanguageProvider{
		providers.NewGoProvider(),
		providers.NewNodeProvider(),
		providers.NewJavaProvider(),
		providers.NewPythonProvider(),
		providers.NewPHPProvider(),
		providers.NewRustProvider(),
	}

	// Filter providers if --lang flag is set
	selectedProviders := filterProviders(allProviders, langFilter)

	if len(selectedProviders) == 0 {
		fmt.Println("No languages selected to scan.")
		return
	}

	// Show scanning message
	if verbose {
		fmt.Println("Scanning development environment...")
		fmt.Println()
	}

	// Scan all providers concurrently
	results := scanProviders(selectedProviders)

	// Render results
	output := output.RenderScanResults(results)
	fmt.Println(output)
}

// filterProviders filters providers based on language filter
func filterProviders(providers []core.LanguageProvider, filter string) []core.LanguageProvider {
	if filter == "" {
		return providers
	}

	// Parse filter
	langs := strings.Split(strings.ToLower(filter), ",")
	langMap := make(map[string]bool)
	for _, lang := range langs {
		langMap[strings.TrimSpace(lang)] = true
	}

	// Filter providers
	var filtered []core.LanguageProvider
	for _, provider := range providers {
		name := strings.ToLower(provider.Name())
		// Check if name contains any of the filter terms
		for filterLang := range langMap {
			if strings.Contains(name, filterLang) {
				filtered = append(filtered, provider)
				break
			}
		}
	}

	return filtered
}

// scanProviders scans all providers concurrently
func scanProviders(providers []core.LanguageProvider) []output.ScanResult {
	var wg sync.WaitGroup
	results := make([]output.ScanResult, len(providers))

	for i, provider := range providers {
		wg.Add(1)
		go func(index int, p core.LanguageProvider) {
			defer wg.Done()
			results[index] = scanProvider(p)
		}(i, provider)
	}

	wg.Wait()
	return results
}

// scanProvider scans a single provider
func scanProvider(provider core.LanguageProvider) output.ScanResult {
	result := output.ScanResult{
		Provider: provider,
	}

	// Detect installation
	installations, err := provider.DetectInstalled()
	if err != nil {
		result.Error = err
		return result
	}

	if len(installations) == 0 {
		result.Error = fmt.Errorf("not installed")
		return result
	}

	// Use first installation (we can extend this later for multiple versions)
	result.Installation = &installations[0]

	// Get disk usage
	diskUsage, err := provider.GetGlobalCacheUsage()
	if err != nil {
		if verbose {
			fmt.Printf("Warning: failed to get disk usage for %s: %v\n", provider.Name(), err)
		}
		// Continue with empty disk usage
		diskUsage = &core.DiskUsage{
			Items: []core.DiskUsageItem{},
			Total: 0,
		}
	}

	result.DiskUsage = diskUsage

	return result
}

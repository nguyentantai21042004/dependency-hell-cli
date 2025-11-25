package cmd

import (
	"fmt"
	"strings"

	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/output"
	"dependency-hell-cli/internal/providers"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <language>",
	Short: "Show detailed information about a language installation",
	Long: `Display detailed information about a specific language including:
  • Version and installation source
  • Binary paths and manager locations
  • Environment variables
  • Cache locations and disk usage

Examples:
  dhell info go       # Show Go information
  dhell info node     # Show Node.js information
  dhell info python   # Show Python information`,
	Args: cobra.ExactArgs(1),
	Run:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) {
	language := strings.ToLower(args[0])

	// Initialize all providers
	allProviders := []core.LanguageProvider{
		providers.NewGoProvider(),
		providers.NewNodeProvider(),
		providers.NewJavaProvider(),
		providers.NewPythonProvider(),
		providers.NewPHPProvider(),
		providers.NewRustProvider(),
	}

	// Find matching provider
	var selectedProvider core.LanguageProvider
	for _, provider := range allProviders {
		providerName := strings.ToLower(provider.Name())
		if strings.Contains(providerName, language) {
			selectedProvider = provider
			break
		}
	}

	if selectedProvider == nil {
		fmt.Printf("Unknown language: %s\n", language)
		fmt.Println("Supported languages: go, node, java, python, php, rust")
		return
	}

	// Get installation info
	installations, err := selectedProvider.DetectInstalled()
	if err != nil {
		fmt.Printf("Error: %s is not installed or not found in PATH\n", selectedProvider.Name())
		return
	}

	if len(installations) == 0 {
		fmt.Printf("%s is not installed\n", selectedProvider.Name())
		return
	}

	installation := &installations[0]

	// Get disk usage
	diskUsage, err := selectedProvider.GetGlobalCacheUsage()
	if err != nil {
		if verbose {
			fmt.Printf("Warning: failed to get disk usage: %v\n", err)
		}
		diskUsage = &core.DiskUsage{
			Items: []core.DiskUsageItem{},
			Total: 0,
		}
	}

	// Render info
	info := output.RenderInfo(selectedProvider, installation, diskUsage)
	fmt.Println(info)
}

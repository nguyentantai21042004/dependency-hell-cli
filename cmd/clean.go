package cmd

import (
	"fmt"
	"strings"

	"dependency-hell-cli/internal/cleaner"
	"dependency-hell-cli/internal/core"
	"dependency-hell-cli/internal/output"
	"dependency-hell-cli/internal/providers"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
	force  bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean <language>",
	Short: "Clean caches for a specific language",
	Long: `Clean caches and temporary files for development tools.

This command helps you reclaim disk space by cleaning:
  • Module/package caches
  • Build caches
  • Package manager stores

Examples:
  dhell clean go                   # Clean Go caches
  dhell clean node --dry-run       # Preview Node.js cleaning
  dhell clean java --force         # Clean Java without confirmation
  dhell clean all                  # Clean all languages`,
	Args: cobra.ExactArgs(1),
	Run:  runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be deleted without actually deleting")
	cleanCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts (use with caution)")
}

func runClean(cmd *cobra.Command, args []string) {
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

	// Select providers based on language argument
	var selectedProviders []core.LanguageProvider
	if language == "all" {
		selectedProviders = allProviders
	} else {
		for _, provider := range allProviders {
			providerName := strings.ToLower(provider.Name())
			if strings.Contains(providerName, language) {
				selectedProviders = append(selectedProviders, provider)
				break
			}
		}
	}

	if len(selectedProviders) == 0 {
		fmt.Printf("Unknown language: %s\n", language)
		fmt.Println("Supported languages: go, node, java, python, php, rust, all")
		return
	}

	// Clean each selected provider
	for _, provider := range selectedProviders {
		if err := cleanProvider(provider); err != nil {
			fmt.Printf("Error cleaning %s: %v\n", provider.Name(), err)
		}
	}
}

func cleanProvider(provider core.LanguageProvider) error {
	// Get cleanable items
	items, err := provider.GetCleanableItems()
	if err != nil {
		return fmt.Errorf("failed to get cleanable items: %w", err)
	}

	if len(items) == 0 {
		fmt.Printf("No cleanable items found for %s\n", provider.Name())
		return nil
	}

	// Calculate total size
	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}

	// Dry-run mode: just show preview
	if dryRun {
		preview := output.RenderCleanPreview(provider.Name(), items)
		fmt.Println(preview)
		return nil
	}

	// Check for unsafe items
	hasUnsafeItems := false
	for _, item := range items {
		if !item.Safe {
			hasUnsafeItems = true
			break
		}
	}

	// Show confirmation unless --force is used
	if !force {
		if hasUnsafeItems {
			fmt.Println()
			fmt.Println("⚠️  WARNING: Some items require careful consideration!")
			fmt.Println()
		}

		if !cleaner.ConfirmClean(items, totalSize) {
			fmt.Println("Cleaning cancelled.")
			return nil
		}
	}

	// Execute cleaning
	if verbose {
		fmt.Printf("Cleaning %s...\n", provider.Name())
	}

	result, err := provider.Clean(items)
	if err != nil {
		return fmt.Errorf("cleaning failed: %w", err)
	}

	// Show results
	resultOutput := output.RenderCleanResult(result, items)
	fmt.Println(resultOutput)

	return nil
}

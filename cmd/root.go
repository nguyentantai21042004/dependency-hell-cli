package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "dhell",
	Short: "Dependency Hell Analyzer - Map, Measure, and Master your Dev Environment",
	Long: `D-Hell CLI is a comprehensive tool to discover, classify, and audit 
development environment dependencies across multiple programming languages.

It helps you understand:
  • What languages/runtimes are installed
  • Where they came from (Homebrew, Version Managers, System)
  • How much disk space they're consuming
  • Environment variables and configurations`,
	Version: version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

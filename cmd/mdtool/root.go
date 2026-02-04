package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mdtool",
	Short: "A CLI tool for Markdown transformations",
	Long: `mdtool is a pure Go CLI application that handles various Markdown transformations:
- Convert HTML to Markdown
- Convert web pages to Markdown (with readability)
- Convert PDF to Markdown
- Convert Markdown to PDF

All conversions use pure Go libraries with no CGO or external dependencies.`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

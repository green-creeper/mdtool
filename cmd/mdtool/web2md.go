package cmd

import (
	"fmt"
	"os"

	"github.com/andrii/mdtool/internal/scraper"
	"github.com/andrii/mdtool/pkg/models"
	"github.com/spf13/cobra"
)

var web2mdCmd = &cobra.Command{
	Use:   "web2md [URL] [output.md]",
	Short: "Convert web page to Markdown",
	Long:  `Fetch a web page, extract the main content using readability, and convert to Markdown.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runWeb2MD,
}

func init() {
	rootCmd.AddCommand(web2mdCmd)
}

func runWeb2MD(cmd *cobra.Command, args []string) error {
	conv := scraper.NewWeb2MDConverter()

	url := args[0]
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	}

	// Setup output
	var output *os.File
	var err error
	if outputFile == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	}

	// Convert
	req := &models.ConvertRequest{
		Input:  nil, // Not used for web scraping
		Output: output,
		Options: map[string]interface{}{
			"url": url,
		},
	}

	fmt.Fprintf(os.Stderr, "Fetching %s...\n", url)
	resp := conv.Convert(req)
	if !resp.Success {
		return resp.Error
	}

	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "✓ Successfully converted %s to %s\n", url, outputFile)
		if title, ok := resp.Metadata["title"]; ok {
			fmt.Fprintf(os.Stderr, "  Title: %s\n", title)
		}
	}

	return nil
}

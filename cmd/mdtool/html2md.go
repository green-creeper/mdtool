package cmd

import (
	"fmt"
	"os"

	"github.com/andrii/mdtool/internal/converter"
	"github.com/andrii/mdtool/pkg/models"
	"github.com/spf13/cobra"
)

var html2mdCmd = &cobra.Command{
	Use:   "html2md [input.html] [output.md]",
	Short: "Convert HTML to Markdown",
	Long:  `Convert an HTML file or stdin to Markdown format.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runHTML2MD,
}

func init() {
	rootCmd.AddCommand(html2mdCmd)
}

func runHTML2MD(cmd *cobra.Command, args []string) error {
	conv := converter.NewHTML2MDConverter()

	var inputFile, outputFile string
	if len(args) > 0 {
		inputFile = args[0]
	}
	if len(args) > 1 {
		outputFile = args[1]
	}

	// Setup input
	var input *os.File
	var err error
	if inputFile == "" {
		input = os.Stdin
	} else {
		input, err = os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer input.Close()
	}

	// Setup output
	var output *os.File
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
		Input:   input,
		Output:  output,
		Options: make(map[string]interface{}),
	}

	resp := conv.Convert(req)
	if !resp.Success {
		return resp.Error
	}

	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "✓ Successfully converted %s to %s\n", inputFile, outputFile)
	}

	return nil
}

package cmd

import (
	"fmt"
	"os"

	"github.com/green-creeper/mdtool/internal/converter"
	"github.com/green-creeper/mdtool/pkg/models"
	"github.com/spf13/cobra"
)

var pdf2mdCmd = &cobra.Command{
	Use:   "pdf2md [input.pdf] [output.md]",
	Short: "Convert PDF to Markdown",
	Long:  `Extract text from a PDF file and convert to Markdown format.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runPDF2MD,
}

func init() {
	rootCmd.AddCommand(pdf2mdCmd)
}

func runPDF2MD(cmd *cobra.Command, args []string) error {
	conv := converter.NewPDF2MDConverter()

	inputFile := args[0]
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	}

	// Setup input
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

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

	fmt.Fprintf(os.Stderr, "Converting PDF...\n")
	resp := conv.Convert(req)
	if !resp.Success {
		return resp.Error
	}

	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "✓ Successfully converted %s to %s\n", inputFile, outputFile)
		if pages, ok := resp.Metadata["pages"]; ok {
			fmt.Fprintf(os.Stderr, "  Pages: %s\n", pages)
		}
	}

	return nil
}

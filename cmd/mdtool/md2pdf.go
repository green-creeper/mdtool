package cmd

import (
	"fmt"
	"os"

	"github.com/green-creeper/mdtool/internal/converter"
	"github.com/green-creeper/mdtool/pkg/models"
	"github.com/spf13/cobra"
)

var md2pdfCmd = &cobra.Command{
	Use:   "md2pdf [input.md] [output.pdf]",
	Short: "Convert Markdown to PDF",
	Long:  `Generate a PDF document from a Markdown file.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runMD2PDF,
}

func init() {
	rootCmd.AddCommand(md2pdfCmd)
}

func runMD2PDF(cmd *cobra.Command, args []string) error {
	conv := converter.NewMD2PDFConverter()

	inputFile := args[0]
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	} else {
		// Default output filename
		outputFile = inputFile + ".pdf"
	}

	// Setup input
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	// Setup output
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Convert
	req := &models.ConvertRequest{
		Input:   input,
		Output:  output,
		Options: make(map[string]interface{}),
	}

	fmt.Fprintf(os.Stderr, "Generating PDF...\n")
	resp := conv.Convert(req)
	if !resp.Success {
		return resp.Error
	}

	fmt.Fprintf(os.Stderr, "✓ Successfully converted %s to %s\n", inputFile, outputFile)

	return nil
}

package converter

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/andrii/mdtool/pkg/models"
)

// MD2PDFConverter converts Markdown to PDF
type MD2PDFConverter struct{}

// NewMD2PDFConverter creates a new Markdown to PDF converter
func NewMD2PDFConverter() *MD2PDFConverter {
	return &MD2PDFConverter{}
}

// Convert converts Markdown to PDF
func (c *MD2PDFConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
	// Read Markdown content
	mdBytes, err := io.ReadAll(req.Input)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to read Markdown input: %w", err),
		}
	}

	// Create PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Parse and render Markdown
	scanner := bufio.NewScanner(strings.NewReader(string(mdBytes)))
	for scanner.Scan() {
		line := scanner.Text()

		// Handle headers
		if strings.HasPrefix(line, "# ") {
			pdf.SetFont("Arial", "B", 18)
			pdf.MultiCell(0, 10, strings.TrimPrefix(line, "# "), "", "", false)
			pdf.Ln(2)
			pdf.SetFont("Arial", "", 12)
		} else if strings.HasPrefix(line, "## ") {
			pdf.SetFont("Arial", "B", 16)
			pdf.MultiCell(0, 10, strings.TrimPrefix(line, "## "), "", "", false)
			pdf.Ln(2)
			pdf.SetFont("Arial", "", 12)
		} else if strings.HasPrefix(line, "### ") {
			pdf.SetFont("Arial", "B", 14)
			pdf.MultiCell(0, 10, strings.TrimPrefix(line, "### "), "", "", false)
			pdf.Ln(2)
			pdf.SetFont("Arial", "", 12)
		} else if strings.HasPrefix(line, "---") {
			// Horizontal rule
			pdf.Ln(2)
			x, y := pdf.GetXY()
			pdf.Line(x, y, 200, y)
			pdf.Ln(4)
		} else if strings.TrimSpace(line) == "" {
			// Empty line
			pdf.Ln(4)
		} else {
			// Regular paragraph
			// Handle bold and italic (simple approach)
			processedLine := line
			processedLine = strings.ReplaceAll(processedLine, "**", "")
			processedLine = strings.ReplaceAll(processedLine, "*", "")
			pdf.MultiCell(0, 6, processedLine, "", "", false)
		}
	}

	if err := scanner.Err(); err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("error reading Markdown: %w", err),
		}
	}

	// Write PDF to output
	err = pdf.Output(req.Output)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to write PDF: %w", err),
		}
	}

	return &models.ConvertResponse{
		Success: true,
		Metadata: map[string]string{
			"converter": "md2pdf",
		},
	}
}

// Name returns the converter name
func (c *MD2PDFConverter) Name() string {
	return "Markdown to PDF Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *MD2PDFConverter) SupportedFormats() (string, string) {
	return "markdown", "pdf"
}

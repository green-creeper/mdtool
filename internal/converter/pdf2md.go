package converter

import (
	"fmt"
	"io"
	"strings"

	"github.com/green-creeper/mdtool/pkg/models"
	"github.com/ledongthuc/pdf"
)

// PDF2MDConverter converts PDF to Markdown
type PDF2MDConverter struct{}

// NewPDF2MDConverter creates a new PDF to Markdown converter
func NewPDF2MDConverter() *PDF2MDConverter {
	return &PDF2MDConverter{}
}

// Convert extracts text from PDF and converts to Markdown
func (c *PDF2MDConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
	// Read PDF content into a temporary buffer
	pdfBytes, err := io.ReadAll(req.Input)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to read PDF input: %w", err),
		}
	}

	// Create a ReaderAt from bytes
	reader := &bytesReaderAt{data: pdfBytes}

	// Open PDF
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to open PDF: %w", err),
		}
	}

	var markdown strings.Builder

	// Extract text from each page
	numPages := pdfReader.NumPage()
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Extract text from page
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		// Add page separator for multi-page docs
		if pageNum > 1 {
			markdown.WriteString("\n\n---\n\n")
		}

		markdown.WriteString(fmt.Sprintf("## Page %d\n\n", pageNum))
		markdown.WriteString(text)
		markdown.WriteString("\n")
	}

	// Write output
	_, err = req.Output.Write([]byte(markdown.String()))
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to write output: %w", err),
		}
	}

	return &models.ConvertResponse{
		Success: true,
		Metadata: map[string]string{
			"converter": "pdf2md",
			"pages":     fmt.Sprintf("%d", numPages),
		},
	}
}

// Name returns the converter name
func (c *PDF2MDConverter) Name() string {
	return "PDF to Markdown Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *PDF2MDConverter) SupportedFormats() (string, string) {
	return "pdf", "markdown"
}

// bytesReaderAt implements io.ReaderAt for byte slices
type bytesReaderAt struct {
	data []byte
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}

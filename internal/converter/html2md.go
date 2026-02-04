package converter

import (
	"fmt"
	"io"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/andrii/mdtool/pkg/models"
)

// HTML2MDConverter converts HTML to Markdown
type HTML2MDConverter struct {
	converter *md.Converter
}

// NewHTML2MDConverter creates a new HTML to Markdown converter
func NewHTML2MDConverter() *HTML2MDConverter {
	converter := md.NewConverter("", true, nil)
	return &HTML2MDConverter{
		converter: converter,
	}
}

// Convert converts HTML to Markdown
func (c *HTML2MDConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
	// Read HTML content
	htmlBytes, err := io.ReadAll(req.Input)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to read HTML input: %w", err),
		}
	}

	// Convert to Markdown
	markdown, err := c.converter.ConvertString(string(htmlBytes))
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to convert HTML to Markdown: %w", err),
		}
	}

	// Write output
	_, err = req.Output.Write([]byte(markdown))
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to write Markdown output: %w", err),
		}
	}

	return &models.ConvertResponse{
		Success: true,
		Metadata: map[string]string{
			"converter": "html2md",
		},
	}
}

// Name returns the converter name
func (c *HTML2MDConverter) Name() string {
	return "HTML to Markdown Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *HTML2MDConverter) SupportedFormats() (string, string) {
	return "html", "markdown"
}

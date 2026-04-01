package converter

import (
	"testing"
)

func TestConverters(t *testing.T) {
	tests := []struct {
		name          string
		converter     Converter
		expectedName  string
		expectedSrc   string
		expectedTgt   string
	}{
		{
			name:          "HTML to Markdown",
			converter:     NewHTML2MDConverter(),
			expectedName:  "HTML to Markdown Converter",
			expectedSrc:   "html",
			expectedTgt:   "markdown",
		},
		{
			name:          "Markdown to PDF",
			converter:     NewMD2PDFConverter(),
			expectedName:  "Markdown to PDF Converter",
			expectedSrc:   "markdown",
			expectedTgt:   "pdf",
		},
		{
			name:          "PDF to Markdown",
			converter:     NewPDF2MDConverter(),
			expectedName:  "PDF to Markdown Converter",
			expectedSrc:   "pdf",
			expectedTgt:   "markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.converter.Name() != tt.expectedName {
				t.Errorf("Expected Name() to be %q, got %q", tt.expectedName, tt.converter.Name())
			}

			src, tgt := tt.converter.SupportedFormats()
			if src != tt.expectedSrc {
				t.Errorf("Expected source format to be %q, got %q", tt.expectedSrc, src)
			}
			if tgt != tt.expectedTgt {
				t.Errorf("Expected target format to be %q, got %q", tt.expectedTgt, tgt)
			}
		})
	}
}

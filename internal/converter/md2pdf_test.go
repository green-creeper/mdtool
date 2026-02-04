package converter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrii/mdtool/pkg/models"
)

func TestMD2PDFConverter_Convert(t *testing.T) {
	c := NewMD2PDFConverter()

	markdown := `
# Document Title
## Section 1
This is a paragraph with **bold** and *italic* text.

---

### End of Document
`

	input := strings.NewReader(markdown)
	var output bytes.Buffer
	req := &models.ConvertRequest{
		Input:  input,
		Output: &output,
	}

	resp := c.Convert(req)

	if !resp.Success {
		t.Fatalf("Convert() failed: %v", resp.Error)
	}

	// Check if output is a PDF (starts with %PDF-)
	pdfBytes := output.Bytes()
	if len(pdfBytes) < 5 || !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Error("Output does not appear to be a valid PDF file")
	}
}

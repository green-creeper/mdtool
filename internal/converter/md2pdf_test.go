package converter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/green-creeper/mdtool/pkg/models"
)

// errorReader is an io.Reader that always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

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

func TestMD2PDFConverter_Convert_ReadError(t *testing.T) {
	c := NewMD2PDFConverter()

	expectedErr := errors.New("read error")
	input := &errorReader{err: expectedErr}
	var output bytes.Buffer
	req := &models.ConvertRequest{
		Input:  input,
		Output: &output,
	}

	resp := c.Convert(req)

	if resp.Success {
		t.Fatal("Convert() should have failed")
	}

	expectedErrMsg := fmt.Sprintf("failed to read Markdown input: %v", expectedErr)
	if resp.Error == nil || resp.Error.Error() != expectedErrMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrMsg, resp.Error)
	}
}

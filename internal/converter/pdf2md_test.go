package converter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/green-creeper/mdtool/pkg/models"
)

func TestPDF2MDConverter_Convert(t *testing.T) {
	c := NewPDF2MDConverter()

	t.Run("invalid pdf bytes", func(t *testing.T) {
		input := bytes.NewReader([]byte("this is not a pdf"))
		var output bytes.Buffer
		req := &models.ConvertRequest{
			Input:  input,
			Output: &output,
		}

		resp := c.Convert(req)

		if resp.Success {
			t.Errorf("Convert() expected success=false, got true")
		}

		if resp.Error == nil {
			t.Fatal("Convert() expected error, got nil")
		}

		expectedError := "failed to open PDF"
		if !strings.Contains(resp.Error.Error(), expectedError) {
			t.Errorf("Convert() error = %v, expected to contain %v", resp.Error, expectedError)
		}
	})
}

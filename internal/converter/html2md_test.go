package converter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrii/mdtool/pkg/models"
)

func TestHTML2MDConverter_Convert(t *testing.T) {
	c := NewHTML2MDConverter()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple html",
			html:     "<h1>Hello</h1><p>World</p>",
			expected: "# Hello\n\nWorld",
		},
		{
			name:     "links",
			html:     "<a href=\"https://example.com\">Example</a>",
			expected: "[Example](https://example.com)",
		},
		{
			name:     "lists",
			html:     "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "- Item 1\n- Item 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.html)
			var output bytes.Buffer
			req := &models.ConvertRequest{
				Input:  input,
				Output: &output,
			}

			resp := c.Convert(req)

			if !resp.Success {
				t.Errorf("Convert() failed: %v", resp.Error)
			}

			if !strings.Contains(output.String(), tt.expected) {
				t.Errorf("Convert() = %v, expected to contain %v", output.String(), tt.expected)
			}
		})
	}
}

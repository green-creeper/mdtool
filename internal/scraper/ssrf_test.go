package scraper

import (
	"bytes"
	"strings"
	"testing"

	"github.com/green-creeper/mdtool/pkg/models"
)

func TestWeb2MDConverter_SSRF_Scheme(t *testing.T) {
	c := NewWeb2MDConverter()

	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: "invalid URL scheme",
		},
		{
			name:    "ftp scheme",
			url:     "ftp://example.com/file",
			wantErr: "invalid URL scheme",
		},
		{
			name:    "no scheme",
			url:     "example.com",
			wantErr: "invalid URL scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			req := &models.ConvertRequest{
				Output: &output,
				Options: map[string]interface{}{
					"url": tt.url,
				},
			}

			resp := c.Convert(req)

			if resp.Success {
				t.Errorf("Expected failure for URL %s, but got success", tt.url)
			} else if resp.Error == nil || !strings.Contains(resp.Error.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantErr, resp.Error)
			}
		})
	}
}

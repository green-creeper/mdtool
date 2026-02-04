package scraper

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrii/mdtool/pkg/models"
)

func TestWeb2MDConverter_Convert(t *testing.T) {
	// Create a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html><head><title>Test Title</title></head><body><article><h1>Main Heading</h1><p>Some content.</p></article></body></html>")
	}))
	defer ts.Close()

	c := NewWeb2MDConverter()

	var output bytes.Buffer
	req := &models.ConvertRequest{
		Output: &output,
		Options: map[string]interface{}{
			"url": ts.URL,
		},
	}

	resp := c.Convert(req)

	if !resp.Success {
		t.Fatalf("Convert() failed: %v", resp.Error)
	}

	result := output.String()

	// Check for title in header
	if !strings.Contains(result, "# Test Title") {
		t.Errorf("Expected result to contain title '# Test Title', got:\n%s", result)
	}

	// Check for content
	if !strings.Contains(result, "Main Heading") {
		t.Errorf("Expected result to contain 'Main Heading', got:\n%s", result)
	}

	if !strings.Contains(result, "Some content.") {
		t.Errorf("Expected result to contain 'Some content.', got:\n%s", result)
	}

	// Check for source link
	if !strings.Contains(result, ts.URL) {
		t.Errorf("Expected result to contain URL, got:\n%s", result)
	}
}

func TestWeb2MDConverter_NoURL(t *testing.T) {
	c := NewWeb2MDConverter()

	var output bytes.Buffer
	req := &models.ConvertRequest{
		Output:  &output,
		Options: map[string]interface{}{},
	}

	resp := c.Convert(req)

	if resp.Success {
		t.Error("Expected Convert() to fail when no URL is provided")
	}

	if resp.Error == nil || !strings.Contains(resp.Error.Error(), "URL is required") {
		t.Errorf("Expected 'URL is required' error, got: %v", resp.Error)
	}
}

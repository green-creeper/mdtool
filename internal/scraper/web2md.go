package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	readability "github.com/go-shiori/go-readability"
	"github.com/green-creeper/mdtool/pkg/models"
)

const (
	// MaxResponseSize limits the response body size to prevent memory exhaustion (50MB)
	MaxResponseSize = 50 * 1024 * 1024
)

// Web2MDConverter fetches web content and converts it to Markdown
type Web2MDConverter struct {
	httpClient  *http.Client
	mdConverter *md.Converter
}

// NewWeb2MDConverter creates a new Web to Markdown converter
func NewWeb2MDConverter() *Web2MDConverter {
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.Table())
	return &Web2MDConverter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
					return fmt.Errorf("invalid redirect URL scheme: %s (only http and https are allowed)", req.URL.Scheme)
				}
				return nil
			},
		},
		mdConverter: converter,
	}
}

// Convert fetches a URL and converts the main content to Markdown
func (c *Web2MDConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
	// Get URL from options
	urlStr, ok := req.Options["url"].(string)
	if !ok || urlStr == "" {
		return &models.ConvertResponse{
			Success: false,
			Error:   errors.New("URL is required in options"),
		}
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to parse URL: %w", err),
		}
	}

	// Only allow http and https schemes to prevent SSRF
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("invalid URL scheme: %s (only http and https are allowed)", parsedURL.Scheme),
		}
	}

	// Fetch the web page
	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to fetch URL: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode),
		}
	}

	// Limit response size to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)

	// Read the response body with size limit
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to read response body: %w", err),
		}
	}

	// Check if we hit the size limit
	if len(bodyBytes) == MaxResponseSize {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("response body exceeds maximum size of %d bytes", MaxResponseSize),
		}
	}

	// Parse with readability to extract main content
	article, err := readability.FromReader(bytes.NewReader(bodyBytes), parsedURL)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to parse article content: %w", err),
		}
	}

	// Convert HTML to Markdown
	markdown, err := c.mdConverter.ConvertString(article.Content)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to convert to Markdown: %w", err),
		}
	}

	// Clear large byte slice to help GC
	bodyBytes = nil

	// Add article metadata as header
	var builder strings.Builder
	// Calculate total size:
	// "# " (2) + title + "\n\n" (2)
	// optional: "*By " (4) + byline + "*\n\n" (3)
	// "*Source: [" (10) + url + "](" (2) + url + ")*\n\n---\n\n" (11)
	// markdown
	size := 2 + len(article.Title) + 2 + 10 + len(urlStr) + 2 + len(urlStr) + 11 + len(markdown)
	if article.Byline != "" {
		size += 4 + len(article.Byline) + 3
	}
	builder.Grow(size)

	builder.WriteString("# ")
	builder.WriteString(article.Title)
	builder.WriteString("\n\n")

	if article.Byline != "" {
		builder.WriteString("*By ")
		builder.WriteString(article.Byline)
		builder.WriteString("*\n\n")
	}

	builder.WriteString("*Source: [")
	builder.WriteString(urlStr)
	builder.WriteString("](")
	builder.WriteString(urlStr)
	builder.WriteString(")*\n\n---\n\n")
	builder.WriteString(markdown)

	fullMarkdown := builder.String()

	// Write output
	_, err = req.Output.Write([]byte(fullMarkdown))
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to write output: %w", err),
		}
	}

	return &models.ConvertResponse{
		Success: true,
		Metadata: map[string]string{
			"converter": "web2md",
			"title":     article.Title,
			"byline":    article.Byline,
			"url":       urlStr,
		},
	}
}

// Name returns the converter name
func (c *Web2MDConverter) Name() string {
	return "Web to Markdown Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *Web2MDConverter) SupportedFormats() (string, string) {
	return "web", "markdown"
}

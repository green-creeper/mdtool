package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/green-creeper/mdtool/pkg/models"
	readability "github.com/go-shiori/go-readability"
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

	// Fetch and validate URL content
	bodyBytes, parsedURL, err := c.fetchURL(urlStr)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   err,
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
	header := fmt.Sprintf("# %s\n\n", article.Title)
	if article.Byline != "" {
		header += fmt.Sprintf("*By %s*\n\n", article.Byline)
	}
	header += fmt.Sprintf("*Source: [%s](%s)*\n\n---\n\n", urlStr, urlStr)

	fullMarkdown := header + markdown

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

// fetchURL handles parsing, validation, and fetching of the URL content
func (c *Web2MDConverter) fetchURL(urlStr string) ([]byte, *url.URL, error) {
	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Only allow http and https schemes to prevent SSRF
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, nil, fmt.Errorf("invalid URL scheme: %s (only http and https are allowed)", parsedURL.Scheme)
	}

	// Fetch the web page
	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Limit response size to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)

	// Read the response body with size limit
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if we hit the size limit
	if len(bodyBytes) == MaxResponseSize {
		return nil, nil, fmt.Errorf("response body exceeds maximum size of %d bytes", MaxResponseSize)
	}

	return bodyBytes, parsedURL, nil
}

// Name returns the converter name
func (c *Web2MDConverter) Name() string {
	return "Web to Markdown Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *Web2MDConverter) SupportedFormats() (string, string) {
	return "web", "markdown"
}

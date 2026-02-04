# mdtool - Markdown Transformation CLI

A pure Go CLI tool for converting between Markdown and various formats without external dependencies.

## Features

✨ **Pure Go** - No CGO, no external system dependencies (wkhtmltopdf, Pandoc, etc.)

🔄 **Multiple Conversions**:
- **PDF → Markdown**: Extract text and structure from PDF files
- **HTML → Markdown**: Convert HTML files or strings to clean Markdown
- **Web → Markdown**: Fetch URLs with readability mode and convert to Markdown
- **Markdown → PDF**: Generate PDF documents from Markdown

## Installation

```bash
# Clone the repository
git clone https://github.com/andrii/mdtool.git
cd mdtool

# Download dependencies
go mod download

# Build the binary
go build -o mdtool main.go

# Optional: Install globally
go install
```

## Usage

### HTML to Markdown

```bash
# Convert a file
mdtool html2md input.html output.md

# From stdin to stdout
cat input.html | mdtool html2md > output.md

# From file to stdout
mdtool html2md input.html
```

### Web to Markdown

```bash
# Fetch and convert a web page
mdtool web2md https://example.com/article output.md

# Output to stdout
mdtool web2md https://example.com/article
```

The web2md command uses **readability** to extract the main content, removing navigation, ads, and other boilerplate.

### PDF to Markdown

```bash
# Convert PDF to Markdown
mdtool pdf2md document.pdf output.md

# Output to stdout
mdtool pdf2md document.pdf
```

### Markdown to PDF

```bash
# Convert Markdown to PDF
mdtool md2pdf input.md output.pdf

# Auto-generate output filename (input.md.pdf)
mdtool md2pdf input.md
```

## Project Structure

```
mdtool/
├── main.go                      # Entry point
├── go.mod                       # Dependencies
├── cmd/
│   └── mdtool/                  # CLI commands
│       ├── root.go              # Root command
│       ├── html2md.go           # HTML → MD command
│       ├── web2md.go            # Web → MD command
│       ├── pdf2md.go            # PDF → MD command
│       └── md2pdf.go            # MD → PDF command
├── internal/
│   ├── converter/               # Format converters
│   │   ├── converter.go         # Converter interface
│   │   ├── html2md.go           # HTML converter
│   │   ├── pdf2md.go            # PDF extractor
│   │   └── md2pdf.go            # PDF generator
│   └── scraper/                 # Web scraping
│       └── web2md.go            # Web fetcher + converter
└── pkg/
    └── models/                  # Data models
        └── models.go            # Request/Response types
```

## Architecture

### Converter Interface

All converters implement a common interface:

```go
type Converter interface {
    Convert(req *ConvertRequest) *ConvertResponse
    Name() string
    SupportedFormats() (source, target string)
}
```

This allows easy extension for new formats.

### Provider Pattern

Each conversion is treated as a **provider** with its own implementation:
- **HTML2MDConverter**: Uses `JohannesKaufmann/html-to-markdown`
- **Web2MDConverter**: Combines HTTP client + `go-readability` + `html-to-markdown`
- **PDF2MDConverter**: Uses `ledongthuc/pdf` for text extraction
- **MD2PDFConverter**: Uses `go-pdf/fpdf` for PDF generation

## Dependencies

All dependencies are **pure Go** libraries:

| Library | Purpose | License |
|---------|---------|---------|
| [JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) | HTML to MD conversion | MIT |
| [go-shiori/go-readability](https://github.com/go-shiori/go-readability) | Readability extraction | MIT |
| [ledongthuc/pdf](https://github.com/ledongthuc/pdf) | PDF text extraction | MIT |
| [go-pdf/fpdf](https://github.com/go-pdf/fpdf) | PDF generation | MIT |
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework | Apache 2.0 |
| [PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery) | HTML parsing | BSD-3 |

## Examples

### Example 1: Convert Blog Post to Markdown

```bash
mdtool web2md https://blog.golang.org/go1.18 go1.18.md
```

### Example 2: Generate PDF Report

```bash
# Create a markdown report
cat > report.md << 'EOF'
# Monthly Report

## Summary
This month we achieved the following goals...

## Metrics
- 100% uptime
- 50% faster response times

---
*Generated on 2024-01-15*
EOF

# Convert to PDF
mdtool md2pdf report.md monthly-report.pdf
```

### Example 3: Pipeline Conversion

```bash
# Fetch web page, convert to MD, then to PDF
mdtool web2md https://example.com/article article.md
mdtool md2pdf article.md article.pdf
```

## Extending mdtool

To add a new converter:

1. Create a new file in `internal/converter/`
2. Implement the `Converter` interface
3. Add a new command in `cmd/mdtool/`
4. Register the command in `root.go`

Example stub:

```go
type DocxToMDConverter struct{}

func (c *DocxToMDConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
    // Implementation here
}

func (c *DocxToMDConverter) Name() string {
    return "DOCX to Markdown Converter"
}

func (c *DocxToMDConverter) SupportedFormats() (string, string) {
    return "docx", "markdown"
}
```

## Limitations

### PDF to Markdown
- **Text-based PDFs only**: Cannot extract text from scanned/image-based PDFs
- **Basic formatting**: Complex layouts may not be preserved
- **No images**: Text extraction only

### Markdown to PDF
- **Basic styling**: Limited to headers, paragraphs, and horizontal rules
- **No advanced markdown**: Tables, code blocks, and images are not yet supported
- **Font limitations**: Uses Arial only

### Web to Markdown
- **JavaScript-rendered content**: Cannot fetch content that requires JavaScript execution
- **Dynamic pages**: Works best with static content

## Contributing

Contributions are welcome! Areas for improvement:
- Add support for tables in MD → PDF
- Improve PDF text extraction (handle more complex layouts)
- Add DOCX/ODT support
- Add image extraction from PDFs
- Enhance Markdown parsing for PDF generation

## License Compliance

This project uses [google/go-licenses](https://github.com/google/go-licenses) to ensure compliance with dependency licenses.

To generate a report of all licenses used:

```bash
go install github.com/google/go-licenses@latest
go-licenses report ./... > licenses.csv
```

## License

MIT License - feel free to use and modify as needed.

## Author

Built with ❤️ using pure Go libraries.

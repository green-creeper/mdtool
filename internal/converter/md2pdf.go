package converter

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/green-creeper/mdtool/pkg/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// emojiRegex matches most emoji characters
// Covers emoji ranges: Emoticons, Dingbats, Symbols, etc.
var emojiRegex = regexp.MustCompile(`[\x{1F300}-\x{1F9FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{FE00}-\x{FE0F}]|[\x{1F000}-\x{1F02F}]`)

// Common emoji replacements for better readability
var emojiReplacements = map[string]string{
	"✨": "[*]",
	"🔄": "[<>]",
	"✓": "[v]",
	"✔": "[v]",
	"❌": "[x]",
	"⚠": "[!]",
	"📝": "[note]",
	"📁": "[dir]",
	"📂": "[dir]",
	"📄": "[file]",
	"🔧": "[tool]",
	"🚀": "[->]",
	"💡": "[i]",
	"🎉": "[!]",
	"👍": "[+]",
	"👎": "[-]",
}

// stripEmojis removes or replaces emoji characters that aren't supported by DejaVu fonts
func stripEmojis(s string) string {
	// First, apply known replacements
	for emoji, replacement := range emojiReplacements {
		s = strings.ReplaceAll(s, emoji, replacement)
	}
	// Then strip any remaining emojis
	return emojiRegex.ReplaceAllString(s, "")
}

// MD2PDFConverter converts Markdown to PDF
type MD2PDFConverter struct{}

// NewMD2PDFConverter creates a new Markdown to PDF converter
func NewMD2PDFConverter() *MD2PDFConverter {
	return &MD2PDFConverter{}
}

// pdfRenderer handles the PDF rendering state
type pdfRenderer struct {
	pdf    *fpdf.Fpdf
	source []byte
}

// Convert converts Markdown to PDF
func (c *MD2PDFConverter) Convert(req *models.ConvertRequest) *models.ConvertResponse {
	// Read Markdown content
	mdBytes, err := io.ReadAll(req.Input)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to read Markdown input: %w", err),
		}
	}

	// Create PDF with embedded Unicode fonts
	pdf := fpdf.New("P", "mm", "A4", "")

	// Add embedded DejaVu fonts for full Unicode support
	pdf.AddUTF8FontFromBytes("DejaVu", "", dejaVuSansFont)
	pdf.AddUTF8FontFromBytes("DejaVu", "B", dejaVuSansBoldFont)
	pdf.AddUTF8FontFromBytes("DejaVuMono", "", dejaVuSansMonoFont)
	pdf.AddUTF8FontFromBytes("DejaVuMono", "B", dejaVuSansMonoBoldFont)

	pdf.AddPage()
	pdf.SetFont("DejaVu", "", 12)

	// Parse Markdown with goldmark including GFM table extension
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	reader := text.NewReader(mdBytes)
	doc := md.Parser().Parse(reader)

	// Render the AST to PDF
	renderer := &pdfRenderer{
		pdf:    pdf,
		source: mdBytes,
	}
	renderer.renderNode(doc)

	// Write PDF to output
	err = pdf.Output(req.Output)
	if err != nil {
		return &models.ConvertResponse{
			Success: false,
			Error:   fmt.Errorf("failed to write PDF: %w", err),
		}
	}

	return &models.ConvertResponse{
		Success: true,
		Metadata: map[string]string{
			"converter": "md2pdf",
		},
	}
}

// renderNode recursively renders AST nodes to PDF
func (r *pdfRenderer) renderNode(n ast.Node) {
	switch node := n.(type) {
	case *ast.Document:
		// Just traverse children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			r.renderNode(child)
		}

	case *ast.Heading:
		r.renderHeading(node)

	case *ast.Paragraph:
		r.renderParagraph(node)

	case *ast.FencedCodeBlock:
		r.renderCodeBlock(node)

	case *ast.CodeBlock:
		r.renderCodeBlock(node)

	case *ast.ThematicBreak:
		r.renderThematicBreak()

	case *ast.List:
		r.renderList(node, 0)

	case *ast.Blockquote:
		r.renderBlockquote(node)

	case *extast.Table:
		r.renderTable(node)

	default:
		// For other block nodes, try to render children
		if n.HasChildren() {
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				r.renderNode(child)
			}
		}
	}
}

// renderHeading renders a heading with appropriate font size
func (r *pdfRenderer) renderHeading(node *ast.Heading) {
	sizes := map[int]float64{1: 20, 2: 17, 3: 14, 4: 12, 5: 11, 6: 10}
	size := sizes[node.Level]
	if size == 0 {
		size = 12
	}

	r.pdf.SetFont("DejaVu", "B", size)
	text := r.extractText(node)
	r.pdf.MultiCell(0, size*0.5, text, "", "", false)
	r.pdf.Ln(3)
	r.pdf.SetFont("DejaVu", "", 12)
}

// renderParagraph renders a paragraph with inline formatting
func (r *pdfRenderer) renderParagraph(node *ast.Paragraph) {
	text := r.extractFormattedText(node)
	r.pdf.MultiCell(0, 6, text, "", "", false)
	r.pdf.Ln(3)
}

// renderCodeBlock renders a fenced code block with monospace font
func (r *pdfRenderer) renderCodeBlock(node ast.Node) {
	// Save current position
	x, y := r.pdf.GetXY()
	pageWidth, _ := r.pdf.GetPageSize()
	marginLeft, _, marginRight, _ := r.pdf.GetMargins()
	contentWidth := pageWidth - marginLeft - marginRight

	// Extract code content
	var codeBuilder strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		codeBuilder.Write(r.source[line.Start:line.Stop])
	}
	code := strings.TrimRight(codeBuilder.String(), "\n")

	// Calculate height needed
	codeLines := strings.Split(code, "\n")
	lineHeight := 5.0
	blockHeight := float64(len(codeLines))*lineHeight + 6 // padding

	// Draw background
	r.pdf.SetFillColor(245, 245, 245) // Light gray background
	r.pdf.Rect(x, y, contentWidth, blockHeight, "F")

	// Draw border
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.Rect(x, y, contentWidth, blockHeight, "D")

	// Set monospace font (DejaVuMono supports Unicode)
	r.pdf.SetFont("DejaVuMono", "", 10)
	r.pdf.SetXY(x+3, y+3)

	// Render each line
	for _, line := range codeLines {
		r.pdf.CellFormat(contentWidth-6, lineHeight, line, "", 0, "", false, 0, "")
		r.pdf.Ln(lineHeight)
		r.pdf.SetX(x + 3)
	}

	// Reset font and position
	r.pdf.SetFont("DejaVu", "", 12)
	r.pdf.SetXY(x, y+blockHeight+3)
	r.pdf.SetDrawColor(0, 0, 0)
}

// renderThematicBreak renders a horizontal rule
func (r *pdfRenderer) renderThematicBreak() {
	r.pdf.Ln(3)
	x, y := r.pdf.GetXY()
	pageWidth, _ := r.pdf.GetPageSize()
	_, _, marginRight, _ := r.pdf.GetMargins()
	r.pdf.SetDrawColor(180, 180, 180)
	r.pdf.Line(x, y, pageWidth-marginRight, y)
	r.pdf.SetDrawColor(0, 0, 0)
	r.pdf.Ln(6)
}

// renderList renders an ordered or unordered list
func (r *pdfRenderer) renderList(node *ast.List, indent int) {
	itemNum := 1
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			indentStr := strings.Repeat("   ", indent)
			var bullet string
			if node.IsOrdered() {
				bullet = fmt.Sprintf("%d. ", itemNum)
				itemNum++
			} else {
				bullet = "• "
			}

			// Get list item text from immediate paragraph children only
			var text string
			for itemChild := listItem.FirstChild(); itemChild != nil; itemChild = itemChild.NextSibling() {
				if para, ok := itemChild.(*ast.Paragraph); ok {
					text = r.extractText(para)
					break
				} else if tc, ok := itemChild.(*ast.TextBlock); ok {
					text = r.extractText(tc)
					break
				}
			}
			if text != "" {
				r.pdf.MultiCell(0, 6, indentStr+bullet+text, "", "", false)
			}

			// Handle nested lists
			for nested := listItem.FirstChild(); nested != nil; nested = nested.NextSibling() {
				if nestedList, ok := nested.(*ast.List); ok {
					r.renderList(nestedList, indent+1)
				}
			}
		}
	}
	r.pdf.Ln(2)
}

// renderBlockquote renders a blockquote with left border
func (r *pdfRenderer) renderBlockquote(node *ast.Blockquote) {
	x, y := r.pdf.GetXY()

	// Draw left border
	r.pdf.SetDrawColor(180, 180, 180)
	r.pdf.SetLineWidth(1)

	// Get quote text
	text := r.extractText(node)
	lines := strings.Split(text, "\n")
	lineHeight := 6.0

	for _, line := range lines {
		r.pdf.Line(x, y, x, y+lineHeight)
		r.pdf.SetX(x + 5)
		r.pdf.SetTextColor(100, 100, 100)
		r.pdf.MultiCell(0, lineHeight, line, "", "", false)
		y += lineHeight
	}

	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetDrawColor(0, 0, 0)
	r.pdf.SetLineWidth(0.2)
	r.pdf.Ln(3)
}

// renderTable renders a GFM table
func (r *pdfRenderer) renderTable(node *extast.Table) {
	// Collect all rows and cells
	var rows [][]string
	var columnCount int

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if row, ok := child.(*extast.TableRow); ok {
			var cells []string
			for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					cells = append(cells, strings.TrimSpace(r.extractText(tableCell)))
				}
			}
			if len(cells) > columnCount {
				columnCount = len(cells)
			}
			rows = append(rows, cells)
		} else if header, ok := child.(*extast.TableHeader); ok {
			var cells []string
			for cell := header.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					cells = append(cells, strings.TrimSpace(r.extractText(tableCell)))
				}
			}
			if len(cells) > columnCount {
				columnCount = len(cells)
			}
			rows = append([][]string{cells}, rows...) // Header first
		}
	}

	if len(rows) == 0 || columnCount == 0 {
		return
	}

	// Calculate column widths
	pageWidth, _ := r.pdf.GetPageSize()
	marginLeft, _, marginRight, _ := r.pdf.GetMargins()
	tableWidth := pageWidth - marginLeft - marginRight
	colWidth := tableWidth / float64(columnCount)

	// Render table
	r.pdf.SetFont("DejaVu", "", 10)
	cellHeight := 7.0

	for rowIdx, row := range rows {
		// Pad row to match column count
		for len(row) < columnCount {
			row = append(row, "")
		}

		// Header row styling
		if rowIdx == 0 {
			r.pdf.SetFont("DejaVu", "B", 10)
			r.pdf.SetFillColor(230, 230, 230)
		} else {
			r.pdf.SetFont("DejaVu", "", 10)
			r.pdf.SetFillColor(255, 255, 255)
		}

		for _, cell := range row {
			r.pdf.CellFormat(colWidth, cellHeight, cell, "1", 0, "L", true, 0, "")
		}
		r.pdf.Ln(cellHeight)
	}

	r.pdf.SetFont("DejaVu", "", 12)
	r.pdf.Ln(3)
}

// extractText extracts plain text from an AST node and strips unsupported characters
func (r *pdfRenderer) extractText(node ast.Node) string {
	var buf bytes.Buffer
	r.extractTextRecursive(node, &buf)
	// Strip emojis that aren't supported by embedded fonts
	return stripEmojis(buf.String())
}

// extractTextRecursive recursively extracts text
func (r *pdfRenderer) extractTextRecursive(node ast.Node, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *ast.Text:
		buf.Write(n.Segment.Value(r.source))
		if n.HardLineBreak() || n.SoftLineBreak() {
			buf.WriteByte('\n')
		}
	case *ast.String:
		buf.Write(n.Value)
	case *ast.CodeSpan:
		// Extract inline code
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			r.extractTextRecursive(child, buf)
		}
	default:
		if node.HasChildren() {
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				r.extractTextRecursive(child, buf)
			}
		}
	}
}

// extractFormattedText extracts text and strips markdown formatting markers
func (r *pdfRenderer) extractFormattedText(node ast.Node) string {
	text := r.extractText(node)
	// Simple cleanup of any remaining formatting markers
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	return strings.TrimSpace(text)
}

// Name returns the converter name
func (c *MD2PDFConverter) Name() string {
	return "Markdown to PDF Converter"
}

// SupportedFormats returns the formats this converter supports
func (c *MD2PDFConverter) SupportedFormats() (string, string) {
	return "markdown", "pdf"
}

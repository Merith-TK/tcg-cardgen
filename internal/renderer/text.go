package renderer

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/Merith-TK/tcg-cardgen/internal/templates"
)

// TextStyle represents text formatting options
type TextStyle struct {
	Bold   bool
	Italic bool
	Size   float64
	Color  color.Color
}

// FormattedText represents a piece of text with styling
type FormattedText struct {
	Content string
	Style   TextStyle
}

// FormattedLine represents a line with multiple formatted text segments
type FormattedLine struct {
	Segments []FormattedText
	Type     string // "normal", "header", "hr" (horizontal rule)
	Level    int    // header level (1-6)
}

// TextProcessor handles all text processing operations
type TextProcessor struct {
	utils *Utils
}

// NewTextProcessor creates a new text processor
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		utils: NewUtils(),
	}
}

// ProcessMarkdown parses markdown content into formatted lines
func (tp *TextProcessor) ProcessMarkdown(content string) []FormattedLine {
	lines := strings.Split(content, "\n")
	var formattedLines []FormattedLine

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines but preserve them for spacing
		if line == "" {
			formattedLines = append(formattedLines, FormattedLine{
				Segments: []FormattedText{},
				Type:     "normal",
			})
			continue
		}

		// Check for horizontal rule
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "***") {
			formattedLines = append(formattedLines, FormattedLine{
				Type: "hr",
			})
			continue
		}

		// Check for headers
		if strings.HasPrefix(line, "#") {
			level := 0
			for i, ch := range line {
				if ch == '#' {
					level++
				} else if ch == ' ' {
					line = line[i+1:]
					break
				} else {
					level = 0
					break
				}
			}

			if level > 0 && level <= 6 {
				formattedLines = append(formattedLines, FormattedLine{
					Segments: tp.parseInlineFormatting(line),
					Type:     "header",
					Level:    level,
				})
				continue
			}
		}

		// Regular line with inline formatting
		formattedLines = append(formattedLines, FormattedLine{
			Segments: tp.parseInlineFormatting(line),
			Type:     "normal",
		})
	}

	return formattedLines
}

// parseInlineFormatting parses inline markdown formatting like **bold** and *italic*
func (tp *TextProcessor) parseInlineFormatting(text string) []FormattedText {
	// Process the text to handle nested and overlapping formats
	return tp.parseFormattingRecursive(text)
}

// parseFormattingRecursive handles nested and overlapping markdown formatting
func (tp *TextProcessor) parseFormattingRecursive(text string) []FormattedText {
	var segments []FormattedText

	// Find the first formatting marker
	pos := -1
	marker := ""
	markerLength := 0

	// Look for ***bold italic***
	if strings.Contains(text, "***") {
		if idx := strings.Index(text, "***"); idx != -1 {
			pos = idx
			marker = "***"
			markerLength = 3
		}
	}

	// Look for **bold** (only if we haven't found *** at this position)
	if (pos == -1 || pos > strings.Index(text, "**")) && strings.Contains(text, "**") {
		if idx := strings.Index(text, "**"); idx != -1 {
			pos = idx
			marker = "**"
			markerLength = 2
		}
	}

	// Look for *italic* (only if we haven't found ** or *** at this position)
	if (pos == -1 || pos > strings.Index(text, "*")) && strings.Contains(text, "*") {
		if idx := strings.Index(text, "*"); idx != -1 {
			pos = idx
			marker = "*"
			markerLength = 1
		}
	}

	if pos == -1 {
		// No formatting found, return as plain text
		if text != "" {
			segments = append(segments, FormattedText{
				Content: text,
				Style:   TextStyle{Bold: false, Italic: false},
			})
		}
		return segments
	}

	// Add text before the marker as plain text
	if pos > 0 {
		segments = append(segments, FormattedText{
			Content: text[:pos],
			Style:   TextStyle{Bold: false, Italic: false},
		})
	}

	// Find the closing marker
	remaining := text[pos+markerLength:]
	closePos := strings.Index(remaining, marker)

	if closePos == -1 {
		// No closing marker, treat as plain text
		segments = append(segments, FormattedText{
			Content: text[pos:],
			Style:   TextStyle{Bold: false, Italic: false},
		})
		return segments
	}

	// Extract the formatted content
	formattedContent := remaining[:closePos]

	// Determine the style
	style := TextStyle{Bold: false, Italic: false}
	switch marker {
	case "***":
		style.Bold = true
		style.Italic = true
	case "**":
		style.Bold = true
	case "*":
		style.Italic = true
	}

	segments = append(segments, FormattedText{
		Content: formattedContent,
		Style:   style,
	})

	// Process the rest of the text
	afterMarker := remaining[closePos+markerLength:]
	if afterMarker != "" {
		segments = append(segments, tp.parseFormattingRecursive(afterMarker)...)
	}

	return segments
}

// SeparateFooter separates footer content from main body content
func (tp *TextProcessor) SeparateFooter(content string) (body string, footer string) {
	lines := strings.Split(content, "\n")
	footerStartIndex := -1

	// Look for "## Footer" header (case insensitive)
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if trimmed == "## footer" {
			footerStartIndex = i
			break
		}
	}

	if footerStartIndex == -1 {
		// No footer found, return original content as body
		return content, ""
	}

	// Split the content
	bodyLines := lines[:footerStartIndex]
	footerLines := lines[footerStartIndex+1:] // Skip the "## Footer" line itself

	// Clean up body (remove trailing empty lines)
	for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
		bodyLines = bodyLines[:len(bodyLines)-1]
	}

	// Clean up footer (remove leading empty lines)
	for len(footerLines) > 0 && strings.TrimSpace(footerLines[0]) == "" {
		footerLines = footerLines[1:]
	}

	body = strings.Join(bodyLines, "\n")
	footer = strings.Join(footerLines, "\n")

	return body, footer
}

// StripMarkdownHeaders removes markdown headers from content
func (tp *TextProcessor) StripMarkdownHeaders(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// DrawFormattedText renders formatted markdown text with proper styling
func (tp *TextProcessor) DrawFormattedText(dc *gg.Context, lines []FormattedLine, x, y, w, h float64, align string, baseFont *templates.Font, vars map[string]string) {
	if len(lines) == 0 {
		return
	}

	// Get base font size
	baseSize := 12.0
	if baseFont.Size != nil {
		switch s := baseFont.Size.(type) {
		case int:
			baseSize = float64(s)
		case float64:
			baseSize = s
		case string:
			resolved := tp.utils.SubstituteVariables(s, vars)
			if parsed, err := strconv.ParseFloat(resolved, 64); err == nil {
				baseSize = parsed
			}
		}
	}

	// Get base color
	var baseColor color.Color = color.Black
	if baseFont.Color != "" {
		colorStr := tp.utils.SubstituteVariables(baseFont.Color, vars)
		if c, err := tp.utils.ParseColor(colorStr); err == nil {
			baseColor = c
		}
	}

	// Calculate line heights and total height
	currentY := y
	lineHeight := baseSize * 1.2

	// First pass: calculate total text height for centering
	totalHeight := 0.0
	for _, line := range lines {
		switch line.Type {
		case "header":
			// Headers are larger
			headerSize := baseSize * (2.0 - float64(line.Level)*0.2) // h1=1.8x, h2=1.6x, etc.
			totalHeight += headerSize * 1.4
		case "hr":
			totalHeight += baseSize * 0.5 // Horizontal rule takes less space
		case "normal":
			if len(line.Segments) == 0 {
				totalHeight += lineHeight * 0.5 // Empty line
			} else {
				totalHeight += lineHeight
			}
		}
	}

	// Center the text block vertically
	startY := y + (h-totalHeight)/2

	// Second pass: render the text
	currentY = startY
	for _, line := range lines {
		switch line.Type {
		case "header":
			// Render header with larger font
			headerSize := baseSize * (2.0 - float64(line.Level)*0.2)
			tp.setFont(dc, headerSize, true, false, baseColor)

			// Render header segments
			lineText := tp.combineSegments(line.Segments)
			tp.drawSingleLine(dc, lineText, x, currentY, w, align)
			currentY += headerSize * 1.4

		case "hr":
			// Draw horizontal rule
			dc.SetColor(color.RGBA{128, 128, 128, 255})
			dc.SetLineWidth(1)
			ruleY := currentY + baseSize*0.25
			dc.DrawLine(x+w*0.1, ruleY, x+w*0.9, ruleY)
			dc.Stroke()
			currentY += baseSize * 0.5

		case "normal":
			if len(line.Segments) == 0 {
				// Empty line - just add spacing
				currentY += lineHeight * 0.5
			} else {
				// Render formatted segments in this line
				currentY = tp.drawFormattedLine(dc, line.Segments, x, currentY, w, baseSize, baseColor, align)
			}
		}
	}
}

// drawFormattedLine renders a single line with multiple formatted segments, with word wrapping
func (tp *TextProcessor) drawFormattedLine(dc *gg.Context, segments []FormattedText, x, y, w, baseSize float64, baseColor color.Color, align string) float64 {
	if len(segments) == 0 {
		return y + baseSize*1.2
	}

	// Convert segments into wrapped lines with formatting preserved
	wrappedLines := tp.wrapFormattedSegments(dc, segments, w, baseSize, baseColor)

	// Render each wrapped line
	currentY := y

	for _, line := range wrappedLines {
		currentY = tp.renderWrappedFormattedLine(dc, line, x, currentY, w, baseSize, baseColor, align)
	}

	return currentY
}

// wrapFormattedSegments wraps formatted text segments across multiple lines
func (tp *TextProcessor) wrapFormattedSegments(dc *gg.Context, segments []FormattedText, maxWidth float64, baseSize float64, baseColor color.Color) [][]FormattedText {
	var wrappedLines [][]FormattedText
	var currentLine []FormattedText
	currentLineWidth := 0.0

	for _, segment := range segments {
		// Set font for this segment to measure accurately
		tp.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)

		// Split segment into words
		words := strings.Fields(segment.Content)
		if len(words) == 0 {
			continue
		}

		for i, word := range words {
			// Add space before word (except for first word in segment)
			testWord := word
			if i > 0 {
				testWord = " " + word
			}

			wordWidth, _ := dc.MeasureString(testWord)

			// Check if adding this word would exceed the line width
			if currentLineWidth+wordWidth > maxWidth && len(currentLine) > 0 {
				// Start a new line
				wrappedLines = append(wrappedLines, currentLine)
				currentLine = []FormattedText{}
				currentLineWidth = 0.0

				// Add the word to the new line (without leading space)
				wordWidth, _ = dc.MeasureString(word)
				currentLine = append(currentLine, FormattedText{
					Content: word,
					Style:   segment.Style,
				})
				currentLineWidth = wordWidth
			} else {
				// Add word to current line
				if i == 0 && len(currentLine) == 0 {
					// First word in first segment on line
					currentLine = append(currentLine, FormattedText{
						Content: word,
						Style:   segment.Style,
					})
				} else {
					// Add word with space prefix if needed
					content := word
					if i > 0 || len(currentLine) > 0 {
						content = " " + word
					}
					currentLine = append(currentLine, FormattedText{
						Content: content,
						Style:   segment.Style,
					})
				}
				currentLineWidth += wordWidth
			}
		}
	}

	// Add the last line if it has content
	if len(currentLine) > 0 {
		wrappedLines = append(wrappedLines, currentLine)
	}

	return wrappedLines
}

// renderWrappedFormattedLine renders a single wrapped line with formatted segments
func (tp *TextProcessor) renderWrappedFormattedLine(dc *gg.Context, segments []FormattedText, x, y, w, baseSize float64, baseColor color.Color, align string) float64 {
	// Check if this is an empty line (paragraph break)
	if len(segments) == 0 {
		return y + baseSize*1.8 // Extra spacing for paragraph breaks
	}

	// Calculate total width of the line for alignment
	totalWidth := 0.0
	for _, segment := range segments {
		tp.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)
		segmentWidth, _ := dc.MeasureString(segment.Content)
		totalWidth += segmentWidth
	}

	// Calculate starting X position based on alignment
	currentX := x
	switch align {
	case "center":
		currentX = x + (w-totalWidth)/2
	case "right":
		currentX = x + w - totalWidth
	}

	// Render each segment with its own formatting
	for _, segment := range segments {
		tp.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)

		// Draw the segment
		dc.DrawStringAnchored(segment.Content, currentX, y, 0.0, 0.0)

		// Move X position forward by the width of this segment
		segmentWidth, _ := dc.MeasureString(segment.Content)
		currentX += segmentWidth
	}

	return y + baseSize*1.5 // Increased line spacing for better readability
}

// combineSegments combines formatted segments into plain text
func (tp *TextProcessor) combineSegments(segments []FormattedText) string {
	var result strings.Builder
	for _, segment := range segments {
		result.WriteString(segment.Content)
	}
	return result.String()
}

// drawSingleLine draws a single line of text with alignment
func (tp *TextProcessor) drawSingleLine(dc *gg.Context, text string, x, y, w float64, align string) {
	switch align {
	case "right":
		dc.DrawStringAnchored(text, x+w, y, 1.0, 0.0)
	case "center":
		dc.DrawStringAnchored(text, x+w/2, y, 0.5, 0.0)
	default: // left
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0)
	}
}

// setFont sets up font with the specified properties
func (tp *TextProcessor) setFont(dc *gg.Context, size float64, bold, italic bool, textColor color.Color) {
	var fontData []byte

	// Choose the appropriate font based on style
	if bold && italic {
		// For bold+italic, use bold font (closest we have)
		fontData = gobold.TTF
	} else if bold {
		fontData = gobold.TTF
	} else if italic {
		fontData = goitalic.TTF
	} else {
		fontData = goregular.TTF
	}

	f, err := truetype.Parse(fontData)
	if err != nil {
		// Fallback to regular font
		f, _ = truetype.Parse(goregular.TTF)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
		DPI:  72,
	})

	dc.SetFontFace(face)
	dc.SetColor(textColor)
}

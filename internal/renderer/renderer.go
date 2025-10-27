package renderer

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
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

// Renderer handles image generation from templates and card data
type Renderer struct {
	imageCache map[string]image.Image
}

// NewRenderer creates a new renderer instance
func NewRenderer() *Renderer {
	return &Renderer{
		imageCache: make(map[string]image.Image),
	}
}

// RenderCard generates a PNG image from a card and template
func (r *Renderer) RenderCard(card *metadata.Card, template *templates.Template, outputPath string) error {
	// Create drawing context
	dc := gg.NewContext(template.Dimensions.Width, template.Dimensions.Height)

	// Set background to white
	dc.SetColor(color.White)
	dc.Clear()

	// Process template variables for this card
	templateVars := r.buildTemplateVariables(card, template)

	// Render each layer in order
	for _, layer := range template.Layers {
		if err := r.renderLayer(dc, layer, templateVars, template); err != nil {
			return fmt.Errorf("error rendering layer '%s': %v", layer.Name, err)
		}
	}

	// Save the image
	if err := dc.SavePNG(outputPath); err != nil {
		return fmt.Errorf("error saving image to %s: %v", outputPath, err)
	}

	return nil
}

// buildTemplateVariables creates a map of all template variables for this card
func (r *Renderer) buildTemplateVariables(card *metadata.Card, template *templates.Template) map[string]string {
	vars := make(map[string]string)

	// Use parsed rules text for body, fall back to full body if needed
	body := card.RulesText
	if body == "" {
		body = card.Body
	}

	// Separate footer from body (in case it wasn't parsed separately)
	bodyContent, footer := r.separateFooter(body)

	// Use parsed flavor text for footer if available
	if card.FlavorText != "" && footer == "" {
		footer = card.FlavorText
	}

	// Basic card fields
	vars["card.title"] = card.Title
	vars["card.type"] = card.Type
	vars["card.rarity"] = card.Rarity
	vars["card.set"] = card.Set
	vars["card.artist"] = card.Artist
	vars["card.body"] = bodyContent
	vars["card.footer"] = footer
	vars["card.rules_text"] = card.RulesText
	vars["card.flavor_text"] = card.FlavorText
	vars["card.mana_cost"] = card.ManaCost
	vars["card.print_this"] = strconv.Itoa(card.PrintThis)
	vars["card.print_total"] = strconv.Itoa(card.PrintTotal)

	// Add all metadata fields
	for key, value := range card.Metadata {
		if str, ok := value.(string); ok {
			vars[key] = str
		} else if num, ok := value.(int); ok {
			vars[key] = strconv.Itoa(num)
		} else if fl, ok := value.(float64); ok {
			vars[key] = strconv.FormatFloat(fl, 'f', -1, 64)
		} else if value != nil {
			vars[key] = fmt.Sprintf("%v", value)
		}
	}

	// Add style tokens
	for key, value := range template.StyleTokens {
		vars["style_tokens."+key] = value
	}

	// Add template optional fields (includes font sizes and other defaults)
	for key, value := range template.Optional {
		if str, ok := value.(string); ok {
			vars[key] = str
		} else if num, ok := value.(int); ok {
			vars[key] = strconv.Itoa(num)
		} else if fl, ok := value.(float64); ok {
			vars[key] = strconv.FormatFloat(fl, 'f', -1, 64)
		} else if value != nil {
			vars[key] = fmt.Sprintf("%v", value)
		}
	}

	// Add template directory
	vars["template_dir"] = template.TemplateDir
	vars["icon_dir"] = filepath.Join(template.TemplateDir, "icons")

	return vars
}

// renderLayer renders a single layer
func (r *Renderer) renderLayer(dc *gg.Context, layer templates.Layer, vars map[string]string, template *templates.Template) error {
	// Check condition if present
	if layer.Condition != "" {
		if !r.evaluateCondition(layer.Condition, vars) {
			return nil // Skip this layer
		}
	}

	switch layer.Type {
	case "image":
		return r.renderImageLayer(dc, layer, vars)
	case "text":
		return r.renderTextLayer(dc, layer, vars, template)
	default:
		return fmt.Errorf("unknown layer type: %s", layer.Type)
	}
}

// renderImageLayer renders an image layer
func (r *Renderer) renderImageLayer(dc *gg.Context, layer templates.Layer, vars map[string]string) error {
	// Resolve image source
	imagePath := r.substituteVariables(layer.Source, vars)
	if imagePath == "" {
		// Try fallback
		if layer.Fallback != "" {
			imagePath = r.substituteVariables(layer.Fallback, vars)
		}
		if imagePath == "" {
			return fmt.Errorf("no image source for layer %s", layer.Name)
		}
	}

	// Load image (with caching)
	img, err := r.loadImage(imagePath)
	if err != nil {
		// Try fallback if main source fails
		if layer.Fallback != "" && imagePath != r.substituteVariables(layer.Fallback, vars) {
			fallbackPath := r.substituteVariables(layer.Fallback, vars)
			img, err = r.loadImage(fallbackPath)
		}
		if err != nil {
			// Create a placeholder rectangle instead of failing
			r.renderPlaceholder(dc, layer, fmt.Sprintf("Missing: %s", filepath.Base(imagePath)))
			return nil
		}
	}

	// Draw image in the specified region
	dc.DrawImageAnchored(img, layer.Region.X+layer.Region.Width/2, layer.Region.Y+layer.Region.Height/2, 0.5, 0.5)

	return nil
}

// renderTextLayer renders a text layer
func (r *Renderer) renderTextLayer(dc *gg.Context, layer templates.Layer, vars map[string]string, template *templates.Template) error {
	// Get text content
	content := r.substituteVariables(layer.Content, vars)
	if content == "" {
		return nil // Skip empty content
	}

	// Strip headers if enabled
	if layer.StripHeaders {
		content = r.stripMarkdownHeaders(content)
	}

	// Process icon replacements if enabled (after variable substitution)
	if layer.IconReplace {
		content = r.processIconReplacements(content, template, vars)
	}

	// Process markdown formatting
	formattedLines := r.processMarkdown(content)

	// Set up base font
	baseFont := &templates.Font{Size: 12.0, Color: "#000000"}
	if layer.Font != nil {
		baseFont = layer.Font
	}

	// Calculate text position
	x := float64(layer.Region.X)
	y := float64(layer.Region.Y)
	w := float64(layer.Region.Width)
	h := float64(layer.Region.Height)

	// Render formatted text
	r.drawFormattedText(dc, formattedLines, x, y, w, h, layer.Align, baseFont, vars)

	return nil
}

// loadImage loads an image with caching
func (r *Renderer) loadImage(path string) (image.Image, error) {
	// Check cache first
	if img, exists := r.imageCache[path]; exists {
		return img, nil
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file not found: %s", path)
	}

	// Load image
	img, err := gg.LoadImage(path)
	if err != nil {
		return nil, err
	}

	// Cache it
	r.imageCache[path] = img
	return img, nil
}

// renderPlaceholder renders a placeholder rectangle with text
func (r *Renderer) renderPlaceholder(dc *gg.Context, layer templates.Layer, text string) {
	// Draw placeholder rectangle
	dc.SetColor(color.RGBA{200, 200, 200, 255})
	dc.DrawRectangle(float64(layer.Region.X), float64(layer.Region.Y),
		float64(layer.Region.Width), float64(layer.Region.Height))
	dc.Fill()

	// Draw border
	dc.SetColor(color.RGBA{100, 100, 100, 255})
	dc.SetLineWidth(2)
	dc.DrawRectangle(float64(layer.Region.X), float64(layer.Region.Y),
		float64(layer.Region.Width), float64(layer.Region.Height))
	dc.Stroke()

	// Draw text
	dc.SetColor(color.RGBA{50, 50, 50, 255})
	dc.DrawStringAnchored(text,
		float64(layer.Region.X+layer.Region.Width/2),
		float64(layer.Region.Y+layer.Region.Height/2),
		0.5, 0.5)
}

// substituteVariables replaces {{variable}} patterns with actual values
func (r *Renderer) substituteVariables(template string, vars map[string]string) string {
	result := template

	// Simple variable substitution for now
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// processIconReplacements handles icon replacement in text
func (r *Renderer) processIconReplacements(content string, template *templates.Template, vars map[string]string) string {
	result := content

	// Look for icon patterns and replace with text placeholders
	// TODO: Implement actual icon rendering
	for iconKey := range template.Icons {
		placeholder := "{{" + iconKey + "}}"
		replacement := "[" + iconKey + "]" // Text placeholder for now
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

// processMarkdown parses markdown content into formatted lines
func (r *Renderer) processMarkdown(content string) []FormattedLine {
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
					Segments: r.parseInlineFormatting(line),
					Type:     "header",
					Level:    level,
				})
				continue
			}
		}

		// Regular line with inline formatting
		formattedLines = append(formattedLines, FormattedLine{
			Segments: r.parseInlineFormatting(line),
			Type:     "normal",
		})
	}

	return formattedLines
}

// parseInlineFormatting parses inline markdown formatting like **bold** and *italic*
func (r *Renderer) parseInlineFormatting(text string) []FormattedText {
	// Process the text to handle nested and overlapping formats
	return r.parseFormattingRecursive(text)
}

// parseFormattingRecursive handles nested and overlapping markdown formatting
func (r *Renderer) parseFormattingRecursive(text string) []FormattedText {
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
		segments = append(segments, r.parseFormattingRecursive(afterMarker)...)
	}

	return segments
}

// separateFooter separates footer content from main body content
func (r *Renderer) separateFooter(content string) (body string, footer string) {
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

// stripMarkdownHeaders removes markdown headers from content
func (r *Renderer) stripMarkdownHeaders(content string) string {
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

// evaluateCondition evaluates a simple condition
func (r *Renderer) evaluateCondition(condition string, vars map[string]string) bool {
	// Simple condition evaluation - check if variables exist and are non-empty
	condition = strings.TrimSpace(condition)

	// Remove {{ }} brackets
	condition = strings.ReplaceAll(condition, "{{", "")
	condition = strings.ReplaceAll(condition, "}}", "")

	// Split on && (simple AND logic)
	parts := strings.Split(condition, "&&")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if value, exists := vars[part]; !exists || value == "" || value == "null" {
			return false
		}
	}

	return true
}

// parseColor parses a color string (hex format)
func (r *Renderer) parseColor(colorStr string) (color.Color, error) {
	if !strings.HasPrefix(colorStr, "#") {
		return color.Black, fmt.Errorf("invalid color format: %s", colorStr)
	}

	colorStr = strings.TrimPrefix(colorStr, "#")

	if len(colorStr) == 6 {
		// RGB format
		r, _ := strconv.ParseUint(colorStr[0:2], 16, 8)
		g, _ := strconv.ParseUint(colorStr[2:4], 16, 8)
		b, _ := strconv.ParseUint(colorStr[4:6], 16, 8)
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
	}

	return color.Black, fmt.Errorf("unsupported color format: %s", colorStr)
}

// getCurrentFontSize extracts the current font size from the drawing context
// drawFormattedText renders formatted markdown text with proper styling
func (r *Renderer) drawFormattedText(dc *gg.Context, lines []FormattedLine, x, y, w, h float64, align string, baseFont *templates.Font, vars map[string]string) {
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
			resolved := r.substituteVariables(s, vars)
			if parsed, err := strconv.ParseFloat(resolved, 64); err == nil {
				baseSize = parsed
			}
		}
	}

	// Get base color
	var baseColor color.Color = color.Black
	if baseFont.Color != "" {
		colorStr := r.substituteVariables(baseFont.Color, vars)
		if c, err := r.parseColor(colorStr); err == nil {
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
			r.setFont(dc, headerSize, true, false, baseColor)

			// Render header segments
			lineText := r.combineSegments(line.Segments)
			r.drawSingleLine(dc, lineText, x, currentY, w, align)
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
				currentY = r.drawFormattedLine(dc, line.Segments, x, currentY, w, baseSize, baseColor, align)
			}
		}
	}
}

// drawFormattedLine renders a single line with multiple formatted segments, with word wrapping
func (r *Renderer) drawFormattedLine(dc *gg.Context, segments []FormattedText, x, y, w, baseSize float64, baseColor color.Color, align string) float64 {
	if len(segments) == 0 {
		return y + baseSize*1.2
	}

	// Convert segments into wrapped lines with formatting preserved
	wrappedLines := r.wrapFormattedSegments(dc, segments, w, baseSize, baseColor)

	// Render each wrapped line
	currentY := y

	for _, line := range wrappedLines {
		currentY = r.renderWrappedFormattedLine(dc, line, x, currentY, w, baseSize, baseColor, align)
	}

	return currentY
}

// wrapFormattedSegments wraps formatted text segments across multiple lines
func (r *Renderer) wrapFormattedSegments(dc *gg.Context, segments []FormattedText, maxWidth float64, baseSize float64, baseColor color.Color) [][]FormattedText {
	var wrappedLines [][]FormattedText
	var currentLine []FormattedText
	currentLineWidth := 0.0

	for _, segment := range segments {
		// Set font for this segment to measure accurately
		r.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)

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
func (r *Renderer) renderWrappedFormattedLine(dc *gg.Context, segments []FormattedText, x, y, w, baseSize float64, baseColor color.Color, align string) float64 {
	// Check if this is an empty line (paragraph break)
	if len(segments) == 0 {
		return y + baseSize*1.8 // Extra spacing for paragraph breaks
	}

	// Calculate total width of the line for alignment
	totalWidth := 0.0
	for _, segment := range segments {
		r.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)
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
		r.setFont(dc, baseSize, segment.Style.Bold, segment.Style.Italic, baseColor)

		// Draw the segment
		dc.DrawStringAnchored(segment.Content, currentX, y, 0.0, 0.0)

		// Move X position forward by the width of this segment
		segmentWidth, _ := dc.MeasureString(segment.Content)
		currentX += segmentWidth
	}

	return y + baseSize*1.5 // Increased line spacing for better readability
}

// combineSegments combines formatted segments into plain text
func (r *Renderer) combineSegments(segments []FormattedText) string {
	var result strings.Builder
	for _, segment := range segments {
		result.WriteString(segment.Content)
	}
	return result.String()
}

// drawSingleLine draws a single line of text with alignment
func (r *Renderer) drawSingleLine(dc *gg.Context, text string, x, y, w float64, align string) {
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
func (r *Renderer) setFont(dc *gg.Context, size float64, bold, italic bool, textColor color.Color) {
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

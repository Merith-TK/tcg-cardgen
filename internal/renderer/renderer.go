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
	"golang.org/x/image/font/gofont/goregular"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
	"github.com/Merith-TK/tcg-cardgen/internal/templates"
)

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

	// Basic card fields
	vars["card.title"] = card.Title
	vars["card.type"] = card.Type
	vars["card.rarity"] = card.Rarity
	vars["card.set"] = card.Set
	vars["card.artist"] = card.Artist
	vars["card.body"] = card.Body
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
	content = r.processMarkdown(content)

	// Set up font
	if layer.Font != nil {
		if err := r.setupFont(dc, layer.Font, vars); err != nil {
			return fmt.Errorf("error setting up font: %v", err)
		}
	}

	// Calculate text position
	x := float64(layer.Region.X)
	y := float64(layer.Region.Y)
	w := float64(layer.Region.Width)
	h := float64(layer.Region.Height)

	// Handle multi-line text properly
	r.drawMultilineText(dc, content, x, y, w, h, layer.Align)

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

// setupFont configures the drawing context font
func (r *Renderer) setupFont(dc *gg.Context, font *templates.Font, vars map[string]string) error {
	// Get font size
	size := 12.0
	if font.Size != nil {
		switch s := font.Size.(type) {
		case int:
			size = float64(s)
		case float64:
			size = s
		case string:
			resolved := r.substituteVariables(s, vars)
			if parsed, err := strconv.ParseFloat(resolved, 64); err == nil {
				size = parsed
			}
		}
	}

	// Debug: print the font size being used
	fmt.Printf("DEBUG: Setting font size to %.1f for layer\n", size)

	// Create a proper font face using Go's font system
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
		DPI:  72, // Match the card DPI for proper scaling
	})

	dc.SetFontFace(face)

	// Set color
	if font.Color != "" {
		colorStr := r.substituteVariables(font.Color, vars)
		if c, err := r.parseColor(colorStr); err == nil {
			dc.SetColor(c)
		} else {
			dc.SetColor(color.Black) // Fallback
		}
	} else {
		dc.SetColor(color.Black) // Default color
	}

	return nil
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

// processMarkdown handles basic markdown formatting
func (r *Renderer) processMarkdown(content string) string {
	// Handle basic markdown formatting - for now just clean it up
	result := content

	// Remove markdown syntax for now - just clean it up
	// **bold** -> bold (remove asterisks)
	result = strings.ReplaceAll(result, "**", "")

	// *italic* -> italic (remove single asterisks)
	result = strings.ReplaceAll(result, "*", "")

	// Clean up extra whitespace but preserve line breaks
	lines := strings.Split(result, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Trim whitespace but keep the line
		cleaned := strings.TrimSpace(line)
		cleanLines = append(cleanLines, cleaned)
	}

	// Join with line breaks and clean up multiple empty lines
	result = strings.Join(cleanLines, "\n")

	// Replace multiple consecutive newlines with double newlines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return result
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

// drawMultilineText draws text with proper line break handling
func (r *Renderer) drawMultilineText(dc *gg.Context, content string, x, y, w, h float64, align string) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return
	}

	// Calculate line height based on font metrics
	_, lineHeight := dc.MeasureString("Tg") // Use characters with ascenders and descenders
	lineHeight *= 1.2                       // Add some line spacing

	// Calculate starting Y position to center the text block vertically
	totalTextHeight := float64(len(lines)) * lineHeight
	startY := y + (h-totalTextHeight)/2 + lineHeight*0.8 // Adjust for baseline

	// Draw each line
	for i, line := range lines {
		lineY := startY + float64(i)*lineHeight

		// Skip empty lines but still advance the position
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Adjust X position based on alignment
		switch align {
		case "right":
			dc.DrawStringAnchored(line, x+w, lineY, 1.0, 0.0)
		case "center":
			dc.DrawStringAnchored(line, x+w/2, lineY, 0.5, 0.0)
		default: // left
			dc.DrawStringAnchored(line, x, lineY, 0.0, 0.0)
		}
	}
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

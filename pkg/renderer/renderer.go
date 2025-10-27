package renderer

import (
	"fmt"
	"image/color"
	"path/filepath"

	"github.com/fogleman/gg"

	"github.com/Merith-TK/tcg-cardgen/pkg/metadata"
	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
)

// Renderer handles image generation from templates and card data
type Renderer struct {
	imageProcessor    *ImageProcessor
	textProcessor     *TextProcessor
	variableProcessor *VariableProcessor
	utils             *Utils
}

// NewRenderer creates a new renderer instance
func NewRenderer() *Renderer {
	return &Renderer{
		imageProcessor:    NewImageProcessor(),
		textProcessor:     NewTextProcessor(),
		variableProcessor: NewVariableProcessor(),
		utils:             NewUtils(),
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
	templateVars := r.variableProcessor.BuildTemplateVariables(card, template)

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

// renderLayer renders a single layer
func (r *Renderer) renderLayer(dc *gg.Context, layer templates.Layer, vars map[string]string, template *templates.Template) error {
	// Check condition if present
	if layer.Condition != "" {
		if !r.utils.EvaluateCondition(layer.Condition, vars) {
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
	imagePath := r.variableProcessor.SubstituteVariables(layer.Source, vars)

	if imagePath == "" {
		// Try fallback
		if layer.Fallback != "" {
			imagePath = r.variableProcessor.SubstituteVariables(layer.Fallback, vars)
		}
		if imagePath == "" {
			return fmt.Errorf("no image source for layer %s", layer.Name)
		}
	}

	// Load image (with caching)
	img, err := r.imageProcessor.LoadImage(imagePath)
	if err != nil {
		// Try fallback if main source fails
		if layer.Fallback != "" && imagePath != r.variableProcessor.SubstituteVariables(layer.Fallback, vars) {
			fallbackPath := r.variableProcessor.SubstituteVariables(layer.Fallback, vars)
			img, err = r.imageProcessor.LoadImage(fallbackPath)
		}
		if err != nil {
			// Create a placeholder rectangle instead of failing
			r.imageProcessor.RenderPlaceholder(dc, layer, fmt.Sprintf("Missing: %s", filepath.Base(imagePath)))
			return nil
		}
	}

	// Draw image fitted to the specified region
	// Priority: card.artwork.fit > template fit_mode > "fill" default
	fitMode := layer.FitMode
	if cardFitMode, exists := vars["card.artwork.fit"]; exists && cardFitMode != "" {
		fitMode = cardFitMode // Card-specific override
	}
	if fitMode == "" {
		fitMode = "fill" // Final default
	}
	fittedImg := r.imageProcessor.CreateFittedImage(img, layer.Region, fitMode)
	dc.DrawImageAnchored(fittedImg, layer.Region.X+layer.Region.Width/2, layer.Region.Y+layer.Region.Height/2, 0.5, 0.5)

	return nil
}

// renderTextLayer renders a text layer
func (r *Renderer) renderTextLayer(dc *gg.Context, layer templates.Layer, vars map[string]string, template *templates.Template) error {
	// Get text content
	content := r.variableProcessor.SubstituteVariables(layer.Content, vars)
	if content == "" {
		return nil // Skip empty content
	}

	// Strip headers if enabled
	if layer.StripHeaders {
		content = r.textProcessor.StripMarkdownHeaders(content)
	}

	// Process icon replacements if enabled (after variable substitution)
	if layer.IconReplace {
		content = r.variableProcessor.ProcessIconReplacements(content, template, vars)
	}

	// Process markdown formatting
	formattedLines := r.textProcessor.ProcessMarkdown(content)

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
	r.textProcessor.DrawFormattedText(dc, formattedLines, x, y, w, h, layer.Align, baseFont, vars)

	return nil
}

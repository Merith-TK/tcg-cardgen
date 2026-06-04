// Package renderer handles image generation from card data and templates.
package renderer

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"

	"github.com/Merith-TK/tcg-cardgen/pkg/card"
	"github.com/Merith-TK/tcg-cardgen/pkg/expr"
	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
)

// RenderCard generates a PNG from a card and template, writing output to outputPath.
func RenderCard(c *card.Card, t *templates.Template, outputPath string) error {
	dc := gg.NewContext(t.Dimensions.Width, t.Dimensions.Height)
	dc.SetColor(color.White)
	dc.Clear()

	vars := buildVars(c, t)
	proc := expr.New(vars)

	for _, layer := range t.Layers {
		if err := renderLayer(dc, layer, proc, vars, t); err != nil {
			return fmt.Errorf("layer %q: %w", layer.Name, err)
		}
	}

	if err := dc.SavePNG(outputPath); err != nil {
		return fmt.Errorf("save %s: %w", outputPath, err)
	}
	return nil
}

// renderLayer dispatches a single layer to the appropriate renderer.
func renderLayer(dc *gg.Context, layer types.Layer, proc *expr.Processor, vars map[string]string, t *templates.Template) error {
	// Evaluate condition — skip layer if falsy.
	if layer.Condition != "" {
		if !proc.EvalBool(layer.Condition) {
			return nil
		}
	}

	switch layer.Type {
	case "image":
		return renderImageLayer(dc, layer, proc, vars, t)
	case "text":
		return renderTextLayer(dc, layer, proc, vars, t)
	case "shape":
		return renderShapeLayer(dc, layer, proc, vars)
	default:
		return fmt.Errorf("unknown layer type %q", layer.Type)
	}
}

// renderImageLayer renders an image layer.
func renderImageLayer(dc *gg.Context, layer types.Layer, proc *expr.Processor, vars map[string]string, t *templates.Template) error {
	imagePath := proc.Substitute(layer.Source)

	if imagePath == "" && layer.Fallback != "" {
		imagePath = proc.Substitute(layer.Fallback)
	}
	if imagePath == "" {
		renderPlaceholder(dc, layer, "no source")
		return nil
	}

	img, err := loadImage(imagePath)
	if err != nil {
		if layer.Fallback != "" {
			fbPath := proc.Substitute(layer.Fallback)
			img, err = loadImage(fbPath)
		}
		if err != nil {
			renderPlaceholder(dc, layer, fmt.Sprintf("missing: %s", filepath.Base(imagePath)))
			return nil
		}
	}

	fitMode := layer.FitMode
	if m, ok := vars["card.artwork.fit"]; ok && m != "" {
		fitMode = m
	}
	if fitMode == "" {
		fitMode = "fill"
	}

	fitted := createFittedImage(img, layer.Region, fitMode)
	cx := layer.Region.X + layer.Region.Width/2
	cy := layer.Region.Y + layer.Region.Height/2
	dc.DrawImageAnchored(fitted, cx, cy, 0.5, 0.5)
	return nil
}

// renderTextLayer renders a text layer.
func renderTextLayer(dc *gg.Context, layer types.Layer, proc *expr.Processor, vars map[string]string, t *templates.Template) error {
	content := proc.Substitute(layer.Content)
	if content == "" {
		return nil
	}

	if layer.StripHeaders {
		content = stripMarkdownHeaders(content)
	}

	if layer.IconReplace {
		content = processIconReplacements(content, t, proc)
	}

	lines := parseMarkdown(content)

	baseFont := &types.Font{Size: 12.0, Color: "#000000"}
	if layer.Font != nil {
		baseFont = layer.Font
	}

	// Resolve font size — may be a variable expression.
	baseSize := 12.0
	switch s := baseFont.Size.(type) {
	case int:
		baseSize = float64(s)
	case float64:
		baseSize = s
	case string:
		resolved := proc.Substitute(s)
		if _, err := fmt.Sscanf(resolved, "%f", &baseSize); err != nil {
			baseSize = 12.0
		}
	}

	// Resolve font color.
	baseColor := color.Color(color.Black)
	if baseFont.Color != "" {
		colorStr := proc.Substitute(baseFont.Color)
		if c, err := parseColor(colorStr); err == nil {
			baseColor = c
		}
	}

	fontFamily := ""
	if baseFont.Family != "" {
		fontFamily = proc.Substitute(baseFont.Family)
	}

	templateDir := vars["template_dir"]

	x := float64(layer.Region.X)
	y := float64(layer.Region.Y)
	w := float64(layer.Region.Width)
	h := float64(layer.Region.Height)

	drawFormattedText(dc, lines, x, y, w, h, layer.Align, baseSize, baseColor, fontFamily, templateDir)
	return nil
}

// processIconReplacements substitutes icon placeholders in text.
func processIconReplacements(content string, t *templates.Template, proc *expr.Processor) string {
	result := content
	for iconKey := range t.Icons {
		placeholder := "{{" + iconKey + "}}"
		result = strings.ReplaceAll(result, placeholder, "["+iconKey+"]")
	}
	return result
}

// buildVars constructs the full variable map for a card/template pair.
func buildVars(c *card.Card, t *templates.Template) map[string]string {
	vars := make(map[string]string, 64)

	// Core card fields.
	vars["card.title"] = c.Title
	vars["card.type"] = c.Type
	vars["card.rarity"] = c.Rarity
	vars["card.set"] = c.Set
	vars["card.artist"] = c.Artist
	vars["card.rules_text"] = c.RulesText
	vars["card.flavor_text"] = c.FlavorText
	vars["card.mana_cost"] = c.ManaCost
	vars["card.body"] = c.RulesText
	vars["card.footer"] = c.FlavorText
	vars["card.print_this"] = fmt.Sprintf("%d", c.PrintThis)
	vars["card.print_total"] = fmt.Sprintf("%d", c.PrintTotal)
	vars["card.tcg"] = c.TCG
	vars["card.cardstyle"] = c.CardStyle

	// Flatten metadata map recursively (one level of nesting).
	flattenMeta(vars, c.Metadata, "")

	// Style tokens (accessible as style_tokens.key).
	for k, v := range t.StyleTokens {
		vars["style_tokens."+k] = v
	}

	// Optional field defaults.
	for k, v := range t.Optional {
		vars[k] = fmt.Sprintf("%v", v)
	}

	// Template directory paths.
	vars["template_dir"] = t.TemplateDir
	vars["icon_dir"] = filepath.Join(t.TemplateDir, "icons")

	return vars
}

// flattenMeta flattens a nested metadata map into dot-path keys.
func flattenMeta(vars map[string]string, m map[string]interface{}, prefix string) {
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			flattenMeta(vars, val, fullKey)
		case string:
			vars[fullKey] = val
		case int:
			vars[fullKey] = fmt.Sprintf("%d", val)
		case float64:
			vars[fullKey] = fmt.Sprintf("%g", val)
		default:
			if val != nil {
				vars[fullKey] = fmt.Sprintf("%v", val)
			}
		}
	}
}

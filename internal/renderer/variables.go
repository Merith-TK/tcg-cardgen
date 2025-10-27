package renderer

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
	"github.com/Merith-TK/tcg-cardgen/internal/templates"
)

// VariableProcessor handles template variable building and substitution
type VariableProcessor struct {
	textProcessor *TextProcessor
}

// NewVariableProcessor creates a new variable processor
func NewVariableProcessor() *VariableProcessor {
	return &VariableProcessor{
		textProcessor: NewTextProcessor(),
	}
}

// BuildTemplateVariables creates a map of all template variables for this card
func (vp *VariableProcessor) BuildTemplateVariables(card *metadata.Card, template *templates.Template) map[string]string {
	vars := make(map[string]string)

	// Use parsed rules text for body, fall back to full body if needed
	body := card.RulesText
	if body == "" {
		body = card.Body
	}

	// Separate footer from body (in case it wasn't parsed separately)
	bodyContent, footer := vp.textProcessor.SeparateFooter(body)

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

	// Add artwork from metadata if present
	// Check for card.artwork in the nested card map
	if cardMap, exists := card.Metadata["card"]; exists {
		if cardMapTyped, ok := cardMap.(map[string]interface{}); ok {
			if artwork, exists := cardMapTyped["artwork"]; exists {
				if artworkStr, ok := artwork.(string); ok {
					// Simple string format: card.artwork: "url"
					vars["card.artwork"] = artworkStr
				} else if artworkMap, ok := artwork.(map[string]interface{}); ok {
					// Nested format: card.artwork: { url: "...", fit: "..." }
					if url, exists := artworkMap["url"]; exists {
						if urlStr, ok := url.(string); ok {
							vars["card.artwork"] = urlStr // Store the URL as card.artwork
						}
					}
					if fit, exists := artworkMap["fit"]; exists {
						if fitStr, ok := fit.(string); ok {
							vars["card.artwork.fit"] = fitStr
						}
					}
				}
			}
		}
	}

	// Add all metadata fields
	for key, value := range card.Metadata {
		// Handle nested maps (like card.artwork being in card map)
		if nestedMap, ok := value.(map[string]interface{}); ok {
			for nestedKey, nestedValue := range nestedMap {
				// Skip artwork as it's handled specially above
				if key == "card" && nestedKey == "artwork" {
					continue
				}

				fullKey := key + "." + nestedKey
				if str, ok := nestedValue.(string); ok {
					vars[fullKey] = str
				} else if num, ok := nestedValue.(int); ok {
					vars[fullKey] = strconv.Itoa(num)
				} else if fl, ok := nestedValue.(float64); ok {
					vars[fullKey] = strconv.FormatFloat(fl, 'f', -1, 64)
				} else if nestedValue != nil {
					vars[fullKey] = fmt.Sprintf("%v", nestedValue)
				}
			}
		} else if str, ok := value.(string); ok {
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

// SubstituteVariables replaces {{variable}} patterns with actual values
func (vp *VariableProcessor) SubstituteVariables(template string, vars map[string]string) string {
	result := template

	// Simple variable substitution for now
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ProcessIconReplacements handles icon replacement in text
func (vp *VariableProcessor) ProcessIconReplacements(content string, template *templates.Template, vars map[string]string) string {
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

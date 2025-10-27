package renderer

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Utils provides utility functions for the renderer
type Utils struct{}

// NewUtils creates a new utils instance
func NewUtils() *Utils {
	return &Utils{}
}

// SubstituteVariables replaces {{variable}} patterns with actual values
func (u *Utils) SubstituteVariables(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ParseColor parses a color string (hex format)
func (u *Utils) ParseColor(colorStr string) (color.Color, error) {
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

// EvaluateCondition evaluates a simple condition
func (u *Utils) EvaluateCondition(condition string, vars map[string]string) bool {
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

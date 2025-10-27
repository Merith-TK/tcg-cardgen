package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
	"gopkg.in/yaml.v3"
)

// Template represents a card template definition
type Template struct {
	Name        string                 `yaml:"name"`
	TCG         string                 `yaml:"tcg"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Extends     string                 `yaml:"extends,omitempty"` // Path to base template
	Dimensions  Dimensions             `yaml:"dimensions"`
	Layers      []Layer                `yaml:"layers"`
	Required    []string               `yaml:"required_fields"`
	Optional    map[string]interface{} `yaml:"optional_fields"`
	Icons       map[string]string      `yaml:"icons"`
	StyleTokens map[string]string      `yaml:"style_tokens"`                // Visual constants
	Overrides   []LayerOverride        `yaml:"overrides,omitempty"`         // Layer modifications
	AddLayers   []Layer                `yaml:"additional_layers,omitempty"` // Extra layers
	Conditions  []Condition            `yaml:"conditions,omitempty"`        // Conditional includes

	// Runtime info
	TemplateDir  string    `yaml:"-"`
	BaseTemplate *Template `yaml:"-"` // Resolved base template
}

// LayerOverride represents modifications to existing layers
type LayerOverride struct {
	Layer   string                 `yaml:"layer"`   // Name of layer to modify
	Updates map[string]interface{} `yaml:",inline"` // Fields to update
}

// Condition represents conditional template inclusion
type Condition struct {
	If      string `yaml:"if"`      // Condition expression
	Include string `yaml:"include"` // Template file to include
}

// Dimensions defines the output image dimensions
type Dimensions struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
	DPI    int `yaml:"dpi"`
}

// Layer represents a single layer in the card template
type Layer struct {
	Name         string `yaml:"name"`
	Role         string `yaml:"role,omitempty"` // Semantic role (title, artwork, etc.)
	Type         string `yaml:"type"`           // "image", "text"
	Source       string `yaml:"source,omitempty"`
	Content      string `yaml:"content,omitempty"`
	Region       Region `yaml:"region"`
	Font         *Font  `yaml:"font,omitempty"`
	FitMode      string `yaml:"fit_mode,omitempty"` // Image fit mode: "fill", "fit", "stretch", "center"
	IconReplace  bool   `yaml:"icon_replace,omitempty"`
	StripHeaders bool   `yaml:"strip_headers,omitempty"`
	Condition    string `yaml:"condition,omitempty"`
	Align        string `yaml:"align,omitempty"`
	Fallback     string `yaml:"fallback,omitempty"`
}

// Region defines a rectangular area on the card
type Region struct {
	X      int `yaml:"x"`
	Y      int `yaml:"y"`
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// Font defines text rendering properties
type Font struct {
	Family string      `yaml:"family"`
	Size   interface{} `yaml:"size"` // Can be int or string template
	Weight string      `yaml:"weight,omitempty"`
	Style  string      `yaml:"style,omitempty"`
	Color  string      `yaml:"color"`
}

// Manager handles template loading and management
type Manager struct {
	customTemplateDir string
	templates         map[string]*Template
}

// NewManager creates a new template manager
func NewManager(customTemplateDir string) *Manager {
	return &Manager{
		customTemplateDir: customTemplateDir,
		templates:         make(map[string]*Template),
	}
}

// LoadTemplate loads a template by TCG and name
func (m *Manager) LoadTemplate(tcg, name string) (*Template, error) {
	key := fmt.Sprintf("%s/%s", tcg, name)

	// Check cache first
	if template, exists := m.templates[key]; exists {
		return template, nil
	}

	// Try custom template directory first
	if m.customTemplateDir != "" {
		templatePath := filepath.Join(m.customTemplateDir, tcg, name+".yaml")
		if template, err := m.loadTemplateFile(templatePath); err == nil {
			m.templates[key] = template
			return template, nil
		}
	}

	// Try built-in templates
	templatePath := filepath.Join("templates", tcg, name+".yaml")
	template, err := m.loadTemplateFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template %s not found: %v", key, err)
	}

	m.templates[key] = template
	return template, nil
}

// loadTemplateFile loads a template from a file
func (m *Manager) loadTemplateFile(filePath string) (*Template, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("error parsing template: %v", err)
	}

	template.TemplateDir = filepath.Dir(filePath)
	return &template, nil
}

// ValidateCard validates a card against this template
func (t *Template) ValidateCard(card *metadata.Card) error {
	// Check TCG match
	if card.TCG != t.TCG {
		return fmt.Errorf("card TCG '%s' doesn't match template TCG '%s'", card.TCG, t.TCG)
	}

	// Check required fields
	for _, field := range t.Required {
		if !t.hasField(card, field) {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Special validation: card.tcg must match template TCG
	if field := "card.tcg"; t.hasRequiredField(field) {
		if card.TCG != t.TCG {
			return fmt.Errorf("card TCG '%s' doesn't match template TCG '%s' - use a %s cardstyle for %s cards", card.TCG, t.TCG, card.TCG, card.TCG)
		}
	}

	return nil
}

// hasRequiredField checks if a field is in the required list
func (t *Template) hasRequiredField(field string) bool {
	for _, req := range t.Required {
		if req == field {
			return true
		}
	}
	return false
}

// hasField checks if a card has a specific field
func (t *Template) hasField(card *metadata.Card, field string) bool {
	switch field {
	case "card.tcg":
		return card.TCG != ""
	case "card.title":
		return card.Title != ""
	case "card.type":
		return card.Type != ""
	case "card.rarity":
		return card.Rarity != ""
	case "card.set":
		return card.Set != ""
	case "card.artist":
		return card.Artist != ""
	default:
		// Check in metadata map
		_, exists := card.Metadata[field]
		return exists
	}
}

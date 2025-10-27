package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
	"gopkg.in/yaml.v3"
)

// Embed built-in templates into the binary
//
//go:embed builtin/*
var builtinTemplates embed.FS

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
	customTemplateDir  string
	customCardstyleDir string
	templates          map[string]*Template
}

// NewManager creates a new template manager
func NewManager(customTemplateDir string) *Manager {
	// Set up custom cardstyle directory
	homeDir, _ := os.UserHomeDir()
	customCardstyleDir := filepath.Join(homeDir, ".tcg-cardgen", "cardstyles")

	return &Manager{
		customTemplateDir:  customTemplateDir,
		customCardstyleDir: customCardstyleDir,
		templates:          make(map[string]*Template),
	}
}

// LoadTemplate loads a template by TCG and cardstyle name
func (m *Manager) LoadTemplate(tcg, cardstyle string) (*Template, error) {
	key := fmt.Sprintf("%s/%s", tcg, cardstyle)

	// Check cache first
	if template, exists := m.templates[key]; exists {
		return template, nil
	}

	template, err := m.findAndLoadTemplate(tcg, cardstyle)
	if err != nil {
		return nil, fmt.Errorf("cardstyle %s/%s not found: %v", tcg, cardstyle, err)
	}

	m.templates[key] = template
	return template, nil
}

// findAndLoadTemplate searches for a template in various locations
func (m *Manager) findAndLoadTemplate(tcg, cardstyle string) (*Template, error) {
	// Search order (first found gets priority):
	// 1. Workspace cardstyles: templates/tcg/cardstyle.yaml (project-specific)
	// 2. User cardstyles: $HOME/.tcg-cardgen/cardstyles/tcg/cardstyle.yaml
	// 3. User cardstyles: $HOME/.tcg-cardgen/cardstyles/cardstyle.yaml (with TCG metadata check)
	// 4. Legacy custom template dir: custom-dir/tcg/cardstyle.yaml (for backwards compatibility)
	// 5. Embedded templates: builtin/tcg/cardstyle.yaml (final fallback)

	// 1. Workspace templates directory (project-specific cardstyles)
	workspacePath := filepath.Join("templates", tcg, cardstyle+".yaml")
	if template, err := m.loadAndProcessTemplate(workspacePath); err == nil {
		return template, nil
	}

	// 2. TCG-specific folder in user cardstyles
	if m.customCardstyleDir != "" {
		tcgPath := filepath.Join(m.customCardstyleDir, tcg, cardstyle+".yaml")
		if template, err := m.loadAndProcessTemplate(tcgPath); err == nil {
			return template, nil
		}

		// 3. Root level in user cardstyles (check TCG metadata)
		rootPath := filepath.Join(m.customCardstyleDir, cardstyle+".yaml")
		if template, err := m.loadAndProcessTemplate(rootPath); err == nil {
			// Verify TCG matches
			if template.TCG == tcg {
				return template, nil
			}
		}
	}

	// 4. Legacy custom template directory (for backwards compatibility)
	if m.customTemplateDir != "" {
		templatePath := filepath.Join(m.customTemplateDir, tcg, cardstyle+".yaml")
		if template, err := m.loadAndProcessTemplate(templatePath); err == nil {
			return template, nil
		}
	}

	// 5. Built-in embedded templates (final fallback)
	return m.loadBuiltinTemplate(tcg, cardstyle)
}

// loadBuiltinTemplate loads a template from embedded builtin templates
func (m *Manager) loadBuiltinTemplate(tcg, cardstyle string) (*Template, error) {
	builtinPath := fmt.Sprintf("builtin/%s/%s.yaml", tcg, cardstyle)

	data, err := builtinTemplates.ReadFile(builtinPath)
	if err != nil {
		return nil, fmt.Errorf("builtin template %s/%s not found: %v", tcg, cardstyle, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("error parsing builtin template: %v", err)
	}

	// Set template directory for builtin templates
	template.TemplateDir = fmt.Sprintf("builtin/%s", tcg)

	// Handle inheritance for builtin templates
	if template.Extends != "" {
		// For builtin templates, resolve relative extends within builtin
		baseTemplate, err := m.resolveBuiltinBaseTemplate(template.Extends, template.TemplateDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load builtin base template '%s': %v", template.Extends, err)
		}
		merged := m.mergeTemplates(baseTemplate, &template)
		template = *merged
	}

	return &template, nil
}

// resolveBuiltinBaseTemplate resolves extends for builtin templates
func (m *Manager) resolveBuiltinBaseTemplate(extendsPath, currentDir string) (*Template, error) {
	// Handle relative paths within builtin templates
	var basePath string
	if strings.HasPrefix(extendsPath, "./") {
		// Relative to current builtin directory
		basePath = filepath.Join(currentDir, extendsPath[2:])
	} else {
		basePath = extendsPath
	}

	// Ensure it's still a builtin path
	if !strings.HasPrefix(basePath, "builtin/") {
		basePath = filepath.Join("builtin", basePath)
	}

	data, err := builtinTemplates.ReadFile(basePath)
	if err != nil {
		return nil, err
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("error parsing builtin base template: %v", err)
	}

	template.TemplateDir = filepath.Dir(basePath)

	// Handle recursive inheritance
	if template.Extends != "" {
		baseTemplate, err := m.resolveBuiltinBaseTemplate(template.Extends, template.TemplateDir)
		if err != nil {
			return nil, err
		}
		template = *m.mergeTemplates(baseTemplate, &template)
	}

	return &template, nil
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

// loadAndProcessTemplate loads a template and handles inheritance
func (m *Manager) loadAndProcessTemplate(filePath string) (*Template, error) {
	// Load the base template
	template, err := m.loadTemplateFile(filePath)
	if err != nil {
		return nil, err
	}

	// If this template extends another, load and merge the base
	if template.Extends != "" {
		baseTemplate, err := m.resolveBaseTemplate(template.Extends, template.TemplateDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load base template '%s': %v", template.Extends, err)
		}

		// Merge base template into this template
		template = m.mergeTemplates(baseTemplate, template)
	}

	return template, nil
}

// resolveBaseTemplate resolves the path to a base template
func (m *Manager) resolveBaseTemplate(extendsPath, currentDir string) (*Template, error) {
	var basePath string

	// Handle relative paths
	if !filepath.IsAbs(extendsPath) {
		basePath = filepath.Join(currentDir, extendsPath)
	} else {
		basePath = extendsPath
	}

	// Load the base template (this will handle recursive inheritance)
	return m.loadAndProcessTemplate(basePath)
}

// mergeTemplates merges a base template with an extending template
func (m *Manager) mergeTemplates(base, extended *Template) *Template {
	// Start with a copy of the extended template
	result := *extended
	result.BaseTemplate = base

	// Merge dimensions if not set in extended
	if result.Dimensions.Width == 0 {
		result.Dimensions = base.Dimensions
	}

	// Merge required fields (base + extended)
	requiredMap := make(map[string]bool)
	for _, field := range base.Required {
		requiredMap[field] = true
	}
	for _, field := range extended.Required {
		requiredMap[field] = true
	}
	result.Required = make([]string, 0, len(requiredMap))
	for field := range requiredMap {
		result.Required = append(result.Required, field)
	}

	// Merge optional fields (base defaults, extended overrides)
	if result.Optional == nil {
		result.Optional = make(map[string]interface{})
	}
	for key, value := range base.Optional {
		if _, exists := result.Optional[key]; !exists {
			result.Optional[key] = value
		}
	}

	// Merge style tokens (base defaults, extended overrides)
	if result.StyleTokens == nil {
		result.StyleTokens = make(map[string]string)
	}
	for key, value := range base.StyleTokens {
		if _, exists := result.StyleTokens[key]; !exists {
			result.StyleTokens[key] = value
		}
	}

	// Merge icons (base defaults, extended overrides)
	if result.Icons == nil {
		result.Icons = make(map[string]string)
	}
	for key, value := range base.Icons {
		if _, exists := result.Icons[key]; !exists {
			result.Icons[key] = value
		}
	}

	// Handle layers - extended layers come after base layers, but can override by name
	baseLayers := make(map[string]Layer)
	for _, layer := range base.Layers {
		baseLayers[layer.Name] = layer
	}

	// Apply overrides first
	for _, override := range result.Overrides {
		if baseLayer, exists := baseLayers[override.Layer]; exists {
			// Apply override to base layer
			modifiedLayer := m.applyLayerOverride(baseLayer, override)
			baseLayers[override.Layer] = modifiedLayer
		}
	}

	// Build final layers list
	finalLayers := make([]Layer, 0)
	layerNames := make(map[string]bool)

	// Add base layers first (with any overrides applied)
	for _, layer := range base.Layers {
		if modifiedLayer, exists := baseLayers[layer.Name]; exists {
			finalLayers = append(finalLayers, modifiedLayer)
			layerNames[layer.Name] = true
		}
	}

	// Add extended layers that don't override base layers
	for _, layer := range extended.Layers {
		if !layerNames[layer.Name] {
			finalLayers = append(finalLayers, layer)
		}
	}

	// Add any additional layers
	finalLayers = append(finalLayers, result.AddLayers...)

	result.Layers = finalLayers
	return &result
}

// applyLayerOverride applies override settings to a layer
func (m *Manager) applyLayerOverride(layer Layer, override LayerOverride) Layer {
	// This is a simplified implementation - in practice you'd want to handle
	// field-specific merging for complex nested structures
	modified := layer

	for key, value := range override.Updates {
		switch key {
		case "source":
			if str, ok := value.(string); ok {
				modified.Source = str
			}
		case "content":
			if str, ok := value.(string); ok {
				modified.Content = str
			}
		case "condition":
			if str, ok := value.(string); ok {
				modified.Condition = str
			}
		case "fit_mode":
			if str, ok := value.(string); ok {
				modified.FitMode = str
			}
			// Add more field overrides as needed
		}
	}

	return modified
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
		return card.TCG != "" || t.hasNestedField(card, "card", "tcg")
	case "card.cardstyle":
		return card.CardStyle != "" || t.hasNestedField(card, "card", "cardstyle")
	case "card.title":
		return card.Title != "" || t.hasNestedField(card, "card", "title")
	case "card.type":
		return card.Type != "" || t.hasNestedField(card, "card", "type")
	case "card.rarity":
		return card.Rarity != "" || t.hasNestedField(card, "card", "rarity")
	case "card.set":
		return card.Set != "" || t.hasNestedField(card, "card", "set")
	case "card.artist":
		return card.Artist != "" || t.hasNestedField(card, "card", "artist")
	default:
		// Check in metadata map (both flat and nested)
		if _, exists := card.Metadata[field]; exists {
			return true
		}

		// Check nested field (e.g., "mtg.cmc" -> card.Metadata["mtg"]["cmc"])
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			return t.hasNestedField(card, parts[0], parts[1])
		}

		return false
	}
}

// hasNestedField checks if a nested field exists in metadata
func (t *Template) hasNestedField(card *metadata.Card, section, field string) bool {
	if sectionData, exists := card.Metadata[section]; exists {
		if sectionMap, ok := sectionData.(map[string]interface{}); ok {
			value, exists := sectionMap[field]
			if exists {
				// Check if the value is not nil and not empty string
				if str, ok := value.(string); ok {
					return str != ""
				}
				return value != nil
			}
		}
	}
	return false
}

// CardStyleInfo represents information about a discovered cardstyle
type CardStyleInfo struct {
	TCG         string
	Name        string
	DisplayName string
	Description string
	Version     string
	Source      string // "built-in" or path to custom cardstyle
	Extends     string // Base template it extends
}

// ListAvailableCardstyles discovers and lists all available cardstyles
func (m *Manager) ListAvailableCardstyles() ([]CardStyleInfo, error) {
	var allCardstyles []CardStyleInfo
	seen := make(map[string]bool) // Track TCG/cardstyle combinations

	// 1. Discover workspace cardstyles from templates/ directory (highest priority)
	workspaceStyles, err := m.discoverWorkspaceCardstyles()
	if err == nil {
		for _, style := range workspaceStyles {
			key := fmt.Sprintf("%s/%s", style.TCG, style.Name)
			if !seen[key] {
				allCardstyles = append(allCardstyles, style)
				seen[key] = true
			}
		}
	}

	// 2. Discover user cardstyles from $HOME/.tcg-cardgen/cardstyles
	if m.customCardstyleDir != "" {
		userStyles, err := m.discoverUserCardstyles()
		if err == nil {
			for _, style := range userStyles {
				key := fmt.Sprintf("%s/%s", style.TCG, style.Name)
				if !seen[key] {
					allCardstyles = append(allCardstyles, style)
					seen[key] = true
				}
			}
		}
	}

	// 3. Discover legacy custom templates (for backwards compatibility)
	if m.customTemplateDir != "" {
		legacyStyles, err := m.discoverLegacyTemplates()
		if err == nil {
			for _, style := range legacyStyles {
				key := fmt.Sprintf("%s/%s", style.TCG, style.Name)
				if !seen[key] {
					allCardstyles = append(allCardstyles, style)
					seen[key] = true
				}
			}
		}
	}

	// 4. Discover embedded built-in cardstyles (fallback)
	embeddedStyles, err := m.discoverEmbeddedCardstyles()
	if err == nil {
		for _, style := range embeddedStyles {
			key := fmt.Sprintf("%s/%s", style.TCG, style.Name)
			if !seen[key] {
				allCardstyles = append(allCardstyles, style)
				seen[key] = true
			}
		}
	}

	return allCardstyles, nil
}

// discoverEmbeddedCardstyles finds embedded built-in cardstyles
func (m *Manager) discoverEmbeddedCardstyles() ([]CardStyleInfo, error) {
	var cardstyles []CardStyleInfo

	// Read the builtin directory from embedded filesystem
	entries, err := builtinTemplates.ReadDir("builtin")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		tcgName := entry.Name()
		tcgPath := "builtin/" + tcgName

		// Read cardstyle files in this TCG directory
		cardstyleEntries, err := builtinTemplates.ReadDir(tcgPath)
		if err != nil {
			continue
		}

		for _, file := range cardstyleEntries {
			if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
				continue
			}

			styleName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			// Create CardStyleInfo for embedded template
			info := &CardStyleInfo{
				TCG:         tcgName,
				Name:        styleName,
				DisplayName: fmt.Sprintf("%s %s", strings.ToUpper(tcgName), strings.Title(styleName)),
				Description: fmt.Sprintf("Built-in %s %s cardstyle", strings.ToUpper(tcgName), styleName),
				Version:     "embedded",
				Source:      "embedded",
				Extends:     "", // Will be determined when loading
			}

			// Try to load the template to get extends information
			if template, err := m.loadEmbeddedTemplateInfo(tcgPath + "/" + file.Name()); err == nil {
				if template.Extends != "" {
					info.Extends = template.Extends
				}
				if template.Name != "" {
					info.DisplayName = template.Name
				}
				if template.Description != "" {
					info.Description = template.Description
				}
				if template.Version != "" {
					info.Version = template.Version
				}
			}

			cardstyles = append(cardstyles, *info)
		}
	}

	return cardstyles, nil
}

// loadEmbeddedTemplateInfo loads template metadata from embedded filesystem
func (m *Manager) loadEmbeddedTemplateInfo(embeddedPath string) (*Template, error) {
	data, err := builtinTemplates.ReadFile(embeddedPath)
	if err != nil {
		return nil, err
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// discoverWorkspaceCardstyles finds workspace cardstyles in templates/ directory
func (m *Manager) discoverWorkspaceCardstyles() ([]CardStyleInfo, error) {
	var cardstyles []CardStyleInfo

	templatesDir := "templates"
	tcgDirs, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	for _, tcgDir := range tcgDirs {
		if !tcgDir.IsDir() {
			continue
		}

		tcgName := tcgDir.Name()
		tcgPath := filepath.Join(templatesDir, tcgName)

		cardstyleFiles, err := os.ReadDir(tcgPath)
		if err != nil {
			continue
		}

		for _, file := range cardstyleFiles {
			if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
				continue
			}

			styleName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			stylePath := filepath.Join(tcgPath, file.Name())

			info, err := m.getCardstyleInfo(stylePath, tcgName, styleName, "workspace")
			if err == nil {
				cardstyles = append(cardstyles, *info)
			}
		}
	}

	return cardstyles, nil
}

// discoverUserCardstyles finds user cardstyles in $HOME/.tcg-cardgen/cardstyles
func (m *Manager) discoverUserCardstyles() ([]CardStyleInfo, error) {
	var cardstyles []CardStyleInfo

	if _, err := os.Stat(m.customCardstyleDir); os.IsNotExist(err) {
		return cardstyles, nil // Directory doesn't exist, return empty list
	}

	// Check for TCG-specific subdirectories
	tcgDirs, err := os.ReadDir(m.customCardstyleDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range tcgDirs {
		if entry.IsDir() {
			// TCG-specific directory (e.g., mtg/, pokemon/)
			tcgName := entry.Name()
			tcgPath := filepath.Join(m.customCardstyleDir, tcgName)

			cardstyleFiles, err := os.ReadDir(tcgPath)
			if err != nil {
				continue
			}

			for _, file := range cardstyleFiles {
				if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
					continue
				}

				styleName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				stylePath := filepath.Join(tcgPath, file.Name())

				info, err := m.getCardstyleInfo(stylePath, tcgName, styleName, "user")
				if err == nil {
					cardstyles = append(cardstyles, *info)
				}
			}
		} else if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml") {
			// Root-level cardstyle file (TCG determined by metadata)
			styleName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			stylePath := filepath.Join(m.customCardstyleDir, entry.Name())

			// Load template to get TCG from metadata
			template, err := m.loadTemplateFile(stylePath)
			if err != nil {
				continue
			}

			info, err := m.getCardstyleInfo(stylePath, template.TCG, styleName, "user")
			if err == nil {
				cardstyles = append(cardstyles, *info)
			}
		}
	}

	return cardstyles, nil
}

// discoverLegacyTemplates finds templates in legacy custom template directory
func (m *Manager) discoverLegacyTemplates() ([]CardStyleInfo, error) {
	var cardstyles []CardStyleInfo

	tcgDirs, err := os.ReadDir(m.customTemplateDir)
	if err != nil {
		return nil, err
	}

	for _, tcgDir := range tcgDirs {
		if !tcgDir.IsDir() {
			continue
		}

		tcgName := tcgDir.Name()
		tcgPath := filepath.Join(m.customTemplateDir, tcgName)

		cardstyleFiles, err := os.ReadDir(tcgPath)
		if err != nil {
			continue
		}

		for _, file := range cardstyleFiles {
			if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
				continue
			}

			styleName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			stylePath := filepath.Join(tcgPath, file.Name())

			info, err := m.getCardstyleInfo(stylePath, tcgName, styleName, "legacy")
			if err == nil {
				cardstyles = append(cardstyles, *info)
			}
		}
	}

	return cardstyles, nil
}

// getCardstyleInfo extracts metadata from a cardstyle file
func (m *Manager) getCardstyleInfo(filePath, tcg, name, source string) (*CardStyleInfo, error) {
	template, err := m.loadTemplateFile(filePath)
	if err != nil {
		return nil, err
	}

	info := &CardStyleInfo{
		TCG:         tcg,
		Name:        name,
		DisplayName: template.Name,
		Description: template.Description,
		Version:     template.Version,
		Source:      source,
		Extends:     template.Extends,
	}

	if source != "built-in" {
		info.Source = filePath
	}

	return info, nil
}

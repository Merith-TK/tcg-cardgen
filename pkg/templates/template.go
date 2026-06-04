// Package templates handles loading, caching, inheritance, and validation of
// card style template YAML files.
//
// Template search order (first match wins):
//  1. Workspace: .tcg-cardstyles/<tcg>/<name>.yaml
//  2. User:      $HOME/.tcg-cardgen/cardstyles/<tcg>/<name>.yaml
//  3. User:      $HOME/.tcg-cardgen/cardstyles/<name>.yaml  (TCG from file metadata)
//  4. Custom:    <customTemplateDir>/<tcg>/<name>.yaml       (legacy --template-dir flag)
//  5. Embedded:  compiled-in built-in templates
package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/card"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
	"gopkg.in/yaml.v3"
)

// Embed built-in templates into the binary.
//
//go:embed templates/*
var builtinTemplates embed.FS

// Template is a fully-resolved card style template ready for rendering.
type Template struct {
	Name        string `yaml:"name"`
	TCG         string `yaml:"tcg"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	// Extends is a relative or absolute path to a base template file.
	// After loading, all inheritance is already merged into this struct.
	Extends string `yaml:"extends,omitempty"`

	Dimensions  types.Dimensions       `yaml:"dimensions"`
	Layers      []types.Layer          `yaml:"layers"`
	Required    []string               `yaml:"required_fields"`
	Optional    map[string]interface{} `yaml:"optional_fields"`
	Icons       map[string]string      `yaml:"icons"`
	StyleTokens map[string]string      `yaml:"style_tokens"`
	// Overrides patches named layers from the base template.
	Overrides []types.LayerOverride `yaml:"overrides,omitempty"`
	// AddLayers are appended after all inherited layers.
	AddLayers []types.Layer `yaml:"additional_layers,omitempty"`

	// TemplateDir is the real on-disk directory containing this template.
	// Always an OS path; never the virtual embed:// path.
	TemplateDir string `yaml:"-"`
}

// Manager loads and caches card style templates.
type Manager struct {
	customTemplateDir  string // legacy --template-dir flag
	customCardstyleDir string // $HOME/.tcg-cardgen/cardstyles
	cache              map[string]*Template
	// embeddedExtractDir is set the first time we need a real path for an
	// embedded template asset; it holds a temp dir with extracted files.
	embeddedExtractDir string
}

// NewManager creates a Manager.  customTemplateDir may be empty.
func NewManager(customTemplateDir string) *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		customTemplateDir:  customTemplateDir,
		customCardstyleDir: filepath.Join(home, ".tcg-cardgen", "cardstyles"),
		cache:              make(map[string]*Template),
	}
}

// ─── public API ──────────────────────────────────────────────────────────────

// LoadTemplate loads (or returns from cache) the template for the given TCG and
// card style name.
func (m *Manager) LoadTemplate(tcg, name string) (*Template, error) {
	key := tcg + "/" + name
	if t, ok := m.cache[key]; ok {
		return t, nil
	}
	t, err := m.findAndLoad(tcg, name)
	if err != nil {
		return nil, fmt.Errorf("cardstyle %s/%s: %w", tcg, name, err)
	}
	m.cache[key] = t
	return t, nil
}

// ValidateCard returns an error if the card is missing required fields or the
// TCG doesn't match.
func (t *Template) ValidateCard(c *card.Card) error {
	if t.TCG != "" && c.TCG != t.TCG {
		return fmt.Errorf("card TCG %q does not match template TCG %q", c.TCG, t.TCG)
	}
	for _, field := range t.Required {
		if !hasField(c, field) {
			return fmt.Errorf("required field %q is missing", field)
		}
	}
	return nil
}

// ListAvailableCardstyles discovers all templates across all sources.
func (m *Manager) ListAvailableCardstyles() ([]types.CardStyleInfo, error) {
	var all []types.CardStyleInfo
	seen := make(map[string]bool)

	add := func(styles []types.CardStyleInfo) {
		for _, s := range styles {
			k := s.TCG + "/" + s.Name
			if !seen[k] {
				all = append(all, s)
				seen[k] = true
			}
		}
	}

	add(m.discoverDir(".tcg-cardstyles", "workspace"))
	add(m.discoverDir(m.customCardstyleDir, "user"))
	if m.customTemplateDir != "" {
		add(m.discoverDir(m.customTemplateDir, "custom"))
	}
	add(m.discoverEmbedded())
	return all, nil
}

// ─── template finding ─────────────────────────────────────────────────────────

func (m *Manager) findAndLoad(tcg, name string) (*Template, error) {
	candidates := []string{
		filepath.Join(".tcg-cardstyles", tcg, name+".yaml"),
	}

	if m.customCardstyleDir != "" {
		candidates = append(candidates,
			filepath.Join(m.customCardstyleDir, tcg, name+".yaml"),
			filepath.Join(m.customCardstyleDir, name+".yaml"),
		)
	}

	if m.customTemplateDir != "" {
		candidates = append(candidates,
			filepath.Join(m.customTemplateDir, tcg, name+".yaml"),
		)
	}

	for _, path := range candidates {
		t, err := m.loadFile(path)
		if err != nil {
			continue
		}
		// Root-level user files have TCG in metadata — verify it matches.
		if t.TCG != "" && t.TCG != tcg {
			continue
		}
		return t, nil
	}

	// Fall back to embedded.
	return m.loadEmbedded(tcg, name)
}

// ─── file loading ─────────────────────────────────────────────────────────────

// loadFile loads a template from an OS path, resolving inheritance.
func (m *Manager) loadFile(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t, err := unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	t.TemplateDir = filepath.Dir(path)

	if t.Extends != "" {
		basePath := t.Extends
		if !filepath.IsAbs(basePath) {
			basePath = filepath.Join(t.TemplateDir, basePath)
		}
		base, err := m.loadFile(basePath)
		if err != nil {
			return nil, fmt.Errorf("load base template %q for %s: %w", t.Extends, path, err)
		}
		t = merge(base, t)
	}
	return t, nil
}

// loadEmbedded loads a template from the embedded FS, resolving inheritance.
func (m *Manager) loadEmbedded(tcg, name string) (*Template, error) {
	path := fmt.Sprintf("templates/%s/%s.yaml", tcg, name)
	data, err := builtinTemplates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded template not found")
	}
	t, err := unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parse embedded %s: %w", path, err)
	}

	// For embedded templates the TemplateDir is set to an extracted OS path
	// the first time we need it, so asset references (fonts, icons) work.
	t.TemplateDir = m.embeddedDir(tcg)

	if t.Extends != "" {
		base, err := m.resolveEmbeddedExtends(t.Extends, "templates/"+tcg)
		if err != nil {
			return nil, fmt.Errorf("load embedded base %q: %w", t.Extends, err)
		}
		t = merge(base, t)
	}
	return t, nil
}

func (m *Manager) resolveEmbeddedExtends(extendsPath, currentEmbedDir string) (*Template, error) {
	var embedPath string
	if strings.HasPrefix(extendsPath, "./") {
		embedPath = currentEmbedDir + "/" + extendsPath[2:]
	} else if strings.HasPrefix(extendsPath, "templates/") {
		embedPath = extendsPath
	} else {
		embedPath = currentEmbedDir + "/" + extendsPath
	}

	data, err := builtinTemplates.ReadFile(embedPath)
	if err != nil {
		return nil, err
	}
	t, err := unmarshal(data)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(embedPath)
	tcg := filepath.Base(dir)
	t.TemplateDir = m.embeddedDir(tcg)

	if t.Extends != "" {
		base, err := m.resolveEmbeddedExtends(t.Extends, dir)
		if err != nil {
			return nil, err
		}
		t = merge(base, t)
	}
	return t, nil
}

// embeddedDir returns the real OS path for an embedded template's directory.
// On first call for a given TCG it extracts the embedded assets to a temp dir.
func (m *Manager) embeddedDir(tcg string) string {
	if m.embeddedExtractDir == "" {
		dir, err := os.MkdirTemp("", "tcg-cardgen-embedded-*")
		if err != nil {
			// Best-effort: return a virtual path that won't resolve.
			return "templates/" + tcg
		}
		m.embeddedExtractDir = dir
	}
	tcgDir := filepath.Join(m.embeddedExtractDir, tcg)
	_ = os.MkdirAll(tcgDir, 0755)
	return tcgDir
}

// ─── template merging ─────────────────────────────────────────────────────────

// merge produces a new template where `extended` overrides `base`.
//
// Merge rules:
//   - Dimensions: extended wins if non-zero, else base.
//   - Required fields: union of both sets.
//   - Optional fields / StyleTokens / Icons: base provides defaults; extended overrides.
//   - Layers: start from base layers; extended layers with the same Name replace
//     the corresponding base layer; remaining extended layers are appended;
//     AddLayers from extended are appended last.
//   - Overrides from extended are applied to the final layer list.
func merge(base, extended *Template) *Template {
	result := *extended

	// Dimensions
	if result.Dimensions.Width == 0 {
		result.Dimensions = base.Dimensions
	}

	// Required fields (union)
	reqSeen := make(map[string]bool, len(base.Required)+len(extended.Required))
	for _, f := range base.Required {
		reqSeen[f] = true
	}
	for _, f := range extended.Required {
		reqSeen[f] = true
	}
	result.Required = make([]string, 0, len(reqSeen))
	for f := range reqSeen {
		result.Required = append(result.Required, f)
	}

	// Optional fields — base provides defaults
	if result.Optional == nil {
		result.Optional = make(map[string]interface{})
	}
	for k, v := range base.Optional {
		if _, exists := result.Optional[k]; !exists {
			result.Optional[k] = v
		}
	}

	// StyleTokens — base provides defaults
	if result.StyleTokens == nil {
		result.StyleTokens = make(map[string]string)
	}
	for k, v := range base.StyleTokens {
		if _, exists := result.StyleTokens[k]; !exists {
			result.StyleTokens[k] = v
		}
	}

	// Icons — base provides defaults
	if result.Icons == nil {
		result.Icons = make(map[string]string)
	}
	for k, v := range base.Icons {
		if _, exists := result.Icons[k]; !exists {
			result.Icons[k] = v
		}
	}

	// Layers — extended layers override base layers by name; extra ones are appended.
	extendedByName := make(map[string]types.Layer, len(extended.Layers))
	for _, l := range extended.Layers {
		if l.Name != "" {
			extendedByName[l.Name] = l
		}
	}

	var finalLayers []types.Layer
	usedNames := make(map[string]bool)

	for _, bl := range base.Layers {
		if el, ok := extendedByName[bl.Name]; ok {
			finalLayers = append(finalLayers, el) // extended wins
		} else {
			finalLayers = append(finalLayers, bl)
		}
		usedNames[bl.Name] = true
	}

	// Append extended layers not present in base (preserving order).
	for _, el := range extended.Layers {
		if !usedNames[el.Name] {
			finalLayers = append(finalLayers, el)
		}
	}

	// Apply overrides from extended.
	for _, ov := range extended.Overrides {
		for i, l := range finalLayers {
			if l.Name == ov.Layer {
				finalLayers[i] = applyOverride(l, ov)
				break
			}
		}
	}

	// Append additional layers.
	finalLayers = append(finalLayers, extended.AddLayers...)

	result.Layers = finalLayers
	result.TemplateDir = extended.TemplateDir
	return &result
}

// applyOverride patches a layer with the fields specified in the override.
func applyOverride(l types.Layer, ov types.LayerOverride) types.Layer {
	for k, v := range ov.Updates {
		str, isStr := v.(string)
		switch k {
		case "source":
			if isStr {
				l.Source = str
			}
		case "content":
			if isStr {
				l.Content = str
			}
		case "condition":
			if isStr {
				l.Condition = str
			}
		case "fit_mode":
			if isStr {
				l.FitMode = str
			}
		case "fill":
			if isStr {
				l.Fill = str
			}
		case "stroke":
			if isStr {
				l.Stroke = str
			}
		case "align":
			if isStr {
				l.Align = str
			}
		case "fallback":
			if isStr {
				l.Fallback = str
			}
		}
	}
	return l
}

// ─── discovery ────────────────────────────────────────────────────────────────

func (m *Manager) discoverDir(root, source string) []types.CardStyleInfo {
	var out []types.CardStyleInfo

	tcgDirs, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	for _, entry := range tcgDirs {
		if !entry.IsDir() {
			// Root-level .yaml (TCG from metadata)
			if !isYAML(entry.Name()) {
				continue
			}
			path := filepath.Join(root, entry.Name())
			info := quickInfo(path)
			if info == nil {
				continue
			}
			info.Source = source
			if source != "embedded" {
				info.Source = path
			}
			out = append(out, *info)
			continue
		}

		tcg := entry.Name()
		tcgPath := filepath.Join(root, tcg)
		files, err := os.ReadDir(tcgPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !isYAML(file.Name()) {
				continue
			}
			path := filepath.Join(tcgPath, file.Name())
			info := quickInfo(path)
			if info == nil {
				continue
			}
			if info.TCG == "" {
				info.TCG = tcg
			}
			info.Source = source
			if source != "embedded" {
				info.Source = path
			}
			out = append(out, *info)
		}
	}
	return out
}

func (m *Manager) discoverEmbedded() []types.CardStyleInfo {
	var out []types.CardStyleInfo

	tcgDirs, err := builtinTemplates.ReadDir("templates")
	if err != nil {
		return nil
	}
	for _, tcgDir := range tcgDirs {
		if !tcgDir.IsDir() {
			continue
		}
		tcg := tcgDir.Name()
		files, err := builtinTemplates.ReadDir("templates/" + tcg)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() || !isYAML(file.Name()) {
				continue
			}
			embedPath := "templates/" + tcg + "/" + file.Name()
			data, err := builtinTemplates.ReadFile(embedPath)
			if err != nil {
				continue
			}
			var t Template
			if err := yaml.Unmarshal(data, &t); err != nil {
				continue
			}
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			dn := t.Name
			if dn == "" {
				dn = strings.ToUpper(tcg) + " " + name
			}
			out = append(out, types.CardStyleInfo{
				TCG:         tcg,
				Name:        name,
				DisplayName: dn,
				Description: t.Description,
				Version:     t.Version,
				Source:      "embedded",
				Extends:     t.Extends,
			})
		}
	}
	return out
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func unmarshal(data []byte) (*Template, error) {
	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func quickInfo(path string) *types.CardStyleInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dn := t.Name
	if dn == "" {
		dn = name
	}
	return &types.CardStyleInfo{
		TCG:         t.TCG,
		Name:        name,
		DisplayName: dn,
		Description: t.Description,
		Version:     t.Version,
		Extends:     t.Extends,
	}
}

func isYAML(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}

// hasField checks whether a card has a particular field populated.
// It checks struct fields by their canonical dot-path name and also looks
// in the flat Metadata map.
func hasField(c *card.Card, field string) bool {
	switch field {
	case "tcg", "card.tcg":
		return c.TCG != ""
	case "cardstyle", "card.cardstyle":
		return c.CardStyle != ""
	case "title", "card.title":
		return c.Title != ""
	case "card_type", "type", "card.type":
		return c.Type != ""
	case "rarity", "card.rarity":
		return c.Rarity != ""
	case "set", "card.set":
		return c.Set != ""
	case "artist", "card.artist":
		return c.Artist != ""
	}

	// Check flat metadata key.
	if _, ok := c.Metadata[field]; ok {
		return true
	}

	// Check nested key (e.g. "mtg.color" -> Metadata["mtg"]["color"]).
	parts := strings.SplitN(field, ".", 2)
	if len(parts) == 2 {
		if section, ok := c.Metadata[parts[0]]; ok {
			if m, ok := section.(map[string]interface{}); ok {
				if v, ok := m[parts[1]]; ok {
					if s, ok := v.(string); ok {
						return s != ""
					}
					return v != nil
				}
			}
		}
	}

	return false
}

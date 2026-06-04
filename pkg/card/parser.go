// Package card handles parsing of .md card definition files.
//
// Card files use YAML frontmatter between --- delimiters, followed by a
// Markdown body.  The frontmatter uses flat keys (not dotted paths):
//
//	---
//	tcg: mtg
//	cardstyle: basic
//	title: Lightning Bolt
//	rarity: common
//	set: Alpha
//	artist: Christopher Rush
//	mtg:
//	  color: red
//	  cmc: 1
//	card:
//	  artwork: path/to/art.png
//	---
//
//	Rules text goes here.
//
//	---
//
//	*Flavor text in italics after a horizontal rule.*
package card

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Card represents a fully parsed card definition.
type Card struct {
	// Core fields — populated from YAML frontmatter.
	TCG       string `yaml:"tcg"`
	CardStyle string `yaml:"cardstyle"`
	Title     string `yaml:"title"`
	Type      string `yaml:"card_type"` // use card_type to avoid clashing with Go keyword
	Rarity    string `yaml:"rarity"`
	Set       string `yaml:"set"`
	Artist    string `yaml:"artist"`

	// Print sheet info.
	PrintThis  int `yaml:"print_this"`
	PrintTotal int `yaml:"print_total"`

	// Content parsed from the Markdown body.
	RulesText  string `yaml:"-"`
	FlavorText string `yaml:"-"`
	ManaCost   string `yaml:"-"` // extracted from "> {{...}}" blockquote

	// All frontmatter fields (including nested maps like mtg.color).
	// Used to build the full variable map for template rendering.
	Metadata map[string]interface{} `yaml:",inline"`

	// Source file path.
	SourceFile string `yaml:"-"`
}

// ParseFile reads a .md file and returns a populated Card.
func ParseFile(filePath string) (*Card, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var frontmatterLines []string
	var bodyLines []string
	hasFrontmatter := false

	// Peek at the first line.
	if scanner.Scan() {
		if scanner.Text() == "---" {
			hasFrontmatter = true
			// Read until closing ---
			for scanner.Scan() {
				line := scanner.Text()
				if line == "---" {
					break
				}
				frontmatterLines = append(frontmatterLines, line)
			}
		} else {
			bodyLines = append(bodyLines, scanner.Text())
		}
	}

	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", filePath, err)
	}

	c := &Card{
		Metadata:   make(map[string]interface{}),
		SourceFile: filePath,
	}

	if hasFrontmatter && len(frontmatterLines) > 0 {
		raw := strings.Join(frontmatterLines, "\n")

		// Parse into Metadata map first (captures all keys including nested ones).
		if err := yaml.Unmarshal([]byte(raw), &c.Metadata); err != nil {
			return nil, fmt.Errorf("parse YAML frontmatter in %s: %w", filePath, err)
		}

		// Parse into the struct for the fields we care about by name.
		if err := yaml.Unmarshal([]byte(raw), c); err != nil {
			return nil, fmt.Errorf("parse YAML struct in %s: %w", filePath, err)
		}

		// If the YAML has a nested "card" key, pull well-known sub-fields up.
		if cardMap, ok := c.Metadata["card"]; ok {
			if m, ok := cardMap.(map[string]interface{}); ok {
				if v, ok := m["tcg"].(string); ok && c.TCG == "" {
					c.TCG = v
				}
				if v, ok := m["cardstyle"].(string); ok && c.CardStyle == "" {
					c.CardStyle = v
				}
				if v, ok := m["title"].(string); ok && c.Title == "" {
					c.Title = v
				}
				if v, ok := m["card_type"].(string); ok && c.Type == "" {
					c.Type = v
				}
				if v, ok := m["type"].(string); ok && c.Type == "" {
					c.Type = v
				}
				if v, ok := m["rarity"].(string); ok && c.Rarity == "" {
					c.Rarity = v
				}
				if v, ok := m["set"].(string); ok && c.Set == "" {
					c.Set = v
				}
				if v, ok := m["artist"].(string); ok && c.Artist == "" {
					c.Artist = v
				}
			}
		}
	}

	parseBody(c, strings.Join(bodyLines, "\n"))
	setDefaults(c, filePath)

	return c, nil
}

// parseBody extracts title (from # header), mana cost (from > {{...}} blockquote),
// card type (from > **...** blockquote), rules text, and flavor text from the
// Markdown body.
func parseBody(c *Card, body string) {
	lines := strings.Split(body, "\n")

	var rulesLines []string
	var flavorLines []string
	inFlavor := false

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		// Title from "# Heading" — only if not already set in frontmatter.
		if c.Title == "" && strings.HasPrefix(line, "# ") {
			c.Title = strings.TrimSpace(line[2:])
			continue
		}

		// Mana cost from "> {{...}}" blockquote.
		if strings.HasPrefix(line, "> {{") && strings.HasSuffix(line, "}}") {
			if c.ManaCost == "" {
				c.ManaCost = strings.TrimSpace(line[2:])
			}
			continue
		}

		// Card type from "> **...**" blockquote.
		if strings.HasPrefix(line, "> **") && strings.HasSuffix(line, "**") {
			if c.Type == "" {
				c.Type = strings.TrimSpace(line[4 : len(line)-2])
			}
			continue
		}

		// Flavor text separator (horizontal rule).
		if line == "---" || line == "-----" || line == "***" {
			inFlavor = true
			continue
		}

		// Skip blank lines and headings inside the body.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if inFlavor {
			// Strip surrounding * or ** from flavor lines.
			if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") && len(line) > 2 {
				flavorLines = append(flavorLines, line[1:len(line)-1])
			} else {
				flavorLines = append(flavorLines, line)
			}
		} else {
			rulesLines = append(rulesLines, line)
		}
	}

	c.RulesText = strings.Join(rulesLines, "\n\n")
	c.FlavorText = strings.Join(flavorLines, "\n")
}

// setDefaults fills in sensible defaults for empty fields.
func setDefaults(c *Card, filePath string) {
	if c.Title == "" {
		base := filepath.Base(filePath)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		name = strings.ReplaceAll(name, "_", " ")
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + name[1:]
		}
		c.Title = name
	}

	if c.PrintThis == 0 {
		c.PrintThis = 1
	}
	if c.PrintTotal == 0 {
		c.PrintTotal = 1
	}
	if c.Rarity == "" {
		c.Rarity = "common"
	}
	if c.Set == "" {
		c.Set = "Unknown"
	}
	if c.Artist == "" {
		c.Artist = "Unknown Artist"
	}
	if c.TCG == "" {
		c.TCG = "mtg"
	}
	if c.CardStyle == "" {
		c.CardStyle = "basic"
	}
}
